package handlers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia/security/rbac"
)

type rbacHandler struct {
	svc rbac.Service
}

func (h *rbacHandler) addRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	var newRoles []rbac.RoleRule
	if err := c.Bind(&newRoles); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body provided.")
	}

	if err := h.svc.AddRole(role, newRoles); err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while adding the role.")
	}

	return c.String(http.StatusOK, "Role created successfully.")
}

func (h *rbacHandler) deleteRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	if err := h.svc.DeleteRole(role); err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while deleting the role.")
	}

	return c.String(http.StatusOK, "Role deleted successfully.")
}

func (h *rbacHandler) getAllRoles(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.GetAllRoles())
}

func (h *rbacHandler) getUserAttachedRoles(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	roles, err := h.svc.GetUserAttachedRoles(username)
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while getting the roles.")
	}

	return c.JSON(http.StatusOK, roles)
}

func (h *rbacHandler) getRolesAttachedUsers(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	roles, err := h.svc.GetRoleAttachedUsers(role)
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while getting the users.")
	}

	return c.JSON(http.StatusOK, roles)
}

func (h *rbacHandler) attachRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	if err := h.svc.AttachRole(username, role); err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while attaching the role.")
	}

	return c.String(http.StatusOK, "Role attached successfully.")
}

func (h *rbacHandler) detatchRole(c echo.Context) error {
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "Must provide role.")
	}

	username := c.Param("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Must provide username.")
	}

	if err := h.svc.DetachRole(username, role); err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while detaching the role.")
	}

	return c.String(http.StatusOK, "Role detached successfully.")
}
