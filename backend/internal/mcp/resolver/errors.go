package resolver

import "fmt"

// AmbiguousDeviceError is returned when multiple devices match a target name.
type AmbiguousDeviceError struct {
	TargetName string
	Candidates []Candidate
}

func (e *AmbiguousDeviceError) Error() string {
	return fmt.Sprintf("ambiguous device: %q matches %d devices", e.TargetName, len(e.Candidates))
}

// DeviceNotFoundError is returned when no device matches a target name.
type DeviceNotFoundError struct {
	TargetName string
}

func (e *DeviceNotFoundError) Error() string {
	return fmt.Sprintf("device not found: %q", e.TargetName)
}

// Candidate represents a device match candidate with its score.
type Candidate struct {
	DeviceID string  `json:"device_id"`
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
}
