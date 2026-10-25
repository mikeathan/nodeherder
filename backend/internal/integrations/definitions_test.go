package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	model "node-herder/models/integrations"
	"node-herder/repository"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefinitionSaveAndOriginReview(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	d, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	changed := strings.Replace(d.YAML, "https://whenisbins.com", "https://new.example.test", 1)
	v, e := s.ValidateDefinition(i.ProviderKey, changed, nil)
	if e != nil || len(v.Affected) != 1 {
		t.Fatal(v, e)
	}
	if _, e = s.SaveDefinition(i.ProviderKey, changed, d.ETag, false, nil); e == nil {
		t.Fatal("destination self authorized")
	}
	saved, e := s.SaveDefinition(i.ProviderKey, changed, d.ETag, false, &DestinationReview{Origin: v.Origin, Reassociate: v.Affected})
	if e != nil {
		t.Fatal(e)
	}
	if saved.ContentHash != saved.ActiveHash || saved.Revision != d.Revision+1 {
		t.Fatal(saved)
	}
	if _, e = s.SaveDefinition(i.ProviderKey, d.YAML, d.ETag, false, nil); e == nil {
		t.Fatal("stale write accepted")
	}
	i, e = s.Instance(i.ID)
	if e != nil || len(i.Resources) != 1 {
		t.Fatal("definition erased last good data", e)
	}
	if e = s.DeleteDefinition(i.ProviderKey, saved.ETag); e == nil {
		t.Fatal("deleted in-use definition")
	}
}
func TestInvalidHostDefinitionRetainsActive(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	d, e := s.Definition("waste_collection/whenisbins")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.files.Write(d.Key, "bad: [yaml", d.ContentHash); e != nil {
		t.Fatal(e)
	}
	_, hash, e := s.files.Read(d.Key)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ReloadDefinition(d.Key, StrongETag(hash), nil); e == nil {
		t.Fatal("invalid reload accepted")
	}
	after, e := s.Definition(d.Key)
	if e != nil || after.ActiveHash != d.ActiveHash || after.DiskError == "" {
		t.Fatal(after, e)
	}
}

func TestHostOriginChangeCanBeReviewedAfterRestart(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	d, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	changed := strings.Replace(d.YAML, "https://whenisbins.com", "https://new.example.test", 1)
	if e = s.files.Write(d.Key, changed, d.ContentHash); e != nil {
		t.Fatal(e)
	}
	if e = s.loadDefinitions(); e != nil {
		t.Fatal(e)
	}
	i, e = s.Instance(i.ID)
	if e != nil || i.Error == nil {
		t.Fatal("unreviewed host destination acquired grant", e)
	}
	v, e := s.ValidateDefinition(d.Key, changed, nil)
	if e != nil || len(v.Affected) != 1 {
		t.Fatal("cannot review blocked binding", v, e)
	}
	disk, e := s.Definition(d.Key)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SaveDefinition(d.Key, changed, disk.ETag, false, &DestinationReview{Origin: v.Origin, Reassociate: v.Affected}); e != nil {
		t.Fatal(e)
	}
	i, e = s.Instance(i.ID)
	if e != nil || i.Error != nil {
		t.Fatal("review did not reconnect", i, e)
	}
}

type faultDefinitionFiles struct {
	DefinitionFiles
	deny       bool
	afterWrite func() error
}

