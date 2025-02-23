package devices

import (
	"errors"
	"fmt"
	"node-herder/utils"
	"sync"
	"time"
)

var availabilityKey = "availability"
var linkQualityKey = "linkquality"
var mainsKey = "mains"
var powerSourceKey = "power_source"
var batterKey = "battery"
var lastSeenKey = "last_seen"
var idKey = "id"
var connectionTypeMqtt = "mqtt"

const online = "online"
const offline = "offline"

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

type ExposeType = string

const (
	NumericExposeType ExposeType = "numeric"
	EnumExposeType    ExposeType = "enum"
	BinaryExposeType  ExposeType = "binary"
)

type FeatureAccessMode = int

const (
	UnknownAccessMode FeatureAccessMode = 0b000
	StateAccessMode   FeatureAccessMode = 0b001
	WriteAccessMode   FeatureAccessMode = 0b010
	ReadAccessMode    FeatureAccessMode = 0b100
)

func ToFeatureAccessMode(access int) (FeatureAccessMode, bool) {
	convertedAccess := FeatureAccessMode(access)
	return convertedAccess, (convertedAccess&StateAccessMode | ReadAccessMode | WriteAccessMode) != 0
}

func isUknownAccessMode(entity *Entity) bool {
	return entity.AccessMode&UnknownAccessMode != 0
}

func isWriteableAccessMode(entity *Entity) bool {
	return entity.AccessMode&WriteAccessMode != 0

}
func isReadAccessMode(entity *Entity) bool {
	return entity.AccessMode&ReadAccessMode != 0
}
func isStateAccessMode(entity *Entity) bool {
	return entity.AccessMode&StateAccessMode != 0
}

var customExposeFeaturesPropertyWhitelist = map[string]int{
	"silence":  1,
	"alarm":    2,
	"melody":   3,
	"duration": 4,
	"volume":   5, // ??
}

var propertiesWhitelist = map[string]int{
	"battery":        1,
	"linkquality":    2,
	"battpercentage": 3,
	//"availability": 3,handled manually
	//"last_seen":    4,handled manually
}

type Device struct {
	Id                      string             `json:"id"`
	FriendlyName            string             `json:"friendly_name"`
	Description             string             `json:"description,omitempty"`
	ConnectionType          string             `json:"connection_type"`
	PowerSource             string             `json:"power_source"`
	Exposes                 map[string]*Entity `json:"exposes"`
	Properties              map[string]any     `json:"properties"`
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
	Id         string         `json:"id"`
	LastSeen   string         `json:"last_seen"`
	Data       PackageData    `json:"data"`
	Properties map[string]any `json:"properties"`
}

func newUpdatePackage(id string) *UpdatePackage {
	return &UpdatePackage{Id: id, LastSeen: getCurrentTime(), Data: make(map[string]any), Properties: make(map[string]any)}
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
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	Data        any               `json:"data"`
	Type        ExposeType        `json:"type"`
	AccessMode  FeatureAccessMode `json:"access_mode"`
	Properties  map[string]any    `json:"properties,omitempty"`
	Attributes  map[string]any    `json:"attributes,omitempty"`
	Presets     map[string]any    `json:"presets,omitempty"`
}

func newEntity() *Entity {
	return &Entity{Attributes: make(map[string]any), Properties: map[string]any{}}
}

func CreateEntityFromExpose(expose BridgeExpose, data any) (*Entity, error) {

	if expose.Property == "" {
		return nil, fmt.Errorf("no expose data")
	}

	if _, ok := exposesWhitelist[expose.Property]; !ok {
		return nil, fmt.Errorf("expose property %v is blacklisted", expose.Property)
	}

	accessMode, ok := ToFeatureAccessMode(expose.Access)
	if !ok {
		return nil, fmt.Errorf("invalid device feature access mode %v", expose.Access)
	}

	newEntity := newEntity()
	newEntity.AccessMode = accessMode
	newEntity.Name = expose.Property
	newEntity.Description = expose.Description
	newEntity.Unit = expose.Unit
	newEntity.Data = data
	newEntity.Type = expose.Type

	// NOTE:
	// exposes are not used to publish events
	// because of that all expose data are published as device attributes

	// TODO: needs refactoring
	switch expose.Type {
	case NumericExposeType:
		if expose.ValueMax != nil {
			newEntity.Attributes["max"] = expose.ValueMax
		}
		if expose.ValueMin != nil {
			newEntity.Attributes["min"] = expose.ValueMin
		}

	case BinaryExposeType:
		if expose.ValueOn != nil {
			newEntity.Attributes["on"] = expose.ValueOn
		}
		if expose.ValueOff != nil {
			newEntity.Attributes["off"] = expose.ValueOff
		}

	case EnumExposeType:
		for _, item := range expose.Values {
			newEntity.Attributes[item] = item
		}
	}
	return newEntity, nil
}

