package automations

import (
	"errors"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/models/devices"
	"node-herder/utils"
	"node-herder/utils/storage"
)

const (
	automationDir = "configs/automations"
)

type Engine interface { // TODO: might need to move it to Models????
	HandleDevice(device *devices.Device)
	HandleManual(automationID string, triggerName string)
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
}

func NewEngine(handlerFactory []AutomationHandler, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient) *AutomationEngine {

	return &AutomationEngine{
		mqttClient: mqtt,
		registrar:  registrar,
		storage: storage.NewJsonDiskStorage[Automation](automationDir, func() Automation {
			return &automationSerialiser{}
		}),
		handlers: handlerFactory,
	}
}

func (a *AutomationEngine) WithStorage(storage storage.Storage[Automation]) {
	a.storage = storage
}

func (a *AutomationEngine) IsAutomationEnabled(id string) bool {
	automation, err := a.storage.LoadFromCache(id)
	if err == nil {
		return automation.IsEnabled()
	}

	return false
}

func (a *AutomationEngine) HandleDevice(device *devices.Device) {
	automation, err := a.storage.LoadFromCache(device.Id)
	if err == nil && automation.IsEnabled() {
		automation.Evaluate(NewDeviceEvent(device))
	}
}

func (a *AutomationEngine) HandleManual(automationID string, triggerName string) {
	automation, err := a.storage.LoadFromCache(automationID)
	if err == nil && automation.IsEnabled() {
		//automation.Evaluate(NewManualEvent(triggerName))
	}
}

func (a *AutomationEngine) GetAllTriggers() []Automation {
	return a.storage.LoadAll()
}

func (a *AutomationEngine) Load(id string) (Automation, error) {
	return a.storage.Load(id)
}

func (a *AutomationEngine) Add(automation Automation) error {

	utils.LogInfof("adding automation id=%s, friendlyName=%s, enabled=%v", automation.GetId(), automation.GetFriendlyName(), automation.IsEnabled())
	err := a.configureAutomation(automation)
	if err != nil {
		utils.LogErrorf("configure automation id %s failed. Error=%s", automation.GetId(), err.Error())
		return err
	}

	a.storage.Store(automation.GetId(), automation)

	return nil
}

func (a *AutomationEngine) DeleteTrigger(id string, triggerId int) error {
	automation, err := a.storage.Load(id)
	if err != nil {
		return err
	}

	if triggerId >= len(automation.GetTriggers()) {
		return errors.New("trigger index out of bounds")
	}

	// remove trigger
	if err := automation.RemoveTrigger(triggerId); err != nil {
		return err
	}

	// store
	a.storage.Store(automation.GetId(), automation)
	return nil
}

func (a *AutomationEngine) Delete(id string) error {

	a.storage.Delete(id)
	return nil
}

func (a *AutomationEngine) Initialize() {

	utils.LogInfof("Initialize automations")
	automations, err := a.storage.Initialize() // <--------------- check if we clear any internal cache in automations after reloading
	if err != nil {
		utils.LogErrorf("Load automations failed. Error=%s", err.Error())
		return
	}

	for _, automation := range automations {

		utils.LogInfof("Loading automation id= %s, friendlyName=%s, Enabled=%t", automation.GetId(), automation.GetFriendlyName(), automation.IsEnabled())

		err := a.configureAutomation(automation)
		if err != nil {
			utils.LogErrorf("configure automation id %s failed. Error=%s", automation.GetId(), err.Error())
			continue
		}
	}
}

func (a *AutomationEngine) configureAutomation(automation Automation) error {

	err := automation.Configure(a.registrar, a.mqttClient)
	if err != nil {
		return err
	}

	for _, handler := range a.handlers {
		err = handler.Process(automation)
		if err != nil {
			return err
		}
	}

	return nil
}
