package controllers

import (
	"context"
	"fmt"
	"node-herder/models/automations"
	"node-herder/models/bridge"
	"node-herder/transport/mqtt"
	"node-herder/utils/pool"
)

type messageHandler struct {
	mqtt        mqtt.MqttClient
	ctx         context.Context
	wp          *pool.WorkerPool
	processFunc func(task pool.Task) error
}

func newMessageHandler(mqtt mqtt.MqttClient, ctx context.Context) *messageHandler {
	h := &messageHandler{
		mqtt: mqtt,
		ctx:  ctx,
		wp:   &pool.WorkerPool{},
	}
	return h
}

func (m *messageHandler) Register(processFunc func(task pool.Task) error) error {
	m.processFunc = processFunc

	m.mqtt.OnMessageHandler(func(name string, payload []byte) {

		// TODO: needs refactoring
		// we need mqtt message handler
		if name == "bridge/devices" {

			devices, err := bridge.Parse(payload)
			if err != nil {
				fmt.Println("parsing devices error: ", err.Error())
				return
			}
			if !automations.IsConfigured() {
				automations.Load(devices, m.mqtt)
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
			m.Enqueue(name, dataMap, "mqtt")
		}
	})

	return nil
}

func (m *messageHandler) Enqueue(name string, payload map[string]interface{}, connType string) {
	m.wp.AddTask(&processorTask{Id: name, Payload: payload, Type: connType})
}
