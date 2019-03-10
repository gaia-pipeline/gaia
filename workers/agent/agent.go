package agent

import (
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

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/workers/agent/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Define the local certificates path
var certFile, keyFile, caCertFile string

func init() {
	// Set path to local certificates
	certFile = filepath.Join(gaia.Cfg.HomePath, "cert.pem")
	keyFile = filepath.Join(gaia.Cfg.HomePath, "key.pem")
	caCertFile = filepath.Join(gaia.Cfg.HomePath, "caCert.pem")

}

// StartAgent starts the agent main loop and waits until SIGINT or SIGTERM
// signal has been received.
func StartAgent() error {
	// Allocate SIG channel
	sigs := make(chan os.Signal, 1)

	// Register the signal channel
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Check if this worker has been already registered at a Gaia instance
	clientTLS, err := generateClientTLSCreds()
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
		if err = ioutil.WriteFile(certFile, cert, 0600); err != nil {
			return fmt.Errorf("cannot write cert to disk: %s", err.Error())
		}
		if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
			return fmt.Errorf("cannot write key to disk: %s", err.Error())
		}
		if err = ioutil.WriteFile(caCertFile, caCert, 0600); err != nil {
			return fmt.Errorf("cannot write ca cert to disk: %s", err.Error())
		}

		// Update the client TLS object
		clientTLS, err = generateClientTLSCreds()
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

	// Block until signal received
	<-sigs
	gaia.Cfg.Logger.Info("exit signal received. Exiting...")
	return nil
}

// generateClientTLSCreds checks if certificates exist in the home directory.
// It will load the certificates and generates TLS creds for mTLS connection.
func generateClientTLSCreds() (credentials.TransportCredentials, error) {
	// Check if all certs exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
		return nil, err
	}

	// Load client key pair
	certs, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(caCertFile)
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
