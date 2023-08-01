package devices

import (
	"errors"
	"fmt"
	"time"
)

var sensorWhitelist = map[string]int{
	"temperature":     1,
	"humidity":        2,
	"pressure":        3,
	"presence":        4,
	"illuminance_lux": 5,
}
var statsWhitelist = map[string]int{
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

const online = "online"
const offline = "offline"

type Device struct {
	Id                      string         `json:"id"`
	ConnectionType          string         `json:"conn"`
	PowerSource             string         `json:"power_source"`
	Sensors                 map[string]any `json:"sensors"`
	Stats                   map[string]any `json:"stats"`
	availabilityTicker      time.Ticker    `json:"-"`
	availablityDone         chan bool
	availabilityTimeoutSecs int
}

func newDevice() *Device {
	return &Device{
		Sensors:                 map[string]any{},
		Stats:                   map[string]any{},
		availabilityTicker:      time.Ticker{},
		availablityDone:         make(chan bool, 1),
		availabilityTimeoutSecs: 3600, // 1 hour check
	}
}

func CreateNewDevice(id string, data map[string]interface{}) *Device {
	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data[availabilityKey] = "online"

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
		} else if _, ok := statsWhitelist[key]; ok {
			newNode.Stats[key] = value
		}
	}
	return newNode
}
func (device *Device) Dispose() {
	if device.AvailabilityTimerRunning() {
		device.availablityDone <- true
		device.availabilityTicker.Stop()
	}
}

func (device *Device) AvailabilityTimerRunning() bool {
	return len(device.availablityDone) != 0
}

func (device *Device) StartAvailabilityTimer(timeoutInSecs int) {

	device.availabilityTicker = *time.NewTicker(1 * time.Second)

	go func() {
		defer close(device.availablityDone)

		for {
			select {
			case <-device.availablityDone:
				device.Stats[availabilityKey] = offline
				fmt.Println("timer killed")
				return

			case <-device.availabilityTicker.C:

				if device.Stats[availabilityKey] == offline {
					return
				}

				lastSeenStr, ok := device.Stats[lastSeenKey].(string)
				if !ok {
					panic("failed to convert")
				}

				lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
				now := time.Now()
				diff := now.Sub(lastSeen)
				if diff.Seconds() >= float64(timeoutInSecs) {

					device.Stats[availabilityKey] = offline

					fmt.Println("we offline!!!!!!!!!!!!!!!!!")
				}
			}
		}
	}()

}
func (node *Device) TryUpdateDevice(data map[string]interface{}) bool {
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}

	var updated = false
	for key, currValue := range node.Sensors {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensors[key] = newValue
			updated = true
		}
	}

	if node.Stats[availabilityKey] != online {
		node.Stats[availabilityKey] = online
		return true
	}

	return updated
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
