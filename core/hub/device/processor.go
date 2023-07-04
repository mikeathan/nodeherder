package device

import (
	"errors"
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
	"linkquality":  2,
	"power_source": 3,
	"availability": 4,
	"last_seen":    5,
}

var lastSeenKey = "last_seen"
var batterKey = "battery"
var mainsKey = "Mains (single phase)"
var powerSourceKey = "power_source"

type NodePayload struct {
	Id          string
	PowerSource string
	Sensor      map[string]any `json:"sensor"`
	Device      map[string]any `json:"device"`
}

func NewNodePayload() *NodePayload {
	return &NodePayload{
		Sensor: map[string]any{},
		Device: map[string]any{},
	}
}

type Processor interface {
	Process(id string, payload interface{}) error
}
type PayloadProcessor struct {
	repo Repository
}

func NewPayloadProcessor(repo Repository) Processor {
	return &PayloadProcessor{repo: repo}
}
func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

// todo:
// collect data and pass them to different process
func (p *PayloadProcessor) Process(id string, payload interface{}) error {

	data, err := convertToMap(payload)
	if err != nil {
		return errors.New("invalid paylaod format")
	}

	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data["availability"] = "online"

	device, _ := p.repo.FindDevice(id)

	if device == nil {
		p.addDevice(id, data)
		return nil
	}

	p.updateDevice(device, data)
	return nil
}

func (p *PayloadProcessor) updateDevice(node *NodePayload, data map[string]interface{}) {
	for key, value := range node.Sensor {
		if val, ok := data[key]; ok && val == value {
			node.Sensor[key] = val
		}
	}
	// need to check if anything has change first, else we will be doing that every time
	//p.repo.Store(node.Id, node)
	// BROADCAST
}

func (p *PayloadProcessor) addDevice(deviceName string, data map[string]interface{}) {
	// TODO:
	// have a timer to see if item is available, if not set offline
	// data["availability"] = "offline"
	// will need to syncronize data access

	powerSource := batterKey

	// check if battery key exist, if not set power_source as mains
	if _, ok := data[batterKey]; !ok {
		powerSource = mainsKey
	}

	var newNode = NewNodePayload()
	newNode.Id = deviceName
	newNode.PowerSource = powerSource

	for key, v := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensor[key] = v
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Device[key] = v
		}
	}

	p.repo.Store(deviceName, newNode)
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
