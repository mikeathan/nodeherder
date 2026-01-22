package resolver_test

import (
	"node-herder/internal/mcp/resolver"
	"testing"
)

func TestScorer_Score(t *testing.T) {
	scorer := resolver.NewScorer(resolver.DefaultTokenConfig)

	tests := []struct {
		name    string
		target  string
		device  string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "exact match",
			target:  "attic temperature sensor",
			device:  "Attic temperature sensor",
			wantMin: 0.9,
			wantMax: 1.0,
		},
		{
			name:    "partial match",
			target:  "attic temperature",
			device:  "Attic temperature sensor",
			wantMin: 0.5,
			wantMax: 1.0,
		},
		{
			name:    "no match",
			target:  "kitchen light",
			device:  "Attic temperature sensor",
			wantMin: 0.0,
			wantMax: 0.3,
		},
		{
			name:    "common words still match when they're the only token",
			target:  "sensor",
			device:  "Attic temperature sensor",
			wantMin: 0.9,
			wantMax: 1.0, // Single token matching gives full ratio
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.Score(tt.target, tt.device)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("Score(%q, %q) = %v, want between %v and %v",
					tt.target, tt.device, score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestScorer_IsAmbiguous(t *testing.T) {
	scorer := resolver.NewScorer(resolver.DefaultTokenConfig)

	tests := []struct {
		name   string
		scores []float64
		want   bool
	}{
		{
			name:   "clear winner",
			scores: []float64{0.9, 0.5, 0.3},
			want:   false,
		},
		{
			name:   "ambiguous - scores too close",
			scores: []float64{0.85, 0.80, 0.3},
			want:   true,
		},
		{
			name:   "single candidate",
			scores: []float64{0.9},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scorer.IsAmbiguous(tt.scores); got != tt.want {
				t.Errorf("IsAmbiguous(%v) = %v, want %v", tt.scores, got, tt.want)
			}
		})
	}
}

type fakeDeviceStore struct {
	devices []resolver.DeviceInfo
}

func (f fakeDeviceStore) AllDeviceInfo() ([]resolver.DeviceInfo, error) {
	return f.devices, nil
}

func TestResolver_Resolve(t *testing.T) {
	store := fakeDeviceStore{
		devices: []resolver.DeviceInfo{
			{ID: "dev1", Name: "Attic temperature sensor"},
			{ID: "dev2", Name: "Living room presence sensor"},
			{ID: "dev3", Name: "Garden temperature"},
			{ID: "dev4", Name: "Kitchen power socket"},
		},
	}

	r := resolver.New(store)

	tests := []struct {
		name      string
		target    string
		wantID    string
		wantErr   bool
		wantAmbig bool
	}{
		{
			name:   "exact match",
			target: "attic temperature sensor",
			wantID: "dev1",
		},
		{
			name:   "partial match",
			target: "garden temperature",
			wantID: "dev3",
		},
		{
			name:    "no match",
			target:  "bedroom light",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, candidates, err := r.Resolve(tt.target)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Resolve(%q) expected error, got none", tt.target)
				}
				return
			}

			if tt.wantAmbig {
				if len(candidates) == 0 {
					t.Errorf("Resolve(%q) expected candidates, got none", tt.target)
				}
				return
			}

			if id != tt.wantID {
				t.Errorf("Resolve(%q) = %q, want %q", tt.target, id, tt.wantID)
			}
		})
	}
}
