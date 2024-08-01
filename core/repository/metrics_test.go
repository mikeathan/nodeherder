package repository_test

import (
	"encoding/json"
	"fmt"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestGenerateMockMetrics(t *testing.T) {
	t.Skip("NOTE: used for generating mock data")
	now := time.Now()

	tempfile := tempfile()
	defer os.Remove(tempfile)

	id := "0xa4c13894070052fc"
	timestamps := utils_test.CreateDateTimeTimestamps(30, 24, 10)
	values := utils_test.CreateBinaryValues(len(timestamps))
	from := time.Date(now.Year(), now.Month(), now.Day()-40, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	deviceName := fmt.Sprintf("device %v", 0)

	fmt.Println("total timestamps: ", len(timestamps))
	dev := createMockDevice(id, deviceName, 1, "binary", time.Now(), nil)
	for tIdx, timestamp := range timestamps {

		dev = createMockDevice(id, deviceName, 1, "binary", timestamp, values[tIdx])
		payload := utils_test.Payload(dev)

		mockClock.SetMockTime(timestamp)

		err = repo.Store(dev.Id, payload)
		if err != nil {
			t.Error("failed to store metrics ", err.Error())
		}
		devices[dev.Id] = dev
	}

	result, err := repo.ViewDeviceTimeRange(dev, from, to)
	if err != nil {
		t.Error("failed to query metrics: ", err.Error())
	}

	//assertDeviceEvents(dev, result, timestamps, values, t)
	bytes, _ := json.Marshal(result)
	fmt.Println(string(bytes))
}

func TestMultipleDeviceTimeRangeMetrics(t *testing.T) {
	now := time.Now()

	tempfile := tempfile()
	defer os.Remove(tempfile)

	testCases := []struct {
		id         string
		timestamps []time.Time
		values     []float32
		from       time.Time
		to         time.Time
	}{
		{id: "x0000",
			timestamps: utils_test.CreateDateTimeTimestamps(1, 24, 1), // 24 events
			values:     utils_test.CreateFloatValues(24),
			from:       time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC),
			to:         time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)},
	}

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	for i, test := range testCases {
		fmt.Println(len(test.timestamps))

		deviceName := fmt.Sprintf("device %v", i)

		fmt.Println("total timestamps: ", len(test.timestamps))
		for tIdx, timestamp := range test.timestamps {

			dev := createMockDevice(test.id, deviceName, 2, "numeric", timestamp, test.values[tIdx])
			payload := utils_test.Payload(dev)

			mockClock.SetMockTime(timestamp)

			err = repo.Store(dev.Id, payload)
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

		assertDeviceAnyDataTypeEvents(dev, result, test.timestamps, test.values, t)
	}
}

func TestDeviceTimeRangeMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
	values := utils_test.CreateFloatValues(24)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	numDevices := 3

	for i := 0; i < numDevices; i++ {

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		fmt.Println("total timestamps: ", len(timestamps))
		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, 2, "numeric", timestamp, values[tIdx])

			payload := utils_test.Payload(dev)
			mockClock.SetMockTime(timestamp)

			err = repo.Store(dev.Id, payload)
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

		assertDeviceAnyDataTypeEvents(dev, result, timestamps, values, t)
	}
}

// func TestMetricsRateLimiter(t *testing.T) {

// 	tempfile := tempfile()
// 	defer os.Remove(tempfile)

// 	repo, err := repository.NewMetricsRepoFromFile(tempfile)
// 	if err != nil {
// 		t.Error("failed to initialise metrics repo", err.Error())
// 	}

// 	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
// 	values := utils_test.CreateFloatValues(24)
// 	var devices map[string]*devices.Device = make(map[string]*devices.Device)
// 	numDevices := 1

// 	for i := 0; i < numDevices; i++ {

// 		deviceId := fmt.Sprintf("x000%v", i)
// 		deviceName := fmt.Sprintf("device %v", i)

// 		fmt.Println("total timestamps: ", len(timestamps))
// 		for tIdx, timestamp := range timestamps {

// 			dev := createMockDevice(deviceId, deviceName, 2, "numeric", *timestamp, values[tIdx])

// 			err = repo.Store(dev)
// 			//fmt.Printf("Add device: %v, data: %v, timestamp: %v \n", deviceId, values[tIdx], dev.Properties["last_seen"])
// 			if err != nil {
// 				t.Error("failed to store metrics ", err.Error())
// 			}
// 			devices[dev.Id] = dev
// 		}
// 	}

// 	// query and assert
// 	for i := 0; i < numDevices; i++ {
// 		deviceId := fmt.Sprintf("x000%v", i)

// 		dev := devices[deviceId]
// 		now := time.Now()

// 		from := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
// 		to := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

// 		result, err := repo.ViewDeviceTimeRange(dev, from, to)
// 		if err != nil {
// 			t.Error("failed to query metrics: ", err.Error())
// 		}

