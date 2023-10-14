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

func (w *MockEventHub) OnLoadAutomations(onLoadAutomations func() []byte) {
	fmt.Println("Empty OnLoadAutomations")
}
func (w *MockEventHub) OnLoadDevices(action func() []byte) {
	fmt.Println("Empty OnLoadDevices")
}
func (w *MockEventHub) OnLoadBridgeFeatures(action func() []byte) {
	fmt.Println("Empty OnLoadBridgeFeatures")
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

func (m *MockMqttClient) AddTopic(topic string) error {
	fmt.Println("Mock ConfigureTopic")
	return nil
}

func (m *MockMqttClient) Disconnect() {
	fmt.Println("Mock Disconnect")
}

func (m *MockMqttClient) OnMessageHandler(handler func(id string, payload []byte)) {
	m.messageHandler = handler
}

func (m *MockMqttClient) messagePubHandler() func(id string, payload []byte) {
	return func(id string, payload []byte) {
		m.messageHandler(id, payload)
	}
}

func (m *MockMqttClient) Publish(topic string, payload interface{}) {
	fmt.Println("Mock Publish")
	data := []byte("mock payload")
	if p, ok := payload.([]byte); ok {
		data = p
	}
	m.messagePubHandler()(topic, data)
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

func (w *NopWsServer) OnLoadAutomations(action func() []byte) {
	fmt.Println("Empty OnLoadAutomations")
}

func (w *NopWsServer) OnLoadDevices(action func() []byte) {
	fmt.Println("Empty OnLoadDevices")
}

func (w *NopWsServer) OnLoadBridgeFeatures(action func() []byte) {
	fmt.Println("Empty OnLoadBridgeFeatures")
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

func (w *NopRepository) StoreV2(deviceName string, payload *devices.DeviceV2) {
	fmt.Println("Empty Store")
}

func (w *NopRepository) ListAllDevicesV2() []*devices.DeviceV2 {

	fmt.Println("Empty ListAllDevices")
	return []*devices.DeviceV2{}
}

func (w *NopRepository) FindDeviceV2(deviceName string) (*devices.DeviceV2, error) {

	fmt.Println("Empty FindDevice")
	return &devices.DeviceV2{}, nil
}
