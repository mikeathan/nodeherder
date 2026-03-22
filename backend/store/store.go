package store

import (
	"fmt"
	metrics "node-herder/internal/metrics/domain"
	"node-herder/models/assistant"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/utils"
	"sort"
	"time"
)

type AppStoreDirtyFlagCallback func()

type AppStore interface {
	StoreDevice(friendlyName string, device *devices.Device) error
	RemoveDeviceById(id string) error

	FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error)
	FindDeviceById(id string) (*devices.Device, error)
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
	AllDevices() ([]*devices.Device, error)

	LoadHubState() (*hub.HubState, error)
	AppConfig() *settings.AppConfigCache

	StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error
	FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error)
	FindBridgeInfoById(id string) (*devices.BridgeInfo, error)

	StoreMetrics(friendlyName string, data map[string]any) error
	ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error)
	QueryDevice(deviceID string, from, to time.Time, filters []metrics.MetricFilter, collectors map[string]metrics.ExposeResult) (*metrics.DeviceMetricsResult, error)
	ResolveFriendlyName(friendlyName string) string
	RegisterIsDirtyCallback(cb AppStoreDirtyFlagCallback)
	AppendAssistantMessage(conversationID string, role assistant.Role, content string) error
	LoadAssistantHistory(conversationID string) (*assistant.Conversation, error)
	ListAssistantConversations() ([]assistant.ConversationSummary, error)
	DeleteAssistantConversation(conversationID string) error
}

type appStore struct {
	metrics          metrics.Repository
	devices          devices.Repository
	config           *settings.AppConfigCache
	assistantRepo    assistant.Repository
	deviceIdMapper   *repository.DeviceIdMapper
	isDirtyCallbacks []AppStoreDirtyFlagCallback
}

func NewAppStore(devices devices.Repository, metrics metrics.Repository, config *settings.AppConfigCache, assistantRepo assistant.Repository) (AppStore, error) {

	app := &appStore{
		metrics:          metrics,
		devices:          devices,
		config:           config,
		assistantRepo:    assistantRepo,
		deviceIdMapper:   repository.NewDeviceIdMapper(devices),
		isDirtyCallbacks: []AppStoreDirtyFlagCallback{},
	}

	return app, nil
}

func (s *appStore) RegisterIsDirtyCallback(cb AppStoreDirtyFlagCallback) {
	s.isDirtyCallbacks = append(s.isDirtyCallbacks, cb)
}

func (s *appStore) markDirty() {
	for _, cb := range s.isDirtyCallbacks {
		cb()
	}
}

func (s *appStore) LoadHubState() (*hub.HubState, error) {
	devs, err := s.AllDevices()
	if err != nil {
		return nil, fmt.Errorf("OnLoadHubState failed during loading devices. Error: %v ", err.Error())
	}
	appConfig, err := s.config.LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("OnLoadHubState failed during loading appconfig. Error: %v ", err.Error())
	}

	return hub.NewHubState(appConfig, devs), nil
}

func (s *appStore) ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {
	utils.LogDebugf("ViewMetrics: deviceId=%s from=%s to=%s exposeCount=%d", device.Id, from.Format(time.RFC3339), to.Format(time.RFC3339), len(device.Exposes))
	result, err := s.metrics.ViewDeviceTimeRange(device, from, to)
	if err != nil {
		utils.LogDebugf("ViewMetrics: error for deviceId=%s: %v", device.Id, err)
		return nil, err
	}
	utils.LogDebugf("ViewMetrics: deviceId=%s returned %d exposes", device.Id, len(result.Exposes))
	return result, nil
}

func (s *appStore) QueryDevice(deviceID string, from, to time.Time, filters []metrics.MetricFilter, collectors map[string]metrics.ExposeResult) (*metrics.DeviceMetricsResult, error) {
	return s.metrics.QueryDevice(deviceID, from, to, filters, collectors)
}

func (s *appStore) AppConfig() *settings.AppConfigCache {
	return s.config
}

func (s *appStore) StoreMetrics(friendlyName string, data map[string]any) error {
	id := s.ResolveFriendlyName(friendlyName)
	utils.LogDebugf("StoreMetrics: friendlyName=%s resolvedId=%s exposeCount=%d", friendlyName, id, len(data))

	config, err := s.config.GetDeviceConfig(id)
	if err != nil {
		utils.LogDebugf("StoreMetrics: config lookup failed for id=%s, skipping", id)
		return nil
	}

	if !config.MetricsEnabled {
		utils.LogDebugf("StoreMetrics: metrics disabled for id=%s, skipping", id)
		return nil
	}

	return s.metrics.Store(id, data)
}

func (s *appStore) RemoveDeviceById(id string) error {
	err := s.devices.Remove(id)
	if err != nil {
		return err
	}

	s.markDirty()
	return nil
}

func (s *appStore) StoreDevice(friendlyName string, device *devices.Device) error {
	id := s.ResolveFriendlyName(friendlyName)

	_, err := s.devices.Store(id, device)
	if err != nil {
		return err
	}

	// if isNew {
	// 	err := s.initialiseDeviceConfig(device)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	s.deviceIdMapper.UpdateId(friendlyName, id)

	s.markDirty()

	return nil
}

func (s *appStore) StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error {
	err := s.devices.StoreBridge(bridgeInfoList)
	if err != nil {
		return err
	}

	s.deviceIdMapper.Configure()
	return nil
}

func (s *appStore) FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error) {

	id := s.ResolveFriendlyName(friendlyName)

	return s.FindDeviceById(id)
}

func (s *appStore) FindDeviceById(id string) (*devices.Device, error) {
	return s.devices.FindDevice(id)
}

func (s *appStore) FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error) {

	id := s.ResolveFriendlyName(friendlyName)
	return s.FindBridgeInfoById(id)
}

func (s *appStore) FindBridgeInfoById(id string) (*devices.BridgeInfo, error) {
	return s.devices.FindBridgeInfo(id)
}

func (s *appStore) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	return s.devices.FindDevices(ids)
}

func (s *appStore) FindDevices(ids []string) ([]*devices.Device, error) {
	return s.devices.FindDevices(ids)
}

func (s *appStore) AllDevices() ([]*devices.Device, error) {

	devs, err := s.devices.AllDevices()

	// maybe that can by done in differntway, with specific query for ordering ?
	if err == nil {
		// sort by friendly name
		sort.Slice(devs, func(i, j int) bool {
			return devs[i].FriendlyName < devs[j].FriendlyName
		})
	}

	return devs, err
}

func (a *appStore) ResolveFriendlyName(friendlyName string) string {
	return a.deviceIdMapper.ResolveFriendlyName(friendlyName)
}
