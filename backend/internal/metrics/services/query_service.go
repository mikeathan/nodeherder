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

	responses := make([]query.MetricsQueryResponse, 0, len(req.Exposes))

	// Iterate over each requested metric
	for _, exposeName := range req.Exposes {

		// Check if any device actually has this expose
		hasExpose := false
		for _, device := range devices {
			if _, ok := device.Exposes[exposeName]; ok {
				hasExpose = true
				break
			}
		}
		if !hasExpose {
			continue // Skip metrics that don't exist on any queried device
		}

		response := query.MetricsQueryResponse{
			Expose: exposeName,
			From:   from.UnixMilli(),
			To:     to.UnixMilli(),
			Values: make([]query.MetricsQueryDeviceResponse, 0, len(req.DeviceIds)),
		}

		for _, device := range devices {
			expose, ok := device.Exposes[exposeName]
			if !ok {
				continue
			}

			collector, err := domain.NewExposeResult(exposeName, expose.Type, req.Aggregation, from, to)
			if err != nil {
				continue
			}

			collectors := map[string]domain.ExposeResult{
				exposeName: collector,
			}

			if _, err := s.metrics.QueryDevice(device.Id, from, to, req.Filters, collectors); err != nil {
				// Don't fail the whole request if one metric is missing (e.g. bucket not found)
				continue
			}

			if req.Aggregation == domain.AggNone && (req.Limit > 0 || req.SortDesc) {
				query.ApplyLimitSortBy(collector, req.Limit, req.SortDesc)
			}

			value, err := buildDeviceQueryResponse(device.Id, req.Aggregation, req.AggregationValue, collector)
			if err != nil {
				return nil, err
			}
			if value == nil {
				continue
			}

			response.Values = append(response.Values, *value)
		}

		responses = append(responses, response)
	}

	return &responses, nil
}

func buildDeviceQueryResponse(deviceID string, aggregation domain.AggregationType, aggregationValue any, expose domain.ExposeResult) (*query.MetricsQueryDeviceResponse, error) {
	value, timestamp, ok, err := query.AggregateExposeValue(aggregation, aggregationValue, expose)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	var formattedTime string
	if timestamp > 0 {
		formattedTime = time.UnixMilli(timestamp).UTC().Format(time.RFC3339)
	}

	return &query.MetricsQueryDeviceResponse{
		DeviceId:      deviceID,
		Value:         value,
		Timestamp:     timestamp,
		FormattedTime: formattedTime,
	}, nil
}
