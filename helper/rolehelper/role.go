package rolehelper

import (
	"fmt"
	"github.com/gaia-pipeline/gaia"
)

// NewUserRoleEndpoint is a constructor for creating new UserRoleEndpoints.
func NewUserRoleEndpoint(method string, path string) *gaia.UserRoleEndpoint {
	return &gaia.UserRoleEndpoint{Path: path, Method: method}
}

// FullUserRoleName returns a full user role name in the form {category}{role}.
func FullUserRoleName(category gaia.UserRoleCategory, role gaia.UserRole) string {
	return fmt.Sprintf("%s%s", category, role)
}

// FlattenUserCategoryRoles flattens the given user categories into a single slice with items in the form off
// {category}{role}s.
func FlattenUserCategoryRoles(cats map[gaia.UserRoleCategory]*gaia.UserRoleCategoryDetails) []string {
	var roles []string
	for categoryName, category := range cats {
		for roleName, _ := range category.Roles {
			roles = append(roles, FullUserRoleName(categoryName, roleName))
		}
	}
		return roles
}

var (
	PipelineCategory       gaia.UserRoleCategory = "Pipeline"
	PipelineRunCategory    gaia.UserRoleCategory = "PipelineRun"
	SecretCategory         gaia.UserRoleCategory = "Secret"
	UserCategory           gaia.UserRoleCategory = "User"
	UserPermissionCategory gaia.UserRoleCategory = "UserPermission"
	WorkerCategory         gaia.UserRoleCategory = "Worker"

	CreateRole gaia.UserRole = "Create"
	ListRole   gaia.UserRole = "List"
	GetRole    gaia.UserRole = "Get"
	UpdateRole gaia.UserRole = "Update"
	DeleteRole gaia.UserRole = "Delete"

	StartRole gaia.UserRole = "Start"
	StopRole  gaia.UserRole = "Stop"
	LogsRole  gaia.UserRole = "Logs"

	ChangePasswordRole gaia.UserRole = "ChangePassword"

	GetRegistrationSecretRole     gaia.UserRole = "GetRegistrationSecret"
	GetOverviewRole               gaia.UserRole = "GetOverview"
	GetWorkerRole                 gaia.UserRole = "GetWorker"
	DeregisterWorkerRole          gaia.UserRole = "DeregisterWorker"
	ResetWorkerRegisterSecretRole gaia.UserRole = "ResetWorkerRegisterSecret"

	// DefaultUserRoles contains all the default user categories and roles.
	DefaultUserRoles = map[gaia.UserRoleCategory]*gaia.UserRoleCategoryDetails{
		PipelineCategory: {
			Description: "Managing and initiating pipelines.",
			Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
				CreateRole: {
					Description: "Create new pipelines.",
				},
				ListRole: {
					Description: "List created pipelines.",
				},
				GetRole: {
					Description: "Get created pipelines.",
				},
				UpdateRole: {
					Description: "Update created pipelines.",
				},
				DeleteRole: {
					Description: "Delete created pipelines.",
				},
				StartRole: {
					Description: "Start created pipelines.",
				},
			},
		},
		PipelineRunCategory: {
			Description: "Managing of pipeline runs.",
			Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
				StopRole: {
					Description: "Stop running pipelines.",
				},
				GetRole: {
					Description: "Get pipeline runs.",
				},
				ListRole: {
					Description: "List pipeline runs.",
				},
				LogsRole: {
					Description: "Get logs for pipeline runs.",
				},
			},
		},
		SecretCategory: {
			Description: "Managing of stored secrets used within pipelines.",
			Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
				ListRole: {
					Description: "List created secrets.",
				},
				DeleteRole: {
					Description: "Delete created secrets.",
				},
				CreateRole: {
					Description: "Create new secrets.",
				},
				UpdateRole: {
					Description: "Update created secrets.",
				},
			},
		},
		UserCategory: {
			Description: "Managing of users.",
			Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
				CreateRole: {
					Description: "Create new users.",
				},
				ListRole: {
					Description: "List created users.",
				},
				ChangePasswordRole: {
					Description: "Change created users passwords.",
				},
				DeleteRole: {
					Description: "Delete created users.",
				},
			},
		},
		UserPermissionCategory: {
			Description: "Managing of user permissions.",
			Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
				GetRole: {
					Description: "Get created users permissions.",
				},
				UpdateRole: {
					Description: "Update created users permissions.",
				},
			},
		},
		WorkerCategory: {
			Description: "Managing of worker permissions.",
			Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
				GetRegistrationSecretRole: {
					Description: "Get global worker registration secret.",
				},
				GetOverviewRole: {
					Description: "Get status overview of all workers.",
				},
				GetWorkerRole: {
					Description: "Get all worker for the worker overview table.",
				},
				DeregisterWorkerRole: {
					Description: "Deregister a worker from the Gaia primary instance.",
				},
				ResetWorkerRegisterSecretRole: {
					Description: "Reset the global worker registration secret.",
				},
			},
		},
	}
)
