package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"node-herder/utils"
	"node-herder/utils/storage"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	automationDir = "configs/automations"
)

var ErrAutomationUpdateInProgress = errors.New("automation update already in progress")

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(device *devices.Device, payload map[string]interface{})
	HandleManual(automationID string, triggerName string) error
	Add(automation Automation) error
	Delete(id string) error
	DeleteTrigger(id string, triggerId int) error
	Initialize()
	Load(id string) (Automation, error)
	IsAutomationEnabled(id string) bool
	GetAllTriggers() []Automation
	WithStorage(storage storage.Storage[Automation])
}

type AutomationEngine struct {
	mqttClient mqtt.MqttClient
	registrar  services.DeviceRegistrar
	storage    storage.Storage[Automation]
	handlers   []AutomationHandler
	// Storage owns recipes; ready is the single source for published runtime reads.
	readyMu  sync.RWMutex
	ready    map[string]Automation
	updateMu sync.Mutex
	// A reload requested while updateMu is owned is deferred, never dropped.
	reloadPending atomic.Bool
}

func NewEngine(handlerFactory []AutomationHandler, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient) *AutomationEngine {

	serializer := NewAutomationSerialiser()
	storage := storage.NewJsonDiskStorage(automationDir, nil, serializer.Unmarshal)

	return &AutomationEngine{
		mqttClient: mqtt,
		registrar:  registrar,
		storage:    storage,
		handlers:   handlerFactory,
		ready:      make(map[string]Automation),
	}

}

func (a *AutomationEngine) WithStorage(storage storage.Storage[Automation]) {
	// Wiring only, before concurrent use.
	a.storage = storage
}

func (a *AutomationEngine) IsAutomationEnabled(id string) bool {
	automation, err := a.Load(id)
	if err == nil {
		return automation.IsEnabled()
	}

	return false
}

func (a *AutomationEngine) HandleDevice(device *devices.Device, payload map[string]interface{}) {
	automation, err := a.Load(device.Id)
	if err == nil && automation.IsEnabled() {
		automation.Evaluate(NewDeviceEvent(device, payload))
	}
}

func (a *AutomationEngine) HandleManual(automationID string, triggerName string) error {
	automation, err := a.Load(automationID)
	if err != nil {
		return err
	}

	if !automation.IsEnabled() {
		return errors.New("automation is disabled")
	}

	device, err := a.registrar.LookupById(automationID)
	if err != nil {
		utils.LogErrorf("Failed to lookup device with id %s. Error=%s", automationID, err.Error())
		return err
	}

	return automation.EvaluateTrigger(NewDeviceEvent(device, nil), triggerName)
}

