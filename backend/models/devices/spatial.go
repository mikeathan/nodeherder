package devices

import (
	"strings"
	"time"
)

// SpatialFrame is a complete, synchronized snapshot of a device's mapped target slots.
// It is presentation data derived from accepted device state: it never replaces confirmed
// state, carries no identity, and never predicts movement. Consumers must treat frame
// silence as "nothing changed", not as an empty room.
type SpatialFrame struct {
	// Id is the existing device identifier.
	Id string `json:"id"`
	// At is server time when the frame was produced, so a consumer can age the frame
	// without depending on arrival time.
	At time.Time `json:"at"`
	// Targets holds one entry per complete mapping row, at most three. It is empty, never
	// nil, when no row is complete, so a consumer can replace its view directly.
	Targets []SpatialTarget `json:"targets"`
}

// SpatialTarget is one mapped slot in a frame. ID is the mapping row number (1-3) and is
// not an identity: firmware reuses slots between people. X and Y are nil when that axis has
// no usable reading; Active always reports the mapped flag so consumers can distinguish
// "cleared" from "unreadable".
type SpatialTarget struct {
	ID     int          `json:"id"`
	Active bool         `json:"active"`
	X      *SpatialAxis `json:"x,omitempty"`
	Y      *SpatialAxis `json:"y,omitempty"`
}

// SpatialAxis is one coordinate reading in the unit declared by its entity. Values are
// passed through unconverted and unrounded; consumers convert for display.
type SpatialAxis struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// spatialLengthUnits are the length units a mapped coordinate may declare. It mirrors the
// browser's length table: an axis in any other unit is not published because a consumer
// could not place it in the plane.
var spatialLengthUnits = map[string]struct{}{
	"mm": {},
	"cm": {},
	"m":  {},
}

// IsSpatialLengthUnit reports whether unit is a length unit that spatial frames can carry.
// Matching ignores case and surrounding spaces, as device metadata often does.
func IsSpatialLengthUnit(unit string) bool {
	_, ok := spatialLengthUnits[strings.ToLower(strings.TrimSpace(unit))]
	return ok
}
