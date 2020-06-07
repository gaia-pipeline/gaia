package rbac

import (
	"github.com/casbin/casbin/v2"

	"github.com/gaia-pipeline/gaia"
)

type enforcerService struct {
	enforcer        casbin.IEnforcer
	rbacapiMappings gaia.RBACAPIMappings
}

// NewEnforcerSvc creates a new enforcerService.
func NewEnforcerSvc(enforcer casbin.IEnforcer, apiMappingsFile string) (*enforcerService, error) {
	rbacapiMappings, err := loadAPIMappings(apiMappingsFile)
	if err != nil {
		return nil, err
	}

	return &enforcerService{
		enforcer:        enforcer,
		rbacapiMappings: rbacapiMappings,
	}, nil
}
