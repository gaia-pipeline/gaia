package handlers

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
	"github.com/michelvocks/gaia"
)

// UserLogin authenticates the user with
// the given credentials.
func UserLogin(ctx iris.Context) {
	u := &gaia.User{}
	if err := ctx.ReadJSON(u); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Authenticate user
	// TODO

	// Remove password from object
	u.Password = ""

	// Wrap User object in JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
	})

	// Sign and get encoded token
	b := []byte{'f', '2', 'f', 'f', 's', 'h', 's'}
	tokenstring, err := token.SignedString(b)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("Error during signing jwt token!")
		fmt.Printf("Error signing jwt token: %s", err.Error())
		return
	}
	u.Tokenstring = tokenstring
	u.DisplayName = "Michel Vocks"

	// Return JWT token and display name
	ctx.JSON(u)

	fmt.Println("Token returned!")
	fmt.Printf("User obj: %+v\n", u)
}
