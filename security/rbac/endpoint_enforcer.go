package rbac

import (
	"fmt"
	"strings"

	"github.com/gaia-pipeline/gaia"
)

// EndpointEnforcer represents the interface for enforcing RBAC using the echo.Context.
type EndpointEnforcer interface {
	Enforce(username, method, path string, params map[string]string) error
}

// Enforce uses the echo.Context to enforce RBAC. Uses the apiLookup to apply policies to specific endpoints.
func (e *enforcerService) Enforce(username, method, path string, params map[string]string) error {
	group := e.rbacapiLookup

	endpoint, ok := group[path]
	if !ok {
		gaia.Cfg.Logger.Warn("path not mapped to api group", "method", method, "path", path)
		return nil
	}

	perm, ok := endpoint.Methods[method]
	if !ok {
		gaia.Cfg.Logger.Warn("method not mapped to api group path", "path", path, "method", method)
		return nil
	}

	splitAction := strings.Split(perm, "/")
	namespace := splitAction[0]
	action := splitAction[1]

	fullResource := "*"
	if endpoint.Param != "" {
		param := params[endpoint.Param]
		if param == "" {
			return fmt.Errorf("error param %s missing", endpoint.Param)
		}
		fullResource = param
	}

	allow, err := e.enforcer.Enforce(username, namespace, action, fullResource)
	if err != nil {
		return fmt.Errorf("error enforcing rbac: %w", err)
	}
	if !allow {
		return NewErrPermissionDenied(namespace, action, fullResource)
	}

	return nil
}

// ErrPermissionDenied is for when the RBAC enforcement check fails.
type ErrPermissionDenied struct {
	namespace string
	action    string
	resource  string
}

// NewErrPermissionDenied creates a new ErrPermissionDenied.
func NewErrPermissionDenied(namespace string, action string, resource string) *ErrPermissionDenied {
	return &ErrPermissionDenied{namespace: namespace, action: action, resource: resource}
}

func (e *ErrPermissionDenied) Error() string {
	msg := fmt.Sprintf("Permission denied. Must have %s/%s", e.namespace, e.action)
	if e.resource != "*" {
		msg = fmt.Sprintf("%s %s", msg, e.resource)
	}
	return msg
}
