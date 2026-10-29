package waste

// Checks collection horizon and refresh decisions.

import (
	model "node-herder/models/integrations"
	"testing"
	"time"
)

func TestHorizonExhaustedUsesSourceCalendar(t *testing.T) {
	// Still 25 October in UTC, but 26 October locally while UK summer time applies in 2024.
	now := time.Date(2024, 10, 25, 23, 30, 0, 0, time.UTC)
	for _, test := range []struct {
		name, zone, state string
		dates             []string
		want              bool
	}{
		{"past horizon", model.WasteTimezone, "known", []string{"2024-10-25"}, true},
		{"today retained", model.WasteTimezone, "known", []string{"2024-10-26"}, false},
		{"missing evidence", model.WasteTimezone, "unavailable", nil, false},
		{"different calendar", "UTC", "known", []string{"2024-10-25"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			schedule := model.WasteSchedule{Timezone: test.zone, Services: []model.WasteService{{DateState: test.state, Dates: test.dates}}}
			if got := horizonExhausted(schedule, now); got != test.want {
				t.Fatal(got)
			}
		})
	}
}
