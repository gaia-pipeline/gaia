package workers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
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
	UniqueID string `json:"uniqueid"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"cacert"`
}

// RegisterWorker allows new workers to register themself at this Gaia instance.
// It accepts a secret and returns valid certificates (base64 encoded) for further mTLS connection.
// @Summary Register a new worker.
// @Description Allows new workers to register themself at this Gaia instance.
// @Tags workers
// @Accept json
// @Produce json
// @Param RegisterWorkerRequest body registerWorker true "Worker details"
// @Success 200 {object} registerResponse "Details of the registered worker."
// @Failure 400 {string} string "Invalid arguments of the worker."
// @Failure 403 {string} string "Wrong global worker secret provided."
// @Failure 500 {string} string "Various internal services like, certs, vault and generating new secrets."
// @Router /worker/register [post]
func (wp *WorkerProvider) RegisterWorker(c echo.Context) error {
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

	// Generate name if none was given
	if worker.Name == "" {
		worker.Name = randomdata.SillyName() + "_" + randomdata.SillyName()
	}

	v4, err := uuid.NewV4()
	if err != nil {
		return c.String(http.StatusInternalServerError, "error generating uuid")
	}
	w := gaia.Worker{
		UniqueID:     uuid.Must(v4, nil).String(),
		Name:         worker.Name,
		Tags:         worker.Tags,
		RegisterDate: time.Now(),
		LastContact:  time.Now(),
		Status:       gaia.WorkerActive,
	}

	// Generate certificates for worker
	crtPath, keyPath, err := wp.deps.Certificate.CreateSignedCertWithValidOpts("", hoursBeforeValid, hoursAfterValid)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create signed certificate", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot create signed certificate")
	}
	defer func() {
		if err := wp.deps.Certificate.CleanupCerts(crtPath, keyPath); err != nil {
			gaia.Cfg.Logger.Error("failed to remove worker certificates", "error", err)
		}
	}()

	// Get public cert from CA (required for mTLS)
	caCertPath, _ := wp.deps.Certificate.GetCACertPath()
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot load CA cert", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot load CA cert")
	}

	// Load certs from disk
	crt, err := ioutil.ReadFile(crtPath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot load cert", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot load cert")
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot load key", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot load key")
	}

	// Encode all certificates base64 to prevent character issues during transportation
	crtB64 := base64.StdEncoding.EncodeToString(crt)
	keyB64 := base64.StdEncoding.EncodeToString(key)
	caCertB64 := base64.StdEncoding.EncodeToString(caCert)

	// Register worker by adding it to the memdb and store
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get memdb service via register worker", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot get memdb service")
	}
	if err = db.UpsertWorker(&w, true); err != nil {
		return c.String(http.StatusInternalServerError, "failed to store worker in memdb/store")
	}

	return c.JSON(http.StatusOK, registerResponse{
		UniqueID: w.UniqueID,
		Cert:     crtB64,
		Key:      keyB64,
		CACert:   caCertB64,
	})
}

// DeregisterWorker deregister a registered worker.
// @Summary Deregister and existing worker.
// @Description Deregister an existing worker.
// @Tags workers
// @Accept json
// @Produce json
// @Param workerid query string true "The id of the worker to deregister."
// @Success 200 {string} string "Worker has been successfully deregistered."
// @Failure 400 {string} string "Worker id is missing or worker not registered."
// @Failure 500 {string} string "Cannot get memdb service from service store or failed to delete worker."
// @Router /worker/{workerid} [delete]
func (wp *WorkerProvider) DeregisterWorker(c echo.Context) error {
	workerID := c.Param("workerid")
	if workerID == "" {
		return c.String(http.StatusBadRequest, "worker id is missing")
	}

	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get memdb service from store", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot get memdb service from service store")
	}

	// Check if worker is still registered
	w, err := db.GetWorker(workerID)
	if err != nil || w == nil {
		return c.String(http.StatusBadRequest, "worker is not registered")
	}

	// Delete worker which basically indicates it is not registered anymore
	if err := db.DeleteWorker(w.UniqueID, true); err != nil {
		gaia.Cfg.Logger.Error("failed to delete worker", "error", err.Error())
		return c.String(http.StatusInternalServerError, "failed to delete worker")
	}

	return c.String(http.StatusOK, "worker has been successfully deregistered")
}

