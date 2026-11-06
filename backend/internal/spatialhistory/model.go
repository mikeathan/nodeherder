// Package spatialhistory stores low-resolution room positions recorded from accepted spatial
// frames. It is gated by the source device's history setting and keeps its own 7-day retention.
package spatialhistory

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"sort"
	"time"
)

const (
	// Retention is intentionally shorter than metrics history: positions are bulkier and reveal
	// occupancy patterns, and the recorder is a bounded convenience rather than an archive.
	Retention = 7 * 24 * time.Hour
	// MaxPointsPerDay bounds a misbehaving sensor; the session is marked capped instead of growing.
	MaxPointsPerDay = 3000
	chunkInterval   = 5 * time.Minute
	chunkPoints     = 300
	minSampleGap    = time.Second
	movementGateMM  = 300
	quantizeMM      = 50
	mergeDistanceMM = 350
)

// Point is one quantised target position. At is Unix milliseconds.
type Point struct {
	At   int64 `json:"at"`
	Slot int   `json:"slot"`
	X    int32 `json:"x"`
	Y    int32 `json:"y"`
}

// Chunk is one bounded write: a session slice with ordered points.
type Chunk struct {
	Session string  `json:"session"`
	From    int64   `json:"from"`
	To      int64   `json:"to"`
	Points  []Point `json:"points"`
	Capped  bool    `json:"capped,omitempty"`
}

// Session is the player's summary of the chunks sharing an id.
type Session struct {
	ID     string `json:"id"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Points int    `json:"points"`
	Slots  []int  `json:"slots"`
	Capped bool   `json:"capped"`
}

// Cell is one aggregated heatmap bucket in millimetres.
type Cell struct {
	X     int32 `json:"x"`
	Y     int32 `json:"y"`
	Count int   `json:"count"`
}

func quantize(value float64) int32 {
	return int32(math.Round(value/quantizeMM) * quantizeMM)
}

func distanceMM(a, b Point) float64 {
	return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y))
}

func newSessionID() string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buffer)
}

// mergeTargets collapses duplicate slot reports that describe the same physical return. The
// recorder never treats a slot as an identity; overlapping slots would otherwise double points.
type rawTarget struct {
	slot   int
	x, y   float64
	active bool
}

func mergeTargets(targets []rawTarget) []rawTarget {
	kept := make([]rawTarget, 0, len(targets))
	for _, target := range targets {
		if !target.active {
			continue
		}
		duplicate := false
		for _, existing := range kept {
			if math.Hypot(target.x-existing.x, target.y-existing.y) < mergeDistanceMM {
				duplicate = true
				break
			}
		}
		if !duplicate {
			kept = append(kept, target)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].slot < kept[j].slot })
	return kept
}
