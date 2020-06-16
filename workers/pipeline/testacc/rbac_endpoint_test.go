package pipeline

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"gotest.tools/assert"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/server"
)

type jwtCustomClaims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.StandardClaims
}

func TestRBACEndpointAcceptanceTestTearUp(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestRBACEndpointAcceptanceTestTearUp")
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.RBACEnabled = true
	gaia.Cfg.JwtPrivateKeyPath = "test.pem"
	gaia.Cfg.Mode = gaia.ModeServer

	claims := jwtCustomClaims{
		Username: "no-roles",
		Roles:    rolehelper.FlattenUserCategoryRoles(rolehelper.DefaultUserRoles),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	file, _ := ioutil.ReadFile("test.pem")
	pk, err := jwt.ParseRSAPrivateKeyFromPEM(file)
	assert.NilError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	tokenstring, err := token.SignedString(pk)
	assert.NilError(t, err)

	defer func() {
		os.RemoveAll(tmp)
	}()

	// Start the server as background process.
	go func() {
		err := server.Start()
		assert.NilError(t, err)
	}()

	// Sleep a bit until all components are initialized and started.
	time.Sleep(2 * time.Second)

	tests := []struct {
		method        string
		endpoint      string
		body          []byte
		permissionErr string
	}{
		{
			method:        http.MethodGet,
			endpoint:      "users",
			body:          nil,
			permissionErr: "users/list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			client := http.Client{}

			req, err := http.NewRequest(tt.method, fmt.Sprintf("http://localhost:8080/api/v1/%s", tt.endpoint), nil)
			assert.NilError(t, err)
			req.Header.Add("Authorization", "Bearer "+tokenstring)

			res, err := client.Do(req)
			assert.NilError(t, err)

			resBody, err := ioutil.ReadAll(res.Body)
			assert.NilError(t, err)

			assert.Equal(t, res.StatusCode, http.StatusForbidden)
			assert.Equal(t, string(resBody), "Permission denied. Must have "+tt.permissionErr)
		})
	}
}
