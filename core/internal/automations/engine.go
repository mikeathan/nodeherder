package automations

import (
	"context"
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
	Add(automation *Device) error
	Delete(id string) error
	DeleteTrigger(id string, triggerId int) error
	Initialize()
	Load(id string) (*Device, error)
	GetAllTriggers() []*Device
	WithStorage(storage storage.Storage[Device])
}

type AutomationEngine struct {
	mqttClient      mqtt.MqttClient
	registrar       services.DeviceRegistrar
	storage         storage.Storage[Device]
	deviceScheduler map[string]*Scheduler
	ctx             context.Context
}

func NewEngine(registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, ctx context.Context) *AutomationEngine {

	return &AutomationEngine{
		mqttClient: mqtt,
		registrar:  registrar,
		storage: storage.NewJsonDiskStorage[Device](automationDir, func() *Device {
			return newDevice()
		}),
		deviceScheduler: map[string]*Scheduler{},
		ctx:             ctx,
	}
}

func (a *AutomationEngine) WithStorage(storage storage.Storage[Device]) {
	a.storage = storage
}

func (a *AutomationEngine) HandleDevice(device *devices.Device) {
	triggerDevice, err := a.storage.LoadFromCache(device.Id)
	if err == nil && triggerDevice.Enabled {
		triggerDevice.Evaluate(device)
	}
}

func (a *AutomationEngine) GetAllTriggers() []*Device {
	return a.storage.LoadAll()
}

func (a *AutomationEngine) Load(id string) (*Device, error) {
	return a.storage.Load(id)
}

func (a *AutomationEngine) Add(automation *Device) error {

	utils.LogInfof("adding automation id=%s, friendlyName=%s, enabled=%v", automation.Id, automation.FriendlyName, automation.Enabled)
	err := a.configureAutomation(automation)
	if err != nil {
		utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
		return err
	}

	a.storage.Store(automation.Id, automation)

	return nil
}

func (a *AutomationEngine) DeleteTrigger(id string, triggerId int) error {
	automation, err := a.storage.Load(id)
	if err != nil {
		return err
	}

	if triggerId >= len(automation.Triggers) {
		return errors.New("trigger index out of bounds")
	}

	// remove trigger
	automation.Triggers = append(automation.Triggers[:triggerId], automation.Triggers[triggerId+1:]...)

	// store
	a.storage.Store(id, automation)
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

		utils.LogInfof("Loading automation id= %s, friendlyName=%s, Enabled=%t", automation.Id, automation.FriendlyName, automation.Enabled)

		err := a.configureAutomation(automation)
		if err != nil {
			utils.LogErrorf("configure automation id %s failed. Error=%s", automation.Id, err.Error())
			continue
		}
	}
}

func (a *AutomationEngine) configureAutomation(automation *Device) error {

	err := automation.configure(a.registrar, a.mqttClient)
	if err != nil {
		return err
	}

	if automation.Schedule != nil && automation.Schedule.Enabled {

		// if we have schedule, disable automation and configure scheduler
		automation.Enabled = false

		utils.LogInfof("adding schedule for automation id=%s, friendlyName=%s, start=%s, end=%s", automation.Id, automation.FriendlyName, automation.Schedule.Start, automation.Schedule.End)

	NOTE:
		// WE NEED TO KEEP THE SCEDULER ALIVE - we ar ehere trigered from the worker pool and when it exits the scheduler will be destroyed

		scheduler := NewScheduler(a.ctx)
		err := scheduler.AddJob(automation.Schedule.Start, func() error {
			automation.Enabled = true
			return nil
		})
		if err != nil {
			return err
		}

		err = scheduler.AddJob(automation.Schedule.End, func() error {
			automation.Enabled = false
			return nil
		})
		if err != nil {
			return err
		}

		scheduler.Start()
		a.deviceScheduler[automation.Id] = scheduler
	}

	return nil
}
