package gaia

var (
	// TODO: Probably load these in via a config file or something.
	UserRoleCategories = []*UserRoleCategory{
		{
			Name: "Pipeline",
			Roles: []*UserRole{
				{
					Name: "Create",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/gitlsremote"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/name"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/githook"),
					},
				},
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/created"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/latest"),
					},
				},
				{
					Name: "Get",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/:pipelineid"),
					},
				},
				{
					Name: "Update",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/pipeline/:pipelineid"),
					},
				},
				{
					Name: "Delete",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/pipeline/:pipelineid"),
					},
				},
				{
					Name: "Start",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/:pipelineid/start"),
					},
				},
			},
		},
		{
			Name: "PipelineRun",
			Roles: []*UserRole{
				{
					Name: "Stop",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipelinerun/:pipelineid/:runid/stop"),
					},
				},
				{
					Name: "Get",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid"),
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/latest"),
					},
				},
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "pipelinerun/:pipelineid"),
					},
				},
				{
					Name: "Logs",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid/latest"),
					},
				},
			},
		},
		{
			Name: "Secret",
			Roles: []*UserRole{
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/secrets"),
					},
				},
				{
					Name: "Delete",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/secret/:key"),
					},
				},
				{
					Name: "Set",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/secret"),
					},
				},
				{
					Name: "Update",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/secret/update"),
					},
				},
			},
		},
		{
			Name: "User",
			Roles: []*UserRole{
				{
					Name: "Create",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user"),
					},
				},
				{
					Name: "List",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/users"),
					},
				},
				{
					Name: "ChangePassword",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user/password"),
					},
				},
				{
					Name: "Delete",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/user/:username"),
					},
				},
			},
		},
		{
			Name: "UserPermission",
			Roles: []*UserRole{
				{
					Name: "Get",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/user/:username/permissions"),
					},
				},
				{
					Name: "Save",
					ApiEndpoint: []*UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/user/:username/permissions"),
					},
				},
			},
		},
	}
)
