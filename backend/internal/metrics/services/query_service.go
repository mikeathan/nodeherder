package services

import (
	"context"
	"errors"
	"fmt"
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

const (
	queryTimeout   = 8 * time.Second
	maxScannedRows = 250000
	maxQueryItems  = 64
	maxQueryPoints = 10001
	maxQueryRange  = 366 * 24 * time.Hour
)

var errQueryBudget = errors.New("metrics scan budget exceeded; narrow the range")

type queryBudget struct {
	ctx       context.Context
	remaining int
}

func (b *queryBudget) check() error {
	if err := b.ctx.Err(); err != nil {
		return err
	}
	if b.remaining <= 0 {
		return errQueryBudget
	}
	b.remaining--
	return nil
}

// Only one collector carries the shared guard, so each scanned row consumes one unit.
type guardedQueryResult struct {
	domain.ExposeResult
	budget *queryBudget
}

func (r *guardedQueryResult) CheckScan() error { return r.budget.check() }

func (s *QueryService) Query(ctx context.Context, req query.MetricsQueryRequest) (*[]query.MetricsQueryResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(req.DeviceIds) > maxQueryItems || len(req.Exposes) > maxQueryItems || req.Limit < 0 || req.Limit > maxQueryPoints {
		return nil, fmt.Errorf("metrics query exceeds device, expose or result limits")
	}
	from, to := query.ResolveTime(req.Time, time.Now().UTC())
	if to.Before(from) || to.Sub(from) > maxQueryRange {
		return nil, fmt.Errorf("metrics query range must be ordered and at most 366 days")
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	found, err := s.devices.FindDeviceByIds(req.DeviceIds)
	if err != nil {
		return nil, err
	}
	budget := &queryBudget{ctx: ctx, remaining: maxScannedRows}
	perDevice := make([]map[string]domain.ExposeResult, len(found))
	for i, device := range found {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		collectors := queryCollectors(device, req, from, to)
		if len(collectors) == 0 {
			continue
		}
		if err := s.scanQuery(device.Id, from, to, req.Filters, collectors, budget); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if errors.Is(err, errQueryBudget) {
				return nil, err
			}
			// Legacy adapters can fail one expose independently. Retry only this error
			// path with fresh per-expose collectors, preserving successful peers.
			collectors = queryCollectors(device, req, from, to)
			for name, collector := range collectors {
				one := map[string]domain.ExposeResult{name: collector}
				if err := s.scanQuery(device.Id, from, to, req.Filters, one, budget); err != nil {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
					if errors.Is(err, errQueryBudget) {
						return nil, err
					}
					delete(collectors, name)
				}
			}
		}
		perDevice[i] = collectors
	}
	responses := make([]query.MetricsQueryResponse, 0, len(req.Exposes))
	seen := make(map[string]bool, len(req.Exposes))
	for _, name := range req.Exposes {
		if seen[name] {
			continue
		}
		seen[name] = true
		response := query.MetricsQueryResponse{Expose: name, From: from.UnixMilli(), To: to.UnixMilli(), Values: make([]query.MetricsQueryDeviceResponse, 0, len(found))}
		exists := false
		for i, device := range found {
			expose, ok := device.Exposes[name]
			if !ok {
				continue
			}
			exists = true
			collector := perDevice[i][name]
			if collector == nil {
				continue
			}
			if req.Aggregation == domain.AggNone && (req.Limit > 0 || req.SortDesc) {
				query.ApplyLimitSortBy(collector, req.Limit, req.SortDesc)
			}
			value, err := buildDeviceQueryResponse(device.Id, req.Aggregation, req.AggregationValue, collector, expose.Unit)
			if err != nil {
				return nil, err
			}
			if value != nil {
				response.Values = append(response.Values, *value)
			}
		}
		if exists {
			responses = append(responses, response)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &responses, nil
}

func queryCollectors(device *devices.Device, req query.MetricsQueryRequest, from, to time.Time) map[string]domain.ExposeResult {
	collectors := make(map[string]domain.ExposeResult, len(req.Exposes))
	for _, name := range req.Exposes {
		if _, exists := collectors[name]; exists {
			continue
		}
		if expose, ok := device.Exposes[name]; ok {
			if result, err := domain.NewExposeResult(name, expose.Type, req.Aggregation, from, to); err == nil {
				collectors[name] = result
			}
		}
	}
	return collectors
}

func (s *QueryService) scanQuery(id string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult, budget *queryBudget) error {
	scanned := make(map[string]domain.ExposeResult, len(collectors))
	first := true
	for name, collector := range collectors {
		if first {
			scanned[name] = &guardedQueryResult{ExposeResult: collector, budget: budget}
			first = false
		} else {
			scanned[name] = collector
		}
	}
	_, err := s.metrics.QueryDevice(id, from, to, filters, scanned)
	return err
}

func buildDeviceQueryResponse(deviceID string, aggregation domain.AggregationType, aggregationValue any, expose domain.ExposeResult, unit string) (*query.MetricsQueryDeviceResponse, error) {
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
		Unit:          unit,
		Timestamp:     timestamp,
		FormattedTime: formattedTime,
	}, nil
}
