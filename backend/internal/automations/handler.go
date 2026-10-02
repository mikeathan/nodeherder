package automations

import (
	"context"
	"fmt"
	"node-herder/utils"
	"slices"
	"strings"
	"sync"
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
	Prepare(automation Automation) (AutomationHandlerChange, error)
	Type() string
	IsRunning(automation Automation) bool
}

// Prepare must not disturb the active generation. Activate may fail; Rollback must
// undo provisional resources. Complete retires old resources after publication.
type AutomationHandlerChange interface {
	Activate() error
	Rollback()
	Complete()
}

type AutomationScheduler struct {
	repeatDuration time.Duration
	actionsMap     map[string]ScheduleFunc
	schedulers     map[string]*scheduleRegistration
	mutex          sync.RWMutex
	updateMu       sync.Mutex
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
		schedulers:     map[string]*scheduleRegistration{},
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
	id := automation.GetId()
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if registration := a.schedulers[id]; registration != nil {
		return registration.scheduler.IsRunning()
	}
	return false
}

func (a *AutomationScheduler) Process(automation Automation) error {
	change, err := a.Prepare(automation)
	if err != nil {
		return err
	}
	if err := change.Activate(); err != nil {
		change.Rollback()
		return err
	}
	change.Complete()
	return nil
}

type scheduleRegistration struct {
	scheduler  *Scheduler
	cancel     context.CancelFunc
	automation Automation
	schedules  []TimeSchedule
}

func (r *scheduleRegistration) stop() {
	if r == nil {
		return
	}
	// Stop also cancels partially started jobs when activation failed.
	r.scheduler.Stop()
	r.cancel()
}

type scheduleChange struct {
	owner     *AutomationScheduler
	id        string
	old, next *scheduleRegistration
	activated bool
	finished  sync.Once
}

func (a *AutomationScheduler) Prepare(automation Automation) (AutomationHandlerChange, error) {
	if !a.updateMu.TryLock() {
		return nil, ErrAutomationUpdateInProgress
	}
	change, err := a.prepareSchedule(automation)
	if err != nil {
		a.updateMu.Unlock()
	}
	return change, err
}

func (a *AutomationScheduler) prepareSchedule(automation Automation) (*scheduleChange, error) {
	if err := a.ctx.Err(); err != nil {
		return nil, err
	}
	id := automation.GetId()
	a.mutex.RLock()
	old := a.schedulers[id]
	a.mutex.RUnlock()
	change := &scheduleChange{owner: a, id: id, old: old}
	schedules := make([]TimeSchedule, 0, len(automation.GetSchedules()))
	for _, schedule := range automation.GetSchedules() {
		if schedule == nil {
			return nil, fmt.Errorf("nil schedule for %s", id)
		}
		if a.actionsMap[schedule.Type] == nil {
			return nil, fmt.Errorf("unknown schedule type %s", schedule.Type)
		}
		if _, err := ConvertStringToTimeUTC(a.clock, schedule.StartAt); err != nil {
			return nil, fmt.Errorf("schedule %s: %w", schedule.StartAt, err)
		}
		schedules = append(schedules, *schedule)
	}
	if len(schedules) == 0 {
		return change, nil
	}
	// Job order is not meaningful (Scheduler stores jobs in a map). Compare the
	// full multiset, including duplicates/types, without treating reordering as edits.
	slices.SortFunc(schedules, func(a, b TimeSchedule) int {
		if a.StartAt != b.StartAt {
			return strings.Compare(a.StartAt, b.StartAt)
		}
		return strings.Compare(a.Type, b.Type)
	})
	ctx, cancel := context.WithCancel(a.ctx)
	next := &scheduleRegistration{scheduler: NewScheduler(a.clock, ctx), cancel: cancel, automation: automation, schedules: schedules}
	for _, schedule := range schedules {
		err := next.scheduler.Name(fmt.Sprintf("%s-%s", id, schedule.Type)).At(schedule.StartAt).Every(a.repeatDuration).Do(a.actionsMap[schedule.Type].Do(automation))
		if err != nil {
			next.stop()
			return nil, err
		}
	}
	change.next = next
	return change, nil
}

func (c *scheduleChange) Activate() error {
	if c.next != nil {
		// Identical schedules retain the current window's enabled state, but jobs
		// always bind the replacement object rather than retaining the old one.
		if c.old != nil && slices.Equal(c.old.schedules, c.next.schedules) {
			c.next.automation.SetEnabled(c.old.automation.IsEnabled())
		} else {
			c.next.automation.SetEnabled(false)
		}
		if err := c.next.scheduler.Start(); err != nil {
			return err
		}
	}
	c.owner.mutex.Lock()
	if c.next == nil {
		delete(c.owner.schedulers, c.id)
	} else {
		c.owner.schedulers[c.id] = c.next
	}
	c.owner.mutex.Unlock()
	c.activated = true
	return nil
}

func (c *scheduleChange) Rollback() {
	c.finished.Do(func() {
		if c.activated {
			c.owner.mutex.Lock()
			if c.old == nil {
				delete(c.owner.schedulers, c.id)
			} else {
				c.owner.schedulers[c.id] = c.old
			}
			c.owner.mutex.Unlock()
		}
		c.next.stop()
		c.owner.updateMu.Unlock()
	})
}

func (c *scheduleChange) Complete() {
	c.finished.Do(func() { c.old.stop(); c.owner.updateMu.Unlock() })
}
