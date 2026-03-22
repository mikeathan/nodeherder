package utils_test

import (
	"fmt"
	metrics "node-herder/internal/metrics/domain"
	metricsrepo "node-herder/internal/metrics/storage"
	"node-herder/mocks"
	"node-herder/models/assistant"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	"node-herder/utils/storage"
	"os"
	"testing"
	"time"
)

func CreateMetricsRepo(filename string) (metrics.Repository, *mocks.MockClock, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filename, "metrics")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise keyvalue db: %s", err.Error())
	}

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})

	repo, err := metricsrepo.NewMetricsRepoFromDatabase(kvdb, mockClock, 0*time.Second)
	if err != nil {
		return nil, nil, err
	}

	return repo, mockClock, nil
}

func CreateMetricsRepoWithTailWindow(filename string, tailWindow time.Duration) (metrics.Repository, *mocks.MockClock, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filename, "metrics")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise keyvalue db: %s", err.Error())
	}

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})

	repo, err := metricsrepo.NewMetricsRepoFromDatabase(kvdb, mockClock, tailWindow)
	if err != nil {
		return nil, nil, err
	}

	return repo, mockClock, nil
}

func CreateMetricsRepoWithClock(filename string, mockClock *mocks.MockClock) (metrics.Repository, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filename, "metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to initialise keyvalue db: %s", err.Error())
	}

	repo, err := metricsrepo.NewMetricsRepoFromDatabase(kvdb, mockClock, 0*time.Second)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func CreateAssistantStore(t *testing.T, assistantRepo assistant.Repository, appConfig *settings.AppConfig) (store.AppStore, func()) {
	settingsTempFile := Tempfile()
	settingsRepo, err := repository.NewFileSettingsRepoFromFile(settingsTempFile)
	if err != nil {
		t.Fatal(err)
	}

	if appConfig != nil {
		settingsRepo.SaveAppConfig(appConfig)
	}

	configCache, err := settings.NewAppConfigCache(settingsRepo, []settings.Task{})
	if err != nil {
		t.Fatal(err)
	}

	metricsRepo := mocks.NopMetricsRepo{}
	deviceRepo := repository.NewMemoryDeviceRepo()

	s, _ := store.NewAppStore(deviceRepo, &metricsRepo, configCache, assistantRepo)

	cleanup := func() {
		os.Remove(settingsTempFile)
		deviceRepo.Close()
	}
	return s, cleanup
}
