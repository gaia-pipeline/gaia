package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	pb "github.com/gaia-pipeline/gaia/workers/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/metadata"
)

// chunkSize is the size of binary chunks transferred to workers.
const chunkSize = 64 * 1024 // 64 KiB

// errNotRegistered is thrown when a worker sends an unauthenticated gRPC request.
var errNotRegistered = errors.New("worker is not registered")

// WorkServer is the implementation of the worker gRPC server interface.
type WorkServer struct{}

// GetWork gets pipeline runs from the store which are not scheduled yet and streams them
// back to the requesting worker. Pipeline runs are filtered by their tags.
func (w *WorkServer) GetWork(workInst *pb.WorkerInstance, serv pb.Worker_GetWorkServer) error {
	// Check if worker is registered
	isRegistered, worker := workerRegistered(serv.Context())
	if !isRegistered {
		md, _ := metadata.FromIncomingContext(serv.Context())
		gaia.Cfg.Logger.Warn("worker tries to get work but is not registered", "metadata", md)
		return errNotRegistered
	}

	// Get memdb instance
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get memdb service via GetWork", "error", err.Error())
		return err
	}
	store, err := services.StorageService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get storage service via GetWork", "error", err.Error())
		return err
	}

	// Update worker information in a separate go routine
	// to avoid getting blocked here.
	worker.LastContact = time.Now()
	worker.Tags = workInst.Tags
	worker.Slots = workInst.WorkerSlots
	go func() {
		if err = db.UpsertWorker(worker, true); err != nil {
			gaia.Cfg.Logger.Error("failed to upsert worker via getwork", "error", err.Error(), "worker", worker)
			return
		}
	}()

	// Get scheduled work from memdb
	for i := int32(0); i < workInst.WorkerSlots; i++ {
		scheduled, err := db.PopPipelineRun(worker.Tags)
		if err != nil {
			return err
		}

		// Check if we have work available
		if scheduled == nil {
			return nil
		}

		// Convert pipeline run to gRPC object
		gRPCPipelineRun := pb.PipelineRun{
			UniqueId:     scheduled.UniqueID,
			Id:           int64(scheduled.ID),
			Status:       string(scheduled.Status),
			PipelineId:   int64(scheduled.PipelineID),
			ScheduleDate: scheduled.ScheduleDate.Unix(),
		}

		// Lookup pipeline from run dependent on the current mode
		switch gaia.Cfg.Mode {
		case gaia.ModeServer:
			for _, p := range pipeline.GlobalActivePipelines.GetAll() {
				if p.ID == scheduled.PipelineID {
					gRPCPipelineRun.ShaSum = p.SHA256Sum
					gRPCPipelineRun.PipelineName = p.Name
					gRPCPipelineRun.PipelineType = string(p.Type)
					break
				}
			}
		case gaia.ModeWorker:
			// Get information directly from the storage
			p, err := store.PipelineGet(scheduled.PipelineID)
			if err != nil {
				gaia.Cfg.Logger.Error("failed to get pipeline via GetWork", "error", err.Error(), "pipeline", scheduled)
				return err
			}
			if p == nil {
				gaia.Cfg.Logger.Error("failed to find related pipeline via GetWork", "error", err.Error(), "pipeline", scheduled)
				return errors.New("failed to find related pipeline in storage")
			}
			shaSum := p.SHA256Sum
			ok, rebuildShaSum, err := store.GetSHAPair(scheduled.PipelineID)
			if err == nil && ok {
				shaSum = rebuildShaSum.Worker
			}

			// Set information
			gRPCPipelineRun.ShaSum = shaSum
			gRPCPipelineRun.PipelineName = p.Name
			gRPCPipelineRun.PipelineType = string(p.Type)
		default:
			gaia.Cfg.Logger.Error("unsupported mode detected via GetWork", "mode", gaia.Cfg.Mode)
			return errors.New("unsupported mode detected")
		}

		// Stream pipeline run back to worker
		if err = serv.Send(&gRPCPipelineRun); err != nil {
			gaia.Cfg.Logger.Error("failed to stream pipeline run to worker instance", "error", err.Error(), "worker", workInst)

			// Insert pipeline run back into memdb since we have popped it
			if errtwo := db.InsertPipelineRun(scheduled); errtwo != nil {
				gaia.Cfg.Logger.Error("failed to insert pipeline run into memdb", "error", errtwo, "originalerr", err)
			}
			return err
		}
	}
	return nil
}

