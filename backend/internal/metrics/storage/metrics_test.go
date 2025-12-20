package storage_test

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	metricsstorage "node-herder/internal/metrics/storage"
	"node-herder/mocks"
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"node-herder/utils"
	utilsstorage "node-herder/utils/storage"
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
	values := utils_test.CreateBinaryBooleanValues(len(timestamps))
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

func TestViewDeviceTimeRangeSkipsUnsupportedExposes(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, mockClock, err := utils_test.CreateMetricsRepo(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo: ", err.Error())
	}

	now := time.Now().UTC()
	from := now.Add(-1 * time.Minute)
	to := now.Add(1 * time.Minute)

	dev := createMockDeviceWithExposes("dev_unknown", "device_unknown", []string{"power"}, "numeric", now, nil)
	unsupported := devices.NewEntity("socket_state")
	unsupported.Type = "unsupported_type"
	unsupported.Data.SetValue(map[string]any{"state": "ON"})
	dev.Exposes["socket_state"] = unsupported

	mockClock.SetMockTime(now)
	err = repo.Store(dev.Id, map[string]any{
		"power":        float64(12.3),
		"socket_state": map[string]any{"state": "ON"},
	})
	if err != nil {
		t.Fatalf("failed to store metrics: %v", err)
	}

	result, err := repo.ViewDeviceTimeRange(dev, from, to)
	if err != nil {
		t.Fatalf("failed to query metrics: %v", err)
	}
	if len(result.Exposes) != 1 {
		t.Fatalf("expected 1 expose, got %d", len(result.Exposes))
	}
	if result.Exposes[0].GetType() != "numeric" {
		t.Fatalf("expected numeric expose, got %s", result.Exposes[0].GetType())
	}
	assertExposeDataCount(t, result.Exposes[0], 1)
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

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	testCases := []struct {
		id         string
		timestamps []time.Time
		values     []float32
	}{
		{id: "x0000",
			timestamps: utils_test.CreateDateTimeTimestamps(2, 24, 1),
			values:     utils_test.CreateFloatValues(48),
		},
		{id: "x0001",
			timestamps: utils_test.CreateDateTimeTimestamps(10, 5, 1),
			values:     utils_test.CreateFloatValues(50),
		},
		{id: "x0002",
			timestamps: utils_test.CreateDateTimeTimestamps(20, 2, 1),
			values:     utils_test.CreateFloatValues(40),
		},
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

	time.Sleep(time.Millisecond * 500)

	// reset clock. we reset it earlier so we can store the mocked timestamp in the repo
	// reset back with now() time for pruning to work
	now := time.Now().UTC()
	mockClock.SetMockTime(now)

	// assert data has been pruned
	repo.Prune(time.Hour * 24)

	time.Sleep(time.Millisecond * 500)

	// assert deletion
	for _, test := range testCases {

		from := time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		dev := devices[test.id]
		result, err := repo.ViewDeviceTimeRange(dev, from, to)
		if err != nil {
			t.Errorf("failed to query metrics for device %v error:%v ", test.id, err.Error())
		}

		for _, expose := range result.Exposes {
			eventResult := domain.ToNumericExposeResults(expose)
			for _, event := range eventResult.Data {
				timestamp := time.UnixMilli(event.X).UTC()

				// assert for any events that are older than 24 hours
				if now.Sub(timestamp) > time.Hour*24 {
					t.Errorf("failed to prune device %v event", test.id)
				}
			}
		}
	}
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
			wantResults: 7}, // 0:00 through 6:00 inclusive = 7 events
		{numEvents: []int{1, 24, 10},
			from:        time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			to:          time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC),
			wantResults: 61}, // Events from 0:00 to 6:00 with 10 per hour
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
		values := utils_test.CreateBinaryBooleanValues(eventsPerDay * eventPerHour * eventsPerMin)

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
			binaryEvent := domain.ToExposeBinaryEventsResult(event)

			if binaryEvent == nil {
				t.Errorf("invalid expose type want binary got %v: ", event.GetType())
				continue
			}

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
				fmt.Printf("value: %v \n", v.Value)
				fmt.Printf("timestamp: %v \n", time.UnixMilli(v.Timestamp).UTC())
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
			switch dataType {
			case "numeric":
				v, _ := values.([]float32)
				value = v[tIdx]

			case "binary":
				if v, ok := values.([]string); ok {
					value = v[tIdx]
				} else if v, ok := values.([]bool); ok {
					value = v[tIdx]
				} else {
					t.Error("failed to create data type")
				}
			case "binaryBool":
				v, _ := values.([]bool)
				value = v[tIdx]

			case "enum":
				v, _ := values.([]string)
				value = v[tIdx]
			}

			// used dev.props['last_seen] previously but now using time.now in domain.Store
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

