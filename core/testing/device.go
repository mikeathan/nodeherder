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
	newEntity.Values = enums
	newEntity.Category = devices.MeasurementCategory

	newEntity.Data = nil
	newEntity.Name = name
	newEntity.Type = "enum"
	newEntity.Unit = "unit_test"
	newEntity.Category = devices.MeasurementCategory
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Attributes = map[string]any{"min": 0, "max": 255}

	return newEntity
}
func CreatePresetsEntity(name string, presets map[string]any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Category = devices.MeasurementCategory

	newEntity.Values = presets
	newEntity.Data = nil
	newEntity.Name = name
	newEntity.Type = "numeric"
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Attributes = map[string]any{"min": 0, "max": 255}

	return newEntity
}

func CreateNumericEntity(name string, data any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Category = devices.MeasurementCategory

	newEntity.Attributes = map[string]any{"max": 0.0, "min": 255.0}
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Type = "numeric"
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Attributes = map[string]any{"min": 0, "max": 255}

	return newEntity
}

func CreateEntity(name string, propType string, data any) *devices.Entity {

	newEntity := &devices.Entity{}
	newEntity.Category = devices.MeasurementCategory
	newEntity.Attributes = map[string]any{"min": 0.0, "max": 255.0}
	newEntity.Values = make(map[string]any)
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Type = propType
	newEntity.Unit = "unit_test"
	newEntity.Description = fmt.Sprintf("description for expose: %s ", name)
	newEntity.Attributes = map[string]any{"min": 0.0, "max": 255.0}

	return newEntity
}

func CreateAlarmDevice(id string, name string, value bool) *devices.Device {

	device2Expose1 := CreateEntity("alarm", "binary", value)
	return CreateDeviceWithExposes(id, name, []*devices.Entity{device2Expose1})
}

func CreateDoorSensorDevice(id string, name string, value bool) *devices.Device {

	device1Expose1 := CreateEntity("contact", "binary", value)
	return CreateDeviceWithExposes(id, name, []*devices.Entity{device1Expose1})
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

func CreatePresenceDevice(deviceId string, friendlyName string, property string, value bool) *devices.Device {
	dev := devices.NewDevice(deviceId)
	dev.Id = deviceId
	dev.FriendlyName = friendlyName
	dev.ConnectionType = "mqtt"
	dev.Description = fmt.Sprintf("Test device %s description", deviceId)
	dev.PowerSource = "mains"
	dev.LastSeen = time.Now().Format(time.RFC3339)

	expose := CreateEntity(property, "binary", value)
	dev.Exposes = make(map[string]*devices.Entity)
	dev.Exposes[property] = expose

	return dev
}

func CreateLightDevice(deviceId string, friendlyName string, property string, value float64) *devices.Device {
	dev := devices.NewDevice(deviceId)
	dev.Id = deviceId
	dev.FriendlyName = friendlyName
	dev.ConnectionType = "mqtt"
	dev.Description = fmt.Sprintf("Test device %s description", deviceId)
	dev.PowerSource = "mains"
	dev.LastSeen = time.Now().Format(time.RFC3339)

	expose := CreateEntity(property, "numeric", value)
	dev.Exposes = make(map[string]*devices.Entity)
	dev.Exposes[property] = expose

	return dev
}

func CreateDeviceWithExposes(deviceId string, friendlyName string, exposes []*devices.Entity) *devices.Device {

	dev := devices.NewDevice(deviceId)
	dev.Id = deviceId
	dev.FriendlyName = friendlyName
	dev.ConnectionType = "mqtt"
	dev.Description = fmt.Sprintf("Test device %s description", deviceId)
	dev.PowerSource = "mains"
	dev.LastSeen = time.Now().Format(time.RFC3339)

	dev.Exposes = make(map[string]*devices.Entity)
	for _, e := range exposes {
		dev.Exposes[e.Name] = e
	}

	return dev

}

func createEntity(name string, description string, data any, unit string, attributes map[string]any) *devices.Entity {
	if attributes == nil {
		attributes = make(map[string]any)
	}

	newEntity := &devices.Entity{}
	newEntity.Attributes = map[string]any{}
	newEntity.Values = map[string]any{}
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Unit = unit
	newEntity.Description = description
	newEntity.Attributes = attributes
	return newEntity
}
