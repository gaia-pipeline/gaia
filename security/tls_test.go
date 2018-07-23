package security

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

func TestGenerateCA(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.DataPath = os.TempDir()

	err := GenerateCA()
	if err != nil {
		t.Fatal(err)
	}

	caCertPath := filepath.Join(gaia.Cfg.DataPath, "ca.crt")
	caKeyPath := filepath.Join(gaia.Cfg.DataPath, "ca.key")

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

	err = cleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateSignedCert(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.DataPath = os.TempDir()

	err := GenerateCA()
	if err != nil {
		t.Fatal(err)
	}

	caCertPath := filepath.Join(gaia.Cfg.DataPath, "ca.crt")
	caKeyPath := filepath.Join(gaia.Cfg.DataPath, "ca.key")

	certPath, keyPath, err := createSignedCert()
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

	err = cleanupCerts(caCertPath, caKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = cleanupCerts(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
}
