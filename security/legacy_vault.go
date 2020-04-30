package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	"github.com/gaia-pipeline/gaia"
)

const (
	secretCheckKey   = "GAIA_CHECK_SECRET"
	secretCheckValue = "!CHECK_ME!"
)

func (v *Vault) legacyDecrypt(data []byte) ([]byte, error) {
	if len(data) < 1 {
		gaia.Cfg.Logger.Info("the vault is empty")
		return []byte{}, nil
	}
	paddedPassword := v.pad(v.cert)
	ci := base64.URLEncoding.EncodeToString(paddedPassword)
	block, err := aes.NewCipher([]byte(ci[:aes.BlockSize]))
	if err != nil {
		return []byte{}, err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(string(data))
	if err != nil {
		return []byte{}, err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return []byte{}, errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := v.unpad(msg)
	if err != nil {
		return []byte{}, err
	}

	if !bytes.Contains(unpadMsg, []byte(secretCheckValue)) {
		return []byte{}, errors.New("possible mistyped password")
	}
	return unpadMsg, nil
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
