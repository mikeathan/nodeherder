package store

import (
	"context"
	metrics "node-herder/internal/metrics/models"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
)

func NewMetricsCleanupTask(ctx context.Context, repo metrics.Repository) settings.Task {
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

func DefaultMetricsCleanupTask(ctx context.Context, repo metrics.Repository) settings.Task {
	return &MetricsCleanupTask{
		clock: utils.NewRealClock(),
		repo:  repo,
		ctx:   ctx,
		wg:    sync.WaitGroup{},
	}
}

func (t *MetricsCleanupTask) Start(config *settings.AppConfig) error {
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
				t.clock.Sleep(config.Hub.History.SleepTimeout.Duration())
				utils.LogInfo("Start metrics cleanup")

				err := t.repo.Prune(config.Hub.History.ExpireAt.Duration())
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

type RemoteLoggerTask struct {
	mutex sync.RWMutex
}

func DefaultRemoteLoggerTask() settings.Task {
	return &RemoteLoggerTask{
		mutex: sync.RWMutex{},
	}
}
func (t *RemoteLoggerTask) Start(config *settings.AppConfig) error {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	utils.EnableRemoteLoggerHook(config.Hub.Logger.EnableRemoteLogger)

	return nil
}

func (t *RemoteLoggerTask) Stop() error {
	return nil
}
