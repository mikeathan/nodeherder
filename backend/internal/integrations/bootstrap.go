package integrations

// Opens integration storage and configuration, or reports a safe degraded startup.

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"node-herder/repository"
	"node-herder/utils"
)

// Bootstrap initializes private storage without provider I/O. Managed native/Docker key
// directories provision first installs only; explicit key files must already exist. An existing
// encrypted DB always requires its original key. Failures disable this feature, not device startup.
func Bootstrap(ctx context.Context) *Service {
	data := os.Getenv("DATA_DIR")
	if data == "" {
		data = "./data"
	}
	keys, e := bootstrapKeyring(data)
	if e != nil {
		return degradedStartup("key_unavailable", e)
	}
	repo, e := repository.OpenIntegrationRepository(data, keys)
	if e != nil {
		return degradedStartup("storage_unavailable", e)
	}
	config := os.Getenv("INTEGRATIONS_CONFIG_DIR")
	if config == "" {
		config = "./configs/integrations"
	}
	files, e := repository.OpenIntegrationDefinitions(config)
	if e != nil {
		_ = repo.Close()
		return degradedStartup("config_directory_unwritable", e)
	}
	service, e := NewService(ctx, Options{Repository: repo, Definitions: files})
	if e != nil {
		_ = repo.Close()
		return degradedStartup("storage_unavailable", e)
	}
	return service
}

// bootstrapKeyring gives explicit externally managed files precedence over managed directories.
// Docker selects a private child in a persistent volume; native runs retain their OS config path.
// LoadOrCreateIntegrationKeyring refuses to replace a missing key if a DB already exists.
func bootstrapKeyring(data string) (repository.IntegrationKeyring, error) {
	if path := os.Getenv("INTEGRATION_KEYRING_FILE"); path != "" {
		return repository.ReadIntegrationKeyring(path)
	}
	if directory := os.Getenv("INTEGRATION_KEYRING_DIR"); directory != "" {
		if !filepath.IsAbs(directory) {
			return repository.IntegrationKeyring{}, errors.New("managed integration key directory must be absolute")
		}
		return repository.LoadOrCreateIntegrationKeyring(
			filepath.Join(directory, "integration-keyring.json"),
			filepath.Join(data, "integrations.db"),
		)
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

// degradedStartup reports only classified operational reasons: raw config/decoder errors could
// contain private values. Existing stable failure codes remain available to the authenticated UI.
func degradedStartup(code string, cause error) *Service {
	utils.LogErrorf("Integrations startup unavailable [%s]: %s", code, startupFailureReason(cause))
	return DegradedService(code)
}

func startupFailureReason(cause error) string {
	switch {
	case errors.Is(cause, os.ErrNotExist):
		return "a required key or storage path is missing; check mounts or restore the matching key backup"
	case errors.Is(cause, os.ErrPermission):
		return "permission denied; check the configured key, data and provider-directory permissions"
	default:
		return "check key validity, private directory permissions and persistent storage; an existing database requires its original key"
	}
}
