package automations

import (
	"node-herder/transport/mqtt"
	"node-herder/transport/ws"
	"time"
)

type Event struct {
	name  string
	value any
}

type DeviceAction struct {
	topic string
	event *Event
}

type AutomationProcessor struct {
	mqtt mqtt.MqttClient
	ws   ws.EventHub
}

func (a *AutomationProcessor) Process() {
}

func RegisterAutomations(mqtt mqtt.MqttClient, ws ws.EventHub) {
	//p := &AutomationProcessor{mqtt: mqtt, ws: ws}
	tt := TriggerFromDuration(5 * time.Second)
	tt.Repeat = false
	trigger := new(TimerTrigger)
	trigger.WithCondition(tt)
	trigger.WithAction("Action triggered!!!!!!!!!!!1")
	trigger.Process()

	// Device Action,
	// friendly_name for mqtt topic or full mqtt topic
	// event with type, value eg "state" = "ON"

	// all device info
	//  zigbee2mqtt/bridge/devices
	// device
	//  zigbee2mqtt/FRIENDLY_NAME/set
	//eg.{
	//   "state": "ON",
	//   "brightness": 215,
	//   "color_temp": 325
	// }

	//https://www.zigbee2mqtt.io/guide/usage/mqtt_topics_and_messages.html#zigbee2mqtt-bridge-request
}
