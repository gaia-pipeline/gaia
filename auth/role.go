package auth

import "github.com/gaia-pipeline/gaia"

// Constructor for creating new UserRoleEndpoints.
func NewUserRoleEndpoint(method string, path string) *gaia.UserRoleEndpoint {
	return &gaia.UserRoleEndpoint{Path: path, Method: method}
}

// Creates a full user role name {category}{role}.
func FullUserRoleName(category *gaia.UserRoleCategory, role *gaia.UserRole) string {
	return category.Name + role.Name
}

// Flattens the given user categories into a single slice of {category}{role}s.
func FlattenUserCategoryRoles(cats []*gaia.UserRoleCategory) []string {
	var roles []string
	for _, category := range cats {
		for _, r := range category.Roles {
			roles = append(roles, FullUserRoleName(category, r))
		}
	}
	return roles
}

var (
	// All the default user categories and roles.
	DefaultUserRoles = []*gaia.UserRoleCategory{
		{
			Name:        "Pipeline",
			Description: "Managing and initiating pipelines.",
			Roles: []*gaia.UserRole{
				{
					Name: "Create",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/gitlsremote"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/name"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/githook"),
					},
					Description: "Create new pipelines.",
				},
				{
					Name: "List",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/created"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/latest"),
					},
					Description: "List created pipelines.",
				},
				{
					Name: "Get",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Get created pipelines.",
				},
				{
					Name: "Update",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Update created pipelines.",
				},
				{
					Name: "Delete",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Delete created pipelines.",
				},
				{
					Name: "Start",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/:pipelineid/start"),
					},
					Description: "Start created pipelines.",
				},
			},
		},
		{
			Name:        "PipelineRun",
			Description: "Managing of pipeline runs.",
			Roles: []*gaia.UserRole{
				{
					Name: "Stop",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipelinerun/:pipelineid/:runid/stop"),
					},
					Description: "Stop running pipelines.",
				},
				{
					Name: "Get",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid"),
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/latest"),
					},
					Description: "Get pipeline runs.",
				},
				{
					Name: "List",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "pipelinerun/:pipelineid"),
					},
					Description: "List pipeline runs.",
				},
				{
					Name: "Logs",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid/latest"),
					},
					Description: "Get logs for pipeline runs.",
				},
			},
		},
		{
			Name:        "Secret",
			Description: "Managing of stored secrets used within pipelines.",
			Roles: []*gaia.UserRole{
				{
					Name: "List",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/secrets"),
					},
					Description: "List created secrets.",
				},
				{
					Name: "Delete",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/secret/:key"),
					},
					Description: "Delete created secrets.",
				},
				{
					Name: "Create",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/secret"),
					},
					Description: "Create new secrets.",
				},
				{
					Name: "Update",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/secret/update"),
					},
					Description: "Update created secrets.",
				},
			},
		},
		{
			Name:        "User",
			Description: "Managing of users.",
			Roles: []*gaia.UserRole{
				{
					Name: "Create",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user"),
					},
					Description: "Create new users.",
				},
				{
					Name: "List",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/users"),
					},
					Description: "List created users.",
				},
				{
					Name: "ChangePassword",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user/password"),
					},
					Description: "Change created users passwords.",
				},
				{
					Name: "Delete",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/user/:username"),
					},
					Description: "Delete created users.",
				},
			},
		},
		{
			Name:        "UserPermission",
			Description: "Managing of user permissions.",
			Roles: []*gaia.UserRole{
				{
					Name: "Get",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/user/:username/permissions"),
					},
					Description: "Get created users permissions.",
				},
				{
					Name: "Update",
					ApiEndpoint: []*gaia.UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/user/:username/permissions"),
					},
					Description: "Update created users permissions.",
				},
			},
		},
	}
)
