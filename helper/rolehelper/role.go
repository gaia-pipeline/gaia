package rolehelper

import (
	"fmt"
	"github.com/gaia-pipeline/gaia"
)

// FullUserRoleName returns a full user role name in the form {category}{role}.
func FullUserRoleName(category gaia.UserRoleCategory, role gaia.UserRole) string {
	return fmt.Sprintf("%s%s", category, role)
}

// FlattenUserCategoryRoles flattens the given user categories into a single slice with items in the form off
// {category}{role}s.
func FlattenUserCategoryRoles(cats map[gaia.UserRoleCategory]*gaia.UserRoleCategoryDetails) []string {
	var roles []string
	for categoryName, category := range cats {
		for roleName := range category.Roles {
			roles = append(roles, FullUserRoleName(categoryName, roleName))
		}
	}
	return roles
}

const (
	// PipelineCategory (DO NOT CHANGE)
	PipelineCategory gaia.UserRoleCategory = "Pipeline"
	// PipelineRunCategory (DO NOT CHANGE)
	PipelineRunCategory gaia.UserRoleCategory = "PipelineRun"
	// SecretCategory (DO NOT CHANGE)
	SecretCategory gaia.UserRoleCategory = "Secret"
	// UserCategory (DO NOT CHANGE)
	UserCategory gaia.UserRoleCategory = "User"
	// UserPermissionCategory (DO NOT CHANGE)
	UserPermissionCategory gaia.UserRoleCategory = "UserPermission"
	// WorkerCategory (DO NOT CHANGE)
	WorkerCategory gaia.UserRoleCategory = "Worker"

	// CreateRole (DO NOT CHANGE)
	CreateRole gaia.UserRole = "Create"
	// ListRole (DO NOT CHANGE)
	ListRole gaia.UserRole = "List"
	// GetRole (DO NOT CHANGE)
	GetRole gaia.UserRole = "Get"
	// UpdateRole (DO NOT CHANGE)
	UpdateRole gaia.UserRole = "Update"
	// DeleteRole (DO NOT CHANGE)
	DeleteRole gaia.UserRole = "Delete"

	// StartRole (DO NOT CHANGE)
	StartRole gaia.UserRole = "Start"
	// StopRole (DO NOT CHANGE)
	StopRole gaia.UserRole = "Stop"
	// LogsRole (DO NOT CHANGE)
	LogsRole gaia.UserRole = "Logs"

	// ChangePasswordRole (DO NOT CHANGE)
	ChangePasswordRole gaia.UserRole = "ChangePassword"

	// GetRegistrationSecretRole (DO NOT CHANGE)
	GetRegistrationSecretRole gaia.UserRole = "GetRegistrationSecret"
	// GetOverviewRole (DO NOT CHANGE)
	GetOverviewRole gaia.UserRole = "GetOverview"
	// GetWorkerRole (DO NOT CHANGE)
	GetWorkerRole gaia.UserRole = "GetWorker"
	// DeregisterWorkerRole (DO NOT CHANGE)
	DeregisterWorkerRole gaia.UserRole = "DeregisterWorker"
	// ResetWorkerRegisterSecretRole (DO NOT CHANGE)
	ResetWorkerRegisterSecretRole gaia.UserRole = "ResetWorkerRegisterSecret"
)

var (
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
