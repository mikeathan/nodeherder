package devices_test

import (
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"testing"
)

func TestEntity_IsEventValue(t *testing.T) {
	tests := []struct {
		name      string
		entity    *devices.Entity
		newValue  any
		wantEvent bool
	}{
		{
			name: "Whitelist property 'action' is always an event",
			entity: &devices.Entity{
				Name: "action",
				Type: bridge.EnumDataType,
			},
			newValue:  "single",
			wantEvent: true,
		},
		{
			name: "Whitelist property 'click' is always an event",
			entity: &devices.Entity{
				Name: "click",
				Type: bridge.EnumDataType,
			},
			newValue:  "double",
			wantEvent: true,
		},
		{
			name: "Schema-defined toggle (string match)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{
					"toggle": "TOGGLE",
				},
			},
			newValue:  "TOGGLE",
			wantEvent: true,
		},
		{
			name: "Schema-defined toggle (case insensitive match)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{
					"toggle": "TOGGLE",
				},
			},
			newValue:  "toggle",
			wantEvent: true,
		},
		{
			name: "Schema-defined toggle (type mismatch string vs float64 fallback match)",
			entity: &devices.Entity{
				Name: "brightness",
				Type: bridge.NumericDataType,
				Values: map[string]any{
					"toggle": 1.0,
				},
			},
			newValue:  "1",
			wantEvent: true,
		},
		{
			name: "Legacy toggle command fallback (no schema definition)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{},
			},
			newValue:  "TOGGLE",
			wantEvent: true,
		},
		{
			name: "Legacy toggle command fallback case insensitive",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{},
			},
			newValue:  "toggle",
			wantEvent: true,
		},
		{
			name: "Standard state update (not an event)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{
					"on":  "ON",
					"off": "OFF",
				},
			},
			newValue:  "ON",
			wantEvent: false,
		},
		{
			name: "Standard numeric update (not an event)",
			entity: &devices.Entity{
				Name: "brightness",
				Type: bridge.NumericDataType,
			},
			newValue:  128.0,
			wantEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entity.IsEventValue(tt.newValue); got != tt.wantEvent {
				t.Errorf("Entity.IsEventValue() = %v, want %v", got, tt.wantEvent)
			}
		})
	}
}
