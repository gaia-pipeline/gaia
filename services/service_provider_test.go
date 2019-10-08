package services

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/gaia-pipeline/gaia/store/memdb"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	"github.com/hashicorp/go-hclog"
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
	_, _ = StorageService()
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
	_, _ = CertificateService()
	_, _ = DefaultVaultService()
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
	_, _ = CertificateService()
	defer func() { certificateService = nil }()
	if certificateService == nil {
		t.Fatal("service should not be nil")
	}
}

func TestMemDBService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestMemDBService")
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
	if memDBService != nil {
		t.Fatal("initial service should be nil. was: ", memDBService)
	}
	if _, err := StorageService(); err != nil {
		t.Fatal(err)
	}
	if _, err := MemDBService(storeService); err != nil {
		t.Fatal(err)
	}
	defer func() {
		memDBService = nil
		storeService = nil
	}()

	if memDBService == nil {
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
	security.GaiaVault
}

type testMockMemDBService struct {
	memdb.GaiaMemDB
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
		if _, ok := s1.(*testMockStorageService); !ok {
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
		if _, ok := s1.(*testMockScheduleService); !ok {
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
		if _, ok := s1.(*testMockCertificateService); !ok {
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
		s1, _ := DefaultVaultService()
		if _, ok := s1.(*testMockVaultService); !ok {
			t.Fatalf("want type: '%s' got: '%s'", "testMockVaultService", reflect.TypeOf(s1).String())
		}
		MockVaultService(nil)
		s2, _ := DefaultVaultService()
		if reflect.TypeOf(s2).String() == "*services.testMockVaultService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})

	t.Run("can mock memdb to nil", func(t *testing.T) {
		mcp := new(testMockMemDBService)
		MockMemDBService(mcp)
		s1, _ := DefaultMemDBService()
		if _, ok := s1.(*testMockMemDBService); !ok {
			t.Fatalf("want type: '%s' got: '%s'", "testMockMemDBService", reflect.TypeOf(s1).String())
		}
		MockMemDBService(nil)
		msp := new(testMockStorageService)
		MockStorageService(msp)
		s2, _ := MemDBService(storeService)
		if reflect.TypeOf(s2).String() == "*services.testMockMemDBService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})
}

func TestDefaultVaultStorer(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestDefaultVaultStorer")
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
	_, _ = CertificateService()
	v, err := DefaultVaultService()
	if err != nil {
		t.Fatal(err)
	}
	if va, ok := v.(security.GaiaVault); !ok {
		t.Fatal("DefaultVaultService should have given back a GaiaVault. was instead: ", va)
	}
}

func TestDefaultMemDBService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestDefaultMemDBService")
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
	_, _ = StorageService()
	v, err := DefaultMemDBService()
	if err != nil {
		t.Fatal(err)
	}
	if va, ok := v.(memdb.GaiaMemDB); !ok {
		t.Fatal("DefaultVaultService should have given back a GaiaVault. was instead: ", va)
	}
}
