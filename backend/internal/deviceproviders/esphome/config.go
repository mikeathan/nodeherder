package esphome

import (
	"encoding/base64"
	"errors"
)

// ValidateEncryptionKey accepts an empty key for plaintext or padded standard
// Base64 representing exactly 32 bytes for Noise. Errors never include the key.
func ValidateEncryptionKey(key string) error {
	if key == "" {
		return nil
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(key)
	if err != nil || len(decoded) != 32 || base64.StdEncoding.EncodeToString(decoded) != key {
		return errors.New("ESPHome encryption key must be padded Base64 encoding of 32 bytes")
	}
	return nil
}
