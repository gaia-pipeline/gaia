package agent

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/workers/agent/api"
	pb "github.com/gaia-pipeline/gaia/workers/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// schedulerTickerSeconds defines the interval in seconds for the scheduler.
const schedulerTickerSeconds = 3

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
}

// InitAgent initiates the agent instance
func InitAgent(scheduler scheduler.GaiaScheduler) *Agent {
	ag := &Agent{
		scheduler: scheduler,
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

	// Get store interface
	store, err := services.StorageService()
	if err != nil {
		return fmt.Errorf("cannot access local store: %s", err.Error())
	}

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
		if err = store.WorkerDeleteAll(); err != nil {
			return fmt.Errorf("failed to clean up worker bucket in store: %s", err.Error())
		}
		w := &gaia.Worker{UniqueID: regResp.UniqueID}
		if err = store.WorkerPut(w); err != nil {
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
	conn, err := grpc.Dial(gaia.Cfg.WorkerHostURL, dialOption)
	if err != nil {
		return fmt.Errorf("failed to connect to remote host: %s", err.Error())
	}
	defer conn.Close()

	// Get worker interface
	a.client = pb.NewWorkerClient(conn)

	// Load worker id from store
	if workerID == "" {
		worker, err := store.WorkerGetAll()
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
	ticker := time.NewTicker(schedulerTickerSeconds * time.Second)
	quitScheduler := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// execute schedule function
				a.scheduleWork()
			case <-quitScheduler:
				ticker.Stop()
				return
			}
		}
	}()

	// Block until signal received
	<-sigs
	gaia.Cfg.Logger.Info("exit signal received. Exiting...")

	// Safely stop scheduler
	close(quitScheduler)

	return nil
}

// scheduleWork is a periodic go routine which continuously pulls work
// from the Gaia master instance. In case the pipeline is not available
// on this machine, the pipeline will be downloaded from the Gaia master instance.
func (a *Agent) scheduleWork() {
	// Check if the agent is busy. Only ask for work when we have the capacity to do it.
	a.self.WorkerSlots = int32(a.scheduler.GetFreeWorkers())
	if a.self.WorkerSlots == 0 {
		return
	}

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), (3 * schedulerTickerSeconds) * time.Second)
	defer cancel()

	// TODO: Add worker tags to a.self
	// Get actual work from remote Gaia instance
	stream, err := a.client.GetWork(ctx, a.self)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to retrieve work from remote instance", "error", err.Error())
		return
	}

	// Read until the stream was closed
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

		// Convert protobuf pipeline run to internal struct
		pipelineRun := &gaia.PipelineRun{
			UniqueID: pipelineRunPB.UniqueId,
			ID: int(pipelineRunPB.Id),
			Status: gaia.PipelineRunStatus(pipelineRunPB.Status),
			PipelineID: int(pipelineRunPB.PipelineId),
		}

		// TODO: Check if we have to download the pipeline binary
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
