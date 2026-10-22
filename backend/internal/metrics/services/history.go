package services

import (
	"context"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/models/devices"
	"time"
)

// ValidateHistory bounds opt-in history requests without changing legacy query behavior.
func ValidateHistory(req query.MetricsQueryRequest) error {
	if len(req.DeviceIds) != 1 {
		return fmt.Errorf("history requires one device")
	}
	if len(req.Exposes) == 0 || len(req.Exposes) > 32 {
		return fmt.Errorf("history requires 1–32 exposes")
	}
	if req.Limit < 1 || req.Limit > maxQueryPoints {
		return fmt.Errorf("history limit must be 1–10001")
	}
	if req.Aggregation != domain.AggNone || len(req.Filters) != 0 || req.SortDesc {
		return fmt.Errorf("history requires ascending raw samples without filters")
	}
	return validateHistoryRange(req)
}

func validateHistoryRange(req query.MetricsQueryRequest) error {
	if req.Time.Lookback != "" {
		return fmt.Errorf("history requires explicit dates")
	}
	from, to := req.Time.From, req.Time.To
	if from.IsZero() || to.IsZero() {
		return fmt.Errorf("history requires explicit dates")
	}
	if !to.After(from) || to.Sub(from) > 31*24*time.Hour {
		return fmt.Errorf("history range must be positive and at most 31 days")
	}
	return nil
}

// QueryHistory scans each device once, sharing scan/time limits across all selected exposes.
func (s *QueryService) QueryHistory(ctx context.Context, req query.MetricsQueryRequest) (*[]query.MetricsQueryResponse, error) {
	if err := ValidateHistory(req); err != nil {
		return nil, err
	}
	found, err := s.devices.FindDeviceByIds(req.DeviceIds)
	if err != nil {
		return nil, err
	}
	if len(found) != 1 {
		return nil, fmt.Errorf("history device not found")
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	budget := &domain.HistoryBudget{Context: ctx, Remaining: maxScannedRows}
	results, collectors := historyCollectors(found[0], req, budget)
	if err := s.readHistory(req, collectors); err != nil {
		return nil, err
	}
	responses := historyResponses(found[0], req, results)
	return &responses, nil
}

func historyCollectors(device *devices.Device, req query.MetricsQueryRequest, budget *domain.HistoryBudget) (map[string]*domain.HistoryResult, map[string]domain.ExposeResult) {
	results := make(map[string]*domain.HistoryResult, len(req.Exposes))
	collectors := make(map[string]domain.ExposeResult, len(req.Exposes))
	for _, name := range req.Exposes {
		expose, ok := device.Exposes[name]
		if !ok {
			continue
		}
		result := &domain.HistoryResult{
			Name: name, Type: expose.Type, From: req.Time.From.UnixMilli(), To: req.Time.To.UnixMilli(),
			Data: []domain.HistoryPoint{}, Limit: req.Limit, Budget: budget,
		}
		results[name] = result
		collectors[name] = result
	}
	return results, collectors
}

func (s *QueryService) readHistory(req query.MetricsQueryRequest, collectors map[string]domain.ExposeResult) error {
	if len(collectors) == 0 {
		return nil
	}
	_, err := s.metrics.QueryDevice(req.DeviceIds[0], req.Time.From, req.Time.To, nil, collectors)
	if errors.Is(err, bolt.ErrBucketNotFound) {
		return nil
	}
	return err
}

func historyResponses(device *devices.Device, req query.MetricsQueryRequest, results map[string]*domain.HistoryResult) []query.MetricsQueryResponse {
	responses := make([]query.MetricsQueryResponse, 0, len(results))
	for _, name := range req.Exposes {
		result, ok := results[name]
		if !ok {
			continue
		}
		responses = append(responses, query.MetricsQueryResponse{
			Expose: name, From: result.From, To: result.To,
			Values: []query.MetricsQueryDeviceResponse{{DeviceId: device.Id, Unit: device.Exposes[name].Unit, Value: result}},
		})
	}
	return responses
}
