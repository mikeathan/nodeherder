package automations_test

import (
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/models/automations"
	"node-herder/models/devices"
	"testing"
	"time"
)

// https://www.home-assistant.io/docs/automation/basics/
// https://www.home-assistant.io/docs/automation/editor/
// mosquitto_pub -h 192.168.179:1883 -u sinkhole -P mqtt2023 -t 'zigbee2mqtt/Hive light 1/set' -m '{ "state": "ON" }'

func TestMqttAction(t *testing.T) {
	mqttConfig := mqtt.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   "192.168.50.179:1883",
		Topics: []string{
			"bridge/devices",
			"bridge/logging",
		},
	}
	mqtt := mqtt.NewMqttClient(mqttConfig)

	var messageHandler = func(id string, payload []byte) {
		if id == "bridge/devices" {

			devices, err := devices.LoadBridgeDevices(payload)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			err = automations.Load(devices, mqtt)
			if err != nil {
				fmt.Println("error loading automations: ", err.Error())
			}

			// fmt.Println(device)
		} else if id == "bridge/logging" {

			fmt.Println(string(payload))
			// handle error ?
			//{"level":"error",
			//"message":"Publish 'set' 'state' to 'Hive light 1' failed: 'Error: Command 0x70ac08fffefafeca/1 genOnOff.off({}, {\"sendWhen\":\"immediate\",\"timeout\":10000,\"disableResponse\":false,\"disableRecovery\":false,\"disableDefaultResponse\":false,\"direction\":0,\"srcEndpoint\":null,\"reservedBits\":0,\"manufacturerCode\":null,\"transactionSequenceNumber\":null,\"writeUndiv\":false}) failed (Data request failed with error: 'MAC no ack' (233))'"}
		}
	}

	mqtt.OnMessageHandler(messageHandler)
	err := mqtt.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mqtt.Publish("zigbee2mqtt/bridge/devices", nil)

	time.Sleep(300 * time.Second)

	fmt.Println("finish")
}
