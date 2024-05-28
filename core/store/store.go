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
	StoreDevice(id string, device *devices.Device) error
	DeviceUpdated(id string, device *devices.Device) error
}

type appStore struct {
	history        metrics.Repository
	devices        devices.Repository
	config         settings.Repository
	deviceConfigs  map[string]*settings.DeviceConfig
	deviceIdMapper map[string]string
}

func newAppStore(devices devices.Repository, metrics metrics.Repository, config settings.Repository) (AppStore, error) {

	appconfig, err := config.Load()
	if err != nil {
		return nil, err
	}

	deviceConfigs := make(map[string]*settings.DeviceConfig)
	for _, dev := range appconfig.Devices {
		deviceConfigs[dev.Id] = dev
	}

	return &appStore{
		history:        metrics,
		devices:        devices,
		config:         config,
		deviceConfigs:  deviceConfigs,
		deviceIdMapper: map[string]string{},
	}, nil
}

func (s *appStore) DeviceUpdated(id string, device *devices.Device) error {
	err := s.StoreDevice(id, device)
	if err != nil {
		return err
	}

	if config, ok := s.deviceConfigs[id]; ok && config.History {
		err := s.history.Store(device)
		if err != nil {
			utils.LogErrorf("storing metrics failed %v", err.Error())
		}
	}

	return nil
}

func (s *appStore) StoreDevice(id string, device *devices.Device) error {
	err := s.devices.Store(id, device)
	if err != nil {
		return err
	}

	if config, ok := s.deviceConfigs[id]; ok && config.History {
		err := s.history.Store(device)
		if err != nil {
			utils.LogErrorf("storing metrics failed %v", err.Error())
		}
	}
	return nil
}

// func (s *appStore) FindDevice(id string) error {
// 	err := s.devices.Store(device.Id, device)
// 	if err != nil {
// 		return err
// 	}

// 	if enabled, ok := s.historyMap[device.Id]; ok && enabled {
// 		err := s.history.Store(device)
// 		if err != nil {
// 			utils.LogErrorf("storing metrics failed %v", err.Error())
// 		}
// 	}

//		return nil
//	}

func (s *appStore) History() metrics.Repository {
	return s.history
}

func (s *appStore) Devices() devices.Repository {
	return s.devices
}
