package metrics

import (
	"context"
	"node-herder/utils"
	"sync"
	"time"
)

type Task interface {
	Start()
	Stop()
}

type CleanupConfig struct {
	SleepTimeout time.Duration
	ExpireAt     time.Duration
}

func NewCleanupConfig(sleepTimeout time.Duration, expireAt time.Duration) *CleanupConfig {
	return &CleanupConfig{
		SleepTimeout: sleepTimeout,
		ExpireAt:     expireAt,
	}
}

func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		SleepTimeout: time.Hour * 12,
		ExpireAt:     time.Hour * 24 * 10,
	}
}

func (c *CleanupConfig) SetExpireAt(expireAt time.Duration) {
	c.ExpireAt = expireAt
}

func (c *CleanupConfig) SetSleepTimeout(sleepTimeout time.Duration) {
	c.SleepTimeout = sleepTimeout
}

type CleanupTask struct {
	clock  utils.Clock
	repo   Repository
	ctx    context.Context
	wg     sync.WaitGroup
	config *CleanupConfig
}

func NewCleanupTask(ctx context.Context, repo Repository, config *CleanupConfig) Task {
	return &CleanupTask{
		clock:  utils.NewRealClock(),
		repo:   repo,
		ctx:    ctx,
		wg:     sync.WaitGroup{},
		config: config,
	}
}

func (t *CleanupTask) Start() {
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
}

func (t *CleanupTask) Stop() {
	t.wg.Wait()
}
