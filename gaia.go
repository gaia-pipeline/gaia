package gaia

import (
	"github.com/hashicorp/go-hclog"
	"github.com/robfig/cron"
	"os"
	"time"
)

// PipelineType represents supported plugin types
type PipelineType string

// CreatePipelineType represents the different status types
// a create pipeline can have.
type CreatePipelineType string

// PipelineRunStatus represents the different status a run
// can have.
type PipelineRunStatus string

// JobStatus represents the different status a job can have.
type JobStatus string

const (
	// PTypeUnknown unknown plugin type
	PTypeUnknown PipelineType = "unknown"

	// PTypeGolang golang plugin type
	PTypeGolang PipelineType = "golang"

	// PTypeJava java plugin type
	PTypeJava PipelineType = "java"

	// PTypePython python plugin type
	PTypePython PipelineType = "python"

	// PTypeCpp C++ plugin type
	PTypeCpp PipelineType = "cpp"

	// CreatePipelineFailed status
	CreatePipelineFailed CreatePipelineType = "failed"

	// CreatePipelineRunning status
	CreatePipelineRunning CreatePipelineType = "running"

	// CreatePipelineSuccess status
	CreatePipelineSuccess CreatePipelineType = "success"

	// RunNotScheduled status
	RunNotScheduled PipelineRunStatus = "not scheduled"

	// RunScheduled status
	RunScheduled PipelineRunStatus = "scheduled"

	// RunFailed status
	RunFailed PipelineRunStatus = "failed"

	// RunSuccess status
	RunSuccess PipelineRunStatus = "success"

	// RunRunning status
	RunRunning PipelineRunStatus = "running"

	// RunCancelled status
	RunCancelled PipelineRunStatus = "cancelled"

	// JobWaitingExec status
	JobWaitingExec JobStatus = "waiting for execution"

	// JobSuccess status
	JobSuccess JobStatus = "success"

	// JobFailed status
	JobFailed JobStatus = "failed"

	// JobRunning status
	JobRunning JobStatus = "running"

	// LogsFolderName represents the Name of the logs folder in pipeline run folder
	LogsFolderName = "logs"

	// LogsFileName represents the file name of the logs output
	LogsFileName = "output.log"

	// APIVersion represents the current API version
	APIVersion = "v1"

	// TmpFolder is the temp folder for temporary files
	TmpFolder = "tmp"

	// TmpPythonFolder is the name of the python temporary folder
	TmpPythonFolder = "python"

	// TmpGoFolder is the name of the golang temporary folder
	TmpGoFolder = "golang"

	// TmpCppFolder is the name of the c++ temporary folder
	TmpCppFolder = "cpp"
)

