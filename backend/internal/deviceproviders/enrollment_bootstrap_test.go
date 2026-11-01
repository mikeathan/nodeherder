package deviceproviders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenConfiguredEnrollmentsMissingKeyFailsClosed(t *testing.T) {
	data, keys := t.TempDir(), t.TempDir()
	if err := os.Chmod(keys, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEVICE_PROVIDER_KEYRING_FILE", "")
	t.Setenv("DEVICE_PROVIDER_KEYRING_DIR", keys)
	r, err := OpenConfiguredEnrollments(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(keys, "device-provider-keyring.json")); err != nil {
		t.Fatal(err)
	}
	if r, err := OpenConfiguredEnrollments(data); err == nil {
		_ = r.Close()
		t.Fatal("missing key was silently replaced")
	}
}

func TestOpenConfiguredEnrollmentsRejectsRelativeKeyDirectory(t *testing.T) {
	t.Setenv("DEVICE_PROVIDER_KEYRING_FILE", "")
	t.Setenv("DEVICE_PROVIDER_KEYRING_DIR", "relative")
	if r, err := OpenConfiguredEnrollments(t.TempDir()); err == nil {
		_ = r.Close()
		t.Fatal("relative key directory accepted")
	}
}
