package utils_test

import (
	"node-herder/mocks"
	"node-herder/models/settings"
	"time"
)

func CreateDebouncer(id string) *settings.DeviceDebouncer {
	app := settings.NewAppConfig()
	cache := settings.NewDeviceConfigCache(app)
	return settings.NewDeviceDebouncer(id, cache, mocks.NewMockClock(func() time.Time { return time.Now() }))
}
