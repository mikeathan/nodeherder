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
	"linkquality":  2,
	"power_source": 3,
	"availability": 4,
	"last_seen":    5,
}

var lastSeenKey = "last_seen"
var batterKey = "battery"
var mainsKey = "Mains (single phase)"

type NodePayload struct {
	Id          string
	PowerSource string
	checksum    string
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
	eventHub hub.EventHub
	repo     Repository
	hasher   Crc32Hasher
}

func NewPayloadProcessor(repo Repository, eventHub hub.EventHub) Processor {
	return &PayloadProcessor{repo: repo, eventHub: eventHub, hasher: *NewCrc32Hasher()}
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
		device = p.addDevice(id, data)
	} else {
		p.updateDevice(device, data)
	}

	h := p.hasher.CalculateHash()
	if h != "" && h != device.checksum {
		device.checksum = h
		p.hasher.Reset()

		p.repo.Store(id, device)
		p.eventHub.Broadcast(hub.DeviceUpdated, device)
	}

	return nil
}

func (p *PayloadProcessor) updateDevice(node *NodePayload, data map[string]interface{}) {
	for key, currValue := range node.Sensor {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensor[key] = newValue
			p.hasher.Write(newValue)
		}
	}
}

func (p *PayloadProcessor) addDevice(id string, data map[string]interface{}) *NodePayload {
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
	newNode.Id = id
	newNode.PowerSource = powerSource

	for key, value := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensor[key] = value
			p.hasher.Write(value)
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Device[key] = value
		}
	}

	return newNode
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
