package handlers

import (
	"crypto/rand"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/kataras/iris"
)

const (
	apiVersion = "v1"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService *store.Store

// cfg is a pointer to the global config
var cfg *gaia.Config

// jwtKey is a random generated key for jwt signing
var jwtKey []byte

// InitHandlers initializes(registers) all handlers
func InitHandlers(c *gaia.Config, i *iris.Application, s *store.Store) error {
	// Set config
	cfg = c

	// Set store instance
	storeService = s

	// Generate signing key for jwt
	jwtKey = make([]byte, 64)
	_, err := rand.Read(jwtKey)
	if err != nil {
		return err
	}

	// Define prefix
	p := "/api/" + apiVersion + "/"

	// Register handlers at iris instance
	i.Post(p+"users/login", UserLogin)
	i.Post(p+"pipelines/gitlsremote", PipelineGitLSRemote)
	i.Post(p+"pipelines/create", PipelineBuildFromSource)
	i.Post(p+"pipelines/nameavailable", PipelineNameAvailable)

	return nil
}
