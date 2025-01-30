package utils_test

import (
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
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

	doorSensorTrigger := &automations.Trigger{}
	doorSensorTrigger.Name = "contact"
	doorSensorTrigger.Actions = []automations.MqttAction{alarmAction}

	deviceAutomation := automations.NewDevice("door sensor")
	deviceAutomation.Id = doorSensorId
	deviceAutomation.FriendlyName = "front door sensor"
	deviceAutomation.Enabled = true
	deviceAutomation.Triggers = []*automations.Trigger{doorSensorTrigger}

	return deviceAutomation
}

func CreateDialTriggerActionsBrightness(actionId string, dialActionName string, mqtt mqtt.MqttClient) *automations.Trigger {
	condition := &automations.Condition{}
	condition.EqualityOperator = "="
	condition.Value = dialActionName
	condition.Name = "action"

	step := &automations.Step{}
	step.Id = actionId
	step.Operator = "+"
	step.Property = "brightness"

	action := &automations.MqttStepAction{}
	action.Id = actionId
	action.Property = "brightness"
	action.Type = "StepAction"
	action.Data = 0.5
	action.Steps = []*automations.Step{step}
	action.Client = mqtt

	trigger := &automations.Trigger{}
	trigger.Name = "action"
	trigger.Actions = []automations.MqttAction{action}
	trigger.Conditions = []*automations.Condition{condition}

	return trigger
}

func CreateDialTriggerStepActionBrightness(lightDeviceId string, dialDeviceId string, dialActionName string, mqtt mqtt.MqttClient) *automations.Trigger {
	condition := &automations.Condition{}
	condition.EqualityOperator = "="
	condition.Value = dialActionName
	condition.Name = "action"

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
	action.Type = "StepAction"
	action.Data = 0.5
	action.Steps = []*automations.Step{step, step2}
	action.Client = mqtt

	trigger := &automations.Trigger{}
	trigger.Name = "action"
	trigger.Actions = []automations.MqttAction{action}
	trigger.Conditions = []*automations.Condition{condition}

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
//	  "type": "StepAction",
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
func CreateSwitchTriggerWithBindingAction(triggerName string, actionProp string, mqtt mqtt.MqttClient) *automations.Trigger {
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
	button1Trigger := &automations.Trigger{}
	button1Trigger.Name = triggerName
	button1Trigger.Actions = []automations.MqttAction{brightnessAction}

	return button1Trigger
}