// User is the user object
type User struct {
	Username    string    `json:"username,omitempty"`
	Password    string    `json:"password,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	Tokenstring string    `json:"tokenstring,omitempty"`
	JwtExpiry   int64     `json:"jwtexpiry,omitempty"`
	LastLogin   time.Time `json:"lastlogin,omitempty"`
}

type UserPermissions struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Groups      []string `json:"groups"`
}

type PermissionCategory struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Permissions []*Permission `json:"permissions"`
}

type Permission struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ApiEndpoint *PermissionApiEndpoint `json:"api_endpoint"`
}

type PermissionApiEndpoint struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func NewPermissionApiEndpoint(path string, method string) *PermissionApiEndpoint {
	return &PermissionApiEndpoint{Path: path, Method: method}
}

type PermissionGroup struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

var (
	PermissionsCategories = []*PermissionCategory{
		{
			Name: "Pipeline",
			Permissions: []*Permission{
				{
					Name:        "Create",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline", "POST"),
				},
				{
					Name:        "GitLSRemote",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/gitlsremote", "POST"),
				},
				{
					Name:        "GetAll",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/created", "GET"),
				},
				{
					Name:        "Get",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*", "GET"),
				},
				{
					Name:        "Update",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*", "PUT"),
				},
				{
					Name:        "Delete",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*", "DELETE"),
				},
				{
					Name:        "Start",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*/start", "POST"),
				},
				{
					Name:        "GetAllWithLatestRun",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/latest", "GET"),
				},
				{
					Name:        "GitHook",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/githook", "POST"),
				},
				{
					Name:        "CheckPeriodicSchedules",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/periodicschedules", "POST"),
				},
			},
		},
		{
			Name: "PipelineRun",
			Permissions: []*Permission{
				{
					Name:        "Stop",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*/*/stop", "POST"),
				},
				{
					Name:        "Get",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*/*", "GET"),
				},
				{
					Name:        "GetAll",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*", "GET"),
				},
				{
					Name:        "GetLatest",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*/latest", "GET"),
				},
				{
					Name:        "GetLogs",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/pipeline/*/*/log", "GET"),
				},
			},
		},
		{
			Name: "Secret",
			Permissions: []*Permission{
				{
					Name:        "List",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/secrets", "GET"),
				},
				{
					Name:        "Remove",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/secret/*", "GET"),
				},
				{
					Name:        "Set",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/secret", "POST"),
				},
				{
					Name:        "Update",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/secret/update", "PUT"),
				},
			},
		},
		{
			Name: "User",
			Permissions: []*Permission{
				{
					Name:        "Create",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/user", "POST"),
				},
				{
					Name:        "List",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/users", "GET"),
				},
				{
					Name:        "ChangePassword",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/user/password", "POST"),
				},
				{
					Name:        "Delete",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/user/*", "DELETE"),
				},
				{
					Name:        "Add",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/user", "POST"),
				},
			},
		},
		{
			Name:        "UserPermission",
			Description: "Permissions relating to User Permissions.",
			Permissions: []*Permission{
				{
					Name:        "Get",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/user/*/permissions", "GET"),
				},
				{
					Name:        "Save",
					ApiEndpoint: NewPermissionApiEndpoint("/api/v1/user/*/permissions", "PUT"),
				},
			},
		},
	}
)

func (p *Permission) FullName(category string) string {
	return category + p.Name
}

// Pipeline represents a single pipeline
type Pipeline struct {
	ID                int          `json:"id,omitempty"`
	Name              string       `json:"name,omitempty"`
	Repo              GitRepo      `json:"repo,omitempty"`
	Type              PipelineType `json:"type,omitempty"`
	ExecPath          string       `json:"execpath,omitempty"`
	SHA256Sum         []byte       `json:"sha256sum,omitempty"`
	Jobs              []Job        `json:"jobs,omitempty"`
	Created           time.Time    `json:"created,omitempty"`
	UUID              string       `json:"uuid,omitempty"`
	IsNotValid        bool         `json:"notvalid,omitempty"`
	PeriodicSchedules []string     `json:"periodicschedules,omitempty"`
	CronInst          *cron.Cron   `json:"-"`
}

// GitRepo represents a single git repository
type GitRepo struct {
	URL            string     `json:"url,omitempty"`
	Username       string     `json:"user,omitempty"`
	Password       string     `json:"password,omitempty"`
	PrivateKey     PrivateKey `json:"privatekey,omitempty"`
	SelectedBranch string     `json:"selectedbranch,omitempty"`
	Branches       []string   `json:"branches,omitempty"`
	LocalDest      string     `json:"-"`
}

// Job represents a single job of a pipeline
type Job struct {
	ID           uint32     `json:"id,omitempty"`
	Title        string     `json:"title,omitempty"`
	Description  string     `json:"desc,omitempty"`
	DependsOn    []*Job     `json:"dependson,omitempty"`
	Status       JobStatus  `json:"status,omitempty"`
	Args         []Argument `json:"args,omitempty"`
	FailPipeline bool       `json:"failpipeline,omitempty"`
}

// Argument represents a single argument of a job
type Argument struct {
	Description string `json:"desc,omitempty"`
	Type        string `json:"type,omitempty"`
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
}

// CreatePipeline represents a pipeline which is not yet
// compiled.
type CreatePipeline struct {
	ID          string             `json:"id,omitempty"`
	Pipeline    Pipeline           `json:"pipeline,omitempty"`
	Status      int                `json:"status,omitempty"`
	StatusType  CreatePipelineType `json:"statustype,omitempty"`
	Output      string             `json:"output,omitempty"`
	Created     time.Time          `json:"created,omitempty"`
	GitHubToken string             `json:"githubtoken,omitempty"`
}

// PrivateKey represents a pem encoded private key
type PrivateKey struct {
	Key      string `json:"key,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// PipelineRun represents a single run of a pipeline.
type PipelineRun struct {
	UniqueID     string            `json:"uniqueid"`
	ID           int               `json:"id"`
	PipelineID   int               `json:"pipelineid"`
	StartDate    time.Time         `json:"startdate,omitempty"`
	FinishDate   time.Time         `json:"finishdate,omitempty"`
	ScheduleDate time.Time         `json:"scheduledate,omitempty"`
	Status       PipelineRunStatus `json:"status,omitempty"`
	Jobs         []Job             `json:"jobs,omitempty"`
}

// Cfg represents the global config instance
var Cfg = &Config{}

// Config holds all config options
type Config struct {
	DevMode           bool
	VersionSwitch     bool
	Poll              bool
	PVal              int
	ListenPort        string
	HomePath          string
	Hostname          string
	VaultPath         string
	DataPath          string
	PipelinePath      string
	WorkspacePath     string
	Worker            string
	JwtPrivateKeyPath string
	JWTKey            interface{}
	Logger            hclog.Logger
	CAPath            string

	Bolt struct {
		Mode os.FileMode
	}
}

// String returns a pipeline type string back
func (p PipelineType) String() string {
	return string(p)
}
