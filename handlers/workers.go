package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

type registerSecret struct {
	Secret string `json:"secret"`
}

// RegisterWorker allows new workers to register themself at this Gaia instance.
// It accepts a secret and returns valid certificates for further mTLS connection.
func RegisterWorker(c echo.Context) error {
	secret := registerSecret{}
	if err := c.Bind(&secret); err != nil {
		return c.String(http.StatusBadRequest, "secret for registration is invalid:"+err.Error())
	}

	// Lookup the global registration secret in our vault.
	globalSecret, err := getWorkerSecret()
	if err != nil {
		return c.String(http.StatusInternalServerError, "cannot get worker secret from vault")
	}

	// Check if given secret is equal with global worker secret.
	if globalSecret != secret.Secret {
		return c.String(http.StatusForbidden, "wrong global worker secret provided")
	}

	return nil
}

// GetWorkerRegisterSecret returns the global secret for registering new worker.
func GetWorkerRegisterSecret(c echo.Context) error {
	return nil
}

// getWorkerSecret returns the global secret for registering new worker.
// If the secret does not exist, it will generate a new one.
func getWorkerSecret() (string, error) {
	v, err := services.VaultService(nil)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get vault instance", "error", err.Error())
		return "", err
	}
	secret, err := v.Get(workerRegisterKey)
	if err != nil {
		// Secret has not been generated yet.
		secret = []byte(uuid.NewV5(uuid.NewV4(), workerRegisterKey).String())
		v.Add(workerRegisterKey, secret)
	}
	return string(secret[:]), nil
}
