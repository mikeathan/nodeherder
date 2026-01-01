package query

import (
	"node-herder/internal/metrics/domain"
	"sort"
)

func ApplyLimitSortBy(expose domain.ExposeResult, limit int, sortDesc bool) {
	if expose == nil {
		return
	}

	switch r := expose.(type) {
	case *domain.ExposeNumericMetricsResult:
		applyLimitSortNumeric(r, limit, sortDesc)
	case *domain.ExposeBinaryEventsResult:
		applyLimitSortBinary(r, limit, sortDesc)
	case *domain.ExposeTimeRangeMetricsResult:
		applyLimitSortTimeRange(r, limit, sortDesc)
	}
}

func applyLimitSortNumeric(result *domain.ExposeNumericMetricsResult, limit int, sortDesc bool) {
	if sortDesc {
		sort.Slice(result.Data, func(i, j int) bool {
			return result.Data[i].X > result.Data[j].X
		})
	}

	if limit > 0 && len(result.Data) > limit {
		result.Data = result.Data[:limit]
	}
}

func applyLimitSortBinary(result *domain.ExposeBinaryEventsResult, limit int, sortDesc bool) {
	if sortDesc {
		sort.Slice(result.Data, func(i, j int) bool {
			return result.Data[i].Timestamp > result.Data[j].Timestamp
		})
	}

	if limit > 0 && len(result.Data) > limit {
		result.Data = result.Data[:limit]
	}
}

func applyLimitSortTimeRange(result *domain.ExposeTimeRangeMetricsResult, limit int, sortDesc bool) {
	if sortDesc {
		sort.Slice(result.Data, func(i, j int) bool {
			return result.Data[i].Y[0] > result.Data[j].Y[0]
		})
	}

	if limit > 0 && len(result.Data) > limit {
		result.Data = result.Data[:limit]
	}
}
