package repository

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"testing"
)

func testIntegrationKeys(id string) IntegrationKeyring {
	return IntegrationKeyring{ActiveKeyID: id, Keys: map[string]string{"a": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)), "b": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32))}}
}
func TestIntegrationRepository(t *testing.T) {
	dir := t.TempDir()
	r, e := OpenIntegrationRepository(dir, testIntegrationKeys("a"))
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	if e = r.Update(func(tx *IntegrationTx) error {
		return tx.Put("instances", "one", map[string]string{"postcode": "PRIVATE_POSTCODE"})
	}); e != nil {
		t.Fatal(e)
	}
	if e = r.Update(func(tx *IntegrationTx) error {
		if e := tx.Put("instances", "two", map[string]string{"secret": "hidden"}); e != nil {
			return e
		}
		return errors.New("abort")
	}); e == nil {
		t.Fatal("no rollback")
	}
	e = r.View(func(tx *IntegrationTx) error {
		var v map[string]string
		ok, e := tx.Get("instances", "two", &v)
		if ok {
			t.Fatal("rollback leaked")
		}
		return e
	})
	if e != nil {
		t.Fatal(e)
	}
	var backup bytes.Buffer
	if e = r.Backup(&backup); e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(backup.Bytes(), []byte("PRIVATE_POSTCODE")) {
		t.Fatal("plaintext backup")
	}
	if e = r.Close(); e != nil {
		t.Fatal(e)
	}
	r, e = OpenIntegrationRepository(dir, testIntegrationKeys("b"))
	if e != nil {
		t.Fatal(e)
	}
	if n, e := r.Rotate(1); e != nil || n != 1 {
		t.Fatal(n, e)
	}
	r.Close()
	keys := testIntegrationKeys("b")
	delete(keys.Keys, "a")
	r, e = OpenIntegrationRepository(dir, keys)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	if e = r.View(func(tx *IntegrationTx) error {
		var v map[string]string
		ok, e := tx.Get("instances", "one", &v)
		if !ok || v["postcode"] != "PRIVATE_POSTCODE" {
			t.Fatal("restart lost record")
		}
		return e
	}); e != nil {
		t.Fatal(e)
	}
	restore := t.TempDir()
	if e = os.WriteFile(filepath.Join(restore, "integrations.db"), backup.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = OpenIntegrationRepository(restore, keys); e == nil {
		t.Fatal("missing old key accepted")
	}
}
func TestIntegrationRepositorySchemaAndAAD(t *testing.T) {
	dir := t.TempDir()
	r, e := OpenIntegrationRepository(dir, testIntegrationKeys("a"))
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Update(func(tx *IntegrationTx) error { return tx.Put("instances", "one", map[string]string{"token": "secret"}) }); e != nil {
		t.Fatal(e)
	}
	if e = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("instances"))
		return b.Put([]byte("two"), b.Get([]byte("one")))
	}); e != nil {
		t.Fatal(e)
	}
	if e = r.View(func(tx *IntegrationTx) error { var v any; _, e := tx.Get("instances", "two", &v); return e }); e == nil {
		t.Fatal("record relocation accepted")
	}
	r.Close()
	newer := t.TempDir()
	db, e := bolt.Open(filepath.Join(newer, "integrations.db"), 0600, nil)
	if e != nil {
		t.Fatal(e)
	}
	e = db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucket([]byte("metadata"))
		if e != nil {
			return e
		}
		return b.Put([]byte("schema"), []byte("99"))
	})
	db.Close()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = OpenIntegrationRepository(newer, testIntegrationKeys("a")); e == nil {
		t.Fatal("new schema accepted")
	}
}

func TestIntegrationRepositoryPreservesOpaqueNumbers(t *testing.T) {
	r, e := OpenIntegrationRepository(t.TempDir(), testIntegrationKeys("a"))
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	input := map[string]any{"state": map[string]any{"opaque": json.Number("9007199254740993")}}
	if e = r.Update(func(tx *IntegrationTx) error { return tx.Put("operations", "opaque", input) }); e != nil {
		t.Fatal(e)
	}
	if e = r.View(func(tx *IntegrationTx) error {
		var decoded map[string]any
		if _, e := tx.Get("operations", "opaque", &decoded); e != nil {
			return e
		}
		n, ok := decoded["state"].(map[string]any)["opaque"].(json.Number)
		if !ok || n.String() != "9007199254740993" {
			t.Fatal("opaque number rounded", n)
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
}
