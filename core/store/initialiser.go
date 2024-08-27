package store

import (
	"fmt"
	"node-herder/models/metrics"
	"node-herder/repository"
	"node-herder/utils"
)

func Create() (AppStore, error) {

	devices := repository.NewMemoryDeviceRepo()

	keyGenerator := metrics.NewTimestampedKeyGenerator(utils.NewRealClock())
	metrics, err := repository.NewMetricsRepo(keyGenerator)
	if err != nil {
		return nil, fmt.Errorf("loading metrics repository failed: %v", err.Error())
	}

	settings, err := repository.NewFileSettingsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading settings repository failed: %v", err.Error())
	}

	return NewAppStore(devices, metrics, settings)
}
