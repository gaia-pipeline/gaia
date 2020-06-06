package rbac

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/gaia-pipeline/gaia"
)

// LoadAPIMappings gets the RBAC API mappings from the defined YAML file.
func LoadAPIMappings() (gaia.RBACAPIMappings, error) {
	file, err := ioutil.ReadFile("security/rbac/rbac-api-mappings.yml")
	if err != nil {
		return gaia.RBACAPIMappings{}, err
	}

	var apiGroup gaia.RBACAPIMappings
	if err := yaml.Unmarshal(file, &apiGroup); err != nil {
		return gaia.RBACAPIMappings{}, err
	}

	return apiGroup, nil
}
