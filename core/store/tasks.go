package store

import (
	"context"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
)

type Task interface {
	Start(cfg *settings.HistoryConfig) error
	Stop() error
}

func NewMetricsCleanupTask(ctx context.Context, repo metrics.Repository) Task {
	return &MetricsCleanupTask{
		clock: utils.NewRealClock(),
		repo:  repo,
		ctx:   ctx,
		wg:    sync.WaitGroup{},
	}
}

type MetricsCleanupTask struct {
	clock utils.Clock
	repo  metrics.Repository
	ctx   context.Context
	wg    sync.WaitGroup
}

func DefaultMetricsCleanupTask(ctx context.Context, repo metrics.Repository) Task {
	return &MetricsCleanupTask{
		clock: utils.NewRealClock(),
		repo:  repo,
		ctx:   ctx,
		wg:    sync.WaitGroup{},
	}
}

func (t *MetricsCleanupTask) Start(config *settings.HistoryConfig) error {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		utils.LogInfo("Metrics cleanup task started")
		for {
			select {
			case <-t.ctx.Done():
				utils.LogInfo("MetricsCleanupTask received cancel signal")
				return

			default:
				t.clock.Sleep(config.SleepTimeout)
				utils.LogInfo("Start metrics cleanup")

				err := t.repo.Prune(config.ExpireAt)
				if err != nil {
					utils.LogErrorf("Error during metrics cleanup: %v", err)
				}
				utils.LogInfo("End metrics cleanup")
			}
		}
	}()

	return nil
}

func (t *MetricsCleanupTask) Stop() error {
	t.wg.Done()

	utils.LogInfo("Metrics cleanup task stopped")

	return nil
}
