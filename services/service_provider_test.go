package services

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	hclog "github.com/hashicorp/go-hclog"
)

func TestStorageService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestStorageService")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.DataPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	if storeService != nil {
		t.Fatal("initial service should be nil. was: ", storeService)
	}
	StorageService()
	defer func() { storeService = nil }()
	if storeService == nil {
		t.Fatal("storage service should not be nil")
	}
}

func TestSchedulerService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSchedulerService")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.DataPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	if schedulerService != nil {
		t.Fatal("initial service should be nil. was: ", schedulerService)
	}
	if _, err := StorageService(); err != nil {
		t.Fatal(err)
	}
	sService, err := SchedulerService()
	if err != nil {
		t.Fatal(err)
	}
	if sService == nil {
		t.Fatal("scheduler service should not be nil")
	}
}

func TestVaultService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestVaultService")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.CAPath = tmp
	gaia.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	if vaultService != nil {
		t.Fatal("initial service should be nil. was: ", vaultService)
	}
	CertificateService()
	VaultService(nil)
	defer func() {
		certificateService = nil
		vaultService = nil
	}()

	if vaultService == nil {
		t.Fatal("service should not be nil")
	}
}

func TestCertificateService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCertificateService")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.CAPath = tmp
	gaia.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	if certificateService != nil {
		t.Fatal("initial service should be nil. was: ", certificateService)
	}
	CertificateService()
	defer func() { certificateService = nil }()
	if certificateService == nil {
		t.Fatal("service should not be nil")
	}
}

type testMockStorageService struct {
	store.GaiaStore
}

type testMockScheduleService struct {
	scheduler.GaiaScheduler
}

type testMockCertificateService struct {
	security.CAAPI
}

type testMockVaultService struct {
	security.VaultAPI
}

func TestCanMockServiceToNil(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCanMockServiceToNil")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.CAPath = tmp
	gaia.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	if certificateService != nil {
		t.Fatal("initial service should be nil. was: ", certificateService)
	}

	t.Run("can mock storage to nil", func(t *testing.T) {
		mcp := new(testMockStorageService)
		MockStorageService(mcp)
		s1, _ := StorageService()
		if reflect.TypeOf(s1).String() != "*services.testMockStorageService" {
			t.Fatalf("want type: '%s' got: '%s'", "testMockStorageService", reflect.TypeOf(s1).String())
		}
		MockStorageService(nil)
		s2, _ := StorageService()
		if reflect.TypeOf(s2).String() == "*services.testMockStorageService" {
			t.Fatalf("want type: '%s' got: '%s'", "BoltStorage", reflect.TypeOf(s2).String())
		}
	})

	t.Run("can mock scheduler to nil", func(t *testing.T) {
		mcp := new(testMockScheduleService)
		MockSchedulerService(mcp)
		s1, _ := SchedulerService()
		if reflect.TypeOf(s1).String() != "*services.testMockScheduleService" {
			t.Fatalf("want type: '%s' got: '%s'", "testMockScheduleService", reflect.TypeOf(s1).String())
		}
		MockSchedulerService(nil)
		s2, _ := SchedulerService()
		if reflect.TypeOf(s2).String() == "*services.testMockScheduleService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})

	t.Run("can mock certificate to nil", func(t *testing.T) {
		mcp := new(testMockCertificateService)
		MockCertificateService(mcp)
		s1, _ := CertificateService()
		if reflect.TypeOf(s1).String() != "*services.testMockCertificateService" {
			t.Fatalf("want type: '%s' got: '%s'", "testMockCertificateService", reflect.TypeOf(s1).String())
		}
		MockCertificateService(nil)
		s2, _ := CertificateService()
		if reflect.TypeOf(s2).String() == "*services.testMockCertificateService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})

	t.Run("can mock vault to nil", func(t *testing.T) {
		mcp := new(testMockVaultService)
		MockVaultService(mcp)
		s1, _ := VaultService(nil)
		if reflect.TypeOf(s1).String() != "*services.testMockVaultService" {
			t.Fatalf("want type: '%s' got: '%s'", "testMockVaultService", reflect.TypeOf(s1).String())
		}
		MockVaultService(nil)
		s2, _ := VaultService(nil)
		if reflect.TypeOf(s2).String() == "*services.testMockVaultService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})
}
