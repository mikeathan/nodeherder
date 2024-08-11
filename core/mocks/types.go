package mocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	"node-herder/utils/storage"
	"sort"
	"strings"
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

func (w *MockEventHub) OnLoadMetrics(action func(interface{}) (interface{}, error)) {
	fmt.Println("Empty OnLoadMetrics")
}

func (w *MockEventHub) OnLoadAppConfig(action func() (interface{}, error)) {
	fmt.Println("Empty OnLoadAppConfig")

}
func (w *MockEventHub) OnSaveDeviceConfig(func(payload interface{}) error) {
	fmt.Println("Empty OnSaveDeviceConfig")
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

func stripAfterSeparator(input string, separator string) (string, bool) {
	parts := strings.SplitN(input, separator, 2)
	if len(parts) > 1 {
		return parts[0], true
	}
	return input, false
}

func (m *MockMqttClient) Publish(topic string, payload interface{}) {
	fmt.Println("Mock Publish")

	// handle setter mqtt messages. strip set and publish, it then gets handled as device update
	topic = strings.Replace(topic, "/set", "", -1)
	if payload == nil {
		data := []byte("mock payload")
		if p, ok := payload.([]byte); ok {
			data = p
		}
		m.messagePubHandler()(topic, data)
	} else if value, ok := payload.([]byte); ok {
		m.messagePubHandler()(topic, value)
	} else {
		bytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Mock Publish error:", err.Error())
			return
		}
		m.messagePubHandler()(topic, bytes)
	}

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

func (w *NopWsServer) OnLoadMetrics(action func(interface{}) (interface{}, error)) {
	fmt.Println("Empty OnLoadMetrics")
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

func (w *NopWsServer) OnLoadAppConfig(action func() (interface{}, error)) {
	fmt.Println("Empty OnLoadAppConfig")

}
func (w *NopWsServer) OnSaveDeviceConfig(func(payload interface{}) error) {
	fmt.Println("Empty OnSaveDeviceConfig")
}

// Mock devices Repository
type NopRepository struct {
}

func (w *NopRepository) Store(deviceName string, payload *devices.Device) (bool, error) {
	fmt.Println("Empty Store")
	return false, nil
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
	storeHandler func(id string, data map[string]any)
}

func (s *NopMetricsRepo) WithStoreHandler(storeHandler func(id string, data map[string]any)) {
	s.storeHandler = storeHandler
}

func (s *NopMetricsRepo) invokeStoreHandler() func(id string, data map[string]any) {
	return func(id string, data map[string]any) {
		s.storeHandler(id, data)
	}
}

func (s *NopMetricsRepo) Store(id string, data map[string]any) error {
	fmt.Println("Mocked Store")
	s.invokeStoreHandler()(id, data)
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

func (s *NopMetricsRepo) Close() error {
	fmt.Println("Mocked Close")
	return nil
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
	return &settings.AppConfig{
		Devices: map[string]*settings.DeviceConfig{},
	}, nil
}

func (s *NopSettingsrepo) FindOrAddDeviceConfigIfNotExists(id string) (*settings.DeviceConfig, error) {
	fmt.Println("Mocked settingsRepo FindOrAddDeviceConfigIfnotExists")
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
	devices        devices.Repository
	metrics        metrics.Repository
	settings       settings.Repository
	deviceIdMapper *repository.DeviceIdMapper
}

func NewMockAppStore() store.AppStore {
	devicesRepo := NopRepository{}
	metricsRepo := NopMetricsRepo{}
	settingsRepo := NopSettingsrepo{}
	return &NopAppStore{
		devices:        &devicesRepo,
		metrics:        &metricsRepo,
		settings:       &settingsRepo,
		deviceIdMapper: repository.NewDeviceIdMapper(&devicesRepo),
	}
}

func NewMockAppStoreFromDevicesRepo(devicesRepo devices.Repository) store.AppStore {
	metricsRepo := &NopMetricsRepo{}
	settingsRepo := &NopSettingsrepo{}
	return &NopAppStore{
		devices:        devicesRepo,
		metrics:        metricsRepo,
		settings:       settingsRepo,
		deviceIdMapper: repository.NewDeviceIdMapper(devicesRepo),
	}
}

func (s *NopAppStore) FindDeviceConfig(id string) (*settings.DeviceConfig, error) {
	fmt.Println("Mocked store LoaFindDeviceConfigdAppConfig")
	return nil, nil
}

func (s *NopAppStore) LoadAppConfig() (*settings.AppConfig, error) {
	fmt.Println("Mocked store LoadAppConfig")
	return nil, nil
}

func (s *NopAppStore) SaveDeviceConfig(deviceconfig *settings.DeviceConfig) error {
	fmt.Println("Mocked store SaveDeviceConfig")
	return nil
}
func (s *NopAppStore) StoreDevice(friendlyName string, device *devices.Device) error {
	fmt.Println("Mocked store StoreDevice")

	return nil
}

func (s *NopAppStore) StoreMetrics(friendlyName string, data map[string]interface{}) error {
	fmt.Println("Mocked store StoreMetrics")

	return nil
}

func (s *NopAppStore) ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {
	fmt.Println("Mocked store ViewMetrics")

	return &metrics.DeviceMetricsResult{}, nil
}

func (s *NopAppStore) FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error) {
	fmt.Println("Mocked store FindDeviceByFriendlyName")
	return &devices.Device{}, nil
}

func (s *NopAppStore) FindDeviceById(id string) (*devices.Device, error) {
	fmt.Println("Mocked store FindDeviceById")
	return &devices.Device{}, nil
}
func (s *NopAppStore) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	fmt.Println("Mocked store FindDeviceByIds")
	return []*devices.Device{}, nil
}
func (s *NopAppStore) AllDevices() ([]*devices.Device, error) {
	fmt.Println("Mocked store AllDevices")
	return []*devices.Device{}, nil
}

func (s *NopAppStore) StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error {
	fmt.Println("Mocked store StoreBridgeInfoList")
	return nil
}
func (s *NopAppStore) FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error) {
	fmt.Println("Mocked store FindBridgeInfoByFriendlyName")
	return &devices.BridgeInfo{}, nil
}
func (s *NopAppStore) FindBridgeInfoById(id string) (*devices.BridgeInfo, error) {
	fmt.Println("Mocked store FindBridgeInfoById")
	return &devices.BridgeInfo{}, nil
}

