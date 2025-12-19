package mocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"node-herder/internal/automations"
	metrics "node-herder/internal/metrics/domain"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/logging"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	"node-herder/utils"
	"node-herder/utils/storage"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

// Mock EventHub

// TODO: get rid of this. we only used it to have a differnet mocked implementation of Publish
type MockEventHub struct {
	context hub.Context

	MockBroadcastEvent func(eventName string, data interface{}) error
}

func NewMockEventHub() *MockEventHub {
	return &MockEventHub{
		context: ws.NewRequestContext(),
	}
}
func (w *MockEventHub) SetMockBroadcastEvent(mock func(eventName string, data interface{}) error) {
	w.MockBroadcastEvent = mock
}

func (w *MockEventHub) Context() hub.Context {

	fmt.Println("EventHub: Mocked Context")
	return w.context
}

func (w *MockEventHub) Start() {
	fmt.Println("EventHub: Mocked Start")
}

func (w *MockEventHub) Close() error {
	fmt.Println("EventHub: Mocked Close")
	return nil
}

func (s *MockEventHub) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("EventHub: Mocked HandleRequest")
	return nil
}

func (w *MockEventHub) Broadcast(eventName string, data interface{}) error {
	return w.MockBroadcastEvent(eventName, data)
}

func (w *MockEventHub) EmitDevice(name string) error {
	fmt.Println("Empty EmitDevice")
	return nil
}

func (w *MockEventHub) EmitBridgeConfig() {
	fmt.Println("MockEventHub: EmitBridgeConfig")
}

func (w *MockEventHub) EmitDevices() {
	fmt.Println("MockEventHub: EmitDevices")
}

func (w *MockEventHub) EmitDeviceList(names []string) {
	fmt.Println("EMockEventHub:mpty EmitDeviceList")
}

func (w *MockEventHub) RegisterNewClient(conn *websocket.Conn) {
	fmt.Println("MockEventHub: RegisterNewClient")
}

func (w *MockEventHub) OnLoadHubState(action func() (interface{}, error)) {
	fmt.Println("MockEventHub: OnLoadHubState")
}

func (w *MockEventHub) OnLoadAutomations(onLoadAutomations func() interface{}) {
	fmt.Println("MockEventHub: OnLoadAutomations")
}

func (w *MockEventHub) OnLoadDevice(action func(id string) (interface{}, error)) {
	fmt.Println("MockEventHub: OnLoadDevice")
}
func (w *MockEventHub) OnLoadDevices(action func() interface{}) {
	fmt.Println("MockEventHub: OnLoadDevices")
}

func (w *MockEventHub) OnLoadDeviceList(action func(names []string) interface{}) {
	fmt.Println("MockEventHub: OnLoadDeviceList")
}

func (w *MockEventHub) OnDeviceSetValue(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnDeviceSetValue")
}

func (w *MockEventHub) OnDeviceRename(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnDeviceRename")
}

func (w *MockEventHub) OnDeviceRemove(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnDeviceRemove")
}

func (w *MockEventHub) OnDeviceInterview(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnDeviceInterview")
}

func (w *MockEventHub) OnBridgePermitJoin(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnBridgePermitJoin")
}

func (w *MockEventHub) OnSaveAutomation(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnSaveAutomation")
}

func (w *MockEventHub) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	fmt.Println("MockEventHub: OnDeleteAutomation")
}

func (w *MockEventHub) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	fmt.Println("MockEventHub: OnDeleteAutomationTrigger")
}

func (w *MockEventHub) OnLoadMetrics(action func(interface{}) (interface{}, error)) {
	fmt.Println("MockEventHub: OnLoadMetrics")
}

func (w *MockEventHub) OnLoadAppConfig(action func() (interface{}, error)) {
	fmt.Println("MockEventHub: OnLoadAppConfig")
}

func (w *MockEventHub) OnLoadBridgeConfig(action func() (interface{}, error)) {
	fmt.Println("MockEventHub: OnLoadBridgeConfig")
}

