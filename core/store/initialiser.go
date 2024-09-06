package store

import (
	"context"
	"fmt"
	"node-herder/models/metrics"
	"node-herder/repository"
)

func Create(ctx context.Context) (AppStore, error) {

	devicesRepo := repository.NewMemoryDeviceRepo()
	metricsRepo, err := repository.NewMetricsRepo()

	// pass task to metrics repo

	// is silly task takes metrics repo and then we add it to it.
	// re think,
	//and also find a way to start it
	metricsCleanupTask := metrics.NewCleanupTask(ctx, metricsRepo, metrics.DefaultCleanupConfig())

	metricsRepo.AddTask(metricsCleanupTask)
	if err != nil {
		return nil, fmt.Errorf("loading metrics repository failed: %v", err.Error())
	}

	settings, err := repository.NewFileSettingsRepo()
	if err != nil {
		return nil, fmt.Errorf("loading settings repository failed: %v", err.Error())
	}

	return NewAppStore(devicesRepo, metricsRepo, settings)
}
