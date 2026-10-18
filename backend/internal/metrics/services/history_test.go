package services_test

import (
	"context"
	"errors"
	bolt "go.etcd.io/bbolt"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/internal/metrics/services"
	"node-herder/models/devices"
	"testing"
	"time"
)

func historyRequest() query.MetricsQueryRequest {
	return query.MetricsQueryRequest{History: true, DeviceIds: []string{"dev1"}, Exposes: []string{"temperature"}, Limit: 10001, Time: domain.TimeQuery{From: time.Unix(100, 0), To: time.Unix(200, 0)}}
}

func TestValidateHistory(t *testing.T) {
	if err := services.ValidateHistory(historyRequest()); err != nil {
		t.Fatal(err)
	}
	cases := []func(*query.MetricsQueryRequest){
		func(r *query.MetricsQueryRequest) { r.DeviceIds = nil }, func(r *query.MetricsQueryRequest) { r.Exposes = nil },
		func(r *query.MetricsQueryRequest) { r.Limit = 0 }, func(r *query.MetricsQueryRequest) { r.Limit = 10002 },
		func(r *query.MetricsQueryRequest) { r.Aggregation = domain.AggLast }, func(r *query.MetricsQueryRequest) { r.SortDesc = true },
		func(r *query.MetricsQueryRequest) { r.Time.Lookback = "1h" }, func(r *query.MetricsQueryRequest) { r.Time.From = time.Time{} },
		func(r *query.MetricsQueryRequest) { r.Time.To = r.Time.From }, func(r *query.MetricsQueryRequest) { r.Time.To = r.Time.From.Add(32 * 24 * time.Hour) },
	}
	for i, change := range cases {
		r := historyRequest()
		change(&r)
		if services.ValidateHistory(r) == nil {
			t.Fatalf("case %d accepted", i)
		}
	}
}

func TestQueryHistoryDistinguishesEmptyAndFailed(t *testing.T) {
	e := devices.NewEntity("temperature")
	e.Type = "numeric"
	e.Unit = "°C"
	d := &devices.Device{Id: "dev1", Exposes: map[string]*devices.Entity{"temperature": e}}
	for _, failure := range []error{nil, bolt.ErrBucketNotFound, errors.New("read failed")} {
		s := services.NewQueryService(fakeMultiMetricStore{devices: []*devices.Device{d}, errorMetrics: map[string]error{"temperature": failure}})
		result, err := s.QueryHistory(context.Background(), historyRequest())
		if failure != nil && !errors.Is(failure, bolt.ErrBucketNotFound) {
			if err == nil {
				t.Fatal("read failure hidden")
			}
			continue
		}
		if err != nil || len(*result) != 1 {
			t.Fatalf("result=%v err=%v", result, err)
		}
		if (*result)[0].Values[0].Unit != "°C" {
			t.Fatal("unit lost")
		}
	}
	r := historyRequest()
	r.Limit = 0
	s := services.NewQueryService(fakeMultiMetricStore{})
	if _, err := s.QueryHistory(context.Background(), r); err == nil {
		t.Fatal("invalid query accepted")
	}
	if _, err := s.QueryHistory(context.Background(), historyRequest()); err == nil {
		t.Fatal("missing device accepted")
	}
}