func TestQueryDeviceFilters(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	kvdb, err := utilsstorage.NewBoltKeyValueDatabase(tempfile, "metrics")
	if err != nil {
		t.Fatal(err)
	}

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})

	repoIface, err := metricsstorage.NewMetricsRepoFromDatabase(kvdb, mockClock, 0*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	repo, ok := repoIface.(*metricsstorage.MetricsRepo)
	if !ok {
		t.Fatal("expected MetricsRepo implementation")
	}

	base := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	timestamps := []time.Time{
		base.Add(1 * time.Minute),
		base.Add(2 * time.Minute),
		base.Add(3 * time.Minute),
		base.Add(4 * time.Minute),
	}
	from := base
	to := base.Add(10 * time.Minute)

	storeExposeValues(t, repo, mockClock, "dev-numeric", "temperature", timestamps, []any{
		float32(1), float32(5), float32(10), float32(15),
	})
	storeExposeValues(t, repo, mockClock, "dev-binary", "presence", timestamps, []any{
		true, false, true, false,
	})
	storeExposeValues(t, repo, mockClock, "dev-enum", "mode", timestamps, []any{
		"ON", "OFF", "ON", "OFF",
	})

	now := time.Date(2024, time.January, 3, 12, 0, 0, 0, time.UTC)
	lookbackTimestamps := []time.Time{
		now.Add(-49 * time.Hour),
		now.Add(-26 * time.Hour),
		now.Add(-2 * time.Hour),
		now.Add(-30 * time.Minute),
		now.Add(-5 * time.Minute),
	}
	storeExposeValues(t, repo, mockClock, "dev-lookback", "temperature", lookbackTimestamps, []any{
		float32(1), float32(2), float32(3), float32(4), float32(5),
	})

	testCases := []struct {
		name        string
		deviceID    string
		exposeName  string
		exposeType  string
		filters     []domain.MetricFilter
		lookback    string
		now         time.Time
		wantExposes int
		wantData    int
	}{
		{
			name:        "numeric/no filters",
			deviceID:    "dev-numeric",
			exposeName:  "temperature",
			exposeType:  "numeric",
			wantExposes: 1,
			wantData:    4,
		},
		{
			name:       "numeric/range filters",
			deviceID:   "dev-numeric",
			exposeName: "temperature",
			exposeType: "numeric",
			filters: []domain.MetricFilter{
				{Op: domain.OpGreaterThan, Value: float64(3)},
				{Op: domain.OpLessThan, Value: float64(12)},
			},
			wantExposes: 1,
			wantData:    2,
		},
		{
			name:       "numeric/equals filter",
			deviceID:   "dev-numeric",
			exposeName: "temperature",
			exposeType: "numeric",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: float64(5)},
			},
			wantExposes: 1,
			wantData:    1,
		},
		{
			name:       "numeric/not equals filter",
			deviceID:   "dev-numeric",
			exposeName: "temperature",
			exposeType: "numeric",
			filters: []domain.MetricFilter{
				{Op: domain.OpNotEquals, Value: float64(10)},
			},
			wantExposes: 1,
			wantData:    3,
		},
		{
			name:       "numeric/invalid filter type",
			deviceID:   "dev-numeric",
			exposeName: "temperature",
			exposeType: "numeric",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "5"},
			},
			wantExposes: 0,
			wantData:    0,
		},
		{
			name:       "binary/equals true",
			deviceID:   "dev-binary",
			exposeName: "presence",
			exposeType: "binary",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: true},
			},
			wantExposes: 1,
			wantData:    2,
		},
		{
			name:       "binary/not equals true",
			deviceID:   "dev-binary",
			exposeName: "presence",
			exposeType: "binary",
			filters: []domain.MetricFilter{
				{Op: domain.OpNotEquals, Value: true},
			},
			wantExposes: 1,
			wantData:    2,
		},
		{
			name:       "binary/invalid filter type",
			deviceID:   "dev-binary",
			exposeName: "presence",
			exposeType: "binary",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "true"},
			},
			wantExposes: 0,
			wantData:    0,
		},
		{
			name:       "enum/not equals missing",
			deviceID:   "dev-enum",
			exposeName: "mode",
			exposeType: "enum",
			filters: []domain.MetricFilter{
				{Op: domain.OpNotEquals, Value: "MISSING"},
			},
			wantExposes: 1,
			wantData:    4,
		},
		{
			name:       "enum/equals ON",
			deviceID:   "dev-enum",
			exposeName: "mode",
			exposeType: "enum",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "ON"},
			},
			wantExposes: 0,
			wantData:    0,
		},
		{
			name:       "enum/equals missing",
			deviceID:   "dev-enum",
			exposeName: "mode",
			exposeType: "enum",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "MISSING"},
			},
			wantExposes: 0,
			wantData:    0,
		},
		{
			name:        "lookback minutes",
			deviceID:    "dev-lookback",
			exposeName:  "temperature",
			exposeType:  "numeric",
			lookback:    "30m",
			now:         now,
			wantExposes: 1,
			wantData:    2,
		},
		{
			name:        "lookback hours",
			deviceID:    "dev-lookback",
			exposeName:  "temperature",
			exposeType:  "numeric",
			lookback:    "2h",
			now:         now,
			wantExposes: 1,
			wantData:    3,
		},
		{
			name:        "lookback days",
			deviceID:    "dev-lookback",
			exposeName:  "temperature",
			exposeType:  "numeric",
			lookback:    "2d",
			now:         now,
			wantExposes: 1,
			wantData:    4,
		},
		{
			name:        "lookback invalid falls back to range",
			deviceID:    "dev-lookback",
			exposeName:  "temperature",
			exposeType:  "numeric",
			lookback:    "bad",
			now:         now,
			wantExposes: 0,
			wantData:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			collectors := makeCollectors(t, tc.exposeName, tc.exposeType, from, to)

			queryFrom := from
			queryTo := to
			if tc.lookback != "" {
				queryFrom, queryTo = query.ResolveTime(domain.TimeQuery{Lookback: tc.lookback}, tc.now)
				collectors = makeCollectors(t, tc.exposeName, tc.exposeType, queryFrom, queryTo)
			}

			result, err := repo.QueryDevice(tc.deviceID, queryFrom, queryTo, tc.filters, collectors)
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}

			if len(result.Exposes) != tc.wantExposes {
				t.Fatalf("expected %d exposes, got %d", tc.wantExposes, len(result.Exposes))
			}

			if tc.wantExposes == 0 {
				return
			}

			assertExposeDataCount(t, result.Exposes[0], tc.wantData)
		})
	}
}

