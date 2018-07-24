package security

import (
	"os"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
)

const (
	vaultName = ".gaia_vault"
)

// Vault is a secret storage for data that gaia needs to store encrypted.
type Vault struct {
	File *os.File
	Path string
	Data map[string]interface{}
}

// NewVault creates a vault which is a simple k/v storage medium with AES encryption.
// The format is:
// KEY=VALUE
// KEY2=VALUE2
func NewVault() (*Vault, error) {
	v := new(Vault)
	vaultPath := filepath.Join(gaia.Cfg.HomePath, vaultName)
	f, err := os.Create(vaultPath)
	if err != nil {
		gaia.Cfg.Logger.Error("failed creating vault file: ", err.Error())
		return nil, err
	}
	v.File = f
	v.Data = make(map[string]interface{}, 0)
	return v, nil
}

// OpenVault decrypts the contents of the vault and fills up a map of data to work with.
func (v *Vault) OpenVault(password string) error {
	return nil
}

// CloseVault encrypts data passed to the vault in a k/v format.
func (v *Vault) CloseVault() error {
	return nil
}

func encrypt(data []byte, password string) ([]byte, error) {
	return []byte{}, nil
}

func decrypt(data []byte, password string) ([]byte, error) {
	return []byte{}, nil
}
