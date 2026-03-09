package store

import (
	"context"
	"fmt"
	"node-herder/models/settings"
	"node-herder/repository"
)

func Create(ctx context.Context) (AppStore, error) {

	devicesRepo := repository.NewMemoryDeviceRepo()
	metricsRepo, err := repository.NewMetricsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading metrics repository failed: %v", err.Error())
	}

	// To Refactor
	configRepo, err := repository.NewFileSettingsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading settings repository failed: %v", err.Error())
	}
	// Build Tasks
	tasks := []settings.Task{DefaultMetricsCleanupTask(ctx, metricsRepo), DefaultRemoteLoggerTask()}

	configCache, err := settings.NewAppConfigCache(configRepo, tasks)
	if err != nil {
		return nil, fmt.Errorf("loading settings cache failed: %v", err.Error())
	}
	///

	return NewAppStore(devicesRepo, metricsRepo, configCache)
}
