package automations

type DeviceQuerier interface {
	IsAutomationEnabled(id string) bool
}
