package services

import (
	"context"
	"fmt"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/models/devices"
	"time"
)

// DeviceResolver exposes the subset of device repository methods that queries need.
type DeviceResolver interface {
	FindDevices(ids []string) ([]*devices.Device, error)
}

type MetricsQuerier interface {
	QueryDevice(deviceID string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error)
}

type QueryService struct {
	devices DeviceResolver
	metrics MetricsQuerier
}

func NewQueryService(devices DeviceResolver, metrics MetricsQuerier) *QueryService {
	return &QueryService{
		devices: devices,
		metrics: metrics,
	}
}

func (s *QueryService) Query(ctx context.Context, req query.MetricsQueryRequest) (*[]query.MetricsQueryResponse, error) {

	devices, err := s.devices.FindDevices(req.DeviceIds)
	if err != nil {
		return nil, err
	}

	from := req.Time.From
	to := req.Time.To

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
	value, timestamp, ok, err := aggregateExposeValue(aggregation, expose)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &query.MetricsQueryDeviceResponse{
		DeviceId:  deviceID,
		Value:     value,
		Timestamp: timestamp, to fix
	}, nil
}

func aggregateExposeValue(aggregation domain.AggregationType, expose domain.ExposeResult) (any, int64, bool, error) {
	switch expose.GetType() {
	case "numeric":
		return aggregateNumericValues(aggregation, expose.(*domain.ExposeNumericMetricsResult))
	case "binary":
		return aggregateBinaryValues(aggregation, expose.(*domain.ExposeBinaryEventsResult))
	case "enum":
		return aggregateTimeRangeValues(aggregation, expose.(*domain.ExposeTimeRangeMetricsResult))
	default:
		return nil, 0, false, fmt.Errorf("unsupported expose result type %T", expose)
	}
}

func aggregateNumericValues(aggregation domain.AggregationType, result *domain.ExposeNumericMetricsResult) (any, int64, bool, error) {
	dataPoints := result.Data
	count := len(dataPoints)

	switch aggregation {
	case domain.AggNone:
		return dataPoints, 0, true, nil
	case domain.AggLast:
		if count == 0 {
			return nil, 0, false, nil
		}
		last := dataPoints[count-1]
		return float64(last.Y), last.X, true, nil
	case domain.AggMin:
		if count == 0 {
			return nil, 0, false, nil
		}
		min := dataPoints[0]
		for _, point := range dataPoints[1:] {
			if point.Y < min.Y {
				min = point
			}
		}
		return float64(min.Y), min.X, true, nil
	case domain.AggMax:
		if count == 0 {
			return nil, 0, false, nil
		}
		max := dataPoints[0]
		for _, point := range dataPoints[1:] {
			if point.Y > max.Y {
				max = point
			}
		}
		return float64(max.Y), max.X, true, nil
	case domain.AggAvg:
		if count == 0 {
			return nil, 0, false, nil
		}
		var sum float64
		for _, point := range dataPoints {
			sum += float64(point.Y)
		}
		avg := sum / float64(count)
		return avg, 0, true, nil
	case domain.AggCount:
		return count, 0, true, nil
	default:
		return nil, 0, false, fmt.Errorf("aggregation %q unsupported for numeric expose", aggregation)
	}
}

func aggregateBinaryValues(aggregation domain.AggregationType, result *domain.ExposeBinaryEventsResult) (any, int64, bool, error) {
	dataPoints := result.Data
	count := len(dataPoints)

	switch aggregation {
	case domain.AggNone:
		return dataPoints, 0, true, nil
	case domain.AggLast:
		if count == 0 {
			return nil, 0, false, nil
		}
		last := dataPoints[count-1]
		return last.Value, last.Timestamp, true, nil
	case domain.AggCount:
		return count, 0, true, nil
	default:
		return nil, 0, false, fmt.Errorf("aggregation %q unsupported for binary expose", aggregation)
	}
}

func aggregateTimeRangeValues(aggregation domain.AggregationType, result *domain.ExposeTimeRangeMetricsResult) (any, int64, bool, error) {
	dataPoints := result.Data
	count := len(dataPoints)

	switch aggregation {
	case domain.AggNone:
		return dataPoints, 0, true, nil
	case domain.AggLast:
		if count == 0 {
			return nil, 0, false, nil
		}
		last := dataPoints[count-1]
		timestamp := last.Y[1]
		if timestamp == 0 {
			timestamp = last.Y[0]
		}
		return last.X, timestamp, true, nil
	case domain.AggCount:
		return count, 0, true, nil
	default:
		return nil, 0, false, fmt.Errorf("aggregation %q unsupported for enum expose", aggregation)
	}
}
