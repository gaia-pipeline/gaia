package assethelper

import (
	rice "github.com/GeertJohan/go.rice"
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
	return loadStaticFile("rbac-policy.csv")
}

// LoadRBACAPIMappings loads the rbac-api-mappings.yml
func LoadRBACAPIMappings() (string, error) {
	return loadStaticFile("rbac-api-mappings.yml")
}

// LoadRBACModel loads the rbac-model.conf
func LoadRBACModel() (string, error) {
	return loadStaticFile("rbac-model.conf")
}
