package rbac

import (
	"fmt"
	"strings"

	"github.com/gaia-pipeline/gaia"
)

// EndpointEnforcer represents the interface for enforcing RBAC using the echo.Context.
type EndpointEnforcer interface {
	Enforce(username, method, path string, params map[string]string) (bool, error)
}

// Enforce uses the echo.Context to enforce RBAC. Uses the apiLookup to apply policies to specific endpoints.
func (e *enforcerService) Enforce(username, method, path string, params map[string]string) (bool, error) {
	group := e.rbacapiLookup

	endpoint, ok := group[path]
	if !ok {
		gaia.Cfg.Logger.Warn("path not mapped to api group", "method", method, "path", path)
		return true, nil
	}

	perm, ok := endpoint.Methods[method]
	if !ok {
		gaia.Cfg.Logger.Warn("method not mapped to api group path", "path", path, "method", method)
		return true, nil
	}

	splitAction := strings.Split(perm, "/")
	namespace := splitAction[0]
	action := splitAction[1]

	fullResource := "*"
	if endpoint.Param != "" {
		param := params[endpoint.Param]
		if param == "" {
			return false, fmt.Errorf("error param %s missing", endpoint.Param)
		}
		fullResource = param
	}

	valid, err := e.enforcer.Enforce(username, namespace, action, fullResource)
	if err != nil {
		return false, fmt.Errorf("error enforcing rbac: %w", err)
	}

	return valid, nil
}
