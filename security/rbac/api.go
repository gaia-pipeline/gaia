package rbac

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/gaia-pipeline/gaia"
)

func loadAPIMappings(apiMappingsFile string) (gaia.RBACAPIMappings, error) {
	file, err := ioutil.ReadFile(apiMappingsFile)
	if err != nil {
		return gaia.RBACAPIMappings{}, err
	}

	var apiGroup gaia.RBACAPIMappings
	if err := yaml.Unmarshal(file, &apiGroup); err != nil {
		return gaia.RBACAPIMappings{}, err
	}

	return apiGroup, nil
}
