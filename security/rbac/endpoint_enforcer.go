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

// Enforce uses the echo.Context to enforce RBAC. Uses the rbacapiMappings to apply policies to specific endpoints.
func (e *enforcerService) Enforce(username, method, path string, params map[string]string) (bool, error) {
	group := e.rbacapiMappings

	endpoint, ok := group.Endpoints[path]
	if !ok {
		gaia.Cfg.Logger.Warn("path not mapped to api group", "path", path)
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
			return false, fmt.Errorf("param %s missing", endpoint.Param)
		}
		fullResource = param
	}

	valid, err := e.enforcer.Enforce(username, namespace, action, fullResource)
	if err != nil {
		return false, err
	}

	return valid, nil
}