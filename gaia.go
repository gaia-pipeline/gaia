package gaia

import (
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/robfig/cron"
)

// PipelineType represents supported plugin types
type PipelineType string

// CreatePipelineType represents the different status types
// a create pipeline can have.
type CreatePipelineType string

// PipelineRunStatus represents the different status a run
// can have.
type PipelineRunStatus string

// JobStatus represents the different status a job can have
type JobStatus string

// GaiaMode represents the different modes for Gaia
type GaiaMode string

// WorkerStatus represents the different status a worker can have
type WorkerStatus string

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

	// PTypeRuby ruby plugin type
	PTypeRuby PipelineType = "ruby"

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

	// ModeServer mode
	ModeServer GaiaMode = "server"

	// ModeWorker mode
	ModeWorker GaiaMode = "worker"

	// WorkerActive status
	WorkerActive WorkerStatus = "active"

	// WorkerInactive status
	WorkerInactive WorkerStatus = "inactive"

	// WorkerSuspended status
	WorkerSuspended WorkerStatus = "suspended"

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

	// TmpRubyFolder is the name of the ruby temporary folder
	TmpRubyFolder = "ruby"

	// WorkerRegisterKey is the used key for worker registration secret
	WorkerRegisterKey = "WORKER_REGISTER_KEY"
)

// User is the user object
type User struct {
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	DisplayName  string    `json:"display_name,omitempty"`
	Tokenstring  string    `json:"tokenstring,omitempty"`
	JwtExpiry    int64     `json:"jwtexpiry,omitempty"`
	LastLogin    time.Time `json:"lastlogin,omitempty"`
	TriggerToken string    `json:"trigger_token,omitempty"`
}

// UserPermission is stored in its own data structure away from the core user. It represents all permission data
// for a single user.
type UserPermission struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Groups   []string `json:"groups"`
}

// UserRoleCategory represents the top-level of the permission role system
type UserRoleCategory struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Roles       []*UserRole `json:"roles"`
}

// UserRole represents a single permission role.
type UserRole struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	APIEndpoint []*UserRoleEndpoint `json:"api_endpoints"`
}

// UserRoleEndpoint represents the path and method of the API endpoint to be secured.
type UserRoleEndpoint struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// Pipeline represents a single pipeline
type Pipeline struct {
	ID                int          `json:"id,omitempty"`
	Name              string       `json:"name,omitempty"`
	Repo              *GitRepo     `json:"repo,omitempty"`
	Type              PipelineType `json:"type,omitempty"`
	ExecPath          string       `json:"execpath,omitempty"`
	SHA256Sum         []byte       `json:"sha256sum,omitempty"`
	Jobs              []*Job       `json:"jobs,omitempty"`
	Created           time.Time    `json:"created,omitempty"`
	UUID              string       `json:"uuid,omitempty"`
	IsNotValid        bool         `json:"notvalid,omitempty"`
	PeriodicSchedules []string     `json:"periodicschedules,omitempty"`
	TriggerToken      string       `json:"trigger_token,omitempty"`
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
	ID           uint32      `json:"id,omitempty"`
	Title        string      `json:"title,omitempty"`
	Description  string      `json:"desc,omitempty"`
	DependsOn    []*Job      `json:"dependson,omitempty"`
	Status       JobStatus   `json:"status,omitempty"`
	Args         []*Argument `json:"args,omitempty"`
	FailPipeline bool        `json:"failpipeline,omitempty"`
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
	Jobs         []*Job            `json:"jobs,omitempty"`
}

// Worker represents a single registered worker.
type Worker struct {
	UniqueID     string            `json:"uniqueid"`
	Name         string            `json:"name"`
	Status       WorkerStatus      `json:"status"`
	RegisterDate time.Time         `json:"registerdate"`
	LastContact  time.Time         `json:"lastcontact"`
	Tags         map[string]string `json:"tags"`
}

// Cfg represents the global config instance
var Cfg = &Config{}

// Config holds all config options
type Config struct {
	DevMode           bool
	ModeRaw           string
	Mode              GaiaMode
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
	Worker            int
	JwtPrivateKeyPath string
	JWTKey            interface{}
	Logger            hclog.Logger
	CAPath            string
	WorkerServerPort  string

	// Worker
	WorkerName        string
	WorkerHostURL     string
	WorkerGRPCHostURL string
	WorkerSecret      string

	Bolt struct {
		Mode os.FileMode
	}
}

// StoreConfig defines config settings to be stored in DB.
type StoreConfig struct {
	ID   int
	Poll bool
}

// String returns a pipeline type string back
func (p PipelineType) String() string {
	return string(p)
}
