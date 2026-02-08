package services_test

import (
	"context"
	"fmt"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/internal/metrics/services"
	"node-herder/models/devices"
	"testing"
	"time"
)

type fakeQueryStore struct {
	devices []*devices.Device
}

func (f fakeQueryStore) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	return f.devices, nil
}

func (f fakeQueryStore) QueryDevice(deviceID string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error) {
	collector, ok := collectors["temperature"].(*domain.ExposeNumericMetricsResult)
	if !ok {
		return nil, fmt.Errorf("collector not found")
	}

	collector.Add(1, time.UnixMilli(1))
	collector.Add(2, time.UnixMilli(2))
	collector.Add(3, time.UnixMilli(3))
	return nil, nil
}

func TestQueryServiceAppliesLimitSort(t *testing.T) {
	entity := devices.NewEntity("temperature")
	entity.Type = "numeric"

	dev := &devices.Device{
		Id:      "dev1",
		Exposes: map[string]*devices.Entity{"temperature": entity},
	}

	service := services.NewQueryService(fakeQueryStore{devices: []*devices.Device{dev}})

	req := query.MetricsQueryRequest{
		DeviceIds: []string{"dev1"},
		Exposes:   []string{"temperature"},
		Time: domain.TimeQuery{
			From: time.UnixMilli(0),
			To:   time.UnixMilli(4),
		},
		Aggregation: domain.AggNone,
		Limit:       2,
		SortDesc:    true,
	}

	responses, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(*responses) != 1 || len((*responses)[0].Values) != 1 {
		t.Fatalf("expected 1 response and 1 value, got %+v", responses)
	}

	values, ok := (*responses)[0].Values[0].Value.([]*domain.NumericValue)
	if !ok {
		t.Fatalf("unexpected value type: %T", (*responses)[0].Values[0].Value)
	}

	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0].X != 3 || values[1].X != 2 {
		t.Fatalf("unexpected order after sort/limit: %+v", values)
	}
}

// fakeMultiMetricStore supports multiple metrics and can simulate errors.
type fakeMultiMetricStore struct {
	devices      []*devices.Device
	errorMetrics map[string]error // Metrics that should return errors
}

func (f fakeMultiMetricStore) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	return f.devices, nil
}

func (f fakeMultiMetricStore) QueryDevice(deviceID string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error) {
	for metricName, collector := range collectors {
		// Check if this metric should return an error
		if f.errorMetrics != nil {
			if err, ok := f.errorMetrics[metricName]; ok {
				return nil, err
			}
		}

		// Add sample data for numeric metrics
		if numericCollector, ok := collector.(*domain.ExposeNumericMetricsResult); ok {
			numericCollector.Add(42, time.UnixMilli(1000))
		}
	}
	return nil, nil
}

func TestQueryServiceFiltersNonExistentMetrics(t *testing.T) {
	// Device only has "temperature", not "humidity" or "pressure"
	entity := devices.NewEntity("temperature")
	entity.Type = "numeric"

	dev := &devices.Device{
		Id:      "dev1",
		Exposes: map[string]*devices.Entity{"temperature": entity},
	}

	store := fakeMultiMetricStore{devices: []*devices.Device{dev}}
	service := services.NewQueryService(store)

	req := query.MetricsQueryRequest{
		DeviceIds: []string{"dev1"},
		Exposes:   []string{"temperature", "humidity", "pressure"}, // Only temperature exists
		Time: domain.TimeQuery{
			From: time.UnixMilli(0),
			To:   time.UnixMilli(2000),
		},
		Aggregation: domain.AggNone,
	}

	responses, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// Should only return 1 response for "temperature"
	if len(*responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(*responses))
	}

	if (*responses)[0].Expose != "temperature" {
		t.Fatalf("expected expose 'temperature', got '%s'", (*responses)[0].Expose)
	}
}

