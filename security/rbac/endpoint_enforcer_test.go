package rbac

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/hashicorp/go-hclog"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"

	"github.com/gaia-pipeline/gaia"
)

func Test_EnforcerService_Enforce_RoleAdmin(t *testing.T) {
	enforcer, err := casbin.NewEnforcer("rbac-model.conf", "rbac-policy.csv")
	assert.NilError(t, err)

	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc, err := NewEnforcerSvc(enforcer, "rbac-api-mappings.yml")
	assert.NilError(t, err)

	getSuccess, err := svc.Enforce("admin", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(getSuccess, true))
}

func Test_EnforcerService_Enforce_RoleReadOnly(t *testing.T) {
	enforcer, err := casbin.NewEnforcer("rbac-model.conf", "rbac-policy.csv")
	assert.NilError(t, err)

	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc, err := NewEnforcerSvc(enforcer, "rbac-api-mappings.yml")
	assert.NilError(t, err)

	getSuccess, err := svc.Enforce("readonly", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(getSuccess, true))

	deleteSuccess, err := svc.Enforce("readonly", "DELETE", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(deleteSuccess, false))
}

func Test_EnforcerService_Enforce_EndpointParamMissing(t *testing.T) {
	enforcer, err := casbin.NewEnforcer("rbac-model.conf", "rbac-policy.csv")
	assert.NilError(t, err)

	gaia.Cfg = &gaia.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		gaia.Cfg = nil
	}()

	svc, err := NewEnforcerSvc(enforcer, "rbac-api-mappings.yml")
	assert.NilError(t, err)

	getSuccess, err := svc.Enforce("readonly", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{})
	assert.Check(t, cmp.Error(err, "param pipelineid missing"))
	assert.Check(t, cmp.Equal(getSuccess, false))
}
