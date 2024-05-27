package utils_test

import (
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/repository"
	"node-herder/store"
)

func CreateStore() store.AppStore {
	repo := repository.NewMemoryDeviceRepo()
	return mocks.NewMockAppStoreFromDevicesRepo(repo)
}

func CreateStoreFromDeviceRepo(repo devices.Repository) store.AppStore {
	return mocks.NewMockAppStoreFromDevicesRepo(repo)
}
