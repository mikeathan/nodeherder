package runtime

import (
	"testing"

	model "node-herder/models/integrations"
)

func TestCatalogStatusSummaries(t *testing.T) {
	tests := []struct {
		instance model.Instance
		want     string
	}{
		{model.Instance{Enabled: true, Snapshot: &model.Resource{}}, "ready"},
		{model.Instance{Enabled: true}, "pending"},
		{model.Instance{Enabled: false, Snapshot: &model.Resource{}}, "disabled"},
		{model.Instance{Enabled: true, LastError: model.Fail("provider_error", "Failed")}, "error"},
	}
	for _, tt := range tests {
		if got := catalogInstanceStatus(tt.instance); got != tt.want {
			t.Fatalf("status = %q, want %q", got, tt.want)
		}
	}
	if got := catalogWorstStatus("ready", "disabled"); got != "disabled" {
		t.Fatalf("worst = %q", got)
	}
	if got := catalogWorstStatus("error", "disabled"); got != "error" {
		t.Fatalf("worst = %q", got)
	}
	if got := catalogWorstStatus("", "pending"); got != "pending" {
		t.Fatalf("worst = %q", got)
	}
}
