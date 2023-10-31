package devices

import (
	"errors"
	"fmt"
	"node-herder/utils"
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
}
var exposesWhitelist = map[string]int{
	"temperature":       1,
	"humidity":          2,
	"pressure":          3,
	"presence":          4,
	"illuminance_lux":   5,
	"occupancy":         6,
	"brightness":        7,
	"state":             8,
	"color_temp":        9,
	"air_quality_score": 10,
	"pm1":               11,
	"pm25":              12,
	"pm10":              13,
}

var propertiesWhitelist = map[string]int{
	"battery":     1,
	"linkquality": 2,
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
	}
}

type updatePackage struct {
	Id         string         `json:"id"`
	LastSeen   string         `json:"last_seen"`
	Data       map[string]any `json:"data"`
	Properties map[string]any `json:"properties"`
}

func newUpdatePackage(id string) *updatePackage {
	return &updatePackage{Id: id, LastSeen: getCurrentTime(), Data: make(map[string]any), Properties: make(map[string]any)}
}

func (u *updatePackage) HasData() bool {
	return len(u.Data) != 0
}

type Entity struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Unit        string         `json:"unit,omitempty"`
	Data        any            `json:"data"`
	Type        string         `json:"type"`
	Properties  map[string]any `json:"properties"`
	Attributes  map[string]any `json:"attributes"`
}

func newEntity() *Entity {
	return &Entity{Attributes: make(map[string]any), Properties: map[string]any{}}
}

func CreateFromFeature(feature BridgeInfoFeature) (*Entity, error) {

	if _, ok := exposesWhitelist[feature.Property]; !ok {
		return nil, fmt.Errorf("feature property %v is blacklisted", feature.Property)
	}

	newEntity := newEntity()
	newEntity.Name = feature.Property
	newEntity.Description = feature.Description
	newEntity.Unit = feature.Unit

	newEntity.Properties["type"] = feature.Type
	newEntity.Properties["feature"] = true

	switch feature.Type {
	case "numeric":
		newEntity.Properties["max"] = feature.ValueMax
		newEntity.Properties["min"] = feature.ValueMin

	case "binary":
		newEntity.Properties["on"] = feature.ValueOn
		newEntity.Properties["off"] = feature.ValueOff
		newEntity.Properties["toggle"] = feature.ValueToggle

	case "enum":
		newEntity.Properties["values"] = feature.Values
	}

	return newEntity, nil
}

func CreateFromExpose(expose BridgeExpose) (*Entity, error) {

	if expose.Property == "" {
		return nil, fmt.Errorf("no expose data")
	}

	if _, ok := exposesWhitelist[expose.Property]; !ok {
		return nil, fmt.Errorf("expose property %v is blacklisted", expose.Property)
	}

	newEntity := newEntity()
	newEntity.Name = expose.Property
	newEntity.Description = expose.Description
	newEntity.Unit = expose.Unit
	newEntity.Properties["type"] = expose.Type
	newEntity.Properties["feature"] = false

	return newEntity, nil
}

func createEntityFromExpose(expose BridgeExpose, data any) (*Entity, error) {

	if expose.Property == "" {
		return nil, fmt.Errorf("no expose data")
	}

	if _, ok := exposesWhitelist[expose.Property]; !ok {
		return nil, fmt.Errorf("expose property %v is blacklisted", expose.Property)
	}

	newEntity := newEntity()
	newEntity.Name = expose.Property
	newEntity.Description = expose.Description
	newEntity.Unit = expose.Unit
	newEntity.Data = data
	newEntity.Type = expose.Type

	return newEntity, nil
}

func createEntityFromFeature(feature BridgeInfoFeature, data any) (*Entity, error) {

	if _, ok := exposesWhitelist[feature.Property]; !ok {
		return nil, fmt.Errorf("feature property %v is blacklisted", feature.Property)
	}

	newEntity := newEntity()
	newEntity.Data = data
	newEntity.Name = feature.Name
	newEntity.Unit = feature.Unit
	newEntity.Description = feature.Description
	newEntity.Data = data
	newEntity.Type = feature.Type

	switch feature.Type {
	case "numeric":
		newEntity.Attributes["max"] = feature.ValueMax
		newEntity.Attributes["min"] = feature.ValueMin

	case "binary":
		newEntity.Properties["on"] = feature.ValueOn
		newEntity.Properties["off"] = feature.ValueOff

	case "enum":
		for index, item := range feature.Values {
			newEntity.Properties[fmt.Sprintf("%s", index)] = item
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
	data[availabilityKey] = online

	return props
}

func createExposures(data map[string]interface{}) map[string]*Entity {
	var entities = make(map[string]*Entity)
	for key, value := range data {
		if _, ok := exposesWhitelist[key]; !ok {
			continue
		}

		newEntity := createEntity(key, "", value, units[key], "", nil)
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
				entity, err := createEntityFromExpose(expose, value)
				if err != nil {
					continue
				}
				entities[expose.Property] = entity
			}
		}
		// load features
		for _, feature := range expose.Features {
			if value, ok := data[feature.Property]; ok {

				entity, err := createEntityFromExpose(expose, value)
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

func (device *Device) Update(payload map[string]interface{}) *updatePackage {

	var updatePackage = newUpdatePackage(device.Id)
	for name, currValue := range device.Exposes {
		if newValue, ok := payload[name]; ok && newValue != currValue.Data {
			device.Exposes[name].Data = newValue
			updatePackage.Data[name] = newValue
		}

	}
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
		utils.LogInfof("device [%s] %s is online", device.Id, device.FriendlyName)
		device.resetAvailabilityTimer()
	}

	if _, ok := payload[lastSeenKey]; !ok {
		payload[lastSeenKey] = getCurrentTime()
	}

	device.Properties[lastSeenKey] = payload[lastSeenKey] // we need that.
	return updatePackage
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

				device.Properties[availabilityKey] = offline
				utils.LogInfof("device %s availability timer killed", device.Id)

				// todo: move it in one place
				if onChangeCallback != nil {
					p := newUpdatePackage(device.Id)
					p.Properties[availabilityKey] = offline
					onChangeCallback(p)
				}

				return

			case <-device.availabilityTicker.C:

				if device.Properties[availabilityKey] == offline {
					return
				}

				lastSeenStr, _ := device.Properties[lastSeenKey].(string)
				lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
				if err != nil {
					utils.LogErrorf("device %s failed to parse time %s", device.Id, err.Error())
					device.Dispose()
				}

				now := time.Now()
				diff := now.Sub(lastSeen)
				if diff.Seconds() >= float64(timeoutInSecs) {

					device.Properties[availabilityKey] = offline
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
