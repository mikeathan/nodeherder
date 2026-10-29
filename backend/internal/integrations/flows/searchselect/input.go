package searchselect

// Normalizes search input and recognizes supported UK postcode queries.

import (
	"regexp"
	"strings"

	model "node-herder/models/integrations"
)

var ukPostcode = regexp.MustCompile(`^(GIR0AA|[A-Z]{1,2}[0-9][A-Z0-9]?[0-9][A-Z]{2})$`)

// NormalizeInput applies the declared search-input type instead of assuming every search
// is a postcode. Text preserves internal spacing/case; UK postcode formatting is explicit.
func NormalizeInput(value string, input model.Input) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 2048 {
		return "", model.Fail("invalid_input", "Enter a search value of at most 2048 characters.")
	}
	switch input.Type {
	case "text":
		return value, nil
	case "postcode":
		compact := strings.ToUpper(strings.Join(strings.Fields(value), ""))
		if !ukPostcode.MatchString(compact) {
			return "", model.Fail("invalid_input", "Enter a UK postcode.")
		}
		return compact[:len(compact)-3] + " " + compact[len(compact)-3:], nil
	default:
		return "", model.Fail("invalid_definition", "The search step requires a text or postcode input.")
	}
}

// IsUKPostcode identifies full UK postcodes without changing place-name searches.
func IsUKPostcode(value string) bool {
	return ukPostcode.MatchString(strings.ToUpper(strings.Join(strings.Fields(value), "")))
}
