package automations

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type Constraint interface {
	Evaluate(parent *DeviceCondition)
}

type TimerConstraint struct {
	Duration time.Duration `json:"duration"`
	mut      sync.RWMutex
	sem      *semaphore.Weighted
	exit     chan bool
	running  bool
}

func NewTimerConstraint() *TimerConstraint {
	return &TimerConstraint{sem: semaphore.NewWeighted(1)}
}

type DeviceConstraint struct {
	Type             string `json:"type"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equalityoperator"`
}

func (c *DeviceConstraint) Evaluate(parent *DeviceCondition) {

	if !parent.isConditionMatchedFromCache() {
		return
	}

	if inputVal, ok := parent.getValue(c.Type); ok {
		if Equalityoperators[c.EqualityOperator](inputVal, c.Value) {
			parent.Action.Run()
		}
	}
	return
}

func (t *TimerConstraint) Reset() {

	// if ok := t.sem.TryAcquire(1); ok {
	// 	fmt.Println("reset - Evaluate is not waiting")
	// 	t.sem.Release(1)
	// 	return
	// }
	// // problem here
	// // deadlocks
	// go func() {
	// 	defer t.mut.Unlock()

	// 	t.mut.Lock()
	// 	t.stop <- true
	// }()
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
	// close(t.exit) ???
}

func (t *TimerConstraint) Evaluate(parent *DeviceCondition) {

	if !parent.isConditionMatchedFromCache() {
		if t.isRunning() {
			t.stop()
		}
		return
	}

	t.start(parent)
}

func (t *TimerConstraint) start(parent *DeviceCondition) {

	// check if its already running
	if ok := t.sem.TryAcquire(1); !ok {
		fmt.Println("time constraint is running")
		return
	}

	t.mut.Lock()
	defer t.mut.Unlock()
	t.exit = make(chan bool, 1)
	t.running = true

	go func() {

		fmt.Println("time constraint started")

		ticker := *time.NewTicker(getDurationFromNow(t) * time.Millisecond)

		defer func() {
			t.sem.Release(1)
			ticker.Stop()

		}()

		select {
		case <-ticker.C:

			parent.Action.Run()
			fmt.Println("timer constraint finished")
			return

		case <-t.exit:
			t.running = false
			fmt.Println("timer constraint stopped")
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
