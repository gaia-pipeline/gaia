package server

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/gaia-pipeline/gaia/helper/filehelper"
	"github.com/gaia-pipeline/gaia/security"

	"github.com/gaia-pipeline/gaia"
	pb "github.com/gaia-pipeline/gaia/workers/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	hoursBeforeValid = 2
	hoursAfterValid  = 87600 // 10 years
)

// Dependencies defines dependencies of this service.
type Dependencies struct {
	Certificate security.CAAPI
}

// WorkerServer represents an instance of the worker server implementation
type WorkerServer struct {
	Dependencies
}

// InitWorkerServer creates a new worker server instance.
func InitWorkerServer(deps Dependencies) *WorkerServer {
	return &WorkerServer{
		Dependencies: deps,
	}
}

// Start starts the gRPC worker server.
// It returns an error when something badly happens.
func (w *WorkerServer) Start() error {
	lis, err := net.Listen("tcp", ":"+gaia.Cfg.WorkerServerPort)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return err
	}

	// Print info message
	gaia.Cfg.Logger.Info("worker gRPC server about to start on port: " + gaia.Cfg.WorkerServerPort)

	// Check if certificates exist for the gRPC server
	certPath := filepath.Join(gaia.Cfg.DataPath, "worker_cert.pem")
	keyPath := filepath.Join(gaia.Cfg.DataPath, "worker_key.pem")
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		// Parse hostname for the certificate
		s := strings.Split(gaia.Cfg.WorkerGRPCHostURL, ":")
		if len(s) != 2 {
			gaia.Cfg.Logger.Error("failed to parse configured gRPC worker host url", "url", gaia.Cfg.WorkerGRPCHostURL)
			return fmt.Errorf("failed to parse configured gRPC worker host url: %s", gaia.Cfg.WorkerGRPCHostURL)
		}

		// Generate certs
		certTmpPath, keyTmpPath, err := w.Certificate.CreateSignedCertWithValidOpts(s[0], hoursBeforeValid, hoursAfterValid)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to generate cert pair for gRPC server", "error", err.Error())
			return err
		}

		// Move certs to correct place
		if err = filehelper.CopyFileContents(certTmpPath, certPath); err != nil {
			gaia.Cfg.Logger.Error("failed to copy gRPC server cert to data folder", "error", err.Error())
			return err
		}
		if err = filehelper.CopyFileContents(keyTmpPath, keyPath); err != nil {
			gaia.Cfg.Logger.Error("failed to copy gRPC server key to data folder", "error", err.Error())
			return err
		}
		if err = os.Remove(certTmpPath); err != nil {
			gaia.Cfg.Logger.Error("failed to remove temporary server cert file", "error", err)
			return err
		}
		if err = os.Remove(keyTmpPath); err != nil {
			gaia.Cfg.Logger.Error("failed to remove temporary key cert file", "error", err)
			return err
		}
	}

	// Generate tls config
	tlsConfig, err := w.Certificate.GenerateTLSConfig(certPath, keyPath)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to generate tls config for gRPC server", "error", err.Error())
		return err
	}

	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterWorkerServer(s, &WorkServer{})
	if err := s.Serve(lis); err != nil {
		gaia.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return err
	}
	return nil
}
