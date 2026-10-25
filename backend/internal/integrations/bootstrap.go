package integrations

import (
	"context"
	"os"
	"path/filepath"

	"node-herder/repository"
)

// Bootstrap initializes native first-run private storage without provider I/O. An explicit
// deployment key path must already exist; missing keys for an existing DB are never replaced.
func Bootstrap(ctx context.Context) *Service {
	data := os.Getenv("DATA_DIR")
	if data == "" {
		data = "./data"
	}
	keys, e := bootstrapKeyring(data)
	if e != nil {
		return DegradedService("key_unavailable")
	}
	repo, e := repository.OpenIntegrationRepository(data, keys)
	if e != nil {
		return DegradedService("storage_unavailable")
	}
	config := os.Getenv("INTEGRATIONS_CONFIG_DIR")
	if config == "" {
		config = "./configs/integrations"
	}
	files, e := repository.OpenIntegrationDefinitions(config)
	if e != nil {
		_ = repo.Close()
		return DegradedService("config_directory_unwritable")
	}
	service, e := NewService(ctx, Options{Repository: repo, Definitions: files})
	if e != nil {
		_ = repo.Close()
		return DegradedService("storage_unavailable")
	}
	return service
}

// bootstrapKeyring separates explicit deployment custody from the native managed default.
func bootstrapKeyring(data string) (repository.IntegrationKeyring, error) {
	if path := os.Getenv("INTEGRATION_KEYRING_FILE"); path != "" {
		return repository.ReadIntegrationKeyring(path)
	}
	directory, err := os.UserConfigDir()
	if err != nil {
		return repository.IntegrationKeyring{}, err
	}
	return repository.LoadOrCreateIntegrationKeyring(
		filepath.Join(directory, "nodeherder", "integration-keyring.json"),
		filepath.Join(data, "integrations.db"),
	)
}
