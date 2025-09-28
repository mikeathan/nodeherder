package automations

import (
	"node-herder/models/devices"
	"sync"
)

type AutomationContext interface {
	SetPayload(payload map[string]*devices.Entity)
	GetPayload(name string) (*devices.Entity, bool)
	GetCurrent(name string) any

	GetPending(name string) any
	SetPending(name string, value any)
	SetCurrent(name string, value any)
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

func (d *DeviceContext) SetPayload(payload map[string]*devices.Entity) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.payload = payload
}

// wIP
func (d *DeviceContext) SetPending(name string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pendingData[name] = value
}

func (d *DeviceContext) GetPending(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pendingData[name]
}

// wIP

func (d *DeviceContext) GetPayload(name string) (*devices.Entity, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, exists := d.payload[name]
	return value, exists
}

// we need current as it hold current state across automation devices not the current device
// payload has the current incoming device state but gets overriden so we can use it to compare with current state
// unless we store the whole thing instad overriding it ?

// pending will be action pending

func (d *DeviceContext) GetCurrent(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentData[name]
}

func (d *DeviceContext) SetCurrent(name string, value any) {
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
