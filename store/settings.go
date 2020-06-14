package store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/gaia-pipeline/gaia"
)

const (
	configSettings     = "gaia_config_settings"
	rbacConfigSettings = "gaia_rbac_settings"
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
		return b.Put([]byte(configSettings), buf)
	})
}

// SettingsGet gets a pipeline by given id.
func (s *BoltStore) SettingsGet() (*gaia.StoreConfig, error) {
	var config = &gaia.StoreConfig{}

	return config, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(settingsBucket)

		// Get pipeline
		v := b.Get([]byte(configSettings))

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

// SettingsRBACPut inserts or updates the rbac config settings.
func (s *BoltStore) SettingsRBACPut(config gaia.RBACConfig) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(settingsBucket)

		buf, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal rbac config: %w", err)
		}

		return b.Put([]byte(rbacConfigSettings), buf)
	})
}

// SettingsRBACGet gets the rbac config settings.
func (s *BoltStore) SettingsRBACGet() (gaia.RBACConfig, error) {
	var config = gaia.RBACConfig{}

	return config, s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(settingsBucket)

		v := b.Get([]byte(rbacConfigSettings))
		if v == nil {
			config = gaia.RBACConfig{Enabled: false}
			return nil
		}

		err := json.Unmarshal(v, &config)
		if err != nil {
			return fmt.Errorf("failed to unmarshal rbac config: %w", err)
		}

		return nil
	})
}
