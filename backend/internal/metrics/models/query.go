package metrics

import "time"

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

	FieldValue MetricField = "value"
	FieldState MetricField = "state"
	FieldTimestamp MetricField = "timestamp"

	RelLatest         RelativeTime = "latest"   
	RelLast5Minutes   RelativeTime = "last_5m"
	RelLast10Minutes  RelativeTime = "last_10m"
	RelLast30Minutes  RelativeTime = "last_30m"
	RelLast1Hour      RelativeTime = "last_1h"
	RelLast24Hours    RelativeTime = "last_24h"
	RelToday          RelativeTime = "today"
	RelYesterday      RelativeTime = "yesterday"
	RelSinceMidnight  RelativeTime = "since_midnight"
)

type MetricsQuery struct {
	DeviceIds []string
	Expose    string

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
