package esphome

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestValidateEncryptionKey(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString(make([]byte, 32))
	for _, key := range []string{"", valid} {
		if err := ValidateEncryptionKey(key); err != nil {
			t.Fatalf("valid key rejected: %v", err)
		}
	}
	for _, key := range []string{"secret", strings.TrimRight(valid, "="), valid + "\n", base64.StdEncoding.EncodeToString(make([]byte, 31))} {
		err := ValidateEncryptionKey(key)
		if err == nil {
			t.Fatal("invalid key accepted")
		}
		if strings.Contains(err.Error(), key) {
			t.Fatal("error reveals key")
		}
	}
}
