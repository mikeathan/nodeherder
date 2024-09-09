package store

import (
	"context"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
	"time"
)

type Task interface {
	Start(cfg *settings.HistoryConfig) error
	Stop() error
}

type MetricsCleanupConfig struct {
	SleepTimeout time.Duration
	ExpireAt     time.Duration
}

func NewMetricsCleanupConfig(sleepTimeout time.Duration, expireAt time.Duration) *MetricsCleanupConfig {
	return &MetricsCleanupConfig{
		SleepTimeout: sleepTimeout,
		ExpireAt:     expireAt,
	}
}

func DefaultCleanupConfig() *MetricsCleanupConfig {
	return &MetricsCleanupConfig{
		SleepTimeout: time.Hour * 12,
		ExpireAt:     time.Hour * 24 * 10,
	}
}

func NewMetricsCleanupTask(ctx context.Context, repo metrics.Repository, config *MetricsCleanupConfig) Task {
	return &MetricsCleanupTask{
		clock:  utils.NewRealClock(),
		repo:   repo,
		ctx:    ctx,
		wg:     sync.WaitGroup{},
		config: config,
	}
}
func (c *MetricsCleanupConfig) SetExpireAt(expireAt time.Duration) {
	c.ExpireAt = expireAt
}

func (c *MetricsCleanupConfig) SetSleepTimeout(sleepTimeout time.Duration) {
	c.SleepTimeout = sleepTimeout
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

TODO use cfg from args
func (t *MetricsCleanupTask) Start(cfg *settings.HistoryConfig) error {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		for {
			select {
			case <-t.ctx.Done():
				utils.LogInfo("MetricsCleanupTask received cancel signal")
				return

			default:
				t.clock.Sleep(t.config.SleepTimeout)
				utils.LogInfo("Start metrics cleanup")

				err := t.repo.Prune(t.config.ExpireAt)
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
	t.wg.Wait()

	return nil
}
