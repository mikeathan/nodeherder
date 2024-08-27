package utils_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	"os"
	"reflect"
	"testing"
	"time"
)

type binaryValue struct {
	Value     string
	Timestamp time.Time
}

func CreateStore() store.AppStore {
	repo := repository.NewMemoryDeviceRepo()
	metricsRepo := mocks.NopMetricsRepo{}
	settingsRepo := mocks.NopSettingsrepo{}

	defer repo.Close()
	store, _ := store.NewAppStore(repo, &metricsRepo, &settingsRepo)
	return store
}

func CreateFileStore() (store.AppStore, func(), error) {

	settingsTempFile := tempfile()
	metricsTempFile := tempfile()
	repo := repository.NewMemoryDeviceRepo()

	cleanup := func() {
		os.Remove(settingsTempFile)
		os.Remove(metricsTempFile)
	}
	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now()
	})
	keyGenerator := metrics.NewTimestampedKeyGenerator(mockClock)
	metricsRepo, err := repository.NewMetricsRepoFromFile(metricsTempFile, keyGenerator)
	if err != nil {
		return nil, nil, err
	}
	settingsRepo, err := repository.NewFileSettingsRepoFromFile(settingsTempFile)
	if err != nil {
		return nil, nil, err
	}

	defer repo.Close()
	store, _ := store.NewAppStore(repo, metricsRepo, settingsRepo)
	return store, cleanup, nil
}

func CreateStoreFromDeviceRepo(repo devices.Repository) store.AppStore {
	metricsRepo := mocks.NopMetricsRepo{}
	settingsRepo := mocks.NopSettingsrepo{}
	store, _ := store.NewAppStore(repo, &metricsRepo, &settingsRepo)
	return store
}

func CreateStoreFromRepos(deviceRepo devices.Repository, metricsRepo metrics.Repository, settingsRepo settings.Repository) store.AppStore {
	store, _ := store.NewAppStore(deviceRepo, metricsRepo, settingsRepo)
	return store
}

func ValidateDevice(t *testing.T, dev1 *devices.Device, dev2 *devices.Device) {

	if dev1.Id != dev2.Id {
		t.Fatalf("device Id mismatch")
	}
	if dev1.FriendlyName != dev2.FriendlyName {
		t.Fatalf("device FriendlyName mismatch")
	}
	if dev1.Description != dev2.Description {
		t.Fatalf("device Description mismatch")
	}
	if dev1.ConnectionType != dev2.ConnectionType {
		t.Fatalf("device ConnectionType mismatch")
	}
	if dev1.PowerSource != dev2.PowerSource {
		t.Fatalf("device PowerSource mismatch")
	}

	for eidx, expose := range dev1.Exposes {
		inputExpose := dev2.Exposes[eidx]

		if expose.Name != inputExpose.Name {
			t.Fatalf("unexpected expose.Name value")
		}
		if expose.Description != inputExpose.Description {
			t.Fatalf("unexpected expose.Description value")
		}

		if !EqualityCheck(expose.Data, inputExpose.Data) {
			t.Fatalf("unexpected expose.Data value")
		}
		if expose.Unit != inputExpose.Unit {
			t.Fatalf("unexpected expose.Unit value")
		}
		for pidx, property := range expose.Properties {
			inputproperty := inputExpose.Properties[pidx]
			if property != inputproperty {
				t.Fatalf("unexpected property value")
			}

		}
	}
}

