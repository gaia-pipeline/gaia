package handlers

import (
	"crypto/rand"

	"github.com/gaia-pipeline/gaia/store"
	"github.com/kataras/iris"
)

const (
	apiVersion = "v1"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService *store.Store

// jwtKey is a random generated key for jwt signing
var jwtKey []byte

// InitHandlers initializes(registers) all handlers
func InitHandlers(i *iris.Application, s *store.Store) error {
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

	// --- Register handlers at iris instance ---

	// Users
	i.Post(p+"users/login", UserLogin)

	// Pipelines
	i.Post(p+"pipelines/gitlsremote", PipelineGitLSRemote)
	i.Post(p+"pipelines/create", CreatePipeline)
	i.Get(p+"pipelines/create", CreatePipelineGetAll)
	i.Post(p+"pipelines/name", PipelineNameAvailable)
	i.Get(p+"pipelines", PipelineGetAll)
	i.Get(p+"pipelines/detail/{id:string}", PipelineGet)
	i.Get(p+"pipelines/start/{id:string}", PipelineStart)

	return nil
}
