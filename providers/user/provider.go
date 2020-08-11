package user

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/gaia-pipeline/gaia/store"
)

// Provider represents the user handlers and contains any dependencies required by the handlers.
type Provider struct {
	Store   store.GaiaStore
	RBACSvc rbac.Service
}

// NewProvider creates a new provider.
func NewProvider(store store.GaiaStore, RBACSvc rbac.Service) *Provider {
	return &Provider{Store: store, RBACSvc: RBACSvc}
}

// UserLogin authenticates the user with the given credentials.
// @Summary User Login
// @Description Returns an authenticated user.
// @Tags users
// @Accept json
// @Produce json
// @Param UserLoginRequest body gaia.User true "UserLogin request"
// @Success 200 {object} gaia.User
// @Failure 400 {string} string "error reading json"
// @Failure 403 {string} string "credentials provided"
// @Failure 500 {string} string "{creating jwt token|signing jwt token}"
// @Router /login [post]
func (h *Provider) UserLogin(c echo.Context) error {
	u := &gaia.User{}
	if err := c.Bind(u); err != nil {
		gaia.Cfg.Logger.Debug("error reading json during UserLogin", "error", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Authenticate user
	user, err := h.Store.UserAuth(u, true)
	if err != nil || user == nil {
		gaia.Cfg.Logger.Info("invalid credentials provided", "username", u.Username)
		return c.String(http.StatusForbidden, "invalid username and/or password")
	}

	perms, err := h.Store.UserPermissionsGet(u.Username)
	if err != nil {
		return err
	}

	// Setup custom claims
	claims := gaia.JwtCustomClaims{
		Username: user.Username,
		Roles:    perms.Roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	var token *jwt.Token
	// Generate JWT token
	switch t := gaia.Cfg.JWTKey.(type) {
	case []byte:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	case *rsa.PrivateKey:
		token = jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	default:
		gaia.Cfg.Logger.Error("invalid jwt key type", "type", t)
		return c.String(http.StatusInternalServerError, "error creating jwt token: invalid jwt key type")
	}

	// Sign and get encoded token
	tokenstring, err := token.SignedString(gaia.Cfg.JWTKey)
	if err != nil {
		gaia.Cfg.Logger.Error("error signing jwt token", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	user.JwtExpiry = claims.ExpiresAt
	user.Tokenstring = tokenstring

	// Return JWT token and display name
	return c.JSON(http.StatusOK, user)
}

// UserGetAll returns all users stored in store.
// @Summary Get all users
// @Description Returns a list of registered users.
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} gaia.User
// @Security ApiKeyAuth
// @Failure 500 {string} string Unable to connect to database.
// @Router /users [get]
func (h *Provider) UserGetAll(c echo.Context) error {
	// Get all users
	users, err := h.Store.UserGetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, users)
}

type changePasswordRequest struct {
	OldPassword     string `json:"oldpassword"`
	NewPassword     string `json:"newpassword"`
	NewPasswordConf string `json:"newpasswordconf"`
	Username        string `json:"username"`
}

// UserChangePassword changes the password from a user.
// @Summary Change password for user.
// @Description Changes the password of the given user.
// @Tags users
// @Accept json
// @Produce plain
// @Param UserChangePasswordRequest body changePasswordRequest true "UserChangePassword request"
// @Success 200 {string} string Password has been changed
// @Failure 400 {string} string "{Invalid parameters given for password change request|Cannot find user with the given username|New password does not match new password confirmation}"
// @Failure 412 {string} string Wrong password given for password change
// @Failure 500 {string} string Cannot update user in store
// @Router /user/password [post]
func (h *Provider) UserChangePassword(c echo.Context) error {
	// Get required parameters
	r := &changePasswordRequest{}
	if err := c.Bind(r); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for password change request")
	}

	// Compare old password with current password of user by simply calling auth method.
	// First get user obj
	user, err := h.Store.UserGet(r.Username)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cannot find user with the given username")
	}

	// Simply call auth by changing password
	user.Password = r.OldPassword
	u, err := h.Store.UserAuth(user, false)
	if err != nil {
		return c.String(http.StatusPreconditionFailed, "Wrong password given for password change")
	}

	// Compare new password with new password confirmation
	if r.NewPassword != r.NewPasswordConf {
		return c.String(http.StatusBadRequest, "New password does not match new password confirmation")
	}

	// Change password
	u.Password = r.NewPassword
	if err := h.Store.UserPut(u, true); err != nil {
		return c.String(http.StatusInternalServerError, "Cannot update user in store")
	}

	return c.String(http.StatusOK, "Password has been changed")
}

// UserResetTriggerToken will generate and save a new Remote trigger token
// for a given user.
// @Summary Generate new remote trigger token.
// @Description Generates and saves a new remote trigger token for a given user.
// @Tags users
// @Accept plain
// @Produce plain
// @Param username query string true "The username to reset the token for"
// @Success 200 {string} string Token reset
// @Failure 400 {string} string "{Invalid username given|Only auto user can have a token reset|User not found|Error retrieving user}"
// @Failure 500 {string} string Error while saving user
// @Router /user/{username}/reset-trigger-token [put]
func (h *Provider) UserResetTriggerToken(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Invalid username given")
	}
	if username != "auto" {
		return c.String(http.StatusBadRequest, "Only auto user can have a token reset")
	}

	user, err := h.Store.UserGet(username)
	if err != nil {
		return c.String(http.StatusBadRequest, "User not found")
	}
	if user == nil {
		return c.String(http.StatusBadRequest, "Error retrieving user")
	}

	user.TriggerToken = security.GenerateRandomUUIDV5()
	err = h.Store.UserPut(user, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error while saving user")
	}

	return c.String(http.StatusOK, "Token reset")
}