func ValidateBridge(t *testing.T, dev1 *devices.BridgeInfo, dev2 *devices.BridgeInfo) {

	if dev1.IeeeAddress != dev2.IeeeAddress {
		t.Fatalf("device IeeeAddress mismatch")
	}
	if dev1.FriendlyName != dev2.FriendlyName {
		t.Fatalf("device FriendlyName mismatch")
	}
	if dev1.DateCode != dev2.DateCode {
		t.Fatalf("device DateCode mismatch")
	}

	if dev1.Manufacturer != dev2.Manufacturer {
		t.Fatalf("device Manufacturer mismatch")
	}

	if dev1.Type != dev2.Type {
		t.Fatalf("device Type mismatch")
	}
	if dev1.SoftwareBuildID != dev2.SoftwareBuildID {
		t.Fatalf("device SoftwareBuildID mismatch")
	}
	if dev1.PowerSource != dev2.PowerSource {
		t.Fatalf("device PowerSource mismatch")
	}
	if dev1.Definition.Description != dev2.Definition.Description {
		t.Fatalf("device Definition.Description mismatch")
	}
	if dev1.Definition.Model != dev2.Definition.Model {
		t.Fatalf("device Definition.Model mismatch")
	}
	if dev1.Definition.Vendor != dev2.Definition.Vendor {
		t.Fatalf("device Definition.Vendor mismatch")
	}
	if dev1.Definition.SupportsOta != dev2.Definition.SupportsOta {
		t.Fatalf("device Definition.SupportsOta mismatch")
	}

	if !reflect.DeepEqual(dev1.Definition.Options, dev2.Definition.Options) {
		t.Fatalf("device Definition.Options mismatch")
	}

	for eidx, expose := range dev1.Definition.Exposes {
		expose2 := dev2.Definition.Exposes[eidx]
		if expose.Name != expose2.Name {
			t.Fatalf("unexpected expose.Name value")
		}

		if expose.Description != expose2.Description {
			t.Fatalf("unexpected expose.Description value")
		}

		if expose.Property != expose2.Property {
			t.Fatalf("unexpected expose.Property value")
		}
		if expose.Type != expose2.Type {
			t.Fatalf("unexpected expose.Type value")
		}
		if expose.Unit != expose2.Unit {
			t.Fatalf("unexpected expose.Unit value")
		}
		if expose.Access != expose2.Access {
			t.Fatalf("unexpected expose.Access value")
		}
		if expose.ValueMax != expose2.ValueMax {
			t.Fatalf("unexpected expose.ValueMax value")
		}
		if expose.ValueMin != expose2.ValueMin {
			t.Fatalf("unexpected expose.ValueMin value")
		}
		if expose.ValueOn != expose2.ValueOn {
			t.Fatalf("unexpected expose.ValueOn value")
		}
		if expose.ValueOff != expose2.ValueOff {
			t.Fatalf("unexpected expose.ValueOff value")
		}

		if !reflect.DeepEqual(expose.Values, expose2.Values) {
			t.Fatalf("unexpected expose.Values value")
		}
	}
}

const epsilon = 1e-9 // Adjust epsilon based on your desired precision

func compareIntFloat(i int, f float64) bool {
	return math.Abs(float64(i)-f) <= epsilon
}
func EqualityCheck(a interface{}, b interface{}) bool {

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	switch va.Kind() {
	case reflect.Float32, reflect.Float64:

		fl1 := math.Float32bits(float32(va.Float()))
		fl2 := math.Float32bits(float32(vb.Float()))

		return fl1 == fl2
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

		if va.Kind() != vb.Kind() {
			return compareIntFloat(a.(int), b.(float64))
		}
		return va.Int() == vb.Int()
	default:
		return a == b
	}
}

func CreateDevice(deviceId string, friendlyName string, property string, data any, min float64, max float64) *devices.Device {
	device1 := &devices.Device{}
	device1.Id = deviceId
	device1.FriendlyName = friendlyName
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", deviceId)
	device1.PowerSource = "mains"
	device1.Properties = map[string]any{}
	device1.Properties["last_seen"] = time.Now().Format(time.RFC3339)
	device1.Properties["link_quality"] = 45.0
	device1.Exposes = make(map[string]*devices.Entity)

	ent1 := &devices.Entity{}
	ent1.Description = fmt.Sprintf("%s readings", property)
	ent1.Name = property
	ent1.Unit = "test"
	ent1.Data = data

	device1.Exposes[property] = ent1
	device1.Exposes[property].Attributes = make(map[string]any)
	device1.Exposes[property].Attributes["min"] = min
	device1.Exposes[property].Attributes["max"] = max
	return device1
}

