package devices

import (
	"errors"
	"node-herder/models/devices"
	"node-herder/utils"
	"time"
)

var availabilityKey = "availability"
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
	"battery":      1,
	"linkquality":  2,
	"availability": 3,
	"last_seen":    4,
}

type DeviceV2 struct {
	Id                      string             `json:"id"`
	FriendlyName            string             `json:"friendlyname"`
	Description             string             `json:"description,omitempty"`
	ConnectionType          string             `json:"connection_type"`
	PowerSource             string             `json:"power_source"`
	Exposes                 map[string]*Entity `json:"exposes"`
	Properties              map[string]any     `json:"properties"`
	availabilityTicker      time.Ticker
	availablityDone         chan bool
	availabilityTimeoutSecs int
}

func build(bridgeDevices []*devices.BridgeInfo) []*DeviceV2 {
	for _, bridgeInfo := range bridgeDevices {
		if !bridgeInfo.IsActive() {
			continue
		}
		d := newDeviceV2(bridgeInfo.id)
		d.ConnectionType = "mqtt"
		d.Description = bridgeInfo.Description
		d.FriendlyName = bridgeInfo.FriendlyName
		d.PowerSource = bridgeInfo.PowerSource

		for _, expose := range bridgeInfo.Definition.Exposes {

			// load exposes
			if expose.Property != "" {

				if _, ok := exposesWhitelist[expose.Property]; !ok {
					continue
				}

				d.Exposes[expose.Property] = createEntity(expose.Property, expose.Description, nil, expose.Unit, expose.Type, nil)
			}

			// load features
			// for _, feature := range expose.Features {
			// 	if value, ok := data[feature.Property]; ok {

			// 		var props map[string]any = make(map[string]any)
			// 		switch feature.Type {
			// 		case "numeric":
			// 			props["type"] = "numeric"
			// 			props["max"] = feature.ValueMax
			// 			props["min"] = feature.ValueMin

			// 		case "binary":
			// 			props["type"] = "binary"
			// 			props["on"] = feature.ValueOn
			// 			props["off"] = feature.ValueOff
			// 			props["toggle"] = feature.ValueToggle

			// 		case "enum":
			// 			props["type"] = "enum"
			// 			props["values"] = feature.Values
			// 		}

			// 		entities[feature.Property] = createEntity(feature.Property, feature.Description, value, feature.Unit, feature.Type, props)
			// 	}
			// }
		}

	}

}
func newDeviceV2(id string) *DeviceV2 {

	return &DeviceV2{
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
	Id       string         `json:"id"`
	LastSeen string         `json:"last_seen"`
	Data     map[string]any `json:"data"`
}

func newUpdatePackage(id string) *updatePackage {
	return &updatePackage{Id: id, LastSeen: getCurrentTime(), Data: make(map[string]any)}
}

func (u *updatePackage) HasData() bool {
	return len(u.Data) != 0
}

type Entity struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Unit        string         `json:"unit,omitempty"`
	Data        any            `json:"data"`
	Properties  map[string]any `json:"properties"`
}

func createFromExpose(expose *BridgeExpose) *Entity {
	newEntity := &Entity{}
 do null checks
	if expose.Property != "" {
		newEntity.Name = expose.Property
		newEntity.Description = expose.Description
		newEntity.Unit = expose.Unit
		var props = map[string]any{}
		props["type"] = expose.Type
		switch expose.Type {
		case "numeric":
			props["type"] = "numeric"
			props["max"] = expose.ValueMax
			props["min"] = expose.ValueMin
			props["step"] = expose.ValueStep

		case "binary":
			props["type"] = "binary"
			props["on"] = expose.ValueOn
			props["off"] = expose.ValueOff

		case "enum":
			props["type"] = "enum"
			props["values"] = expose.Values
		}
	}
	for _, feature := range expose.Features {
		switch feature.Type {
		case "numeric":
			props["type"] = "numeric"
			props["max"] = feature.ValueMax
			props["min"] = feature.ValueMin

		case "binary":
			props["type"] = "binary"
			props["on"] = feature.ValueOn
			props["off"] = feature.ValueOff
			props["toggle"] = feature.ValueToggle

		case "enum":
			props["type"] = "enum"
			props["values"] = feature.Values
		}
	}
}
func createEntity(name string, description string, data any, unit string, dataType string, props map[string]any) *Entity {
	if props == nil {
		props = make(map[string]any)
	}
	newEntity := &Entity{}
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Unit = unit
	newEntity.Description = description
	newEntity.Properties = props
	return newEntity
}

