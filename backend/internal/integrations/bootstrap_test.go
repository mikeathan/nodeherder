package integrations

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"node-herder/repository"
)

func TestBootstrapMissingKeyIsDegraded(t *testing.T) {
	t.Setenv("INTEGRATION_KEYRING_FILE", t.TempDir()+"/absent.json")
	service := Bootstrap(context.Background())
	catalog, e := service.Catalog()
	if e != nil || catalog.Available || catalog.Error == nil || catalog.Error.Code != "key_unavailable" {
		t.Fatal(catalog, e)
	}
	if e = service.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
}

func TestBootstrapNativeFirstRun(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("INTEGRATION_KEYRING_FILE", "")
	t.Setenv("INTEGRATION_KEYRING_DIR", "")
	t.Setenv("DATA_DIR", root+"/data")
	t.Setenv("INTEGRATIONS_CONFIG_DIR", root+"/definitions")
	for n := 0; n < 2; n++ {
		service := Bootstrap(context.Background())
		catalog, err := service.Catalog()
		if err != nil || !catalog.Available || len(catalog.Providers) != 1 {
			t.Fatalf("native startup %d: %+v, %v", n, catalog, err)
		}
		if err := service.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	keys, err := bootstrapKeyring(root + "/data")
	if err != nil || keys.ActiveKeyID == "" {
		t.Fatal("native key not reused", err)
	}
}

func TestBootstrapManagedDirectorySurvivesRestart(t *testing.T) {
	root := t.TempDir()
	volume := filepath.Join(root, "volume")
	if err := os.Mkdir(volume, 0755); err != nil {
		t.Fatal(err)
	}
	// Match Docker's volume root even when the test runner has a stricter umask.
	if err := os.Chmod(volume, 0755); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(volume, "managed")
	data := filepath.Join(root, "data")
	t.Setenv("INTEGRATION_KEYRING_FILE", "")
	t.Setenv("INTEGRATION_KEYRING_DIR", managed)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "") // Distroless must not depend on a writable/persistent HOME.
	t.Setenv("DATA_DIR", data)
	t.Setenv("INTEGRATIONS_CONFIG_DIR", filepath.Join(root, "configs"))
	var original []byte
	for attempt := 0; attempt < 2; attempt++ {
		service := Bootstrap(context.Background())
		catalog, err := service.Catalog()
		if err != nil || !catalog.Available {
			t.Fatalf("startup %d: %+v, %v", attempt, catalog, err)
		}
		if err := service.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
		keyPath := filepath.Join(managed, "integration-keyring.json")
		content, err := os.ReadFile(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		keyInfo, _ := os.Stat(keyPath)
		directoryInfo, _ := os.Stat(managed)
		if keyInfo.Mode().Perm() != 0600 || directoryInfo.Mode().Perm() != 0700 {
			t.Fatal("managed key permissions")
		}
		if attempt == 0 {
			original = content
		} else if !bytes.Equal(original, content) {
			t.Fatal("restart replaced key")
		}
		keys, err := repository.ReadIntegrationKeyring(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		repo, err := repository.OpenIntegrationRepository(data, keys)
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 0 {
			err = repo.Update(func(tx *repository.IntegrationTx) error {
				return tx.Put("identities", "fixture", map[string]string{"owner": "persisted", "kind": "resource"})
			})
		} else {
			err = repo.View(func(tx *repository.IntegrationTx) error {
				var saved map[string]string
				found, err := tx.Get("identities", "fixture", &saved)
				if err == nil && (!found || saved["owner"] != "persisted") {
					t.Error("encrypted state was not retained")
				}
				return err
			})
		}
		if closeErr := repo.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestBootstrapManagedDirectoryRejectsUnsafeRecovery(t *testing.T) {
	for _, scenario := range []string{"existing_database", "malformed_key", "public_directory", "relative_directory", "explicit_missing"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			managed := filepath.Join(root, "managed")
			keyPath := filepath.Join(managed, "integration-keyring.json")
			data := filepath.Join(root, "data")
			t.Setenv("INTEGRATION_KEYRING_FILE", "")
			t.Setenv("INTEGRATION_KEYRING_DIR", managed)
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "native-config"))
			t.Setenv("DATA_DIR", data)
			t.Setenv("INTEGRATIONS_CONFIG_DIR", filepath.Join(root, "configs"))
			var beforeKey, beforeDB []byte
			switch scenario {
			case "existing_database":
				if err := os.Mkdir(data, 0700); err != nil {
					t.Fatal(err)
				}
				beforeDB = []byte("existing encrypted data must not change")
				if err := os.WriteFile(filepath.Join(data, "integrations.db"), beforeDB, 0600); err != nil {
					t.Fatal(err)
				}
			case "malformed_key":
				if err := os.Mkdir(managed, 0700); err != nil {
					t.Fatal(err)
				}
				beforeKey = []byte("invalid-private-content")
				if err := os.WriteFile(keyPath, beforeKey, 0600); err != nil {
					t.Fatal(err)
				}
			case "public_directory":
				if err := os.Mkdir(managed, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(managed, 0755); err != nil {
					t.Fatal(err)
				}
			case "relative_directory":
				t.Setenv("INTEGRATION_KEYRING_DIR", "relative-managed-key")
			case "explicit_missing":
				t.Setenv("INTEGRATION_KEYRING_FILE", filepath.Join(root, "explicit-absent.json"))
			}
			service := Bootstrap(context.Background())
			catalog, err := service.Catalog()
			if err != nil || catalog.Available || catalog.Error == nil || catalog.Error.Code != "key_unavailable" {
				t.Fatal(catalog, err)
			}
			if err := service.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			afterKey, _ := os.ReadFile(keyPath)
			if !bytes.Equal(beforeKey, afterKey) {
				t.Fatal("key was created/replaced during unsafe recovery")
			}
			if beforeDB != nil {
				afterDB, _ := os.ReadFile(filepath.Join(data, "integrations.db"))
				if !bytes.Equal(beforeDB, afterDB) {
					t.Fatal("database changed during key recovery")
				}
			}
		})
	}
}

func TestBootstrapExplicitKeyTakesPrecedence(t *testing.T) {
	root := t.TempDir()
	keyPath := filepath.Join(root, "external", "integration-keyring.json")
	data := filepath.Join(root, "data")
	if _, err := repository.LoadOrCreateIntegrationKeyring(keyPath, filepath.Join(data, "integrations.db")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("INTEGRATION_KEYRING_FILE", keyPath)
	t.Setenv("INTEGRATION_KEYRING_DIR", "ignored-relative-managed-path")
	t.Setenv("DATA_DIR", data)
	t.Setenv("INTEGRATIONS_CONFIG_DIR", filepath.Join(root, "configs"))
	service := Bootstrap(context.Background())
	defer service.Close(context.Background())
	catalog, err := service.Catalog()
	if err != nil || !catalog.Available {
		t.Fatal(catalog, err)
	}
}

func TestStartupFailureReasonsDoNotLeakRawValues(t *testing.T) {
	const private = "sensitive-provider-payload"
	for _, scenario := range []struct {
		name  string
		cause error
		want  string
	}{
		{"missing", fmt.Errorf("%s: %w", private, os.ErrNotExist), "missing"},
		{"permission", fmt.Errorf("%s: %w", private, os.ErrPermission), "permission denied"},
		{"invalid", errors.New(private), "existing database requires its original key"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			message := startupFailureReason(scenario.cause)
			if strings.Contains(message, private) || !strings.Contains(message, scenario.want) {
				t.Fatalf("unsafe or unhelpful diagnostic: %s", message)
			}
		})
	}
}

// Opt-in production-runtime fixture: the caller owns three disposable named volumes and must
// disable container networking. Fixed paths prevent accidental tests against a user's data.
func TestBootstrapDockerVolumeLifecycle(t *testing.T) {
	phase := os.Getenv("NODEHERDER_INTEGRATIONS_VOLUME_TEST")
	if phase == "" {
		t.Skip("disposable Docker volume fixture only")
	}
	if phase != "write" && phase != "read" && phase != "missing" {
		t.Fatal("unknown fixture phase")
	}
	if os.Getenv("DATA_DIR") != "/fixture/data" || os.Getenv("INTEGRATIONS_CONFIG_DIR") != "/fixture/configs" || os.Getenv("INTEGRATION_KEYRING_DIR") != "/fixture/keys/managed" {
		t.Fatal("this test requires isolated fixed-path fixture volumes")
	}
	const keyPath = "/fixture/keys/managed/integration-keyring.json"
	if explicit := os.Getenv("INTEGRATION_KEYRING_FILE"); explicit != "" && explicit != keyPath {
		t.Fatal("unexpected explicit key fixture")
	}
	const dbPath = "/fixture/data/integrations.db"
	beforeDB, _ := os.ReadFile(dbPath)
	service := Bootstrap(context.Background())
	catalog, err := service.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr := service.Close(context.Background()); closeErr != nil {
		t.Fatal(closeErr)
	}
	if phase == "missing" {
		if len(beforeDB) == 0 || catalog.Available || catalog.Error == nil || catalog.Error.Code != "key_unavailable" {
			t.Fatal("missing key did not fail closed", catalog)
		}
		if _, err := os.Lstat(keyPath); !os.IsNotExist(err) {
			t.Fatal("replacement key was created", err)
		}
		afterDB, err := os.ReadFile(dbPath)
		if err != nil || !bytes.Equal(beforeDB, afterDB) {
			t.Fatal("missing key changed existing encrypted data", err)
		}
		return
	}
	if !catalog.Available {
		t.Fatal("managed fixture could not start", catalog)
	}
	keys, err := repository.ReadIntegrationKeyring(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Dir(keyPath))
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatal("key directory must be private", err)
	}
	repo, err := repository.OpenIntegrationRepository("/fixture/data", keys)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if phase == "write" {
		err = repo.Update(func(tx *repository.IntegrationTx) error {
			return tx.Put("identities", "docker-fixture", map[string]string{"kind": "resource", "owner": keys.ActiveKeyID})
		})
	} else {
		err = repo.View(func(tx *repository.IntegrationTx) error {
			var record map[string]string
			found, err := tx.Get("identities", "docker-fixture", &record)
			if err == nil && (!found || record["owner"] != keys.ActiveKeyID) {
				t.Error("container replacement lost encrypted data or key identity")
			}
			return err
		})
	}
	if err != nil {
		t.Fatal(err)
	}
}
