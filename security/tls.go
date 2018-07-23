package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/gaia-pipeline/gaia"
)

const (
	rsaBits      = 2048
	maxValidCA   = 17520 // 2 years
	maxValidCERT = 48    // 48 hours
	orgName      = "gaia-pipeline"
	orgDNS       = "gaia-pipeline.io"

	// CA key name
	certName = "ca.crt"
	keyName  = "ca.key"
)

// GenerateCA generates the CA and puts it into the data folder.
// The CA will be always overwritten on startup.
func GenerateCA() error {
	// Cleanup old certs if existing.
	// We ignore the error here cause files might be non existend.
	caCertPath := filepath.Join(gaia.Cfg.DataPath, certName)
	caKeyPath := filepath.Join(gaia.Cfg.DataPath, keyName)
	cleanupCerts(caCertPath, caKeyPath)

	// Generate the key
	key, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return err
	}

	// Set time range for cert validation
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * maxValidCA)

	// Generate serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	// Generate CA template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{orgName},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{orgDNS},
	}

	// Create certificate authority
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, key.PublicKey, key)
	if err != nil {
		return err
	}

	// Write out the ca.crt file
	certOut, err := os.Create(caCertPath)
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// Write out the ca.key file
	keyOut, err := os.OpenFile(caKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	keyOut.Close()

	return nil
}

// createSignedCert creates a new key pair which is signed by the CA.
func createSignedCert() (string, string, error) {
	caCertPath := filepath.Join(gaia.Cfg.DataPath, "ca.crt")
	caKeyPath := filepath.Join(gaia.Cfg.DataPath, "ca.key")

	// Load CA plain
	caPlain, err := tls.LoadX509KeyPair(caCertPath, caKeyPath)
	if err != nil {
		return "", "", err
	}

	// Parse certificate
	ca, err := x509.ParseCertificate(caPlain.Certificate[0])
	if err != nil {
		return "", "", err
	}

	// Generate serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", err
	}

	// Set time range for cert validation
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * maxValidCERT)

	// Prepare certificate
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{orgName},
		},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	priv, _ := rsa.GenerateKey(rand.Reader, rsaBits)
	pub := &priv.PublicKey

	// Sign the certificate
	certSigned, err := x509.CreateCertificate(rand.Reader, cert, ca, pub, caPlain.PrivateKey)

	// Public key
	certOut, err := ioutil.TempFile("", "crt")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certSigned})
	certOut.Close()

	// Private key
	keyOut, err := ioutil.TempFile("", "key")
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return certOut.Name(), keyOut.Name(), nil
}

// cleanupCerts removes certificates at the given path.
func cleanupCerts(crt, key string) error {
	if err := os.Remove(crt); err != nil {
		return err
	}
	return os.Remove(key)
}