func CreateCustomFeatureFromExpose(expose BridgeExpose, data any) (*Entity, error) {

	// NOTE:
	// create custom feature from devices that can be triggered but dont have feature description
	// only support alarm for now
	// Features are used to publish events
	// because of that all expose data are published as device properties

	if expose.Property == "" {
		return nil, fmt.Errorf("no expose data")
	}

	if _, ok := customExposeFeaturesPropertyWhitelist[expose.Property]; !ok {
		return nil, fmt.Errorf("custom expose property %v is blacklisted", expose.Property)
	}

	accessMode, ok := ToFeatureAccessMode(expose.Access)
	if !ok {
		return nil, fmt.Errorf("invalid device feature access mode %v", expose.Access)
	}
	newEntity := newEntity()
	newEntity.AccessMode = accessMode
	newEntity.Data = data
	newEntity.Name = expose.Name
	newEntity.Unit = expose.Unit
	newEntity.Description = expose.Description
	newEntity.Data = data
	newEntity.Type = expose.Type

	// NOTE:
	// features are used to publish events
	// because of that all expose data are published as device properties
	switch expose.Type {
	case NumericExposeType:
		newEntity.Attributes["max"] = expose.ValueMax
		newEntity.Attributes["min"] = expose.ValueMin
		newEntity.Properties[expose.Name] = 0

	case BinaryExposeType:
		newEntity.Properties["on"] = expose.ValueOn
		newEntity.Properties["off"] = expose.ValueOff

	case EnumExposeType:
		for index, item := range expose.Values {
			newEntity.Properties[fmt.Sprintf("%d", index)] = item
		}
	}
	return newEntity, nil

}

todo
// create a single entity constructor driven by the AccessMode 

func CreateEntityFromFeature(feature BridgeInfoFeature, data any) (*Entity, error) {

	if _, ok := exposesWhitelist[feature.Property]; !ok {
		return nil, fmt.Errorf("feature property %v is blacklisted", feature.Property)
	}

	accessMode, ok := ToFeatureAccessMode(feature.Access)
	if !ok {
		return nil, fmt.Errorf("invalid device feature access mode %v", feature.Access)
	}

	newEntity := newEntity()
	newEntity.Data = data
	newEntity.AccessMode = accessMode
	newEntity.Name = feature.Name
	newEntity.Unit = feature.Unit
	newEntity.Description = feature.Description
	newEntity.Data = data
	newEntity.Type = feature.Type

	// populate feature presets
	if len(feature.Presets) != 0 {
		newEntity.Presets = make(map[string]any)
		for _, preset := range feature.Presets {
			newEntity.Presets[preset.Name] = preset.Value
		}
	}

	switch feature.Type {
	case NumericExposeType:
		newEntity.Attributes["max"] = feature.ValueMax
		newEntity.Attributes["min"] = feature.ValueMin
		newEntity.Properties[feature.Name] = 0

	case BinaryExposeType:
		newEntity.Properties["on"] = feature.ValueOn
		newEntity.Properties["off"] = feature.ValueOff
		newEntity.Properties["toggle"] = feature.ValueToggle

	case EnumExposeType:
		for index, item := range feature.Values {
			newEntity.Properties[fmt.Sprintf("%d", index)] = item
		}
	}
	return newEntity, nil
}

func createProperties(data map[string]interface{}) map[string]any {

	var props = map[string]any{}
	for key, value := range data {
		if _, ok := propertiesWhitelist[key]; ok {
			props[key] = value
		}
	}

	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}

	props[lastSeenKey] = data[lastSeenKey]
	props[availabilityKey] = online

	return props
}

func createExposures(data map[string]interface{}) map[string]*Entity {
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
		newEntity.Type = NumericExposeType // TODO: make this dynamic
		entities[key] = newEntity
	}
	return entities
}

func createExposuresFromBridge(data map[string]interface{}, bridgeInfo *BridgeInfo) map[string]*Entity {

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
		newDevice.Exposes = createExposuresFromBridge(data, bridgeInfo)
	} else { // device not in hub bridge
		newDevice.Exposes = createExposures(data)
	}

	newDevice.Properties = createProperties(data)
	if len(newDevice.Exposes) == 0 {
		return nil, errors.New("invalid payload - no exposed entries found")
	}

	return newDevice, nil
}

func (device *Device) Update(payload map[string]interface{}) *UpdatePackage {

	var updatePackage = newUpdatePackage(device.Id)
	for name, currValue := range device.Exposes {
		if newValue, ok := payload[name]; ok && newValue != currValue.Data {
			device.Exposes[name].Data = newValue
			updatePackage.Data[name] = newValue
		}
	}

	// Update device properties
	defer device.mutex.Unlock()
	device.mutex.Lock()

	if updatePackage.HasData() {

		for name := range propertiesWhitelist {
			currValue := device.Properties[name]
			if newValue, ok := payload[name]; ok && newValue != currValue {

				device.Properties[name] = newValue
				updatePackage.Properties[name] = newValue
			}
		}
	}

	if device.Properties[availabilityKey] != online {
		device.Properties[availabilityKey] = online
		updatePackage.Properties[availabilityKey] = online // we handle it manually for now

		utils.LogInfof("device [%s] %s is online", device.Id, device.FriendlyName)
		device.resetAvailabilityTimer()
	}

	if _, ok := payload[lastSeenKey]; !ok {
		payload[lastSeenKey] = getCurrentTime()
	}

	device.Properties[lastSeenKey] = payload[lastSeenKey] // we need that.
	return updatePackage
}

func (device *Device) LastSeen() (time.Time, error) {

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

	return device.Properties[availabilityKey] == online
}

func (device *Device) setAvailable(value bool) {

	defer device.mutex.Unlock()
	device.mutex.Lock()
	if value {
		device.Properties[availabilityKey] = online
	} else {
		device.Properties[availabilityKey] = offline
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
					p.Properties[availabilityKey] = offline
					onChangeCallback(p)
				}

				return

			case <-device.availabilityTicker.C:

				if !device.isAvailable() {
					return
				}

				lastSeen, err := device.LastSeen()
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
						p.Properties[availabilityKey] = offline
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
