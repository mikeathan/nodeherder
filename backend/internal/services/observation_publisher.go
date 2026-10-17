package services

import (
	"context"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
	"time"
)

type pendingObservation struct {
	expose    *devices.Entity
	value     any
	threshold float64
	due       time.Time
}

// ObservationPublisher owns live-output state, never raw device state or automation.
// One timer per device flushes the latest qualifying reading at the original deadline.
// Its mutex serializes immediate/delayed emission and cancellation; incoming packets
// replace pending values without allocating timers or postponing publication.
type ObservationPublisher struct {
	mu           sync.Mutex
	clock        utils.Clock
	filter       *devices.ObservationFilter
	publishedAt  map[string]time.Time
	pending      map[string]pendingObservation
	timer        utils.Timer
	deadline     time.Time
	generation   uint64
	closed       bool
	stopShutdown func() bool
	ctx          context.Context
	device       *devices.Device
	policy       func(*devices.Entity) settings.ReportingPolicy
	emit         func(*devices.UpdatePackage)
}

// NewObservationPublisher binds publication lifetime to the hub context.
func NewObservationPublisher(ctx context.Context, device *devices.Device, clock utils.Clock, policy func(*devices.Entity) settings.ReportingPolicy, emit func(*devices.UpdatePackage)) *ObservationPublisher {
	p := &ObservationPublisher{ctx: ctx, device: device, clock: clock, policy: policy, emit: emit,
		filter: devices.NewObservationFilter(), publishedAt: map[string]time.Time{}, pending: map[string]pendingObservation{}}
	for _, expose := range device.Exposes {
		p.filter.Prime(expose, expose.Data.Value())
		p.publishedAt[expose.Name] = clock.Now()
	}
	// A cancelled context can invoke Close immediately; publish the stop handle
	// under the same lock so registration cannot race with retirement.
	p.mu.Lock()
	p.stopShutdown = context.AfterFunc(ctx, p.Close)
	p.mu.Unlock()
	return p
}

// Prime aligns the baseline with a full snapshot already sent to observers.
func (p *ObservationPublisher) Prime(expose *devices.Entity, value any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.filter.Prime(expose, value)
	p.publishedAt[expose.Name] = p.clock.Now()
}

// Publish consumes a candidate update after raw state and automation are current.
// Callbacks run under the output mutex and must not re-enter this publisher.
func (p *ObservationPublisher) Publish(update *devices.UpdatePackage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.publishLocked(update)
}

func (p *ObservationPublisher) publishLocked(update *devices.UpdatePackage) {
	if p.closed || p.ctx.Err() != nil {
		return
	}
	now := p.clock.Now()
	for name, value := range update.Data {
		expose, ok := p.device.GetExpose(name)
		if !ok {
			delete(update.Data, name)
			continue
		}
		policy := p.policy(expose)
		if !p.filter.Allows(expose, value, policy.MinimumChange) {
			delete(p.pending, name)
			delete(update.Data, name)
			continue
		}
		due := p.publishedAt[name].Add(policy.LiveInterval)
		if devices.IsNumericObservation(expose, value) && p.filter.HasBaseline(name) && now.Before(due) {
			p.pending[name] = pendingObservation{expose: expose, value: value, threshold: policy.MinimumChange, due: due}
			delete(update.Data, name)
			continue
		}
		delete(p.pending, name)
		p.filter.Accept(expose, value, policy.MinimumChange)
		p.publishedAt[name] = now
	}
	p.scheduleLocked(now)
	p.emitLocked(update)
}

func (p *ObservationPublisher) emitLocked(update *devices.UpdatePackage) {
	if p.device.IsAvailable() && (update.HasData() || update.Availability != "") && p.emit != nil {
		p.emit(update)
	}
}

func (p *ObservationPublisher) flush(generation uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if generation != p.generation {
		return
	}
	p.timer = nil
	p.deadline = time.Time{}
	if p.closed || p.ctx.Err() != nil {
		return
	}
	update := devices.NewUpdatePackage(p.device.Id)
	// Another field may have reported since this observation was queued. Carry the
	// current raw last-seen timestamp so a delayed field never moves it backwards.
	if seen, err := p.device.LastSeenTime(); err == nil {
		update.LastSeen = seen.Format(time.RFC3339Nano)
	}
	now := p.clock.Now()
	for name, item := range p.pending {
		if now.Before(item.due) {
			continue
		}
		delete(p.pending, name)
		if p.filter.Accept(item.expose, item.value, item.threshold) {
			update.Data[name] = item.value
			p.publishedAt[name] = now
		}
	}
	p.scheduleLocked(now)
	p.emitLocked(update)
}

func (p *ObservationPublisher) scheduleLocked(now time.Time) {
	var earliest time.Time
	for _, item := range p.pending {
		if earliest.IsZero() || item.due.Before(earliest) {
			earliest = item.due
		}
	}
	if earliest.Equal(p.deadline) {
		return
	}
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	p.generation++
	p.deadline = earliest
	if !earliest.IsZero() {
		generation := p.generation
		p.timer = p.clock.AfterFunc(earliest.Sub(now), func() { p.flush(generation) })
	}
}

// Reconfigure applies new rules to retained readings without requiring another
// sensor report. Unrelated settings edits therefore cannot discard the final value.
// The last publication time remains the interval anchor, so edits never starve it.
func (p *ObservationPublisher) Reconfigure() {
	p.mu.Lock()
	defer p.mu.Unlock()
	update := devices.NewUpdatePackage(p.device.Id)
	if seen, err := p.device.LastSeenTime(); err == nil {
		update.LastSeen = seen.Format(time.RFC3339Nano)
	}
	for name, item := range p.pending {
		update.Data[name] = item.value
	}
	p.clearLocked()
	p.publishLocked(update)
}

// Clear drops deferred output on disable/offline transitions. The published
// baseline stays intact; the next accepted report is evaluated with current policy.
func (p *ObservationPublisher) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clearLocked()
}

func (p *ObservationPublisher) clearLocked() {
	p.generation++
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	clear(p.pending)
	p.deadline = time.Time{}
}

// Close releases pending output and its shutdown registration. Removed devices
// must not remain reachable through the hub context until the hub itself stops.
func (p *ObservationPublisher) Close() { _ = p.closeAfter(nil) }

// closeAfter is the retirement fence used by the lifetime owner. Output callbacks
// also hold mu, so persistence deletion cannot race an older snapshot write. Failed
// deletion leaves timers and publication active; success closes the generation.
// The action must not re-enter this publisher.
func (p *ObservationPublisher) closeAfter(action func() error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if action != nil {
		if err := action(); err != nil {
			return err
		}
	}
	p.closed = true
	p.clearLocked()
	stop := p.stopShutdown
	p.stopShutdown = nil
	// AfterFunc stop unregisters without waiting for an already-started callback.
	if stop != nil {
		stop()
	}
	return nil
}
