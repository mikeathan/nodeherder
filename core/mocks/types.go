package mocks

import (
	"errors"
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

func (w *MockEventHub) OnLoadAutomations(onLoadAutomations func() interface{}) {
	fmt.Println("Empty OnLoadAutomations")
}

func (w *MockEventHub) OnLoadDevices(action func() interface{}) {
	fmt.Println("Empty OnLoadDevices")
}

func (w *MockEventHub) OnDeviceSetValue(P func(payload interface{}) error) {
	fmt.Println("Empty OnDeviceSetValue")
}

func (w *MockEventHub) OnSaveAutomation(action func(p interface{}) error) {
	fmt.Println("Empty OnSaveAutomation")
}

func (w *MockEventHub) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	fmt.Println("Empty OnDeleteAutomation")
}

func (w *MockEventHub) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	fmt.Println("Empty OnDeleteAutomationTrigger")
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

func (w *NopWsServer) OnLoadAutomations(action func() interface{}) {
	fmt.Println("Empty OnLoadAutomations")
}

func (w *NopWsServer) OnDeviceSetValue(action func(payload interface{}) error) {
	fmt.Println("Empty OnDeviceSetValue")
}

func (w *NopWsServer) OnLoadDevices(action func() interface{}) {
	fmt.Println("Empty OnLoadDevices")
}

func (w *NopWsServer) OnSaveAutomation(action func(p interface{}) error) {
	fmt.Println("Empty OnSaveAutomation")
}

func (w *NopWsServer) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	fmt.Println("Empty OnDeleteAutomation")
}

func (w *NopWsServer) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	fmt.Println("Empty OnDeleteAutomationTrigger")
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

func (w *NopRepository) AllDevices() []*devices.Device {

	fmt.Println("Empty ListAllDevices")
	return []*devices.Device{}
}

func (w *NopRepository) FindDevice(deviceName string) (*devices.Device, error) {

	fmt.Println("Empty FindDevice")
	return nil, errors.New("device not found")
}

type NopDeviceRegistrar struct {
}

func (w *NopDeviceRegistrar) LookupByName(name string) (*devices.Device, error) {
	fmt.Println("Mocked LookupByName")

	return nil, errors.New("mocked object")
}
func (w *NopDeviceRegistrar) LookupById(id string) (*devices.Device, error) {
	fmt.Println("Mocked LookupById")

	return nil, errors.New("mocked object")
}

func (w *NopDeviceRegistrar) CreateNewDevice(friendlyName string, connType string, data map[string]interface{}) (*devices.Device, error) {
	fmt.Println("Mocked CreateNewDevice")
	return nil, errors.New("mocked object")

}
func (w *NopDeviceRegistrar) FindBridgeInfo(id string) *devices.BridgeInfo {
	fmt.Println("Mocked FindBridgeInfo")
	return nil
}

func (w *NopDeviceRegistrar) RegisterBridge(bridgeInfoList []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int) {
	fmt.Println("Mocked RegisterBridge")
}
