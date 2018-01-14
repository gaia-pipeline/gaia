package bolt

import (
	bolt "github.com/coreos/bbolt"
	"github.com/michelvocks/gaia"
)

// Bolt represents the access type for bolt
type Bolt struct {
	DB *bolt.DB
}

// NewBolt creates a new instance of bolt.
func NewBolt() *Bolt {
	b := &Bolt{}

	return b
}

// Init initalizes the connection to the bolt file.
// This should be called only once per database file
// because bolt holds a lock on the database file.
func (b *Bolt) Init(cfg *gaia.Config) error {
	db, err := bolt.Open(cfg.Bolt.Path, cfg.Bolt.Mode, nil)
	if err != nil {
		return err
	}

	// Pointer cached for later use
	b.DB = db
	return nil
}
