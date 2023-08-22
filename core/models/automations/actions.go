package automations

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/mqtt"
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
		a.Property: a.Value,
	}
	payload, err := json.Marshal(jp)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("%s/set", a.Friendlyname)
	a.client.Publish(msg, payload)
	return nil
}
