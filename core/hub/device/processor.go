package device

import (
	"errors"
	"node-herder/hub"
	"node-herder/models"
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
	"availability": 4,
}

var lastSeenKey = "last_seen"
var batterKey = "battery"
var mainsKey = "Mains (single phase)"
var powerSourceKey = "power_source"

type NodePayload struct {
	Id     string
	Sensor map[string]any `json:"sensor"`
	Device map[string]any `json:"device"`
}

func NewNodePayload() *NodePayload {
	return &NodePayload{
		Sensor: map[string]any{},
		Device: map[string]any{},
	}
}

type Processor interface {
	Process(payload interface{}) error
}
type PayloadProcessor struct {
	repo hub.Repository
}

// todo:
// collect data and pass them to different process
func (p *PayloadProcessor) Process(payload interface{}) error {

	data, err := convertToMap(payload)
	if err != nil {
		return errors.New("invalid paylaod format")
	}
	if _, ok := data["name"]; !ok {
		return errors.New("no name key in payload")
	}

	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data["lastSeenKey"] = time.Now() // TODO: fix format
	}
	data["availability"] = "online"
	//

	deviceName := data["name"].(string)
	device, _ := p.repo.FindDevice(deviceName)

	if device == nil {
		p.addDevice(deviceName, data)
		return nil
	}

	p.updateDevice(device, data)
	return nil
}

func (p *PayloadProcessor) updateDevice(device *models.Device, data map[string]interface{}) {
	node := device.Payload.(*NodePayload)
	for key, value := range node.Sensor {
		if val, ok := data[key]; ok && val == value {
			p.repo.StoreObject(node.Id, data)
			// BROADCAST
			break
		}
	}
}

func (p *PayloadProcessor) addDevice(deviceName string, data map[string]interface{}) {
	// TODO:
	// have a timer to see if item is available, if not set offline
	// data["availability"] = "offline"
	// will need to syncronize data access

	data[powerSourceKey] = batterKey

	// check if battery key exist, if not set power_source as mains
	if _, ok := data[batterKey]; !ok {
		data[powerSourceKey] = mainsKey
	}

	var newNode = NewNodePayload()
	newNode.Id = deviceName

	for key, v := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensor[key] = v
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Device[key] = v
		}
	}

	p.repo.StoreObject(deviceName, newNode)
}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
