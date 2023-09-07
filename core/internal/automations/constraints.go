package automations

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type Constraint interface {
	Evaluate(parent *DeviceCondition, exit chan bool)
}

type TimerConstraint struct {
	Duration time.Duration `json:"duration"`
	stop     chan bool
	mut      sync.RWMutex
	sem      *semaphore.Weighted
}

func NewTimerConstraint() *TimerConstraint {
	return &TimerConstraint{stop: make(chan bool), sem: semaphore.NewWeighted(1)}
}

type DeviceConstraint struct {
	Type             string `json:"type"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equalityoperator"`
}

func (c *DeviceConstraint) Evaluate(parent *DeviceCondition, exit chan bool) {

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

	if ok := t.sem.TryAcquire(1); ok {
		fmt.Println("reset - Evaluate is not waiting")
		t.sem.Release(1)
		return
	}
	// problem here
	// deadlocks
	go func() {
		defer t.mut.Unlock()

		t.mut.Lock()
		t.stop <- true
	}()
}

func (t *TimerConstraint) Evaluate(parent *DeviceCondition, exit chan bool) {

	// check if its already running
	if ok := t.sem.TryAcquire(1); !ok {
		fmt.Println("time constraint is running")
		return
	}

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

		case <-exit:

			fmt.Println("timer constraint stopped")
			return
		}
	}()

	///

	// 	defer t.mut.Unlock()
	// 	t.mut.Lock()

	// 	// check if its already running
	// 	if ok := t.sem.TryAcquire(1); !ok {
	// 		fmt.Println("time constraint is running")
	// 		return
	// 	}

	// 	go func() {

	// 		fmt.Println("time constraint started")

	// 		ticker := *time.NewTicker(getDurationFromNow(t) * time.Millisecond)

	// 		defer func() {
	// 			t.sem.Release(1)
	// 			ticker.Stop()
	// 		}()

	// 		select {
	// 		case <-ticker.C:

	// 			parent.Action.Run()
	// 			fmt.Println("timer constraint finished")
	// 			return

	// 		case <-t.stop:

	//			fmt.Println("timer constraint stopped")
	//			return
	//		}
	//	}()
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
