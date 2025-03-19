package devices

import (
	"errors"
	"fmt"
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
var exposesWhitelist = map[string]int{
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

type ExposeDataType = string
type ExposeCategory = string
type ExposeAccessMode = string

const (
	ReadAccessMode      ExposeAccessMode = "read"
	WriteAccessMode     ExposeAccessMode = "write"
	ReadWriteAccessMode ExposeAccessMode = "readwrite"
	UnknownAccessMode   ExposeAccessMode = "unknown"

	MeasurementCategory ExposeCategory = "measurement"
	DiagnosticCategory  ExposeCategory = "diagnostic"
	ConfigCategory      ExposeCategory = "config"

	NumericDataType   ExposeDataType = "numeric"
	EnumDataType      ExposeDataType = "enum"
	BinaryDataType    ExposeDataType = "binary"
	CompositeDataType ExposeDataType = "composite"
)

func getExposeAccessMode(entity *BridgeExpose) ExposeAccessMode {

	if HasReadWriteAccessMode(entity) {
		return ReadWriteAccessMode
	}
	if HasWriteAccessMode(entity) {
		return WriteAccessMode
	}
	if HasReadAccessMode(entity) {
		return ReadAccessMode
	}

	return UnknownAccessMode
}

func getExposeCategory(entity BridgeExpose) string {
	if entity.Category == "" {
		return MeasurementCategory
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

type EntityPreset struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Value       int    `json:"value"`
}

type Entity struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Unit        string           `json:"unit,omitempty"`
	Data        any              `json:"data"`
	Type        ExposeDataType   `json:"type"`
	AccessMode  ExposeAccessMode `json:"access_mode"`
	Category    ExposeCategory   `json:"category,omitempty"`
	Attributes  map[string]any   `json:"attributes,omitempty"`
	Presets     map[string]any   `json:"presets,omitempty"`
	Values      map[string]any
}

func newEntity() *Entity {
	return &Entity{Attributes: make(map[string]any), Values: make(map[string]any)}
}

func CreateEntityFromExpose(expose BridgeExpose, data any) (*Entity, error) {

	if expose.Property == "" {
		return nil, fmt.Errorf("no expose data")
	}

	// if _, ok := exposesWhitelist[expose.Property]; !ok {
	// 	return nil, fmt.Errorf("expose property %v is blacklisted", expose.Property)
	// }

	accessMode := getExposeAccessMode(&expose)
	if accessMode == UnknownAccessMode {
		return nil, fmt.Errorf("invalid device feature access mode %v", expose.Access)
	}

	// DEBUG
	if expose.Name == "target_distance" {
		fmt.Println("target_distance")
	}

	newEntity := newEntity()
	newEntity.AccessMode = accessMode
	newEntity.Category = getExposeCategory(expose)
	newEntity.Name = expose.Property
	newEntity.Description = expose.Description
	newEntity.Unit = expose.Unit
	newEntity.Data = data
	newEntity.Type = expose.Type

	switch expose.Type {
	case NumericDataType:
		if expose.ValueMax != nil {
			newEntity.Attributes["max"] = expose.ValueMax
		}
		if expose.ValueMin != nil {
			newEntity.Attributes["min"] = expose.ValueMin
		}

		if len(expose.Presets) != 0 {
			newEntity.Presets = make(map[string]any)
			for _, preset := range expose.Presets {
				newEntity.Values[preset.Name] = preset.Value
			}
		} else {
			newEntity.Values[expose.Name] = 0 // ????????? - i dont think i need this
		}

	case BinaryDataType:

		newEntity.Values["on"] = expose.ValueOn
		newEntity.Values["off"] = expose.ValueOff

		if expose.ValueToggle != "" {
			newEntity.Values["toggle"] = expose.ValueToggle
		}

	case EnumDataType:
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
		if _, ok := exposesWhitelist[key]; !ok {
			continue
		}

		newEntity := newEntity()
		newEntity.AccessMode = UnknownAccessMode
		newEntity.Category = MeasurementCategory
		newEntity.Name = key
		newEntity.Data = value
		newEntity.Unit = units[key]
		newEntity.Type = NumericDataType // TODO: make this dynamic
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
