package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationDefinitions(t *testing.T) {
	root := t.TempDir()
	r, e := OpenIntegrationDefinitions(root)
	if e != nil {
		t.Fatal(e)
	}
	key := "waste_collection/example"
	if e = r.Write(key, "id: example\n", ""); e != nil {
		t.Fatal(e)
	}
	original, hash, e := r.Read(key)
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Write(key, "replacement", ""); !errors.Is(e, ErrDefinitionConflict) {
		t.Fatal("seed overwrote file", e)
	}
	if e = r.Write(key, "new", "wrong"); !errors.Is(e, ErrDefinitionConflict) {
		t.Fatal("stale accepted", e)
	}
	if e = r.Write(key, "new", hash); e != nil {
		t.Fatal(e)
	}
	_, hash, e = r.Read(key)
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Write("../escape", "bad", ""); e == nil {
		t.Fatal("traversal")
	}
	if e = os.Symlink(filepath.Join(root, "waste_collection", "example.yaml"), filepath.Join(root, "waste_collection", "alias.yaml")); e != nil {
		t.Fatal(e)
	}
	if _, _, e = r.Read("waste_collection/alias"); e == nil {
		t.Fatal("symlink")
	}
	if original != "id: example\n" {
		t.Fatal(original)
	}
	if e = r.Delete(key, hash); e != nil {
		t.Fatal(e)
	}
}
