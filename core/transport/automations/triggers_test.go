package automations_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	automation "node-herder/transport/automations"
	"node-herder/transport/mqtt"
	"testing"
	"time"
)

func TestTriggers(t *testing.T) {

	tt := automation.TriggerFromDuration(5 * time.Second)
	tt.Repeat = false
	trigger := new(automation.TimerTrigger)
	trigger.WithCondition(tt)
	trigger.WithAction("Action triggered!!!!!!!!!!!1")
	trigger.Process()

	// https://www.home-assistant.io/docs/automation/basics/
	// https://www.home-assistant.io/docs/automation/editor/
}

type Features struct {
	//Type string `json:"type"`
	Data map[string]interface{}
}

func TestMqttAction(t *testing.T) {
	mqttConfig := mqtt.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   "192.168.50.179:1883",
		Topics: []string{
			"bridge/devices",
		},
	}
	//\bh := &automations.BridgeHandler{}
	var messageHandler = func(id string, payload []byte) {
		if id == "bridge/devices" {
			var deviceMap []map[string]interface{}
			err := json.Unmarshal(payload, &deviceMap)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			for _, value := range deviceMap {
				friendly_name := value["friendly_name"]

				if friendly_name == "Hive light 1" {
					definition := value["definition"]

					def, _ := definition.(map[string]interface{})

					exposes := def["exposes"]
					// ???
					ex, _ := exposes.([]map[string]interface{})
					fmt.Println(ex) // //byteKey := []byte(fmt.Sprintf("%v", definition))
					// var features Features
					// err := json.Unmarshal(byteKey, &features)
					// if err != nil {
					// 	fmt.Println("error: ", err.Error())
					// }

				}
			}
			// err := bh.ProcessMessage(payload)
			// if err != nil {
			// 	fmt.Println("error: ", err.Error())
			// }
		}
	}

	mqtt := mqtt.NewMqttClient(mqttConfig)
	mqtt.OnMessageHandler(messageHandler)
	err := mqtt.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mqtt.Publish("zigbee2mqtt/bridge/devices", nil)

	time.Sleep(10 * time.Second)

	fmt.Println("finish")
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
