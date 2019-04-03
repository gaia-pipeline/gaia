package server

import (
	"context"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	pb "github.com/gaia-pipeline/gaia/workers/worker"
)

// WorkServer is the implementation of the worker gRPC server interface.
type WorkServer struct{}

func (w *WorkServer) Ping(ctx context.Context, workInst *pb.WorkerInstance) (*pb.Empty, error) {
	// TODO: Store ping received.
	return &pb.Empty{}, nil
}

// GetWork gets pipeline runs from the store which are not scheduled yet and streams them
// back to the requesting worker. Pipeline runs are filtered by their tags.
func (w *WorkServer) GetWork(workInst *pb.WorkerInstance, serv pb.Worker_GetWorkServer) error {
	// Get memdb instance
	db, err := services.MemDBService(nil)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get memdb service via GetWork", "error", err.Error())
		return err
	}

	// Get scheduled work from memdb
	for i := uint32(0); i < workInst.WorkerSlots; i++ {
		// TODO: Pop pipeline runs by their tags & worker tags
		scheduled, err := db.PopPipelineRun()
		if err != nil {
			return err
		}

		// Convert pipeline run to gRPC object
		gRPCPipelineRun := pb.PipelineRun{
			UniqueId: scheduled.UniqueID,
			Id: uint32(scheduled.ID),
			Status: string(scheduled.Status),
		}

		// Stream pipeline run back to worker
		if err = serv.Send(&gRPCPipelineRun); err != nil {
			gaia.Cfg.Logger.Error("failed to stream pipeline run to worker instance", "error", err.Error(), "worker", workInst)

			// Insert pipeline run back into memdb since we have popped it
			db.InsertPipelineRun(scheduled)
		}
	}
	return nil
}

func (w *WorkServer) Deregister(ctx context.Context, workInst *pb.WorkerInstance) (*pb.Empty, error) {
	// TODO: Remove worker from store
	return &pb.Empty{}, nil
}