// GetGitRepo retrieves repository information associated with a pipline.
func (w *WorkServer) GetGitRepo(ctx context.Context, in *pb.PipelineID) (*pb.GitRepo, error) {
	repo := &pb.GitRepo{}

	// Check if worker is registered
	isRegistered, _ := workerRegistered(ctx)
	if !isRegistered {
		md, _ := metadata.FromIncomingContext(ctx)
		gaia.Cfg.Logger.Warn("worker tries to get work but is not registered", "metadata", md)
		return repo, errNotRegistered
	}

	store, err := services.StorageService()
	if err != nil {
		return repo, err
	}

	repoInfo, err := store.PipelineGet(int(in.Id))
	if err != nil {
		return repo, err
	}

	if repoInfo == nil {
		return nil, fmt.Errorf("pipeline for id %d not found", int(in.Id))
	}

	pk := pb.PrivateKey{}
	pk.Key = repoInfo.Repo.PrivateKey.Key
	pk.Username = repoInfo.Repo.PrivateKey.Username
	pk.Password = repoInfo.Repo.PrivateKey.Password

	repo.PrivateKey = &pk
	repo.Username = repoInfo.Repo.Username
	repo.Password = repoInfo.Repo.Password
	repo.SelectedBranch = repoInfo.Repo.SelectedBranch
	repo.Url = repoInfo.Repo.URL

	return repo, err
}

// UpdateWork updates work from a worker.
func (w *WorkServer) UpdateWork(ctx context.Context, pipelineRun *pb.PipelineRun) (*empty.Empty, error) {
	e := &empty.Empty{}

	// Check if worker is registered
	isRegistered, worker := workerRegistered(ctx)
	if !isRegistered {
		md, _ := metadata.FromIncomingContext(ctx)
		gaia.Cfg.Logger.Warn("worker tries to update work but is not registered", "metadata", md)
		return e, errNotRegistered
	}

	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get memdb service via updatework", "error", err.Error())
		return e, err
	}

	// Check the status of the pipeline run
	switch gaia.PipelineRunStatus(pipelineRun.Status) {
	case gaia.RunReschedule:
		store, err := services.StorageService()
		if err != nil {
			gaia.Cfg.Logger.Error("failed to get storage service via updatework", "error", err.Error())
			return e, err
		}
		run, err := store.PipelineGetRunByPipelineIDAndID(int(pipelineRun.PipelineId), int(pipelineRun.Id))
		if err != nil {
			gaia.Cfg.Logger.Error("failed to load pipeline run via updatework", "error", err.Error(), "pipelinerun", pipelineRun)
			return e, err
		}
		if run == nil {
			gaia.Cfg.Logger.Error("unable to find pipeline run in store", "pipelinerun", pipelineRun)
			return e, fmt.Errorf("unable to find pipeline run in store: %#v", pipelineRun)
		}

		// Set new status
		run.Status = gaia.RunScheduled
		if err = store.PipelinePutRun(run); err != nil {
			gaia.Cfg.Logger.Error("failed to store pipeline run via updatework", "error", err.Error(), "pipelinerun", run)
			return e, err
		}

		// Put pipeline run back into memdb. This adds the pipeline run to the stack again.
		if err = db.InsertPipelineRun(run); err != nil {
			gaia.Cfg.Logger.Error("failed to insert pipeline run into memdb via updatework", "error", err.Error())
			return e, err
		}

		// Print information output
		gaia.Cfg.Logger.Debug("failed to execute work at worker. Run has been rescheduled...", "runid", run.ID)
	default:
		// Transform protobuf object to internal struct
		run := &gaia.PipelineRun{
			UniqueID:     pipelineRun.UniqueId,
			Status:       gaia.PipelineRunStatus(pipelineRun.Status),
			PipelineID:   int(pipelineRun.PipelineId),
			ID:           int(pipelineRun.Id),
			ScheduleDate: time.Unix(pipelineRun.ScheduleDate, 0),
			StartDate:    time.Unix(pipelineRun.StartDate, 0),
			FinishDate:   time.Unix(pipelineRun.FinishDate, 0),
			Docker:       pipelineRun.Docker,
		}
		run.Jobs = make([]*gaia.Job, 0, len(pipelineRun.Jobs))

		// It can happen that the run is in state "NotScheduled" and waits for the worker
		// scheduler to be picked up. To prevent a rescheduling here at the primary instance,
		// we obfuscate the pipeline run state.
		if run.Status == gaia.RunNotScheduled {
			run.Status = gaia.RunScheduled
		}

		// Transform pipeline run jobs
		jobsMap := make(map[uint32]*gaia.Job)
		for _, job := range pipelineRun.Jobs {
			j := &gaia.Job{
				ID:          job.UniqueId,
				Title:       job.Title,
				Status:      gaia.JobStatus(job.Status),
				Description: job.Description,
			}
			run.Jobs = append(run.Jobs, j)

			// Fill helper map for job dependency search
			jobsMap[j.ID] = j

			// Convert arguments
			j.Args = make([]*gaia.Argument, 0, len(job.Args))
			for _, arg := range job.Args {
				a := &gaia.Argument{
					Description: arg.Description,
					Type:        arg.Type,
					Key:         arg.Key,
					Value:       arg.Value,
				}
				j.Args = append(j.Args, a)
			}
		}

		// Convert dependencies
		for _, job := range pipelineRun.Jobs {
			// Get job
			j := jobsMap[job.UniqueId]

			// Iterate all dependencies
			j.DependsOn = make([]*gaia.Job, 0, len(job.DependsOn))
			for _, depJob := range job.DependsOn {
				// Get dependency
				depJ := jobsMap[depJob.UniqueId]

				// Set dependency
				j.DependsOn = append(j.DependsOn, depJ)
			}
		}

		// Get old pipeline run object first
		store, err := services.StorageService()
		if err != nil {
			gaia.Cfg.Logger.Error("failed to get storage service via updatework", "error", err.Error())
			return e, err
		}
		oldPipelineRun, err := store.PipelineGetRunByID(run.UniqueID)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to get old pipeline run from storage vbia updatework", "error", err.Error())
			return e, err
		}

		// Store pipeline run
		if err = store.PipelinePutRun(run); err != nil {
			gaia.Cfg.Logger.Error("failed to store pipeline run via updatework", "error", err.Error())
			return e, err
		}

		// Update worker information if needed
		switch run.Status {
		case gaia.RunSuccess, gaia.RunFailed, gaia.RunCancelled:
			// Check if this was a docker worker run
			if oldPipelineRun.Docker {
				go func() {
					// Get docker worker object
					dockerWorker, err := db.GetDockerWorker(run.DockerWorkerID)
					if err != nil {
						return
					}

					// Kill and remove docker worker
					_ = db.DeleteDockerWorker(run.DockerWorkerID)
					if err := dockerWorker.KillDockerWorker(); err != nil {
						return
					}
				}()
				return e, nil
			}

			// Update statistics
			worker.FinishedRuns++

			// Store worker object but don't block here
			go func() {
				if err = db.UpsertWorker(worker, true); err != nil {
					gaia.Cfg.Logger.Error("failed to upsert worker via updatework", "error", err.Error(), "worker", worker)
					return
				}
			}()
		}
	}

	return e, nil
}

