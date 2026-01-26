package intent_test

import (
	"node-herder/internal/mcp/intent"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid intent",
			input: `{
				"target_name": "attic temperature sensor",
				"metrics": ["temperature"],
				"time_scope": "today",
				"aggregation": "latest_value"
			}`,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   "not json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := intent.ParseString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		intent  *intent.Intent
		ctx     *intent.DeviceContext
		wantErr bool
	}{
		{
			name: "valid intent",
			intent: &intent.Intent{
				TargetName:  "attic sensor",
				Metrics:     []string{"temperature"},
				TimeScope:   "today",
				Aggregation: "latest_value",
			},
			wantErr: false,
		},
		{
			name: "missing target_name",
			intent: &intent.Intent{
				TargetName: "",
				Metrics:    []string{"temperature"},
			},
			wantErr: true,
		},
		{
			name: "missing metrics",
			intent: &intent.Intent{
				TargetName: "attic sensor",
				Metrics:    []string{},
			},
			wantErr: true,
		},
		{
			name: "count_events on numeric metric",
			intent: &intent.Intent{
				TargetName:  "attic sensor",
				Metrics:     []string{"temperature"},
				Aggregation: "count_events",
			},
			ctx: &intent.DeviceContext{
				ID:   "dev1",
				Name: "Attic sensor",
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature", Type: "numeric"},
				},
			},
			wantErr: true,
		},
		{
			name: "latest_value with large range",
			intent: &intent.Intent{
				TargetName:  "attic sensor",
				Metrics:     []string{"temperature"},
				TimeScope:   "last_7_days",
				Aggregation: "latest_value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := intent.Validate(tt.intent, tt.ctx)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTranslate(t *testing.T) {
	in := &intent.Intent{
		TargetName:  "attic sensor",
		Metrics:     []string{"temperature"},
		TimeScope:   "today",
		Aggregation: "latest_value",
	}

	req := intent.Translate(in, "dev123", intent.DefaultTranslateOptions)

	if len(req.DeviceIds) != 1 || req.DeviceIds[0] != "dev123" {
		t.Errorf("unexpected DeviceIds: %v", req.DeviceIds)
	}
	if len(req.Exposes) != 1 || req.Exposes[0] != "temperature" {
		t.Errorf("unexpected Exposes: %v", req.Exposes)
	}
	if req.Aggregation != "last" {
		t.Errorf("unexpected Aggregation: %v", req.Aggregation)
	}
}
