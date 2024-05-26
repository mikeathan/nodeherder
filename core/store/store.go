package store

import (
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
)

type Store interface {
	History() metrics.Repository
	Devices() devices.Repository
	Config() settings.Repository
}

type AppStore struct {
	history metrics.Repository
	devices devices.Repository
	config  settings.Repository
}

func NewAppStore(devices devices.Repository, metrics metrics.Repository, config settings.Repository) Store {
	return &AppStore{
		history: metrics,
		devices: devices,
		config:  config,
	}
}

func (s *AppStore) History() metrics.Repository {
	return s.history
}

func (s *AppStore) Devices() devices.Repository {
	return s.devices
}

func (s *AppStore) Config() settings.Repository {
	return s.config
}
