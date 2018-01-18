package gaia

import (
	"os"

	hclog "github.com/hashicorp/go-hclog"
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
	Name string  `json:"pipelinename"`
	Repo GitRepo `json:"gitrepo"`
}

// GitRepo represents a single git repository
type GitRepo struct {
	URL        string   `json:"giturl"`
	Username   string   `json:"gituser"`
	Password   string   `json:"gitpassword"`
	PrivateKey string   `json:"gitkey"`
	Branches   []string `json:"branches"`
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
