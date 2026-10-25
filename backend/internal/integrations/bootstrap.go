package integrations

import (
	"context"
	"os"

	"node-herder/repository"
)

// Bootstrap keeps unavailable optional Connections storage independent of device management.
// Provision keys explicitly outside the checkout; runtime never generates replacement keys.
func Bootstrap(ctx context.Context) *Service {
	keyPath := os.Getenv("INTEGRATION_KEYRING_FILE")
	if keyPath == "" {
		keyPath = "/run/nodeherder-secrets/integration-keyring.json"
	}
	keys, e := repository.ReadIntegrationKeyring(keyPath)
	if e != nil {
		return DegradedService("key_unavailable")
	}
	data := os.Getenv("DATA_DIR")
	if data == "" {
		data = "./data"
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
