package operations

import (
	"strconv"
	"strings"
)

func ValidPointer(p string) bool {
	if p == "" {
		return true
	}
	if len(p) > 256 || p[0] != '/' {
		return false
	}
	for i := 0; i < len(p); i++ {
		if p[i] == '~' {
			i++
			if i == len(p) || (p[i] != '0' && p[i] != '1') {
				return false
			}
		}
	}
	return true
}

// Pointer reads RFC 6901 paths from decoded JSON, preserving missing versus null.
func Pointer(root any, p string) (any, bool) {
	if !ValidPointer(p) {
		return nil, false
	}
	if p == "" {
		return root, true
	}
	for _, part := range strings.Split(p[1:], "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch x := root.(type) {
		case map[string]any:
			var ok bool
			root, ok = x[part]
			if !ok {
				return nil, false
			}
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || strconv.Itoa(i) != part || i >= len(x) {
				return nil, false
			}
			root = x[i]
		default:
			return nil, false
		}
	}
	return root, true
}
