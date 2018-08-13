package golang

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
)

// generateTLSConfig generates a new TLS config based on given
// certificate path and key path.
func generateTLSConfig(certPath, keyPath, caCertPath string) (*tls.Config, error) {
	// Load certificate
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	// Create certificate pool
	certPool := x509.NewCertPool()
	rootCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	// Append cert to cert pool
	ok := certPool.AppendCertsFromPEM(rootCert)
	if !ok {
		return nil, errCertNotAppended
	}

	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}, nil
}

// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

func (l *rmListener) Close() error {
	// Close the listener itself
	if err := l.Listener.Close(); err != nil {
		return err
	}

	// Remove the file
	return os.Remove(l.Path)
}