func (w *MockEventHub) OnSaveDeviceConfigOverride(func(payload interface{}) error) {
	fmt.Println("MockEventHub: OnSaveDeviceConfigOverride")
}
func (w *MockEventHub) OnSaveDeviceConfigDefaults(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnSaveDeviceConfigDefaults")
}

func (w *MockEventHub) OnDeleteDeviceConfigOverride(action func(p interface{}) error) {
	fmt.Println("MockEventHub: OnDeleteDeviceConfigOverride")
}

func (w *MockEventHub) OnSaveHistoryConfig(func(payload interface{}) error) {
	fmt.Println("MockEventHub: OnSaveHistoryConfig")
}

func (w MockEventHub) OnSaveLoggerConfig(action func(payload interface{}) error) {
	fmt.Println("MockEventHub: OnSaveLoggerConfig")
}

func (h *MockEventHub) OnSaveDashboardGroup(action func(payload interface{}) error) {
	fmt.Println("MockEventHub OnSaveDashboardGroup")
}

func (h *MockEventHub) OnRenameDashboardGroup(action func(payload interface{}) (interface{}, error)) {
	fmt.Println("MockEventHub OnRenameDashboardGroup")
}

func (h *MockEventHub) OnDeleteDashboardGroup(action func(payload interface{}) error) {
	fmt.Println("MockEventHub OnDeleteDashboardGroup")
}

func (h *MockEventHub) OnImportDashboardGroups(action func(payload interface{}) error) {
	fmt.Println("MockEventHub OnImportDashboardGroups")
}

func (h *MockEventHub) OnLoadDashboardGroups(action func() (interface{}, error)) {
	fmt.Println("MockEventHub OnLoadDashboardGroups")
}

// Mock MqttClient
type MockMqttClient struct {
	messageHandler func(string, []byte)
	responses      map[string]interface{}
	responseDelay  time.Duration
}

func (m *MockMqttClient) SetResponseDelay(delay time.Duration) {
	m.responseDelay = delay
}

func (m *MockMqttClient) Connect() error {
	fmt.Println("Mock Connect")
	return nil
}

func (m *MockMqttClient) AddResponse(topic string, payload interface{}) {
	if m.responses == nil {
		m.responses = make(map[string]interface{})
	}
	m.responses[topic] = payload
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

	// Check if this is a command message (ends with /set)
	isSetCommand := strings.HasSuffix(topic, "/set")

	var data []byte
	switch v := payload.(type) {
	case nil:
		data = []byte("mock payload")
	case []byte:
		data = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println("Mock Publish error:", err.Error())
			return
		}
		data = b
	}

	// For non-set messages, process them normally (like bridge/device messages, sensor updates)
	if !isSetCommand {
		m.messagePubHandler()(topic, data)
		return
	}

	// For /set commands: strip /set to get the device topic
	topic = strings.Replace(topic, "/set", "", -1)

	// Check if there's an auto-response configured for this topic
	hasAutoResponse := false
	if _, ok := m.responses[topic]; ok {
		hasAutoResponse = true
	}

	// Only simulate immediate device response if there's no auto-response configured
	// This prevents double responses that can cause feedback loops
	if !hasAutoResponse {
		go func() {
			time.Sleep(50 * time.Millisecond) // Delay to let pending state be set
			m.messagePubHandler()(topic, data)
		}()
	}

	// Send configured auto-response if available (for specific test scenarios)
	if resp, ok := m.responses[topic]; ok {
		go func() {
			delay := m.responseDelay
			if delay == 0 {
				delay = 500 * time.Millisecond // Default to 500ms
			}

			var respBytes []byte
			switch v := resp.(type) {
			case []byte:
				respBytes = v
			default:
				b, err := json.Marshal(v)
				if err != nil {
					fmt.Println("Mock Auto-response error:", err)
					return
				}
				respBytes = b
			}

			m.messagePubHandler()(topic, respBytes)
		}()
	}
}

// Mock WsServer
type NopWsServer struct {
	context hub.Context
}

func NewNoWsServer() *NopWsServer {
	return &NopWsServer{
		context: ws.NewRequestContext(),
	}
}
func (w *NopWsServer) Context() hub.Context {
	fmt.Println("WsServer: Mocked Context")
	return w.context
}

