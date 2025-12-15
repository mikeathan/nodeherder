package services

import (
	"context"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/models/devices"
)

// DeviceResolver exposes the subset of device repository methods that queries need.
type DeviceResolver interface {
	FindDevices(ids []string) ([]*devices.Device, error)
}

type MetricsQuerier interface {
	Query(query.MetricsQueryRequest) (*[]query.MetricsQueryResponse, error)
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

	// build expose collectors
	collectors := make(map[string]domain.ExposeResult, len(req.DeviceIds))

	for _, device := range devices {

		expose, ok := device.Exposes[req.Expose]
		if !ok {
			continue
		}

		result, err := domain.NewExposeResult(req.Expose, expose.Type, req.Aggregation, from, to)
		if err != nil {
			return nil, err
		}
		collectors[device.Id] = result
	}

	//

	return nil, nil
}
