package utils_test

import (
	"fmt"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"reflect"
	"sort"
	"testing"
	"time"
)

func AssertDeviceAnyDataTypeEvents(device *devices.Device, result *metrics.DeviceMetricsResult, timestamps []time.Time, values any, t *testing.T) {

	if result.DeviceId != device.Id {
		t.Errorf("deviceId mismatch want %v got %v: ", device.Id, result.DeviceId)
	}

	gotNumExposes := len(result.Exposes)
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
		event := result.Exposes[idx]

		if event.GetType() == "numeric" {
			AssertNumericExposeEvent(expose, event, timestamps, values.([]float32), t)
		} else if event.GetType() == "binary" {
			AssertBinaryExposeEvent(expose, event, timestamps, values.([]string), t)
		} else if event.GetType() == "enum" {
			AssertEnumExposeEvent(expose, event, timestamps, values.([]string), t)
		} else {
			t.Errorf("invalid expose type %v: ", event.GetType())
		}
		idx++ // ?????

	}
}
func AssertNumericExposeEvent(expose *devices.Entity, event metrics.ExposeResult, timestamps []time.Time, values []float32, t *testing.T) {
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
			t.Fatalf(fmt.Sprintf("Timestamp not found %v,  timestamp: %v", eventTimestamp, time.UnixMilli(eventTimestamp).UTC()))
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

func AssertBinaryExposeEvent(expose *devices.Entity, event metrics.ExposeResult, timestamps []time.Time, values []string, t *testing.T) {
	binaryEvent := metrics.ToTimeRangeExposeResults(event)

	if binaryEvent == nil {
		t.Errorf("invalid expos type want binary got %v: ", event.GetType())
	}

	if binaryEvent.Name != expose.Name {
		t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, binaryEvent.Name)
	}

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
			t.Fatalf(fmt.Sprintf("Start timestamp not found %v, timestamp: %v", eventStart, time.UnixMilli(eventStart).UTC()))
		}

		// we are expecting the end timestamp to be the next one
		if eventEnd != timestamps[dataIdx+1].UnixMilli() {
			t.Fatalf(fmt.Sprintf("End timestamp not matching %v", eventEnd))
		}

		wantValue := values[dataIdx]
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

