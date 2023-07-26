package services

import (
	"context"
	"errors"
	"fmt"
	"node-herder/device"
	"node-herder/hub"
	"node-herder/hub/pool"
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

type Processor interface {
	Equeue(id string, payload interface{}) error
}

type processorTask struct {
	Id      string
	Payload interface{}
}

func (m *processorTask) OnFailure(err error) {
	fmt.Printf("Job: %s Error: %s", m.Id, err.Error())
}

type PayloadProcessor struct {
	eventHub   hub.EventHub
	repo       device.Repository
	workerPool pool.WorkerPool
	ctx        context.Context
	procFunc   func(t pool.Task) error
}

func NewPayloadProcessor(repo device.Repository, eventHub hub.EventHub, ctx context.Context) Processor {
	p := &PayloadProcessor{
		repo:     repo,
		eventHub: eventHub,
		ctx:      ctx,
	}
	p.procFunc = func(t pool.Task) error {
		return p.process(t)
	}

	p.workerPool = *pool.NewWorkerPool(1, ctx, p.procFunc)
	p.workerPool.Start()
	return p
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func (p *PayloadProcessor) process(task pool.Task) error {
	pTask, ok := task.(*processorTask)
	if !ok {
		return errors.New("invalid task type")
	}
	id := pTask.Id
	payload := pTask.Payload
	data, err := convertToMap(payload)
	if err != nil {
		return errors.New("invalid payload format")
	}

	// sanitize payload,
	// TODO: need optimization
	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data["availability"] = "online"

	device, _ := p.repo.FindDevice(id)
	if device != nil && !p.updateDevice(device, data) {
		return nil
	} else {
		device = p.addDevice(id, data)
	}
	p.repo.Store(id, device)
	p.eventHub.Broadcast(hub.DeviceUpdated, device)

	return nil
}

func (p *PayloadProcessor) Equeue(id string, payload interface{}) error {

	return p.workerPool.AddTask(&processorTask{Id: id, Payload: payload})
}

func (p *PayloadProcessor) updateDevice(node *device.Payload, data map[string]interface{}) bool {
	var updated = false
	for key, currValue := range node.Sensors {
		if newValue, ok := data[key]; ok && newValue != currValue {
			node.Sensors[key] = newValue
			updated = true
		}
	}

	return updated
}

func (p *PayloadProcessor) addDevice(id string, data map[string]interface{}) *device.Payload {
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

	var newNode = device.NewPayload()
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