func createMockDeviceWithExposes(id string, name string, exposeNames []string, exposeType string, timestamp time.Time, data any) *devices.Device {
	device1 := &devices.Device{}
	device1.Id = id
	device1.FriendlyName = name
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", id)
	device1.PowerSource = "mains"
	device1.LastSeen = timestamp.Format(time.RFC3339)
	device1.Exposes = make(map[string]*devices.Entity)

	for i := 0; i < len(exposeNames); i++ {

		property := exposeNames[i]

		ent1 := devices.NewEntity(property)
		ent1.Description = fmt.Sprintf("%s readings", property)
		ent1.Name = property
		ent1.Unit = "test"
		ent1.Data.SetValue(data)
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
	device1.LastSeen = timestamp.Format(time.RFC3339)
	device1.Exposes = make(map[string]*devices.Entity)

	for i := 0; i < numOfExposes; i++ {

		property := fmt.Sprintf("property_%v_%v", id, (i + 1))

		ent1 := devices.NewEntity(property)
		ent1.Description = fmt.Sprintf("%s readings", property)
		ent1.Name = property
		ent1.Unit = "test"
		ent1.Data.SetValue(data)
		ent1.Type = exposeType

		device1.Exposes[property] = ent1
		device1.Exposes[property].Attributes = make(map[string]any)
		device1.Exposes[property].Attributes["min"] = 0.0
		device1.Exposes[property].Attributes["max"] = 255.0
	}

	return device1
}

func TestExposeBinaryEventsResultCollect(t *testing.T) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	result := domain.NewExposeBinaryEventsResult("state", from, to)

	testCases := []struct {
		timestamp time.Time
		value     string
	}{
		{timestamp: from.Add(time.Hour), value: "on"},
		{timestamp: from.Add(time.Hour * 2), value: "off"},
		{timestamp: from.Add(time.Hour * 3), value: "on"},
		{timestamp: from.Add(time.Hour * 4), value: "off"},
	}

	for _, tc := range testCases {
		boolVal, _ := utils.ConvertToBool(tc.value)
		valueBytes, err := utils.AnyToByteArray(boolVal)
		if err != nil {
			t.Fatalf("failed to marshal value: %v", err)
		}

		err = result.Collect(tc.timestamp, valueBytes)
		if err != nil {
			t.Fatalf("failed to collect data: %v", err)
		}
	}

	if len(result.Data) != len(testCases) {
		t.Errorf("expected %d data points, got %d", len(testCases), len(result.Data))
	}

	for i, tc := range testCases {
		if result.Data[i].Timestamp != tc.timestamp.UnixMilli() {
			t.Errorf("data[%d] timestamp mismatch: want %v, got %v",
				i, tc.timestamp.UnixMilli(), result.Data[i].Timestamp)
		}
		val, _ := utils.ConvertToBool(tc.value)
		if result.Data[i].Value != val {
			t.Errorf("data[%d] value mismatch: want %s, got %v",
				i, tc.value, result.Data[i].Value)
		}
	}
}

