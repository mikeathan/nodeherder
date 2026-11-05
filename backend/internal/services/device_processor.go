package services

import (
	"context"
	"fmt"
	"maps"
	"node-herder/internal/spatial"
	"node-herder/models/automations"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"

	"node-herder/store"
	"node-herder/utils"
	"sync"
)

type DeviceProcessor struct {
	ctx               context.Context
	registrar         *HubRegisterService
	deviceServices    map[string]*deviceServiceEntry
	configCache       *settings.DeviceConfigCache
	events            *devices.DeviceRequestEvents
	automationQueries automations.AutomationQuerier
	mutex             *sync.RWMutex
}

// Only the creator drains this queue. Registry locking protects enqueue/publication;
// Seed, Update and configuration callbacks always run outside that lock.
type deviceServiceEntry struct {
	service  *DeviceLifetimeService
	creating bool
	pending  []deviceCreationEvent
}

type deviceCreationEvent struct {
	payload map[string]interface{}
	config  *settings.DeviceConfig
}

func (event deviceCreationEvent) apply(service *DeviceLifetimeService) {
	if event.config != nil {
		service.OnConfigUpdated(event.config)
	} else {
		service.Update(event.payload)
	}
}

func (event deviceCreationEvent) snapshot() deviceCreationEvent {
	if event.config != nil {
		// Lifetime configuration consumes only Disabled.
		event.config = &settings.DeviceConfig{Disabled: event.config.Disabled}
	} else {
		event.payload = maps.Clone(event.payload)
	}
	return event
}

func newDeviceProcessor(registrar *HubRegisterService, store store.AppStore, events *devices.DeviceRequestEvents, automationRetreiver automations.AutomationQuerier) *DeviceProcessor {
	return &DeviceProcessor{
		registrar:         registrar,
		deviceServices:    make(map[string]*deviceServiceEntry),
		configCache:       store.AppConfig().GetDeviceConfigCache(),
		events:            events,
		automationQueries: automationRetreiver,
		mutex:             &sync.RWMutex{},
	}
}
func (dm *DeviceProcessor) OnDeviceConfigUpdated(cfg *settings.DeviceConfig) {

	if cfg.Id == "" {
		// device config defaults
		for _, id := range dm.getDeviceIDs() {
			effective, err := dm.configCache.Get(id)
			if err == nil {
				dm.configureDeviceLifetime(id, effective)
			}
		}
		return
	}

	// device config overrides
	dm.configureDeviceLifetime(cfg.Id, cfg)
}

func (dm *DeviceProcessor) configureDeviceLifetime(id string, cfg *settings.DeviceConfig) {
	event := deviceCreationEvent{config: cfg}
	if entry, _ := dm.routeDeviceEvent(id, nil, event); entry != nil {
		event.apply(entry.service)
	}
}

// DeviceName resolves an active generation even while its first snapshot is still
// being published and has not reached storage. Removal must join that same lane.
func (dm *DeviceProcessor) DeviceName(id string) (string, bool) {
	dm.mutex.RLock()
	entry := dm.deviceServices[id]
	dm.mutex.RUnlock()
	if entry == nil {
		return "", false
	}
	return entry.service.device.GetFriendlyName(), true
}

// RemoveDevice retires the exact active generation before evicting its registry
// entry. The caller serializes this operation with reports on the device lane.
// No registry lock surrounds persistence or callbacks. A later accepted discovery
// creates a fresh generation rather than reusing a closed lifetime.
// HasCompleteSpatialMapping reports whether at least one saved row maps X, Y and Active to
// three distinct entities, the minimum a usable derived position needs.
func HasCompleteSpatialMapping(mappings []settings.SpatialMapping) bool {
	for _, mapping := range mappings {
		if mapping.X == "" || mapping.Y == "" || mapping.Active == "" {
			continue
		}
		if mapping.X == mapping.Y || mapping.Active == mapping.X || mapping.Active == mapping.Y {
			continue
		}
		return true
	}
	return false
}

