package settings

import (
	"node-herder/models/bridge"
	"node-herder/utils"
)

type DeviceSettings struct {
	Defaults  *DeviceConfig            `json:"defaults"`
	Overrides map[string]*DeviceConfig `json:"overrides"`
}

func (d *DeviceSettings) AddOverride(deviceConfig *DeviceConfig) {
	if deviceConfig.Id == "" {
		return
	}

	if deviceConfig.DebounceOverrides == nil {
		deviceConfig.DebounceOverrides = make(map[string]*utils.TimeInterval)
	}

	d.Overrides[deviceConfig.Id] = deviceConfig
}

func (d *DeviceSettings) DeleteOverride(deviceId string) bool {
	if _, ok := d.Overrides[deviceId]; !ok {
		return false
	}

	delete(d.Overrides, deviceId)
	return true
}

func NewDeviceSettings() *DeviceSettings {
	return &DeviceSettings{
		Defaults:  DefaultDeviceConfig(),
		Overrides: map[string]*DeviceConfig{},
	}
}

func DefaultDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		Disabled:                     false,
		MetricsEnabled:               false,
		CommandConfirmationTimeoutMs: newDefaultConfirmationTimeout(),
		LiveIntervalByCategory:       defaultLiveIntervals(),
		DefaultDebounceByCategory: map[bridge.ExposeCategory]*utils.TimeInterval{
			bridge.DiagnosticCategory:  utils.IntervalFromSeconds(300),
			bridge.MeasurementCategory: utils.IntervalFromSeconds(60),
		},
		DebounceOverrides: nil,
	}
}

type DeviceConfig struct {
	// Live intervals quiet numeric observations; legacy debounce maps sample history.
	LiveIntervalByCategory       map[bridge.ExposeCategory]*utils.TimeInterval `json:"liveIntervalByCategory"`
	LiveIntervalOverrides        map[string]*utils.TimeInterval                `json:"liveIntervalOverrides,omitempty"`
	CommandConfirmationTimeoutMs *int                                          `json:"commandConfirmationTimeoutMs,omitempty"`
	MetricOverrides              map[string]bool                               `json:"metricOverrides,omitempty"`
	ReportingThresholds          map[string]float64                            `json:"reportingThresholds,omitempty"`
	Id                           string                                        `json:"id,omitempty"`
	Disabled                     bool                                          `json:"disabled"`
	MetricsEnabled               bool                                          `json:"metricsEnabled"`
	DefaultDebounceByCategory    map[bridge.ExposeCategory]*utils.TimeInterval `json:"defaultDebounceByCategory,omitempty"`
	DebounceOverrides            map[string]*utils.TimeInterval                `json:"debounceOverrides,omitempty"`
}

func NewDeviceConfig(id string) *DeviceConfig {
	return &DeviceConfig{
		Id:                        id,
		Disabled:                  false,
		MetricsEnabled:            false,
		DebounceOverrides:         map[string]*utils.TimeInterval{},
		DefaultDebounceByCategory: nil,
	}
}

// NewDeviceConfigFrom derives an owned settings snapshot using the same copy path
// as saves and app snapshots, including every optional policy map and interval.
func NewDeviceConfigFrom(config *DeviceConfig) *DeviceConfig {
	return cloneDeviceConfig(config)
}

type HistoryConfig struct {
	SleepTimeout *utils.TimeInterval `json:"sleepTimeout"`
	ExpireAt     *utils.TimeInterval `json:"expireAt"`
}

func NewHistoryConfig(sleepTimeout *utils.TimeInterval, expireAt *utils.TimeInterval) *HistoryConfig {
	return &HistoryConfig{
		SleepTimeout: sleepTimeout,
		ExpireAt:     expireAt,
	}
}

func DefaultHistoryConfig() *HistoryConfig {
	return &HistoryConfig{
		SleepTimeout: utils.IntervalFromHours(12),
		ExpireAt:     utils.IntervalFromDays(10),
	}
}

type LoggerConfig struct {
	EnableRemoteLogger bool   `json:"enableRemoteLogger"`
	Level              string `json:"level"`
}

func NewLoggerConfig(enableRemoteLogger bool, level string) *LoggerConfig {
	return &LoggerConfig{
		EnableRemoteLogger: enableRemoteLogger,
		Level:              level,
	}
}

func DefaultLoggingConfig() *LoggerConfig {
	return &LoggerConfig{
		EnableRemoteLogger: false,
		Level:              "info",
	}
}

type MCPConfig struct {
	Enabled bool `json:"enabled"`
}

func DefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Enabled: false,
	}
}

func DefaultBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		TimeExpireAt: utils.IntervalFromSeconds(120),
		PermitJoin:   false,
	}
}

type AssistantConfig struct {
	Url string `json:"url"`
}

func DefaultAssistantConfig() *AssistantConfig {
	return &AssistantConfig{
		Url: "",
	}
}

type HubConfig struct {
	Devices         *DeviceSettings            `json:"devices"`
	History         *HistoryConfig             `json:"history"`
	Logger          *LoggerConfig              `json:"logger"`
	MCP             *MCPConfig                 `json:"mcp"`
	Assistant       *AssistantConfig           `json:"assistant"`
	DashboardGroups map[string]*DashboardGroup `json:"dashboardGroups"`
}

type BridgeConfig struct {
	TimeExpireAt *utils.TimeInterval `json:"maxTimeAllowed"`
	PermitJoin   bool                `json:"permitJoin"`
}

func NewBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		TimeExpireAt: utils.IntervalFromSeconds(120),
		PermitJoin:   false,
	}
}

func NewHubConfig() *HubConfig {
	return &HubConfig{
		Devices:         NewDeviceSettings(),
		History:         DefaultHistoryConfig(),
		Logger:          DefaultLoggingConfig(),
		MCP:             DefaultMCPConfig(),
		Assistant:       DefaultAssistantConfig(),
		DashboardGroups: map[string]*DashboardGroup{},
	}
}

type AppConfig struct {
	Hub    *HubConfig    `json:"hub"`
	Bridge *BridgeConfig `json:"bridge"`
	// HomeLayout is persisted privately; AppConfigCache strips it from ordinary snapshots.
	HomeLayout *HomeLayout `json:"homeLayout,omitempty"`
}

func (s *AppConfig) AddDeviceConfig(cfg *DeviceConfig) {
	s.Hub.Devices.AddOverride(cfg)
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Hub: &HubConfig{
			Devices:         NewDeviceSettings(),
			History:         DefaultHistoryConfig(),
			Logger:          DefaultLoggingConfig(),
			MCP:             DefaultMCPConfig(),
			Assistant:       DefaultAssistantConfig(),
			DashboardGroups: map[string]*DashboardGroup{},
		},
		Bridge: DefaultBridgeConfig(),
	}
}
