package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// SettingsPut puts settings into the store.
func (s *BoltStore) SettingsPut(c *gaia.StoreConfig) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get settings bucket
		b := tx.Bucket(settingsBucket)

		// Marshal pipeline data into bytes.
		buf, err := json.Marshal(c)
		if err != nil {
			return err
		}

		// Persist bytes to settings bucket.
		return b.Put([]byte("gaia_config_settings"), buf)
	})
}

// SettingsGet gets a pipeline by given id.
func (s *BoltStore) SettingsGet() (*gaia.StoreConfig, error) {
	var config = &gaia.StoreConfig{}

	return config, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(settingsBucket)

		// Get pipeline
		v := b.Get([]byte("gaia_config_settings"))

		// Check if we found the pipeline
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		err := json.Unmarshal(v, config)
		if err != nil {
			return err
		}

		return nil
	})
}
