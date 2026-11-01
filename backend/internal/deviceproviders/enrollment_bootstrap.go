package deviceproviders

// This file resolves the dedicated provider keyring and opens enrollment storage.
// It fails closed when an existing database has no matching key and never
// generates replacement keys for existing data.
import (
	"errors"
	"os"
	"path/filepath"

	"node-herder/repository"
)

// OpenConfiguredEnrollments opens persistent provider storage using a separate
// configured keyring. Managed keys are created only if the database is absent;
// missing keys for existing databases fail closed. The caller owns the result.
func OpenConfiguredEnrollments(dataDirectory string) (*EnrollmentRepository, error) {
	if dataDirectory == "" {
		return nil, errors.New("device data directory is required")
	}
	var keys repository.IntegrationKeyring
	var err error
	if path := os.Getenv("DEVICE_PROVIDER_KEYRING_FILE"); path != "" {
		keys, err = repository.ReadIntegrationKeyring(path)
	} else {
		directory := os.Getenv("DEVICE_PROVIDER_KEYRING_DIR")
		if directory != "" && !filepath.IsAbs(directory) {
			return nil, errors.New("device key directory must be absolute")
		}
		if directory == "" {
			directory, err = os.UserConfigDir()
			if err != nil {
				return nil, err
			}
			directory = filepath.Join(directory, "nodeherder")
		}
		keys, err = repository.LoadOrCreateIntegrationKeyring(filepath.Join(directory, "device-provider-keyring.json"), filepath.Join(dataDirectory, "device-providers", "integrations.db"))
	}
	if err != nil {
		return nil, err
	}
	return OpenEnrollments(dataDirectory, keys)
}
