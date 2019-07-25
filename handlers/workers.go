package handlers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gaia-pipeline/gaia/security"

	"github.com/Pallinder/go-randomdata"
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
	UniqueID string `json:"uniqueid"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"cacert"`
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

	// Generate name if none was given
	if worker.Name == "" {
		worker.Name = randomdata.SillyName() + "_" + randomdata.SillyName()
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
	crtPath, keyPath, err := cert.CreateSignedCertWithValidOpts("", hoursBeforeValid, hoursAfterValid)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create signed certificate", "error", err.Error())
		return c.String(http.StatusInternalServerError, "cannot create signed certificate")
	}
	defer func() {
		if err := cert.CleanupCerts(crtPath, keyPath); err != nil {
			gaia.Cfg.Logger.Error("failed to remove worker certificates", "error", err)
		}
	}()

	// Get public cert from CA (required for mTLS)
	caCertPath, _ := cert.GetCACertPath()
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
	crtB64 := base64.StdEncoding.EncodeToString([]byte(crt))
	keyB64 := base64.StdEncoding.EncodeToString([]byte(key))
	caCertB64 := base64.StdEncoding.EncodeToString([]byte(caCert))

	// Register worker by adding it to the memdb and store
	db, err := services.MemDBService(nil)
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
func DeregisterWorker(c echo.Context) error {
	workerID := c.Param("workerid")
	if workerID == "" {
		return c.String(http.StatusBadRequest, "worker id is missing")
	}

	// Get memdb service
	db, err := services.MemDBService(nil)
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

type workerStatusOverviewRespoonse struct {
	ActiveWorker    int   `json:"activeworker"`
	SuspendedWorker int   `json:"suspendedworker"`
	InactiveWorker  int   `json:"inactiveworker"`
	FinishedRuns    int64 `json:"finishedruns"`
	QueueSize       int   `json:"queuesize"`
}

// GetWorkerStatusOverview returns general status information about all workers.
func GetWorkerStatusOverview(c echo.Context) error {
	response := workerStatusOverviewRespoonse{}

	// Get memdb service
	db, err := services.MemDBService(nil)
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

	// Get scheduler service
	scheduler, err := services.SchedulerService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get scheduler service from service store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Get pipeline queue size
	response.QueueSize = scheduler.CountScheduledRuns()

	// Send response back
	return c.JSON(http.StatusOK, response)
}

// GetWorker returns all workers.
func GetWorker(c echo.Context) error {
	// Get memdb service
	db, err := services.MemDBService(nil)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get memdb service from service store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, db.GetAllWorker())
}

// ResetWorkerRegisterSecret generates a new global worker registration secret
func ResetWorkerRegisterSecret(c echo.Context) error {
	// Get vault service
	v, err := services.VaultService(nil)
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

func GetPipelineRepositoryInformation(c echo.Context) error {
	store, err := services.StorageService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to initialise store")
	}
	pipelineName := c.Param("name")

	repo, err := store.PipelineGetByName(pipelineName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to get pipeline")
	}

	return c.JSON(http.StatusOK, repo.Repo)
}
