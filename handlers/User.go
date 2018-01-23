package handlers

import (
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
		cfg.Logger.Debug("error reading json during UserLogin", "error", err.Error())
		return
	}

	// Authenticate user
	user, err := storeService.UserAuth(u)
	if err != nil {
		cfg.Logger.Error("error during UserAuth", "error", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString("invalid username and/or password")
		return
	}

	// Setup custom claims
	claims := jwtCustomClaims{
		user.Username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get encoded token
	tokenstring, err := token.SignedString(jwtKey)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		cfg.Logger.Error("error signing jwt token", "error", err.Error())
		return
	}
	user.JwtExpiry = claims.ExpiresAt
	user.Tokenstring = tokenstring

	// Return JWT token and display name
	ctx.JSON(user)
}
