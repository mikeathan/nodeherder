package query

import (
	"node-herder/internal/metrics/domain"
	"node-herder/utils"
)

func MatchesFilters(raw []byte, filters []domain.MetricFilter, exposeType string) bool {
	if len(filters) == 0 {
		return true
	}

	switch exposeType {
	case "numeric":
		var v float64
		if err := utils.DecodeGobValue(raw, &v); err != nil {
			return false
		}
		return matchNumericFilters(v, filters)

	case "binary":
		var v bool
		if err := utils.DecodeGobValue(raw, &v); err != nil {
			return false
		}
		return matchBoolFilters(v, filters)

	case "enum":
		var v string
		if err := utils.DecodeGobValue(raw, &v); err != nil {
			return false
		}
		return matchStringFilters(v, filters)

	default:
		// Unknown expose type
		return false
	}
}

func matchNumericFilters(v float64, filters []domain.MetricFilter) bool {
	for _, f := range filters {
		fv, ok := f.Value.(float64)
		if !ok {
			return false
		}

		switch f.Op {
		case domain.OpGreaterThan:
			if v <= fv {
				return false
			}
		case domain.OpLessThan:
			if v >= fv {
				return false
			}
		case domain.OpEquals:
			if v != fv {
				return false
			}
		case domain.OpNotEquals:
			if v == fv {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func matchStringFilters(v string, filters []domain.MetricFilter) bool {
	for _, f := range filters {
		fv, ok := f.Value.(string)
		if !ok {
			return false
		}

		switch f.Op {
		case domain.OpEquals:
			if v != fv {
				return false
			}
		case domain.OpNotEquals:
			if v == fv {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func matchBoolFilters(v bool, filters []domain.MetricFilter) bool {
	for _, f := range filters {
		fv, ok := f.Value.(bool)
		if !ok {
			return false
		}

		switch f.Op {
		case domain.OpEquals:
			if v != fv {
				return false
			}
		case domain.OpNotEquals:
			if v == fv {
				return false
			}
		default:
			return false
		}
	}
	return true
}
