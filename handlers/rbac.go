package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/resourcehelper"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/labstack/echo"
)

type rbacHandler struct {
	store          store.GaiaStore
	rbacMarshaller resourcehelper.Marshaller
}

func newRBACHandler(store store.GaiaStore, rbacMarshaller resourcehelper.Marshaller) *rbacHandler {
	return &rbacHandler{store: store, rbacMarshaller: rbacMarshaller}
}

// RBACPolicyPut creates or updates a new authorization.rbac resource.
func (h rbacHandler) RBACPolicyPut(c echo.Context) error {
	bts, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	var spec gaia.RBACPolicyV1
	if err := h.rbacMarshaller.Unmarshal(bts, &spec); err != nil {
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	if err := h.store.ResourceAuthRBACPut(spec); err != nil {
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	return c.String(http.StatusOK, "Policy saved successfully.")
}

// RBACPolicyPut gets an authorization.rbac resource.
func (h rbacHandler) RBACPolicyGet(c echo.Context) error {
	name := c.Param("name")

	policy, err := h.store.ResourceAuthRBACGet(name)
	if err != nil {
		return c.String(http.StatusBadRequest, "Error getting policy.")
	}

	bts, err := h.rbacMarshaller.Marshal(policy)
	if err != nil {
		return c.String(http.StatusBadRequest, "Error getting policy.")
	}

	return c.String(http.StatusOK, string(bts))
}
