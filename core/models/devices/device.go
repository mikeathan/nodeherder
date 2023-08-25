package devices

import (
	"errors"
	"fmt"
	"time"
)

var sensorWhitelist = map[string]int{
	"temperature":       1,
	"humidity":          2,
	"pressure":          3,
	"presence":          4,
	"illuminance_lux":   5,
	"occupancy":         6,
	"brightness":        7,
	"state":             8,
	"color_temp":        9,
	"air_quality_score": 10,
	"pm1":               11,
	"pm25":              12,
	"pm10":              13,
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
var connectionTypeMqtt = "mqtt"

const online = "online"
const offline = "offline"

type Device struct {
	Id                      string         `json:"id"`
	ConnectionType          string         `json:"conn"`
	PowerSource             string         `json:"power_source"`
	Sensors                 map[string]any `json:"sensors"`
	Stats                   map[string]any `json:"stats"`
	availabilityTicker      time.Ticker
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

func CreateNewDevice(id string, connType string, data map[string]interface{}) (*Device, error) {
	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data[availabilityKey] = online

	var newNode = newDevice()
	newNode.Id = id
	newNode.ConnectionType = connType

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

	if len(newNode.Sensors) == 0 {
		return nil, errors.New("invalid payload - no sensor data")
	}

	return newNode, nil
}

func (device *Device) Dispose() {
	device.availablityDone <- true
	device.availabilityTicker.Stop()
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

				lastSeenStr, _ := device.Stats[lastSeenKey].(string)
				lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
				if err != nil {
					fmt.Println("failed to parse time", err)
					device.Dispose()
				}

				now := time.Now()
				diff := now.Sub(lastSeen)
				if diff.Seconds() >= float64(timeoutInSecs) {
					device.Stats[availabilityKey] = offline
				}
			}
		}
	}()
}

func (node *Device) TryUpdateDevice(data map[string]interface{}) bool {

	// TODO: needs refactoring. maybe keep reference and assign only if changes arefound
	//
	var updated = false
	for key, currValue := range node.Sensors {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensors[key] = newValue
			updated = true
		}
	}

	if node.Stats[availabilityKey] != online {
		node.Stats[availabilityKey] = online
		updated = true
	}

	if updated {
		if _, ok := data[lastSeenKey]; !ok {
			data[lastSeenKey] = getCurrentTime()
		}
		node.Stats[lastSeenKey] = data[lastSeenKey]
	}
	return updated
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