func AssertEnumExposeEvent(expose *devices.Entity, event metrics.ExposeResult, timestamps []time.Time, values []string, t *testing.T) {
	enumEvent := metrics.ToTimeRangeExposeResults(event)

	if enumEvent == nil {
		t.Errorf("invalid expos type want enum got %v: ", event.GetType())
	}

	if enumEvent.Name != expose.Name {
		t.Errorf("exposeName mismatch want %v got %v: ", expose.Name, enumEvent.Name)
	}

	for idx, enumData := range enumEvent.Data {
		tsFound := false
		dataIdx := 0

		eventValue := enumData.X
		eventTimestamps := enumData.Y

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
			t.Fatalf(fmt.Sprintf("Start timestamp not found %v, timestamp: %v", eventStart, time.UnixMilli(eventStart).UTC()))
		}

		// we are expecting the end timestamp to be the next one
		if eventEnd != timestamps[dataIdx+1].UnixMilli() {
			// if last item check to see if its been padded with previous item value and end timestamp
			if idx+1 == len(enumEvent.Data) {
				if eventEnd != timestamps[dataIdx].UnixMilli() {
					t.Fatalf(fmt.Sprintf("End timestamp not matching %v", eventEnd))
				}
			} else {
				t.Fatalf(fmt.Sprintf("End timestamp not matching %v", eventEnd))
			}
		}

		wantValue := values[dataIdx]
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
func AssertBinaryExposeMetricResults(wantResults metrics.ExposeResult, gotResults metrics.ExposeResult, t *testing.T) {
	if wantResults.GetType() != gotResults.GetType() {
		t.Fatalf("Expected type %v', got '%v'", wantResults.GetType(), gotResults.GetType())
	}

	wantBinaryResults := metrics.ToTimeRangeExposeResults(wantResults)
	gotBinaryResults := metrics.ToTimeRangeExposeResults(gotResults)

	if wantBinaryResults.Name != gotBinaryResults.Name {
		t.Fatalf("Expected name %v', got '%v'", wantBinaryResults.Name, gotBinaryResults.Name)
	}

	if wantBinaryResults.From != gotBinaryResults.From {
		t.Fatalf("Expected from %v', got '%v'", wantBinaryResults.From, gotBinaryResults.From)
	}

	if wantBinaryResults.To != gotBinaryResults.To {
		t.Fatalf("Expected to %v', got '%v'", wantBinaryResults.To, gotBinaryResults.To)

	}
	for idx, wantEvent := range wantBinaryResults.Data {

		gotEvent := gotBinaryResults.Data[idx]

		if wantEvent.X != gotEvent.X {
			t.Fatalf("Expected X %v', got '%v'", wantEvent.X, gotEvent.X)
		}

		for tIdx, wantTimestamp := range wantEvent.Y {
			gotTimestamp := gotEvent.Y[tIdx]
			if wantTimestamp != gotTimestamp {
				t.Fatalf("Expected Y want %v', got '%v'", wantTimestamp, gotTimestamp)
			}
		}
	}
}

func AssertEnumExposeMetricResults(wantResults metrics.ExposeResult, gotResults metrics.ExposeResult, t *testing.T) {
	if wantResults.GetType() != gotResults.GetType() {
		t.Fatalf("Expected type %v', got '%v'", wantResults.GetType(), gotResults.GetType())
	}

	wantEnumResults := metrics.ToTimeRangeExposeResults(wantResults)
	gotEnumResults := metrics.ToTimeRangeExposeResults(gotResults)

	if wantEnumResults.Name != gotEnumResults.Name {
		t.Fatalf("Expected name %v', got '%v'", wantEnumResults.Name, gotEnumResults.Name)
	}

	if wantEnumResults.From != gotEnumResults.From {
		t.Fatalf("Expected from %v', got '%v'", wantEnumResults.From, gotEnumResults.From)
	}

	if wantEnumResults.To != gotEnumResults.To {
		t.Fatalf("Expected to %v', got '%v'", wantEnumResults.To, gotEnumResults.To)

	}
	for idx, wantEvent := range wantEnumResults.Data {

		gotEvent := gotEnumResults.Data[idx]

		if wantEvent.X != gotEvent.X {
			t.Fatalf("Expected X %v', got '%v'", wantEvent.X, gotEvent.X)
		}

		for tIdx, wantTimestamp := range wantEvent.Y {
			gotTimestamp := gotEvent.Y[tIdx]
			if wantTimestamp != gotTimestamp {
				t.Fatalf("Expected Y want %v', got '%v'", wantTimestamp, gotTimestamp)
			}
		}
	}
}
func AssertNumericExposeMetricResults(wantResults metrics.ExposeResult, gotResults metrics.ExposeResult, t *testing.T) {
	if wantResults.GetType() != gotResults.GetType() {
		t.Fatalf("Expected type %v', got '%v'", wantResults.GetType(), gotResults.GetType())
	}

	wantNumericResults := metrics.ToNumericExposeResults(wantResults)
	gotNumericResults := metrics.ToNumericExposeResults(gotResults)

	if wantNumericResults.Name != gotNumericResults.Name {
		t.Fatalf("Expected name %v', got '%v'", wantNumericResults.Name, gotNumericResults.Name)
	}

	if wantNumericResults.From != gotNumericResults.From {
		t.Fatalf("Expected from %v', got '%v'", wantNumericResults.From, gotNumericResults.From)
	}

	if wantNumericResults.To != gotNumericResults.To {
		t.Fatalf("Expected to %v', got '%v'", wantNumericResults.To, gotNumericResults.To)

	}
	for idx, wantEvent := range wantNumericResults.Data {

		gotEvent := gotNumericResults.Data[idx]

		if wantEvent.X != gotEvent.X {
			t.Fatalf("Expected X %v', got '%v'", wantEvent.X, gotEvent.X)
		}

		if wantEvent.Y != gotEvent.Y {
			t.Fatalf("Expected Y %v', got '%v'", wantEvent.Y, gotEvent.Y)
		}
	}
}
