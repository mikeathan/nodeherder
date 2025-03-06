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
	config, err := repository.NewFileSettingsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading settings repository failed: %v", err.Error())
	}

	configCache, err := settings.NewAppConfigCache(config)
	if err != nil {
		return nil, fmt.Errorf("loading settings cache failed: %v", err.Error())
	}
	/// 

	// Build Tasks
	tasks := []Task{DefaultMetricsCleanupTask(ctx, metricsRepo), DefaultRemoteLoggerTask()}

	return NewAppStore(devicesRepo, metricsRepo, configCache, tasks)
}
