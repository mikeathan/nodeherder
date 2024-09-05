package store

import (
	"node-herder/models/metrics"
	"node-herder/utils"
	"time"
)

type MetricsCleanupTask struct {
	clock        utils.Clock
	repo         metrics.Repository
	sleepTimeout time.Duration
	expireAt     time.Duration
}


TODO
func (t *MetricsCleanupTask) Start() {

	go func() {
		for {

			t.clock.Sleep(t.sleepTimeout)
			utils.LogInfo("Start metrics cleanup")

			err := t.repo.Prune(t.expireAt)
			if err != nil {
				utils.LogErrorf("Error during metrics cleanup: %v", err)
			}
			utils.LogInfo("End metrics cleanup")
		}
	}()
}
