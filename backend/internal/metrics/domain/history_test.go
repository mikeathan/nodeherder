package domain

import (
	"context"
	"node-herder/utils"
	"testing"
	"time"
)

func TestHistoryCollectPreservesEventsAndBounds(t *testing.T) {
	for _, tc := range []struct {
		kind  string
		value any
	}{{"enum", "press"}, {"binary", true}, {"numeric", float32(0)}} {
		t.Run(tc.kind, func(t *testing.T) {
			h := &HistoryResult{Type: tc.kind, Limit: 2}
			raw, err := utils.AnyToByteArray(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				if err := h.Collect(time.UnixMilli(int64(i)), raw); err != nil {
					t.Fatal(err)
				}
			}
			if len(h.Data) != 2 || !h.Partial || h.GetType() != tc.kind || h.Size() != 2 {
				t.Fatalf("unexpected history: %+v", h)
			}
			h.Flush()
		})
	}
}

func TestHistoryFailureAndCancellation(t *testing.T) {
	for _, kind := range []string{"numeric", "binary", "enum", "unknown"} {
		h := &HistoryResult{Type: kind, Limit: 1}
		if err := h.Collect(time.Now(), []byte("invalid")); err == nil {
			t.Fatal("invalid data accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	b := &HistoryBudget{Context: ctx, Remaining: 1}
	h := &HistoryResult{Budget: b}
	if err := h.CheckScan(); err != nil {
		t.Fatal(err)
	}
	if err := h.CheckScan(); err == nil {
		t.Fatal("budget not enforced")
	}
	cancel()
	if err := h.CheckScan(); err != context.Canceled {
		t.Fatalf("cancellation: %v", err)
	}
}
