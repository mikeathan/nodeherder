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

var lastSeenKey = "last_seen"

type NodePayload struct {
	Id       string `json:"id"`
	checksum string
	Sensors  map[string]any `json:"sensors"`
	Stats    map[string]any `json:"stats"`
}

func NewNodePayload() *NodePayload {
	return &NodePayload{
		Sensors: map[string]any{},
		Stats:   map[string]any{},
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
	for key, currValue := range node.Sensors {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensors[key] = newValue
			p.hasher.Write(newValue)
		}
	}
}

func (p *PayloadProcessor) addDevice(id string, data map[string]interface{}) *NodePayload {
	// TODO:
	// have a timer to see if item is available, if not set offline
	// data["availability"] = "offline"
	// will need to syncronize data access

	var newNode = NewNodePayload()
	newNode.Id = id
	for key, value := range data {
		if _, ok := sensorWhitelist[key]; ok {
			newNode.Sensors[key] = value
			p.hasher.Write(value)
		} else if _, ok := deviceWhitelist[key]; ok {
			newNode.Stats[key] = value
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
