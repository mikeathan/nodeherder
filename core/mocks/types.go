package mocks

import (
	"errors"
	"fmt"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/store"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

// Mock EventHub
type MockEventHub struct {
	MockBroadcastEvent func(eventName string, data interface{}) error
}

func (w *MockEventHub) Broadcast(eventName string, data interface{}) error {
	return w.MockBroadcastEvent(eventName, data)
}

func (w *MockEventHub) EmitDevice(name string) error {
	fmt.Println("Empty EmitDevice")
	return nil
}

func (w *MockEventHub) EmitDevices() {
	fmt.Println("Empty EmitDevices")
}

func (w *MockEventHub) EmitDeviceList(names []string) {
	fmt.Println("Empty EmitDeviceList")
}

func (w *MockEventHub) RegisterNewClient(conn *websocket.Conn) {
	fmt.Println("Empty RegisterNewClient")
}

func (w *MockEventHub) OnLoadAutomations(onLoadAutomations func() interface{}) {
	fmt.Println("Empty OnLoadAutomations")
}

func (w *MockEventHub) OnLoadDevice(action func(id string) (interface{}, error)) {
	fmt.Println("Empty OnLoadDevice")
}
func (w *MockEventHub) OnLoadDevices(action func() interface{}) {
	fmt.Println("Empty OnLoadDevices")
}

func (w *MockEventHub) OnLoadDeviceList(action func(names []string) interface{}) {
	fmt.Println("Empty OnLoadDeviceList")
}

func (w *MockEventHub) OnDeviceSetValue(P func(payload interface{}) error) {
	fmt.Println("Empty OnDeviceSetValue")
}

func (w *MockEventHub) OnDeviceRename(P func(payload interface{}) error) {
	fmt.Println("Empty OnDeviceRename")
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

// Mock MqttClient
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

func (m *MockMqttClient) RemoveTopic(topic string) error {
	fmt.Println("Mock RemoveTopic")
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

// Mock WsServer
type NopWsServer struct {
}

func (w *NopWsServer) Broadcast(eventName string, data interface{}) error {
	fmt.Println("Empty Broadcast")
	return nil
}
func (w *NopWsServer) EmitDevice(name string) error {
	fmt.Println("Empty EmitDevice")
	return nil
}
func (w *NopWsServer) EmitDevices() {
	fmt.Println("Empty EmitDevices")
}

func (w *NopWsServer) EmitDeviceList(names []string) {
	fmt.Println("Empty EmitDeviceList")
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

func (w *NopWsServer) OnDeviceRename(P func(payload interface{}) error) {
	fmt.Println("Empty OnDeviceRename")
}

func (w *NopWsServer) OnLoadDevice(action func(id string) (interface{}, error)) {
	fmt.Println("Empty OnLoadDevice")
}

func (w *NopWsServer) OnLoadDeviceList(action func(names []string) interface{}) {
	fmt.Println("Empty OnLoadDeviceList")
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

// Mock devices Repository
type NopRepository struct {
}

func (w *NopRepository) Store(deviceName string, payload *devices.Device) error {
	fmt.Println("Empty Store")
	return nil
}

func (w *NopRepository) AllDevices() ([]*devices.Device, error) {

	fmt.Println("Empty ListAllDevices")
	return []*devices.Device{}, nil
}

func (w *NopRepository) FindDevices(ids []string) ([]*devices.Device, error) {
	fmt.Println("Empty FindDevices")
	return []*devices.Device{}, nil
}

func (w *NopRepository) FindDevice(deviceName string) (*devices.Device, error) {

	fmt.Println("Empty FindDevice")
	return nil, errors.New("device not found")
}

func (w *NopRepository) StoreBridge(brigeInfo []*devices.BridgeInfo) error {

	fmt.Println("Empty StoreBridge")
	return nil
}

func (w *NopRepository) AllBridgeInfo() ([]*devices.BridgeInfo, error) {

	fmt.Println("Empty AllBridgeInfo")
	return nil, nil
}

func (w *NopRepository) FindBridgeInfo(key string) (*devices.BridgeInfo, error) {

	fmt.Println("Empty FindBridgeInfo")
	return nil, nil
}

func (w *NopRepository) Close() error {

	fmt.Println("Empty close")
	return nil
}

// Mock DeviceRegistrar
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

func (w *NopDeviceRegistrar) RetrieveEntityData(id string, property string) (any, error) {
	fmt.Println("Mocked RetrieveEntityData")
	return nil, errors.New("mocked object")

}

// Mock metrics repo
type NopMetricsRepo struct {
}

func (s *NopMetricsRepo) Store(device *devices.Device) error {
	fmt.Println("Mocked Store")
	return nil
}
func (s *NopMetricsRepo) ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {
	fmt.Println("Mocked ViewDeviceTimeRange")
	return nil, nil
}

func (s *NopMetricsRepo) ViewExposeTimeRange(device *devices.Device, exposeName string, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {

	fmt.Println("Mocked ViewExposeTimeRange")
	return nil, nil
}

func (s *NopMetricsRepo) Close() {
	fmt.Println("Mocked Close")
}

// Mock seting repo
type NopSettingsrepo struct {
}

func (s *NopSettingsrepo) Save(config *settings.AppConfig) error {
	fmt.Println("Mocked settingsRepo Save")
	return nil
}

func (s *NopSettingsrepo) Load() (*settings.AppConfig, error) {
	fmt.Println("Mocked settingsRepo Load")
	return nil, nil
}
func (s *NopSettingsrepo) FindDeviceConfig(id string) (*settings.DeviceConfig, error) {
	fmt.Println("Mocked settingsRepo FindDeviceConfig")
	return nil, nil
}
func (s *NopSettingsrepo) SaveDeviceConfig(deviceConfig *settings.DeviceConfig) error {
	fmt.Println("Mocked settingsRepo SaveDeviceConfig")
	return nil
}
func (s *NopSettingsrepo) Close() error {
	fmt.Println("Mocked settingsRepo Close")
	return nil
}

// Mock appstore
type NopAppStore struct {
	devices  devices.Repository
	metrics  metrics.Repository
	settings settings.Repository
}

func NewMockAppStore() store.AppStore {
	devicesRepo := NopRepository{}
	metricsRepo := NopMetricsRepo{}
	settingsRepo := NopSettingsrepo{}
	return &NopAppStore{devices: &devicesRepo, metrics: &metricsRepo, settings: &settingsRepo}
}

func (s *NopAppStore) History() metrics.Repository {
	return s.metrics
}

func (s *NopAppStore) Devices() devices.Repository {
	return s.devices
}

func (s *NopAppStore) Config() settings.Repository {
	return s.settings
}
