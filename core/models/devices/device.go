package devices

import (
	"errors"
	"time"
)

var sensorWhitelist = map[string]int{
	"temperature":     1,
	"humidity":        2,
	"pressure":        3,
	"presence":        4,
	"illuminance_lux": 5,
}
var deviceWhitelist = map[string]int{
	"battery":      1,
	"linkquality":  2,
	"availability": 3,
	"last_seen":    4,
}

var availabilityKey = "availability"
var mainsKey = "mains"
var powerSourceKey = "power_source"
var batterKey = "battery"
var lastSeenKey = "last_seen"
var connectionTypeKey = "conn"
var connectionTypeMqtt = "mqtt"

type Device struct {
	Id                 string         `json:"id"`
	ConnectionType     string         `json:"conn"`
	PowerSource        string         `json:"power_source"`
	Sensors            map[string]any `json:"sensors"`
	Stats              map[string]any `json:"stats"`
	availabilityTicker time.Ticker
}

func newDevice() *Device {
	return &Device{
		Sensors:            map[string]any{},
		Stats:              map[string]any{},
		availabilityTicker: time.Ticker{},
	}
}

func CreateNewDevice(id string, data map[string]interface{}) *Device {
	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data["availability"] = "online"

	// Todo: need to pass in payload
	data[connectionTypeKey] = connectionTypeMqtt

	var newNode = newDevice()
	newNode.Id = id
	newNode.ConnectionType = connectionTypeMqtt

	if _, ok := data[batterKey]; !ok {
		newNode.PowerSource = mainsKey
	} else {
		newNode.PowerSource = batterKey
	}

	for key, value := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensors[key] = value
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Stats[key] = value
		}
	}

	return newNode
}

func (node *Device) TryUpdateDevice(data map[string]interface{}) bool {
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data["availability"] = "online"
	var updated = false
	for key, currValue := range node.Sensors {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensors[key] = newValue
			updated = true
		}
	}

	return updated
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

// newNode.availabilityTicker = *time.NewTicker(1 * time.Hour)
// done := make(chan bool)

// go func() {
// 	for {
// 		select {
// 		// use context to kill goroutine
// 		case <-done:
// 			//newNode.Stats[availabilityKey] = "offline"
// 		case t := <-newNode.availabilityTicker.C:
// 			// if t >= newNode.last_seen
// 			// flag offline
// 			//newNode.Stats[availabilityKey] = "online"
// 		}
// 	}
// }()

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
