package automations

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type Constraint interface {
	Evaluate(parent *DeviceCondition)
	Reset()
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

func (c *DeviceConstraint) Evaluate(parent *DeviceCondition) {

	if inputVal, ok := parent.getValue(c.Type); ok {
		if Equalityoperators[c.EqualityOperator](inputVal, c.Value) {
			parent.Action.Run()
		}
	}
}

func (c *DeviceConstraint) Reset() {
	// nothing to reset
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

func (t *TimerConstraint) Evaluate(parent *DeviceCondition) {
	defer t.mut.Unlock()
	t.mut.Lock()

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

		case <-t.stop:

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

var Equalityoperators = map[string]func(any, any) bool{
	"=": func(v1 any, v2 any) bool {
		return v1 == v2
	},
	">=": func(v1 any, v2 any) bool {
		return ToFloat(v1) >= ToFloat(v2)
	},
	"<=": func(v1 any, v2 any) bool {
		return ToFloat(v1) <= ToFloat(v2)
	},
	">": func(v1 any, v2 any) bool {
		return ToFloat(v1) > ToFloat(v2)
	},
	"<": func(v1 any, v2 any) bool {
		return ToFloat(v1) < ToFloat(v2)
	},
}

func ToFloat(value any) float32 {
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

type DeviceCondition struct {
	Friendlyname     string      `json:"friendlyname"`
	Type             string      `json:"type"`
	Value            any         `json:"value"`
	Constraint       Constraint  `json:"constraint"` // TODO
	Action           *MqttAction `json:"action"`
	EqualityOperator string      `json:"equalityoperator"`
	cache            map[string]any
}

func NewDeviceCondition() *DeviceCondition {
	return &DeviceCondition{cache: make(map[string]any), EqualityOperator: "="}
}

func (m *DeviceCondition) getValue(sensor string) (any, bool) {
	if value, ok := m.cache[sensor]; ok {
		return value, true
	}

	return nil, false
}

func (m *DeviceCondition) Evaluate(data map[string]any) {

	value, ok := data[m.Type]
	if !ok {
		fmt.Printf("[DEBUG] sensor type %s not in input payload \n", m.Type)
		return
	}

	m.cache = data // cache any values, we need them for constraints
	if Equalityoperators[m.EqualityOperator](m.Value, value) {

		if m.Constraint != nil {
			m.Constraint.Evaluate(m)
			return
		}

		err := m.Action.Run()
		if err != nil {
			fmt.Printf("device sensor action failed %s", err.Error())
		}
	} else {

		// reset any existing constraint state
		if m.Constraint != nil {
			m.Constraint.Reset()
		}
	}
}

type TimerCondition interface {
	GetSchedule() time.Time
	IsRepeat() bool
}

type TimestampCondition struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Repeat    bool      `json:"repeat"`
}

type TimeDurationCondition struct {
	Type      string        `json:"type"`
	Duration  time.Duration `json:"duration"`
	Repeat    bool          `json:"repeat"`
	Timestamp time.Time
}

func (tc *TimeDurationCondition) GetSchedule() time.Time {
	tc.Timestamp = time.Now().Add(tc.Duration)

	fmt.Printf("Sceduled for %v \n", tc.Timestamp.Format(time.RFC3339))
	return tc.Timestamp
}

func (tc *TimestampCondition) GetSchedule() time.Time {

	if time.Since(tc.Timestamp) < 0 {
		tc.Timestamp = getTomorrow(tc.Timestamp)
	}

	fmt.Printf("Sceduled for %v \n", tc.Timestamp.Format(time.RFC3339))
	return tc.Timestamp
}

func (tc *TimeDurationCondition) IsRepeat() bool {
	return tc.Repeat
}

func (tc *TimestampCondition) IsRepeat() bool {
	return tc.Repeat
}

func getTomorrow(ts time.Time) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day()+1, ts.Hour(), ts.Minute(), 0, 0, ts.Location())
}
