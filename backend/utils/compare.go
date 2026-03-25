package utils

import (
	"fmt"
	"strings"
)

func CompareValuesCaseInsensitive(val1, val2 any) bool {
	// FAST PATH: If both are strings, use strings.EqualFold (Case-Insensitive, zero allocation)
	if str1, ok1 := val1.(string); ok1 {
		if str2, ok2 := val2.(string); ok2 {
			return strings.EqualFold(str1, str2)
		}
	}

	// SLOW PATH FALLBACK: If types mismatch (e.g., int vs string) or aren't strings
	// Safe conversion using fmt.Sprintf to compare string representations
	return fmt.Sprintf("%v", val1) == fmt.Sprintf("%v", val2)
}
