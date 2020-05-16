package rbac

import (
	"github.com/gaia-pipeline/gaia"
)

const (
	// PipelineNamespace (DO NOT CHANGE)
	PipelineNamespace gaia.RBACPolicyNamespace = "pipelines"
	// PipelineRunNamespace (DO NOT CHANGE)
	PipelineRunNamespace gaia.RBACPolicyNamespace = "pipeline-runs"
	// SecretNamespace (DO NOT CHANGE)
	SecretNamespace gaia.RBACPolicyNamespace = "secrets"
	// UserNamespace (DO NOT CHANGE)
	UserNamespace gaia.RBACPolicyNamespace = "users"
	// UserPermissionNamespace (DO NOT CHANGE)
	UserPermissionNamespace gaia.RBACPolicyNamespace = "user-permissions"
	// WorkerNamespace (DO NOT CHANGE)
	WorkerNamespace gaia.RBACPolicyNamespace = "workers"

	// CreateAction (DO NOT CHANGE)
	CreateAction gaia.RBACPolicyAction = "create"
	// ListAction (DO NOT CHANGE)
	ListAction gaia.RBACPolicyAction = "list"
	// GetAction (DO NOT CHANGE)
	GetAction gaia.RBACPolicyAction = "get"
	// UpdateAction (DO NOT CHANGE)
	UpdateAction gaia.RBACPolicyAction = "update"
	// DeleteAction (DO NOT CHANGE)
	DeleteAction gaia.RBACPolicyAction = "delete"

	// StartAction (DO NOT CHANGE)
	StartAction gaia.RBACPolicyAction = "start"
	// StopAction (DO NOT CHANGE)
	StopAction gaia.RBACPolicyAction = "stop"
	// LogsAction (DO NOT CHANGE)
	LogsAction gaia.RBACPolicyAction = "logs"

	// ChangePasswordAction (DO NOT CHANGE)
	ChangePasswordAction gaia.RBACPolicyAction = "change-password"

	// GetRegistrationSecretAction (DO NOT CHANGE)
	GetRegistrationSecretAction gaia.RBACPolicyAction = "get-registration-secret"
	// GetOverviewAction (DO NOT CHANGE)
	GetOverviewAction gaia.RBACPolicyAction = "get-overview"
	// GetWorkerAction (DO NOT CHANGE)
	GetWorkerAction gaia.RBACPolicyAction = "get-worker"
	// DeregisterWorkerAction (DO NOT CHANGE)
	DeregisterWorkerAction gaia.RBACPolicyAction = "deregister-worker"
	// ResetWorkerRegisterSecretAction (DO NOT CHANGE)
	ResetWorkerRegisterSecretAction gaia.RBACPolicyAction = "register-worker-secret"
)
