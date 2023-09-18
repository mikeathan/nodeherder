package automations

import (
	"node-herder/utils"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type Constraint interface {
	Evaluate(parent *DeviceCondition)
}
type TimerConstraint struct {
	Type     string        `json:"type"`
	Duration time.Duration `json:"duration"`
	mut      sync.RWMutex
	sem      *semaphore.Weighted
	exit     chan bool
	running  bool
}

func NewTimerConstraint() *TimerConstraint {
	return &TimerConstraint{sem: semaphore.NewWeighted(1), Type: "timer"}
}

type DeviceConstraint struct {
	Type             string `json:"type"`
	Sensor           string `json:"sensor"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equalityoperator"`
}

func NewDeviceConstraint() *DeviceConstraint {
	return &DeviceConstraint{Type: "device", EqualityOperator: "="}
}

func (c *DeviceConstraint) Evaluate(parent *DeviceCondition) {

	if parent.conditionWithConstraintIsMatched(c) {
		err := parent.Action.Run()
		if err != nil {
			utils.LogErrorf("Action failed %s", err.Error())
		}
	}
}

func (t *TimerConstraint) isRunning() bool {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.running
}

func (t *TimerConstraint) stop() {
	t.mut.Lock()
	defer t.mut.Unlock()

	t.exit <- true
	//close(t.exit)
}

func (t *TimerConstraint) Evaluate(parent *DeviceCondition) {

	if !parent.conditionFromCacheIsMatched() {
		if t.isRunning() {
			t.stop()
		}
		return
	}

	utils.LogInfo("TimerConstraint evaluated")
	t.start(parent)
}

func (t *TimerConstraint) start(parent *DeviceCondition) {

	// check if its already running
	if ok := t.sem.TryAcquire(1); !ok {
		utils.LogDebug("time constraint is currently running.")
		return
	}

	t.mut.Lock()
	defer t.mut.Unlock()
	t.exit = make(chan bool, 1)
	t.running = true

	go func() {

		utils.LogInfo("time constraint started")

		ticker := *time.NewTicker(getDurationFromNow(t) * time.Millisecond)

		defer func() {
			t.sem.Release(1)
			ticker.Stop()
			t.running = false
		}()

		select {
		case <-ticker.C:

			parent.Action.Run()
			utils.LogInfo("timer constraint finished")
			return

		case <-t.exit:

			utils.LogInfo("timer constraint stopped")
			return
		}
	}()
}

func getDurationFromNow(t *TimerConstraint) time.Duration {

	timestamp := time.Now().Add(t.Duration)
	diff := time.Until(timestamp).Milliseconds()

	return time.Duration(diff)
}

func toFloat(value any) float32 {
	switch v := value.(type) {
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case float32:
		return float32(v)
	default:
		return float32(0)
	}
}