func (a *AutomationEngine) GetAllTriggers() []Automation {
	a.readyMu.RLock()
	defer a.readyMu.RUnlock()
	ids := make([]string, 0, len(a.ready))
	for id := range a.ready {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	all := make([]Automation, 0, len(ids))
	for _, id := range ids {
		all = append(all, a.ready[id])
	}
	return all
}

func (a *AutomationEngine) Load(id string) (Automation, error) {
	a.readyMu.RLock()
	defer a.readyMu.RUnlock()
	if automation, ok := a.ready[id]; ok {
		return automation, nil
	}
	return nil, fmt.Errorf("automation %s is not ready", id)
}

func (a *AutomationEngine) withUpdate(update func() error) error {
	if !a.updateMu.TryLock() {
		return ErrAutomationUpdateInProgress
	}
	defer a.runPendingReload()
	defer a.updateMu.Unlock()
	return update()
}

// runPendingReload runs requested reloads unless another owner holds the gate;
// that owner observes the request after releasing it.
func (a *AutomationEngine) runPendingReload() {
	for a.reloadPending.Load() && a.updateMu.TryLock() {
		if a.reloadPending.Swap(false) {
			if err := a.reload(); err != nil {
				utils.LogErrorf("Load automations failed: %v", err)
			}
		}
		a.updateMu.Unlock()
	}
}

func (a *AutomationEngine) Add(automation Automation) error {
	return a.withUpdate(func() error { return a.publishAutomation(automation, true) })
}

func (a *AutomationEngine) DeleteTrigger(id string, triggerId int) error {
	return a.withUpdate(func() error {
		automation, err := a.Load(id)
		if err != nil {
			return err
		}

		if triggerId < 0 || triggerId >= len(automation.GetTriggers()) {
			return errors.New("trigger index out of bounds")
		}

		// Edit an unpublished recipe, never the currently executing generation.
		automation, err = cloneAutomation(automation)
		if err != nil {
			return err
		}
		if err := automation.RemoveTrigger(triggerId); err != nil {
			return err
		}

		return a.publishAutomation(automation, true)
	})
}

func (a *AutomationEngine) Delete(id string) error {
	return a.withUpdate(func() error { return a.removeAutomation(id, true) })
}

// Initialize never blocks: while another update owns the gate, the reload is
// deferred until that owner releases it.
func (a *AutomationEngine) Initialize() {
	a.reloadPending.Store(true)
	a.runPendingReload()
}

func (a *AutomationEngine) reload() error {
	utils.LogInfof("Initialize automations")
	automations, err := a.storage.Initialize()
	if err != nil {
		return err
	}

	seen := make(map[string]bool, len(automations))
	for _, automation := range automations {
		seen[automation.GetId()] = true
		err := a.publishAutomation(automation, false)
		if err != nil {
			utils.LogErrorf("configure automation id %s failed. Error=%s", automation.GetId(), err.Error())
			continue
		}
	}
	for _, old := range a.GetAllTriggers() {
		if !seen[old.GetId()] {
			if err := a.removeAutomation(old.GetId(), false); err != nil {
				utils.LogErrorf("remove automation %s failed: %v", old.GetId(), err)
			}
		}
	}
	return nil
}

func cloneAutomation(automation Automation) (Automation, error) {
	data, err := json.Marshal(automation)
	if err != nil {
		return nil, err
	}
	serializer := NewAutomationSerialiser()
	return serializer.Unmarshal(data)
}

func (a *AutomationEngine) publishAutomation(automation Automation, persist bool) error {
	if automation == nil {
		return errors.New("automation is required")
	}
	// Test/in-memory storage may reload the very same instance. Disk creates fresh
	// instances; only clone here when the caller supplied the published object.
	if old, err := a.Load(automation.GetId()); err == nil && old == automation {
		var err error
		automation, err = cloneAutomation(automation)
		if err != nil {
			return err
		}
	}
	if err := automation.Configure(a.registrar, a.mqttClient); err != nil {
		if errors.Is(err, ErrAutomationSourceDisabled) {
			id := automation.GetId()
			// Disable is authoritative: cleanup rollback must not restore execution.
			// Keep the recipe so a later successful reload can re-enable this ID.
			a.readyMu.Lock()
			delete(a.ready, id)
			a.readyMu.Unlock()
			if cleanupErr := a.removeAutomation(id, false); cleanupErr != nil {
				return errors.Join(err, fmt.Errorf("retire disabled automation: %w", cleanupErr))
			}
		}
		return err
	}
	id := automation.GetId()
	return a.changeHandlers(automation, func() error {
		if persist {
			if err := a.storage.Store(id, automation); err != nil {
				return fmt.Errorf("save automation: %w", err)
			}
		}
		a.readyMu.Lock()
		a.ready[id] = automation
		a.readyMu.Unlock()
		return nil
	})
}

func (a *AutomationEngine) removeAutomation(id string, persist bool) error {
	// An empty recipe prepares removal of this ID's handler resources.
	return a.changeHandlers(NewDevice(id), func() error {
		if persist {
			if err := a.storage.Delete(id); err != nil {
				return fmt.Errorf("delete automation: %w", err)
			}
		}
		a.readyMu.Lock()
		delete(a.ready, id)
		a.readyMu.Unlock()
		return nil
	})
}

func (a *AutomationEngine) changeHandlers(automation Automation, publish func() error) error {
	changes := make([]AutomationHandlerChange, 0, len(a.handlers))
	complete := false
	defer func() {
		if !complete {
			for i := len(changes) - 1; i >= 0; i-- {
				changes[i].Rollback()
			}
		}
	}()
	for _, handler := range a.handlers {
		change, err := handler.Prepare(automation)
		if err != nil {
			return fmt.Errorf("prepare %s: %w", handler.Type(), err)
		}
		changes = append(changes, change)
	}
	for _, change := range changes {
		if err := change.Activate(); err != nil {
			return err
		}
	}
	if err := publish(); err != nil {
		return err
	}
	complete = true
	for _, change := range changes {
		change.Complete()
	}
	return nil
}
