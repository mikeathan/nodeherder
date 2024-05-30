package repository_test

import (
	"fmt"
	"math"
	"math/rand"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/repository"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestMultipleDeviceTimeRangeMetrics(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		id         string
		timestamps []*time.Time
		values     []float32
		from       time.Time
		to         time.Time
	}{
		{id: "x0000",
			timestamps: CreateDateTimeTimestamps(1, 24, 1), // 24 events
			values:     CreateFloatValues(24),
			from:       time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC),
			to:         time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)},
		{id: "x0001",
			timestamps: CreateDateTimeTimestamps(10, 24, 1), values: CreateFloatValues(240), // 240 events
			from: time.Date(now.Year(), now.Month()-6, now.Day()-5, 15, 0, 0, 0, time.UTC),
			to:   time.Date(now.Year(), now.Month()-2, now.Day()-2, 20, 0, 0, 0, time.UTC)},
		{id: "x0002",
			timestamps: CreateDateTimeTimestamps(60, 2, 1), values: CreateFloatValues(120), // 240 events
			from: time.Date(now.Year(), now.Month()-20, now.Day()-5, 15, 0, 0, 0, time.UTC),
			to:   time.Date(now.Year(), now.Month()-5, now.Day()-2, 20, 0, 0, 0, time.UTC)},
	}

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	for i, test := range testCases {
		fmt.Println(len(test.timestamps))

		deviceName := fmt.Sprintf("device %v", i)

		fmt.Println("total timestamps: ", len(test.timestamps))
		for tIdx, timestamp := range test.timestamps {

			dev := createMockDevice(test.id, deviceName, 2, "numeric", *timestamp, test.values[tIdx])

			err := repo.Store(dev)

			//fmt.Printf("Add device: %v, data: %v, timestamp: %v \n", deviceId, values[tIdx], dev.Properties["last_seen"])
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	for _, test := range testCases {

		dev := devices[test.id]

		result, err := repo.ViewDeviceTimeRange(dev, test.from, test.to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		assertDeviceEvents(dev, result, test.timestamps, test.values, t)
	}
}

func TestDeviceTimeRangeMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := CreateDateTimeTimestamps(1, 24, 1)
	values := CreateFloatValues(24)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	numDevices := 3

	for i := 0; i < numDevices; i++ {

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		fmt.Println("total timestamps: ", len(timestamps))
		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, 2, "numeric", *timestamp, values[tIdx])

			err = repo.Store(dev)
			//fmt.Printf("Add device: %v, data: %v, timestamp: %v \n", deviceId, values[tIdx], dev.Properties["last_seen"])
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	// query and assert
	for i := 0; i < numDevices; i++ {
		deviceId := fmt.Sprintf("x000%v", i)

		dev := devices[deviceId]
		now := time.Now()

		from := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)

		result, err := repo.ViewDeviceTimeRange(dev, from, to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		assertDeviceEvents(dev, result, timestamps, values, t)
	}
}

func TestMetricsRateLimiter(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := CreateDateTimeTimestamps(1, 24, 1)
	values := CreateFloatValues(24)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	numDevices := 1

	for i := 0; i < numDevices; i++ {

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		fmt.Println("total timestamps: ", len(timestamps))
		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, 2, "numeric", *timestamp, values[tIdx])

			err = repo.Store(dev)
			//fmt.Printf("Add device: %v, data: %v, timestamp: %v \n", deviceId, values[tIdx], dev.Properties["last_seen"])
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	// query and assert
	for i := 0; i < numDevices; i++ {
		deviceId := fmt.Sprintf("x000%v", i)

		dev := devices[deviceId]
		now := time.Now()

		from := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

		result, err := repo.ViewDeviceTimeRange(dev, from, to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		for _, expose := range result.Expose {
			if len(expose.Values) > 1 {
				t.Errorf("rate limiter registered more than expected device hits want 1 got %v: ", len((expose.Values)))
			}
		}
	}
}

func TestDeviceTimeRangeDataTypesMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := CreateDateTimeTimestamps(1, 24, 1)
	data := map[string]any{

		"numeric": CreateFloatValues(24),
		"binary":  CreateBinaryValues(24),
		"enum":    CreateEnumValues(24),
	}

	// sort data keys
	dataKeys := make([]string, 0, len(data))
	for k := range data {
		dataKeys = append(dataKeys, k)
	}

	sort.Strings(dataKeys)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	for i, dataType := range dataKeys {

		values := data[dataType]
		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		for tIdx, timestamp := range timestamps {
			var value any = 0
			if dataType == "numeric" {
				v, _ := values.([]float32)
				value = v[tIdx]

			} else if dataType == "binary" {
				v, _ := values.([]string)
				value = v[tIdx]

			} else if dataType == "enum" {
				v, _ := values.([]int)
				value = v[tIdx]
			}

			dev := createMockDevice(deviceId, deviceName, 2, dataType, *timestamp, value)
			err = repo.Store(dev)
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	// query and assert

	for i, dataType := range dataKeys {

		values := data[dataType]
		deviceId := fmt.Sprintf("x000%v", i)

		dev := devices[deviceId]
		now := time.Now()

		from := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)

		result, err := repo.ViewDeviceTimeRange(dev, from, to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		assertDeviceAnyDataTypeEvents(dev, result, timestamps, values, t)
	}
}

func TestExposeTimeRangeMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := CreateDateTimeTimestamps(1, 24, 1)
	values := CreateFloatValues(24)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	numDevices := 3

	for i := 0; i < numDevices; i++ {

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, 5, "numeric", *timestamp, values[tIdx])
			err = repo.Store(dev)
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	// query and assert
	for i := 0; i < numDevices; i++ {
		deviceId := fmt.Sprintf("x000%v", i)

		dev := devices[deviceId]
		now := time.Now()

		from := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)

		property := fmt.Sprintf("property_%v_%v", deviceId, 1)

		result, err := repo.ViewExposeTimeRange(dev, property, from, to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		assertDeviceExposeFloatDataEvents(dev, result, property, timestamps, values, t)
	}
}

func TestMultipleExposeTimeRangeMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := CreateDateTimeTimestamps(1, 24, 1)
	values := CreateFloatValues(24)
	numOfExposes := 10
	deviceIds := []string{"x0000"}
	exposes := []string{"property_x0000_1", "property_x0000_2", "property_x0000_4"}
	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	for _, deviceId := range deviceIds {

		deviceName := fmt.Sprintf("device %v", deviceId)

		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, numOfExposes, "numeric", *timestamp, values[tIdx])
			err = repo.Store(dev)
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	// query and assert
	for _, deviceId := range deviceIds {
		for _, exposeName := range exposes {
			dev := devices[deviceId]
			now := time.Now()

			from := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC)
			to := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)

			result, err := repo.ViewExposeTimeRange(dev, exposeName, from, to)
			if err != nil {
				t.Error("failed to query metrics: ", err.Error())
			}

			assertDeviceExposeFloatDataEvents(dev, result, exposeName, timestamps, values, t)
		}
	}
}

