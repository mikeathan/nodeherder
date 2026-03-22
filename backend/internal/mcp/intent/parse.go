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

	// Support unwrapping if a client incorrectly sends `{"name": "tool", "arguments": {...}}`
	var wrapper struct {
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Arguments != nil {
		raw = wrapper.Arguments
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