func TestExposeBinaryEventsResultJSON(t *testing.T) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	original := domain.NewExposeBinaryEventsResult("state", from, to)

	// Add some test data
	testData := []struct {
		timestamp time.Time
		value     string
	}{
		{timestamp: from.Add(time.Hour), value: "on"},
		{timestamp: from.Add(time.Hour * 2), value: "off"},
		{timestamp: from.Add(time.Hour * 3), value: "on"},
	}

	for _, td := range testData {
		boolVal, _ := utils.ConvertToBool(td.value)
		valueBytes, _ := utils.AnyToByteArray(boolVal)
		original.Collect(td.timestamp, valueBytes)
	}

	// Test marshaling
	_, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Test as part of DeviceMetricsResult
	deviceResult := domain.NewDeviceMetricsResult("test-device")
	deviceResult.Add(original)

	deviceJSON, err := json.Marshal(deviceResult)
	if err != nil {
		t.Fatalf("failed to marshal device result: %v", err)
	}

	// Test unmarshaling
	var unmarshaled domain.DeviceMetricsResult
	err = json.Unmarshal(deviceJSON, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.DeviceId != "test-device" {
		t.Errorf("device ID mismatch: want test-device, got %s", unmarshaled.DeviceId)
	}

	if len(unmarshaled.Exposes) != 1 {
		t.Fatalf("expected 1 expose, got %d", len(unmarshaled.Exposes))
	}

	binaryResult := domain.ToExposeBinaryEventsResult(unmarshaled.Exposes[0])
	if binaryResult == nil {
		t.Fatal("failed to convert to binary events result")
	}

	if binaryResult.Name != "state" {
		t.Errorf("name mismatch: want state, got %s", binaryResult.Name)
	}

	if binaryResult.Type != "binary" {
		t.Errorf("type mismatch: want binary, got %s", binaryResult.Type)
	}

	if binaryResult.From != from.UnixMilli() {
		t.Errorf("from mismatch: want %v, got %v", from.UnixMilli(), binaryResult.From)
	}

	if binaryResult.To != to.UnixMilli() {
		t.Errorf("to mismatch: want %v, got %v", to.UnixMilli(), binaryResult.To)
	}

	if len(binaryResult.Data) != len(testData) {
		t.Errorf("data length mismatch: want %d, got %d", len(testData), len(binaryResult.Data))
	}

	for i, td := range testData {
		if binaryResult.Data[i].Timestamp != td.timestamp.UnixMilli() {
			t.Errorf("data[%d] timestamp mismatch: want %v, got %v",
				i, td.timestamp.UnixMilli(), binaryResult.Data[i].Timestamp)
		}
		val, _ := utils.ConvertToBool(td.value)

		if binaryResult.Data[i].Value != val {
			t.Errorf("data[%d] value mismatch: want %s, got %v",
				i, td.value, binaryResult.Data[i].Value)
		}
	}
}

