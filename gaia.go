package gaia

import (
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
)

// PipelineType represents supported plugin types
type PipelineType string

const (
	// UNKNOWN plugin type
	UNKNOWN PipelineType = "unknown"

	// GOLANG plugin type
	GOLANG PipelineType = "golang"
)

// User is the user object
type User struct {
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"`
	DisplayName string `json:"display_name"`
	Tokenstring string `json:"tokenstring"`
	JwtExpiry   int64  `json:"jwtexpiry"`
}

// Pipeline represents a single pipeline
type Pipeline struct {
	Name    string       `json:"name"`
	Repo    GitRepo      `json:"repo"`
	Type    PipelineType `json:"type"`
	Created time.Time    `json:"created"`
}

// GitRepo represents a single git repository
type GitRepo struct {
	URL            string     `json:"url"`
	Username       string     `json:"user"`
	Password       string     `json:"password"`
	PrivateKey     PrivateKey `json:"privatekey"`
	SelectedBranch string     `json:"selectedbranch"`
	Branches       []string   `json:"branches"`
	LocalDest      string
}

// CreatePipeline represents a pipeline which is not yet
// compiled.
type CreatePipeline struct {
	ID       string    `json:"id"`
	Pipeline Pipeline  `json:"pipeline"`
	Status   int       `json:"status"`
	Output   string    `json:"errmsg"`
	Created  time.Time `json:"created"`
}

// PrivateKey represents a pem encoded private key
type PrivateKey struct {
	Key      string `json:"key"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Cfg represents the global config instance
var Cfg *Config

// Config holds all config options
type Config struct {
	ListenPort   string
	HomePath     string
	DataPath     string
	PipelinePath string
	Logger       hclog.Logger

	Bolt struct {
		Path string
		Mode os.FileMode
	}
}

// String returns a pipeline type string back
func (p PipelineType) String() string {
	return string(p)
}
