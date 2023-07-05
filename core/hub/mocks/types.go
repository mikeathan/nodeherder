package mocks

import (
	"fmt"
	"node-herder/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

type MockEventHub struct {
	MockBroadcastEvent func(eventName string, data interface{}) error
}

func (w *MockEventHub) Broadcast(eventName string, data interface{}) error {
	return w.MockBroadcastEvent(eventName, data)
}

func (w *MockEventHub) RegisterNewClient(conn *websocket.Conn) {
	fmt.Println("Empty RegisterNewClient")
}

func (w *MockEventHub) OnConnected(onConnected func() interface{}) {
	fmt.Println("Empty OnConnected")
}

type NopWsServer struct {
}

func (w *NopWsServer) Broadcast(eventName string, data interface{}) error {
	fmt.Println("Empty Broadcast")
	return nil
}

func (w *NopWsServer) RegisterNewClient(conn *websocket.Conn) {
	fmt.Println("Empty RegisterNewClient")
}

func (w *NopWsServer) OnConnected(onConnected func() interface{}) {
	fmt.Println("Empty OnConnected")
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

func (m *NopMqttClient) OnMessageHandler(handler func(client mqtt.Client, msg mqtt.Message)) {
	fmt.Println("Empty OnMessageHandler")
}

type NopRepository struct {
}

func (w *NopRepository) StoreJson(deviceName string, payload []byte) error {
	fmt.Println("Empty StoreJson")
	return nil
}

func (w *NopRepository) StoreObject(deviceName string, payload interface{}) error {
	fmt.Println("Empty StoreObject")
	return nil
}

func (w *NopRepository) ListAllDevices() []*models.Device {

	fmt.Println("Empty ListAllDevices")
	return []*models.Device{}
}

func (w *NopRepository) FindDevice(deviceName string) (*models.Device, error) {

	fmt.Println("Empty FindDevice")
	return &models.Device{}, nil
}
