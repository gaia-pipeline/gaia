package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/gaia-pipeline/gaia"
)

const (
	vaultName        = ".gaia_vault"
	secretCheckKey   = "GAIA_CHECK_SECRET"
	secretCheckValue = "!CHECK_ME!"
)

// VaultAPI defines a set of apis that a Vault must provide in order to be a Gaia Vault.
type VaultAPI interface {
	LoadSecrets() error
	GetAll() []string
	SaveSecrets() error
	Add(key string, value []byte)
	Remove(key string)
	Get(key string) ([]byte, error)
}

// VaultStorer defines a storage medium for the Vault.
type VaultStorer interface {
	// Init initializes the medium by creating the file, or bootstrapping the
	// db or simply setting up an in-memory mock storage device. The Init
	// function of a storage medium should be idempotent. Meaning it should
	// be callable multiple times without changing the underlying medium.
	Init() error
	// Read will read bytes from the storage medium and return it to the caller.
	Read() (data []byte, err error)
	// Write will store the passed in encrypted data. How, is up to the implementor.
	Write(data []byte) error
}

// FileVaultStorer implements VaultStorer as a simple file based storage device.
type FileVaultStorer struct {
	path string
}

// Vault is a secret storage for data that gaia needs to store encrypted.
type Vault struct {
	storer VaultStorer
	cert   []byte
	data   map[string][]byte
	sync.RWMutex
	counter uint64
}

// NewVault creates a vault which is a simple k/v storage medium with AES encryption.
// The format is:
// KEY=VALUE
// KEY2=VALUE2
// NewVault also can take a storer which is an implementation of VaultStorer.
// This defines a storage medium for the vault. If it's left to nil the vault
// will use a default FileVaultStorer.
func NewVault(ca CAAPI, storer VaultStorer) (*Vault, error) {
	v := new(Vault)

	if storer == nil {
		storer = new(FileVaultStorer)
	}
	err := storer.Init()
	if err != nil {
		return nil, err
	}
	// Setting up certificate key content
	_, certKey := ca.GetCACertPath()
	data, err := ioutil.ReadFile(certKey)
	if err != nil {
		return nil, err
	}
	v.storer = storer
	v.cert = data
	v.data = make(map[string][]byte, 0)
	return v, nil
}

// LoadSecrets decrypts the contents of the vault and fills up a map of data to work with.
func (v *Vault) LoadSecrets() error {
	r, err := v.storer.Read()
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
	// clear the hash after saving so the system always has a fresh view of the vault.
	v.data = make(map[string][]byte, 0)
	return v.storer.Write(encryptedData)
}

// GetAll returns all keys and values in a copy of the internal data.
func (v *Vault) GetAll() []string {
	v.RLock()
	defer v.RUnlock()
	m := make([]string, 0)
	for k := range v.data {
		m = append(m, k)
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

// Init initializes the FileVaultStorer.
func (fvs *FileVaultStorer) Init() error {
	vaultPath := filepath.Join(gaia.Cfg.VaultPath, vaultName)
	if _, osErr := os.Stat(vaultPath); os.IsNotExist(osErr) {
		gaia.Cfg.Logger.Info("vault file doesn't exist. creating...")
		_, err := os.Create(vaultPath)
		if err != nil {
			gaia.Cfg.Logger.Error("failed creating vault file: ", "error", err.Error())
			return err
		}
	}
	fvs.path = vaultPath
	return nil
}

// Read defines a read for the FileVaultStorer.
func (fvs *FileVaultStorer) Read() ([]byte, error) {
	r, err := ioutil.ReadFile(fvs.path)
	return r, err
}

// Write defines a read for the FileVaultStorer.
func (fvs *FileVaultStorer) Write(data []byte) error {
	return ioutil.WriteFile(fvs.path, []byte(data), 0400)
}

// encrypt uses an aes cipher provided by the certificate file for encryption.
// We don't store the password in the file. an error will be thrown in case the encryption
// operation encounters a problem which will most likely be due to a mistyped password.
// We will return this possibility but we won't know for sure if that's the cause.
// The password is padded with 0x04 to Blocklenght. IV randomized to blocksize and length of the message.
// In the end we encrypt the whole thing to Base64 for ease of saving an handling.
func (v *Vault) encrypt(data []byte) ([]byte, error) {
	if len(data) < 1 {
		// User has deleted all the secrets. the file will be empty.
		return []byte{}, nil
	}
	secretCheck := fmt.Sprintf("\n%s=%s", secretCheckKey, secretCheckValue)
	data = append(data, []byte(secretCheck)...)
	key := v.pad(v.cert)
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	v.counter++
	nonce := make([]byte, 12)
	binary.LittleEndian.PutUint64(nonce, v.counter)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	src := fmt.Sprintf("%s||%s", nonce, ciphertext)
	var dst []byte
	base64.URLEncoding.Encode(dst, []byte(src))
	return dst, nil
}

func (v *Vault) decrypt(encodedData []byte) ([]byte, error) {
	if len(encodedData) < 1 {
		gaia.Cfg.Logger.Info("the vault is empty")
		return []byte{}, nil
	}
	key := v.pad(v.cert)
	var dst []byte
	base64.URLEncoding.Decode(dst, encodedData)
	var nonce string
	var data string
	fmt.Sscanf(string(dst), "%s||%s", &nonce, &data)
	b := []byte(nonce)
	v.counter = binary.LittleEndian.Uint64(b)

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	plaintext, err := aesgcm.Open(nil, []byte(nonce), []byte(data), nil)
	if err != nil {
		return []byte{}, err
	}

	if !bytes.Contains(plaintext, []byte(secretCheckValue)) {
		return []byte{}, errors.New("possible mistyped password")
	}
	return plaintext, nil
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
		if bytes.Equal(d[0], []byte(secretCheckKey)) {
			continue
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