func assertDeviceExposeFloatDataEvents(device *devices.Device, result *metrics.DeviceMetricsResult, exposeName string, timestamps []*time.Time, values []float32, t *testing.T) {
	if result.DeviceId != device.Id {
		t.Errorf("deviceId mismatch want %v got %v: ", device.Id, result.DeviceId)
	}

	gotNumExposes := len(result.Expose)
	wantNumExposes := 1
	if gotNumExposes != wantNumExposes {
		t.Errorf("Exposes mismatch want %v got %v: ", wantNumExposes, gotNumExposes)
	}

	expose := device.Exposes[exposeName]
	event := result.Expose[0]
	if event.Name != expose.Name {
		t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, event.Name)
	}
	dataType, ok := kindFromString(event.Type)
	if !ok {
		t.Errorf("invalid event.type want %v got %v: ", event.Type, dataType)
	}

	for fId, foundTs := range event.Timestamp {
		tsFound := false
		dataIdx := 0
		foundTsUnix := foundTs.Unix()

		for insertIdx, insertTs := range timestamps {
			insertTsUnix := insertTs.Unix()

			if foundTsUnix == insertTsUnix {
				tsFound = true
				dataIdx = insertIdx
				break
			}
		}

		if !tsFound {
			t.Fatalf(fmt.Sprintf("Timestamp not found %v", foundTs))
		}

		wantKind := reflect.TypeOf(values[dataIdx]).Kind()
		if dataType != wantKind {
			t.Fatalf(fmt.Sprintf("Type mismatch want: %v got: %v", wantKind.String(), dataType.String()))
		}

		if values[dataIdx] != event.Values[fId] {
			t.Fatalf(fmt.Sprintf("Value mismatch want:%v got: %v", values[dataIdx], event.Values[fId]))
		}
		//fmt.Printf("found event %v with data: %v, timestamp: %v \n", expose.Name, event.Values[fId], foundTs)
	}

}
func assertDeviceAnyDataTypeEvents(device *devices.Device, result *metrics.DeviceMetricsResult, timestamps []*time.Time, values any, t *testing.T) {

	if result.DeviceId != device.Id {
		t.Errorf("deviceId mismatch want %v got %v: ", device.Id, result.DeviceId)
	}

	gotNumExposes := len(result.Expose)
	wantNumExposes := len(device.Exposes)
	if gotNumExposes != wantNumExposes {
		t.Errorf("Exposes mismatch want %v got %v: ", wantNumExposes, gotNumExposes)
	}

	// sort exposekeys in same sequence as results.
	exposekeys := make([]string, 0, len(device.Exposes))
	for k := range device.Exposes {
		exposekeys = append(exposekeys, k)
	}

	sort.Strings(exposekeys)

	idx := 0
	for _, key := range exposekeys {

		expose := device.Exposes[key]
		event := result.Expose[idx]

		if event.Name != expose.Name {
			t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, event.Name)
		}
		dataType, ok := kindFromString(event.Type)
		if !ok {
			t.Errorf("invalid event.type want %v got %v: ", event.Type, dataType)
		}

		for fId, foundTs := range event.Timestamp {
			tsFound := false
			dataIdx := 0
			foundTsUnix := foundTs.Unix()

			for insertIdx, insertTs := range timestamps {
				insertTsUnix := insertTs.Unix()

				if foundTsUnix == insertTsUnix {
					tsFound = true
					dataIdx = insertIdx
					break
				}
			}

			if !tsFound {
				t.Fatalf(fmt.Sprintf("Timestamp not found %v", foundTs))
			}
			var value any = 0
			if dataType == reflect.Float32 {
				v, _ := values.([]float32)
				value = v[dataIdx]

			} else if dataType == reflect.String {
				v, _ := values.([]string)
				value = v[dataIdx]

			} else if dataType == reflect.Int {
				v, _ := values.([]int)
				value = v[dataIdx]
			}

			wantKind := reflect.TypeOf(value).Kind()
			if dataType != wantKind {
				t.Fatalf(fmt.Sprintf("Type mismatch want: %v got: %v", wantKind.String(), dataType.String()))
			}

			if value != event.Values[fId] {
				t.Fatalf(fmt.Sprintf("Value mismatch want:%v got: %v", value, event.Values[fId]))
			}
			//fmt.Printf("found event %v with data: %v, timestamp: %v \n", expose.Name, event.Values[fId], foundTs)
		}

		idx++
	}
}
func assertDeviceEvents(device *devices.Device, result *metrics.DeviceMetricsResult, timestamps []*time.Time, values []float32, t *testing.T) {

	if result.DeviceId != device.Id {
		t.Errorf("deviceId mismatch want %v got %v: ", device.Id, result.DeviceId)
	}

	gotNumExposes := len(result.Expose)
	wantNumExposes := len(device.Exposes)
	if gotNumExposes != wantNumExposes {
		t.Errorf("Exposes mismatch want %v got %v: ", wantNumExposes, gotNumExposes)
	}

	// sort exposekeys in same sequence as results.
	exposekeys := make([]string, 0, len(device.Exposes))
	for k := range device.Exposes {
		exposekeys = append(exposekeys, k)
	}

	sort.Strings(exposekeys)

	idx := 0
	for _, key := range exposekeys {

		expose := device.Exposes[key]
		event := result.Expose[idx]

		if event.Name != expose.Name {
			t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, event.Name)
		}
		dataType, ok := kindFromString(event.Type)
		if !ok {
			t.Errorf("invalid event.type want %v got %v: ", event.Type, dataType)
		}

		for fId, foundTs := range event.Timestamp {
			tsFound := false
			dataIdx := 0
			foundTsUnix := foundTs.Unix()

			for insertIdx, insertTs := range timestamps {
				insertTsUnix := insertTs.Unix()

				if foundTsUnix == insertTsUnix {
					tsFound = true
					dataIdx = insertIdx
					break
				}
			}

			if !tsFound {
				t.Fatalf(fmt.Sprintf("Timestamp not found %v", foundTs))
			}

			wantKind := reflect.TypeOf(values[dataIdx]).Kind()
			if dataType != wantKind {
				t.Fatalf(fmt.Sprintf("Type mismatch want: %v got: %v", wantKind.String(), dataType.String()))
			}

			if values[dataIdx] != event.Values[fId] {
				t.Fatalf(fmt.Sprintf("Value mismatch want:%v got: %v", values[dataIdx], event.Values[fId]))
			}
			//fmt.Printf("found event %v with data: %v, timestamp: %v \n", expose.Name, event.Values[fId], foundTs)
		}

		idx++
	}
}
func createMockDevice(id string, name string, numOfExposes int, exposeType string, timestamp time.Time, data any) *devices.Device {
	device1 := &devices.Device{}
	device1.Id = id
	device1.FriendlyName = name
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", id)
	device1.PowerSource = "mains"
	device1.Properties = map[string]any{}
	device1.Properties["last_seen"] = timestamp.Format(time.RFC3339)
	device1.Properties["link_quality"] = 45.0
	device1.Exposes = make(map[string]*devices.Entity)

	for i := 0; i < numOfExposes; i++ {

		property := fmt.Sprintf("property_%v_%v", id, (i + 1))

		ent1 := &devices.Entity{}
		ent1.Description = fmt.Sprintf("%s readings", property)
		ent1.Name = property
		ent1.Unit = "test"
		ent1.Data = data
		ent1.Type = exposeType

		device1.Exposes[property] = ent1
		device1.Exposes[property].Attributes = make(map[string]any)
		device1.Exposes[property].Attributes["min"] = 0.0
		device1.Exposes[property].Attributes["max"] = 255.0
	}

	return device1
}

