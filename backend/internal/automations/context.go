package automations

import "node-herder/models/devices"

type AutomationContext interface {
	SetPayload(payload map[string]*devices.Entity)
	GetPayload(name string) (*devices.Entity, bool)
	GetCurrent(name string) any
	SetCurrent(name string, value any)
}
