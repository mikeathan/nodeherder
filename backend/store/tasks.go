package store

import (
	"context"
	metrics "node-herder/internal/metrics/domain"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
	"time"
)

func NewMetricsCleanupTask(ctx context.Context, repo metrics.Repository) settings.Task {
	return &MetricsCleanupTask{
		clock: utils.NewRealClock(),
		repo:  repo,
		ctx:   ctx,
	}
}

type MetricsCleanupTask struct {
	clock  utils.Clock
	repo   metrics.Repository
	ctx    context.Context
	mutex  sync.Mutex // serializes worker replacement and shutdown
	cancel context.CancelFunc
	done   chan struct{}
}

type cleanupDurations struct {
	sleep  time.Duration
	expire time.Duration
}

func DefaultMetricsCleanupTask(ctx context.Context, repo metrics.Repository) settings.Task {
	return NewMetricsCleanupTask(ctx, repo)
}

func (t *MetricsCleanupTask) Start(config *settings.AppConfig) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.stopLocked()
	if err := t.ctx.Err(); err != nil {
		return err
	}
	// Each worker owns immutable durations; Start joins the predecessor first.
	history := cleanupDurations{
		sleep:  config.Hub.History.SleepTimeout.Duration(),
		expire: config.Hub.History.ExpireAt.Duration(),
	}
	ctx, cancel := context.WithCancel(t.ctx)
	done := make(chan struct{})
	t.cancel, t.done = cancel, done

	go func() {
		defer close(done)
		defer cancel()

		utils.LogInfo("Metrics cleanup task started")
		for {
			wake := make(chan struct{}, 1)
			timer := t.clock.AfterFunc(history.sleep, func() { wake <- struct{}{} })
			select {
			case <-ctx.Done():
				timer.Stop()
				utils.LogInfo("MetricsCleanupTask received cancel signal")
				return
			case <-wake:
				timer.Stop()
			}
			if ctx.Err() != nil {
				return
			}
			utils.LogInfo("Start metrics cleanup")
			if err := t.repo.Prune(history.expire); err != nil {
				utils.LogErrorf("Error during metrics cleanup: %v", err)
			}
			utils.LogInfo("End metrics cleanup")
		}
	}()

	return nil
}

func (t *MetricsCleanupTask) Stop() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.stopLocked()

	utils.LogInfo("Metrics cleanup task stopped")

	return nil
}

// Caller owns mutex; the worker never needs it to finish.
func (t *MetricsCleanupTask) stopLocked() {
	if t.cancel == nil {
		return
	}
	t.cancel()
	<-t.done
	t.cancel, t.done = nil, nil
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
	utils.SetLogLevel(config.Hub.Logger.Level)

	return nil
}

func (t *RemoteLoggerTask) Stop() error {
	return nil
}
