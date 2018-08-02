package services

import (
	"bytes"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

func TestStorageService(t *testing.T) {
	tmp := os.TempDir()
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
	tmp := os.TempDir()
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
	SchedulerService()
	defer func() { schedulerService = nil }()
	if schedulerService == nil {
		t.Fatal("service should not be nil")
	}
}

func TestVaultService(t *testing.T) {
	tmp := os.TempDir()
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
	tmp := os.TempDir()
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
