package services

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"node-herder/models/devices"
	"node-herder/utils"
)

// emptyFrameKey is an impossible frame key, used so the first cleared frame is not mistaken
// for "nothing changed" before any frame was emitted.
const emptyFrameKey = "\x00empty"

// spatialFrameInterval bounds how often one device emits a spatial frame. Device providers
// report at most every 100 ms, so the cap only matters for faster or chatty devices.
const spatialFrameInterval = 100 * time.Millisecond

// spatialEmitter rate-limits and deduplicates spatial frames for one device lifetime.
// Emission runs on the device lane while the trailing-flush timer runs on its own
// goroutine, so every field is guarded by mu. The emit callback runs under that mutex and
// must not re-enter the lifetime.
type spatialEmitter struct {
	mu         sync.Mutex
	clock      utils.Clock
	deviceId   string
	emit       func(*devices.SpatialFrame)
	interval   time.Duration
	lastKey    string
	lastAt     time.Time
	pendingKey string
	pending    []devices.SpatialTarget
	timer      utils.Timer
	closed     bool
}

// newSpatialEmitter creates an idle emitter for one device. A nil emit function makes every
// publication a no-op, which keeps tests and unconsumed devices cheap.
func newSpatialEmitter(clock utils.Clock, deviceId string, emit func(*devices.SpatialFrame)) *spatialEmitter {
	return &spatialEmitter{
		clock:    clock,
		deviceId: deviceId,
		emit:     emit,
		interval: spatialFrameInterval,
		lastKey:  emptyFrameKey,
	}
}

// Publish offers one complete frame. Identical content is dropped, a frame outside the rate
// limit replaces any pending one and is flushed when the interval elapses, so the newest
// position is never lost to throttling.
func (e *spatialEmitter) Publish(targets []devices.SpatialTarget) {
	key := spatialFrameKey(targets)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return
	}
	if key == e.lastKey {
		// The offered content is already current; a queued flush for older content is stale.
		e.clearPendingLocked()
		return
	}
	now := e.clock.Now()
	if e.lastAt.IsZero() || now.Sub(e.lastAt) >= e.interval {
		// A pending flush could fire after this newer frame and publish an older position with a
		// later timestamp, so it must be cancelled before emitting.
		e.clearPendingLocked()
		e.emitLocked(now, key, targets)
		return
	}
	e.pendingKey = key
	e.pending = targets
	e.scheduleLocked(now)
}

// Discard drops pending work and forgets the last content, so the next offer emits even when
// it matches what was shown before an offline period. Use when a device becomes unavailable.
func (e *spatialEmitter) Discard() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clearPendingLocked()
	e.lastKey = emptyFrameKey
}

// Close stops the trailing flush permanently. Call from the owning lifetime's Close.
func (e *spatialEmitter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	e.clearPendingLocked()
}

func (e *spatialEmitter) emitLocked(now time.Time, key string, targets []devices.SpatialTarget) {
	e.lastKey = key
	e.lastAt = now
	if e.emit == nil {
		return
	}
	e.emit(&devices.SpatialFrame{Id: e.deviceId, At: now, Targets: targets})
}

// scheduleLocked arms the single flush timer at the end of the current interval. Frames are
// newest-wins, so one timer is enough regardless of how many offers arrive meanwhile.
func (e *spatialEmitter) scheduleLocked(now time.Time) {
	if e.timer != nil {
		return
	}
	// Publish only schedules while the interval has not elapsed, so the delay is positive.
	e.timer = e.clock.AfterFunc(e.lastAt.Add(e.interval).Sub(now), e.flush)
}

func (e *spatialEmitter) flush() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return
	}
	e.timer = nil
	if e.pending == nil {
		return
	}
	// time.AfterFunc never fires before the requested delay, so the interval has elapsed.
	now := e.clock.Now()
	key, targets := e.pendingKey, e.pending
	e.pendingKey, e.pending = "", nil
	e.emitLocked(now, key, targets)
}

func (e *spatialEmitter) clearPendingLocked() {
	if e.timer != nil {
		e.timer.Stop()
		e.timer = nil
	}
	e.pendingKey, e.pending = "", nil
}

// spatialFrameKey builds a comparable summary of frame content so repeated reports of an
// unchanged position do not produce traffic. Axis values use exact formatting: a different
// reading must never collapse into an identical key.
func spatialFrameKey(targets []devices.SpatialTarget) string {
	var builder strings.Builder
	for _, target := range targets {
		builder.WriteString(strconv.Itoa(target.ID))
		if target.Active {
			builder.WriteString(":1:")
		} else {
			builder.WriteString(":0:")
		}
		writeSpatialAxis(&builder, target.X)
		builder.WriteByte('|')
		writeSpatialAxis(&builder, target.Y)
		builder.WriteByte(';')
	}
	return builder.String()
}

func writeSpatialAxis(builder *strings.Builder, axis *devices.SpatialAxis) {
	if axis == nil {
		builder.WriteByte('-')
		return
	}
	builder.WriteString(strconv.FormatFloat(axis.Value, 'g', -1, 64))
	builder.WriteByte(':')
	builder.WriteString(axis.Unit)
}
