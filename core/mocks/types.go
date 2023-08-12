package mocks

import (
	"fmt"
	"node-herder/models/devices"

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

type MockMqttClient struct {
	messageHandler func(string, []byte)
}

func (m *MockMqttClient) Connect() error {
	fmt.Println("Mock Connect")
	return nil
}

func (m *MockMqttClient) WithMessageHandler(messageHandler func(client mqtt.Client, msg mqtt.Message)) {
	fmt.Println("Mock WithMessageHandler")
}
func (m *MockMqttClient) AddTopic(topic string) {
	fmt.Println("Mock BroaAddTopicdcast")

}
func (m *MockMqttClient) Disconnect() {
	fmt.Println("Mock Disconnect")
}

func (m *MockMqttClient) OnMessageHandler(handler func(id string, payload []byte)) {
	m.messageHandler = handler
}

func (m *MockMqttClient) PublishMessage(id string, payload []byte) {
	m.messagePubHandler()(id, payload)
}
func (m *MockMqttClient) messagePubHandler() func(id string, payload []byte) {
	return func(id string, payload []byte) {
		m.messageHandler(id, payload)
	}
}

func (m *MockMqttClient) Publish(topic string, payload interface{}) {
	fmt.Println("Mock Publish")
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

func (m *NopMqttClient) OnMessageHandler(handler func(string, []byte)) {
	fmt.Println("Empty OnMessageHandler")
}

type NopRepository struct {
}

func (w *NopRepository) Store(deviceName string, payload *devices.Device) {
	fmt.Println("Empty Store")
}

func (w *NopRepository) ListAllDevices() []*devices.Device {

	fmt.Println("Empty ListAllDevices")
	return []*devices.Device{}
}

func (w *NopRepository) FindDevice(deviceName string) (*devices.Device, error) {

	fmt.Println("Empty FindDevice")
	return &devices.Device{}, nil
}
