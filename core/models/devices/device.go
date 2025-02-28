package devices

import (
	"errors"
	"fmt"
	"node-herder/utils"
	"sync"
	"time"
)

var availabilityKey = "availability"
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
type ExposeAccessMode = int

const (
	UnknownAccessMode ExposeAccessMode = 0b000
	StateAccessMode   ExposeAccessMode = 0b001 // although ican request, the device send updates about its value
	WriteAccessMode   ExposeAccessMode = 0b010 // it will request to set the value to device
	ReadAccessMode    ExposeAccessMode = 0b100 // it will request the read the value from device

	MeasurementCategory ExposeCategory = "measurement"
	DiagnosticCategory  ExposeCategory = "diagnostic"
	ConfigCategory      ExposeCategory = "config"

	NumericDataType   ExposeDataType = "numeric"
	EnumDataType      ExposeDataType = "enum"
	BinaryDataType    ExposeDataType = "binary"
	CompositeDataType ExposeDataType = "composite"
)

func ToFeatureAccessMode(access int) (ExposeAccessMode, bool) {
	convertedAccess := ExposeAccessMode(access)
	return convertedAccess, (convertedAccess&StateAccessMode | ReadAccessMode | WriteAccessMode) != 0
}

func IsUknownAccessMode(entity BridgeExpose) bool {
	return entity.Access&UnknownAccessMode != 0
}

func IsWriteableAccessMode(entity BridgeExpose) bool {
	return entity.Access&WriteAccessMode != 0

}
func IsReadAccessMode(entity BridgeExpose) bool {
	return entity.Access&ReadAccessMode != 0
}
func IsStateAccessMode(entity *BridgeExpose) bool {
	return entity.Access&StateAccessMode != 0
}

func getExposeCategory(entity BridgeExpose) string {
	if entity.Category == "" {
		return MeasurementCategory
	}
	return entity.Category
}

type Device struct {
	Id                      string             `json:"id"`
	FriendlyName            string             `json:"friendly_name"`
	Description             string             `json:"description,omitempty"`
	ConnectionType          string             `json:"connection_type"`
	PowerSource             string             `json:"power_source"`
	Exposes                 map[string]*Entity `json:"exposes"`
	Properties              map[string]any     `json:"properties"`
	LastSeen                string             `json:"last_seen"`
	Availability            AvailabilityType   `json:"availability"`
	availabilityTicker      time.Ticker
	availablityDone         chan bool
	availabilityTimeoutSecs int
	mutex                   sync.RWMutex
}

func NewDevice(id string) *Device {

	return &Device{
		Id:                      id,
		FriendlyName:            "",
		Description:             "",
		ConnectionType:          "",
		PowerSource:             "",
		Availability:            UnknownAvailability,
		LastSeen:                "",
		Exposes:                 map[string]*Entity{},
		Properties:              map[string]any{},
		availabilityTicker:      time.Ticker{},
		availablityDone:         make(chan bool, 1),
		availabilityTimeoutSecs: 3600,
		mutex:                   sync.RWMutex{},
	}
}

type PackageData map[string]any
type UpdatePackage struct {
	Id           string           `json:"id"`
	LastSeen     string           `json:"last_seen"`
	Availability AvailabilityType `json:"availability,omitempty"`
	Data         PackageData      `json:"data"`
}

func newUpdatePackage(id string) *UpdatePackage {
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
	Properties  map[string]any   `json:"properties,omitempty"` // NEEDS REMOVING !!!!!!!!!!!!!!!!!!1
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

	accessMode, ok := ToFeatureAccessMode(expose.Access)
	if !ok {
		return nil, fmt.Errorf("invalid device feature access mode %v", expose.Access)
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

		if IsWriteableAccessMode(expose) && expose.ValueToggle != "" {
			newEntity.Values["toggle"] = expose.ValueToggle
		}

	case EnumDataType:
		for id, item := range expose.Values {
			newEntity.Values[fmt.Sprintf("%d", id)] = item
		}
	}

	return newEntity, nil
}