func createProperties(data map[string]interface{}) map[string]any {

	var props = map[string]any{}
	for key, value := range data {
		if _, ok := propertiesWhitelist[key]; ok {
			props[key] = value
		}
	}

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

			if _, ok := exposesWhitelist[expose.Property]; !ok {
				continue
			}

			if value, ok := data[expose.Property]; ok {
				entities[expose.Property] = createEntity(expose.Property, expose.Description, value, expose.Unit, expose.Type, nil)
			}
		}
		// load features
		for _, feature := range expose.Features {
			if value, ok := data[feature.Property]; ok {

				var props map[string]any = make(map[string]any)
				switch feature.Type {
				case "numeric":
					props["type"] = "numeric"
					props["max"] = feature.ValueMax
					props["min"] = feature.ValueMin

				case "binary":
					props["type"] = "binary"
					props["on"] = feature.ValueOn
					props["off"] = feature.ValueOff
					props["toggle"] = feature.ValueToggle

				case "enum":
					props["type"] = "enum"
					props["values"] = feature.Values
				}

				entities[feature.Property] = createEntity(feature.Property, feature.Description, value, feature.Unit, feature.Type, props)
			}
		}
	}

	return entities
}

func CreateNewDeviceV2(repo Repository, friendlyName string, connType string, data map[string]interface{}) (*DeviceV2, error) {

	if _, ok := data[lastSeenKey]; !ok {
		data[lastSeenKey] = getCurrentTime()
	}
	data[availabilityKey] = online

	id := repo.ResolveId(friendlyName)

	var newDevice = newDeviceV2(id)
	newDevice.FriendlyName = friendlyName
	newDevice.ConnectionType = connType

	if _, ok := data[batterKey]; !ok {
		newDevice.PowerSource = mainsKey
	} else {
		newDevice.PowerSource = batterKey
	}

	bridgeInfo := repo.FindBridgeInfo(id)
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

func (device *DeviceV2) TryUpdate(payload map[string]interface{}) *updatePackage {

	var updatePackage = newUpdatePackage(device.Id)
	for name, currValue := range device.Exposes {
		if newValue, ok := payload[name]; ok && newValue != currValue.Data {
			device.Exposes[name].Data = newValue
			updatePackage.Data[name] = newValue
		}
	}

	if device.Properties[availabilityKey] != online {
		device.Properties[availabilityKey] = online
		utils.LogInfof("device %s is online", device.Id)
		device.resetAvailabilityTimer()
	}

	if _, ok := payload[lastSeenKey]; !ok {
		payload[lastSeenKey] = getCurrentTime()
	}

	device.Properties[lastSeenKey] = payload[lastSeenKey] // we need that.
	return updatePackage
}

func (device *DeviceV2) Dispose() {
	device.availablityDone <- true
	device.availabilityTicker.Stop()
	utils.LogDebugf("device %s disposed", device.Id)
}

func (device *DeviceV2) resetAvailabilityTimer() {

	device.availabilityTicker.Reset(1 * time.Second)
}

func (device *DeviceV2) Monitor(timeoutInSecs int, onChangeCallback func(p interface{})) {

	device.availabilityTicker = *time.NewTicker(1 * time.Second)

	go func() {
		defer close(device.availablityDone)
		for {
			select {
			case <-device.availablityDone:

				device.Properties[availabilityKey] = offline
				utils.LogInfof("device %s availability timer killed", device.Id)

				if onChangeCallback != nil {
					p := newUpdatePackage(device.Id)
					p.Data[availabilityKey] = offline
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

					if onChangeCallback != nil {
						p := newUpdatePackage(device.Id)
						p.Data[availabilityKey] = offline
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