func CreateBridgeInfoList(deviceList []*devices.Device) []*devices.BridgeInfo {

	bridgeInfoList := []*devices.BridgeInfo{}
	for _, dev := range deviceList {
		bridge := &devices.BridgeInfo{}
		bridge.IeeeAddress = dev.Id
		bridge.Definition.Description = dev.Description
		bridge.FriendlyName = dev.FriendlyName
		bridge.Type = "EndDevice"
		bridge.PowerSource = dev.PowerSource
		bridge.Disabled = false
		bridge.InterviewCompleted = true

		e := devices.BridgeExpose{}
		e.Name = fmt.Sprintf("expose for %v", dev.FriendlyName)

		// NOTE: We make all features for now
		for _, expose := range dev.Exposes {
			f := devices.BridgeInfoFeature{}
			f.Name = expose.Name
			f.Property = expose.Name
			f.Type = expose.Type
			f.Unit = expose.Unit
			f.ValueMin = expose.Attributes["min"]
			f.ValueMax = expose.Attributes["max"]
			f.Description = expose.Description
			e.Features = append(e.Features, f)
		}

		bridge.Definition.Exposes = append(bridge.Definition.Exposes, e)
		bridgeInfoList = append(bridgeInfoList, bridge)
	}

	return bridgeInfoList
}

func Payload(device *devices.Device) map[string]any {
	payload := map[string]any{}
	for k, v := range device.Exposes {
		payload[k] = v.Data
	}
	return payload
}

func Tempfile() string {
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

func CreateDateTimeTimestamps(numberOfDays int, numberOfHours int, numberOfMinutes int) []time.Time {
	var timestamps []time.Time
	now := time.Now()
	year := now.Year()
	month := now.Month()
	today := now.Day()
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
				timestamps = append(timestamps, timestamp)
			}
		}
		currentDay++
	}

	return timestamps
}

func CreateFullDateTimeTimestamps(numberOfDays int, numberOfHours int, numberOfMinutes int) []time.Time {
	var timestamps []time.Time
	now := time.Now()
	year := now.Year()
	month := now.Month()
	today := now.Day()
	hours := 0
	minutes := 0

	if numberOfMinutes <= 0 {
		numberOfMinutes = 1
	}
	random := NewRanomValueGenerator(time.Now().UnixNano())
	currentDay := (today + 1) - numberOfDays

	for d := 1; d <= numberOfDays; d++ {
		for h := 0; h < numberOfHours; h++ {
			for m := 0; m < numberOfMinutes; m++ {
				seconds := random.GenerateRandomValue(0, 59)
				nanoseconds := random.GenerateRandomValue(0, 999999999)
				timestamp := time.Date(year, month, currentDay, hours+h, minutes+m, seconds, nanoseconds, time.UTC)
				timestamps = append(timestamps, timestamp)
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

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomWord(length int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func CreateEnumValues(numOfItems int) []string {
	enums := []string{}
	for i := 0; i < numOfItems; i++ {

		word := generateRandomWord(5)
		enums = append(enums, word)
	}
	return enums
}
func CreateBinaryBooleanValues(numOfItems int) []bool {
	var values []bool = make([]bool, numOfItems)
	var currentValue bool = false
	for i := 0; i < numOfItems; i++ {
		currentValue = !currentValue

		values[i] = currentValue

	}
	return values
}

func CreateBinaryValues(numOfItems int) []string {
	var values []string = make([]string, numOfItems)
	var currentValue bool = false
	for i := 0; i < numOfItems; i++ {
		currentValue = !currentValue
		if currentValue {
			values[i] = "on"
		} else {
			values[i] = "off"
		}
	}
	return values
}

func AddBinaryDataToExposeMetricsResult(expose *metrics.ExposeTimeRangeMetricsResult, values []string, timestamps []time.Time) *metrics.ExposeTimeRangeMetricsResult {
	var prevValue *binaryValue = nil
	for idx, value := range values {
		if prevValue == nil {
			prevValue = &binaryValue{value, timestamps[idx]}
			continue
		}
		if prevValue.Value != value {
			expose.Add(prevValue.Value, prevValue.Timestamp, timestamps[idx])
			prevValue = &binaryValue{value, timestamps[idx]}
		}
	}

	return expose
}

func StructToBytes(p interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
