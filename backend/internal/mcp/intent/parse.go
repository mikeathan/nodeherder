package intent

import (
	"encoding/json"
	"fmt"
)

// Parse converts raw JSON to an Intent struct.
func Parse(raw json.RawMessage) (*Intent, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty intent data")
	}

	var intent Intent
	if err := json.Unmarshal(raw, &intent); err != nil {
		return nil, fmt.Errorf("invalid intent JSON: %w", err)
	}

	return &intent, nil
}

// ParseString converts a JSON string to an Intent struct.
func ParseString(s string) (*Intent, error) {
	return Parse(json.RawMessage(s))
}
