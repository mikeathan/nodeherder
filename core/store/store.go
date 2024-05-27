package store

import (
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/utils"
)

type AppStore interface {
	History() metrics.Repository
	Devices() devices.Repository
	Config() settings.Repository
	StoreDevice(device *devices.Device) error
}

type appStore struct {
	history    metrics.Repository
	devices    devices.Repository
	settings   settings.Repository
	appConfig  *settings.AppConfig
	historyMap map[string]bool
}

func newAppStore(devices devices.Repository, metrics metrics.Repository, settings settings.Repository) (AppStore, error) {

	appconfig, err := settings.Load()
	if err != nil {
		return nil, err
	}

	var historyMap map[string]bool = make(map[string]bool)
	for _, config := range appconfig.Devices {
		historyMap[config.Id] = config.History
	}

	return &appStore{
		history:    metrics,
		devices:    devices,
		settings:   settings,
		appConfig:  appconfig,
		historyMap: historyMap,
	}, nil
}

func (s *appStore) StoreDevice(device *devices.Device) error {
	err := s.devices.Store(device.Id, device)
	if err != nil {
		return err
	}

	if enabled, ok := s.historyMap[device.Id]; ok && enabled {
		err := s.history.Store(device)
		if err != nil {
			utils.LogErrorf("storing metrics failed %v", err.Error())
		}
	}

	return nil
}

func (s *appStore) History() metrics.Repository {
	return s.history
}

func (s *appStore) Devices() devices.Repository {
	return s.devices
}

func (s *appStore) Config() settings.Repository {
	return s.settings
}