// UserDelete deletes the given user
// @Summary Delete user.
// @Description Delete a given user.
// @Tags users
// @Accept plain
// @Produce plain
// @Param username query string true "The username to delete"
// @Success 200 {string} string User has been deleted
// @Failure 400 {string} string "{Invalid username given|Auto user cannot be deleted}"
// @Failure 404 {string} string "{User not found|Permission not found|Rbac not found}"
// @Router /user/{username}/delete [delete]
func (h *Provider) UserDelete(c echo.Context) error {
	username := c.Param("username")

	if username == "" {
		return c.String(http.StatusBadRequest, "Invalid username given")
	}

	if username == "auto" {
		return c.String(http.StatusBadRequest, "Auto user cannot be deleted")
	}

	if err := h.Store.UserDelete(username); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	if err := h.Store.UserPermissionsDelete(username); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	if err := h.RBACSvc.DeleteUser(username); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.String(http.StatusOK, "User has been deleted")
}

// UserAdd adds a new user to the store.
// @Summary Add user.
// @Description Adds a new user.
// @Tags users
// @Accept json
// @Produce plain
// @Param UserAddRequest body gaia.User true "UserAdd request"
// @Success 200 {string} string "User has been added"
// @Failure 400 {string} string "Invalid parameters given for add user request"
// @Failure 500 {string} string "{User put failed|User permission put error}"
// @Router /user [post]
func (h *Provider) UserAdd(c echo.Context) error {
	// Get user information required for add
	u := &gaia.User{}
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for add user request")
	}

	// Add user
	u.LastLogin = time.Now()
	err := h.Store.UserPut(u, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Add default perms
	perms := &gaia.UserPermission{
		Username: u.Username,
		Roles:    rolehelper.FlattenUserCategoryRoles(rolehelper.DefaultUserRoles),
		Groups:   []string{},
	}
	err = h.Store.UserPermissionsPut(perms)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusCreated, "User has been added")
}

// UserGetPermissions returns the permissions for a user.
// @Summary Get permission of the user.
// @Description Get permissions of the user.
// @Tags users
// @Accept plain
// @Produce json
// @Param username query string true "The username to get permission for"
// @Success 200 {object} gaia.UserPermission
// @Failure 400 {string} string "Failed to get permission"
// @Router /user/{username}/permissions [get]
func (h *Provider) UserGetPermissions(c echo.Context) error {
	u := c.Param("username")

	perms, err := h.Store.UserPermissionsGet(u)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, perms)
}

// UserPutPermissions adds or updates permissions for a user.
// @Summary Adds or updates permissions for a user.
// @Description Adds or updates permissions for a user..
// @Tags users
// @Accept json
// @Produce plain
// @Param username query string true "The username to get permission for"
// @Success 200 {string} string "Permissions have been updated"
// @Failure 400 {string} string "{Invalid parameters given for request|Permissions put failed}"
// @Router /user/{username}/permissions [put]
func (h *Provider) UserPutPermissions(c echo.Context) error {
	var perms *gaia.UserPermission
	if err := c.Bind(&perms); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for request")
	}

	if err := h.Store.UserPermissionsPut(perms); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Permissions have been updated")
}
