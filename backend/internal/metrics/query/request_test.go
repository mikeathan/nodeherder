package query_test

import (
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"testing"
	"time"
)

func TestResolveTime(t *testing.T) {
	now := time.Date(2024, time.January, 2, 15, 4, 5, 0, time.UTC)
	dayStart := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name string
		in   domain.TimeQuery
		want domain.TimeQuery
	}{
		{
			name: "no range",
			in:   domain.TimeQuery{From: dayStart, To: now},
			want: domain.TimeQuery{From: dayStart, To: now},
		},
		{
			name: "lookback hours",
			in:   domain.TimeQuery{Lookback: "6h"},
			want: domain.TimeQuery{From: now.Add(-6 * time.Hour), To: now},
		},
		{
			name: "lookback days",
			in:   domain.TimeQuery{Lookback: "3d"},
			want: domain.TimeQuery{From: now.Add(-72 * time.Hour), To: now},
		},
		{
			name: "lookback invalid",
			in:   domain.TimeQuery{Lookback: "bad", From: dayStart, To: now},
			want: domain.TimeQuery{From: dayStart, To: now},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			from, to := query.ResolveTime(tc.in, now)

			if !from.Equal(tc.want.From) {
				t.Fatalf("from mismatch: want %v got %v", tc.want.From, from)
			}
			if !to.Equal(tc.want.To) {
				t.Fatalf("to mismatch: want %v got %v", tc.want.To, to)
			}
		})
	}
}
