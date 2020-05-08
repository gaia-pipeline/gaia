package rbac

import (
	"github.com/gaia-pipeline/gaia"
)

const (
	// PipelineNamespace (DO NOT CHANGE)
	PipelineNamespace gaia.AuthPolicyNamespace = "pipelines"
	// PipelineRunNamespace (DO NOT CHANGE)
	PipelineRunNamespace gaia.AuthPolicyNamespace = "pipeline-runs"
	// SecretNamespace (DO NOT CHANGE)
	SecretNamespace gaia.AuthPolicyNamespace = "secrets"
	// UserNamespace (DO NOT CHANGE)
	UserNamespace gaia.AuthPolicyNamespace = "users"
	// UserPermissionNamespace (DO NOT CHANGE)
	UserPermissionNamespace gaia.AuthPolicyNamespace = "user-permissions"
	// WorkerNamespace (DO NOT CHANGE)
	WorkerNamespace gaia.AuthPolicyNamespace = "workers"

	// CreateAction (DO NOT CHANGE)
	CreateAction gaia.AuthPolicyAction = "create"
	// ListAction (DO NOT CHANGE)
	ListAction gaia.AuthPolicyAction = "list"
	// GetAction (DO NOT CHANGE)
	GetAction gaia.AuthPolicyAction = "get"
	// UpdateAction (DO NOT CHANGE)
	UpdateAction gaia.AuthPolicyAction = "update"
	// DeleteAction (DO NOT CHANGE)
	DeleteAction gaia.AuthPolicyAction = "delete"

	// StartAction (DO NOT CHANGE)
	StartAction gaia.AuthPolicyAction = "start"
	// StopAction (DO NOT CHANGE)
	StopAction gaia.AuthPolicyAction = "stop"
	// LogsAction (DO NOT CHANGE)
	LogsAction gaia.AuthPolicyAction = "logs"

	// ChangePasswordAction (DO NOT CHANGE)
	ChangePasswordAction gaia.AuthPolicyAction = "change-password"

	// GetRegistrationSecretAction (DO NOT CHANGE)
	GetRegistrationSecretAction gaia.AuthPolicyAction = "get-registration-secret"
	// GetOverviewAction (DO NOT CHANGE)
	GetOverviewAction gaia.AuthPolicyAction = "get-overview"
	// GetWorkerAction (DO NOT CHANGE)
	GetWorkerAction gaia.AuthPolicyAction = "get-worker"
	// DeregisterWorkerAction (DO NOT CHANGE)
	DeregisterWorkerAction gaia.AuthPolicyAction = "deregister-worker"
	// ResetWorkerRegisterSecretAction (DO NOT CHANGE)
	ResetWorkerRegisterSecretAction gaia.AuthPolicyAction = "register-worker-secret"
)
