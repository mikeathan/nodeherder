package device

import (
	"errors"
	"node-herder/hub"
	"time"
)

// sensor whitelist map
var powerSource = map[string]int{
	"battery":              1,
	"Mains (single phase)": 2,
	"DC source":            3,
}

var sensorWhitelist = map[string]int{
	"temperature":     1,
	"humidity":        2,
	"pressure":        3,
	"presence":        4,
	"illuminance_lux": 5,
}
var deviceWhitelist = map[string]int{
	"battery":      1,
	"linquality":   2,
	"power_source": 3,
}

var lastSeenKey = "last_seen"
var batterKey = "battery"
var mainsKey = "Mains (single phase)"
var powerSourceKey = "power_source"

var repo = hub.Repository

type NodePayload struct {
	Sensor map[string]any `json:"sensor"`
	Device map[string]any `json:"device"`
}

func NewNodePayload() *NodePayload {
	return &NodePayload{
		Sensor: map[string]any{},
		Device: map[string]any{},
	}
}

// todo:
// collect data and pass them to different process
func Process(payload interface{}) {

	data, err := convertToMap(payload)
	if err != nil {
		panic("fooked")
	}
	if _, ok := data["name"]; !ok {
		panic("invalid payload")
	}
	name := data["name"]
	device := repo.FindDevice(name)
	if device == nil {

		data[powerSourceKey] = batterKey
		if _, ok := data[batterKey]; !ok {
			data[powerSourceKey] = mainsKey
		}

		var newNode = NewNodePayload()
		// new device
		for key, v := range data {
			if _, ok := sensorWhitelist[key]; ok {
				// store in sensor data
			} else if _, ok := deviceWhitelist[key]; ok {
				// store in device data
			}
		}
	}

	// todo
	// have a timer to see if item is available, if not set offline

	data["availability"] = "offline"
	if lastSeen, ok := data[lastSeenKey]; ok {
		data["availability"] = "online"
		// if first time, store
		// else get previous item and compare if anything changed  ?
		// if sth changed then broadcast
		// else flag as online and dont broadcast

	} else {
		data["lastSeenKey"] = time.Now() // fix format
	}

}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")

}