// GetWorkerRegisterSecret returns the global secret for registering new worker.
// @Summary Get worker register secret.
// @Description Returns the global secret for registering new worker.
// @Tags workers
// @Produce json
// @Success 200 {string} string
// @Failure 500 {string} string "Cannot get worker secret from vault."
// @Router /worker/secret [get]
func (wp *WorkerProvider) GetWorkerRegisterSecret(c echo.Context) error {
	globalSecret, err := getWorkerSecret()
	if err != nil {
		return c.String(http.StatusInternalServerError, "cannot get worker secret from vault")
	}
	return c.String(http.StatusOK, globalSecret)
}

// getWorkerSecret returns the global secret for registering new worker.
func getWorkerSecret() (string, error) {
	v, err := services.DefaultVaultService()
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get vault instance", "error", err.Error())
		return "", err
	}
	if err = v.LoadSecrets(); err != nil {
		gaia.Cfg.Logger.Debug("cannot load secrets from vault", "error", err.Error())
		return "", err
	}
	secret, err := v.Get(gaia.WorkerRegisterKey)
	if err != nil {
		gaia.Cfg.Logger.Debug("global worker secret not found", "error", err.Error())
		return "", err
	}
	return string(secret[:]), nil
}

type workerStatusOverviewResponse struct {
	ActiveWorker    int   `json:"activeworker"`
	SuspendedWorker int   `json:"suspendedworker"`
	InactiveWorker  int   `json:"inactiveworker"`
	FinishedRuns    int64 `json:"finishedruns"`
	QueueSize       int   `json:"queuesize"`
}

// GetWorkerStatusOverview returns general status information about all workers.
// @Summary Get worker status overview.
// @Description Returns general status information about all workers.
// @Tags workers
// @Produce json
// @Success 200 {object} workerStatusOverviewResponse "The worker status overview response."
// @Failure 500 {string} string "Cannot get memdb service from service store."
// @Router /worker/status [get]
func (wp *WorkerProvider) GetWorkerStatusOverview(c echo.Context) error {
	response := workerStatusOverviewResponse{}

	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get memdb service from service store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Get all worker objects
	workers := db.GetAllWorker()
	for _, w := range workers {
		switch w.Status {
		case gaia.WorkerActive:
			response.ActiveWorker++
		case gaia.WorkerInactive:
			response.InactiveWorker++
		case gaia.WorkerSuspended:
			response.SuspendedWorker++
		}

		// Store overall finished runs
		response.FinishedRuns += w.FinishedRuns
	}
	// Get pipeline queue size
	response.QueueSize = wp.deps.Scheduler.CountScheduledRuns()

	// Send response back
	return c.JSON(http.StatusOK, response)
}

// GetWorker returns all workers.
// @Summary Get all workers.
// @Description Gets all workers.
// @Tags workers
// @Produce json
// @Success 200 {array} gaia.Worker "A list of workers."
// @Failure 500 {string} string "Cannot get memdb service from service store."
// @Router /worker [get]
func (wp *WorkerProvider) GetWorker(c echo.Context) error {
	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get memdb service from service store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.GetAllWorker())
}

// ResetWorkerRegisterSecret generates a new global worker registration secret
// @Summary Reset worker register secret.
// @Description Generates a new global worker registration secret.
// @Tags workers
// @Produce plain
// @Success 200 {string} string "global worker registration secret has been successfully reset"
// @Failure 500 {string} string "Vault related internal problems."
// @Router /worker/secret [post]
func (wp *WorkerProvider) ResetWorkerRegisterSecret(c echo.Context) error {
	// Get vault service
	v, err := services.DefaultVaultService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get vault service from service store", "error", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Load all secrets
	err = v.LoadSecrets()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to load secrets from vault", "error", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Generate a new global worker secret
	secret := []byte(security.GenerateRandomUUIDV5())

	// Add secret and store it
	v.Add(gaia.WorkerRegisterKey, secret)
	if err := v.SaveSecrets(); err != nil {
		gaia.Cfg.Logger.Error("failed to store secrets in vault", "error", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "global worker registration secret has been successfully reset")
}
