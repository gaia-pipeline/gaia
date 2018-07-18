package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestPipelineGitLSRemote(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		t.Fatalf("error creating data dir %v", err.Error())
	}

	defer func() {
		gaia.Cfg = nil
		os.RemoveAll(dataDir)
	}()

	gaia.Cfg = &gaia.Config{
		Logger:   hclog.NewNullLogger(),
		DataPath: dataDir,
	}

	dataStore := store.NewStore()
	assert.NoError(t, dataStore.Init())

	e := echo.New()
	InitHandlers(e, dataStore, nil)

	t.Run("fails with invalid data", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/gitlsremote", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, PipelineGitLSRemote(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("fails with invalid access", func(t *testing.T) {
		repoURL := "https://example.com"
		body := map[string]string{
			"url":      repoURL,
			"username": "admin",
			"password": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, PipelineGitLSRemote(c))
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("otherwise succeed", func(t *testing.T) {
		repoURL := "https://github.com/gaia-pipeline/gaia"
		body := map[string]string{
			"url":      repoURL,
			"username": "admin",
			"password": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, PipelineGitLSRemote(c))
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
