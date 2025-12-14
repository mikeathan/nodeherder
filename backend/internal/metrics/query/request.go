package query

import (
	"time"
)

type AggregationType string

type OperationType string

type MetricField string

type RelativeTime string

const (
	AggNone  AggregationType = ""
	AggLast  AggregationType = "last"
	AggMin   AggregationType = "min"
	AggMax   AggregationType = "max"
	AggAvg   AggregationType = "avg"
	AggCount AggregationType = "count"

	OpEquals      OperationType = "="
	OpNotEquals   OperationType = "!="
	OpGreaterThan OperationType = ">"
	OpLessThan    OperationType = "<"

	FieldValue     MetricField = "value"
	FieldState     MetricField = "state"
	FieldTimestamp MetricField = "timestamp"

	RelLatest        RelativeTime = "latest"
	RelLast5Minutes  RelativeTime = "last_5m"
	RelLast10Minutes RelativeTime = "last_10m"
	RelLast30Minutes RelativeTime = "last_30m"
	RelLast1Hour     RelativeTime = "last_1h"
	RelLast24Hours   RelativeTime = "last_24h"
	RelToday         RelativeTime = "today"
	RelYesterday     RelativeTime = "yesterday"
	RelSinceMidnight RelativeTime = "since_midnight"
)

type MetricsQueryRequest struct {
	DeviceIds []string
	Expose    string
	//ExposeType bridge.ExposeDataType // to be resolved internally eg metadata

	Time TimeQuery

	Aggregation AggregationType

	Filters []MetricFilter

	Limit    int
	SortDesc bool
}

type TimeQuery struct {
	From     *time.Time
	To       *time.Time
	Relative string
}

type MetricFilter struct {
	Field MetricField
	Op    OperationType
	Value any
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
