package utils_test

import (
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/repository"
	"node-herder/store"
)

func CreateStore() store.AppStore {
	repo := repository.NewMemoryDeviceRepo()
	metricsRepo := mocks.NopMetricsRepo{}
	settingsRepo := mocks.NopSettingsrepo{}
	store, _ := store.NewAppStore(repo, &metricsRepo, &settingsRepo)
	return store
}

func CreateStoreFromDeviceRepo(repo devices.Repository) store.AppStore {
	metricsRepo := mocks.NopMetricsRepo{}
	settingsRepo := mocks.NopSettingsrepo{}
	store, _ := store.NewAppStore(repo, &metricsRepo, &settingsRepo)
	return store
}
