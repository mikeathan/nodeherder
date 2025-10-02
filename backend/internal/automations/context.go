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

	// stores pending state
	SetPendingState(name string, value any)
	GetPendingState(name string) any
}

var ContextIgnoreList = []string{"action"}

// Device Context
type DeviceContext struct {
	currentData map[string]any
	payload     map[string]*devices.Entity
	pendingData map[string]any
	mu          sync.RWMutex
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{
		currentData: map[string]any{},
		payload:     make(map[string]*devices.Entity),
		pendingData: map[string]any{},
	}
}

func (d *DeviceContext) SetDevicePayload(payload map[string]*devices.Entity) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.payload = payload
}

// wIP
func (d *DeviceContext) SetPendingState(name string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pendingData[name] = value
}

func (d *DeviceContext) GetPendingState(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pendingData[name]
}

// wIP

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