//			for _, expose := range result.Expose {
//				if len(expose.Values) > 1 {
//					t.Errorf("rate limiter registered more than expected device hits want 1 got %v: ", len((expose.Values)))
//				}
//			}
//		}
//	}
func TestDeviceTimeRangeBinaryDataMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}
	now := time.Now()

	testCases := []struct {
		numEvents   []int
		from        time.Time
		to          time.Time
		wantResults int
	}{
		{numEvents: []int{1, 7, 1},
			from:        time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			to:          time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC),
			wantResults: 6},
		{numEvents: []int{1, 24, 10},
			from:        time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			to:          time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC),
			wantResults: 6 * 10},
		{numEvents: []int{1, 24, 2},
			from:        time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC),
			to:          time.Date(now.Year(), now.Month(), now.Day(), 80, 0, 0, 0, time.UTC),
			wantResults: 89},
	}

	for _, testCase := range testCases {
		eventsPerDay := testCase.numEvents[0]
		eventPerHour := testCase.numEvents[1]
		eventsPerMin := testCase.numEvents[2]
		timestamps := utils_test.CreateDateTimeTimestamps(eventsPerDay, eventPerHour, eventsPerMin)
		values := utils_test.CreateBinaryValues(eventsPerDay * eventPerHour * eventsPerMin)

		var devices map[string]*devices.Device = make(map[string]*devices.Device)

		deviceId := fmt.Sprintf("x000%v", 0)
		deviceName := fmt.Sprintf("device %v", 0)

		for tIdx, timestamp := range timestamps {
			value := values[tIdx]
			dev := createMockDevice(deviceId, deviceName, 1, "binary", timestamp, value)
			payload := utils_test.Payload(dev)
			mockClock.SetMockTime(timestamp)
			err = repo.Store(dev.Id, payload)
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}

		dev := devices[deviceId]

		result, err := repo.ViewDeviceTimeRange(dev, testCase.from, testCase.to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		for _, event := range result.Expose {
			binaryEvent := metrics.ToBinaryExposeResults(event)

			if testCase.from != time.UnixMilli(binaryEvent.From).UTC() {
				t.Errorf("from time mismatch want %v got %v", testCase.from.UnixMilli(), binaryEvent.From)
			}
			if testCase.to != time.UnixMilli(binaryEvent.To).UTC() {
				t.Errorf("to time mismatch want %v got %v", testCase.to.UnixMilli(), binaryEvent.To)
			}

			fmt.Printf("Name: %v \n", binaryEvent.Name)
			fmt.Printf("From: %v \n", time.UnixMilli(binaryEvent.From))
			fmt.Printf("To: %v \n", time.UnixMilli(binaryEvent.To))

			if len(binaryEvent.Data) != testCase.wantResults {
				t.Errorf("number of data points mismatch want %v got %v", testCase.wantResults, len(binaryEvent.Data))
			}
			for _, v := range binaryEvent.Data {
				fmt.Printf("value: %v \n", v.X)
				fmt.Printf("Time range : %v  - %v \n", time.UnixMilli(v.Y[0]).UTC(), time.UnixMilli(v.Y[1]).UTC())
			}
		}

	}

	// TODO:
	// add a test case for when we have repeated events
	//values := []string{"on", "off", "on", "off", "on", "on", "off"}

}

func TestDeviceTimeRangeDataTypesMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
	data := map[string]any{

		"numeric": utils_test.CreateFloatValues(24),
		"binary":  utils_test.CreateBinaryValues(24),
	}

	// "enum":    utils_test.CreateEnumValues(24), NOT SUPPORTED YET

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

			// used dev.props['last_seen] previously but now using time.now in metrics.Store
			// so i cant test timestamps
			dev := createMockDevice(deviceId, deviceName, 2, dataType, timestamp, value)
			payload := utils_test.Payload(dev)
			mockClock.SetMockTime(timestamp)
			err = repo.Store(dev.Id, payload)
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
	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
	values := utils_test.CreateFloatValues(24)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	numDevices := 3

	for i := 0; i < numDevices; i++ {

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, 5, "numeric", timestamp, values[tIdx])
			payload := utils_test.Payload(dev)
			mockClock.SetMockTime(timestamp)
			err = repo.Store(dev.Id, payload)
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

		exposeName := fmt.Sprintf("property_%v_%v", deviceId, 1)

		result, err := repo.ViewExposeTimeRange(dev, exposeName, from, to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		assertDeviceExportAnyDataTypeEvents(dev, exposeName, result, timestamps, values, t)
	}
}

func TestMultipleExposeTimeRangeMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, err := repository.NewMetricsRepoFromFile(tempfile, mockClock)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}

	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
	values := utils_test.CreateFloatValues(24)
	numOfExposes := 10
	deviceIds := []string{"x0000"}
	exposes := []string{"property_x0000_1", "property_x0000_2", "property_x0000_4"}
	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	for _, deviceId := range deviceIds {

		deviceName := fmt.Sprintf("device %v", deviceId)

		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, numOfExposes, "numeric", timestamp, values[tIdx])

			payload := utils_test.Payload(dev)
			mockClock.SetMockTime(timestamp)

			err = repo.Store(dev.Id, payload)
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

			assertDeviceExportAnyDataTypeEvents(dev, exposeName, result, timestamps, values, t)
		}
	}
}

