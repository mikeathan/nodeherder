package utils

import "math"

func TruncateFloat32(num float32, decimalPlaces int) float32 {
	factor := float32(math.Pow10(decimalPlaces))
	truncated := float32(int32(num*factor)) / factor
	return truncated
}
