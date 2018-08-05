package security

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

func TestInitCA(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	tmp, _ := ioutil.TempDir("", "TestInitCA")
	gaia.Cfg.DataPath = tmp

	c, err := InitCA()
	if err != nil {
		t.Fatal(err)
	}

	// Get root CA cert path
	caCertPath, caKeyPath := c.GetCACertPath()

	// Load CA plain
	caPlain, err := tls.LoadX509KeyPair(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse certificate
	ca, err := x509.ParseCertificate(caPlain.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	// Create cert pool and load ca root
	certPool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		t.Fatal(err)
	}

	ok := certPool.AppendCertsFromPEM(rootCA)
	if !ok {
		t.Fatalf("Cannot append root cert to cert pool!\n")
	}

	_, err = ca.Verify(x509.VerifyOptions{
		Roots:   certPool,
		DNSName: orgDNS,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = c.CleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateSignedCert(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.DataPath = os.TempDir()

	c, err := InitCA()
	if err != nil {
		t.Fatal(err)
	}

	// Get root ca cert path
	caCertPath, caKeyPath := c.GetCACertPath()

	certPath, keyPath, err := c.CreateSignedCert()
	if err != nil {
		t.Fatal(err)
	}

	// Load CA plain
	caPlain, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse certificate
	ca, err := x509.ParseCertificate(caPlain.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	// Create cert pool and load ca root
	certPool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		t.Fatal(err)
	}

	ok := certPool.AppendCertsFromPEM(rootCA)
	if !ok {
		t.Fatalf("Cannot append root cert to cert pool!\n")
	}

	_, err = ca.Verify(x509.VerifyOptions{
		Roots:   certPool,
		DNSName: orgDNS,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = c.CleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = c.CleanupCerts(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateTLSConfig(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.DataPath = os.TempDir()

	c, err := InitCA()
	if err != nil {
		t.Fatal(err)
	}

	// Get root ca cert path
	caCertPath, caKeyPath := c.GetCACertPath()

	certPath, keyPath, err := c.CreateSignedCert()
	if err != nil {
		t.Fatal(err)
	}

	// Generate TLS Config
	_, err = c.GenerateTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}

	err = c.CleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = c.CleanupCerts(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
}
