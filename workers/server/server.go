package server

import (
	"github.com/gaia-pipeline/gaia"
	pb "github.com/gaia-pipeline/gaia/workers/worker"
	"google.golang.org/grpc"
	"net"
)

// WorkerServer represents an instance of the worker server implementation
type WorkerServer struct {}

// InitWorkerServer creates a new worker server instance.
func InitWorkerServer() *WorkerServer {
	return &WorkerServer{}
}

// Start starts the gRPC worker server.
// It returns an error when something badly happens.
func (w *WorkerServer) Start() {
	lis, err := net.Listen("tcp", ":" + gaia.Cfg.WorkerServerPort)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return
	}

	// Print info message
	gaia.Cfg.Logger.Info("worker gRPC server about to start on port "+gaia.Cfg.WorkerServerPort)

	s := grpc.NewServer()
	pb.RegisterWorkerServer(s, &WorkServer{})
	if err := s.Serve(lis); err != nil {
		gaia.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return
	}
}
