package events

import "node-herder/models/settings"

type UpdateHandlerFilter 	
type UpdateHandler interface {
	HandleUpdate(name string, newValue any)
}

type DebouncedUpdateHandler struct {
	deviceConfigs []*settings.DeviceConfig
}

func NewDebouncedUpdateHandler(deviceConfigs []*settings.DeviceConfig) *DebouncedUpdateHandler {
	return &DebouncedUpdateHandler{deviceConfigs: deviceConfigs}
}

func rundebouncer() bool { return true }

func (d *DebouncedUpdateHandler) HandleUpdate(name string, newValue any) bool {

	if rundebouncer() {
		return false
	}

	return true
}
