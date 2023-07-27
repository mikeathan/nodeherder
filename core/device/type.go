package device

import "time"

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

type Repository interface {
	Store(deviceName string, payload *Payload)
	ListAllDevices() []*Payload
	FindDevice(deviceName string) (*Payload, error)
}

type Payload struct {
	Id                 string         `json:"id"`
	ConnectionType     string         `json:"conn"`
	PowerSource        string         `json:"power_source"`
	Sensors            map[string]any `json:"sensors"`
	Stats              map[string]any `json:"stats"`
	availabilityTicker time.Ticker
}

func NewPayload() *Payload {
	return &Payload{
		Sensors:            map[string]any{},
		Stats:              map[string]any{},
		availabilityTicker: time.Ticker{},
	}
}

func CreateDevicePayload(id string, data map[string]interface{}) *Payload {
	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data["availability"] = "online"
	if _, ok := data[batterKey]; !ok {
		data[powerSourceKey] = mainsKey
	}

	// Todo: need to pass in payload
	data[connectionTypeKey] = connectionTypeMqtt

	var newNode = NewPayload()
	newNode.Id = id
	for key, value := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensors[key] = value
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Stats[key] = value
		}
	}

	return newNode
}

func TryUpdateDevicePayload(node *Payload, data map[string]interface{}) bool {
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