func (w *NopWsServer) Start() {
	fmt.Println("WsServer: Mocked Start")
}

func (w *NopWsServer) Close() error {
	fmt.Println("WsServer: Mocked Close")
	return nil
}

func (s *NopWsServer) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("WsServer: Mocked HandleRequest")
	return nil
}

func (w *NopWsServer) Broadcast(eventName string, data interface{}) error {
	fmt.Println("WsServer: Mocked Broadcast")
	return nil
}
func (w *NopWsServer) EmitDevice(name string) error {
	fmt.Println("WsServer: Mocked EmitDevice")
	return nil
}
func (w *NopWsServer) EmitDevices() {
	fmt.Println("WsServer: Mocked EmitDevices")
}

func (w *NopWsServer) EmitBridgeConfig() {
	fmt.Println("WsServer: Mocked EmitBridgeConfig")
}

func (w *NopWsServer) EmitDeviceList(names []string) {
	fmt.Println("WsServer: Mocked EmitDeviceList")
}

func (w *NopWsServer) RegisterNewClient(conn *websocket.Conn) {
	fmt.Println("WsServer: Mocked RegisterNewClient")
}

func (w *NopWsServer) OnLoadAutomations(action func() interface{}) {
	fmt.Println("WsServer: Mocked OnLoadAutomations")
}

func (w *NopWsServer) OnLoadHubState(event func() (interface{}, error)) {
	fmt.Println("WsServer: Mocked OnLoadHubState")
}

func (w *NopWsServer) OnDeviceSetValue(action func(payload interface{}) error) {
	fmt.Println("WsServer: Mocked OnDeviceSetValue")
}

func (w *NopWsServer) OnLoadMetrics(action func(interface{}) (interface{}, error)) {
	fmt.Println("WsServer: Mocked OnLoadMetrics")
}

func (w *NopWsServer) OnDeviceRename(event func(p interface{}) error) {
	fmt.Println("WsServer: Mocked OnDeviceRename")
}

func (w *NopWsServer) OnDeviceRemove(event func(p interface{}) error) {
	fmt.Println("WsServer: Mocked OnDeviceRemove")
}

func (w *NopWsServer) OnDeviceInterview(event func(p interface{}) error) {
	fmt.Println("WsServer: Mocked OnDeviceInterview")
}

func (w *NopWsServer) OnBridgePermitJoin(event func(p interface{}) error) {
	fmt.Println("WsServer: Mocked OnBridgePermitJoin")
}

func (w *NopWsServer) OnLoadDevice(action func(id string) (interface{}, error)) {
	fmt.Println("WsServer: Mocked OnLoadDevice")
}

func (w *NopWsServer) OnLoadDeviceList(action func(names []string) interface{}) {
	fmt.Println("WsServer: Mocked OnLoadDeviceList")
}

func (w *NopWsServer) OnLoadDevices(action func() interface{}) {
	fmt.Println("WsServer: Mocked OnLoadDevices")
}

func (w *NopWsServer) OnSaveAutomation(action func(p interface{}) error) {
	fmt.Println("WsServer: Mocked OnSaveAutomation")
}

func (w *NopWsServer) OnDeleteAutomation(action func(p interface{}) (interface{}, error)) {
	fmt.Println("WsServer: Mocked OnDeleteAutomation")
}

func (w *NopWsServer) OnDeleteAutomationTrigger(action func(p interface{}) (interface{}, error)) {
	fmt.Println("Empty OnDeleteAutomationTrigger")
}

func (w *NopWsServer) OnLoadAppConfig(action func() (interface{}, error)) {
	fmt.Println("WsServer: Mocked OnLoadAppConfig")
}

func (w *NopWsServer) OnLoadBridgeConfig(action func() (interface{}, error)) {
	fmt.Println("WsServer: Mocked OnLoadBridgeConfig")
}

func (w *NopWsServer) OnSaveDeviceConfigOverride(func(payload interface{}) error) {
	fmt.Println("WsServer: Mocked OnSaveDeviceConfigOverride")
}

