package gaia

import (
	"os"
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

	Bolt struct {
		Path string
		Mode os.FileMode
	}
}
