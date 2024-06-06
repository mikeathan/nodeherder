package utils_test

import (
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
)

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
