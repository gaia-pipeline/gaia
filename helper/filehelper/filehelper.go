package filehelper

import (
	"crypto/sha256"
	"io"
	"os"
)

// GetSHA256Sum accepts a path to a file.
// It load's the file and calculates a SHA256 Checksum and returns it.
func GetSHA256Sum(path string) ([]byte, error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create sha256 obj and insert bytes
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	// return sha256 checksum
	return h.Sum(nil), nil
}
