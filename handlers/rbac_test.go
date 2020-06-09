package handlers

import (
	"bytes"
	"errors"
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/labstack/echo"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRBACSvc struct {
	rbac.Service
}

func (e *mockRBACSvc) GetAllRoles() []string {
	return []string{"role-a", "role-b"}
}

func (e *mockRBACSvc) AddRole(role string, roleRules []rbac.RoleRule) error {
	if role == "success" {
		return nil
	}
	return errors.New("add error")
}

func (e *mockRBACSvc) DeleteRole(role string) error {
	if role == "delme" {
		return nil
	}
	return errors.New("delete error")
}

func (e *mockRBACSvc) GetUserAttachedRoles(username string) ([]string, error) {
	if username == "test" {
		return []string{"role-a", "role-b"}, nil
	}
	return nil, errors.New("an error")
}

func (e *mockRBACSvc) GetRoleAttachedUsers(role string) ([]string, error) {
	if role == "test" {
		return []string{"user-a", "user-b"}, nil
	}
	return nil, errors.New("an error")
}

func (e *mockRBACSvc) AttachRole(username string, role string) error {
	if role == "test-role" && username == "test-user" {
		return nil
	}
	return errors.New("an error")
}

func (e *mockRBACSvc) DetachRole(username string, role string) error {
	if role == "test-role" && username == "test-user" {
		return nil
	}
	return errors.New("an error")
}

func Test_rbacHandler_addRole(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	t.Run("success (200) if add is successful", func(t *testing.T) {
		body := `[
	{
		"namespace": "secrets",
		"action": "delete",
		"resource": "*",
		"effect": "deny"
	},
	{
		"namespace": "secrets",
		"action": "get",
		"resource": "*",
		"effect": "allow"
	}
]`
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("success")

		err := handler.addRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Role created successfully."))
	})

	t.Run("error (400) if role is not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")

		err := handler.addRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide role."))
	})

	t.Run("error (400) if body is invalid", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("success")

		err := handler.addRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Invalid body provided."))
	})

	t.Run("error (500) if error occurs adding the role", func(t *testing.T) {
		body := `[]`
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("error")

		err := handler.addRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusInternalServerError))
		assert.Check(t, cmp.Equal(rec.Body.String(), "An error occurred while adding the role."))
	})
}

func Test_rbacHandler_deleteRole(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("delme")

		err := handler.deleteRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Role deleted successfully."))
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")

		err := handler.deleteRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide role."))
	})

	t.Run("error (400) if error occurs deleting role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("error")

		err := handler.deleteRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusInternalServerError))
		assert.Check(t, cmp.Equal(rec.Body.String(), "An error occurred while deleting the role."))
	})
}

func Test_rbacHandler_getAllRoles(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/rbac/roles")

	err := handler.getAllRoles(c)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
	assert.Check(t, cmp.Equal(rec.Body.String(), "[\"role-a\",\"role-b\"]\n"))
}

func Test_rbacHandler_getUserAttachedRoles(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	t.Run("success (200) if user is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:username/rbac/roles")
		c.SetParamNames("username")
		c.SetParamValues("test")

		err := handler.getUserAttachedRoles(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
		assert.Check(t, cmp.Equal(rec.Body.String(), "[\"role-a\",\"role-b\"]\n"))
	})

	t.Run("error (400) if no username is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:username/rbac/roles")

		err := handler.getUserAttachedRoles(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide username."))
	})

	t.Run("error (500) if error occurs getting roles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:username/rbac/roles")
		c.SetParamNames("username")
		c.SetParamValues("error")

		err := handler.getUserAttachedRoles(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusInternalServerError))
		assert.Check(t, cmp.Equal(rec.Body.String(), "An error occurred while getting the roles."))
	})
}

func Test_rbacHandler_getRolesAttachedUsers(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attached")
		c.SetParamNames("role")
		c.SetParamValues("test")

		err := handler.getRolesAttachedUsers(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
		assert.Check(t, cmp.Equal(rec.Body.String(), "[\"user-a\",\"user-b\"]\n"))
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attached")

		err := handler.getRolesAttachedUsers(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide role."))
	})

	t.Run("error (500) if an error occurs getting users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attached")
		c.SetParamNames("role")
		c.SetParamValues("error")

		err := handler.getRolesAttachedUsers(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusInternalServerError))
		assert.Check(t, cmp.Equal(rec.Body.String(), "An error occurred while getting the users."))
	})
}

func Test_rbacHandler_attachRole(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("test-role", "test-user")

		err := handler.attachRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Role attached successfully."))
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")

		err := handler.attachRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide role."))
	})

	t.Run("error (400) if no username is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role")
		c.SetParamValues("test-role")

		err := handler.attachRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide username."))
	})

	t.Run("error (500) if error occurs attaching the role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("error", "error")

		err := handler.attachRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusInternalServerError))
		assert.Check(t, cmp.Equal(rec.Body.String(), "An error occurred while attaching the role."))
	})
}

func Test_rbacHandler_detachRole(t *testing.T) {
	handlerService := NewGaiaHandler(Dependencies{})

	handler := rbacHandler{
		svc: &mockRBACSvc{},
	}

	e := echo.New()
	_ = handlerService.InitHandlers(e)

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("test-role", "test-user")

		err := handler.detatchRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusOK))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Role detached successfully."))
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")

		err := handler.detatchRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide role."))
	})

	t.Run("error (400) if no username is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role")
		c.SetParamValues("test-role")

		err := handler.detatchRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusBadRequest))
		assert.Check(t, cmp.Equal(rec.Body.String(), "Must provide username."))
	})

	t.Run("error (500) if error occurs detaching the role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("error", "error")

		err := handler.detatchRole(c)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(rec.Code, http.StatusInternalServerError))
		assert.Check(t, cmp.Equal(rec.Body.String(), "An error occurred while detaching the role."))
	})
}
