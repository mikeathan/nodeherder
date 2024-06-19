package utils_test

import (
	"fmt"
	"node-herder/models/devices"
	"time"
)

func CreateExposuresFromMap(data map[string]interface{}) map[string]*devices.Entity {
	var entities = make(map[string]*devices.Entity)
	for key, value := range data {

		newEntity := createEntity(key, "", value, "", nil)
		entities[key] = newEntity
	}
	return entities
}
func CreateEnumEntity(name string, enums map[string]any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Attributes = enums
	newEntity.Presets = map[string]any{}
	newEntity.Data = nil
	newEntity.Name = name
	newEntity.Type = "enums"
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Properties = map[string]any{"min": 0, "max": 255}

	return newEntity
}
func CreatePresetsEntity(name string, presets map[string]any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Attributes = map[string]any{"min": 0.0, "max": 255.0}
	newEntity.Presets = presets
	newEntity.Data = nil
	newEntity.Name = name
	newEntity.Type = "numeric"
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Properties = map[string]any{"min": 0, "max": 255}

	return newEntity
}

func CreateNumericEntity(name string, data any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Attributes = map[string]any{"max": 0.0, "min": 255.0}
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Type = "numeric"
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Properties = map[string]any{"min": 0, "max": 255}

	return newEntity
}

func CreateEntity(name string, propType string, data any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Attributes = map[string]any{"min": 0.0, "max": 255.0}
	newEntity.Presets = make(map[string]any)
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Type = propType
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Properties = map[string]any{"min": 0.0, "max": 255.0}

	return newEntity
}

func CreateDialActionEnums() map[string]any {

	enums := map[string]any{

		"button_1_press":         "button_1_press",
		"button_1_press_release": "button_1_press_release",
		"button_2_press":         "button_2_press",
		"button_2_press_release": "button_2_press_release",
		"dial_rotate_left_fast":  "dial_rotate_left_fast",
		"dial_rotate_left_slow":  "dial_rotate_left_slow",
		"dial_rotate_left_step":  "dial_rotate_left_step",
		"dial_rotate_right_fast": "dial_rotate_right_fast",
		"dial_rotate_right_slow": "dial_rotate_right_slow",
		"dial_rotate_right_step": "dial_rotate_right_step"}

	return enums
}

func CreateColorTempPresets() map[string]any {
	presets := map[string]any{"cool": 250,
		"coolest": 150,
		"neutral": 370,
		"warm":    454,
		"warmest": 500}

	return presets
}

func CreateDeviceWithExposes(deviceId string, friendlyName string, exposes []*devices.Entity) *devices.Device {

	dev := devices.NewDevice(deviceId)
	dev.Id = deviceId
	dev.FriendlyName = friendlyName
	dev.ConnectionType = "mqtt"
	dev.Description = fmt.Sprintf("Test device %s description", deviceId)
	dev.PowerSource = "mains"
	dev.Properties = map[string]any{}
	dev.Properties["last_seen"] = time.Now().Format(time.RFC3339)
	dev.Properties["link_quality"] = 45.0

	dev.Exposes = make(map[string]*devices.Entity)
	for _, e := range exposes {
		dev.Exposes[e.Name] = e
	}

	return dev

}

func createEntity(name string, description string, data any, unit string, props map[string]any) *devices.Entity {
	if props == nil {
		props = make(map[string]any)
	}

	newEntity := &devices.Entity{}
	newEntity.Attributes = map[string]any{}
	newEntity.Presets = map[string]any{}
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Unit = unit
	newEntity.Description = description
	newEntity.Properties = props
	return newEntity
}
