package handlers

import (
	"github.com/kataras/iris"
	"github.com/michelvocks/gaia/store"
)

const (
	apiVersion = "v1"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService *store.Store

// InitHandlers initializes(registers) all handlers
func InitHandlers(i *iris.Application, s *store.Store) {
	// Set store instance
	storeService = s

	// Define prefix
	p := "/api/" + apiVersion + "/"

	i.Post(p+"users/login", UserLogin)
}
