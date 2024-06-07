package store

import (
	"fmt"
	"node-herder/repository"
)

func Create() (AppStore, error) {

	devices := repository.NewMemoryDeviceRepo()
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
