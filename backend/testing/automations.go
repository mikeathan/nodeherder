package utils_test

import (
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/utils"
)

func CreateDoorContactWithAlarmTriggerAutomation(doorSensorId string, alarmId string, mqtt mqtt.MqttClient) *automations.Device {
	// setup automations
	alarmAction := automations.NewTriggerAction()
	alarmAction.Id = alarmId

	alarmAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "alarm",
			Data: true,
		},
	}
	alarmAction.Type = automations.TriggerAction
	alarmAction.Client = mqtt

	doorSensorTrigger := automations.NewDeviceTrigger("contact")
	doorSensorTrigger.Actions = []automations.MqttAction{alarmAction}

	deviceAutomation := automations.NewDevice("door sensor")
	deviceAutomation.Id = doorSensorId
	deviceAutomation.FriendlyName = "front door sensor"
	deviceAutomation.Enabled = true
	deviceAutomation.Triggers = []automations.Trigger{doorSensorTrigger}

	return deviceAutomation
}

func CreateDoorContactDurationWithAlarmTriggerAutomation(doorSensorId string, alarmId string, mqtt mqtt.MqttClient) *automations.Device {
	// setup automations

	// door open triggers alarm
	openDoorTrigger := automations.NewDeviceTrigger("contact")
	openDoorTrigger.Conditions = []automations.Condition{
		NewExposeCondition("contact", true, "="),
	}

	alarmOnAction := automations.NewTriggerAction()
	alarmOnAction.Id = alarmId
	alarmOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "alarm",
			Data: true,
		},
		{
			Name: "duration",
			Data: 2,
		},
	}
	alarmOnAction.Type = automations.TriggerAction
	alarmOnAction.Client = mqtt
	openDoorTrigger.Actions = []automations.MqttAction{alarmOnAction}

	// door close turns off alarm
	closeDoorTrigger := automations.NewDeviceTrigger("contact")
	closeDoorTrigger.Conditions = []automations.Condition{
		NewExposeCondition("contact", false, "="),
	}
	alarmOffAction := automations.NewTriggerAction()
	alarmOffAction.Id = alarmId
	alarmOffAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "alarm",
			Data: false,
		},
	}
	alarmOffAction.Type = automations.TriggerAction
	alarmOffAction.Client = mqtt
	closeDoorTrigger.Actions = []automations.MqttAction{alarmOffAction}

	// create device automation
	deviceAutomation := automations.NewDevice("door sensor")
	deviceAutomation.Id = doorSensorId
	deviceAutomation.FriendlyName = "front door sensor"
	deviceAutomation.Enabled = true
	deviceAutomation.Triggers = []automations.Trigger{openDoorTrigger, closeDoorTrigger}

	return deviceAutomation
}

func CreateDialTriggerActionsBrightness(actionId string, dialActionName string, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	condition := NewExposeCondition("action", dialActionName, "=")
	step := &automations.Step{}
	step.Id = actionId
	step.Operator = "+"
	step.Property = "brightness"

	action := &automations.MqttStepAction{}
	action.Id = actionId
	action.Property = "brightness"
	action.Type = "step"
	action.Data = 0.5
	action.Steps = []*automations.Step{step}
	action.Client = mqtt

	trigger := automations.NewDeviceTrigger("action")
	trigger.Actions = []automations.MqttAction{action}
	trigger.Conditions = []automations.Condition{condition}

	return trigger
}

func CreateDialTriggerStepActionBrightness(lightDeviceId string, dialDeviceId string, dialActionName string, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	condition := NewExposeCondition("action", dialActionName, "=")
	step := &automations.Step{}
	step.Id = lightDeviceId
	step.Operator = "+"
	step.Property = "brightness"

	step2 := &automations.Step{}
	step2.Id = dialDeviceId
	step2.Operator = "*"
	step2.Property = "action_time"

	action := &automations.MqttStepAction{}
	action.Id = lightDeviceId
	action.Property = "brightness"
	action.Type = "step"
	action.Data = 0.5
	action.Steps = []*automations.Step{step, step2}
	action.Client = mqtt

	trigger := automations.NewDeviceTrigger("action")
	trigger.Actions = []automations.MqttAction{action}
	trigger.Conditions = []automations.Condition{condition}

	return trigger
}

// "name": "action",
// "conditions": [
//
//	  {
//		"name": "action",
//		"value": "dial_rotate_right_slow",
//		"equality": "="
//	  }
//
// ],
//
//	"action": {
//	  "id": "0x00158d0005a23c38",
//	  "friendlyname": "Living Room",
//	  "property": "brightness",
//	  "type": "step",
//	  "data": 0.5,
//	  "steps": [
//		{
//		  "property": "brightness",
//		  "operator": "+",
//		  "id": "0x00158d0005a23c38"
//		},
//		{
//		  "property": "action_time",
//		  "operator": "*",
//		  "id": "0x001788010d7d9d3f"
//		}
//	  ]
//	}
//
// },
func CreateSwitchTriggerWithBindingAction(triggerName string, actionProp string, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	// action = turn off light
	brightnessAction := automations.NewTriggerAction()
	brightnessAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: actionProp,
			Data: 0,
		},
	}
	brightnessAction.Client = mqtt

	// Turn off sensor trigger
	button1Trigger := automations.NewDeviceTrigger(triggerName)
	button1Trigger.Actions = []automations.MqttAction{brightnessAction}

	return button1Trigger
}

// Conditions
func NewExposeConditionwithTimeRange(name string, value any, operation utils.EqualityOperator, timeRange *automations.TimeRange, clock utils.Clock) *automations.ExposeCondition {
	cond := &automations.ExposeCondition{
		Name:          name,
		Value:         value,
		BaseCondition: *automations.NewBaseCondition(automations.ExposeConditionType, operation, timeRange),
	}

	err := cond.InitHandlers(clock)
	if err != nil {
		utils.LogErrorf("initialising expose timeRange failed. Error: %s", err.Error())
		return nil
	}
	return cond
}


func NewExposeCondition(name string, value any, operation utils.EqualityOperator) *automations.ExposeCondition {
	
	cond := &automations.ExposeCondition{
		Name:          name,
		Value:         value,
		BaseCondition: *automations.NewBaseCondition(automations.ExposeConditionType, operation, nil),
	}

	err := cond.InitHandlers(utils.NewRealClock())
	if err != nil {
		utils.LogErrorf("error initialising expose condition %s", err.Error())
		return nil
	}

	return cond
}

func NewManualCondition(timeRange *automations.TimeRange, clock utils.Clock) *automations.ManualCondition {
	cond := &automations.ManualCondition{
		BaseCondition: *automations.NewBaseCondition(automations.ManualConditionType, utils.Equals, timeRange),
	}

	err := cond.InitHandlers(clock)
	if err != nil {
		utils.LogErrorf("error initialising expose condition %s", err.Error())
		return nil
	}

	return cond
}
