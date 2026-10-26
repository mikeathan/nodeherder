package waste

import (
	"time"

	model "node-herder/models/integrations"
)

// horizonExhausted treats missing future coverage as a renewal opportunity, never cancellation.
// Dates are canonical sorted local dates; the last date is the provider's latest known evidence.
func horizonExhausted(snapshot model.WasteSchedule, now time.Time) bool {
	if snapshot.Timezone != model.WasteTimezone {
		return false
	}
	today, err := model.WasteCalendarDate(now)
	if err != nil {
		return false
	}
	for _, service := range snapshot.Services {
		if service.DateState == "known" && len(service.Dates) > 0 && service.Dates[len(service.Dates)-1] < today {
			return true
		}
	}
	return false
}
