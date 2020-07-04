package store

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gaia-pipeline/gaia"
)

func TestBoltStore_Settings(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestBoltStore_SettingsGet")
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

	empty := &gaia.StoreConfig{
		ID:          0,
		Poll:        false,
		RBACEnabled: false,
	}

	config, err := store.SettingsGet()
	assert.NoError(t, err)
	assert.EqualValues(t, empty, config)

	cfg := &gaia.StoreConfig{
		ID:          1,
		Poll:        true,
		RBACEnabled: true,
	}

	err = store.SettingsPut(cfg)
	assert.NoError(t, err)

	config, err = store.SettingsGet()
	assert.NoError(t, err)
	assert.EqualValues(t, config, cfg)
}
