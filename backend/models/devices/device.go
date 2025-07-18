package devices

import (
	"errors"
	"fmt"
	"node-herder/models/bridge"
	"node-herder/utils"
	"sync"
	"time"
)

var mainsKey = "mains"
var batterKey = "battery"
var lastSeenKey = "last_seen"

type AvailabilityType string

const (
	UnknownAvailability AvailabilityType = "unknown"
	OnlineAvailability  AvailabilityType = "online"
	OfflineAvailability AvailabilityType = "offline"
)

var units = map[string]string{
	"temperature":     "°C",
	"pressure":        "hPa",
	"humidity":        "%",
	"voltage":         "mV",
	"linkquality":     "LQI",
	"illuminance_lux": "lux",
	"illuminance":     "lux",
	"battpercentage":  "%",
	"co":              "ppm",
}
var nonBridgeExposesWhitelist = map[string]int{
	"temperature":         1,
	"humidity":            2,
	"pressure":            3,
	"presence":            4,
	"illuminance_lux":     5,
	"occupancy":           6,
	"brightness":          7,
	"state":               8,
	"color_temp":          9,
	"air_quality_score":   10,
	"pm1":                 11,
	"pm25":                12,
	"pm10":                13,
	"action":              14,
	"action_direction":    15,
	"action_type":         16,
	"action_time":         17,
	"tamper":              18,
	"voc":                 19,
	"smoke":               20,
	"smoke_concentration": 21,
	"test":                22,
	"device_fault":        23,
	"power":               24,
	"voltage":             25,
	"current":             26,
	"energy":              27,
	"alarm":               28,
	"volume":              29,
	"contact":             30,
	"co":                  31,
	"carbon_monoxide":     32,
	"self_test_result":    33,
	"illuminance":         34,
}

var measurementWhitelist = map[string]int{
	"state":            1,
	"brightness":       2,
	"color_temp":       3,
	"mode":             4,
	"sound":            5,
	"occupancy":        6,
	"tamper":           7,
	"alarm":            8,
	"action":           9,
	"action_direction": 10,
	"action_type":      11,
	"action_time":      12,
	"contact":          13,
}

var configWhitelist = map[string]int{
	"color_temp_startup": 1,
	"color_temp_max":     2,
	"color_temp_min":     3,
	"color_options":      4,
	"options":            5,
}
var diagnosticWhitelist = map[string]int{
	"linkquality":     1,
	"battery":         2,
	"voltage":         3,
	"battery_low":     4,
	"battery_state":   5,
	"strength":        6,
	"target_distance": 7,
}

// read access mode is set as measurement category
// unless is in blacklist

// write access mode is set as config category
// unless is in the whitelist

// if diagnostics is set in expose check if is in blacklist so it can be moved to measurmeent eg switch action

// categories
// measurement
// exposes that can be shown in a card as a sensor value

// diagnostics
// exposes	that have diagnostics information (battery, link, voltage, batterypercentage tc...))

// config
// exposes that are used to configure the device

func getExposeCategory(entity BridgeExpose) string {

	if _, ok := configWhitelist[entity.Property]; ok {
		return bridge.ConfigCategory
	}
	if _, ok := diagnosticWhitelist[entity.Property]; ok {
		return bridge.DiagnosticCategory
	}
	if _, ok := measurementWhitelist[entity.Property]; ok {
		return bridge.MeasurementCategory
	}
	if entity.Access&WriteBridgeAccessMode != 0 {
		return bridge.ConfigCategory
	}

	if entity.Category == "" {

		if entity.Access&ReadBridgeAccessMode != 0 || entity.Access&StateBridgeAccessMode != 0 {
			return bridge.MeasurementCategory
		}
	}
	return entity.Category
}

type Device struct {
	Id             string             `json:"id"`
	FriendlyName   string             `json:"friendly_name"`
	Description    string             `json:"description,omitempty"`
	ConnectionType string             `json:"connection_type"`
	PowerSource    string             `json:"power_source"`
	Exposes        map[string]*Entity `json:"exposes"`
	LastSeen       string             `json:"last_seen"`
	Availability   AvailabilityType   `json:"availability"`
	mutex          sync.RWMutex
}

func NewDevice(id string) *Device {

	return &Device{
		Id:             id,
		FriendlyName:   "",
		Description:    "",
		ConnectionType: "",
		PowerSource:    "",
		Availability:   UnknownAvailability,
		LastSeen:       "",
		Exposes:        map[string]*Entity{},
		mutex:          sync.RWMutex{},
	}
}

type PackageData map[string]any
type UpdatePackage struct {
	Id           string           `json:"id"`
	LastSeen     string           `json:"last_seen"`
	Availability AvailabilityType `json:"availability,omitempty"`
	Data         PackageData      `json:"data"`
}

func NewUpdatePackage(id string) *UpdatePackage {
	return &UpdatePackage{Id: id, LastSeen: getCurrentTime(), Data: make(map[string]any)}
}

func (u *UpdatePackage) HasData() bool {
	return len(u.Data) != 0
}

type Entity struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Unit        string                  `json:"unit,omitempty"`
	Data        any                     `json:"data"`
	Type        bridge.ExposeDataType   `json:"type"`
	Category    bridge.ExposeCategory   `json:"category,omitempty"`
	Attributes  map[string]any          `json:"attributes,omitempty"`
	AccessMode  bridge.ExposeAccessMode `json:"access_mode"`
	Values      map[string]any          `json:"values,omitempty"`
}

func newEntity() *Entity {
	return &Entity{
		Attributes: make(map[string]any),
		AccessMode: bridge.UnknownAccessMode,
		Values:     make(map[string]any),
	}
}