func (dm *DeviceProcessor) RemoveDevice(id string) error {
	dm.mutex.RLock()
	entry := dm.deviceServices[id]
	dm.mutex.RUnlock()
	remove := func() error { return dm.registrar.RemoveDevice(id) }
	if entry == nil {
		return remove()
	}
	if err := entry.service.Retire(remove); err != nil {
		return err
	}
	dm.mutex.Lock()
	if dm.deviceServices[id] == entry {
		clear(entry.pending)
		entry.pending = nil
		delete(dm.deviceServices, id)
	}
	dm.mutex.Unlock()
	return nil
}

func (dm *DeviceProcessor) CreateOrUpdateDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, _ := dm.registrar.LookupByName(friendlyName)
	if device == nil {
		return dm.createNewDevice(friendlyName, connType, dataMap)
	}

	dm.processDeviceUpdate(device, dataMap)
	return nil
}

// RegisterProviderDevice installs a fresh provider inventory generation under its
// stable ID. The caller must serialize it with reports and removal on the device
// lane. Storage failure leaves the previous lifetime untouched.
func (dm *DeviceProcessor) RegisterProviderDevice(device *devices.Device) error {
	if device == nil || device.Id == "" || device.Provider == "" || device.FriendlyName == "" {
		return fmt.Errorf("provider device identity is required")
	}
	dm.mutex.RLock()
	old := dm.deviceServices[device.Id]
	dm.mutex.RUnlock()
	storeDevice := func() error { return dm.registrar.store.StoreDevice(device.FriendlyName, device) }
	var err error
	if old != nil {
		// Change-only binary sensors may not republish their current state after a
		// reconnect. Retain the last confirmed binary value for matching entities
		// before fencing/replacing the old generation. Derived devices are projections:
		// replaying their old occupancy would present stale presence as current.
		if device.Provider != spatial.ProviderName {
			preserveProviderBinaryValues(device, old.service.device)
		}
		// Fence delayed publication before replacing the persisted generation.
		err = old.service.Retire(storeDevice)
	} else {
		err = storeDevice()
	}
	if err != nil {
		return fmt.Errorf("store provider device: %w", err)
	}
	dm.mutex.Lock()
	if dm.deviceServices[device.Id] == old {
		delete(dm.deviceServices, device.Id)
	}
	dm.mutex.Unlock()
	dm.processDeviceUpdate(device, map[string]interface{}{})
	return nil
}

// preserveProviderBinaryValues carries matching binary states across a provider
// inventory replacement. Numeric readings are re-published on subscription and are
// deliberately not carried, so stale measurements are never presented as live.
func preserveProviderBinaryValues(next, previous *devices.Device) {
	if next == nil || previous == nil {
		return
	}
	for name, entity := range next.Exposes {
		if entity == nil || entity.Type != bridge.BinaryDataType {
			continue
		}
		oldEntity := previous.Exposes[name]
		if oldEntity == nil || oldEntity.Type != entity.Type {
			continue
		}
		if value := oldEntity.Data.Value(); value != nil {
			entity.Data.SetValue(value)
		}
	}
}

// UpdateProviderDevice applies a report to an already registered provider
// generation. The caller serializes updates with inventory and availability.
func (dm *DeviceProcessor) UpdateProviderDevice(id string, payload map[string]interface{}) error {
	dm.mutex.RLock()
	entry := dm.deviceServices[id]
	dm.mutex.RUnlock()
	if entry == nil || entry.service.device.Provider == "" {
		return fmt.Errorf("provider device not registered")
	}
	dm.processDeviceUpdate(entry.service.device, payload)
	return nil
}

// ProviderOffline marks an existing provider generation unavailable and clears
// queued presentation updates. The caller serializes it on the device lane.
func (dm *DeviceProcessor) ProviderOffline(id string) error {
	dm.mutex.RLock()
	entry := dm.deviceServices[id]
	dm.mutex.RUnlock()
	if entry == nil || entry.service.device.Provider == "" {
		return fmt.Errorf("provider device not registered")
	}
	entry.service.ProviderOffline()
	if err := dm.registrar.store.StoreDevice(entry.service.device.FriendlyName, entry.service.device); err != nil {
		return fmt.Errorf("store provider availability: %w", err)
	}
	return nil
}