func (w *NopWsServer) OnSaveDeviceConfigDefaults(action func(payload interface{}) error) {
	fmt.Println("WsServer: Mocked OnSaveDeviceConfigDefaults")
}
func (w *NopWsServer) OnSaveHistoryConfig(func(payload interface{}) error) {
	fmt.Println("WsServer: Mocked OnSaveHistoryConfig")
}
func (w *NopWsServer) OnDeleteDeviceConfigOverride(action func(payload interface{}) error) {
	fmt.Println("WsServer: Mocked OnDeleteDeviceConfigOverride")
}
func (w NopWsServer) OnSaveLoggerConfig(action func(payload interface{}) error) {
	fmt.Println("WsServer: Mocked OnSaveLoggerConfig")
}

func (h *NopWsServer) OnSaveDashboardGroup(action func(payload interface{}) error) {
	fmt.Println("WsServer OnSaveDashboardGroup")
}

func (h *NopWsServer) OnRenameDashboardGroup(action func(payload interface{}) (interface{}, error)) {
	fmt.Println("WsServer OnRenameDashboardGroup")
}

func (h *NopWsServer) OnDeleteDashboardGroup(action func(payload interface{}) error) {
	fmt.Println("WsServer OnDeleteDashboardGroup")
}

func (h *NopWsServer) OnImportDashboardGroups(action func(payload interface{}) error) {
	fmt.Println("WsServer OnImportDashboardGroups")
}

func (h *NopWsServer) OnLoadDashboardGroups(action func() (interface{}, error)) {
	fmt.Println("WsServer OnLoadDashboardGroups")
}

// Mock devices Repository
type NopRepository struct {
}

func (w *NopRepository) Store(deviceName string, payload *devices.Device) (bool, error) {
	fmt.Println("Empty Store")
	return false, nil
}

func (w *NopRepository) Remove(deviceName string) error {
	fmt.Println("Empty Remove")
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

func (s *NopMetricsRepo) Prune(expireAt time.Duration) error {
	fmt.Println("Mocked Prune")
	return nil
}

func (s *NopMetricsRepo) Store(id string, data map[string]any) error {
	fmt.Println("Mocked Store")
	s.invokeStoreHandler()(id, data)
	return nil
}

func (s *NopMetricsRepo) QueryDevice(deviceID string, from, to time.Time, filters []metrics.MetricFilter, collectors map[string]metrics.ExposeResult) (*metrics.DeviceMetricsResult, error) {
	fmt.Println("Mocked QueryDevice")
	return nil, nil
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
	return settings.NewAppConfig(), nil
}

func (s *NopSettingsrepo) LoadOrDefaultDeviceConfig(id string) (*settings.DeviceConfig, error) {
	fmt.Println("Mocked settingsRepo LoadOrDefaultDeviceConfig")
	return settings.NewDeviceConfig(id), nil
}
func (s *NopSettingsrepo) SaveDeviceConfig(deviceConfig *settings.DeviceConfig) error {
	fmt.Println("Mocked settingsRepo SaveDeviceConfig")
	return nil
}

func (s *NopSettingsrepo) DeleteDeviceConfig(id string) error {
	fmt.Println("Mocked settingsRepo DeleteDeviceConfig")
	return nil
}

func (s *NopSettingsrepo) Close() error {
	fmt.Println("Mocked settingsRepo Close")
	return nil
}

func (s *NopSettingsrepo) LoadBridgeConfig() (*settings.BridgeConfig, error) {

	fmt.Println("Mocked settingsRepo LoadBridgeConfig")
	return nil, nil
}
func (s *NopSettingsrepo) SaveBridgeConfig(bridgeConfig *settings.BridgeConfig) error {

	fmt.Println("Mocked settingsRepo SaveBridgeConfig")
	return nil
}
func (s *NopSettingsrepo) SaveHubConfig(hubConfig *settings.HubConfig) error {

	fmt.Println("Mocked settingsRepo SaveHubConfig")
	return nil
}

func (s *NopSettingsrepo) SaveAppConfig(appConfig *settings.AppConfig) error {
	fmt.Println("Mocked settingsRepo SaveAppConfig")
	return nil
}

// Mock appstore
type NopAppStore struct {
	devices             devices.Repository
	metrics             metrics.Repository
	config              *settings.AppConfigCache
	deviceIdMapper      *repository.DeviceIdMapper
	loadHubStateFunc    func() (*hub.HubState, error)
	registerIsDirtyFunc func()
	TriggerDirty        func()
}

func NewMockAppStoreWithLoadStateFunc(loadHubStateFunc func() (*hub.HubState, error)) *NopAppStore {

	devicesRepo := NopRepository{}
	metricsRepo := NopMetricsRepo{}
	config := &settings.AppConfigCache{}
	store := &NopAppStore{
		devices:          &devicesRepo,
		metrics:          &metricsRepo,
		config:           config,
		deviceIdMapper:   repository.NewDeviceIdMapper(&devicesRepo),
		loadHubStateFunc: loadHubStateFunc,
	}

	store.TriggerDirty = func() {
		if store.registerIsDirtyFunc != nil {
			store.registerIsDirtyFunc()
		}
	}

	return store
}

func NewMockAppStore() store.AppStore {
	devicesRepo := NopRepository{}
	metricsRepo := NopMetricsRepo{}
	config := &settings.AppConfigCache{}

	return &NopAppStore{
		devices:        &devicesRepo,
		metrics:        &metricsRepo,
		config:         config,
		deviceIdMapper: repository.NewDeviceIdMapper(&devicesRepo),
	}
}

func (s *NopAppStore) WithLoadHubStateFunc(loadHubStateFunc func() (*hub.HubState, error)) {
	s.loadHubStateFunc = loadHubStateFunc
}

func NewMockAppStoreFromDevicesRepo(devicesRepo devices.Repository) store.AppStore {
	metricsRepo := &NopMetricsRepo{}
	config := &settings.AppConfigCache{}
	return &NopAppStore{
		devices:        devicesRepo,
		metrics:        metricsRepo,
		config:         config,
		deviceIdMapper: repository.NewDeviceIdMapper(devicesRepo),
	}
}

func (s *NopAppStore) RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback) {
	fmt.Println("Mocked store RegisterIsDirtyCallback")
	s.registerIsDirtyFunc = cb

}

