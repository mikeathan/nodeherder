package operations

// Validates and evaluates JSON Pointers against provider response data.

import (
	"strconv"
	"strings"
)

// ValidPointer validates the bounded RFC 6901 field-location syntax. Empty means the entire
// document; / separates fields, ~1 escapes a literal slash and ~0 escapes a tilde.
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

// Pointer reads an RFC 6901 JSON field location; it is unrelated to Go memory pointers.
// For example, /result/collections/0/name selects the first collection's name. Empty selects
// root; ~1 and ~0 encode literal slash/tilde field names. Array indexes are zero-based canonical
// integers, so 01 and - are invalid. The returned bool distinguishes a missing field (false)
// from an explicitly present JSON null (nil, true). Returned maps/slices are borrowed read-only.
func Pointer(root any, p string) (any, bool) {
	if !ValidPointer(p) {
		return nil, false
	}
	if p == "" {
		return root, true
	}
	for _, part := range strings.Split(p[1:], "/") {
		// Decode escapes in this order: ~01 denotes the literal text ~1, not a slash.
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
