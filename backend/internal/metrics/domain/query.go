package domain

import "time"

type AggregationType string

type OperationType string

type MetricField string

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
)

type TimeQuery struct {
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Lookback string    `json:"lookback"`
}

type MetricFilter struct {
	Field MetricField   `json:"field"`
	Op    OperationType `json:"op"`
	Value any           `json:"value"`
}
