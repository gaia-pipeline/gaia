package server

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	pb "github.com/gaia-pipeline/gaia/workers/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"path/filepath"
)

const (
	hoursBeforeValid = 2
	hoursAfterValid  = 87600 // 10 years
)

// WorkerServer represents an instance of the worker server implementation
type WorkerServer struct{}

// InitWorkerServer creates a new worker server instance.
func InitWorkerServer() *WorkerServer {
	return &WorkerServer{}
}

// Start starts the gRPC worker server.
// It returns an error when something badly happens.
func (w *WorkerServer) Start() {
	lis, err := net.Listen("tcp", ":"+gaia.Cfg.WorkerServerPort)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return
	}

	// Print info message
	gaia.Cfg.Logger.Info("worker gRPC server about to start on port: " + gaia.Cfg.WorkerServerPort)

	// Setup TLS
	certService, err := services.CertificateService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to initiate certificate service", "error", err.Error())
		return
	}

	// Check if certificates exist for the gRPC server
	certPath := filepath.Join(gaia.Cfg.DataPath, "worker_cert.pem")
	keyPath := filepath.Join(gaia.Cfg.DataPath, "worker_key.pem")
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		// Generate certs
		certTmpPath, keyTmpPath, err := certService.CreateSignedCertWithValidOpts(hoursBeforeValid, hoursAfterValid)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to generate cert pair for gRPC server", "error", err.Error())
			return
		}

		// Move certs to correct place
		if err = os.Rename(certTmpPath, certPath); err != nil {
			gaia.Cfg.Logger.Error("failed to move gRPC server cert to data folder", "error", err.Error())
			return
		}
		if err = os.Rename(keyTmpPath, keyPath); err != nil {
			gaia.Cfg.Logger.Error("failed to move gRPC server key to data folder", "error", err.Error())
			return
		}
	}

	// Generate tls config
	tlsConfig, err := certService.GenerateTLSConfig(certPath, keyPath)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to generate tls config for gRPC server", "error", err.Error())
		return
	}

	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterWorkerServer(s, &WorkServer{})
	if err := s.Serve(lis); err != nil {
		gaia.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return
	}
}
