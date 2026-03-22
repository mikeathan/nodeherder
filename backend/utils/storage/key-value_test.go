package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBoltKeyValueDatabase_DeleteKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kvtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := NewBoltKeyValueDatabase(dbPath, "test_bucket")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	key := []byte("test-key")
	value := []byte("test-value")

	err = db.Set(key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := db.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(val) != string(value) {
		t.Fatalf("Expected value %s, got %s", string(value), string(val))
	}

	err = db.DeleteKey(key)
	if err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	val, err = db.Get(key)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if val != nil {
		t.Fatalf("Expected nil value after deletion, got %s", string(val))
	}

	err = db.DeleteKey([]byte("non-existent-key"))
	if err != nil {
		t.Fatalf("DeleteKey with non-existent key failed: %v", err)
	}
}

func TestBoltKeyValueDatabase_GetAll(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kvtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := NewBoltKeyValueDatabase(dbPath, "test_bucket")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	data := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}

	for k, v := range data {
		if err := db.Set([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Set failed for %s: %v", k, err)
		}
	}

	foundData := make(map[string]string)
	err = db.GetAll(func(k, v []byte) error {
		foundData[string(k)] = string(v)
		return nil
	})
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(foundData) != len(data) {
		t.Fatalf("Expected %d keys, got %d", len(data), len(foundData))
	}

	for k, v := range data {
		if foundData[k] != v {
			t.Errorf("Expected value %s for key %s, got %s", v, k, foundData[k])
		}
	}
}

func TestBoltKeyValueDatabase_Prune(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kvtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := NewBoltKeyValueDatabase(dbPath, "test_bucket")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	err = db.SetBatch("sub1", map[string]any{"k1": "v1", "k2": "v2"}, func(k string, v any) ([]byte, []byte, error) {
		return []byte(k), []byte(v.(string)), nil
	})
	if err != nil {
		t.Fatalf("SetBatch failed: %v", err)
	}

	err = db.Prune(func(key []byte) (bool, error) {
		if string(key) == "k1" {
			return true, nil 
		}
		return false, nil 
	})
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}

	exists, err := db.HasDataInRange("sub1", []byte("k1"), []byte("k1"))
	if err != nil {
		t.Fatalf("HasDataInRange failed: %v", err)
	}
	if exists {
		t.Errorf("Expected k1 to be pruned from sub1, but it still exists")
	}

	exists, err = db.HasDataInRange("sub1", []byte("k2"), []byte("k2"))
	if err != nil {
		t.Fatalf("HasDataInRange failed for k2: %v", err)
	}
	if !exists {
		t.Errorf("Expected k2 to still exist in sub1")
	}
}
