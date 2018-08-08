package security

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

type MockVaultStorer struct {
	Error error
}

var store []byte

func (mvs *MockVaultStorer) Init() error {
	store = make([]byte, 0)
	return mvs.Error
}

func (mvs *MockVaultStorer) Read() ([]byte, error) {
	return store, mvs.Error
}

func (mvs *MockVaultStorer) Write(data []byte) error {
	store = data
	return mvs.Error
}

func TestNewVault(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestNewVault")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	mvs := new(MockVaultStorer)
	v.storer = mvs
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddAndGet(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestAddAndGet")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	mvs := new(MockVaultStorer)
	v.storer = mvs
	if err != nil {
		t.Fatal(err)
	}
	v.Add("key", []byte("value"))
	val, err := v.Get("key")
	if bytes.Compare(val, []byte("value")) != 0 {
		t.Fatal("value didn't match expected of 'value'. was: ", string(val))
	}
}

func TestCloseLoadSecrets(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCloseLoadSecrets")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	mvs := new(MockVaultStorer)
	v.storer = mvs
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte, 0)
	err = v.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	val, err := v.Get("key1")
	if bytes.Compare(val, []byte("value1")) != 0 {
		t.Fatal("could not properly retrieve value for key1. was:", string(val))
	}
}

func TestCloseLoadSecretsWithInvalidPassword(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCloseLoadSecretsWithInvalidPassword")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	mvs := new(MockVaultStorer)
	v.storer = mvs
	v.cert = []byte("test")
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte, 0)
	v.cert = []byte("invalid")
	err = v.LoadSecrets()
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
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	mvs := new(MockVaultStorer)
	v.storer = mvs
	defer os.Remove(vaultName)
	defer os.Remove("ca.crt")
	defer os.Remove("ca.key")
	v.cert = []byte("test")
	v.Add("test", []byte("value"))
	v.SaveSecrets()
	v2, _ := NewVault(c, nil)
	v2.storer = mvs
	v2.cert = []byte("test")
	v2.LoadSecrets()
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
	tmp, _ := ioutil.TempDir("", "TestRemovingFromTheVault")
	gaia.Cfg = &gaia.Config{}
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	mvs := new(MockVaultStorer)
	v.storer = mvs
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte, 0)
	err = v.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	val, err := v.Get("key1")
	if bytes.Compare(val, []byte("value1")) != 0 {
		t.Fatal("could not properly retrieve value for key1. was:", string(val))
	}
	v.Remove("key1")
	v.SaveSecrets()
	v.data = make(map[string][]byte, 0)
	v.LoadSecrets()
	_, err = v.Get("key1")
	if err == nil {
		t.Fatal("should have failed to retrieve non-existant key")
	}
	expected := "key 'key1' not found in vault"
	if err.Error() != expected {
		t.Fatalf("got the wrong error message. expected: \n'%s'\n was: \n'%s'\n", expected, err.Error())
	}
}

func TestGetAll(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGetAll")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, err := NewVault(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	mvs := new(MockVaultStorer)
	v.storer = mvs
	v.Add("key1", []byte("value1"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	err = v.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"key1"}
	actual := v.GetAll()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("actual did not equal expected. actual was: %+v, expected: %+v.", actual, expected)
	}
}

func TestEditValueWithAddingItAgain(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestEditValueWithAddingItAgain")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, _ := NewVault(c, nil)
	mvs := new(MockVaultStorer)
	v.storer = mvs
	v.Add("key1", []byte("value1"))
	v.SaveSecrets()
	v.data = make(map[string][]byte, 0)
	v.LoadSecrets()
	v.Add("key1", []byte("value2"))
	v.SaveSecrets()
	v.data = make(map[string][]byte, 0)
	v.LoadSecrets()
	val, _ := v.Get("key1")
	if bytes.Compare(val, []byte("value2")) != 0 {
		t.Fatal("value should have equaled expected 'value2'. was: ", string(val))
	}
}

func TestReadErrorForVault(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestReadErrorForVault")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, _ := NewVault(c, nil)
	mvs := new(MockVaultStorer)
	mvs.Error = errors.New("get vault data error")
	v.storer = mvs
	err := v.LoadSecrets()
	if err == nil {
		t.Fatal("error expected on LoadSecret but got none")
	}
	if err.Error() != "get vault data error" {
		t.Fatal("got a different error than expected. was: ", err.Error())
	}
}

func TestWriteErrorForVault(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestWriteErrorForVault")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, _ := NewVault(c, nil)
	mvs := new(MockVaultStorer)
	mvs.Error = errors.New("write vault data error")
	v.storer = mvs
	err := v.SaveSecrets()
	if err == nil {
		t.Fatal("error expected on LoadSecret but got none")
	}
	if err.Error() != "write vault data error" {
		t.Fatal("got a different error than expected. was: ", err.Error())
	}
}

func TestDefaultStorerIsAFileStorer(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestDefaultStorerIsAFileStorer")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	c, _ := InitCA()
	v, _ := NewVault(c, nil)
	if _, ok := v.storer.(*FileVaultStorer); !ok {
		t.Fatal("default filestorer not created when nil is passed in")
	}
}
