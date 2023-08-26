package controllers

import (
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
)

type messageTask struct {
	Id      string
	Payload []byte
	Type    string
	h       handler
}

func (m *messageTask) OnFailure(err error) {
	utils.LogErrorf("Job: %s Error: %s", m.Id, err.Error())
}

func (m *messageTask) Process() error {
	return m.h.ProcessPayload(m.Id, m.Type, m.Payload)
}

type handler interface {
	ProcessPayload(id string, connType string, payload []byte) error
}

type bridgeHandler struct {
	mqtt       mqtt.MqttClient
	configured bool
}

func newBridgeHandler(mqtt mqtt.MqttClient) *bridgeHandler {
	return &bridgeHandler{
		mqtt: mqtt,
	}
}

func (b *bridgeHandler) ProcessPayload(id string, connType string, payload []byte) error {

	if id == "bridge/devices" {

		devices, err := devices.LoadBridgeDevices(payload)
		if err != nil {
			return err
		}

		if !b.configured {
			// WIP
			// automations.Load(devices, b.mqtt)
			// if err != nil {
			// 	fmt.Println("loading automations error: ", err.Error())
			// 	return err
			// }

			for _, device := range devices {
				if device.Disabled || device.Type == "Coordinator" || !device.InterviewCompleted {
					continue
				}

				err := b.mqtt.AddTopic(device.FriendlyName)
				if err != nil {
					utils.LogErrorf("error %s conffigure topic %s", device.FriendlyName, err.Error())
				}
			}
			b.configured = true
		}

	} else if id == "bridge/logging" {
		// todo: handle
		fmt.Println(string(payload))
	}
	return nil
}

type deviceHandler struct {
	AvailabilityTimeoutinSeconds int
	repo                         devices.Repository
	eventHub                     ws.EventHub
}

func newDeviceHandler(repo devices.Repository, eventHub ws.EventHub) *deviceHandler {
	return &deviceHandler{
		repo:                         repo,
		eventHub:                     eventHub,
		AvailabilityTimeoutinSeconds: 3600, // 1 Hour
	}
}

func (c *deviceHandler) ProcessPayload(id string, connType string, payload []byte) error {

	dataMap, err := convertToMap(payload)
	if err != nil {
		return err
	}

	device, _ := c.repo.FindDevice(id)
	if device == nil {
		device, err = devices.CreateNewDevice(id, connType, dataMap)
		if err != nil {
			return err
		}
		device.StartAvailabilityTimer(c.AvailabilityTimeoutinSeconds)
	} else {
		if !device.TryUpdateDevice(dataMap) {
			return nil
		}
	}

	c.repo.Store(id, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device) // ????
	return nil
}
