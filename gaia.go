package gaia

import (
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
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

// Pipeline represents a single pipeline
type Pipeline struct {
	ID        int          `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Repo      GitRepo      `json:"repo,omitempty"`
	Type      PipelineType `json:"type,omitempty"`
	ExecPath  string       `json:"execpath,omitempty"`
	SHA256Sum []byte       `json:"sha256sum,omitempty"`
	Jobs      []Job        `json:"jobs,omitempty"`
	Created   time.Time    `json:"created,omitempty"`
	UUID      string       `json:"uuid,omitempty"`
}

// GitRepo represents a single git repository
type GitRepo struct {
	URL            string     `json:"url,omitempty"`
	Username       string     `json:"user,omitempty"`
	Password       string     `json:"password,omitempty"`
	PrivateKey     PrivateKey `json:"privatekey,omitempty"`
	SelectedBranch string     `json:"selectedbranch,omitempty"`
	Branches       []string   `json:"branches,omitempty"`
	LocalDest      string
}

// Job represents a single job of a pipeline
type Job struct {
	ID          uint32    `json:"id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"desc,omitempty"`
	Priority    int64     `json:"priority"`
	Status      JobStatus `json:"status,omitempty"`
}

// CreatePipeline represents a pipeline which is not yet
// compiled.
type CreatePipeline struct {
	ID         string             `json:"id,omitempty"`
	Pipeline   Pipeline           `json:"pipeline,omitempty"`
	Status     int                `json:"status,omitempty"`
	StatusType CreatePipelineType `json:"statustype,omitempty"`
	Output     string             `json:"output,omitempty"`
	Created    time.Time          `json:"created,omitempty"`
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
var Cfg *Config

// Config holds all config options
type Config struct {
	DevMode           bool
	VersionSwitch     bool
	Poll              bool
	PVal              int
	ListenPort        string
	HomePath          string
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
