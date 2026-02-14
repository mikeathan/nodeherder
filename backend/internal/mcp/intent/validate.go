package intent

import (
	"fmt"
	"strings"
)

// Rule is a validation function that returns an error if validation fails.
type Rule func(intent *Intent, ctx *DeviceContext) error

// ValidationRules contains the default validation rules.
var ValidationRules = []Rule{
	RequireTargetName,
	RequireMetrics,
	RejectCountEventsForNumeric,
	RejectRangeWithLatestValue,
}

// Validate runs all validation rules against an intent.
func Validate(intent *Intent, ctx *DeviceContext) error {
	for _, rule := range ValidationRules {
		if err := rule(intent, ctx); err != nil {
			return err
		}
	}
	return nil
}

// ValidateWithRules runs custom validation rules against an intent.
func ValidateWithRules(intent *Intent, ctx *DeviceContext, rules []Rule) error {
	for _, rule := range rules {
		if err := rule(intent, ctx); err != nil {
			return err
		}
	}
	return nil
}

// RequireTargetName ensures target_name is provided.
func RequireTargetName(intent *Intent, ctx *DeviceContext) error {
	if strings.TrimSpace(intent.TargetName) == "" {
		return fmt.Errorf("target_name is required")
	}
	return nil
}

// RequireMetrics ensures at least one metric is requested.
func RequireMetrics(intent *Intent, ctx *DeviceContext) error {
	if len(intent.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	return nil
}

// RejectCountEventsForNumeric prevents count_events aggregation on numeric metrics.
func RejectCountEventsForNumeric(intent *Intent, ctx *DeviceContext) error {
	if ctx == nil || intent.Aggregation != "count_events" {
		return nil
	}

	for _, metric := range intent.Metrics {
		if info, ok := ctx.Metrics[metric]; ok {
			if info.Type == "numeric" {
				return fmt.Errorf("count_events aggregation is not valid for numeric metric %q", metric)
			}
		}
	}
	return nil
}

// RejectRangeWithLatestValue prevents using time range scopes with latest_value aggregation.
func RejectRangeWithLatestValue(intent *Intent, ctx *DeviceContext) error {
	if intent.Aggregation != "latest_value" {
		return nil
	}

	rangeScopes := map[string]bool{
		"last_7_days":  true,
		"last_30_days": true,
	}

	if rangeScopes[intent.TimeScope] {
		return fmt.Errorf("latest_value aggregation should not be used with %q time scope", intent.TimeScope)
	}
	return nil
}

// ValidateMetricsExist ensures all requested metrics exist on the device.
func ValidateMetricsExist(intent *Intent, ctx *DeviceContext) error {
	if ctx == nil {
		return nil
	}

	for _, metric := range intent.Metrics {
		if _, ok := ctx.Metrics[metric]; !ok {
			return fmt.Errorf("metric %q not found on device %q", metric, ctx.Name)
		}
	}
	return nil
}

// ValidateAggregationSupported ensures the aggregation is supported for requested metrics.
func ValidateAggregationSupported(intent *Intent, ctx *DeviceContext) error {
	if ctx == nil || intent.Aggregation == "" {
		return nil
	}

	for _, metric := range intent.Metrics {
		info, ok := ctx.Metrics[metric]
		if !ok {
			continue
		}

		supported := false
		for _, agg := range info.Aggregations {
			if agg == intent.Aggregation {
				supported = true
				break
			}
		}

		if !supported {
			return fmt.Errorf("aggregation %q not supported for metric %q", intent.Aggregation, metric)
		}
	}
	return nil
}
