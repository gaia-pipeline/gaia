package security

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

func TestNewVault(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.HomePath = tmp
	v, err := NewVault()
	if err != nil {
		t.Fatal(err)
	}
	if v.Path != filepath.Join(gaia.Cfg.HomePath, vaultName) {
		t.Fatal("file path of vault file did not equal expected. was:", v.Path)
	}
}

func TestAddAndGet(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.HomePath = tmp
	v, err := NewVault()
	if err != nil {
		t.Fatal(err)
	}
	v.Add("key", []byte("value"))
	val, err := v.Get("key")
	if bytes.Compare(val, []byte("value")) != 0 {
		t.Fatal("value didn't match expected of 'value'. was: ", string(val))
	}
}

func TestCloseOpenVault(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.HomePath = tmp
	v, err := NewVault()
	if err != nil {
		t.Fatal(err)
	}
	v.Cert = []byte("test")
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.CloseVault()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte, 0)
	err = v.OpenVault()
	if err != nil {
		t.Fatal(err)
	}
	val, err := v.Get("key1")
	if bytes.Compare(val, []byte("value1")) != 0 {
		t.Fatal("could not properly retrieve value for key1. was:", string(val))
	}
}

func TestCloseOpenVaultWithInvalidPassword(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.HomePath = tmp
	v, err := NewVault()
	if err != nil {
		t.Fatal(err)
	}
	v.Cert = []byte("test")
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.CloseVault()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte, 0)
	v.Cert = []byte("invalid")
	err = v.OpenVault()
	if err == nil {
		t.Fatal("error should not have been nil.")
	}
	expected := "possible mistyped password"
	if err.Error() != expected {
		t.Fatalf("didn't get the right error. expected: \n'%s'\n error was: \n'%s'\n", expected, err.Error())
	}
}

func TestAnExistingVaultFileIsNotOverwritten(t *testing.T) {
	tmp := "."
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	v, err := NewVault()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(vaultName)
	v.Cert = []byte("test")
	v.OpenVault()
	v.Add("test", []byte("value"))
	v.CloseVault()
	v2, _ := NewVault()
	if v2.Path != v.Path {
		t.Fatal("paths should have equaled. were: ", v2.Path, v.Path)
	}
	v2.Cert = []byte("test")
	v2.OpenVault()
	if err != nil {
		t.Fatal(err)
	}
	value, err := v2.Get("test")
	if err != nil {
		t.Fatal("couldn't retrieve value: ", err)
	}
	if bytes.Compare(value, []byte("value")) != 0 {
		t.Fatal("test value didn't equal expected of 'value'. was:", string(value))
	}
}

func TestRemovingFromTheVault(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.HomePath = tmp
	v, err := NewVault()
	if err != nil {
		t.Fatal(err)
	}
	v.Cert = []byte("test")
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.CloseVault()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte, 0)
	err = v.OpenVault()
	if err != nil {
		t.Fatal(err)
	}
	val, err := v.Get("key1")
	if bytes.Compare(val, []byte("value1")) != 0 {
		t.Fatal("could not properly retrieve value for key1. was:", string(val))
	}
	v.Remove("key1")
	v.CloseVault()
	v.data = make(map[string][]byte, 0)
	v.OpenVault()
	_, err = v.Get("key1")
	if err == nil {
		t.Fatal("should have failed to retrieve non-existant key")
	}
	expected := "key 'key1' not found in vault"
	if err.Error() != expected {
		t.Fatalf("got the wrong error message. expected: \n'%s'\n was: \n'%s'\n", expected, err.Error())
	}
}
