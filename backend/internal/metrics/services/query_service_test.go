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
		Expose:    "temperature",
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
