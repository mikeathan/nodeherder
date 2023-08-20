package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/automations"
	"node-herder/models/devices"
	"node-herder/utils/pool"
)

type processorTask struct {
	Id      string
	Payload map[string]interface{}
	Type    string
}

func (m *processorTask) OnFailure(err error) {
	fmt.Printf("Job: %s Error: %s", m.Id, err.Error())
}

type HubController struct {
	eventHub                     ws.EventHub
	mqtt                         mqtt.MqttClient
	repo                         devices.Repository
	ctx                          context.Context
	pool                         pool.WorkerPool
	procFunc                     func(t pool.Task) error
	AvailabilityTimeoutinSeconds int
}

// if we have handlers, in device handler
// run logic
// new data, store...
// else do nothing....

// pass callback to handler on  success, so on device success send mqtt
// on sth else success do sth else

func RegisterHubController(ws ws.EventHub, mqtt mqtt.MqttClient, repo devices.Repository, ctx context.Context) *HubController {
	h := &HubController{
		eventHub:                     ws,
		mqtt:                         mqtt,
		repo:                         repo,
		ctx:                          ctx,
		AvailabilityTimeoutinSeconds: 3600, // 1 hour
	}

	// TODO:
	// do i need to store procFunc as member variable
	h.procFunc = func(task pool.Task) error {
		pTask, ok := task.(*processorTask)
		if !ok {
			return errors.New("invalid task type")
		}

		return h.processPayload(pTask.Id, pTask.Type, pTask.Payload)
	}

	h.pool = *pool.NewWorkerPool(1, h.ctx)
	h.pool.Run(h.procFunc)

	h.eventHub.OnConnected(func() interface{} {
		return h.repo.ListAllDevices()
	})

	h.mqtt.OnMessageHandler(func(name string, payload []byte) {

		// TODO: needs refactoring
		// we need mqtt message handler
		if name == "bridge/devices" {

			devices, err := devices.LoadBridgeDevices(payload)
			if err != nil {
				fmt.Println("parsing devices error: ", err.Error())
				return
			}
			if !automations.IsConfigured() {
				automations.Load(devices, h.mqtt)
				if err != nil {
					fmt.Println("loading automations error: ", err.Error())
					return
				}
			}
		} else if name == "bridge/logging" {
			// todo: handle
			fmt.Println(string(payload))
		} else {
			dataMap, err := convertToMap(payload)
			if err != nil {
				fmt.Println("error: failed to convert mqtt payload to map")
				return
			}
			h.Enqueue(name, dataMap, "mqtt")
		}
	})

	// setup
	h.mqtt.Connect()
	h.mqtt.Publish("zigbee2mqtt/bridge/devices", nil) // get devices for setup stuff
	return h
}

func (c *HubController) Enqueue(name string, payload map[string]interface{}, connType string) {
	c.pool.AddTask(&processorTask{Id: name, Payload: payload, Type: connType})
}

func (c *HubController) processPayload(id string, connType string, payload map[string]interface{}) error {

	var err error
	device, _ := c.repo.FindDevice(id)
	if device == nil {
		device, err = devices.CreateNewDevice(id, connType, payload)
		if err != nil {
			return err
		}
		device.StartAvailabilityTimer(c.AvailabilityTimeoutinSeconds)
	} else {
		if !device.TryUpdateDevice(payload) {
			return nil
		}
	}

	c.repo.Store(id, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device)
	return nil
}

func convertToMap(payload []byte) (map[string]interface{}, error) {

	deviceMap := make(map[string]interface{})
	err := json.Unmarshal(payload, &deviceMap)
	if err != nil {
		return nil, errors.New("invalid device data")
	}
	return deviceMap, nil
}
