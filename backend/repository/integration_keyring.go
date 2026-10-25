package repository

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LoadOrCreateIntegrationKeyring provisions a native first installation only. An existing DB
// always requires its original key, even if that DB is empty, unreadable or a symlink. Existing
// key files are read and validated rather than overwritten or silently repaired.
func LoadOrCreateIntegrationKeyring(path, databasePath string) (IntegrationKeyring, error) {
	keyring, err := ReadIntegrationKeyring(path)
	if err == nil {
		return keyring, syncIntegrationKeyDirectory(filepath.Dir(path))
	}
	if !errors.Is(err, os.ErrNotExist) {
		return keyring, err
	}
	if err := requireAbsentIntegrationDB(databasePath); err != nil {
		return keyring, err
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return keyring, fmt.Errorf("create private key directory: %w", err)
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return keyring, err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return keyring, fmt.Errorf("integration key directory must be private (0700)")
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return keyring, fmt.Errorf("generate integration key: %w", err)
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return keyring, fmt.Errorf("generate integration key ID: %w", err)
	}
	keyring = IntegrationKeyring{ActiveKeyID: id.String(), Keys: map[string]string{id.String(): base64.StdEncoding.EncodeToString(raw)}}
	content, err := json.Marshal(keyring)
	if err != nil {
		return keyring, err
	}
	if err := publishIntegrationKeyring(path, databasePath, content); err != nil {
		return IntegrationKeyring{}, err
	}
	// Another process may have published first. Always load the actual winning file.
	return ReadIntegrationKeyring(path)
}

func requireAbsentIntegrationDB(path string) error {
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		if err != nil {
			return fmt.Errorf("check existing integration database: %w", err)
		}
		return fmt.Errorf("integration database exists but its key is missing; restore the matching key backup")
	}
	return nil
}

// publishIntegrationKeyring exposes fully written bytes without replacing an existing path.
// O_EXCL on the final file alone would expose a partially written key to a competing startup.
func publishIntegrationKeyring(path, databasePath string, content []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".integration-keyring-*.tmp")
	if err != nil {
		return fmt.Errorf("stage integration key: %w", err)
	}
	defer os.Remove(temporary.Name())
	if _, err = temporary.Write(content); err == nil {
		err = temporary.Sync()
	}
	closeErr := temporary.Close()
	if err != nil {
		return fmt.Errorf("write integration key: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close staged integration key: %w", closeErr)
	}
	// Recheck after staging; a new DB must never be paired with a newly invented key.
	if err := requireAbsentIntegrationDB(databasePath); err != nil {
		// A competing first startup may already have published its key and opened the DB.
		// Reuse that key if present; otherwise preserve the missing-key recovery error.
		if _, readErr := ReadIntegrationKeyring(path); readErr == nil {
			return syncIntegrationKeyDirectory(directory)
		}
		return err
	}
	if err := os.Link(temporary.Name(), path); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("publish integration key: %w", err)
	}
	return syncIntegrationKeyDirectory(directory)
}

// Sync even a competing winner's directory entry before any caller creates the database.
func syncIntegrationKeyDirectory(directory string) error {
	parent, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open integration key directory: %w", err)
	}
	defer parent.Close()
	if err := parent.Sync(); err != nil {
		return fmt.Errorf("sync integration key directory: %w", err)
	}
	return nil
}
