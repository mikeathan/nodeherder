package repository_test

import (
	"encoding/json"
	"fmt"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	utils_test "node-herder/testing"
	"os"
	"testing"
	"time"
)

func TestGenerateMockMetrics(t *testing.T) {
	t.Skip("NOTE: used for generating mock data")
	now := time.Now()

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	id := "0xa4c13894070052fc"
	timestamps := utils_test.CreateDateTimeTimestamps(2, 24, 1)
	values := utils_test.CreateBinaryValues(len(timestamps))
	from := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
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

func TestNewVersionSingleExposeValueUpdatesDeviceTimeRangeMetrics(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
	}
	now := time.Now()

	timestamps := utils_test.CreateFullDateTimeTimestamps(1, 24, 1)
	values := utils_test.CreateFloatValues(24)
	from := time.Date(now.Year(), now.Month(), now.Day()-2, 15, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 30, 0, 0, time.UTC)

	exposeNames := []string{"temperature", "humidity", "presence", "power", "voltage", "current"}

	id := "x0001"
	deviceName := "device_1"
	dev := createMockDeviceWithExposes(id, deviceName, exposeNames, "numeric", now, nil)

	exposeIdx := 0

	// fire data changes one by one
	for tIdx, timestamp := range timestamps {

		exposeName := exposeNames[exposeIdx]
		exposeIdx++

		payload := map[string]any{exposeName: values[tIdx]}
		mockClock.SetMockTime(timestamp)

		if exposeIdx >= len(exposeNames) {
			exposeIdx = 0
		}

		err = repo.Store(dev.Id, payload)
		if err != nil {
			t.Error("failed to store metrics ", err.Error())
		}
	}

	result, err := repo.ViewDeviceTimeRange(dev, from, to)
	if err != nil {
		t.Error("failed to query metrics: ", err.Error())
	}

	utils_test.AssertDeviceAnyDataTypeEvents(dev, result, timestamps, values, t)
}
func TestSingleExposeValueUpdatesDeviceTimeRangeMetrics(t *testing.T) {

	now := time.Now()

	tempfile := utils_test.Tempfile()

	defer os.Remove(tempfile)

	id := "x0000"
	timestamps := utils_test.CreateFullDateTimeTimestamps(1, 24, 1)
	values := utils_test.CreateFloatValues(24)
	from := time.Date(now.Year(), now.Month(), now.Day()-2, 15, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 30, 0, 0, time.UTC)

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
	}

	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	deviceName := fmt.Sprintf("device %v", 1)

	fmt.Println("total timestamps: ", len(timestamps))

	exposeNames := []string{"temperature", "humidity", "presence", "power", "voltage", "current"}

	dev := createMockDeviceWithExposes(id, deviceName, exposeNames, "numeric", now, nil)

	exposeIdx := 0

	// fire data changes one by one
	for tIdx, timestamp := range timestamps {

		exposeName := exposeNames[exposeIdx]
		exposeIdx++

		payload := map[string]any{exposeName: values[tIdx]}
		mockClock.SetMockTime(timestamp)

		if exposeIdx >= len(exposeNames) {
			exposeIdx = 0
		}
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

	utils_test.AssertDeviceAnyDataTypeEvents(dev, result, timestamps, values, t)

}

func TestMultipleDeviceTimeRangeMetrics(t *testing.T) {

	now := time.Now()
	tempfile := utils_test.Tempfile()

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

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
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

		utils_test.AssertDeviceAnyDataTypeEvents(dev, result, test.timestamps, test.values, t)
	}
}

func TestMetricsPruning(t *testing.T) {
	now := time.Now()

	tempfile := utils_test.Tempfile()
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

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
	}

	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	for i, test := range testCases {

		deviceName := fmt.Sprintf("device %v", i)
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

	// assert data has been pruned
	time.Sleep(time.Millisecond * 500)

	repo.Prune(time.Now().Add(-time.Hour * 24))

	time.Sleep(time.Minute * 500)

}
func TestDeviceTimeRangeMetrics(t *testing.T) {

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
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

		utils_test.AssertDeviceAnyDataTypeEvents(dev, result, timestamps, values, t)
	}
}

// func TestMetricsRateLimiter(t *testing.T) {

// 	tempfile := utils_test.Tempfile()
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

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
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
			wantResults: 90},
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

		for _, event := range result.Exposes {
			binaryEvent := metrics.ToTimeRangeExposeResults(event)

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

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
	}

	testCases := []struct {
		dataType string
		data     any
	}{
		{dataType: "numeric", data: utils_test.CreateFloatValues(24)},
		{dataType: "binary", data: utils_test.CreateBinaryValues(24)},
		{dataType: "binary", data: utils_test.CreateBinaryBooleanValues(24)},
		{dataType: "enum", data: utils_test.CreateEnumValues(24)},
	}

	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)

	for i, testCase := range testCases {
		values := testCase.data
		dataType := testCase.dataType

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)

		for tIdx, timestamp := range timestamps {
			var value any = 0
			if dataType == "numeric" {
				v, _ := values.([]float32)
				value = v[tIdx]

			} else if dataType == "binary" {
				if v, ok := values.([]string); ok {
					value = v[tIdx]
				} else if v, ok := values.([]bool); ok {
					value = v[tIdx]
				} else {
					t.Error("failed to create data type")
				}
			} else if dataType == "binaryBool" {
				v, _ := values.([]bool)
				value = v[tIdx]

			} else if dataType == "enum" {
				v, _ := values.([]string)
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
	for i, testCase := range testCases {
		values := testCase.data
		deviceId := fmt.Sprintf("x000%v", i)

		dev := devices[deviceId]
		now := time.Now()

		from := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)

		result, err := repo.ViewDeviceTimeRange(dev, from, to)
		if err != nil {
			t.Error("failed to query metrics: ", err.Error())
		}

		utils_test.AssertDeviceAnyDataTypeEvents(dev, result, timestamps, values, t)
	}
}

func createMockDeviceWithExposes(id string, name string, exposeNames []string, exposeType string, timestamp time.Time, data any) *devices.Device {
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

	for i := 0; i < len(exposeNames); i++ {

		property := exposeNames[i]

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
