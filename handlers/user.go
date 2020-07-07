package handlers

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/gaia-pipeline/gaia/store"
)

// jwtExpiry defines how long the produced jwt tokens
// are valid. By default 12 hours.
const jwtExpiry = 12 * 60 * 60

type jwtCustomClaims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.StandardClaims
}

// UserHandler represents the user handlers and contains any dependencies required by the handlers.
type UserHandler struct {
	Store   store.GaiaStore
	RBACSvc rbac.Service
}

// NewUserHandler creates a new UserHandlers.
func NewUserHandler(store store.GaiaStore, RBACSvc rbac.Service) *UserHandler {
	return &UserHandler{Store: store, RBACSvc: RBACSvc}
}

// UserLogin authenticates the user with
// the given credentials.
func (h *UserHandler) UserLogin(c echo.Context) error {
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
	claims := jwtCustomClaims{
		Username: user.Username,
		Roles:    perms.Roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
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
func (h *UserHandler) UserGetAll(c echo.Context) error {
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
func (h *UserHandler) UserChangePassword(c echo.Context) error {
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
func (h *UserHandler) UserResetTriggerToken(c echo.Context) error {
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
func (h *UserHandler) UserDelete(c echo.Context) error {
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
func (h *UserHandler) UserAdd(c echo.Context) error {
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
func (h *UserHandler) UserGetPermissions(c echo.Context) error {
	u := c.Param("username")

	perms, err := h.Store.UserPermissionsGet(u)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, perms)
}

// UserPutPermissions adds or updates permissions for a user.
func (h *UserHandler) UserPutPermissions(c echo.Context) error {
	var perms *gaia.UserPermission
	if err := c.Bind(&perms); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for request")
	}

	if err := h.Store.UserPermissionsPut(perms); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Permissions have been updated")
}
