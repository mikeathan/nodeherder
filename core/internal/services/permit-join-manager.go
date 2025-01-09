package services

import (
	"errors"
	"node-herder/utils"
	"sync"
	"time"
)

type PermitJoinManager struct {
	mutex    sync.Mutex
	active   bool
	endTime  time.Time
	timer    *time.Timer
	callback func(bool)
}

func NewPermitJoinManager(timeoutHandler func(bool)) *PermitJoinManager {
	return &PermitJoinManager{
		callback: timeoutHandler,
	}
}

func (p *PermitJoinManager) Start(duration time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.active {
		return errors.New("permit join already active")
	}

	p.active = true
	p.endTime = time.Now().Add(duration)

	if p.timer != nil {
		p.timer.Stop()
	}

	// set inital state to true
	p.callback(p.active)

	p.timer = time.AfterFunc(duration, func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()

		p.active = false
		utils.LogInfo("Permit join timeout")

		// on timeout set to false
		p.callback(p.active)
	})

	utils.LogInfof("Permit join started for %s, ends at %s\n", duration, p.endTime)

	return nil
}

func (p *PermitJoinManager) Stop(invokeCallback bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	p.active = false

	if invokeCallback {
		p.callback(p.active)
	}

	utils.LogInfof("Permit join stopped manually")
}
