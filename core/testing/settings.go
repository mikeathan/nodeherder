package utils_test

import (
	"node-herder/mocks"
	"node-herder/models/settings"
	"node-herder/utils"
	"time"
)

func CreateDebouncer(id string) *settings.DeviceDebouncer {
	app := settings.NewAppConfig()
	cache := settings.NewDeviceConfigCache(app)
	return settings.NewDeviceDebouncer(id, cache, mocks.NewMockClock(func() time.Time { return time.Now() }))
}

func CreateDebouncerFromAppConfig(id string, appConfig *settings.AppConfig, clock utils.Clock) *settings.DeviceDebouncer {
	cache := settings.NewDeviceConfigCache(appConfig)
	return settings.NewDeviceDebouncer(id, cache, clock)
}
