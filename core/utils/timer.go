package utils

import (
	"errors"
	"sync"
	"time"
)

type ActiveStateTimer struct {
	mutex    sync.Mutex
	active   bool
	endTime  time.Time
	timer    *time.Timer
	callback func(bool) error
}

func NewActiveStateTimer() *ActiveStateTimer {
	return &ActiveStateTimer{}
}

// Timer for active state
// - if inactive, set state to active and start timer. state is set back to inactive after duration
// - if active, exit early
func (p *ActiveStateTimer) Start(callback func(bool) error, duration time.Duration) error {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.active {
		return errors.New("state already active")
	}

	p.callback = callback
	p.active = true
	LogInfo("state set to active")

	err := p.callback(p.active)
	if err != nil {
		return err
	}

	p.endTime = time.Now().Add(duration)
	if p.timer != nil {
		p.timer.Stop()
	}

	p.timer = time.AfterFunc(duration, func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()

		p.active = false
		LogInfo("state set to inactive")

		// on timeout set to false
		p.callback(p.active)
	})

	LogInfof("active state timer started for %s, ends at %s\n", duration, p.endTime)

	return nil
}

func (p *ActiveStateTimer) Stop(invokeCallback bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	p.active = false

	if invokeCallback {
		err := p.callback(p.active)
		if err != nil {
			return err
		}
	}

	LogInfof("timer stopped manually")

	return nil
}
