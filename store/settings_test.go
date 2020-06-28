package store

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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
	assert.NoError(t, err)
	assert.EqualValues(t, config, gaia.RBACConfig{Enabled: false})

	err = store.SettingsRBACPut(gaia.RBACConfig{
		Enabled: true,
	})
	assert.NoError(t, err)

	config, err = store.SettingsRBACGet()
	assert.NoError(t, err)
	assert.EqualValues(t, config, gaia.RBACConfig{Enabled: true})
}