func (s *NopAppStore) LoadHubState() (*hub.HubState, error) {

	fmt.Println("Mocked store LoadHubState")
	if s.loadHubStateFunc != nil {
		return s.loadHubStateFunc()
	}

	return nil, nil
}
func (s *NopAppStore) SaveHistoryConfig(historyConfig *settings.HistoryConfig) error {
	fmt.Println("Mocked store SaveHistoryConfig")
	return nil
}

func (s *NopAppStore) LoadDeviceConfig(id string) (*settings.DeviceConfig, error) {
	fmt.Println("Mocked store LoaLoadDeviceConfigdAppConfig")
	return nil, nil
}

func (s *NopAppStore) SaveLoggerConfig(loggerConfig *settings.LoggerConfig) error {
	fmt.Println("Mocked store SaveLoggerConfig")
	return nil
}

func (s *NopAppStore) AppConfig() *settings.AppConfigCache {
	fmt.Println("Mocked store AppConfig")
	return nil
}

func (s *NopAppStore) SaveDeviceConfig(deviceconfig *settings.DeviceConfig) error {
	fmt.Println("Mocked store SaveDeviceConfig")
	return nil
}

func (s *NopAppStore) StoreDevice(friendlyName string, device *devices.Device) error {
	fmt.Println("Mocked store StoreDevice")

	return nil
}

func (s *NopAppStore) RemoveDeviceById(id string) error {
	fmt.Println("Mocked store RemoveDeviceById")
	return nil
}

func (s *NopAppStore) StoreMetrics(friendlyName string, data map[string]interface{}) error {
	fmt.Println("Mocked store StoreMetrics")

	return nil
}

func (s *NopAppStore) SaveBridgePermitJoin(enabled bool) error {
	fmt.Println("Mocked store SaveBridgeConfig")
	return nil
}

