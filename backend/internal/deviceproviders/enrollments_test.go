package deviceproviders

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"node-herder/repository"
)

func TestEnrollmentRepositoryBackupAndIsolation(t *testing.T) {
	keys := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))}}
	dir := t.TempDir()
	r, err := OpenEnrollments(dir, keys)
	if err != nil {
		t.Fatal(err)
	}
	item := Enrollment{ID: uuid.NewString(), Provider: "esphome", Address: "192.168.50.160", Port: 6053, Credential: "private-test-credential", PhysicalID: "private-device-identity"}
	if err := r.Save(item); err != nil {
		t.Fatal(err)
	}
	duplicate := item
	duplicate.ID = uuid.NewString()
	duplicate.Address = "::ffff:192.168.50.160"
	if err := r.Save(duplicate); err == nil {
		t.Fatal("duplicate endpoint accepted")
	}
	items, err := r.List()
	if err != nil || len(items) != 1 || items[0].Credential != item.Credential {
		t.Fatalf("list: %+v %v", items, err)
	}
	public, err := json.Marshal(items)
	if err != nil || bytes.Contains(public, []byte(item.Credential)) || bytes.Contains(public, []byte(item.PhysicalID)) {
		t.Fatal("private enrollment data exposed in JSON")
	}
	var backup bytes.Buffer
	if err := r.Backup(&backup); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(backup.Bytes(), []byte(item.Credential)) || bytes.Contains(backup.Bytes(), []byte(item.PhysicalID)) {
		t.Fatal("private enrollment data stored unencrypted")
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	restore := t.TempDir()
	if err := os.MkdirAll(filepath.Join(restore, "device-providers"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(restore, "device-providers", "integrations.db"), backup.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	wrong := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32))}}
	if invalid, err := OpenEnrollments(restore, wrong); err == nil {
		_ = invalid.Close()
		t.Fatal("wrong key accepted")
	}
	restored, err := OpenEnrollments(restore, keys)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	items, err = restored.List()
	if err != nil || len(items) != 1 || items[0].Credential != item.Credential {
		t.Fatalf("restore: %+v %v", items, err)
	}
	if err := restored.Delete(item.ID); err != nil {
		t.Fatal(err)
	}
	items, err = restored.List()
	if err != nil || len(items) != 0 {
		t.Fatalf("delete: %+v %v", items, err)
	}
}
