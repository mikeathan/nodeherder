package settings

type DashboardGroup struct {
	Name        string                  `json:"name"`
	DeviceGroup map[string]*DeviceGroup `json:"deviceGroup"`
}

type DeviceGroup struct {
	DeviceId string   `json:"deviceId"`
	Exposes  []string `json:"exposes"`
}

func NewDeviceGroup(deviceId string) *DeviceGroup {
	return &DeviceGroup{
		DeviceId: deviceId,
		Exposes:  []string{},
	}
}

func (e *DeviceGroup) AddExpose(expose string) {
	e.Exposes = append(e.Exposes, expose)
}

func NewDashboardGroup(name string) *DashboardGroup {
	return &DashboardGroup{
		Name:        name,
		DeviceGroup: map[string]*DeviceGroup{},
	}
}

func (e *DashboardGroup) AddDeviceExpose(deviceId string, expose string) {
	if _, ok := e.DeviceGroup[deviceId]; !ok {
		e.DeviceGroup[deviceId] = NewDeviceGroup(deviceId)
	}
	e.DeviceGroup[deviceId].AddExpose(expose)
}

