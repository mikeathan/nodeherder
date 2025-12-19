package services

import (
	"context"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/models/devices"
	"time"
)

type DeviceResolver interface {
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
}

type MetricsQuerier interface {
	QueryDevice(deviceID string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error)
}

type QueryStore interface {
	DeviceResolver
	MetricsQuerier
}

type QueryService struct {
	devices DeviceResolver
	metrics MetricsQuerier
}

func NewQueryService(store QueryStore) *QueryService {
	return &QueryService{
		devices: store,
		metrics: store,
	}
}

func (s *QueryService) Query(ctx context.Context, req query.MetricsQueryRequest) (*[]query.MetricsQueryResponse, error) {

	devices, err := s.devices.FindDeviceByIds(req.DeviceIds)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	from, to := query.ResolveTime(req.Time, now)

	response := query.MetricsQueryResponse{
		Expose: req.Expose,
		From:   from.UnixMilli(),
		To:     to.UnixMilli(),
		Values: make([]query.MetricsQueryDeviceResponse, 0, len(req.DeviceIds)),
	}

	for _, device := range devices {

		expose, ok := device.Exposes[req.Expose]
		if !ok {
			continue
		}

		collector, err := domain.NewExposeResult(req.Expose, expose.Type, req.Aggregation, from, to)
		if err != nil {
			return nil, err
		}

		collectors := map[string]domain.ExposeResult{
			req.Expose: collector,
		}

		if _, err := s.metrics.QueryDevice(device.Id, from, to, req.Filters, collectors); err != nil {
			return nil, err
		}

		if req.Aggregation == domain.AggNone && (req.Limit > 0 || req.SortDesc) {
			query.ApplyLimitSortBy(collector, req.Limit, req.SortDesc)
		}

		value, err := buildDeviceQueryResponse(device.Id, req.Aggregation, collector)
		if err != nil {
			return nil, err
		}
		if value == nil {
			continue
		}

		response.Values = append(response.Values, *value)
	}

	responses := []query.MetricsQueryResponse{response}
	return &responses, nil
}

func buildDeviceQueryResponse(deviceID string, aggregation domain.AggregationType, expose domain.ExposeResult) (*query.MetricsQueryDeviceResponse, error) {
	value, timestamp, ok, err := query.AggregateExposeValue(aggregation, expose)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &query.MetricsQueryDeviceResponse{
		DeviceId:  deviceID,
		Value:     value,
		Timestamp: timestamp,
	}, nil
}
