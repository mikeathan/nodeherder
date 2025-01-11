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
