package handlers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

const (
	hoursBeforeValid = 2
	hoursAfterValid  = 87600 // 10 years
)

type registerWorker struct {
	Secret string   `json:"secret"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
}

type registerResponse struct {
	Cert   string `json:"cert"`
	Key    string `json:"key"`
	CACert string `json:"cacert"`
}

// RegisterWorker allows new workers to register themself at this Gaia instance.
// It accepts a secret and returns valid certificates (base64 encoded) for further mTLS connection.
func RegisterWorker(c echo.Context) error {
	worker := registerWorker{}
	if err := c.Bind(&worker); err != nil {
		return c.String(http.StatusBadRequest, "secret for registration is invalid:"+err.Error())
	}

	// Lookup the global registration secret in our vault
	globalSecret, err := getWorkerSecret()
	if err != nil {
		return c.String(http.StatusInternalServerError, "cannot get worker secret from vault")
	}

	// Check if given secret is equal with global worker secret
	if globalSecret != worker.Secret {
		return c.String(http.StatusForbidden, "wrong global worker secret provided")
	}

	w := gaia.Worker{
		UniqueID:     uuid.Must(uuid.NewV4(), nil).String(),
		Name:         worker.Name,
		Tags:         worker.Tags,
		RegisterDate: time.Now(),
		LastContact:  time.Now(),
		Status:       gaia.WorkerActive,
	}

	// Generate certificates for worker
	cert, err := services.CertificateService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get certificate service", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot get certificate service")
	}
	crt, key, err := cert.CreateSignedCertWithValidOpts(hoursBeforeValid, hoursAfterValid)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create signed certificate", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot create signed certificate")
	}

	// Get public cert from CA (required for mTLS)
	caCertPath, _ := cert.GetCACertPath()
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot load CA cert", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot load CA cert")
	}

	// Encode all certificates base64 to prevent inconsistency during transport
	crtB64 := base64.StdEncoding.EncodeToString([]byte(crt))
	keyB64 := base64.StdEncoding.EncodeToString([]byte(key))
	caCertB64 := base64.StdEncoding.EncodeToString([]byte(caCert))

	// Register worker by adding it to store
	store, err := services.StorageService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get store service via register worker", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot get store service")
	}
	if err = store.WorkerPut(&w); err != nil {
		gaia.Cfg.Logger.Error("cannot store worker object", "error", err.Error(), "worker", w)
		return c.String(http.StatusInternalServerError, "cannot store worker object")
	}

	return c.JSON(http.StatusOK, registerResponse{
		Cert:   crtB64,
		Key:    keyB64,
		CACert: caCertB64,
	})
}

// GetWorkerRegisterSecret returns the global secret for registering new worker.
func GetWorkerRegisterSecret(c echo.Context) error {
	globalSecret, err := getWorkerSecret()
	if err != nil {
		return c.String(http.StatusInternalServerError, "cannot get worker secret from vault")
	}
	return c.String(http.StatusOK, globalSecret)
}

// getWorkerSecret returns the global secret for registering new worker.
func getWorkerSecret() (string, error) {
	v, err := services.VaultService(nil)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get vault instance", "error", err.Error())
		return "", err
	}
	secret, err := v.Get(gaia.WorkerRegisterKey)
	if err != nil {
		gaia.Cfg.Logger.Debug("global worker secret not found", "error", err.Error())
		return "", err
	}
	return string(secret[:]), nil
}