func assertDeviceExportAnyDataTypeEvents(device *devices.Device, exposeName string, result *metrics.DeviceMetricsResult, timestamps []time.Time, values any, t *testing.T) {

	if result.DeviceId != device.Id {
		t.Errorf("deviceId mismatch want %v got %v: ", device.Id, result.DeviceId)
	}
	expose := device.Exposes[exposeName]
	for _, event := range result.Expose {
		if event.GetType() == "numeric" {
			assertNumericExposeEvent(expose, event, timestamps, values.([]float32), t)

		} else if event.GetType() == "binary" {
			assertBinaryExposeEvent(expose, event, timestamps, values.([]string), t)

		} else {
			t.Errorf("invalid expose type %v: ", event.GetType())
		}
	}

}
func assertDeviceAnyDataTypeEvents(device *devices.Device, result *metrics.DeviceMetricsResult, timestamps []time.Time, values any, t *testing.T) {

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

		if event.GetType() == "numeric" {
			assertNumericExposeEvent(expose, event, timestamps, values.([]float32), t)
		} else if event.GetType() == "binary" {
			assertBinaryExposeEvent(expose, event, timestamps, values.([]string), t)
		} else {
			t.Errorf("invalid expose type %v: ", event.GetType())
		}
		idx++ // ?????

	}
}
func assertNumericExposeEvent(expose *devices.Entity, event metrics.ExposeMetricsResult, timestamps []time.Time, values []float32, t *testing.T) {
	numericEvent := metrics.ToNumericExposeResults(event)

	if numericEvent == nil {
		t.Errorf("invalid expose type want numeric got %v: ", event.GetType())
	}

	if numericEvent.Name != expose.Name {
		t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, numericEvent.Name)
	}

	for _, numericData := range numericEvent.Data {
		tsFound := false
		dataIdx := 0

		eventTimestamp := numericData.X
		eventValue := numericData.Y
		for insertIdx, insertTs := range timestamps {
			insertTsUnix := insertTs.UnixMilli()

			if eventTimestamp == insertTsUnix {
				tsFound = true
				dataIdx = insertIdx
				break
			}
		}
		if !tsFound {
			t.Fatalf(fmt.Sprintf("Timestamp not found %v", eventTimestamp))
		}

		wantValue := values[dataIdx]
		wantKind := reflect.Float32
		gotKind := reflect.TypeOf(eventValue).Kind()
		if gotKind != wantKind {
			t.Fatalf(fmt.Sprintf("Type mismatch want: %v got: %v", wantKind.String(), gotKind.String()))
		}

		if wantValue != eventValue {
			t.Fatalf(fmt.Sprintf("Value mismatch want:%v got: %v", wantValue, eventValue))
		}
		//fmt.Printf("found event %v with data: %v, timestamp: %v \n", expose.Name, event.Values[fId], foundTs)
	}

}

func assertBinaryExposeEvent(expose *devices.Entity, event metrics.ExposeMetricsResult, timestamps []time.Time, values []string, t *testing.T) {
	binaryEvent := metrics.ToBinaryExposeResults(event)

	if binaryEvent == nil {
		t.Errorf("invalid expos type want binary got %v: ", event.GetType())
	}

	if binaryEvent.Name != expose.Name {
		t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, binaryEvent.Name)
	}

	// on = 1
	// off = 2
	// on = 3
	// off = 4

	// on = [1,2]
	// off = [2,3]
	// on = [3,4]
	for _, binaryData := range binaryEvent.Data {
		tsFound := false
		dataIdx := 0

		eventValue := binaryData.X
		eventTimestamps := binaryData.Y

		eventStart := eventTimestamps[0]
		eventEnd := eventTimestamps[1]

		for insertIdx, insertTs := range timestamps {
			insertTsUnix := insertTs.UnixMilli()

			if eventStart == insertTsUnix {
				tsFound = true
				dataIdx = insertIdx
				break
			}
		}

		if !tsFound {
			t.Fatalf(fmt.Sprintf("Start timestamp not found %v", eventStart))
		}

		// we are expecting the end timestamp to be the next one
		if eventEnd != timestamps[dataIdx+1].UnixMilli() {
			t.Fatalf(fmt.Sprintf("End timestamp not matching %v", eventEnd))
		}

		// NOTE: not sure if dataindex is correct here . could be + or -1
		// NEEDS TESTING
		wantValue := values[dataIdx] // !!!!!!!
		wantKind := reflect.String
		gotKind := reflect.TypeOf(eventValue).Kind()
		if gotKind != wantKind {
			t.Fatalf(fmt.Sprintf("Type mismatch want: %v got: %v", wantKind.String(), gotKind.String()))
		}

		if wantValue != eventValue {
			t.Fatalf(fmt.Sprintf("Value mismatch want:%v got: %v", wantValue, eventValue))
		}
		//fmt.Printf("found event %v with data: %v, timestamp: %v \n", expose.Name, event.Values[fId], foundTs)
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