func (s *NopAppStore) LoadBridgeConfig() (*settings.BridgeConfig, error) {
	fmt.Println("Mocked store LoadBridgeConfig")
	return nil, nil
}

func (s *NopAppStore) ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {
	fmt.Println("Mocked store ViewMetrics")

	return &metrics.DeviceMetricsResult{}, nil
}

func (s *NopAppStore) QueryDevice(deviceID string, from, to time.Time, filters []metrics.MetricFilter, collectors map[string]metrics.ExposeResult) (*metrics.DeviceMetricsResult, error) {
	fmt.Println("Mocked store QueryDevice")
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

func (s *NopAppStore) FindDevices(ids []string) ([]*devices.Device, error) {
	fmt.Println("Mocked store FindDevices")
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
type MockAutomationEngine[T automations.Automation] struct {
	cache    map[string]automations.Automation
	mockData []automations.Automation
}

func NewMockAutomationStorage[T automations.Automation](mockData []automations.Automation) storage.Storage[automations.Automation] {
	d := new(MockAutomationEngine[T])
	d.cache = make(map[string]automations.Automation)
	d.mockData = mockData
	return d
}

func (d *MockAutomationEngine[T]) Initialize() ([]automations.Automation, error) {

	d.ClearCache()

	// initialize with mock data
	for _, mockItem := range d.mockData {
		d.Store(mockItem.GetId(), mockItem)
	}

	return d.LoadAll(), nil
}

func (d *MockAutomationEngine[T]) LoadAll() []automations.Automation {

	keys := make([]string, 0, len(d.cache))
	values := make([]automations.Automation, 0, len(d.cache))

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

func (d *MockAutomationEngine[T]) Store(name string, item automations.Automation) error {

	d.addToCache(name, item)
	return nil
}
func (d *MockAutomationEngine[T]) LoadFromCache(name string) (automations.Automation, error) {
	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}

	return nil, errors.New("not in cache")
}

func (d *MockAutomationEngine[T]) Load(name string) (automations.Automation, error) {

	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}

	for _, mockItem := range d.mockData {
		if mockItem.GetId() == name {
			return mockItem, nil
		}
	}

	return nil, errors.New("item not found")
}

func (d *MockAutomationEngine[T]) addToCache(name string, item automations.Automation) {
	d.cache[name] = item
}

func (d *MockAutomationEngine[T]) loadFromCache(name string) automations.Automation {
	if item, ok := d.cache[name]; ok {
		return item
	}

	return nil
}

func (d *MockAutomationEngine[T]) deleteFromCache(name string) {
	delete(d.cache, name)
}

// Mock  Automation Device Querier

type MockAutomationDeviceQuerier struct {
	mockReturnValues map[string]bool
}

// NewMockAutomationDeviceQuerier creates a new mock instance with predefined return values
func NewMockAutomationDeviceQuerierWithValues(mockReturnValues map[string]bool) *MockAutomationDeviceQuerier {
	return &MockAutomationDeviceQuerier{
		mockReturnValues: mockReturnValues,
	}
}
func NewMockAutomationDeviceQuerier() *MockAutomationDeviceQuerier {
	return &MockAutomationDeviceQuerier{
		mockReturnValues: map[string]bool{},
	}
}

func (s *MockAutomationDeviceQuerier) IsAutomationEnabled(id string) bool {

	fmt.Println("Mocked AutomationDeviceQuerier IsAutomationEnabled")

	if value, ok := s.mockReturnValues[id]; ok {
		return value
	}
	return false
}

// Mock Clock
type MockClock struct {
	callback      func() time.Time
	sleepDuration time.Duration
	timers        []*mockTimer
}

func NewMockClock(callback func() time.Time) *MockClock {
	return &MockClock{callback: callback}
}

func (m *MockClock) Advance(duration time.Duration) {
	oldNow := m.Now()
	newNow := oldNow.Add(duration)

	// update Now()
	m.callback = func() time.Time { return newNow }

	// tick timers
	var remaining []*mockTimer
	for _, t := range m.timers {
		if !t.fired {
			if t.duration <= duration {
				t.fired = true
				t.f() // run inline for determinism
			} else {
				t.duration -= duration
				remaining = append(remaining, t)
			}
		}
	}
	m.timers = remaining
}

func (m *MockClock) SetMockTime(t time.Time) {
	m.callback = func() time.Time {
		return t
	}
}

func (m *MockClock) SetMockSleepDuration(d time.Duration) {
	m.sleepDuration = d
}

// needs refactoring - shouldnt replicate the logic
func (m *MockClock) IsInRange(from time.Time, to time.Time) bool {
	now := m.Now()
	fromTime := time.Date(now.Year(), now.Month(), now.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
	toTime := time.Date(now.Year(), now.Month(), now.Day(), to.Hour(), to.Minute(), to.Second(), to.Nanosecond(), to.Location())

	return utils.CompareTimeRange(now, fromTime, utils.GreaterThanEqual) && utils.CompareTimeRange(now, toTime, utils.LessThanEqual)
}

// needs refactoring - shouldnt replicate the logic
func (r *MockClock) CompareWithNow(t time.Time, operator utils.EqualityOperator) bool {
	now := r.Now()
	toTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	return utils.CompareTimeRange(now, toTime, operator)

}
func (m *MockClock) Now() time.Time {
	return m.callback()
}

func (m *MockClock) Sleep(d time.Duration) {
	//time.Sleep(m.sleepDuration) // ignore the passed in duration and use the mock duration
	m.Advance(d)
}

func (m *MockClock) AfterFunc(d time.Duration, f func()) utils.Timer {
	mt := &mockTimer{duration: d, f: f, parent: m}
	m.timers = append(m.timers, mt)
	return mt
}

type mockTimer struct {
	duration time.Duration
	f        func()
	fired    bool
	parent   *MockClock
}

func (t *mockTimer) Stop() bool {
	if t.fired {
		return false
	}
	t.fired = true
	return true
}

func (t *mockTimer) Reset(d time.Duration) bool {
	t.duration = d
	t.fired = false
	// put it back into parent’s active timers
	t.parent.timers = append(t.parent.timers, t)
	return true
}

// Mock RemoteLogger emitter
type MockRemoteLoggerEmitter struct {
	callback func(eventName string, data interface{}) error
}

func NewMockRemoteLoggerEmitter(callback func(eventName string, data interface{}) error) logging.RemoteHookEmitter {
	return &MockRemoteLoggerEmitter{
		callback: callback,
	}
}

func (e *MockRemoteLoggerEmitter) Broadcast(eventName string, data interface{}) error {
	return e.callback(eventName, data)
}

// File Walker

type MockWalker struct {
	mockFiles []string
}

func NewMockWalker(mockFiles []string) *MockWalker {
	return &MockWalker{mockFiles: mockFiles}
}

func (m *MockWalker) Walk(root string, walkFn filepath.WalkFunc) error {

	for _, file := range m.mockFiles {
		fileInfo := mockFileInfo{name: file, isDir: false}
		if err := walkFn(file, fileInfo, nil); err != nil {
			return err
		}
	}
	return nil
}

type mockFileInfo struct {
	name  string
	isDir bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

// File loader

type MockFileLoader struct {
	fileBuffer []byte
	callback   func() []byte
}

func NewMockFileLoaderWithCallback(callback func() []byte) *MockFileLoader {
	return &MockFileLoader{callback: callback}
}

func NewMockFileLoader(fileBuffer []byte) *MockFileLoader {
	return &MockFileLoader{fileBuffer: fileBuffer}
}

func (f MockFileLoader) Load(file string) ([]byte, error) {
	if f.callback != nil {
		return f.callback(), nil
	}
	return f.fileBuffer, nil
}

// Automation trigger handler
type MockAutomationTrigger struct {
	callback func(automationId string, triggerName string) error
}

func NewMockAutomationTrigger(callback func(automationId string, triggerName string) error) *MockAutomationTrigger {
	return &MockAutomationTrigger{callback: callback}
}

func (m *MockAutomationTrigger) TriggerManual(automationId string, triggerName string) error {
	return m.callback(automationId, triggerName)
}
