package utils_test

import (
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
)

func CreateDialTriggerActionsBrightness(actionId string, dialActionName string, mqtt mqtt.MqttClient) *automations.Trigger {
	condition := &automations.Condition{}
	condition.EqualityOperator = "="
	condition.Value = dialActionName
	condition.Name = "action"

	step := &automations.Step{}
	step.Id = actionId
	step.Operator = "="
	step.Property = "brightness"

	action := &automations.MqttAction{}
	action.Id = actionId
	action.FriendlyName = "Attic light"
	action.Property = "brightness"
	action.Type = "StepAction"
	action.Data = 0.5
	action.Steps = []automations.Step{*step}
	action.Client = mqtt

	trigger := &automations.Trigger{}
	trigger.Name = "action"
	trigger.Action = action
	trigger.Conditions = []*automations.Condition{condition}

	return trigger
}

func CreateSwitchTriggerWithBindingAction(triggerName string, actionProp string, mqtt mqtt.MqttClient) *automations.Trigger {
	// action = turn off light
	brightnessAction := &automations.MqttAction{}
	brightnessAction.FriendlyName = "Attic light"
	brightnessAction.Property = actionProp
	brightnessAction.Client = mqtt

	// Turn off sensor trigger
	button1Trigger := &automations.Trigger{}
	button1Trigger.Name = triggerName
	button1Trigger.Action = brightnessAction

	return button1Trigger
}