func TestQueryServiceMultipleValidMetrics(t *testing.T) {
	// Device has both "temperature" and "humidity"
	tempEntity := devices.NewEntity("temperature")
	tempEntity.Type = "numeric"

	humidEntity := devices.NewEntity("humidity")
	humidEntity.Type = "numeric"

	dev := &devices.Device{
		Id: "dev1",
		Exposes: map[string]*devices.Entity{
			"temperature": tempEntity,
			"humidity":    humidEntity,
		},
	}

	store := fakeMultiMetricStore{devices: []*devices.Device{dev}}
	service := services.NewQueryService(store)

	req := query.MetricsQueryRequest{
		DeviceIds: []string{"dev1"},
		Exposes:   []string{"temperature", "humidity"},
		Time: domain.TimeQuery{
			From: time.UnixMilli(0),
			To:   time.UnixMilli(2000),
		},
		Aggregation: domain.AggNone,
	}

	responses, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// Should return 2 responses
	if len(*responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(*responses))
	}

	// Verify both metrics are present
	foundTemp, foundHumid := false, false
	for _, resp := range *responses {
		if resp.Expose == "temperature" {
			foundTemp = true
		}
		if resp.Expose == "humidity" {
			foundHumid = true
		}
	}

	if !foundTemp || !foundHumid {
		t.Fatalf("expected both temperature and humidity, got %+v", responses)
	}
}

func TestQueryServiceIgnoresQueryDeviceErrors(t *testing.T) {
	// Device has "temperature" and "presence"
	tempEntity := devices.NewEntity("temperature")
	tempEntity.Type = "numeric"

	presenceEntity := devices.NewEntity("presence")
	presenceEntity.Type = "binary"

	dev := &devices.Device{
		Id: "dev1",
		Exposes: map[string]*devices.Entity{
			"temperature": tempEntity,
			"presence":    presenceEntity,
		},
	}

	// Simulate "bucket not found" error for temperature
	store := fakeMultiMetricStore{
		devices:      []*devices.Device{dev},
		errorMetrics: map[string]error{"temperature": fmt.Errorf("bucket not found")},
	}
	service := services.NewQueryService(store)

	req := query.MetricsQueryRequest{
		DeviceIds: []string{"dev1"},
		Exposes:   []string{"temperature", "presence"},
		Time: domain.TimeQuery{
			From: time.UnixMilli(0),
			To:   time.UnixMilli(2000),
		},
		Aggregation: domain.AggNone,
	}

	responses, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("query should not fail when individual metrics error: %v", err)
	}

	// Should still return responses (possibly with empty values for the errored metric)
	if len(*responses) < 1 {
		t.Fatalf("expected at least 1 response, got %d", len(*responses))
	}
}

func TestQueryServiceFormattedTime(t *testing.T) {
	// Setup generic device/entity
	entity := devices.NewEntity("temperature")
	entity.Type = "numeric"
	dev := &devices.Device{
		Id:      "dev1",
		Exposes: map[string]*devices.Entity{"temperature": entity},
	}

	// This store mock injects sample data (42) at 1000ms
	store := fakeMultiMetricStore{devices: []*devices.Device{dev}}
	service := services.NewQueryService(store)

	req := query.MetricsQueryRequest{
		DeviceIds: []string{"dev1"},
		Exposes:   []string{"temperature"},
		Time: domain.TimeQuery{
			From: time.UnixMilli(0),
			To:   time.UnixMilli(2000),
		},
		Aggregation: domain.AggLast, // Aggregating to single value
	}

	responses, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(*responses) != 1 || len((*responses)[0].Values) != 1 {
		t.Fatalf("expected 1 response value")
	}

	val := (*responses)[0].Values[0]

	// Verify timestamp is 1000
	if val.Timestamp != 1000 {
		t.Errorf("expected timestamp 1000, got %d", val.Timestamp)
	}

	// Verify formatted time string
	expected := time.UnixMilli(1000).UTC().Format(time.RFC3339)
	if val.FormattedTime != expected {
		t.Errorf("expected formatted time %q, got %q", expected, val.FormattedTime)
	}
}
