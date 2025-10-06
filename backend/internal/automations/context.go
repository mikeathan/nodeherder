package automations

import (
	"node-herder/models/devices"
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

type pendingEntry struct {
	value   any
	expires time.Time
}

// Device Context
type DeviceContext struct {
	currentData map[string]any
	payload     map[string]*devices.Entity
	pendingData map[string]pendingEntry
	mu          sync.RWMutex
	ttl         time.Duration
	stopCleanup chan struct{}
}

func NewDeviceContext(opts ...func(*DeviceContext)) *DeviceContext {
	dc := &DeviceContext{
		currentData: map[string]any{},
		payload:     make(map[string]*devices.Entity),
		pendingData: map[string]pendingEntry{},
		ttl:         1 * time.Second, // default TTL for pending state
		stopCleanup: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(dc)
	}
	go dc.cleanupLoop()
	return dc
}

func (d *DeviceContext) GetPendingState(name string) any {
	d.mu.RLock()
	entry, exists := d.pendingData[name]
	isExpired := exists && time.Now().After(entry.expires)
	d.mu.RUnlock()

	if !exists {
		return nil
	}

	if isExpired {
		d.removeExpiredPending(name)
		return nil
	}

	return entry.value
}

func (d *DeviceContext) SetPendingState(name string, value any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if value == nil {
		delete(d.pendingData, name)
		return
	}

	d.pendingData[name] = pendingEntry{
		value:   value,
		expires: time.Now().Add(d.ttl),
	}
}

func (d *DeviceContext) removeExpiredPending(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Double-check under write lock
	if entry, ok := d.pendingData[name]; ok && time.Now().After(entry.expires) {
		delete(d.pendingData, name)
	}
}

func (d *DeviceContext) cleanupLoop() {
	// Reduced cleanup frequency - less CPU usage
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			now := time.Now()
			for name, entry := range d.pendingData {
				if now.After(entry.expires) {
					delete(d.pendingData, name)
				}
			}
			d.mu.Unlock()
		case <-d.stopCleanup:
			return
		}
	}
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
