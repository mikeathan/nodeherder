package automations

import (
	"node-herder/models/devices"
	"sync"
)

type AutomationContext interface {
	// stores incoming updated device state
	SetDevicePayload(payload map[string]*devices.Entity)
	GetDevicePayload(name string) (*devices.Entity, bool)

	// stores automation state across all devices, if included.
	SetCurrentState(name string, value any)
	GetCurrentState(name string) any

	// source tracking
	SetManualTrigger(manual bool)
	IsManualTrigger() bool
}

var ContextIgnoreList = []string{"action"}

func WithContextIgnoreList(ignoreList []string) func(*DeviceContext) {
	return func(dc *DeviceContext) {
		ContextIgnoreList = ignoreList
	}
}

// Device Context
type DeviceContext struct {
	currentData map[string]any
	payload     map[string]*devices.Entity
	isManual    bool
	mu          sync.RWMutex
}

func NewDeviceContext(opts ...func(*DeviceContext)) *DeviceContext {
	dc := &DeviceContext{
		currentData: map[string]any{},
		payload:     make(map[string]*devices.Entity),
	}
	for _, opt := range opts {
		opt(dc)
	}
	return dc
}

func (d *DeviceContext) SetDevicePayload(payload map[string]*devices.Entity) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.payload = payload
}

func (d *DeviceContext) GetDevicePayload(name string) (*devices.Entity, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, exists := d.payload[name]
	return value, exists
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

func (d *DeviceContext) SetManualTrigger(manual bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.isManual = manual
}

func (d *DeviceContext) IsManualTrigger() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isManual
}
