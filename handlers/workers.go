package handlers

import (
	"net/http"

	"github.com/labstack/echo"
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

	return nil
}
