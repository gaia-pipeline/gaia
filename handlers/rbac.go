package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/resourcehelper"
	"github.com/gaia-pipeline/gaia/security/rbac"
	gStore "github.com/gaia-pipeline/gaia/store"
)

type rbacHandler struct {
	store          gStore.GaiaStore
	rbacSvc        rbac.Service
	rbacMarshaller resourcehelper.Marshaller
}

func newRBACHandler(store gStore.GaiaStore, rbacService rbac.Service, rbacMarshaller resourcehelper.Marshaller) *rbacHandler {
	return &rbacHandler{store: store, rbacSvc: rbacService, rbacMarshaller: rbacMarshaller}
}

// AuthPolicyResourcePut creates or updates a new authorization.policy resource.
func (h rbacHandler) AuthPolicyResourcePut(c echo.Context) error {
	bts, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to read request body: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	var policy gaia.AuthPolicyResourceV1
	if err := h.rbacMarshaller.Unmarshal(bts, &policy); err != nil {
		gaia.Cfg.Logger.Error("failed to unmarshal auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	if err := h.rbacSvc.PutPolicy(policy); err != nil {
		gaia.Cfg.Logger.Error("failed to put auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	return c.String(http.StatusOK, "Policy saved successfully.")
}

// AuthPolicyResourcePut gets an authorization.policy resource.
func (h rbacHandler) AuthPolicyResourceGet(c echo.Context) error {
	name := c.Param("name")

	policy, err := h.store.AuthPolicyResourceGet(name)
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

func (h rbacHandler) AuthPolicyAssignmentPut(c echo.Context) error {
	name := c.Param("name")
	username := c.Param("username")

	policies, err := h.store.AuthPolicyAssignmentGet(username)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get auth assignment: " + err.Error())
		return c.String(http.StatusBadRequest, "Error getting policy assignment.")
	}

	newAssignment := gaia.AuthPolicyAssignment{
		Username: username,
		Policies: append(policies.Policies, name),
	}

	if err := h.store.AuthPolicyAssignmentPut(newAssignment); err != nil {
		gaia.Cfg.Logger.Error("failed to put auth assignment: " + err.Error())
		return c.String(http.StatusBadRequest, "Error getting policy.")
	}

	return c.String(http.StatusOK, "Successfully assignment role")
}