// StreamBinary streams a pipeline binary in chunks back to the worker.
func (w *WorkServer) StreamBinary(pipelineRun *pb.PipelineRun, serv pb.Worker_StreamBinaryServer) error {
	// Check if worker is registered
	if isRegistered, _ := workerRegistered(serv.Context()); !isRegistered {
		md, _ := metadata.FromIncomingContext(serv.Context())
		gaia.Cfg.Logger.Warn("worker tries to request for binary but is not registered", "metadata", md)
		return errNotRegistered
	}

	// Get storage service
	store, err := services.StorageService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get storage service via StreamBinary", "error", err)
		return err
	}

	// Lookup related pipeline
	var foundPipeline *gaia.Pipeline
	switch gaia.Cfg.Mode {
	case gaia.ModeServer:
		pipelines := pipeline.GlobalActivePipelines.GetAll()
		for id := range pipelines {
			if pipelines[id].ID == int(pipelineRun.PipelineId) {
				foundPipeline = &pipelines[id]
				break
			}
		}
	case gaia.ModeWorker:
		// Load pipeline object from storage
		foundPipeline, err = store.PipelineGet(int(pipelineRun.PipelineId))
		if err != nil {
			gaia.Cfg.Logger.Error("failed to load pipeline object from store via StreamBinary", "error", err)
		}
	}

	// Failed to find the pipeline
	if foundPipeline == nil {
		gaia.Cfg.Logger.Error("failed to stream binary. Failed to find related pipeline", "pipelinerun", pipelineRun)
		return errors.New("failed to find related pipeline with given id")
	}

	// Open pipeline binary file
	file, err := os.Open(foundPipeline.ExecPath)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to open pipeline binary for streambinary", "error", err.Error(), "pipelinerun", pipelineRun)
		return errors.New("failed to open pipeline binary for streaming")
	}
	defer file.Close()

	// Stream back the binary in chunks
	chunk := &pb.FileChunk{}
	buffer := make([]byte, chunkSize)
	for {
		bytesread, err := file.Read(buffer)

		// Check for errors
		if err != nil {
			if err != io.EOF {
				gaia.Cfg.Logger.Error("error occurred during pipeline binary disk read", "error", err.Error(), "pipelinerun", pipelineRun)
				return err
			}
			break
		}

		// Set bytes
		chunk.Chunk = buffer[:bytesread]

		// Stream it back to worker
		if err = serv.Send(chunk); err != nil {
			gaia.Cfg.Logger.Error("failed to stream binary chunk back to worker", "error", err.Error(), "pipelinerun", pipelineRun)
			return err
		}
	}

	return nil
}

