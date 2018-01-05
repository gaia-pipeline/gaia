package handlers

import (
	"github.com/kataras/iris"
)

const (
	apiVersion = "v1"
)

// InitHandlers initializes(registers) all handlers
func InitHandlers(i *iris.Application) {
	// Define prefix
	p := "/api/" + apiVersion + "/"

	i.Post(p+"users/login", UserLogin)
}
