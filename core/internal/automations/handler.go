package automations

import (
	"context"
	"fmt"
	"node-herder/utils"
	"time"
)

const (
	EnableScheduleType  = "enable"
	DisableScheduleType = "disable"
)

type TimeSchedule struct {
	Enabled bool   `json:"enabled"`
	StartAt string `json:"startAt"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

func NewTimeSchedule() *TimeSchedule {
	return &TimeSchedule{
		Enabled: false,
		StartAt: "",
		Name:    "",
		Type:    "",
	}
}

type ScheduleFunc interface {
	Name() string
	Do(a *Device) func() error
}

type AutomationScheduleFunc struct {
	action func(a *Device) error
	name   string
}

func NewAutomationScheduleFunc(name string, action func(a *Device) error) ScheduleFunc {
	return &AutomationScheduleFunc{
		action: action,
		name:   name,
	}
}

func (s *AutomationScheduleFunc) Name() string {
	return s.name
}

func (s *AutomationScheduleFunc) Do(a *Device) func() error {
	return func() error {
		return s.action(a)
	}
}

type AutomationHandler interface {
	Process(automation *Device) error
	Type() string
}

type AutomationScheduler struct {
	repeatDuration time.Duration
	scheduler      *Scheduler
	actionsMap     map[string]ScheduleFunc
	ctx            context.Context
}

func WithScheduleFunc(name string, action func(a *Device) error) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.actionsMap[name] = NewAutomationScheduleFunc(name, action)
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
		actionsMap:     make(map[string]ScheduleFunc),
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

func (a *AutomationScheduler) Process(automation *Device) error {

	if len(automation.Schedules) == 0 {
		return nil
	}
	// if !automation.Schedule.Enabled {
	// 	if a.scheduler != nil && a.scheduler.IsRunning() {
	// 		a.scheduler.Stop()
	// 	}

	// 	return nil
	// }

	if a.scheduler.IsRunning() {

		// check if schedule has changed and determine logic
		// TODO
		return nil
	}

	// if a.scheduler != nil {
	// 	a.scheduler.Stop()
	// }

	// if we have schedule, disable automation and configure scheduler
	automation.Enabled = false

	for _, schedule := range automation.Schedules {

		actionFunc := a.actionsMap[schedule.Type]
		if actionFunc == nil {
			utils.LogErrorf("No action func for type %s in schedule %s", schedule.Type, schedule.Name)
			continue
		}

		err := a.scheduler.
			Name(fmt.Sprintf("%s job for %v automation", schedule.Name, automation.FriendlyName)).
			At(schedule.StartAt).
			Every(a.repeatDuration).
			Do(actionFunc.Do(automation))

		if err != nil {
			utils.LogErrorf("Error adding %s job: %v", schedule.Name, err)
			return err
		}
	}

	a.scheduler.Start()

	return nil
}
