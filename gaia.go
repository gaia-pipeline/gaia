package gaia

import (
	"os"

	hclog "github.com/hashicorp/go-hclog"
)

// PluginType represents supported plugin types
type PluginType int

const (
	// GOLANG plugin type
	GOLANG PluginType = iota
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
	Name string     `json:"pipelinename"`
	Repo GitRepo    `json:"gitrepo"`
	Type PluginType `json:"plugintype"`
}

// GitRepo represents a single git repository
type GitRepo struct {
	URL        string     `json:"giturl"`
	Username   string     `json:"gituser"`
	Password   string     `json:"gitpassword"`
	PrivateKey PrivateKey `json:"privatekey"`
	Branches   []string   `json:"gitbranches"`
	LocalDest  string
}

// PrivateKey represents a pem encoded private key
type PrivateKey struct {
	Key      string `json:"key"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config holds all config options
type Config struct {
	ListenPort string
	DataPath   string
	Logger     hclog.Logger

	Bolt struct {
		Path string
		Mode os.FileMode
	}
}
