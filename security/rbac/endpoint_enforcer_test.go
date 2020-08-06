package rbac

import (
	"errors"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/gaia-pipeline/gaia"
)

type mockEnforcer struct {
	casbin.IEnforcer
}

var mappings = APILookup{
	"/api/v1/pipeline/:pipelineid": {
		Methods: map[string]string{
			"GET": "pipelines/get",
		},
		Param: "pipelineid",
	},
}

func (m *mockEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	role := rvals[0].(string)
	if role == "admin" {
		return true, nil
	}
	if role == "failed" {
		return false, nil
	}
	return false, errors.New("error test")
}

func Test_EnforcerService_Enforce_ValidEnforcement(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("admin", "GET", "/api/v1/pipelines/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NoError(t, err)
}

func Test_EnforcerService_Enforce_FailedEnforcement(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("failed", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.EqualError(t, err, "Permission denied. Must have pipelines/get test")
}

func Test_EnforcerService_Enforce_ErrorEnforcement(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("error", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.EqualError(t, err, "error enforcing rbac: error test")
}

func Test_EnforcerService_Enforce_EndpointParamMissing(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("readonly", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{})
	assert.EqualError(t, err, "error param pipelineid missing")
}
