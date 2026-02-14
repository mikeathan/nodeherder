package resources

import (
	"node-herder/models/hub"
	"node-herder/store"
)

// DeviceStore provides access to devices for resource generation.
type DeviceStore interface {
	LoadHubState() (*hub.HubState, error)
	RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback)
}
