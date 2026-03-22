package intent_test

import (
	"fmt"
	"node-herder/internal/mcp/intent"
	"testing"
	"time"
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
			name: "valid intent with wrapper",
			input: `{
				"name": "query_device",
				"arguments": {
					"target_name": "attic temperature sensor",
					"metrics": ["temperature"],
					"time_scope": "today",
					"aggregation": "latest_value"
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid intent with nested wrapper",
			input: `{
				"arguments": {
					"target_name": "attic temperature sensor",
					"metrics": ["temperature"],
					"time_scope": "today",
					"aggregation": "latest_value"
				}
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

func TestValidateWithRules(t *testing.T) {
	alwaysPass := func(i *intent.Intent, ctx *intent.DeviceContext) error {
		return nil
	}
	alwaysFail := func(i *intent.Intent, ctx *intent.DeviceContext) error {
		return fmt.Errorf("always fails")
	}

	tests := []struct {
		name    string
		rules   []intent.Rule
		wantErr bool
	}{
		{
			name:    "empty rules",
			rules:   []intent.Rule{},
			wantErr: false,
		},
		{
			name:    "all pass",
			rules:   []intent.Rule{alwaysPass, alwaysPass},
			wantErr: false,
		},
		{
			name:    "first fails",
			rules:   []intent.Rule{alwaysFail, alwaysPass},
			wantErr: true,
		},
		{
			name:    "second fails",
			rules:   []intent.Rule{alwaysPass, alwaysFail},
			wantErr: true,
		},
	}

	i := &intent.Intent{TargetName: "test", Metrics: []string{"m1"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := intent.ValidateWithRules(i, nil, tt.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMetricsExist(t *testing.T) {
	tests := []struct {
		name    string
		intent  *intent.Intent
		ctx     *intent.DeviceContext
		wantErr bool
	}{
		{
			name:    "nil context skips validation",
			intent:  &intent.Intent{TargetName: "test", Metrics: []string{"foo"}},
			ctx:     nil,
			wantErr: false,
		},
		{
			name:   "metric exists",
			intent: &intent.Intent{TargetName: "test", Metrics: []string{"temperature"}},
			ctx: &intent.DeviceContext{
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature"},
				},
			},
			wantErr: false,
		},
		{
			name:   "metric does not exist",
			intent: &intent.Intent{TargetName: "test", Metrics: []string{"humidity"}},
			ctx: &intent.DeviceContext{
				Name: "Test Device",
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature"},
				},
			},
			wantErr: true,
		},
		{
			name:   "multiple metrics - one missing",
			intent: &intent.Intent{TargetName: "test", Metrics: []string{"temperature", "pressure"}},
			ctx: &intent.DeviceContext{
				Name: "Test Device",
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := intent.ValidateMetricsExist(tt.intent, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetricsExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAggregationSupported(t *testing.T) {
	tests := []struct {
		name    string
		intent  *intent.Intent
		ctx     *intent.DeviceContext
		wantErr bool
	}{
		{
			name:    "nil context skips validation",
			intent:  &intent.Intent{TargetName: "test", Metrics: []string{"foo"}, Aggregation: "avg"},
			ctx:     nil,
			wantErr: false,
		},
		{
			name:    "empty aggregation skips validation",
			intent:  &intent.Intent{TargetName: "test", Metrics: []string{"temperature"}, Aggregation: ""},
			ctx:     &intent.DeviceContext{},
			wantErr: false,
		},
		{
			name:   "supported aggregation",
			intent: &intent.Intent{TargetName: "test", Metrics: []string{"temperature"}, Aggregation: "avg"},
			ctx: &intent.DeviceContext{
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature", Aggregations: []string{"avg", "min", "max"}},
				},
			},
			wantErr: false,
		},
		{
			name:   "unsupported aggregation",
			intent: &intent.Intent{TargetName: "test", Metrics: []string{"temperature"}, Aggregation: "median"},
			ctx: &intent.DeviceContext{
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature", Aggregations: []string{"avg", "min", "max"}},
				},
			},
			wantErr: true,
		},
		{
			name:   "metric not in context skips validation for that metric",
			intent: &intent.Intent{TargetName: "test", Metrics: []string{"unknown"}, Aggregation: "avg"},
			ctx: &intent.DeviceContext{
				Metrics: map[string]intent.MetricInfo{
					"temperature": {Name: "temperature", Aggregations: []string{"avg"}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := intent.ValidateAggregationSupported(tt.intent, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAggregationSupported() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTranslate_AggregationMapping(t *testing.T) {
	tests := []struct {
		name        string
		aggregation string
		want        string // domain.AggregationType.String() equivalent
	}{
		{"count_events maps to count", "count_events", "count"},
		{"count maps to count", "count", "count"},
		{"latest_value maps to last", "latest_value", "last"},
		{"last maps to last", "last", "last"},
		{"last_event maps to last_event", "last_event", "last_event"},
		{"min_value maps to min", "min_value", "min"},
		{"min maps to min", "min", "min"},
		{"max_value maps to max", "max_value", "max"},
		{"max maps to max", "max", "max"},
		{"avg_value maps to avg", "avg_value", "avg"},
		{"avg maps to avg", "avg", "avg"},
		{"empty defaults to last", "", "last"},
		{"unknown defaults to last", "unknown_agg", "last"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &intent.Intent{
				TargetName:  "test",
				Metrics:     []string{"m"},
				Aggregation: tt.aggregation,
			}
			req := intent.Translate(in, "dev1", intent.DefaultTranslateOptions)
			if string(req.Aggregation) != tt.want {
				t.Errorf("Translate() aggregation = %v, want %v", req.Aggregation, tt.want)
			}
		})
	}
}

func TestTranslate_WithFilters(t *testing.T) {
	in := &intent.Intent{
		TargetName: "test",
		Metrics:    []string{"temperature"},
		Filters: []intent.Filter{
			{Field: "value", Op: ">", Value: 20},
			{Field: "state", Op: "=", Value: "on"},
			{Field: "timestamp", Op: "!=", Value: 12345},
		},
	}

	req := intent.Translate(in, "dev1", intent.DefaultTranslateOptions)

	if len(req.Filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(req.Filters))
	}

	// Verify field mapping
	expectedFields := []string{"value", "state", "timestamp"}
	for i, expected := range expectedFields {
		if string(req.Filters[i].Field) != expected {
			t.Errorf("filter[%d].Field = %v, want %v", i, req.Filters[i].Field, expected)
		}
	}
}

func TestTranslate_CustomTimezone(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	opts := intent.TranslateOptions{
		Timezone: time.FixedZone("EST", -5*60*60),
		Clock:    func() time.Time { return fixedTime },
	}

	in := &intent.Intent{
		TargetName: "test",
		Metrics:    []string{"m"},
		TimeScope:  "today",
	}

	req := intent.Translate(in, "dev1", opts)

	// The time range should reflect the EST timezone
	if req.Time.From.Location().String() != "EST" {
		t.Errorf("Time.From should be in EST timezone, got %v", req.Time.From.Location())
	}
}
