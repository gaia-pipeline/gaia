package rbac

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security/rbac"
)

// Provider represents the RBAC provider.
type Provider struct {
	svc rbac.Service
}

// NewProvider creates a new Provider.
func NewProvider(svc rbac.Service) *Provider {
	return &Provider{svc: svc}
}

// AddRole adds an RBAC role using the RBAC service.
// @Summary Adds an RBAC role.
// @Description Adds an RBAC role using the RBAC service.
// @Tags rbac
// @Accept plain
// @Produce plain
// @Param role query string true "Name of the role"
// @Success 200 {string} string "Role created successfully."
// @Failure 400 {string} string "Must provide role."
// @Failure 500 {string} string "An error occurred while adding the role."
// @Router /rbac/roles/{role} [put]
func (h *Provider) AddRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	var newRoles []rbac.RoleRule
	if err := c.Bind(&newRoles); err != nil {
		gaia.Cfg.Logger.Error("error parsing role body", "role", role, "error", err.Error())
		return c.String(http.StatusBadRequest, "Invalid body provided.")
	}

	if err := h.svc.AddRole(role, newRoles); err != nil {
		gaia.Cfg.Logger.Error("error adding role", "role", role, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while adding the role.")
	}

	return c.String(http.StatusOK, "Role created successfully.")
}

// DeleteRole deletes an RBAC role using the RBAC service.
// @Summary Delete an RBAC role.
// @Description Deletes an RBAC role using the RBAC service.
// @Tags rbac
// @Accept plain
// @Produce plain
// @Param role query string true "The name of the rule"
// @Success 200 {string} string "Role deleted successfully."
// @Failure 400 {string} string "Must provide role."
// @Failure 500 {string} string "An error occurred while deleting the role."
// @Router /rbac/roles/{role} [delete]
func (h *Provider) DeleteRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	if err := h.svc.DeleteRole(role); err != nil {
		gaia.Cfg.Logger.Error("error deleting role", "role", role, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while deleting the role.")
	}

	return c.String(http.StatusOK, "Role deleted successfully.")
}

// GetAllRoles gets all RBAC roles.
// @Summary Gets all RBAC roles.
// @Description Gets all RBAC roles.
// @Tags rbac
// @Produce plain
// @Success 200 {array} string "All the roles."
// @Router /rbac/roles [get]
func (h *Provider) GetAllRoles(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.GetAllRoles())
}

// GetUserAttachedRoles gets all roles for a user.
// @Summary Gets all roles for a user.
// @Description Gets all roles for a user.
// @Tags rbac
// @Accept plain
// @Produce json
// @Param username query string true "The username of the user"
// @Success 200 {array} string "Attached roles to a user"
// @Failure 400 {string} string "Must provide username."
// @Failure 500 {string} string "An error occurred while getting the roles."
// @Router /users/{username}/rbac/roles [get]
func (h *Provider) GetUserAttachedRoles(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	roles, err := h.svc.GetUserAttachedRoles(username)
	if err != nil {
		gaia.Cfg.Logger.Error("error getting user attached roles", "username", username, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while getting the roles.")
	}

	return c.JSON(http.StatusOK, roles)
}

// GetRoleAttachedUsers gets a user attached to a role.
// @Summary Gets a user attached to a role.
// @Description Gets a user attached to a role.
// @Tags rbac
// @Accept plain
// @Produce json
// @Param role query string true "The role for the user"
// @Success 200 {array} string "Attached users for the role"
// @Failure 400 {string} string "Must provide role."
// @Failure 500 {string} string "An error occurred while getting the user."
// @Router /rbac/roles/{role}/attached [get]
func (h *Provider) GetRoleAttachedUsers(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	users, err := h.svc.GetRoleAttachedUsers(role)
	if err != nil {
		gaia.Cfg.Logger.Error("error users attached to user", "role", role, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while getting the users.")
	}

	return c.JSON(http.StatusOK, users)
}

// AttachRole attaches a role to a user.
// @Summary Attach role to user.
// @Description Attach role to user.
// @Tags rbac
// @Accept plain
// @Produce plain
// @Param role query string true "The role"
// @Param username query string true "The username of the user"
// @Success 200 {string} string "Role attached successfully."
// @Failure 400 {string} string "Must provide role or username."
// @Failure 500 {string} string "An error occurred while attaching the role."
// @Router /rbac/roles/{role}/attach/{username} [put]
func (h *Provider) AttachRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	if err := h.svc.AttachRole(username, role); err != nil {
		gaia.Cfg.Logger.Error("error attaching role", "role", role, "username", username, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while attaching the role.")
	}

	return c.String(http.StatusOK, "Role attached successfully.")
}

// DetachRole detaches a role from a user.
// @Summary Detach role to user.
// @Description Detach role to user.
// @Tags rbac
// @Accept plain
// @Produce plain
// @Param role query string true "The role"
// @Param username query string true "The username of the user"
// @Success 200 {string} string "Role detached successfully."
// @Failure 400 {string} string "Must provide role or username."
// @Failure 500 {string} string "An error occurred while detaching the role."
// @Router /rbac/roles/{role}/attach/{username} [delete]
func (h *Provider) DetachRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	if err := h.svc.DetachRole(username, role); err != nil {
		gaia.Cfg.Logger.Error("error detaching role", "role", role, "username", username, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while detaching the role.")
	}

	return c.String(http.StatusOK, "Role detached successfully.")
}
