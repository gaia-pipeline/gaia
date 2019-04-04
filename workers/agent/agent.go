package agent

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
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
}

// InitAgent initiates the agent instance
func InitAgent() *Agent {
	ag := &Agent{}

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
	clientTLS, err := a.generateClientTLSCreds()
	if err != nil {
		// If there is an error, no matter if no certificates exist or
		// we cannot load them, we try the registration process to register
		// the worker again.
		regResp, err := api.RegisterWorker(gaia.Cfg.WorkerHostURL, gaia.Cfg.WorkerSecret)
		if err != nil {
			return err
		}

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
			return err
		}
	}

	dialOption := grpc.WithTransportCredentials(clientTLS)
	conn, err := grpc.Dial(gaia.Cfg.WorkerHostURL, dialOption)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot establish connection to Gaia instance", "error", err)
		return err
	}
	defer conn.Close()

	// Get worker interface
	a.client = pb.NewWorkerClient(conn)

	// Setup information object about the current agent
	a.self = &pb.WorkerInstance{
		UniqueId:
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
// on this machine, the pipeline will be downloaded from the Gaia instance.
func (a *Agent) scheduleWork() {
	// Get actual work from remote Gaia instance
	stream, err := a.client.GetWork(context.Background(), )

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
