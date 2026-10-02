package storage

import (
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestInitializationFailureReleasesDatabase(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "test.db")
	if db, err := NewBoltKeyValueDatabase(filename, ""); err == nil {
		_ = db.Close()
		t.Fatal("empty bucket name should fail initialization")
	}
	db, err := NewBoltKeyValueDatabase(filename, "valid")
	if err != nil {
		t.Fatalf("failed initialization did not release database lock: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// Synthetic database created with Bolt v1.3.1 before the driver replacement.
func copyLegacyBoltFixture(t *testing.T) string {
	t.Helper()
	const data = "H4sIAAAAAAAA/+zdvWoUURQH8LOTL4yJpBRswhRWEVSsAqKgCD6DSNjduQlDvmRnNySELXwSCx/AJ7Cx9wUsLK2MlViL2Rs/AgEDxhX392v+98K9M2fZ+pyJbDrn0buFo+Lb4v5oPxO/ms05l7PI+frz+y+XPnTeBAAAAAAAAAAAAAAAAAAAAHBurZxn9f9PnTpfnDp/cv/F3bdXP1759OqCywUAAAAAAAAAAAAAAAAAAID/0kk//9KY6wAAAAAAAAAAAAAAAAAAAIBJdvJ9/2LMdQAAAAAAAAAAAAAAAAAAAMAkm85Z/DQBYDki5iNiISJaEXE9n7sXEVtpo909uLGZDvJqr701SM2gc+usNxTHzxhZiojFPG0g399O/V7dvXP7Yn8mAAAAAAAAAAAAAAAAAAAA/NNmchYxdZytiLgZEbMRMcz7lxExFxHP8/5aK+Jy7ubv9OpqI5319NO9//MR8eD7rcc767tPDss6pbTWrqpeappytcxzAaq0V3dTuVKu9+q0U20drO20t1O5Wj6q9/uDXlrOB4ZPR4vmd2tYjIiHP+YPjG4flnV1/nevlGn/2W6TmnL1cDjsp6a/1hl0N1N/+g/9OwAAAAAAAAAAAAAAAAAAAEyK2ZxL+fv/Rd5Pja0iAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAPj7vgYAAP//83mKIQAAAgA="
	r, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(data)))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	decoded, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "legacy.db")
	if err := os.WriteFile(filename, decoded, 0600); err != nil {
		t.Fatal(err)
	}
	return filename
}

func TestBoltKeyValueDatabase_LegacyReadWriteReopen(t *testing.T) {
	filename := copyLegacyBoltFixture(t)
	for pass := 0; pass < 2; pass++ {
		db, err := NewBoltKeyValueDatabase(filename, "test_bucket")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		value, err := db.Get([]byte("legacy-key"))
		if err != nil || string(value) != "legacy-value" {
			t.Fatalf("legacy root value: %q, %v", value, err)
		}
		found := false
		err = db.ViewInRange("sub1", []byte("legacy-metric"), []byte("legacy-metric"), func(k, v []byte) error {
			found = string(k) == "legacy-metric" && string(v) == "42"
			return nil
		})
		if err != nil || !found {
			t.Fatalf("legacy nested value: found=%v, err=%v", found, err)
		}
		if pass == 0 {
			if err := db.Set([]byte("new-key"), []byte("new-value")); err != nil {
				t.Fatal(err)
			}
			if err := db.SetBatch("sub1", map[string]any{"new-metric": "43"}, func(k string, v any) ([]byte, []byte, error) {
				return []byte(k), []byte(v.(string)), nil
			}); err != nil {
				t.Fatal(err)
			}
		} else {
			v, err := db.Get([]byte("new-key"))
			if err != nil || string(v) != "new-value" {
				t.Fatalf("new root value after reopen: %q, %v", v, err)
			}
			foundNew := false
			err = db.ViewInRange("sub1", []byte("new-metric"), []byte("new-metric"), func(k, v []byte) error {
				foundNew = string(k) == "new-metric" && string(v) == "43"
				return nil
			})
			if err != nil || !foundNew {
				t.Fatalf("new nested value after reopen: found=%v, err=%v", foundNew, err)
			}
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
