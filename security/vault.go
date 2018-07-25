package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/gaia-pipeline/gaia"
)

const (
	vaultName = ".gaia_vault"
)

// Vault is a secret storage for data that gaia needs to store encrypted.
type Vault struct {
	Path string
	Cert []byte
	data map[string][]byte
	sync.RWMutex
}

// NewVault creates a vault which is a simple k/v storage medium with AES encryption.
// The format is:
// KEY=VALUE
// KEY2=VALUE2
func NewVault() (*Vault, error) {
	v := new(Vault)
	vaultPath := filepath.Join(gaia.Cfg.VaultPath, vaultName)
	if _, osErr := os.Stat(vaultPath); os.IsNotExist(osErr) {
		gaia.Cfg.Logger.Info("vault file doesn't exist. creating...")
		_, err := os.Create(vaultPath)
		if err != nil {
			gaia.Cfg.Logger.Error("failed creating vault file: ", err.Error())
			return nil, err
		}
	}
	v.Cert = []byte("readthecertificatehere")
	v.Path = vaultPath
	v.data = make(map[string][]byte, 0)
	return v, nil
}

// LoadSecrets decrypts the contents of the vault and fills up a map of data to work with.
func (v *Vault) LoadSecrets() error {
	r, err := ioutil.ReadFile(v.Path)
	if err != nil {
		return err
	}
	data, err := v.decrypt(r)
	if err != nil {
		return err
	}
	return v.parseToMap(data)
}

// SaveSecrets encrypts data passed to the vault in a k/v format and saves it to the vault file.
func (v *Vault) SaveSecrets() error {
	// open f
	data := v.parseFromMap()
	encryptedData, err := v.encrypt(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(v.Path, []byte(encryptedData), 0400)
	return err
}

// GetAll returns all keys and values in a copy of the internal data.
func (v *Vault) GetAll() map[string][]byte {
	m := make(map[string][]byte, 0)
	for k, v := range v.data {
		m[k] = v
	}
	return m
}

// Add adds a value to the vault. This operation is safe to use concurrently.
// Add will overwrite if the key already exists and not warn.
func (v *Vault) Add(key string, value []byte) {
	v.Lock()
	defer v.Unlock()
	v.data[key] = value
}

// Remove removes a key from the vault. This operation is safe to use concurrently.
// Remove is a no-op if the data doesn't exist.
func (v *Vault) Remove(key string) {
	v.Lock()
	defer v.Unlock()
	delete(v.data, key)
}

// Get returns a value for a key. This operation is safe to use concurrently.
// Get will return an error if the data doesn't exist.
func (v *Vault) Get(key string) ([]byte, error) {
	v.RLock()
	defer v.RUnlock()
	val, ok := v.data[key]
	if !ok {
		message := fmt.Sprintf("key '%s' not found in vault", key)
		return []byte{}, errors.New(message)
	}

	return val, nil
}

// encrypt uses an aes cipher provided by the certificate file for encryption.
// We don't store the password in the file. an error will be thrown in case the encryption
// operation encounters a problem which will most likely be due to a mistyped password.
// We will return this possibilitiy but we won't know for sure if that's the cause.
// The password is padded with 0x04 to Blocklenght. IV randomized to blocksize and length of the message.
// In the end we encrypt the whole thing to Base64 for ease of saving an handling.
func (v *Vault) encrypt(data []byte) (string, error) {
	if len(data) < 1 {
		return "", errors.New("tried to save an empty data set")
	}
	paddedPassword := v.pad(v.Cert)
	block, err := aes.NewCipher(paddedPassword)
	if err != nil {
		return "", err
	}

	msg := v.pad(data)
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(msg))
	finalMsg := base64.URLEncoding.EncodeToString(ciphertext)
	return finalMsg, nil
}

func (v *Vault) decrypt(data []byte) ([]byte, error) {
	if len(data) < 1 {
		gaia.Cfg.Logger.Info("the vault is empty")
		return []byte{}, nil
	}
	paddedPassword := v.pad(v.Cert)
	block, err := aes.NewCipher(paddedPassword)
	if err != nil {
		return []byte{}, err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(string(data))
	if err != nil {
		return []byte{}, err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return []byte{}, errors.New("blocksize must be multipe of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := v.unpad(msg)
	if err != nil {
		return []byte{}, err
	}
	return unpadMsg, nil
}

// ParseToMap will update the Vault data map with values from
// an encrypted file content.
func (v *Vault) parseToMap(data []byte) error {
	if len(data) < 1 {
		return nil
	}
	row := bytes.Split(data, []byte("\n"))
	for _, r := range row {
		d := bytes.Split(r, []byte("="))
		if len(d) < 2 {
			// It is possible that if there is a password failure it's not caught
			// by the padding process. Here it will be caught because we can't
			// marshal the data into proper k/v pairs.
			return errors.New("possible mistype of password")
		}
		v.data[string(d[0])] = d[1]
	}
	return nil
}

// ParseFromMap will create a joined by new line set of key value
// pairs ready to be saved.
func (v *Vault) parseFromMap() []byte {
	data := make([][]byte, 0)
	for key, value := range v.data {
		s := fmt.Sprintf("%s=%s", key, value)
		data = append(data, []byte(s))
	}

	return bytes.Join(data, []byte("\n"))
}

// Pad pads the src with 0x04 until block length.
func (v *Vault) pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// Unpad removes the padding from pad.
func (v *Vault) unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("possible mistyped password")
	}

	return src[:(length - unpadding)], nil
}
