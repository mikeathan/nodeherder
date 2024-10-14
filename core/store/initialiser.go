package store

import (
	"context"
	"fmt"
	"node-herder/repository"
)

func Create(ctx context.Context) (AppStore, error) {

	devicesRepo := repository.NewMemoryDeviceRepo()
	metricsRepo, err := repository.NewMetricsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading metrics repository failed: %v", err.Error())
	}

	settings, err := repository.NewFileSettingsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading settings repository failed: %v", err.Error())
	}

	// Build Tasks
	tasks := []Task{DefaultMetricsCleanupTask(ctx, metricsRepo), DefaultRemoteLoggerTask()}

	return NewAppStore(devicesRepo, metricsRepo, settings, tasks)
}
