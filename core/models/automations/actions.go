package automations

import "node-herder/transport/mqtt"

type MqttAction struct {
	Friendlyname string `json:"friendlyname"`
	Type         string `json:"type"`
	Property     string `json:"name"`
	Value        any    `json:"value"`
	client       mqtt.MqttClient
}

func (a *MqttAction) Run() error {

	var payload []byte

	topic := a.Friendlyname // TODO: build topic
	a.client.Publish(topic, payload)

	return nil
}
