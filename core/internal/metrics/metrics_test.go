package metrics_test

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"node-herder/internal/metrics"
	"node-herder/models/devices"
	"os"
	"testing"
	"time"
)

func TestDeviceTimeRangeMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := metrics.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}
	// testCases := []struct {
	// 	timestamps []*time.Time
	// 	values     []float32
	// }{
	// 	{timestamps: CreateDateTimeTimestamps(1, 24, 1), values: CreateValues(24)}, // 1 day, 24 hours, 1 min = 1 event per hour = 24 total
	// }

	timestamps := CreateDateTimeTimestamps(1, 24, 1)
	values := CreateValues(24)
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	numDevices := 1

	for i := 0; i < numDevices; i++ {

		deviceId := fmt.Sprintf("x000%v", i)
		deviceName := fmt.Sprintf("device %v", i)
		property := fmt.Sprintf("property_%v", deviceId)

		fmt.Println("total timestamps: ", len(timestamps))
		for tIdx, timestamp := range timestamps {

			dev := createMockDevice(deviceId, deviceName, property, *timestamp, values[tIdx])

			err = repo.Store(dev)
			fmt.Printf("Add device: %v, data: %v, timestamp: %v \n", deviceId, values[tIdx], dev.Properties["last_seen"])
			if err != nil {
				t.Error("failed to store metrics ", err.Error())
			}
			devices[dev.Id] = dev
		}
	}

	id1 := fmt.Sprintf("x000%v", 0)
	dev1 := devices[id1]
	now := time.Now()

	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 5, 0, 0, 0, time.UTC)

	result, err := repo.ViewDeviceTimeRange(dev1, from, to)
	if err != nil {
		t.Error("failed to query metrics: ", err.Error())
	}
	if result.DeviceId != id1 {
		t.Errorf("deviceId mismatch want %v got %v: ", id1, result.DeviceId)
	}

	gotNumExposes := len(result.Expose)
	wantNumExposes := 1
	if gotNumExposes != wantNumExposes {
		t.Errorf("num of exposes mismatch want %v got %v: ", wantNumExposes, gotNumExposes)
	}

	idx := 0
	for _, expose := range dev1.Exposes {

		events := result.Expose[idx]
		gotNumEvents := len(events.Values)
		wantNumEvents := 6

		if gotNumEvents != wantNumEvents {
			t.Errorf("num of events mismatch want %v got %v: ", wantNumEvents, gotNumExposes)
		}

		if events.Name != expose.Name {
			t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, events.Name)
		}

		/// -----------------------------------------

		for fId, foundTs := range events.Timestamp {
			tsFound := false
			dataIdx := 0
			foundTsUnix := foundTs.Unix()

			for _, insertTs := range timestamps {
				insertTsUnix := insertTs.Unix()

				if foundTsUnix == insertTsUnix {
					tsFound = true
					dataIdx = idx
					break
				}
			}

			if !tsFound {
				t.Fatalf(fmt.Sprintf("Timestamp not found %v", foundTs))
			}

			if values[dataIdx] != events.Values[fId] {
				t.Fatalf(fmt.Sprintf("Value mismatch want:%v got: %v", values[dataIdx], events.Values[fId]))
			}
		}

		// ---------------------------------------
		// TODO -
		// verify values
		// make assert function to reuse

		idx++
	}

	//property1 := fmt.Sprintf("property_%v", id1)
	// err = repo.ViewExposeTimeRange(id1, property1, time.Now().Add(-(time.Minute * 1)), time.Now())
	// if err != nil {
	// 	t.Error("failed to query metrics: ", err.Error())
	// }
}

func createMockDevice(id string, name string, property string, timestamp time.Time, data any) *devices.Device {
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

	ent1 := &devices.Entity{}
	ent1.Description = fmt.Sprintf("%s readings", property)
	ent1.Name = property
	ent1.Unit = "test"
	ent1.Data = data

	device1.Exposes[property] = ent1
	device1.Exposes[property].Attributes = make(map[string]any)
	device1.Exposes[property].Attributes["min"] = 0.0
	device1.Exposes[property].Attributes["max"] = 255.0
	return device1
}

func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

func createDayTimestamps(numberOfDays int, numberOfEvents int) []*time.Time {

	var eventDays []*time.Time
	now := time.Now()
	year, month, today := now.Date()
	currentDay := (today + 1) - numberOfDays

	for day := 1; day <= numberOfDays; day++ {
		hour := 0
		for i := 0; i < numberOfEvents; i++ {
			timestamp := time.Date(year, month, currentDay, hour+i, 0, 0, 0, time.UTC)
			eventDays = append(eventDays, &timestamp)
		}

		currentDay++
	}

	return eventDays
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

func CreateValues(numOfItems int) []float32 {
	var values []float32 = make([]float32, numOfItems)
	for i := 0; i < numOfItems; i++ {
		values[i] = floatrandom(10, 100)
	}
	return values
}

func floatrandom(value_1, value_2 float32) float32 {
	randomValue := value_1 + value_2 + rand.Float32()

	ratio := math.Pow(10, float64(1))
	return float32(math.Round(float64(randomValue)*ratio) / ratio)
}
