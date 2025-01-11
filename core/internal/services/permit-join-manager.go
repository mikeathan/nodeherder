package services

import (
	"errors"
	"node-herder/utils"
	"sync"
	"time"
)

type ActiveStateTimer struct {
	mutex    sync.Mutex
	active   bool
	endTime  time.Time
	timer    *time.Timer
	callback func(bool)
}

func NewActiveStateTimer(callback func(bool)) *ActiveStateTimer {
	return &ActiveStateTimer{
		callback: callback,
	}
}

func (p *ActiveStateTimer) Start(duration time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.active {
		return errors.New("state already active")
	}

	err:= p.callback(p.active)

	check for error

	p.active = true
	utils.LogInfo("state set to active")

	p.endTime = time.Now().Add(duration)

	if p.timer != nil {
		p.timer.Stop()
	}

	p.timer = time.AfterFunc(duration, func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()

		p.active = false
		utils.LogInfo("state set to inactive")

		// on timeout set to false
		p.callback(p.active)
	})

	utils.LogInfof("active state timer started for %s, ends at %s\n", duration, p.endTime)

	return nil
}

func (p *ActiveStateTimer) Stop(invokeCallback bool) {
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

	utils.LogInfof("timer stopped manually")
}
