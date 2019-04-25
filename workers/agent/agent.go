package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"google.golang.org/grpc/metadata"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/pipelinehelper"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/agent/api"
	pb "github.com/gaia-pipeline/gaia/workers/proto"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// schedulerTickerSeconds defines the interval in seconds for the scheduler.
	schedulerTickerSeconds = 3

	// updateTickerSeconds defines the interval in seconds to send updates.
	updateTickerSeconds = 2

	// chunkSize is the size of binary chunks transferred to workers.
	chunkSize = 64 * 1024 // 64 KiB

	// idMDKey is the key used for the gRPC metadata map.
	idMDKey = "uniqueid"
)

// Agent represents an instance of an agent
type Agent struct {
	// client represents the interface for the worker client
	client pb.WorkerClient

	// self represents the current agent instance information
	self *pb.WorkerInstance

	// certFile represents the local path to the agent cert
	certFile string

	// keyFile represents the local path to the agent key
	keyFile string

	// caCertFile represents the local path to the agent ca cert
	caCertFile string

	// Instance of scheduler
	scheduler scheduler.GaiaScheduler

	// Instance of store
	store store.GaiaStore
}

// InitAgent initiates the agent instance
func InitAgent(scheduler scheduler.GaiaScheduler, store store.GaiaStore) *Agent {
	ag := &Agent{
		scheduler: scheduler,
		store:     store,
	}

	// Set path to local certificates
	ag.certFile = filepath.Join(gaia.Cfg.HomePath, "cert.pem")
	ag.keyFile = filepath.Join(gaia.Cfg.HomePath, "key.pem")
	ag.caCertFile = filepath.Join(gaia.Cfg.HomePath, "caCert.pem")

	// return instance
	return ag
}

// StartAgent starts the agent main loop and waits until SIGINT or SIGTERM
// signal has been received.
func (a *Agent) StartAgent() error {
	// Allocate SIG channel
	sigs := make(chan os.Signal, 1)

	// Register the signal channel
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Check if this worker has been already registered at a Gaia instance
	workerID := ""
	clientTLS, err := a.generateClientTLSCreds()
	if err != nil {
		// If there is an error, no matter if no certificates exist or
		// we cannot load them, we try the registration process to register
		// the worker again.
		regResp, err := api.RegisterWorker(gaia.Cfg.WorkerHostURL, gaia.Cfg.WorkerSecret)
		if err != nil {
			return fmt.Errorf("failed to register worker: %s", err.Error())
		}

		// The registration process was successful.
		// Store the generated worker id since we need it later.
		if err = a.store.WorkerDeleteAll(); err != nil {
			return fmt.Errorf("failed to clean up worker bucket in store: %s", err.Error())
		}
		w := &gaia.Worker{UniqueID: regResp.UniqueID}
		if err = a.store.WorkerPut(w); err != nil {
			return fmt.Errorf("failed to store worker obj in store: %s", err.Error())
		}
		workerID = regResp.UniqueID

		// Decode received certificates
		cert, err := base64.StdEncoding.DecodeString(regResp.Cert)
		if err != nil {
			return fmt.Errorf("cannot decode certificate: %s", err.Error())
		}
		key, err := base64.StdEncoding.DecodeString(regResp.Key)
		if err != nil {
			return fmt.Errorf("cannot decode key: %s", err.Error())
		}
		caCert, err := base64.StdEncoding.DecodeString(regResp.CACert)
		if err != nil {
			return fmt.Errorf("cannot decode ca cert: %s", err.Error())
		}

		// Store received certificates locally
		if err = ioutil.WriteFile(a.certFile, cert, 0600); err != nil {
			return fmt.Errorf("cannot write cert to disk: %s", err.Error())
		}
		if err = ioutil.WriteFile(a.keyFile, key, 0600); err != nil {
			return fmt.Errorf("cannot write key to disk: %s", err.Error())
		}
		if err = ioutil.WriteFile(a.caCertFile, caCert, 0600); err != nil {
			return fmt.Errorf("cannot write ca cert to disk: %s", err.Error())
		}

		// Update the client TLS object
		clientTLS, err = a.generateClientTLSCreds()
		if err != nil {
			return fmt.Errorf("failed to generate TLS credentials: %s", err.Error())
		}
	}

	dialOption := grpc.WithTransportCredentials(clientTLS)
	conn, err := grpc.Dial(gaia.Cfg.WorkerGRPCHostURL, dialOption)
	if err != nil {
		return fmt.Errorf("failed to connect to remote host: %s", err.Error())
	}
	defer conn.Close()

	// Get worker interface
	a.client = pb.NewWorkerClient(conn)

	// Load worker id from store
	if workerID == "" {
		worker, err := a.store.WorkerGetAll()
		if err != nil {
			return fmt.Errorf("failed to load worker id from store: %s", err.Error())
		}

		// Only one worker obj should exist
		if len(worker) != 1 {
			return fmt.Errorf("failed to load worker obj from store. Expected one object but got %d", len(worker))
		}

		// Set worker id
		workerID = worker[0].UniqueID
	}

	// Setup information object about the current agent
	a.self = &pb.WorkerInstance{
		UniqueId: workerID,
	}

	// Start periodic go routine which schedules the worker work
	workTicker := time.NewTicker(schedulerTickerSeconds * time.Second)
	quitScheduler := make(chan struct{})
	go func() {
		for {
			select {
			case <-workTicker.C:
				// execute schedule function
				a.scheduleWork()
			case <-quitScheduler:
				workTicker.Stop()
				return
			}
		}
	}()

	// Start periodic go routine which sends back information to the Gaia master instance
	updateTicker := time.NewTicker(updateTickerSeconds * time.Second)
	quitUpdate := make(chan struct{})
	go func() {
		for {
			select {
			case <-updateTicker.C:
				// run update function
				a.updateWork()
			case <-quitUpdate:
				updateTicker.Stop()
				return
			}
		}
	}()

	// Block until signal received
	<-sigs
	gaia.Cfg.Logger.Info("exit signal received. Exiting...")

	// Safely stop scheduler
	close(quitScheduler)
	close(quitUpdate)

	return nil
}

