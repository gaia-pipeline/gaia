package gaia

var (
	// TODO: Load these in via a file
	UserRoleCategories = []*UserRoleCategory{
		{
			Name:        "Pipeline",
			Description: "Managing and initiating pipelines.",
			Roles: []*UserRole{
				{
					Name: "Create",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/gitlsremote"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/name"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/githook"),
					},
					Description: "Create new pipelines.",
				},
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/created"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/latest"),
					},
					Description: "List created pipelines.",
				},
				{
					Name: "Get",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Get created pipelines.",
				},
				{
					Name: "Update",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Update created pipelines.",
				},
				{
					Name: "Delete",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Delete created pipelines.",
				},
				{
					Name: "Start",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/:pipelineid/start"),
					},
					Description: "Start created pipelines.",
				},
			},
		},
		{
			Name:        "PipelineRun",
			Description: "Managing of pipeline runs.",
			Roles: []*UserRole{
				{
					Name: "Stop",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipelinerun/:pipelineid/:runid/stop"),
					},
					Description: "Stop running pipelines.",
				},
				{
					Name: "Get",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid"),
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/latest"),
					},
					Description: "Get pipeline runs.",
				},
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "pipelinerun/:pipelineid"),
					},
					Description: "List pipeline runs.",
				},
				{
					Name: "Logs",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid/latest"),
					},
					Description: "Get logs for pipeline runs.",
				},
			},
		},
		{
			Name:        "Secret",
			Description: "Managing of stored secrets used within pipelines.",
			Roles: []*UserRole{
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/secrets"),
					},
					Description: "List created secrets.",
				},
				{
					Name: "Delete",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/secret/:key"),
					},
					Description: "Delete created secrets.",
				},
				{
					Name: "Create",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/secret"),
					},
					Description: "Create new secrets.",
				},
				{
					Name: "Update",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/secret/update"),
					},
					Description: "Update created secrets.",
				},
			},
		},
		{
			Name:        "User",
			Description: "Managing of users.",
			Roles: []*UserRole{
				{
					Name: "Create",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user"),
					},
					Description: "Create new users.",
				},
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/users"),
					},
					Description: "List created users.",
				},
				{
					Name: "ChangePassword",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user/password"),
					},
					Description: "Change created users passwords.",
				},
				{
					Name: "Delete",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/user/:username"),
					},
					Description: "Delete created users.",
				},
			},
		},
		{
			Name:        "UserPermission",
			Description: "Managing of user permissions.",
			Roles: []*UserRole{
				{
					Name: "Get",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/user/:username/permissions"),
					},
					Description: "Get created users permissions.",
				},
				{
					Name: "Update",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/user/:username/permissions"),
					},
					Description: "Update created users permissions.",
				},
			},
		},
	}
)
