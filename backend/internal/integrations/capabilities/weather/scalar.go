package weather

// Converts provider scalar values and timestamps into weather-friendly representations.

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"node-herder/internal/integrations/operations"
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
func instant(v any, format string, zone *time.Location) (time.Time, error) {
	if format == "unix" {
		n := number(v)
		if n == nil || math.IsNaN(*n) || math.IsInf(*n, 0) || math.Trunc(*n) != *n || *n < 0 || *n > 7258118400 {
			return time.Time{}, fmt.Errorf("invalid time")
		}
		return time.Unix(int64(*n), 0).UTC(), nil
	}
	if format == "local" {
		return time.ParseInLocation("2006-01-02T15:04", text(v), zone)
	}
	timestamp, err := time.Parse(time.RFC3339, text(v))
	if err == nil {
		return timestamp, nil
	}
	return time.Parse("2006-01-02T15:04Z07:00", text(v))
}