// scheduleWork is a periodic go routine which continuously pulls work
// from the Gaia master instance. In case the pipeline is not available
// on this machine, the pipeline will be downloaded from the Gaia master instance.
func (a *Agent) scheduleWork() {
	// Print info output
	gaia.Cfg.Logger.Debug("try to pull work from Gaia master instance...")

	// Check if the agent is busy. Only ask for work when we have the capacity to do it.
	a.self.WorkerSlots = int32(a.scheduler.GetFreeWorkers())
	if a.self.WorkerSlots == 0 {
		gaia.Cfg.Logger.Debug("all workers are currently busy. Will try it again after a while...")
		return
	}

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), (3*schedulerTickerSeconds)*time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, idMDKey, a.self.UniqueId)
	defer cancel()

	// TODO: Add worker tags to a.self
	// Get actual work from remote Gaia instance
	stream, err := a.client.GetWork(ctx, a.self)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to retrieve work from remote instance", "error", err.Error())
		return
	}

	// Read until the stream was closed
	workCounter := 0
	for {
		pipelineRunPB, err := stream.Recv()

		// Stream was closed
		if err == io.EOF {
			break
		}
		if err != nil {
			gaia.Cfg.Logger.Error("failed to stream work from remote instance", "error", err.Error())
			return
		}

		gaia.Cfg.Logger.Info("received work from Gaia master instance...")
		workCounter++

		// Convert protobuf pipeline run to internal struct
		pipelineRun := &gaia.PipelineRun{
			UniqueID:     pipelineRunPB.UniqueId,
			ID:           int(pipelineRunPB.Id),
			Status:       gaia.PipelineRunStatus(pipelineRunPB.Status),
			PipelineID:   int(pipelineRunPB.PipelineId),
			ScheduleDate: time.Unix(pipelineRunPB.ScheduleDate, 0),
		}

		// Convert jobs
		jobsMap := make(map[uint32]*gaia.Job)
		for _, job := range pipelineRunPB.Jobs {
			j := &gaia.Job{
				ID:          job.UniqueId,
				Title:       job.Title,
				Status:      gaia.JobStatus(job.Status),
				Description: job.Description,
			}
			jobsMap[j.ID] = j
			pipelineRun.Jobs = append(pipelineRun.Jobs, j)

			// Arguments
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
		for _, pbJob := range pipelineRunPB.Jobs {
			// Get job
			j := jobsMap[pbJob.UniqueId]

			// Iterate all dependencies
			j.DependsOn = make([]*gaia.Job, 0, len(pbJob.DependsOn))
			for _, depJob := range pbJob.DependsOn {
				// Get dependency
				depJ := jobsMap[depJob.UniqueId]

				// Set dependency
				j.DependsOn = append(j.DependsOn, depJ)
			}
		}

		// Get pipeline binary name and SHA256SUM
		pipelineName := pipelineRunPB.PipelineName
		pipelineSHA256SUM := pipelineRunPB.ShaSum

		// Setup reschedule of pipeline in case something goes wrong
		reschedulePipeline := func() {
			pipelineRunPB.Status = string(gaia.RunReschedule)
			a.client.UpdateWork(context.Background(), pipelineRunPB)
		}

		// Check if the binary is already stored locally
		pipelineFullPath := filepath.Join(gaia.Cfg.PipelinePath, pipelineName)
		if _, err := os.Stat(pipelineFullPath); err != nil {
			// Download binary from remote gaia instance
			if err = a.streamBinary(pipelineRunPB, pipelineFullPath); err != nil {
				gaia.Cfg.Logger.Error("failed to download pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
		}

		// Validate SHA256 sum to make sure the integrity is provided
		sha256Sum, err := getSHA256Sum(pipelineFullPath)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to determine SHA256Sum of pipeline file", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}
		if bytes.Compare(sha256Sum, pipelineSHA256SUM) != 0 {
			// A possible scenario is that the pipeline has been updated and the old binary still exists here.
			// Let us try to delete the binary and re-download the pipeline.
			if err := os.Remove(pipelineFullPath); err != nil {
				gaia.Cfg.Logger.Error("failed to remove inconsistent pipeline binary", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
			if err := a.streamBinary(pipelineRunPB, pipelineFullPath); err != nil {
				gaia.Cfg.Logger.Error("failed to download pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}

			// Validate SHA256 sum again to make sure the integrity is provided
			sha256Sum, err := getSHA256Sum(pipelineFullPath)
			if err != nil {
				gaia.Cfg.Logger.Error("failed to determine SHA256Sum of pipeline file", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
			if bytes.Compare(sha256Sum, pipelineSHA256SUM) != 0 {
				gaia.Cfg.Logger.Error("pipeline binary SHA256Sum mismatch", "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
		}

		// Check if the pipeline has been already stored
		var pipeline *gaia.Pipeline
		pipeline, err = a.store.PipelineGet(pipelineRun.PipelineID)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to load pipeline from store", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}
		if pipeline == nil {
			// Create a new pipeline object
			pipelineType := gaia.PipelineType(pipelineRunPB.PipelineType)
			pipeline = &gaia.Pipeline{
				ID:       pipelineRun.PipelineID,
				Name:     pipelinehelper.GetRealPipelineName(pipelineRunPB.PipelineName, pipelineType),
				Type:     pipelineType,
				ExecPath: pipelineFullPath,
			}
		}

		// Doesn't matter if we created a new pipeline object or load it from store,
		// we always set the correct SHA256Sum to make sure this is always the newest.
		pipeline.SHA256Sum = pipelineSHA256SUM

		// Let us try to start the plugin and receive all implemented jobs
		if err = a.scheduler.SetPipelineJobs(pipeline); err != nil {
			gaia.Cfg.Logger.Error("cannot get pipeline jobs", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}
		pipelineRun.Jobs = pipeline.Jobs

		// Store pipeline
		if err = a.store.PipelinePut(pipeline); err != nil {
			gaia.Cfg.Logger.Error("failed to store pipeline in store", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}

		// The scheduler picks only runs up which are in state "NotScheduled".
		// Since the scheduler from the Gaia master instance set the state already to "scheduled",
		// we have to reset the state here so that the scheduler will pick it up.
		pipelineRun.Status = gaia.RunNotScheduled

		// Store finally pipeline run
		if err = a.store.PipelinePutRun(pipelineRun); err != nil {
			gaia.Cfg.Logger.Error("failed to store pipeline run in store", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}
	}

	// Check if we received work at all
	if workCounter == 0 {
		gaia.Cfg.Logger.Debug("got no work from Gaia master instance. Will try it again after a while...")
	}
}

// streamBinary streams the binary in chunks from the remote instance to the given path.
func (a *Agent) streamBinary(pipelineRunPB *pb.PipelineRun, pipelinePath string) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, idMDKey, a.self.UniqueId)
	defer cancel()

	// Initiate streaming
	stream, err := a.client.StreamBinary(ctx, pipelineRunPB)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to stream pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
		return err
	}

	// Open output file
	pipelineFile, err := os.Create(pipelinePath)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to create new pipeline binary file during remote binary stream", "error", err.Error(), "pipelinerun", pipelineRunPB)
		return err
	}
	defer pipelineFile.Close()

	// Read whole stream
	for {
		streamChunk, err := stream.Recv()

		// Check if stream was closed remotely
		if err == io.EOF {
			break
		}
		if err != nil {
			gaia.Cfg.Logger.Error("failed to stream full pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
			return err
		}

		// Write chunk to file
		if _, err := pipelineFile.Write(streamChunk.Chunk); err != nil {
			gaia.Cfg.Logger.Error("failed to write chunk to local disk during stream binary", "error", err.Error(), "pipelinerun", pipelineRunPB)
			return err
		}
	}

	// Set pipeline executable rights
	return os.Chmod(pipelinePath, 0766)
}

// updateWork is function that is periodically called and it is used to
// send new information about a pipeline run to the Gaia master instance.
func (a *Agent) updateWork() {
	// Read all pipeline runs from the store. The number of pipeline runs
	// should be relatively low since we delete pipeline runs after successful
	// execution.
	runs, err := a.store.PipelineGetAllRuns()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to load pipeline runs from store", "error", err.Error())
		return
	}

	// Send all pipeline runs to the remote primary instance
	for _, run := range runs {
		// Transform to protobuf struct
		runPB := &pb.PipelineRun{
			Id:           int64(run.ID),
			UniqueId:     run.UniqueID,
			Status:       string(run.Status),
			PipelineId:   int64(run.PipelineID),
			ScheduleDate: run.ScheduleDate.Unix(),
			StartDate:    run.StartDate.Unix(),
			FinishDate:   run.FinishDate.Unix(),
		}

		// Transform pipeline run jobs
		jobsMap := make(map[uint32]*pb.Job)
		for _, job := range run.Jobs {
			j := &pb.Job{
				UniqueId:    job.ID,
				Title:       job.Title,
				Status:      string(job.Status),
				Description: job.Description,
			}
			runPB.Jobs = append(runPB.Jobs, j)

			// Fill helper map for job dependency search
			jobsMap[j.UniqueId] = j

			// Convert arguments
			j.Args = make([]*pb.Argument, 0, len(job.Args))
			for _, arg := range job.Args {
				a := &pb.Argument{
					Description: arg.Description,
					Type:        arg.Type,
					Key:         arg.Key,
					Value:       arg.Value,
				}
				j.Args = append(j.Args, a)
			}
		}

		// Convert dependencies
		for _, job := range run.Jobs {
			// Get job
			j := jobsMap[job.ID]

			// Iterate all dependencies
			j.DependsOn = make([]*pb.Job, 0, len(job.DependsOn))
			for _, depJob := range job.DependsOn {
				// Get dependency
				depJ := jobsMap[depJob.ID]

				// Set dependency
				j.DependsOn = append(j.DependsOn, depJ)
			}
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), (updateTickerSeconds*3)*time.Second)
		ctx = metadata.AppendToOutgoingContext(ctx, idMDKey, a.self.UniqueId)
		defer cancel()

		// Send to remote instance
		if _, err := a.client.UpdateWork(ctx, runPB); err != nil {
			gaia.Cfg.Logger.Error("failed to send update information to remote instance", "error", err.Error())
			return
		}

		// Remove pipeline run from store when the state is finalized
		if run.Status == gaia.RunFailed || run.Status == gaia.RunSuccess || run.Status == gaia.RunCancelled {
			if err = a.store.PipelineRunDelete(run.UniqueID); err != nil {
				gaia.Cfg.Logger.Error("failed to remove pipeline run from store", "error", err.Error(), "pipelinerun", run)
			}
		}

		// Check if log file exists for pipeline run
		logFilePath := filepath.Join(gaia.Cfg.WorkspacePath, strconv.Itoa(run.PipelineID), strconv.Itoa(run.ID), gaia.LogsFolderName, gaia.LogsFileName)
		if _, err := os.Stat(logFilePath); err != nil {
			continue
		}

		// Open file handle
		file, err := os.Open(logFilePath)
		if err != nil {
			gaia.Cfg.Logger.Warn("failed to open pipeline run log file via updatework", "error", err.Error(), "pipelinerun", run)
			continue
		}
		defer file.Close()

		// Open streaming session to primary instance
		streamCtx := context.Background()
		streamCtx = metadata.AppendToOutgoingContext(streamCtx, idMDKey, a.self.UniqueId)
		stream, err := a.client.StreamLogs(streamCtx)
		if err != nil {
			gaia.Cfg.Logger.Warn("failed to open stream session to primary instance to ship logs via updatework", "error", err.Error(), "pipelinerun", run)
			continue
		}

		chunk := &pb.LogChunk{
			PipelineId: int64(run.PipelineID),
			RunId:      int64(run.ID),
		}
		buffer := make([]byte, chunkSize)
		for {
			bytesread, err := file.Read(buffer)

			// Check for errors
			if err != nil {
				if err != io.EOF {
					gaia.Cfg.Logger.Warn("error occurred during pipeline run log disk read", "error", err.Error(), "pipelinerun", run)
					continue
				}
				break
			}

			// Set bytes
			chunk.Chunk = buffer[:bytesread]

			// Stream it to primary instance
			if err = stream.Send(chunk); err != nil {
				gaia.Cfg.Logger.Error("failed to stream log chunk to primary instance", "error", err.Error(), "pipelinerun", run)
				continue
			}
		}
		if err = stream.CloseSend(); err != nil {
			gaia.Cfg.Logger.Warn("failed to safely close gRPC connection via updatework", "error", err.Error())
		}
	}
}

// generateClientTLSCreds checks if certificates exist in the home directory.
// It will load the certificates and generates TLS creds for mTLS connection.
func (a *Agent) generateClientTLSCreds() (credentials.TransportCredentials, error) {
	// Check if all certs exist
	if _, err := os.Stat(a.certFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(a.keyFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(a.caCertFile); os.IsNotExist(err) {
		return nil, err
	}

	// Load client key pair
	certs, err := tls.LoadX509KeyPair(a.certFile, a.keyFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(a.caCertFile)
	if err != nil {
		return nil, err
	}

	// Add certificate to cert pool
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errors.New("cannot append ca cert to cert pool")
	}

	return credentials.NewTLS(&tls.Config{
		ServerName:   gaia.Cfg.WorkerName,
		Certificates: []tls.Certificate{certs},
		RootCAs:      certPool,
	}), nil
}

// --- TODO: move to helper package since this is also used in ticker.go ---

// getSHA256Sum accepts a path to a file.
// It load's the file and calculates a SHA256 Checksum and returns it.
func getSHA256Sum(path string) ([]byte, error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create sha256 obj and insert bytes
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	// return sha256 checksum
	return h.Sum(nil), nil
}
