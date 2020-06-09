package rbac

import (
	"errors"
	"github.com/casbin/casbin/v2"
	"testing"

	"github.com/hashicorp/go-hclog"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"

	"github.com/gaia-pipeline/gaia"
)

type mockEnforcer struct {
	casbin.IEnforcer
}

var mappings = gaia.RBACAPIMappings{
	Endpoints: map[string]gaia.RBACAPIMappingEndpoint{
		"/api/v1/pipeline/:pipelineid": {
			Methods: map[string]string{
				"GET": "pipelines/get",
			},
			Param: "pipelineid",
		},
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
		enforcer:        &mockEnforcer{},
		rbacapiMappings: mappings,
	}

	getSuccess, err := svc.Enforce("admin", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(getSuccess, true))
}

func Test_EnforcerService_Enforce_FailedEnforcement(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:        &mockEnforcer{},
		rbacapiMappings: mappings,
	}

	getSuccess, err := svc.Enforce("failed", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(getSuccess, false))
}

func Test_EnforcerService_Enforce_ErrorEnforcement(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:        &mockEnforcer{},
		rbacapiMappings: mappings,
	}

	getSuccess, err := svc.Enforce("error", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.Check(t, cmp.Error(err, "error test"))
	assert.Check(t, cmp.Equal(getSuccess, false))
}

func Test_EnforcerService_Enforce_EndpointParamMissing(t *testing.T) {
	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:        &mockEnforcer{},
		rbacapiMappings: mappings,
	}

	getSuccess, err := svc.Enforce("readonly", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{})
	assert.Check(t, cmp.Error(err, "param pipelineid missing"))
	assert.Check(t, cmp.Equal(getSuccess, false))
}