func TestNewExposeResultBinaryType(t *testing.T) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	result, err := domain.NewExposeResult("state", "binary", domain.AggNone, from, to)
	if err != nil {
		t.Fatalf("failed to create binary expose result: %v", err)
	}

	if result.GetType() != "binary" {
		t.Errorf("type mismatch: want binary, got %s", result.GetType())
	}

	binaryResult := domain.ToExposeBinaryEventsResult(result)
	if binaryResult == nil {
		t.Fatal("failed to convert to binary events result")
	}

	if binaryResult.Name != "state" {
		t.Errorf("name mismatch: want state, got %s", binaryResult.Name)
	}

	if binaryResult.From != from.UnixMilli() {
		t.Errorf("from mismatch: want %v, got %v", from.UnixMilli(), binaryResult.From)
	}

	if binaryResult.To != to.UnixMilli() {
		t.Errorf("to mismatch: want %v, got %v", to.UnixMilli(), binaryResult.To)
	}

	// Test that it properly collects data
	testValue := "on"
	boolVal, _ := utils.ConvertToBool(testValue)
	valueBytes, _ := utils.AnyToByteArray(boolVal)
	testTime := from.Add(time.Hour)

	err = result.Collect(testTime, valueBytes)
	if err != nil {
		t.Fatalf("failed to collect data: %v", err)
	}

	if binaryResult.Size() != 1 {
		t.Errorf("size mismatch: want 1, got %d", binaryResult.Size())
	}
}

func TestExposeBinaryEventsResultFlush(t *testing.T) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	result := domain.NewExposeBinaryEventsResult("state", from, to)

	// Add some test data
	testValue := "on"
	boolVal, _ := utils.ConvertToBool(testValue)
	valueBytes, _ := utils.AnyToByteArray(boolVal)
	testTime := from.Add(time.Hour)

	result.Collect(testTime, valueBytes)

	// Flush should do nothing for binary events (no state to finalize)
	result.Flush()

	// Verify data is still intact
	if len(result.Data) != 1 {
		t.Errorf("expected 1 data point after flush, got %d", len(result.Data))
	}

	if result.Data[0].Value != boolVal {
		t.Errorf("value changed after flush: want %s, got %v", testValue, result.Data[0].Value)
	}
}

func TestTailCacheReturnsRecentData(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	tailWindow := 2 * time.Minute
	repo, mockClock, err := utils_test.CreateMetricsRepoWithTailWindow(tempfile, tailWindow)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	// Build device
	dev := createMockDevice("dev1", "testdev", 1, "numeric", now, nil)

	ts1 := now.Add(-30 * time.Second)
	ts2 := now.Add(-10 * time.Second)

	mockClock.SetMockTime(ts1)
	repo.Store(dev.Id, map[string]any{"property_dev1_1": float32(10)})

	mockClock.SetMockTime(ts2)
	repo.Store(dev.Id, map[string]any{"property_dev1_1": float32(20)})

	// Query inside the tail window
	result, err := repo.ViewDeviceTimeRange(dev, now.Add(-1*time.Minute), now)
	if err != nil {
		t.Fatal(err)
	}

	events := result.Exposes[0]
	numeric := domain.ToNumericExposeResults(events)

	if len(numeric.Data) != 2 {
		t.Fatalf("expected 2 tail-cache entries, got %d", len(numeric.Data))
	}

	if numeric.Data[0].Y != 10 || numeric.Data[1].Y != 20 {
		t.Fatalf("unexpected values: %+v", numeric.Data)
	}
}

