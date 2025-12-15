package query

import "node-herder/internal/metrics/domain"

type MetricsQueryRequest struct {
	DeviceIds []string
	Expose    string

	Time domain.TimeQuery

	Aggregation domain.AggregationType

	Filters []domain.MetricFilter

	Limit    int
	SortDesc bool
}

//
// func resolveTime(q TimeQuery, now time.Time) (from, to time.Time) {
// 	to = now

// 	switch q.Relative {
// 	case RelToday, RelSinceMidnight:
// 		from = startOfDay(now)
// 	case RelYesterday:
// 		from = startOfDay(now.Add(-24 * time.Hour))
// 		to = startOfDay(now)
// 	case RelLast10Minutes:
// 		from = now.Add(-10 * time.Minute)
// 	case RelLatest, "":
// 		from = time.Time{} // repo decides “latest”
// 	default:
// 		panic("unsupported relative time")
// 	}
// 	return
// }
