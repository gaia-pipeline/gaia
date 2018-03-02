package gaia

import (
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
)

// PipelineType represents supported plugin types
type PipelineType string

// PipelineRunStatus represents the different status a run
// can have.
type PipelineRunStatus string

// JobStatus represents the different status a job can have.
type JobStatus string

const (
	// UNKNOWN plugin type
	UNKNOWN PipelineType = "unknown"

	// GOLANG plugin type
	GOLANG PipelineType = "golang"

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
)

// User is the user object
type User struct {
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Tokenstring string `json:"tokenstring,omitempty"`
	JwtExpiry   int64  `json:"jwtexpiry,omitempty"`
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
	ID       string    `json:"id,omitempty"`
	Pipeline Pipeline  `json:"pipeline,omitempty"`
	Status   int       `json:"status,omitempty"`
	Output   string    `json:"errmsg,omitempty"`
	Created  time.Time `json:"created,omitempty"`
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
	ListenPort   string
	HomePath     string
	DataPath     string
	PipelinePath string
	Workers      int
	Logger       hclog.Logger

	Bolt struct {
		Path string
		Mode os.FileMode
	}

	TestMode bool
}

// String returns a pipeline type string back
func (p PipelineType) String() string {
	return string(p)
}
