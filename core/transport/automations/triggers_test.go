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

type BridgeDevice struct {
	DateCode   string `json:"date_code"`
	Definition struct {
		Description string `json:"description"`
		Exposes     []struct {
			Features []struct {
				Access      int    `json:"access"`
				Description string `json:"description"`
				Name        string `json:"name"`
				Property    string `json:"property"`
				Type        string `json:"type"`
				ValueOff    string `json:"value_off,omitempty"`
				ValueOn     string `json:"value_on,omitempty"`
				ValueToggle string `json:"value_toggle,omitempty"`
				ValueMax    any    `json:"value_max,omitempty"`
				ValueMin    any    `json:"value_min,omitempty"`
				Presets     []struct {
					Description string `json:"description"`
					Name        string `json:"name"`
					Value       int    `json:"value"`
				} `json:"presets,omitempty"`
				Unit string `json:"unit,omitempty"`
			} `json:"features,omitempty"`
			Type        string   `json:"type"`
			Access      int      `json:"access,omitempty"`
			Description string   `json:"description,omitempty"`
			Name        string   `json:"name,omitempty"`
			Property    string   `json:"property,omitempty"`
			Values      []string `json:"values,omitempty"`
			Unit        string   `json:"unit,omitempty"`
			ValueMax    any      `json:"value_max,omitempty"`
			ValueMin    any      `json:"value_min,omitempty"`
		} `json:"exposes"`
		Model   string `json:"model"`
		Options []struct {
			Access      int    `json:"access"`
			Description string `json:"description"`
			Name        string `json:"name"`
			Property    string `json:"property"`
			Type        string `json:"type"`
			ValueMin    int    `json:"value_min,omitempty"`
			ValueOff    bool   `json:"value_off,omitempty"`
			ValueOn     bool   `json:"value_on,omitempty"`
		} `json:"options"`
		SupportsOta bool   `json:"supports_ota"`
		Vendor      string `json:"vendor"`
	} `json:"definition"`
	Disabled           bool   `json:"disabled"`
	FriendlyName       string `json:"friendly_name"`
	IeeeAddress        string `json:"ieee_address"`
	InterviewCompleted bool   `json:"interview_completed"`
	Interviewing       bool   `json:"interviewing"`
	Manufacturer       string `json:"manufacturer"`
	ModelID            string `json:"model_id"`
	NetworkAddress     int    `json:"network_address"`
	PowerSource        string `json:"power_source"`
	SoftwareBuildID    string `json:"software_build_id"`
	Supported          bool   `json:"supported"`
	Type               string `json:"type"`
}

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

			var bridgeDevice []BridgeDevice
			err := json.Unmarshal(payload, &bridgeDevice)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}
			for _, value := range bridgeDevice {

				for _, e := range value.Definition.Exposes {

					if e.Type == "light" {

						fmt.Println("FriendlyName:", value.FriendlyName)
						fmt.Println("IeeeAddress:", value.IeeeAddress)
						fmt.Println("PowerSource:", value.PowerSource)
						fmt.Println("ModelID:", value.ModelID)
						fmt.Println("Type:", e.Type)
						fmt.Println("Type:", value.Type)

						fmt.Println("Features")
						for _, f := range e.Features {
							if f.Name == "state" {
								fmt.Println("Name:", f.Name)
								fmt.Println("Description:", f.Description)
								fmt.Println("Type:", f.Type)
								if f.Type == "binary" {
									fmt.Println("value_off:", f.ValueOff)
									fmt.Println("value_on:", f.ValueOn)
								}
								fmt.Println("---------------------")
							}
						}
					}
				}

			}
			// var deviceMap []map[string]interface{}
			// err := json.Unmarshal(payload, &deviceMap)
			// if err != nil {
			// 	fmt.Println("error: ", err.Error())
			// }

			// for _, value := range deviceMap {
			// 	friendly_name := value["friendly_name"]

			// 	if friendly_name == "Hive light 1" {
			// 		definition := value["definition"]

			// 		def, _ := definition.(map[string]interface{})

			// 		exposes := def["exposes"]
			// 		// ???
			// 		ex, _ := exposes.([]map[string]interface{})
			// 		fmt.Println(ex) // //byteKey := []byte(fmt.Sprintf("%v", definition))
			// 		// var features Features
			// 		// err := json.Unmarshal(byteKey, &features)
			// 		// if err != nil {
			// 		// 	fmt.Println("error: ", err.Error())
			// 		// }

			// 	}
			// }
			// // err := bh.ProcessMessage(payload)
			// // if err != nil {
			// // 	fmt.Println("error: ", err.Error())
			// // }
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
