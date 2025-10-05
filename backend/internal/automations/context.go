package automations

import (
	"node-herder/models/devices"
	"reflect"
	"sync"
	"time"
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

func WithContextIgnoreList(ignoreList []string) func(*DeviceContext) {
	return func(dc *DeviceContext) {
		ContextIgnoreList = ignoreList
	}
}

func WithDeviceContextTtl(d time.Duration) func(*DeviceContext) {
	return func(dc *DeviceContext) {
		dc.ttl = d
	}
}

// Device Context
type DeviceContext struct {
	currentData map[string]any
	payload     map[string]*devices.Entity
	pendingData map[string]any
	mu          sync.RWMutex
	ttl         time.Duration
}

func NewDeviceContext(opts ...func(*DeviceContext)) *DeviceContext {
	dc := &DeviceContext{
		currentData: map[string]any{},
		payload:     make(map[string]*devices.Entity),
		pendingData: map[string]any{},
		ttl:         1 * time.Second, // default TTL for pending state
	}
	for _, opt := range opts {
		opt(dc)
	}
	return dc
}

func (d *DeviceContext) GetPendingState(name string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pendingData[name]
}

func (d *DeviceContext) SetPendingState(name string, value any) {
	if value == nil {
		d.clearPending(name)
		return
	}
	d.setPendingWithTTL(name, value, d.ttl)
}

func (d *DeviceContext) setPendingWithTTL(name string, value any, ttl time.Duration) {
	d.mu.Lock()
	d.pendingData[name] = value
	d.mu.Unlock()

	time.AfterFunc(ttl, func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		// Only clear if still the same value
		if current, exists := d.pendingData[name]; exists && reflect.DeepEqual(current, value) {
			delete(d.pendingData, name)
		}
	})
}

func (d *DeviceContext) clearPending(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.pendingData, name)
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
