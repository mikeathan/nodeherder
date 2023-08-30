package controllers

import (
	"fmt"
	"node-herder/internal/automations"
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

type bridgeLoggingHandler struct {
	ws ws.EventHub
}

func newBridgeLoggingHandler(ws ws.EventHub) *bridgeLoggingHandler {
	return &bridgeLoggingHandler{ws: ws}
}

func (b *bridgeLoggingHandler) ProcessPayload(id string, connType string, payload []byte) error {

	// if id == "bridge/devices" {
	// 	// TODO: maybe we update bridge devices later
	// 	// for now do nothing

	// } else
	if id == "bridge/logging" {
		// todo: handle
		fmt.Println(string(payload))
	}
	return nil
}

type deviceHandler struct {
	AvailabilityTimeoutInSeconds int
	repo                         devices.Repository
	eventHub                     ws.EventHub
	automationEngine             automations.Engine
}

func newDeviceHandler(repo devices.Repository, eventHub ws.EventHub, automationEngine automations.Engine) *deviceHandler {
	return &deviceHandler{
		repo:                         repo,
		eventHub:                     eventHub,
		automationEngine:             automationEngine,
		AvailabilityTimeoutInSeconds: 3600, // 1 Hour
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

		device.StartAvailabilityTimer(c.AvailabilityTimeoutInSeconds, func() {
			c.eventHub.Broadcast(ws.DeviceUpdated, device)
		})
	} else {
		if !device.TryUpdateDevice(dataMap) {
			return nil
		}
	}

	c.repo.Store(id, device)
	c.eventHub.Broadcast(ws.DeviceUpdated, device)

	// check to see if we have an automation for current device
	c.automationEngine.HandleDevice(device)
	return nil
}