func (s *NopAppStore) ResolveFriendlyName(friendlyName string) string {
	fmt.Println("Mocked store ResolveFriendlyName")
	return ""
}

// Mock engine
type MockAutomationEngine[T any] struct {
	cache    map[string]*automations.Device
	mockData []*automations.Device
}

func NewMockAutomationStorage[T automations.Device](mockData []*automations.Device) storage.Storage[automations.Device] {
	d := new(MockAutomationEngine[automations.Device])
	d.cache = make(map[string]*automations.Device)
	d.mockData = mockData
	return d
}

func (d *MockAutomationEngine[T]) Initialize() ([]*automations.Device, error) {

	d.ClearCache()

	// initialize with mock data
	for _, mockItem := range d.mockData {
		d.Store(mockItem.Id, mockItem)
	}

	return d.LoadAll(), nil
}

func (d *MockAutomationEngine[T]) LoadAll() []*automations.Device {

	keys := make([]string, 0, len(d.cache))
	values := make([]*automations.Device, 0, len(d.cache))

	for k, _ := range d.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values = append(values, d.cache[k])
	}

	return values
}

func (d *MockAutomationEngine[T]) Delete(name string) error {

	d.deleteFromCache(name)
	return nil
}

func (d *MockAutomationEngine[T]) ClearCache() {

	for k := range d.cache {
		delete(d.cache, k)
	}
}

func (d *MockAutomationEngine[T]) Store(name string, item *automations.Device) error {

	d.addToCache(name, item)
	return nil
}
func (d *MockAutomationEngine[T]) LoadFromCache(name string) (*automations.Device, error) {
	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}

	return nil, errors.New("not in cache")
}

func (d *MockAutomationEngine[T]) Load(name string) (*automations.Device, error) {

	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}

	for _, mockItem := range d.mockData {
		if mockItem.Id == name {
			return mockItem, nil
		}
	}

	return nil, errors.New("item not found")
}

func (d *MockAutomationEngine[T]) addToCache(name string, item *automations.Device) {
	d.cache[name] = item
}

func (d *MockAutomationEngine[T]) loadFromCache(name string) *automations.Device {
	if item, ok := d.cache[name]; ok {
		return item
	}

	return nil
}

func (d *MockAutomationEngine[T]) deleteFromCache(name string) {
	delete(d.cache, name)
}

// Mock Clock
type MockClock struct {
	callback func() time.Time
}

func NewMockClock(callback func() time.Time) *MockClock {
	return &MockClock{callback: callback}
}

func (m *MockClock) SetMockTime(t time.Time) {
	m.callback = func() time.Time {
		return t
	}
}
func (m *MockClock) Now() time.Time {
	return m.callback()
}