func CreateDateTimeTimestamps(numberOfDays int, numberOfHours int, numberOfMinutes int) []*time.Time {
	var timestamps []*time.Time
	year := time.Now().Year()
	month := time.Now().Month()
	today := time.Now().Day()
	hours := 0
	minutes := 0

	if numberOfMinutes <= 0 {
		numberOfMinutes = 1
	}

	currentDay := (today + 1) - numberOfDays

	for d := 1; d <= numberOfDays; d++ {
		for h := 0; h < numberOfHours; h++ {
			for m := 0; m < numberOfMinutes; m++ {
				timestamp := time.Date(year, month, currentDay, hours+h, minutes+m, 0, 0, time.UTC)
				timestamps = append(timestamps, &timestamp)
			}
		}
		currentDay++
	}

	return timestamps
}

func CreateFloatValues(numOfItems int) []float32 {
	var values []float32 = make([]float32, numOfItems)
	for i := 0; i < numOfItems; i++ {
		values[i] = floatrandom(10, 100)
	}
	return values
}

func CreateEnumValues(numOfItems int) []int {
	var values []int = make([]int, numOfItems)
	for i := 0; i < numOfItems; i++ {
		values[i] = intrandom(100)
	}
	return values
}

func CreateBinaryValues(numOfItems int) []string {
	var values []string = make([]string, numOfItems)
	for i := 0; i < numOfItems; i++ {
		val := intrandom(1)
		if val == 0 {
			values[i] = "on"
		} else {
			values[i] = "off"
		}
	}
	return values
}

func intrandom(max int) int {
	rand.Seed(time.Now().UnixNano())
	val := rand.Intn(max)

	return val
}

func floatrandom(value_1, value_2 float32) float32 {
	randomValue := value_1 + value_2 + rand.Float32()

	ratio := math.Pow(10, float64(1))
	return float32(math.Round(float64(randomValue)*ratio) / ratio)
}

func kindFromString(kindStr string) (reflect.Kind, bool) {
	kindStr = strings.ToLower(kindStr)
	switch kindStr {

	case "int", "int8", "int16", "int32", "int64":
		return reflect.Int, true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return reflect.Uint, true
	case "float32":
		return reflect.Float32, true
	case "float64":
		return reflect.Float64, true
	case "bool":
		return reflect.Bool, true
	case "string":
		return reflect.String, true
	default:
		return reflect.Invalid, false
	}
}