func (f faultDefinitionFiles) Write(key, source, hash string) error {
	if f.deny {
		return os.ErrPermission
	}
	if e := f.DefinitionFiles.Write(key, source, hash); e != nil {
		return e
	}
	if f.afterWrite != nil {
		return f.afterWrite()
	}
	return nil
}
func TestDefinitionWriteFailureKeepsActiveAndDisk(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	d, e := s.Definition("waste_collection/whenisbins")
	if e != nil {
		t.Fatal(e)
	}
	original := s.files
	s.files = faultDefinitionFiles{DefinitionFiles: original, deny: true}
	if _, e = s.SaveDefinition(d.Key, strings.Replace(d.YAML, "name: WhenIsBins", "name: Changed", 1), d.ETag, false, nil); e == nil || failure(e).Code != "config_directory_unwritable" {
		t.Fatal(e)
	}
	after, e := s.Definition(d.Key)
	if e != nil || after.ContentHash != d.ContentHash || after.ActiveHash != d.ActiveHash || after.Status != "active" {
		t.Fatal(after, e)
	}
}
func TestDefinitionSavedNotAppliedRecoversFromJournal(t *testing.T) {
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	d, e := s.Definition("waste_collection/whenisbins")
	if e != nil {
		t.Fatal(e)
	}
	source := strings.Replace(d.YAML, "name: WhenIsBins", "name: Recovered", 1)
	original := s.files
	var snapshot bytes.Buffer
	s.files = faultDefinitionFiles{DefinitionFiles: original, afterWrite: func() error {
		if e := s.repo.Backup(&snapshot); e != nil {
			return e
		}
		return s.repo.Close()
	}}
	_, e = s.SaveDefinition(d.Key, source, d.ETag, false, nil)
	if e == nil {
		t.Fatal("activation failure was hidden")
	}
	f := failure(e)
	if f.Code != "saved_not_applied" || f.ActiveHash != d.ActiveHash || f.ContentHash != ContentHash([]byte(source)) {
		t.Fatal(f)
	}
	dir := t.TempDir()
	if e = os.WriteFile(filepath.Join(dir, "integrations.db"), snapshot.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	repo, e := repository.OpenIntegrationRepository(dir, repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))}})
	if e != nil {
		t.Fatal(e)
	}
	recovered, e := NewService(context.Background(), Options{Repository: repo, Definitions: original, Clock: clock, HTTP: providerFunc(standardProvider)})
	if e != nil {
		t.Fatal(e)
	}
	defer recovered.Close(context.Background())
	after, e := recovered.Definition(d.Key)
	if e != nil || after.Status != "active" || after.ContentHash != after.ActiveHash || after.ContentHash != f.ContentHash {
		t.Fatal(after, e)
	}
}

func TestDefinitionFlushFailureReportsWrittenFile(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	d, e := s.Definition("waste_collection/whenisbins")
	if e != nil {
		t.Fatal(e)
	}
	source := strings.Replace(d.YAML, "name: WhenIsBins", "name: Written", 1)
	s.files = faultDefinitionFiles{DefinitionFiles: s.files, afterWrite: func() error { return os.ErrPermission }}
	_, e = s.SaveDefinition(d.Key, source, d.ETag, false, nil)
	if e == nil || failure(e).Code != "saved_not_applied" {
		t.Fatal("written file reported as unsaved", e)
	}
	after, e := s.Definition(d.Key)
	if e != nil || after.ContentHash != ContentHash([]byte(source)) || after.ActiveHash != d.ActiveHash || after.Status != "saved_not_applied" {
		t.Fatal(after, e)
	}
}

func TestDestinationReviewDoesNotTransferUnusedToken(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	if _, e := s.SetCredential("instances", i.ID, "api_token", "synthetic-private", i.ETag, false); e != nil {
		t.Fatal(e)
	}
	d, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	changed := strings.ReplaceAll(strings.Replace(d.YAML, "https://whenisbins.com", "https://new.example.test", 1), "secret_slot: api_token", "secret_slot: new_token")
	v, e := s.ValidateDefinition(d.Key, changed, nil)
	if e != nil || len(v.Affected) != 1 || v.Affected[0].CredentialRevision != 0 {
		t.Fatal(v, e)
	}
	if _, e = s.SaveDefinition(d.Key, changed, d.ETag, false, &DestinationReview{Origin: v.Origin, Reassociate: v.Affected}); e != nil {
		t.Fatal(e)
	}
	var stored model.Instance
	if e = s.repo.View(func(tx *repository.IntegrationTx) error { _, e := tx.Get("instances", i.ID, &stored); return e }); e != nil {
		t.Fatal(e)
	}
	if stored.Credentials["api_token"].Origin != "https://whenisbins.com" {
		t.Fatal("input-only review moved an unused token")
	}
	current, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SetCredential("instances", i.ID, "api_token", "", current.ETag, true); e != nil {
		t.Fatal("old token cannot be removed", e)
	}
	back := strings.ReplaceAll(changed, "secret_slot: new_token", "secret_slot: api_token")
	if _, e = s.ValidateDefinition(d.Key, back, nil); e != nil {
		t.Fatal(e)
	}
}

func TestMissingProviderFileCanBeRecreatedWithoutLosingData(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	d, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.files.Delete(d.Key, d.ContentHash); e != nil {
		t.Fatal(e)
	}
	missing, e := s.Definition(d.Key)
	if e != nil || missing.Status != "missing" || missing.YAML != d.YAML {
		t.Fatal(missing, e)
	}
	restored, e := s.SaveDefinition(d.Key, missing.YAML, "", true, nil)
	if e != nil || restored.ContentHash != d.ContentHash {
		t.Fatal(restored, e)
	}
	after, e := s.Instance(i.ID)
	if e != nil || after.Resources[0].ResourceID != i.Resources[0].ResourceID {
		t.Fatal("restoring a file lost instance data", e)
	}
}
