package security

// CreateVault creates the vault file which is a simple k/v storage medium with AES encryption.
// The format is:
// KEY=VALUE
// KEY2=VALUE2
func CreateVault() error {
	return nil
}

// OpenVault decrypts the contents of the vault and fills up a map of data to work with.
func OpenVault(password string) map[string]interface{} {
	var vault map[string]interface{}
	return vault
}

// SaveVault encrypts data passed to the vault in a k/v format.
func SaveVault(data map[string]interface{}, password string) {

}
