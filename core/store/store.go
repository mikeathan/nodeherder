package store

import (
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
)

type AppStore interface {
	History() metrics.Repository
	Devices() devices.Repository
	Config() settings.Repository
}

type appStore struct {
	history metrics.Repository
	devices devices.Repository
	config  settings.Repository
}

func newAppStore(devices devices.Repository, metrics metrics.Repository, config settings.Repository) AppStore {
	return &appStore{
		history: metrics,
		devices: devices,
		config:  config,
	}
}

func (s *appStore) History() metrics.Repository {
	return s.history
}

func (s *appStore) Devices() devices.Repository {
	return s.devices
}

func (s *appStore) Config() settings.Repository {
	return s.config
}
