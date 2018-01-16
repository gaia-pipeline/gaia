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
