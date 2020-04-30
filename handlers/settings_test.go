package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia/workers/pipeline"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	gStore "github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

type status struct {
	Status bool
}

type mockSettingStoreService struct {
	gStore.GaiaStore
	get func() (*gaia.StoreConfig, error)
	put func(*gaia.StoreConfig) error
}

func (m mockSettingStoreService) SettingsGet() (*gaia.StoreConfig, error) {
	return m.get()
}

func (m mockSettingStoreService) SettingsPut(c *gaia.StoreConfig) error {
	return m.put(c)
}

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

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// // Initialize echo
	e := echo.New()
	_ = handlerService.InitHandlers(e)
	get := func() (*gaia.StoreConfig, error) {
		return nil, nil
	}
	put := func(*gaia.StoreConfig) error {
		return nil
	}
	m := mockSettingStoreService{get: get, put: put}
	services.MockStorageService(&m)
	defer func() {
		services.MockStorageService(nil)
	}()

	t.Run("switching it on twice should fail", func(t2 *testing.T) {
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/on")

		_ = SettingsPollOn(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}

		req2 := httptest.NewRequest(echo.POST, "/", nil)
		req2.Header.Set("Content-Type", "application/json")
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		c2.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/on")

		_ = SettingsPollOn(c2)
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

		_ = SettingsPollOn(c)
		retStatus := http.StatusBadRequest
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
	})
	t.Run("switching it off while the setting is on should pass", func(t *testing.T) {
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
		c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/off")

		_ = SettingsPollOff(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
		if gaia.Cfg.Poll != false {
			t.Fatalf("poll value should have been set to false. was: %v", gaia.Cfg.Poll)
		}
	})
	t.Run("switching it off while the setting is off should fail", func(t *testing.T) {
		gaia.Cfg = &gaia.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         false,
		}
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/off")

		_ = SettingsPollOff(c)
		retStatus := http.StatusBadRequest
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
	})
	t.Run("getting the value should return the correct setting", func(t *testing.T) {
		gaia.Cfg = &gaia.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         true,
		}
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll")

		_ = SettingsPollGet(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
		var s status
		_ = json.NewDecoder(rec.Body).Decode(&s)
		if s.Status != true {
			t.Fatalf("expected returned status to be true. was: %v", s.Status)
		}
	})
}

func TestGettingSettingFromDBTakesPrecedence(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGettingSettingFromDBTakesPrecedence")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})
	// // Initialize echo
	e := echo.New()
	_ = handlerService.InitHandlers(e)
	get := func() (*gaia.StoreConfig, error) {
		return &gaia.StoreConfig{
			Poll: true,
		}, nil
	}
	put := func(*gaia.StoreConfig) error {
		return nil
	}
	m := mockSettingStoreService{get: get, put: put}
	services.MockStorageService(&m)
	defer services.MockStorageService(nil)
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/")

	_ = SettingsPollGet(c)
	retStatus := http.StatusOK
	if rec.Code != retStatus {
		t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
	}
	var s status
	_ = json.NewDecoder(rec.Body).Decode(&s)
	if s.Status != true {
		t.Fatalf("expected returned status to be true from storage. was: %v", s.Status)
	}
}

func TestSettingPollerOnAlsoSavesSettingsInDB(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSettingPollerOnAlsoSavesSettingsInDB")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// // Initialize echo
	e := echo.New()
	_ = handlerService.InitHandlers(e)
	get := func() (*gaia.StoreConfig, error) {
		return &gaia.StoreConfig{
			Poll: true,
		}, nil
	}
	putCalled := false
	put := func(*gaia.StoreConfig) error {
		putCalled = true
		return nil
	}
	m := mockSettingStoreService{get: get, put: put}
	services.MockStorageService(&m)
	defer services.MockStorageService(nil)
	req := httptest.NewRequest(echo.POST, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/" + gaia.APIVersion + "/setttings/poll/on")

	_ = SettingsPollOn(c)
	retStatus := http.StatusOK
	if rec.Code != retStatus {
		t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
	}

	if putCalled != true {
		t.Fatal("SettingPut should have been called. Was not.")
	}
	putCalled = false
	_ = SettingsPollOff(c)
	if putCalled != true {
		t.Fatal("SettingPut should have been called. Was not.")
	}
}
