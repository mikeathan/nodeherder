package weather

// Limits and timestamp layouts shared by provider series parsing and mapping validation.

const (
	// maxUnitScale bounds a provider-declared multiplication factor.
	maxUnitScale = 1000
	// maxSeriesRows bounds each hourly, current or daily provider series.
	maxSeriesRows = 384
	// maxNestedParents bounds the outer list used by nested provider series.
	maxNestedParents = 16
	// maxClockOffsetSeconds bounds a fixed daily UTC offset of ±14 hours.
	maxClockOffsetSeconds = 50400
	// maxUnixTimestamp accepts integer instants through the end of 2199.
	maxUnixTimestamp = 7258118400
)

// timeFormat identifies how a provider encodes a series timestamp.
type timeFormat string

const (
	formatUnix    timeFormat = "unix"
	formatRFC3339 timeFormat = "rfc3339"
	formatDate    timeFormat = "date"
	formatLocal   timeFormat = "local"
)

// Timestamp layouts accepted from provider payloads.
const (
	localMinuteLayout = "2006-01-02T15:04"
	localSecondLayout = "2006-01-02T15:04:05"
	localOffsetLayout = "2006-01-02T15:04Z07:00"
)
