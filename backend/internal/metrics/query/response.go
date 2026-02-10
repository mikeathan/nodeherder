package query

type MetricsQueryResponse struct {
	Expose string                       `json:"expose"`
	From   int64                        `json:"query_start_timestamp"`
	To     int64                        `json:"query_end_timestamp"`
	Values []MetricsQueryDeviceResponse `json:"values"`
}

type MetricsQueryDeviceResponse struct {
	DeviceId      string `json:"device_id"`
	Value         any    `json:"value"`
	Timestamp     int64  `json:"timestamp"`
	FormattedTime string `json:"formatted_time,omitempty"`
}
