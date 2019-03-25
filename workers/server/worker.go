package server

import (
	"context"
	pb "github.com/gaia-pipeline/gaia/workers/worker"
)

// WorkServer is the implementation of the worker gRPC server interface.
type WorkServer struct{}

func (w *WorkServer) Ping(ctx context.Context, workInst *pb.WorkerInstance) (*pb.Empty, error) {
	// TODO: Store ping received.
	return &pb.Empty{}, nil
}

func (w *WorkServer) GetWork(workInst *pb.WorkerInstance, serv pb.Worker_GetWorkServer) error {
	// TODO: Get work from the store which was scheduled for this worker instance
	// TODO: Filter by worker instance
	return nil
}

func (w *WorkServer) Deregister(ctx context.Context, workInst *pb.WorkerInstance) (*pb.Empty, error) {
	// TODO: Remove worker from store
	return &pb.Empty{}, nil
}
