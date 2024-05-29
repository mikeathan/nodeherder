package store

import (
	"fmt"
	"node-herder/repository"
)

func Create() (AppStore, error) {
	devices, err := repository.NewFileDeviceRepo()
	if err != nil {
		return nil, fmt.Errorf("loading device repository failed: %v", err.Error())
	}

	metrics, err := repository.NewMetricsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading metrics repository failed: %v", err.Error())
	}

	settings, err := repository.NewFileSettingsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading settings repository failed: %v", err.Error())
	}

	return NewAppStore(devices, metrics, settings)
}
