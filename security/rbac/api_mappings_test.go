package rbac

import (
	"fmt"
	"net/http"
	"testing"
)

// This test loads in the rbac-api-mappings.yml.
// It allows us to verify any changes made in the future.
// CHanging something like a namespace/mapping could have severe consequences.
func Test_RBACAPIMappings(t *testing.T) {
	test := []struct {
		path         string
		expectedPerm string
		method       string
	}{
		{
			path:         "/api/v1/pipeline",
			method:       http.MethodGet,
			expectedPerm: "pipelines/create",
		},
		{
			path:         "/api/v1/pipeline/gitlsremote",
			method:       http.MethodPost,
			expectedPerm: "pipelines/create",
		},
		{
			path:         "/api/v1/pipeline/name",
			method:       http.MethodGet,
			expectedPerm: "pipelines/create",
		},
		{
			path:         "/api/v1/pipeline/periodicschedules",
			method:       http.MethodPost,
			expectedPerm: "pipelines/create",
		},
		{
			path:         "/api/v1/pipeline/created",
			method:       http.MethodGet,
			expectedPerm: "pipelines/list-created",
		},
		{
			path:         "/api/v1/pipeline",
			method:       http.MethodGet,
			expectedPerm: "pipelines/list",
		},
		{
			path:         "/api/v1/pipeline/latest",
			method:       http.MethodGet,
			expectedPerm: "pipelines/list-latest",
		},
		{
			path:         "/api/v1/pipeline/:pipelineid",
			method:       http.MethodGet,
			expectedPerm: "pipelines/get",
		},
		{
			path:         "/api/v1/pipeline/:pipelineid",
			method:       http.MethodPut,
			expectedPerm: "pipelines/update",
		},
		{
			path:         "/api/v1/pipeline/:pipelineid",
			method:       http.MethodDelete,
			expectedPerm: "pipelines/delete",
		},
		{
			path:         "/api/v1/pipeline/:pipelineid/pull",
			method:       http.MethodPost,
			expectedPerm: "pipelines/pull",
		},
		{
			path:         "/api/v1/pipeline/:pipelineid/reset-trigger-token",
			method:       http.MethodPut,
			expectedPerm: "pipelines/reset-trigger-token",
		},
		{
			path:         "/api/v1/pipeline/:pipelineid/start",
			method:       http.MethodPost,
			expectedPerm: "pipelines/start",
		},
		{
			path:         "/api/v1/pipelinerun/:pipelineid/:runid/stop",
			method:       http.MethodPost,
			expectedPerm: "pipelines:runs/stop",
		},
		{
			path:         "/api/v1/pipelinerun/:pipelineid/:runid/latest",
			method:       http.MethodGet,
			expectedPerm: "pipelines:runs/get-latest-run",
		},
		{
			path:         "/api/v1/pipelinerun/:pipelineid/:runid",
			method:       http.MethodGet,
			expectedPerm: "pipelines:runs/get-run",
		},
		{
			path:         "/api/v1/pipelinerun/:pipelineid/:runid/log",
			method:       http.MethodGet,
			expectedPerm: "pipelines:runs/get-run",
		},
		{
			path:         "/api/v1/pipelinerun/:pipelineid/latest",
			method:       http.MethodGet,
			expectedPerm: "pipelines:runs/get-latest",
		},
		{
			path:         "/api/v1/pipelinerun/:pipelineid",
			method:       http.MethodGet,
			expectedPerm: "pipelines:runs/get",
		},
		{
			path:         "/api/v1/secret",
			method:       http.MethodPost,
			expectedPerm: "secrets/create",
		},
		{
			path:         "/api/v1/secrets",
			method:       http.MethodGet,
			expectedPerm: "secrets/list",
		},
		{
			path:         "/api/v1/secret/update",
			method:       http.MethodPut,
			expectedPerm: "secrets/update",
		},
		{
			path:         "/api/v1/secret/:key",
			method:       http.MethodDelete,
			expectedPerm: "secrets/delete",
		},
		{
			path:         "/api/v1/user",
			method:       http.MethodPost,
			expectedPerm: "users/create",
		},
		{
			path:         "/api/v1/users",
			method:       http.MethodGet,
			expectedPerm: "users/list",
		},
		{
			path:         "/api/v1/user/password",
			method:       http.MethodPost,
			expectedPerm: "users/change-password",
		},
		{
			path:         "/api/v1/user/:username",
			method:       http.MethodDelete,
			expectedPerm: "users/delete",
		},
		{
			path:         "/api/v1/user/:username/reset-trigger-token",
			method:       http.MethodPut,
			expectedPerm: "users/reset-trigger-token",
		},
		{
			path:         "/api/v1/worker/secret",
			method:       http.MethodPost,
			expectedPerm: "workers/create-secret",
		},
		{
			path:         "/api/v1/worker/status",
			method:       http.MethodGet,
			expectedPerm: "workers/status-list",
		},
		{
			path:         "/api/v1/worker",
			method:       http.MethodGet,
			expectedPerm: "workers/list",
		},
		{
			path:         "/api/v1/worker/secret",
			method:       http.MethodGet,
			expectedPerm: "workers/get-secret",
		},
		{
			path:         "/api/v1/worker/:workerid",
			method:       http.MethodDelete,
			expectedPerm: "workers/deregister",
		},
		{
			path:         "/api/v1/worker/status",
			method:       http.MethodGet,
			expectedPerm: "workers/get-status",
		},
		{
			path:         "/api/v1/settings/poll",
			method:       http.MethodGet,
			expectedPerm: "settings/get",
		},
		{
			path:         "/api/v1/settings/rbac",
			method:       http.MethodGet,
			expectedPerm: "settings/get",
		},
		{
			path:         "/api/v1/settings/poll/on",
			method:       http.MethodPost,
			expectedPerm: "settings/update",
		},
		{
			path:         "/api/v1/settings/poll/off",
			method:       http.MethodPost,
			expectedPerm: "settings/update",
		},
		{
			path:         "/api/v1/settings/rbac",
			method:       http.MethodPut,
			expectedPerm: "settings/update",
		},
		{
			path:         "/api/v1/rbac/roles",
			method:       http.MethodGet,
			expectedPerm: "rbac:roles/list",
		},
		{
			path:         "/api/v1/rbac/roles/:role",
			method:       http.MethodPut,
			expectedPerm: "rbac:roles/create",
		},
		{
			path:         "/api/v1/rbac/roles/:role",
			method:       http.MethodDelete,
			expectedPerm: "rbac:roles/delete",
		},
		{
			path:         "/api/v1/rbac/roles/:role/attach/:username",
			method:       http.MethodPut,
			expectedPerm: "rbac:roles/attach",
		},
		{
			path:         "/api/v1/rbac/roles/:role/attach/:username",
			method:       http.MethodDelete,
			expectedPerm: "rbac:roles/detach",
		},
		{
			path:         "/api/v1/rbac/roles/:role/attached",
			method:       http.MethodGet,
			expectedPerm: "rbac:roles/get-attached",
		},
		{
			path:         "/api/v1/users/:username/rbac/roles",
			method:       http.MethodGet,
			expectedPerm: "users/get-roles",
		},
	}

	mappings, _ := LoadAPILookup()

	for _, tt := range test {
		t.Run(fmt.Sprintf("test mapping route:%s:%s", tt.method, tt.path), func(t *testing.T) {
			mapping, hasMapping := mappings[tt.path]
			if !hasMapping {
				t.Errorf("no route found for: %s - %s", tt.method, tt.path)
				return
			}
			_, hasMethod := mapping.Methods[tt.method]
			if !hasMethod {
				t.Errorf("no method found for: %s - %s", tt.method, tt.path)
				return
			}
		})
	}
}
