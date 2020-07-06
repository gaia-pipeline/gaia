package assethelper

import (
	rice "github.com/GeertJohan/go.rice"
)

const (
	rbacBuiltinPolicy = "rbac-policy.csv"
	rbacModel         = "rbac-model.conf"
	rbacAPIMappings   = "rbac-api-mappings.yml"
)

func loadStaticFile(filename string) (string, error) {
	box, err := rice.FindBox("../../static")
	if err != nil {
		return "", err
	}
	filestr, err := box.String(filename)
	if err != nil {
		return "", err
	}
	return filestr, nil
}

// LoadRBACBuiltinPolicy loads the builtin rbac-policy.csv
func LoadRBACBuiltinPolicy() (string, error) {
	return loadStaticFile(rbacBuiltinPolicy)
}

// LoadRBACAPIMappings loads the rbac-api-mappings.yml
func LoadRBACAPIMappings() (string, error) {
	return loadStaticFile(rbacAPIMappings)
}

// LoadRBACModel loads the rbac-model.conf
func LoadRBACModel() (string, error) {
	return loadStaticFile(rbacModel)
}