func CreateEntityFromExpose(expose BridgeExpose, data any) (*Entity, error) {

	if expose.Property == "" {
		return nil, fmt.Errorf("no expose data")
	}

	// if _, ok := exposesWhitelist[expose.Property]; !ok {
	// 	return nil, fmt.Errorf("expose property %v is blacklisted", expose.Property)
	// }

	accessMode := expose.AccessMode()
	if accessMode == bridge.UnknownAccessMode {
		return nil, fmt.Errorf("invalid device feature access mode %v", expose.Access)
	}

	newEntity := newEntity()
	newEntity.Category = getExposeCategory(expose)
	newEntity.Name = expose.Property
	newEntity.AccessMode = accessMode
	newEntity.Description = expose.Description
	newEntity.Unit = expose.Unit
	newEntity.Data = data
	newEntity.Type = expose.Type

	switch expose.Type {
	case bridge.NumericDataType:
		if expose.ValueMax != nil {
			newEntity.Attributes["max"] = expose.ValueMax
		}
		if expose.ValueMin != nil {
			newEntity.Attributes["min"] = expose.ValueMin
		}

		if len(expose.Presets) != 0 {
			for _, preset := range expose.Presets {
				newEntity.Values[preset.Name] = preset.Value
			}
		}
		// else {
		// 	newEntity.Values[expose.Name] = 0 // ????????? - i dont think i need this
		// }

	case bridge.BinaryDataType:

		newEntity.Values["on"] = expose.ValueOn
		newEntity.Values["off"] = expose.ValueOff

		if expose.ValueToggle != "" {
			newEntity.Values["toggle"] = expose.ValueToggle
		}

	case bridge.EnumDataType:
		for id, item := range expose.Values {
			newEntity.Values[fmt.Sprintf("%d", id)] = item
		}
	}

	return newEntity, nil
}

// Not used yet, is for handling non bridge devices which we havent tested yet
func createExpose(data map[string]interface{}) map[string]*Entity {
	var entities = make(map[string]*Entity)
	for key, value := range data {
		if _, ok := nonBridgeExposesWhitelist[key]; !ok {
			continue
		}

		newEntity := newEntity()
		newEntity.Category = bridge.MeasurementCategory
		newEntity.Name = key
		newEntity.AccessMode = bridge.ReadAccessMode
		newEntity.Data = value
		newEntity.Unit = units[key]
		newEntity.Type = bridge.NumericDataType // TODO: make this dynamic
		entities[key] = newEntity
	}
	return entities
}

func createExposeFromBridge(data map[string]interface{}, bridgeInfo *BridgeInfo) map[string]*Entity {

	var entities = map[string]*Entity{}
	for _, expose := range bridgeInfo.Definition.Exposes {
		// load exposes
		if expose.Property != "" {
			if value, ok := data[expose.Property]; ok {
				entity, err := CreateEntityFromExpose(expose, value)
				if err != nil {
					continue
				}
				entities[expose.Property] = entity
			}
		}

		// load features
		for _, feature := range expose.Features {
			if value, ok := data[feature.Property]; ok {

				entity, err := CreateEntityFromExpose(feature, value)
				if err != nil {
					continue
				}
				entities[feature.Property] = entity
			}
		}
	}

	return entities
}

func CreateNewDevice(id string, friendlyName string, connType string, bridgeInfo *BridgeInfo, data map[string]interface{}) (*Device, error) {

	var newDevice = NewDevice(id)
	newDevice.FriendlyName = friendlyName
	newDevice.ConnectionType = connType

	if _, ok := data[batterKey]; !ok {
		newDevice.PowerSource = mainsKey
	} else {
		newDevice.PowerSource = batterKey
	}

	if bridgeInfo != nil {
		newDevice.Exposes = createExposeFromBridge(data, bridgeInfo)
	} else { // device not in hub bridge
		newDevice.Exposes = createExpose(data)
	}

	if len(newDevice.Exposes) == 0 {
		return nil, errors.New("invalid payload - no exposed entries found")
	}

	newDevice.LastSeen = getLastSeen(data)
	newDevice.Availability = OnlineAvailability
	return newDevice, nil
}

func getLastSeen(data map[string]interface{}) string {
	if val, ok := data[lastSeenKey]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	utils.LogDebug("lastSeen not in payload, using current time.")
	return getCurrentTime()
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func (d *Device) GetFriendlyName() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.FriendlyName
}

func (d *Device) GetExpose(name string) (*Entity, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	expose, ok := d.Exposes[name]
	return expose, ok
}

func (d *Device) SetExposeData(name string, data interface{}) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Exposes[name].Data = data
}

func (d *Device) GetAvailability() AvailabilityType {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.Availability
}

func (d *Device) SetAvailability(availability AvailabilityType) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Availability = availability
}

func (d *Device) SetLastSeen(lastSeen string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.LastSeen = lastSeen
}

func (device *Device) LastSeenTime() (time.Time, error) {

	defer device.mutex.RUnlock()
	device.mutex.RLock()

	lastSeen, err := time.Parse(time.RFC3339, device.LastSeen)
	if err != nil {
		return time.Time{}, err
	}

	return lastSeen, nil
}

func (device *Device) IsAvailable() bool {

	defer device.mutex.RUnlock()
	device.mutex.RLock()

	return device.Availability == OnlineAvailability
}

func (device *Device) SetAvailable(value bool) {

	defer device.mutex.Unlock()
	device.mutex.Lock()
	if value {
		device.Availability = OnlineAvailability
	} else {
		device.Availability = OfflineAvailability
	}
}
