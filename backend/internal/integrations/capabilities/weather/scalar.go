package weather

// Converts provider scalar values and timestamps into weather-friendly representations.

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

func pointer(root any, p string) any { v, _ := operations.Pointer(root, p); return v }

// number accepts supported provider number representations and returns nil for other values.
func number(v any) *float64 {
	var n float64
	var e error
	switch x := v.(type) {
	case json.Number:
		n, e = x.Float64()
	case float64:
		n = x
	case string:
		n, e = strconv.ParseFloat(x, 64)
	default:
		return nil
	}
	if e != nil {
		return nil
	}
	return &n
}

// text converts supported provider scalar values into condition or location text.
func text(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return ""
}

// instant parses supported provider timestamp formats using the supplied location for local times.
func instant(v any, format timeFormat, zone *time.Location) (time.Time, error) {
	if format == formatUnix {
		n := number(v)
		if n == nil || math.IsNaN(*n) || math.IsInf(*n, 0) || math.Trunc(*n) != *n || *n < 0 || *n > maxUnixTimestamp {
			return time.Time{}, fmt.Errorf("invalid time")
		}
		return time.Unix(int64(*n), 0).UTC(), nil
	}
	if format == formatLocal {
		return time.ParseInLocation(localMinuteLayout, text(v), zone)
	}
	timestamp, err := time.Parse(time.RFC3339, text(v))
	if err == nil {
		return timestamp, nil
	}
	return time.Parse(localOffsetLayout, text(v))
}

// clockValue reduces a provider clock value to local HH:MM. Providers may send a local
// timestamp/clock string or a Unix instant, depending on their time-format setting.
func clockValue(value any, zone *time.Location) (string, error) {
	switch typed := value.(type) {
	case string:
		return localClock(typed, zone)
	case json.Number:
		instant, err := typed.Float64()
		if err != nil {
			return "", fmt.Errorf("invalid clock")
		}
		return unixClock(instant, zone)
	case float64:
		return unixClock(typed, zone)
	default:
		return "", fmt.Errorf("invalid clock")
	}
}

// unixClock converts a whole-second Unix instant into the resource timezone's HH:MM clock.
func unixClock(value float64, zone *time.Location) (string, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value || value < 0 || value > maxUnixTimestamp {
		return "", fmt.Errorf("invalid clock")
	}
	return time.Unix(int64(value), 0).In(zone).Format(model.WeatherClockLayout), nil
}

// localClock reduces a provider local timestamp or HH:MM clock string to a zero-padded
// HH:MM clock in the supplied location. Providers either repeat the local date or already
// send a clock, so both accepted shapes stay explicit.
func localClock(value string, zone *time.Location) (string, error) {
	if clock, err := time.Parse(model.WeatherClockLayout, value); err == nil && clock.Format(model.WeatherClockLayout) == value {
		return value, nil
	}
	for _, layout := range []string{localMinuteLayout, localSecondLayout} {
		if instant, err := time.ParseInLocation(layout, value, zone); err == nil {
			return instant.In(zone).Format(model.WeatherClockLayout), nil
		}
	}
	if instant, err := time.Parse(time.RFC3339, value); err == nil {
		return instant.In(zone).Format(model.WeatherClockLayout), nil
	}
	return "", fmt.Errorf("invalid local clock")
}
