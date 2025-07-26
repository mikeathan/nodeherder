package utils

import (
	"math"
	"strconv"
)

func TruncateFloat32(num float32, decimalPlaces int) float32 {
	factor := float32(math.Pow10(decimalPlaces))
	truncated := float32(int32(num*factor)) / factor
	return truncated
}

func ParseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 32)
}

func ConvertInt32(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

type EqualityOperator string

const (
	Equals           EqualityOperator = "="
	GreaterThanEqual EqualityOperator = ">="
	LessThanEqual    EqualityOperator = "<="
	GreaterThan      EqualityOperator = ">"
	LessThan         EqualityOperator = "<"
)

var EqualityOperators = map[EqualityOperator]func(any, any) bool{
	Equals: func(v1 any, v2 any) bool {
		return v1 == v2
	},
	GreaterThanEqual: func(v1 any, v2 any) bool {
		return ToFloat(v1) >= ToFloat(v2)
	},
	LessThanEqual: func(v1 any, v2 any) bool {
		return ToFloat(v1) <= ToFloat(v2)
	},
	GreaterThan: func(v1 any, v2 any) bool {
		return ToFloat(v1) > ToFloat(v2)
	},
	LessThan: func(v1 any, v2 any) bool {
		return ToFloat(v1) < ToFloat(v2)
	},
}

func ToFloat(value any) float32 {
	switch v := value.(type) {
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case float32:
		return float32(v)
	default:
		return float32(0)
	}
}
