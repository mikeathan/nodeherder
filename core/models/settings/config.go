package settings

import (
	"node-herder/models/bridge"
	"node-herder/utils"
	"time"
)

type DeviceBaseConfig struct {
	Disabled         bool                                          `json:"disabled"`
	MetricsEnabled   bool                                          `json:"history"`
	RateLimit        *utils.TimeInterval                           `json:"rateLimit"`
	CategoryDebounce map[bridge.ExposeCategory]*utils.TimeInterval `json:"categoryDebounce"`
}

//eg
//DefaultDebounce map[ExposeCategory]*utils.TimeInterval `json:"defaultDebounce"`
//DefaultDebounce[ExposeCategory.DiagnosticCategory] = utils.IntervalFromSeconds(300)

// certain types can have default debouncer eg diagnostic
// TODO: remove id from DeviceConfig - we dont need it

// default values:
// 		disabled": false,
// 		"history": false,
// 		RateLimit default to 60 seconds

//// for diagnostic entities, set default debounce to 5 min
// else empty
// deviceConfig.Debounce[entity.Name] = utils.IntervalFromSeconds(300)

func DefaultDeviceBaseConfig() *DeviceBaseConfig {
	return &DeviceBaseConfig{
		Disabled:         false,
		MetricsEnabled:   false,
		RateLimit:        utils.IntervalFromSeconds(60), // default to 60 seconds
		CategoryDebounce: map[bridge.ExposeCategory]*utils.TimeInterval{bridge.DiagnosticCategory: utils.IntervalFromSeconds(300)},
	}
}

type DevicesConfig struct {
	BaseConfig *DeviceBaseConfig `json:"baseConfig"`
	Config     []*DeviceConfig   `json:"config"`
}

func (d *DevicesConfig) Find(id string) *DeviceConfig {
	for _, device := range d.Config {
		if device.Id == id {
			return device
		}
	}
	return nil
}

func (d *DevicesConfig) Save(deviceConfig *DeviceConfig) {

	cfg := d.Find(deviceConfig.Id)
	if cfg != nil {
		// update
		*cfg = *deviceConfig
		return
	}

	// add new
	d.Config = append(d.Config, deviceConfig)
}

func NewDevicesConfig() *DevicesConfig {
	return &DevicesConfig{
		BaseConfig: DefaultDeviceBaseConfig(),
		Config:     []*DeviceConfig{},
	}
}

type DeviceConfig struct {
	Id             string                         `json:"id"`
	Disabled       bool                           `json:"disabled"`
	MetricsEnabled bool                           `json:"history"`
	RateLimit      *utils.TimeInterval            `json:"rateLimit"`
	Debounce       map[string]*utils.TimeInterval `json:"debounce"`
}

func (d *DeviceConfig) RateLimitDuration() time.Duration {
	if d.RateLimit == nil {
		return 0
	}
	return d.RateLimit.Duration()
}

func NewDeviceConfig(id string) *DeviceConfig {
	return &DeviceConfig{
		Id:             id,
		Disabled:       false,
		MetricsEnabled: false,
		RateLimit:      &utils.TimeInterval{}, // default to 60 seconds
		Debounce:       map[string]*utils.TimeInterval{},
	}
}

type HistoryConfig struct {
	SleepTimeout *utils.TimeInterval `json:"sleeptTmeout"`
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
	EnableRemoteLogger bool `json:"enableRemoteLogger"`
}

func NewLoggerConfig(enableRemoteLogger bool) *LoggerConfig {
	return &LoggerConfig{
		EnableRemoteLogger: enableRemoteLogger,
	}
}

func DefaultLoggingConfig() *LoggerConfig {
	return &LoggerConfig{
		EnableRemoteLogger: false,
	}
}

func DefaultBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		TimeExpireAt: utils.IntervalFromSeconds(120),
		PermitJoin:   false,
	}
}

type HubConfig struct {
	//Devices         map[string]*DeviceConfig   `json:"devices"`
	Devices         *DevicesConfig             `json:"devices"`
	History         *HistoryConfig             `json:"history"`
	Logger          *LoggerConfig              `json:"logger"`
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
		Devices:         NewDevicesConfig(),
		History:         DefaultHistoryConfig(),
		Logger:          DefaultLoggingConfig(),
		DashboardGroups: map[string]*DashboardGroup{},
	}
}

type AppConfig struct {
	Hub    *HubConfig    `json:"hub"`
	Bridge *BridgeConfig `json:"bridge"`
}

func (s *AppConfig) AddDeviceConfig(cfg *DeviceConfig) {

	s.Hub.Devices.Config = append(s.Hub.Devices.Config, cfg)
	//s.Hub.Devices[cfg.Id] = cfg
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Hub: &HubConfig{
			Devices:         NewDevicesConfig(),
			History:         DefaultHistoryConfig(),
			Logger:          DefaultLoggingConfig(),
			DashboardGroups: map[string]*DashboardGroup{},
		},
		Bridge: DefaultBridgeConfig(),
	}
}
