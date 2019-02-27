package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

func TestSetPollerToggle(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSetPollerON")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	// // Initialize echo
	e := echo.New()
	InitHandlers(e)

	t.Run("switching it on twice should fail", func(t2 *testing.T) {
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/on")

		SettingsPollOn(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}

		req2 := httptest.NewRequest(echo.POST, "/", nil)
		req2.Header.Set("Content-Type", "application/json")
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		c2.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/on")

		SettingsPollOn(c2)
		secondRetStatus := http.StatusBadRequest
		if rec2.Code != secondRetStatus {
			t.Fatalf("expected response code %v got %v", secondRetStatus, rec2.Code)
		}
	})
	t.Run("switching it on while the setting is on should fail", func(t *testing.T) {
		gaia.Cfg = &gaia.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         true,
		}
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/on")

		SettingsPollOn(c)
		retStatus := http.StatusBadRequest
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
	})
}
