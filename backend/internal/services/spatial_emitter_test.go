package services

import (
	"sync"
	"testing"
	"time"

	"node-herder/models/devices"
	"node-herder/utils"
)

func emitterTargets(x float64) []devices.SpatialTarget {
	return []devices.SpatialTarget{{
		ID:     1,
		Active: true,
		X:      &devices.SpatialAxis{Value: x, Unit: "mm"},
		Y:      &devices.SpatialAxis{Value: 2000, Unit: "mm"},
	}}
}

// fakeClock keeps timer behaviour deterministic without importing the shared mocks package,
// which would cycle back into this package through the automation engine.
type fakeClock struct {
	utils.RealClock
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

func newFakeClock(now time.Time) *fakeClock {
	return &fakeClock{now: now}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) AfterFunc(d time.Duration, f func()) utils.Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	timer := &fakeTimer{clock: c, remaining: d, callback: f}
	c.timers = append(c.timers, timer)
	return timer
}

// Advance moves time forward and runs due callbacks outside the clock lock, matching a real
// timer: callbacks take their own locks and may arm further timers.
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	var due []*fakeTimer
	for _, timer := range c.timers {
		if timer.stopped || timer.fired {
			continue
		}
		timer.remaining -= d
		if timer.remaining <= 0 {
			timer.fired = true
			due = append(due, timer)
		}
	}
	c.mu.Unlock()
	for _, timer := range due {
		timer.callback()
	}
}

// PassTime moves the clock without running due callbacks, modelling a timer goroutine that
// has not been scheduled yet while the device lane keeps publishing.
func (c *fakeClock) PassTime(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

type fakeTimer struct {
	clock     *fakeClock
	remaining time.Duration
	callback  func()
	stopped   bool
	fired     bool
}

func (t *fakeTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	if t.stopped || t.fired {
		return false
	}
	t.stopped = true
	return true
}

func (t *fakeTimer) Reset(d time.Duration) bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	wasActive := !t.stopped && !t.fired
	t.stopped, t.fired, t.remaining = false, false, d
	return wasActive
}

func newTestEmitter(frames *[]*devices.SpatialFrame) (*spatialEmitter, *fakeClock, time.Time) {
	base := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	clock := newFakeClock(base)
	emitter := newSpatialEmitter(clock, "device", func(frame *devices.SpatialFrame) {
		*frames = append(*frames, frame)
	})
	return emitter, clock, base
}

func TestSpatialEmitterEmitsFirstFrameAndDropsRepeats(t *testing.T) {
	var frames []*devices.SpatialFrame
	emitter, clock, base := newTestEmitter(&frames)
	t.Cleanup(emitter.Close)

	emitter.Publish(emitterTargets(100))
	emitter.Publish(emitterTargets(100))
	clock.Advance(time.Second)
	emitter.Publish(emitterTargets(100))

	if len(frames) != 1 {
		t.Fatalf("repeat emissions = %d, want 1", len(frames))
	}
	if !frames[0].At.Equal(base) || frames[0].Id != "device" {
		t.Fatalf("frame envelope = %+v", frames[0])
	}
}

func TestSpatialEmitterRateLimitsWithNewestWins(t *testing.T) {
	var frames []*devices.SpatialFrame
	emitter, clock, start := newTestEmitter(&frames)
	t.Cleanup(emitter.Close)

	emitter.Publish(emitterTargets(100))
	clock.Advance(40 * time.Millisecond)
	emitter.Publish(emitterTargets(200))
	emitter.Publish(emitterTargets(300))
	if len(frames) != 1 {
		t.Fatalf("emitted inside the interval: %d", len(frames))
	}

	clock.Advance(60 * time.Millisecond)
	if len(frames) != 2 {
		t.Fatalf("trailing flush missing: %d", len(frames))
	}
	if got := frames[1].Targets[0].X.Value; got != 300 {
		t.Fatalf("flushed value = %v, want the newest (300)", got)
	}
	if !frames[1].At.Equal(start.Add(spatialFrameInterval)) {
		t.Fatalf("flush time = %v, want %v", frames[1].At, start.Add(spatialFrameInterval))
	}
}

func TestSpatialEmitterDropsPendingFlushWhenContentReverts(t *testing.T) {
	var frames []*devices.SpatialFrame
	emitter, clock, _ := newTestEmitter(&frames)
	t.Cleanup(emitter.Close)

	emitter.Publish(emitterTargets(100))
	clock.Advance(10 * time.Millisecond)
	emitter.Publish(emitterTargets(200))
	emitter.Publish(emitterTargets(100))
	clock.Advance(time.Second)

	if len(frames) != 1 {
		t.Fatalf("stale pending frame was emitted: %d", len(frames))
	}
}

func TestSpatialEmitterDiscardForgetsContent(t *testing.T) {
	var frames []*devices.SpatialFrame
	emitter, clock, _ := newTestEmitter(&frames)
	t.Cleanup(emitter.Close)

	emitter.Publish(emitterTargets(100))
	clock.Advance(10 * time.Millisecond)
	emitter.Publish(emitterTargets(200))
	emitter.Discard()
	clock.Advance(time.Second)
	if len(frames) != 1 {
		t.Fatalf("discarded frame was emitted: %d", len(frames))
	}

	// A device returning from offline must render even when nothing moved meanwhile.
	emitter.Publish(emitterTargets(200))
	if len(frames) != 2 {
		t.Fatalf("frame after discard = %d, want 2", len(frames))
	}
}

func TestSpatialEmitterNewerFrameCancelsPendingFlush(t *testing.T) {
	var frames []*devices.SpatialFrame
	emitter, clock, start := newTestEmitter(&frames)
	t.Cleanup(emitter.Close)

	emitter.Publish(emitterTargets(100))
	clock.Advance(40 * time.Millisecond)
	emitter.Publish(emitterTargets(200))
	// The flush timer is due at +100 ms but has not run; a newer frame arrives first.
	clock.PassTime(60 * time.Millisecond)
	emitter.Publish(emitterTargets(300))
	clock.Advance(time.Second)

	if len(frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(frames))
	}
	if value := frames[1].Targets[0].X.Value; value != 300 {
		t.Fatalf("last frame carried %v, want the newer 300", value)
	}
	if !frames[1].At.Equal(start.Add(spatialFrameInterval)) {
		t.Fatalf("last frame time = %v, want %v", frames[1].At, start.Add(spatialFrameInterval))
	}
}

func TestSpatialEmitterCloseStopsEmission(t *testing.T) {
	var frames []*devices.SpatialFrame
	emitter, clock, _ := newTestEmitter(&frames)

	emitter.Publish(emitterTargets(100))
	clock.Advance(10 * time.Millisecond)
	emitter.Publish(emitterTargets(200))
	emitter.Close()
	clock.Advance(time.Second)
	emitter.Publish(emitterTargets(300))

	if len(frames) != 1 {
		t.Fatalf("closed emitter produced %d frames, want 1", len(frames))
	}
}
