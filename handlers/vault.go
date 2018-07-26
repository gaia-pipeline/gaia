package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/security"

	"github.com/labstack/echo"
)

type secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type updateSecret struct {
	Key   string `json:"key"`
	Value string `json:"newvalue"`
}

// UpdateSecret updates a secret using the vault.
func UpdateSecret(c echo.Context) error {
	s := new(updateSecret)
	err := c.Bind(s)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	cert, err := security.InitCA()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v, err := security.NewVault(cert)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Add(s.Key, []byte(s.Value))
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "secret successfully updated")
}

// AddSecret creates a secret using the vault.
func AddSecret(c echo.Context) error {
	s := new(secret)
	err := c.Bind(s)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	cert, err := security.InitCA()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v, err := security.NewVault(cert)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Add(s.Key, []byte(s.Value))
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusCreated, "secret successfully added")
}

// ListSecrets retrieves all secrets from the vault.
func ListSecrets(c echo.Context) error {
	secrets := make([]secret, 0)
	cert, err := security.InitCA()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v, err := security.NewVault(cert)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	kvs := v.GetAll()
	for _, k := range kvs {
		s := secret{Key: k, Value: "**********"}
		secrets = append(secrets, s)
	}
	return c.JSON(http.StatusOK, secrets)
}

// RemoveSecret removes a secret from the vault.
func RemoveSecret(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "invalid key given")
	}
	cert, err := security.InitCA()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v, err := security.NewVault(cert)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Remove(key)
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "secret successfully deleted")
}
