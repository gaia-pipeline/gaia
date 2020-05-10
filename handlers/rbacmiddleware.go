package handlers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security/rbac"
)

type policyEnforcerMiddleware struct {
	enforcer rbac.PolicyEnforcer
}

func newPolicyEnforcerMiddleware(enforcer rbac.PolicyEnforcer) *policyEnforcerMiddleware {
	return &policyEnforcerMiddleware{
		enforcer: enforcer,
	}
}

func (pe *policyEnforcerMiddleware) do(namespace gaia.RBACPolicyNamespace, action gaia.RBACPolicyAction) echo.MiddlewareFunc {
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
