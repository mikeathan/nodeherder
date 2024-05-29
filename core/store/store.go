package store

import (
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/utils"
)

type AppStore interface {
	StoreDevice(friendlyName string, device *devices.Device) error
	UpdateDevice(friendlyName string, device *devices.Device) error
	FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error)
	FindDeviceById(id string) (*devices.Device, error)
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
	AllDevices() ([]*devices.Device, error)

	StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error
	FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error)
	FindBridgeInfoById(id string) (*devices.BridgeInfo, error)

	ResolveFriendlyName(friendlyName string) string
}

type appStore struct {
	metrics        metrics.Repository
	devices        devices.Repository
	config         settings.Repository
	deviceConfigs  map[string]*settings.DeviceConfig
	deviceIdMapper *repository.DeviceIdMapper
}

func NewAppStore(devices devices.Repository, metrics metrics.Repository, config settings.Repository) (AppStore, error) {

	appconfig, err := config.Load()
	if err != nil {
		return nil, err
	}

	deviceConfigs := make(map[string]*settings.DeviceConfig)
	for _, dev := range appconfig.Devices {
		deviceConfigs[dev.Id] = dev
	}

	return &appStore{
		metrics:        metrics,
		devices:        devices,
		config:         config,
		deviceConfigs:  deviceConfigs,
		deviceIdMapper: repository.NewDeviceIdMapper(devices),
	}, nil
}

func (s *appStore) UpdateDevice(friendlyName string, device *devices.Device) error {

	id := s.ResolveFriendlyName(friendlyName)

	err := s.StoreDevice(id, device)
	if err != nil {
		return err
	}

	if s.IsMetricsEnabled(id) {
		err := s.metrics.Store(device)
		if err != nil {
			utils.LogErrorf("storing metrics failed %v", err.Error())
		}
	}

	s.deviceIdMapper.UpdateId(friendlyName, id)
	return nil
}

func (s *appStore) IsMetricsEnabled(id string) bool {
	if config, ok := s.deviceConfigs[id]; ok && config.MetricsEnabled {
		return true
	}

	return false
}

func (s *appStore) StoreDevice(friendlyName string, device *devices.Device) error {
	id := s.ResolveFriendlyName(friendlyName)

	err := s.devices.Store(id, device)
	if err != nil {
		return err
	}

	s.deviceIdMapper.UpdateId(friendlyName, id)
	return nil
}

func (s *appStore) StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error {
	err := s.devices.StoreBridge(bridgeInfoList)
	if err != nil {
		return err
	}

	s.deviceIdMapper.Configure()
	return nil
}

func (s *appStore) FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error) {

	id := s.ResolveFriendlyName(friendlyName)

	return s.FindDeviceById(id)
}

func (s *appStore) FindDeviceById(id string) (*devices.Device, error) {
	return s.devices.FindDevice(id)
}

func (s *appStore) FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error) {

	id := s.ResolveFriendlyName(friendlyName)
	return s.FindBridgeInfoById(id)
}

func (s *appStore) FindBridgeInfoById(id string) (*devices.BridgeInfo, error) {
	return s.devices.FindBridgeInfo(id)
}

func (s *appStore) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	return s.devices.FindDevices(ids)
}

func (s *appStore) AllDevices() ([]*devices.Device, error) {
	return s.devices.AllDevices()
}

func (a *appStore) ResolveFriendlyName(friendlyName string) string {

	return a.deviceIdMapper.ResolveFriendlyName(friendlyName)
}
