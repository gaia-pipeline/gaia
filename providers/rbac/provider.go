package rbac

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security/rbac"
)

type Provider struct {
	svc rbac.Service
}

func NewProvider(svc rbac.Service) *Provider {
	return &Provider{svc: svc}
}

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

func (h *Provider) GetAllRoles(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.GetAllRoles())
}

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

func (h *Provider) GetRolesAttachedUsers(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	roles, err := h.svc.GetRoleAttachedUsers(role)
	if err != nil {
		gaia.Cfg.Logger.Error("error roles attached to user", "role", role, "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while getting the users.")
	}

	return c.JSON(http.StatusOK, roles)
}

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
