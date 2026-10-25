package repository

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLoadOrCreateIntegrationKeyringFirstRunAndRestore(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "private", "integration-keyring.json")
	data := filepath.Join(root, "data")
	first, err := LoadOrCreateIntegrationKeyring(path, filepath.Join(data, "integrations.db"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("key permissions", info, err)
	}
	directory, _ := os.Stat(filepath.Dir(path))
	if directory.Mode().Perm() != 0700 {
		t.Fatal("directory must be private")
	}
	original, _ := os.ReadFile(path)
	repo, err := OpenIntegrationRepository(data, first)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(func(tx *IntegrationTx) error {
		return tx.Put("instances", "test", map[string]string{"value": "private"})
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := LoadOrCreateIntegrationKeyring(path, filepath.Join(data, "integrations.db"))
	if err != nil || second.ActiveKeyID != first.ActiveKeyID {
		t.Fatal("key changed on restart", err)
	}
	repo, err = OpenIntegrationRepository(data, second)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if err := repo.View(func(tx *IntegrationTx) error {
		var value map[string]string
		ok, err := tx.Get("instances", "test", &value)
		if !ok || value["value"] != "private" {
			t.Error("encrypted state lost")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(original, after) {
		t.Fatal("startup rewrote the key file")
	}
	// Losing a key after real encrypted data exists must not create a replacement.
	if err := repo.Close(); err != nil {
		t.Fatal(err)
	}
	beforeDB, err := os.ReadFile(filepath.Join(data, "integrations.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadOrCreateIntegrationKeyring(path, filepath.Join(data, "integrations.db")); err == nil {
		t.Fatal("replaced a lost encryption key")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("unexpected replacement key", err)
	}
	afterDB, err := os.ReadFile(filepath.Join(data, "integrations.db"))
	if err != nil || !bytes.Equal(beforeDB, afterDB) {
		t.Fatal("lost-key recovery modified encrypted data", err)
	}
}

func TestLoadOrCreateIntegrationKeyringFailsClosed(t *testing.T) {
	for _, scenario := range []string{"existing_db", "malformed_key", "public_key", "public_directory"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "integration-keyring.json")
			db := filepath.Join(root, "integrations.db")
			switch scenario {
			case "existing_db":
				if err := os.WriteFile(db, nil, 0600); err != nil {
					t.Fatal(err)
				}
			case "malformed_key":
				if err := os.WriteFile(path, []byte("invalid"), 0600); err != nil {
					t.Fatal(err)
				}
			case "public_key":
				if err := os.WriteFile(path, []byte("invalid"), 0644); err != nil {
					t.Fatal(err)
				}
			case "public_directory":
				if err := os.Chmod(root, 0755); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := os.ReadFile(path)
			if _, err := LoadOrCreateIntegrationKeyring(path, db); err == nil {
				t.Fatal("unsafe provisioning accepted")
			}
			after, _ := os.ReadFile(path)
			if !bytes.Equal(before, after) {
				t.Fatal("key replaced during recovery")
			}
			if scenario == "existing_db" {
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Fatal("invented replacement key", err)
				}
			}
		})
	}
}

func TestLoadOrCreateIntegrationKeyringConcurrentPublication(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "private", "integration-keyring.json")
	db := filepath.Join(root, "integrations.db")
	const count = 12
	keys := make([]IntegrationKeyring, count)
	errors := make([]error, count)
	var wg sync.WaitGroup
	for index := range count {
		wg.Add(1)
		go func() { defer wg.Done(); keys[index], errors[index] = LoadOrCreateIntegrationKeyring(path, db) }()
	}
	wg.Wait()
	for index := range count {
		if errors[index] != nil {
			t.Fatal(errors[index])
		}
		if keys[index].ActiveKeyID != keys[0].ActiveKeyID || keys[index].Keys[keys[0].ActiveKeyID] != keys[0].Keys[keys[0].ActiveKeyID] {
			t.Fatal("startups used different keys")
		}
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatal("staged key file leaked", entries, err)
	}
}
