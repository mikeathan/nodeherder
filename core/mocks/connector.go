package mocks

import (
	"fmt"
	"node-herder/hub"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type NopWsServer struct {
}

func (w *NopWsServer) Broadcast(eventName string, data interface{}) error {
	fmt.Println("Empty Broadcast")
	return nil
}

type NopMqttClient struct {
}

func (w *NopMqttClient) Connect() error {
	fmt.Println("Empty Connect")
	return nil
}

func (w *NopMqttClient) WithMessageHandler(messageHandler func(client mqtt.Client, msg mqtt.Message)) {
	fmt.Println("Empty WithMessageHandler")
}
func (w *NopMqttClient) AddTopic(topic string) {
	fmt.Println("Empty BroaAddTopicdcast")

}
func (w *NopMqttClient) Disconnect() {
	fmt.Println("Empty Disconnect")
}

type NopRepository struct {
}

func (w *NopMqttClient) Store(deviceName string, payload interface{}) {
	fmt.Println("Empty Store")
}
func (w *NopMqttClient) ListAllDevices() *hub.Device {
	fmt.Println("Empty ListAllDevices")
	return nil
}
