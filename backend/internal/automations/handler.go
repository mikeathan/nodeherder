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
	StartAt string `json:"startAt"`
	Type    string `json:"type"`
}

func NewTimeSchedule() *TimeSchedule {
	return &TimeSchedule{
		StartAt: "",
		Type:    "",
	}
}

type ScheduleFunc interface {
	Name() string
	Do(automation Automation) func() error
}

type AutomationScheduleFunc struct {
	action func(automation Automation) error
	name   string
}

func NewAutomationScheduleFunc(name string, action func(automation Automation) error) ScheduleFunc {
	return &AutomationScheduleFunc{
		action: action,
		name:   name,
	}
}

func (s *AutomationScheduleFunc) Name() string {
	return s.name
}

func (s *AutomationScheduleFunc) Do(automation Automation) func() error {
	return func() error {
		return s.action(automation)
	}
}

type AutomationHandler interface {
	Process(automation Automation) error
	Type() string
	IsRunning(automation Automation) bool
}

type AutomationScheduler struct {
	repeatDuration time.Duration
	actionsMap     map[string]ScheduleFunc
	schedulers     map[string]*Scheduler
	ctx            context.Context
	clock          utils.Clock
}

func DefaultAutomationHandlers(ctx context.Context) []AutomationHandler {
	return []AutomationHandler{
		NewAutomationScheduler(
			WithContext(ctx),
			WithAutomationsFuncs(),
		),
	}
}

var scheduleFuncMap = map[string]ScheduleFunc{
	"enable": NewAutomationScheduleFunc(EnableScheduleType, func(automation Automation) error {
		automation.SetEnabled(true)
		utils.LogInfof("AutomationScheduler: enable automation %s", automation.GetId())
		return nil
	}),
	"disable": NewAutomationScheduleFunc(DisableScheduleType, func(automation Automation) error {
		automation.SetEnabled(false)
		utils.LogInfof("AutomationScheduler: disable automation %s", automation.GetId())
		return nil
	}),
}

func WithAutomationsFuncs() func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		for name, scheduleFunc := range scheduleFuncMap {
			as.actionsMap[name] = scheduleFunc
		}
	}
}

func WithCustomScheduleFuncs(funcMap map[string]func(automation Automation) error) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		for name, fn := range funcMap {
			as.actionsMap[name] = NewAutomationScheduleFunc(name, fn)
		}
	}
}

func WithScheduleFunc(name string, action func(automation Automation) error) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.actionsMap[name] = NewAutomationScheduleFunc(name, action)
	}
}

func WithContext(ctx context.Context) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.ctx = ctx
	}
}

func WithSchedulerClock(clock utils.Clock) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.clock = clock
	}
}
func WithRepeatDuration(duration time.Duration) func(*AutomationScheduler) {
	return func(as *AutomationScheduler) {
		as.repeatDuration = duration
	}
}

func NewAutomationScheduler(opts ...func(as *AutomationScheduler)) AutomationHandler {

	as := &AutomationScheduler{
		repeatDuration: 24 * time.Hour,
		actionsMap:     make(map[string]ScheduleFunc),
		schedulers:     map[string]*Scheduler{},
		ctx:            context.Background(),
		clock:          utils.NewRealClock(),
	}

	for _, opt := range opts {
		opt(as)
	}
	return as
}

func (a *AutomationScheduler) Type() string {
	return "scheduler"
}

func (a *AutomationScheduler) IsRunning(automation Automation) bool {
	if scheduler := a.schedulers[automation.GetId()]; scheduler != nil {
		return scheduler.IsRunning()
	}
	return false
}

func (a *AutomationScheduler) Process(automation Automation) error {

	scheduler := a.schedulers[automation.GetId()]

	if len(automation.GetSchedules()) == 0 {

		// remove scheduler if exists
		if scheduler != nil {

			scheduler.Stop()
			delete(a.schedulers, automation.GetId())
		}

		return nil
	}

	if scheduler != nil {

		scheduleUpdated := false

		for _, schedule := range automation.GetSchedules() {
			job := scheduler.FindJobByStartTime(schedule.StartAt)
			if job == nil {
				scheduleUpdated = true
				break
			}
		}

		if !scheduleUpdated {
			// nothing to do, schedules are the same
			return nil
		}

		if scheduleUpdated {
			err := scheduler.Stop()
			if err != nil {
				utils.LogError("Error stopping scheduler: ", err)
			}

			// delete and configure new scheduler
			scheduler = nil
			delete(a.schedulers, automation.GetId())
		}
	}

	// new scheduler
	scheduler = NewScheduler(a.clock, a.ctx)

	// disable automation and configure scheduler
	automation.SetEnabled(false)

	schedules := automation.GetSchedules()

	fmt.Printf("[DEBUG-SCHEDULER] Num of schedules %v\n", len(schedules))
	for _, schedule := range schedules {

		name := fmt.Sprintf("%s-%s", automation.GetId(), schedule.Type)

		actionFunc := a.actionsMap[schedule.Type]
		if actionFunc == nil {
			utils.LogErrorf("No action func for type %s in schedule %s", schedule.Type, name)
			continue
		}

		err := scheduler.
			Name(name).
			At(schedule.StartAt).
			Every(a.repeatDuration).
			Do(actionFunc.Do(automation))

		if err != nil {
			utils.LogErrorf("Error adding %s job: %v", name, err)
			return err
		}
	}

	scheduler.Start()
	a.schedulers[automation.GetId()] = scheduler

	return nil
}