func createExpose(data map[string]interface{}) map[string]*Entity {
	var entities = make(map[string]*Entity)
	for key, value := range data {
		if _, ok := exposesWhitelist[key]; !ok {
			continue
		}

		newEntity := newEntity()
		newEntity.AccessMode = UnknownAccessMode
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

				entity, err := CreateEntityFromExpose(expose, value)
				if err != nil {
					continue
				}
				entities[expose.Property] = entity
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

// TODO: do we need to sanitize/valdate time?
func getLastSeen(data map[string]interface{}) string {
	if val, ok := data[lastSeenKey]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	utils.LogDebug("lastSeen not in payload, using current time.")
	return getCurrentTime()
}

func (device *Device) Update(payload map[string]interface{}) *UpdatePackage {

	var updatePackage = newUpdatePackage(device.Id)
	for name, newValue := range payload {
		if expose, ok := device.Exposes[name]; ok && expose.Data != newValue {

			// we only care about measurements to determine
			// if there has been sensor changes. if we do have new sensor data
			// then collect everything that has changed

			if len(updatePackage.Data) != 0 {
				updatePackage.Data[name] = newValue
			} else if expose.Category == MeasurementCategory {
				updatePackage.Data[name] = newValue
			}

			// update device expose with updated data
			device.Exposes[name].Data = newValue
		}
	}

	defer device.mutex.Unlock()
	device.mutex.Lock()

	if updatePackage.HasData() {
		updatePackage.LastSeen = getLastSeen(payload)
	}

	if device.Availability == OfflineAvailability {
		device.Availability = OnlineAvailability

		// TODO: handle this below better
		// updatePackage contains Avaailability only if we have a change oin Device Availability. else its ommited.
		// thats because we use updatePackage for either measurement data or device availability change
		updatePackage.Availability = OnlineAvailability // we handle it manually for now.

		utils.LogInfof("device [%s] %s is online", device.Id, device.FriendlyName)
		device.resetAvailabilityTimer()
	}

	device.LastSeen = getLastSeen(payload) // we need that.
	return updatePackage
}

func (device *Device) LastSeenTime() (time.Time, error) {

	defer device.mutex.RUnlock()
	device.mutex.RLock()

	lastSeenStr, _ := device.Properties[lastSeenKey].(string)
	lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
	if err != nil {
		return time.Time{}, err
	}

	return lastSeen, nil
}

func (device *Device) isAvailable() bool {

	defer device.mutex.RUnlock()
	device.mutex.RLock()

	return device.Availability == OnlineAvailability
}

func (device *Device) setAvailable(value bool) {

	defer device.mutex.Unlock()
	device.mutex.Lock()
	if value {
		device.Availability = OnlineAvailability
	} else {
		device.Availability = OfflineAvailability
	}
}

func (device *Device) Dispose() {
	device.availablityDone <- true
	device.availabilityTicker.Stop()
	utils.LogDebugf("device %s disposed", device.Id)
}

func (device *Device) resetAvailabilityTimer() {

	device.availabilityTicker.Reset(1 * time.Second)
}

func (device *Device) Monitor(timeoutInSecs int, onChangeCallback func(p interface{})) {

	device.availabilityTicker = *time.NewTicker(1 * time.Second)

	go func() {
		defer close(device.availablityDone)
		for {
			select {
			case <-device.availablityDone:

				device.setAvailable(false)
				utils.LogInfof("device %s availability timer killed", device.Id)

				// todo: move it in one place
				if onChangeCallback != nil {
					p := newUpdatePackage(device.Id)
					p.Availability = OfflineAvailability
					onChangeCallback(p)
				}

				return

			case <-device.availabilityTicker.C:

				if !device.isAvailable() {
					return
				}

				lastSeen, err := device.LastSeenTime()
				if err != nil {
					utils.LogErrorf("device %s failed to parse time %s", device.Id, err.Error())

					device.Dispose()
				}

				now := time.Now()
				diff := now.Sub(lastSeen)
				if diff.Seconds() >= float64(timeoutInSecs) {

					device.setAvailable(false)
					utils.LogInfof("device %s is offine", device.Id)

					// todo: move it in one place
					if onChangeCallback != nil {
						p := newUpdatePackage(device.Id)
						p.Availability = OfflineAvailability
						onChangeCallback(p)
					}

					device.availabilityTicker.Stop()
				}

			}
		}
	}()
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
