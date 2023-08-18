package automations

import (
	"encoding/json"
	"node-herder/transport/mqtt"
)

type MqttAction struct {
	Friendlyname string `json:"friendlyname"`
	Type         string `json:"type"`
	Property     string `json:"name"`
	Value        any    `json:"value"`
	client       mqtt.MqttClient
}

func (a *MqttAction) Run() error {

	jp := map[string]any{
		a.Property: a.Property,
	}
	payload, err := json.Marshal(jp)
	if err != nil {
		return err
	}

	a.client.Publish(a.Friendlyname, payload)
	return nil
}
