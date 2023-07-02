package device

import (
	"errors"
	"time"
)

// sensor whitelist map
var powerSource = map[string]int{
	"battery":              1,
	"Mains (single phase)": 2,
	"DC source":            3,
}

var sensorWhitelist = map[string]int{
	"temperature":     1,
	"humidity":        2,
	"pressure":        3,
	"presence":        4,
	"illuminance_lux": 5,
}
var lastSeenKey = "last_seen"
var batterKey = "battery"
var mainsKey = "Mains (single phase)"
var powerSourceKey = "power_source"

// todo:
// collect data and pass them to different process
func Process(payload interface{}) {

	data, err := convertToMap(payload)
	if err != nil {
		panic("fooked")
	}

	// todo:
	// check if data exists in store before continue with logic

	// todo
	// have a timer to see if item is available, if not set offline
	data[powerSourceKey] = batterKey
	if _, ok := data[batterKey]; !ok {
		data[powerSourceKey] = mainsKey
	}

	data["availability"] = "offline"
	if lastSeen, ok := data[lastSeenKey]; ok {
		data["availability"] = "online"
		// if first time, store
		// else get previous item and compare if anything changed  ?
		// if sth changed then broadcast
		// else flag as online and dont broadcast

	} else {
		data["lastSeenKey"] = time.Now() // fix format
	}

}

func convertToMap(payload interface{}) (map[string]interface{}, error) {
	if data, ok := payload.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("invalid device data")

}