// StreamLogs streams logs in chunks from the client to the primary instance.
func (w *WorkServer) StreamLogs(stream pb.Worker_StreamLogsServer) error {
	defer stream.SendAndClose(&empty.Empty{})

	// Check if worker is registered
	if isRegistered, _ := workerRegistered(stream.Context()); !isRegistered {
		md, _ := metadata.FromIncomingContext(stream.Context())
		gaia.Cfg.Logger.Warn("worker tries to stream logs but is not registered", "metadata", md)
		return errNotRegistered
	}

	// Read first chunk which must have content
	firstLogChunk, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil
		}

		gaia.Cfg.Logger.Error("corrupted stream opened via streamlogs", "error", err.Error())
		return err
	}

	// Create logs folder for this run
	logFolderPath := filepath.Join(gaia.Cfg.WorkspacePath, strconv.Itoa(int(firstLogChunk.PipelineId)), strconv.Itoa(int(firstLogChunk.RunId)), gaia.LogsFolderName)
	err = os.MkdirAll(logFolderPath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create pipeline run log folder", "error", err.Error(), "path", logFolderPath)
		return err
	}

	// Open output file
	logFilePath := filepath.Join(logFolderPath, gaia.LogsFileName)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to create new log file via streamlogs", "error", err.Error(), "logobj", firstLogChunk)
		return err
	}
	defer logFile.Close()

	// Write chunk to file
	if _, err := logFile.Write(firstLogChunk.Chunk); err != nil {
		gaia.Cfg.Logger.Error("failed to write chunk to local disk during streamlogs", "error", err.Error(), "logobj", firstLogChunk)
		return err
	}

	// Read whole stream
	for {
		logChunk, err := stream.Recv()

		// Check if stream was closed remotely
		if err == io.EOF {
			break
		}
		if err != nil {
			gaia.Cfg.Logger.Error("failed to stream pipeline run log file from remote instance", "error", err.Error(), "logobj", logChunk)
			return err
		}

		// Defense in depth check. Should never happen!
		if logChunk.RunId != firstLogChunk.RunId || logChunk.PipelineId != firstLogChunk.PipelineId {
			gaia.Cfg.Logger.Error("corrupted chunk found in stream during streamlogs", "logobj", logChunk, "firstlogobj", firstLogChunk)
			return errors.New("corrupted chunk found in stream")
		}

		// Write chunk to file
		if _, err := logFile.Write(logChunk.Chunk); err != nil {
			gaia.Cfg.Logger.Error("failed to write chunk to local disk during streamlogs", "error", err.Error(), "logobj", logChunk)
			return err
		}
	}
	return nil
}

// Deregister removes a worker from this primary instance by deleting the object from store.
func (w *WorkServer) Deregister(ctx context.Context, workInst *pb.WorkerInstance) (*empty.Empty, error) {
	e := &empty.Empty{}

	// Check if worker is registered
	if isRegistered, _ := workerRegistered(ctx); !isRegistered {
		gaia.Cfg.Logger.Warn("worker tries to deregister but is already unregistered", "id", workInst.UniqueId)
		return e, errNotRegistered
	}

	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get memdb service via deregister", "error", err.Error())
		return e, err
	}

	// Delete worker
	if err = db.DeleteWorker(workInst.UniqueId, true); err != nil {
		gaia.Cfg.Logger.Error("failed to delete worker from store via deregister", "error", err.Error(), "worker", workInst)
		return e, err
	}
	return e, nil
}

// workerRegistered checks if a worker by the given context is registered.
// It returns true when the worker is registered and the worker object.
func workerRegistered(ctx context.Context) (bool, *gaia.Worker) {
	var w *gaia.Worker

	// Get metadata information
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		gaia.Cfg.Logger.Debug("failed to get metadata from context")
		return false, w
	}

	// Get identifier
	id := md.Get("uniqueid")
	if len(id) != 1 {
		gaia.Cfg.Logger.Debug("metadata objects contains wrong number of values", "expected", 1, "got", len(id))
		return false, w
	}

	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get memdb service via isWorkerRegistered", "error", err.Error())
		return false, w
	}

	// Lookup worker
	w, err = db.GetWorker(id[0])
	if err != nil {
		gaia.Cfg.Logger.Debug("failed to load worker from memdb via isWorkerRegistered", "error", err.Error(), "id", id)
		return false, w
	}

	// Worker not registered
	if w == nil {
		gaia.Cfg.Logger.Debug("worker is not registered at primary instance but has valid mTLS certificates", "id", id)
		return false, w
	}
	return true, w
}
