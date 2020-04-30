package security

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/hashicorp/go-hclog"
)

type MockVaultStorer struct {
	ReadError  error
	InitError  error
	WriteError error
}

var store []byte

func (mvs *MockVaultStorer) Init() error {
	if len(store) < 1 {
		store = make([]byte, 0)
	}
	return mvs.InitError
}

func (mvs *MockVaultStorer) Read() ([]byte, error) {
	return store, mvs.ReadError
}

func (mvs *MockVaultStorer) Write(data []byte) error {
	store = data
	return mvs.WriteError
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
	mvs := new(MockVaultStorer)
	_, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewVaultNilStorer(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestNewVaultNilStorer")
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
	mvs := new(MockVaultStorer)
	_, err := NewVault(c, mvs)
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	v.Add("key", []byte("value"))
	val, _ := v.Get("key")
	if !bytes.Equal(val, []byte("value")) {
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte)
	err = v.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	val, _ := v.Get("key1")
	if !bytes.Equal(val, []byte("value1")) {
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	v.key = []byte("change this password to a secret")
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte)
	v.key = []byte("change this pa00word to a secret")
	err = v.LoadSecrets()
	if err == nil {
		t.Fatal("error should not have been nil.")
	}
	expected := "cipher: message authentication failed"
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(vaultName)
	defer os.Remove("ca.crt")
	defer os.Remove("ca.key")
	v.key = []byte("change this password to a secret")
	v.Add("test", []byte("value"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v2, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	v2.key = []byte("change this password to a secret")
	err = v2.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	value, err := v2.Get("test")
	if err != nil {
		t.Fatal("couldn't retrieve value: ", err)
	}
	if !bytes.Equal(value, []byte("value")) {
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	v.Add("key1", []byte("value1"))
	v.Add("key2", []byte("value2"))
	err = v.SaveSecrets()
	if err != nil {
		t.Fatal(err)
	}
	v.data = make(map[string][]byte)
	err = v.LoadSecrets()
	if err != nil {
		t.Fatal(err)
	}
	val, _ := v.Get("key1")
	if !bytes.Equal(val, []byte("value1")) {
		t.Fatal("could not properly retrieve value for key1. was:", string(val))
	}
	v.Remove("key1")
	_ = v.SaveSecrets()
	v.data = make(map[string][]byte)
	_ = v.LoadSecrets()
	_, err = v.Get("key1")
	if err == nil {
		t.Fatal("should have failed to retrieve non-existent key")
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
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
	mvs := new(MockVaultStorer)
	v, _ := NewVault(c, mvs)
	v.Add("key1", []byte("value1"))
	_ = v.SaveSecrets()
	v.data = make(map[string][]byte)
	_ = v.LoadSecrets()
	v.Add("key1", []byte("value2"))
	_ = v.SaveSecrets()
	v.data = make(map[string][]byte)
	_ = v.LoadSecrets()
	val, _ := v.Get("key1")
	if !bytes.Equal(val, []byte("value2")) {
		t.Fatal("value should have equaled expected 'value2'. was: ", string(val))
	}
}

func TestInitErrorForVault(t *testing.T) {
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
	mvs := new(MockVaultStorer)
	mvs.InitError = errors.New("init error")
	_, err := NewVault(c, mvs)
	if err == nil {
		t.Fatal("error expected on NewVault but got none")
	}
	if err.Error() != "init error" {
		t.Fatal("got a different error than expected. was: ", err.Error())
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
	mvs := new(MockVaultStorer)
	mvs.ReadError = errors.New("get vault data error")
	v, _ := NewVault(c, mvs)
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
	mvs := new(MockVaultStorer)
	mvs.WriteError = errors.New("write vault data error")
	v, _ := NewVault(c, mvs)
	err := v.SaveSecrets()
	if err == nil {
		t.Fatal("error expected on LoadSecret but got none")
	}
	if err.Error() != "write vault data error" {
		t.Fatal("got a different error than expected. was: ", err.Error())
	}
}

func TestNonceCounter(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestNonceCounter")
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
	mvs := new(MockVaultStorer)
	v, err := NewVault(c, mvs)
	if err != nil {
		t.Fatal(err)
	}
	v.Add("key1", []byte("value1"))
	beginCounter := v.counter
	for i := 0; i < 3; i++ {
		err = v.SaveSecrets()
		if err != nil {
			t.Fatal(err)
		}
		err = v.LoadSecrets()
		if err != nil {
			t.Fatal(err)
		}
	}
	if v.counter == beginCounter {
		t.Fatal("counter should have not equaled to the count at the begin of the test.")
	}
	want := uint64(3)
	if v.counter != want {
		t.Fatalf("counter should have been %d. got: %d\n", want, v.counter)
	}
}

func TestEmptyVault(t *testing.T) {
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	v := Vault{}
	t.Run("empty vault", func(t *testing.T) {
		var data []byte
		_, err := v.decrypt(data)
		if err != nil {
			t.Fatal("was not expecting an error. was: ", err)
		}
		want := "the vault is empty"
		if strings.Contains(want, buf.String()) {
			t.Fatalf("wanted log message '%s'. Got: %s", want, buf.String())
		}
	})
}

func TestDefaultMemDBService(t *testing.T) {
	_, err := NewVault(nil, nil)
	if err == nil {
		t.Fatal("NewVault without a storer should have thrown an error. Got no error.")
	}
	if err.Error() != "vault must be created with a valid VaultStore" {
		t.Fatal("error did not equal expected error. got: ", err.Error())
	}
}

func TestAllTheHexDecrypts(t *testing.T) {
	v := Vault{}
	t.Run("encoded data", func(t *testing.T) {
		data := []byte("invalid")
		_, err := v.decrypt(data)
		if err == nil {
			t.Fatal("should have failed since data is not valid hex string")
		}
	})
	t.Run("invalid data format", func(t *testing.T) {
		d := []byte("asdf&&asdf")
		data := []byte(hex.EncodeToString(d))
		_, err := v.decrypt(data)
		if err == nil {
			t.Fatal("should have failed since data did not contain delimiter")
		}
		want := "invalid number of returned splits from data. was:  1\n"
		if err.Error() != want {
			t.Fatalf("want: %s, got: %s", want, err.Error())
		}
	})
	t.Run("invalid nonce", func(t *testing.T) {
		d := []byte("asdf||asdf")
		data := []byte(hex.EncodeToString(d))
		_, err := v.decrypt(data)
		if err == nil {
			t.Fatal("should have failed since data did not contain delimiter")
		}
	})
	t.Run("invalid data", func(t *testing.T) {
		nonce := hex.EncodeToString([]byte("valid"))
		d := []byte(nonce + "||asdf")
		data := []byte(hex.EncodeToString(d))
		_, err := v.decrypt(data)
		if err == nil {
			t.Fatal("should have failed since data did not contain delimiter")
		}
	})
}

func TestLegacyDecryptOfOldVaultFile(t *testing.T) {
	oldVault, err := ioutil.ReadFile("./testdata/gaia_vault")
	if err != nil {
		t.Fatal(err)
	}
	key, err := ioutil.ReadFile("./testdata/ca.key")
	if err != nil {
		t.Fatal(err)
	}
	v := Vault{
		cert: key,
	}
	content, err := v.legacyDecrypt(oldVault)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "test=secret") {
		t.Fatal("was expecting content to have 'test=secret'. it was: ", string(content))
	}
}