func (dm *DeviceProcessor) createNewDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, err := dm.registrar.CreateNewDevice(friendlyName, connType, dataMap)
	if err != nil {
		return fmt.Errorf("failed to create new device: %w", err)
	}

	dm.processDeviceUpdate(device, dataMap)
	return nil
}

// A nil device means configuration-only: do not create an unknown service.
// A nil result means absent or queued; true grants exclusive Seed/drain ownership.
func (dm *DeviceProcessor) routeDeviceEvent(id string, device *devices.Device, event deviceCreationEvent) (*deviceServiceEntry, bool) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	if entry, ok := dm.deviceServices[id]; ok {
		if entry.creating {
			entry.pending = append(entry.pending, event.snapshot())
			return nil, false
		}
		return entry, false
	}
	if device == nil {
		return nil, false
	}
	entry := &deviceServiceEntry{
		service:  NewDeviceLifetimeService(device, dm.events, dm.configCache, dm.automationQueries, utils.NewRealClock(), dm.ctx),
		creating: true,
	}
	dm.deviceServices[id] = entry
	return entry, true
}

func (dm *DeviceProcessor) processDeviceUpdate(device *devices.Device, dataMap map[string]interface{}) {
	event := deviceCreationEvent{payload: dataMap}
	entry, creator := dm.routeDeviceEvent(device.Id, device, event)
	if entry == nil {
		return
	}
	if !creator {
		event.apply(entry.service)
		return
	}
	entry.service.Seed(dataMap)
	dm.drainDeviceEvents(entry)
}

func (dm *DeviceProcessor) drainDeviceEvents(entry *deviceServiceEntry) {
	for {
		pending := dm.takePendingDeviceEvents(entry)
		if len(pending) == 0 {
			return
		}
		for _, event := range pending {
			event.apply(entry.service)
		}
	}
}

func (dm *DeviceProcessor) takePendingDeviceEvents(entry *deviceServiceEntry) []deviceCreationEvent {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	pending := entry.pending
	entry.pending = nil
	if len(pending) == 0 {
		entry.creating = false
	}
	return pending
}

func (dm *DeviceProcessor) getDeviceIDs() []string {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	ids := make([]string, 0, len(dm.deviceServices))
	for id := range dm.deviceServices {
		ids = append(ids, id)
	}
	return ids
}

// DeviceProcessor builder
type DeviceProcessorBuilder struct {
	ctx                 context.Context
	registrar           *HubRegisterService
	store               store.AppStore
	events              *devices.DeviceRequestEvents
	automationRetreiver automations.AutomationQuerier
}

func NewDeviceProcessorBuilder() *DeviceProcessorBuilder {
	return &DeviceProcessorBuilder{}
}

func (b *DeviceProcessorBuilder) WithRegistrar(r *HubRegisterService) *DeviceProcessorBuilder {
	b.registrar = r
	return b
}

func (b *DeviceProcessorBuilder) WithStore(s store.AppStore) *DeviceProcessorBuilder {
	b.store = s
	return b
}

func (b *DeviceProcessorBuilder) WithEvents(e *devices.DeviceRequestEvents) *DeviceProcessorBuilder {
	b.events = e
	return b
}

func (b *DeviceProcessorBuilder) WithAutomationQuerier(a automations.AutomationQuerier) *DeviceProcessorBuilder {
	b.automationRetreiver = a
	return b
}

func (b *DeviceProcessorBuilder) WithContext(ctx context.Context) *DeviceProcessorBuilder {
	b.ctx = ctx
	return b
}

func (b *DeviceProcessorBuilder) Build() *DeviceProcessor {
	processor := newDeviceProcessor(
		b.registrar,
		b.store,
		b.events,
		b.automationRetreiver,
	)
	processor.ctx = b.ctx
	if processor.ctx == nil {
		processor.ctx = context.Background()
	}
	return processor
}
