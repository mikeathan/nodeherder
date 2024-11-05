package automations

import (
	"context"
	"fmt"
	"node-herder/utils"
	"time"
)

type AutomationHandler interface {
	Process(automation *Device) error
	Type() string
}

type AutomationScheduler struct {
	repeatDuration time.Duration
	scheduler      *Scheduler
	startFunc      func() error
	stopFunc       func() error
	ctx            context.Context
}

func WithStartFunc(start func() error) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.startFunc = start
	}
}
func WithStopFunc(stop func() error) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.stopFunc = stop
	}
}

func WithContext(ctx context.Context) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.ctx = ctx
	}
}

func WithRepeatDuration(duration time.Duration) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.repeatDuration = duration
	}
}

func NewAutomationScheduler(opts ...func(as *AutomationScheduler)) AutomationHandler {

	as := &AutomationScheduler{
		repeatDuration: 24 * time.Hour, // default to 1 day
		ctx:            context.Background(),
		startFunc:      func() error { return nil },
		stopFunc:       func() error { return nil },
	}
	for _, opt := range opts {
		opt(as)
	}
	as.scheduler = NewScheduler(as.ctx)
	return as
}

func (a *AutomationScheduler) Type() string {
	return "scheduler"
}

need to use the startFunc and stopFunc so i can be configured
func (a *AutomationScheduler) Process(automation *Device) error {

	if automation.Schedule == nil {
		return nil
	}
	if !automation.Schedule.Enabled {
		if a.scheduler != nil && a.scheduler.IsRunning() {
			a.scheduler.Stop()
		}

		return nil
	}

	if a.scheduler != nil && a.scheduler.IsRunning() {

		// check if schedule has changed and determine logic
		// TODO
		return nil
	}

	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	// if we have schedule, disable automation and configure scheduler
	automation.Enabled = false

	err := a.scheduler.
		Name(fmt.Sprintf("Start job for %v automation", automation.FriendlyName)).
		At(automation.Schedule.Start).
		Every(a.repeatDuration).
		Do(func() error {

			utils.LogInfof("Schedule enable %v automation", automation.FriendlyName)
			automation.Enabled = true
			return nil
		})

	if err != nil {
		utils.LogError("Error adding start job:", err)
		return err
	}

	err = a.scheduler.
		Name(fmt.Sprintf("End job for %v automation", automation.FriendlyName)).
		At(automation.Schedule.End).
		Every(a.repeatDuration).
		Do(func() error {

			utils.LogInfof("Schedule disable %v automation", automation.FriendlyName)
			automation.Enabled = false
			return nil
		})

	if err != nil {
		utils.LogError("Error adding end job:", err)
		return err
	}

	a.scheduler.Start()

	return nil
}
