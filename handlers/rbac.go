package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gaia-pipeline/gaia/security/rbac"

	gStore "github.com/gaia-pipeline/gaia/store"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/resourcehelper"
	"github.com/labstack/echo"
)

type rbacHandler struct {
	store          gStore.GaiaStore
	rbacMarshaller resourcehelper.Marshaller
}

func newRBACHandler(store gStore.GaiaStore, rbacMarshaller resourcehelper.Marshaller) *rbacHandler {
	return &rbacHandler{store: store, rbacMarshaller: rbacMarshaller}
}

// AuthPolicyResourcePut creates or updates a new authorization.policy resource.
func (h rbacHandler) AuthPolicyResourcePut(c echo.Context) error {
	bts, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to read request body: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	var spec gaia.AuthPolicyResourceV1
	if err := h.rbacMarshaller.Unmarshal(bts, &spec); err != nil {
		gaia.Cfg.Logger.Error("failed to unmarshal auth policy: " + err.Error())
		return c.String(http.StatusBadRequest, "Error saving new policy.")
	}

	if err := h.store.AuthPolicyResourcePut(spec); err != nil {
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

type policyEnforcerMiddleware struct {
	enforcer rbac.PolicyEnforcer
}

func newPolicyEnforcerMiddleware(enforcer rbac.PolicyEnforcer) *policyEnforcerMiddleware {
	return &policyEnforcerMiddleware{
		enforcer: enforcer,
	}
}

func (pe *policyEnforcerMiddleware) do(namespace gaia.AuthPolicyNamespace, action gaia.AuthPolicyAction) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ctx, ok := c.(AuthContext); ok {
				if !pe.enforcer.Enforce(ctx.policies, namespace, action) {
					return c.String(http.StatusForbidden, "You do not have the required permissions.")
				}
				return next(c)
			}
			return c.String(http.StatusInternalServerError, "An error has occurred.")
		}
	}
}
