package services

import (
	"fmt"
	"maps"
	"node-herder/models/automations"
	"node-herder/models/devices"
	"node-herder/models/settings"

	"node-herder/store"
	"node-herder/utils"
	"sync"
)

type DeviceProcessor struct {
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
			dm.configureDeviceLifetime(id, cfg)
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

func (dm *DeviceProcessor) CreateOrUpdateDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, _ := dm.registrar.LookupByName(friendlyName)
	if device == nil {
		return dm.createNewDevice(friendlyName, connType, dataMap)
	}

	dm.processDeviceUpdate(device, dataMap)
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
		service:  NewDeviceLifetimeService(device, dm.events, dm.configCache, dm.automationQueries, utils.NewRealClock()),
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

func (b *DeviceProcessorBuilder) Build() *DeviceProcessor {
	return newDeviceProcessor(
		b.registrar,
		b.store,
		b.events,
		b.automationRetreiver,
	)
}
