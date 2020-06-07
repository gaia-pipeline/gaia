package handlers

import (
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/labstack/echo"
	"net/http"
)

type RBACHandler struct {
	svc rbac.EnforcerService
}

func (h *RBACHandler) AddRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	var newRoles []rbac.RoleRule
	if err := c.Bind(&newRoles); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse rules.")
	}

	if err := h.svc.AddRole(role, newRoles); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Role created successfully.")
}

func (h *RBACHandler) DeleteRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	if err := h.svc.DeleteRole(role); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Role deleted successfully.")
}

func (h *RBACHandler) GetAllRoles(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.GetAllRoles())
}

func (h *RBACHandler) GetUserAttachedRoles(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	roles, err := h.svc.GetUserAttachedRoles(username)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) GetRolesAttachedUsers(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide Role.")
	}

	roles, err := h.svc.GetRoleAttachedUsers(role)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) AttachRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	if err := h.svc.AttachRole(username, role); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Role attached successfully.")
}

func (h *RBACHandler) DetatchRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	if err := h.svc.DetachRole(username, role); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Role detached successfully.")
}
