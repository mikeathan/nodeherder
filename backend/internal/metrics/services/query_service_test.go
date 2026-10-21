package services_test

import (
	"context"
	"errors"
	"fmt"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/internal/metrics/services"
	"node-herder/models/devices"
	"node-herder/utils"
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
	collector, ok := collectors["temperature"]
	if !ok {
		return nil, fmt.Errorf("collector not found")
	}

	for n := 1; n <= 3; n++ {
		raw, _ := utils.AnyToByteArray(float32(n))
		if err := collector.Collect(time.UnixMilli(int64(n)), raw); err != nil {
			return nil, err
		}
	}
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
		if collector.GetType() == "numeric" {
			raw, _ := utils.AnyToByteArray(float32(42))
			if err := collector.Collect(time.UnixMilli(1000), raw); err != nil {
				return nil, err
			}
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

type countingQueryStore struct {
	fakeMultiMetricStore
	calls int
	scan  func(map[string]domain.ExposeResult) error
}

func (s *countingQueryStore) QueryDevice(id string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error) {
	s.calls++
	if s.scan != nil {
		return nil, s.scan(collectors)
	}
	return s.fakeMultiMetricStore.QueryDevice(id, from, to, filters, collectors)
}
func TestQueryServiceSharesOneScanAndHonorsCancellation(t *testing.T) {
	device := &devices.Device{Id: "d", Exposes: map[string]*devices.Entity{}}
	for _, name := range []string{"temperature", "humidity"} {
		entity := devices.NewEntity(name)
		entity.Type = "numeric"
		device.Exposes[name] = entity
	}
	source := &countingQueryStore{fakeMultiMetricStore: fakeMultiMetricStore{devices: []*devices.Device{device}}}
	service := services.NewQueryService(source)
	req := query.MetricsQueryRequest{DeviceIds: []string{"d"}, Exposes: []string{"temperature", "humidity", "temperature"}, Aggregation: domain.AggLast}
	result, err := service.Query(context.Background(), req)
	if err != nil || len(*result) != 2 || source.calls != 1 {
		t.Fatalf("result=%v, error=%v, scans=%d", result, err, source.calls)
	}
	for _, response := range *result {
		if len(response.Values) != 1 || response.Values[0].Value != float64(42) {
			t.Fatalf("changed result: %+v", response)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Query(ctx, req); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled query: %v", err)
	}
	if source.calls != 1 {
		t.Fatal("canceled query scanned storage")
	}
}
func TestQueryServicePropagatesScanBudgetAndCancellation(t *testing.T) {
	entity := devices.NewEntity("temperature")
	entity.Type = "numeric"
	device := &devices.Device{Id: "d", Exposes: map[string]*devices.Entity{"temperature": entity}}
	for _, exhausted := range []bool{false, true} {
		t.Run(fmt.Sprint(exhausted), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			source := &countingQueryStore{fakeMultiMetricStore: fakeMultiMetricStore{devices: []*devices.Device{device}}}
			source.scan = func(collectors map[string]domain.ExposeResult) error {
				if !exhausted {
					cancel()
				}
				guard := collectors["temperature"].(domain.ScanGuard)
				for i := 0; i <= 250000; i++ {
					if err := guard.CheckScan(); err != nil {
						return err
					}
				}
				return nil
			}
			_, err := services.NewQueryService(source).Query(ctx, query.MetricsQueryRequest{DeviceIds: []string{"d"}, Exposes: []string{"temperature"}, Aggregation: domain.AggNone, Limit: 1})
			if err == nil || source.calls != 1 {
				t.Fatalf("bounded query error=%v scans=%d", err, source.calls)
			}
			if !exhausted && !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation: %v", err)
			}
		})
	}
}