func TestTailCacheMixedWithDatabaseData(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	tailWindow := 1 * time.Minute
	repo, mockClock, err := utils_test.CreateMetricsRepoWithTailWindow(tempfile, tailWindow)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	dev := createMockDevice("dev1", "testdev", 1, "numeric", now, nil)

	// Bolt event (older than tail window)
	tsBolt := now.Add(-2 * time.Minute)
	mockClock.SetMockTime(tsBolt)
	repo.Store(dev.Id, map[string]any{"property_dev1_1": float32(10)})

	// Tail-cache event (recent)
	tsTail := now.Add(-10 * time.Second)
	mockClock.SetMockTime(tsTail)
	repo.Store(dev.Id, map[string]any{"property_dev1_1": float32(20)})

	result, err := repo.ViewDeviceTimeRange(dev, now.Add(-3*time.Minute), now)
	if err != nil {
		t.Fatal(err)
	}

	numeric := domain.ToNumericExposeResults(result.Exposes[0])

	if len(numeric.Data) != 2 {
		t.Fatalf("expected 2 mixed events (Bolt + tail), got %d", len(numeric.Data))
	}

	if numeric.Data[0].Y != 10 || numeric.Data[1].Y != 20 {
		t.Fatalf("unexpected merge order: %+v", numeric.Data)
	}
}

func TestTailCachePruning(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	tailWindow := 30 * time.Second
	repo, mockClock, err := utils_test.CreateMetricsRepoWithTailWindow(tempfile, tailWindow)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	dev := createMockDevice("dev1", "testdev", 1, "numeric", now, nil)

	// old entry (should be pruned)
	tsOld := now.Add(-1 * time.Minute)
	mockClock.SetMockTime(tsOld)
	repo.Store(dev.Id, map[string]any{"property_dev1_1": float32(10)})

	// new entry (inside window)
	tsNew := now.Add(-5 * time.Second)
	mockClock.SetMockTime(tsNew)
	repo.Store(dev.Id, map[string]any{"property_dev1_1": float32(20)})

	// Query entire range
	result, _ := repo.ViewDeviceTimeRange(dev, now.Add(-2*time.Minute), now)

	numeric := domain.ToNumericExposeResults(result.Exposes[0])

	if len(numeric.Data) != 2 {
		t.Fatalf("expected 2 results: one Bolt (old), one tail-cache (new); got %d", len(numeric.Data))
	}

	if numeric.Data[1].Y != 20 {
		t.Fatalf("new tail-cache event missing or incorrect: %+v", numeric.Data)
	}
}

func storeExposeValues(t *testing.T, repo *metricsstorage.MetricsRepo, mockClock *mocks.MockClock, deviceID string, exposeName string, timestamps []time.Time, values []any) {
	t.Helper()

	for i, timestamp := range timestamps {
		mockClock.SetMockTime(timestamp)
		if err := repo.Store(deviceID, map[string]any{exposeName: values[i]}); err != nil {
			t.Fatalf("failed to store metrics: %v", err)
		}
	}
}

func makeCollectors(t *testing.T, exposeName string, exposeType string, from time.Time, to time.Time) map[string]domain.ExposeResult {
	t.Helper()

	collector, err := domain.NewExposeResult(exposeName, exposeType, domain.AggNone, from, to)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}

	return map[string]domain.ExposeResult{
		exposeName: collector,
	}
}

func assertExposeDataCount(t *testing.T, expose domain.ExposeResult, want int) {
	t.Helper()

	if numeric := domain.ToNumericExposeResults(expose); numeric != nil {
		if len(numeric.Data) != want {
			t.Fatalf("expected %d numeric points, got %d", want, len(numeric.Data))
		}
		return
	}

	if binary := domain.ToExposeBinaryEventsResult(expose); binary != nil {
		if len(binary.Data) != want {
			t.Fatalf("expected %d binary points, got %d", want, len(binary.Data))
		}
		return
	}

	if enum := domain.ToTimeRangeExposeResults(expose); enum != nil {
		if len(enum.Data) != want {
			t.Fatalf("expected %d enum points, got %d", want, len(enum.Data))
		}
		return
	}

	t.Fatalf("unknown expose result type")
}
