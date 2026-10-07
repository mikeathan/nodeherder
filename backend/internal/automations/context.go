package automations

import (
	"node-herder/models/devices"
	"sync"
)

// AutomationContext is what conditions and actions see during one automation
// run: the triggering device's exposes and the run's origin, over state that
// persists across runs.
type AutomationContext interface {
	// device exposes as seen by this run
	GetDevicePayload(name string) (*devices.Entity, bool)

	// stores automation state across all devices, if included.
	SetCurrentState(name string, value any)
	GetCurrentState(name string) any

	// IsManualTrigger reports whether this run was started manually rather
	// than by a device message.
	IsManualTrigger() bool
}

var ContextIgnoreList = []string{"action"}

func WithContextIgnoreList(ignoreList []string) func(*DeviceContext) {
	return func(dc *DeviceContext) {
		ContextIgnoreList = ignoreList
	}
}

// DeviceContext is an automation's state that persists across runs.
type DeviceContext struct {
	currentData map[string]any
	mu          sync.RWMutex
}

func NewDeviceContext(opts ...func(*DeviceContext)) *DeviceContext {
	dc := &DeviceContext{
		currentData: map[string]any{},
	}
	for _, opt := range opts {
		opt(dc)
	}
	return dc
}

func (d *DeviceContext) GetCurrentState(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentData[name]
}

func (d *DeviceContext) SetCurrentState(name string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// if trigger is in ignore list, we  want to trigger it again
	for _, item := range ContextIgnoreList {
		if item == name {
			return
		}
	}
	d.currentData[name] = value
}

// runContext is one run's view: its own exposes and origin layered over the
// automation's persistent DeviceContext. It is built per run and never shared,
// so concurrent runs cannot overwrite each other's origin or exposes.
type runContext struct {
	*DeviceContext
	exposes map[string]*devices.Entity
	manual  bool
}

// NewRunContext returns the context for one run over persistent state.
func NewRunContext(state *DeviceContext, exposes map[string]*devices.Entity, manual bool) AutomationContext {
	return &runContext{DeviceContext: state, exposes: exposes, manual: manual}
}

func (r *runContext) GetDevicePayload(name string) (*devices.Entity, bool) {
	value, exists := r.exposes[name]
	return value, exists
}

func (r *runContext) IsManualTrigger() bool {
	return r.manual
}
