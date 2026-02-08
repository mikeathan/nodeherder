package query

type MetricsQueryResponse struct {
	Expose string
	From   int64
	To     int64
	Values []MetricsQueryDeviceResponse
}

type MetricsQueryDeviceResponse struct {
	DeviceId      string `json:"deviceId"`
	Value         any    `json:"value"`
	Timestamp     int64  `json:"timestamp"`
	FormattedTime string `json:"formatted_time,omitempty"`
}
