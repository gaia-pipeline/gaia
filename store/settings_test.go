package store

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"

	"github.com/gaia-pipeline/gaia"
)

func TestBoltStore_SettingsRBACGet(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestBoltStore_SettingsRBACGet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	config, err := store.SettingsRBACGet()
	assert.DeepEqual(t, config, gaia.RBACConfig{Enabled: false})
	assert.NilError(t, err)

	err = store.SettingsRBACPut(gaia.RBACConfig{
		Enabled: true,
	})
	assert.NilError(t, err)

	config, err = store.SettingsRBACGet()
	assert.DeepEqual(t, config, gaia.RBACConfig{Enabled: true})
	assert.NilError(t, err)
}
