package gaia

import (
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
)

// PipelineType represents supported plugin types
type PipelineType string

const (
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
	Name         string       `json:"pipelinename"`
	Repo         GitRepo      `json:"gitrepo"`
	Type         PipelineType `json:"pipelinetype"`
	Status       int          `json:"status"`
	ErrMsg       string       `json:"errmsg"`
	CreationDate time.Time    `json:"creationdate"`
}

// GitRepo represents a single git repository
type GitRepo struct {
	URL            string     `json:"giturl"`
	Username       string     `json:"gituser"`
	Password       string     `json:"gitpassword"`
	PrivateKey     PrivateKey `json:"privatekey"`
	SelectedBranch string     `json:"selectedbranch"`
	Branches       []string   `json:"gitbranches"`
	LocalDest      string
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
	ListenPort string
	HomePath   string
	Logger     hclog.Logger

	Bolt struct {
		Path string
		Mode os.FileMode
	}
}
