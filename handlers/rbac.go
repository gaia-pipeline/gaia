package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/resourcehelper"
	"github.com/gaia-pipeline/gaia/security/rbac"
)

type rbacHandler struct {
	svc            rbac.Service
	rbacMarshaller resourcehelper.Marshaller
}

func newRBACHandler(svc rbac.Service, rbacMarshaller resourcehelper.Marshaller) *rbacHandler {
	return &rbacHandler{svc: svc, rbacMarshaller: rbacMarshaller}
}

// RBACPolicyResourcePut creates or updates a new authorization.policy resource.
func (h *rbacHandler) RBACPolicyResourcePut(c echo.Context) error {
	bts, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to read request body: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	var policy gaia.RBACPolicyResourceV1
	if err := h.rbacMarshaller.Unmarshal(bts, &policy); err != nil {
		gaia.Cfg.Logger.Error("failed to unmarshal auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	if err := h.svc.PutPolicy(policy); err != nil {
		gaia.Cfg.Logger.Error("failed to put auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	return c.String(http.StatusOK, "Policy saved successfully.")
}

// RBACPolicyResourcePut gets an authorization.policy resource.
func (h *rbacHandler) RBACPolicyResourceGet(c echo.Context) error {
	name := c.Param("name")

	policy, err := h.svc.GetPolicy(name)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error getting policy.")
	}

	bts, err := h.rbacMarshaller.Marshal(policy)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to marshal auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error getting policy.")
	}

	return c.String(http.StatusOK, string(bts))
}

func (h *rbacHandler) RBACPolicyBindingPut(c echo.Context) error {
	name := c.Param("name")
	username := c.Param("username")

	if err := h.svc.PutUserBinding(username, name); err != nil {
		gaia.Cfg.Logger.Error("failed to put policy binding: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving policy binding.")
	}

	return c.String(http.StatusOK, "Successfully assignment role")
}
