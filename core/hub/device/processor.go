package device

import (
	"errors"
	"node-herder/hub"
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

type NodePayload struct {
	Id                 string         `json:"id"`
	ConnectionType     string         `json:"conn"`
	PowerSource        string         `json:"power_source"`
	Sensors            map[string]any `json:"sensors"`
	Stats              map[string]any `json:"stats"`
	availabilityTicker time.Ticker
}

func NewNodePayload() *NodePayload {
	return &NodePayload{
		Sensors:            map[string]any{},
		Stats:              map[string]any{},
		availabilityTicker: time.Ticker{},
	}
}

type Processor interface {
	Process(id string, payload interface{}) error
}
type PayloadProcessor struct {
	eventHub hub.EventHub
	repo     Repository
}

func NewPayloadProcessor(repo Repository, eventHub hub.EventHub) Processor {
	return &PayloadProcessor{repo: repo, eventHub: eventHub}
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
		if !p.updateDevice(device, data) {
			return nil
		}
	}

	p.repo.Store(id, device)
	p.eventHub.Broadcast(hub.DeviceUpdated, device)

	return nil
}

func (p *PayloadProcessor) updateDevice(node *NodePayload, data map[string]interface{}) bool {
	var updated = false
	for key, currValue := range node.Sensors {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensors[key] = newValue
			updated = true
		}
	}

	return updated
}

func (p *PayloadProcessor) addDevice(id string, data map[string]interface{}) *NodePayload {
	// TODO:
	// have a timer to see if item is available, if not set offline
	// data["availability"] = "offline"
	// will need to syncronize data access

	// check if battery key exist, if not set source as mains
	if _, ok := data[batterKey]; !ok {
		data[powerSourceKey] = mainsKey
	}

	// Todo: need to pass in payload
	data[connectionTypeKey] = connectionTypeMqtt

	var newNode = NewNodePayload()
	newNode.Id = id
	for key, value := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensors[key] = value
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Stats[key] = value
		}
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
	return newNode
}

// ticker := time.NewTicker(500 * time.Millisecond)

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")
}
