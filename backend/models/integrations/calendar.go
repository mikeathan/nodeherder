package integrations

import (
	"fmt"
	"sync"
	"time"
	_ "time/tzdata" // Native/minimal-container builds retain the source calendar without external zoneinfo.
)

// WasteTimezone is the source calendar required by waste_schedule/v1. It is not the
// installation's display/automation timezone; changing that must not reinterpret UK bin dates.
const WasteTimezone = "Europe/London"

// Timezone rules are immutable; share one load instead of repeated filesystem reads per resource.
var wasteCalendarLocation = sync.OnceValues(func() (*time.Location, error) {
	return time.LoadLocation(WasteTimezone)
})

// WasteCalendarDate converts an operational instant to a YYYY-MM-DD UK collection date.
// Callers share this boundary for midnight/horizon decisions, independently of process TZ.
// A location error is returned rather than silently substituting UTC and choosing a wrong day.
func WasteCalendarDate(now time.Time) (string, error) {
	location, err := wasteCalendarLocation()
	if err != nil {
		return "", fmt.Errorf("load waste source calendar: %w", err)
	}
	return now.In(location).Format(time.DateOnly), nil
}
