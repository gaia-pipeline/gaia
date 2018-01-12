package handlers

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
	"github.com/michelvocks/gaia"
)

// jwtExpiry defines how long the produced jwt tokens
// are valid. By default 12 hours.
const jwtExpiry = (12 * 60 * 60)

type jwtCustomClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

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

	// Setup custom claims
	claims := jwtCustomClaims{
		u.Username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get encoded token
	b := []byte{'f', '2', 'f', 'f', 's', 'h', 's'}
	tokenstring, err := token.SignedString(b)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("Error during signing jwt token!")
		fmt.Printf("Error signing jwt token: %s", err.Error())
		return
	}
	u.JwtExpiry = claims.ExpiresAt
	u.Tokenstring = tokenstring
	u.DisplayName = "Michel Vocks"

	// Return JWT token and display name
	ctx.JSON(u)
}
