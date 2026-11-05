package spatial

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
)

// Derived publication limits: occupancy and counts publish on change with a short guard
// against bursts; slow counters publish on a fixed cadence so history cannot grow at frame
// rate (NFR-01).
const (
	defaultQueueSize    = 64
	publishGap          = 250 * time.Millisecond
	counterInterval     = 10 * time.Second
	workerTick          = time.Second
	trackFreshnessLimit = time.Second
	// maxCounterStep bounds how much counter time one flush can add, so a suspended hub does
	// not publish a huge dwell or empty jump on resume.
	maxCounterStep = 3 * counterInterval
)

// DeviceSink receives derived device lifecycle and reports. The hub controller implements it
// with the device processor, so derived values reach automation, metric and WebSocket
// consumers through the same processor seam as provider reports. The engine serializes all
// calls under one lock and is the derived device's only writer. Implementations must be safe
// to call from the engine worker.
type DeviceSink interface {
	Register(device *devices.Device) error
	Update(id string, payload map[string]interface{}) error
	Offline(id string) error
	Remove(id string) error
}

// Config supplies the saved spatial settings the engine needs. It is read on every sync, so a
// saved mapping or zone edit takes effect without a restart.
type Config interface {
	Mapped(deviceID string) bool
	Zones(deviceID string) []settings.SpatialZone
}

// Deps are the engine's collaborators. Source returns the current inventory of a source
// device, which supplies the derived device name and power metadata.
type Deps struct {
	// Context bounds the engine lifetime; nil uses context.Background.
	Context context.Context
	Config  Config
	Source  func(id string) (*devices.Device, bool)
	Sink    DeviceSink
	Clock   utils.Clock
	Queue   int
	// Dropped observes frames discarded because the worker was busy.
	Dropped func()
}

type frameJob struct {
	id      string
	samples []Sample
	at      time.Time
}

type sourceState struct {
	name          string
	zones         []settings.SpatialZone
	deviceKey     string
	geometryKey   string
	tracker       *Tracker
	registered    bool
	available     bool
	lastImmediate map[string]interface{}
	lastSignals   Signals
	lastSignalsAt time.Time
	lastPublish   time.Time
	lastCounters  time.Time
	pending       map[string]interface{}
}

// Engine derives presence signals for every mapped source device. One worker consumes a
// bounded queue, so a slow store or automation never blocks a device lane; frames that arrive
// while it is busy are dropped rather than queued without limit.
type Engine struct {
	deps   Deps
	mu     sync.Mutex
	states map[string]*sourceState
	// sinkMu serializes device lifecycle and report delivery, so a removal can never race a
	// report already in flight. It is always taken outside mu.
	sinkMu  sync.Mutex
	queue   chan frameJob
	ctx     context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
	started bool
	closed  bool
}

// NewEngine creates an idle engine. Call Start before supplying frames.
func NewEngine(deps Deps) *Engine {
	queueSize := deps.Queue
	if queueSize <= 0 {
		queueSize = defaultQueueSize
	}
	parent := deps.Context
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &Engine{deps: deps, states: map[string]*sourceState{}, queue: make(chan frameJob, queueSize), ctx: ctx, cancel: cancel}
}

// Start launches the single worker once. Later calls are no-ops; a closed engine ignores it.
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.started || e.closed {
		return
	}
	e.started = true
	e.workers.Add(1)
	go e.run()
}

// Observe accepts one synchronized frame without blocking the device lane. A full queue
// drops the frame because presence is best effort and the next frame supersedes it.
func (e *Engine) Observe(frame *devices.SpatialFrame) {
	if frame == nil || frame.Id == "" || isDerivedDeviceID(frame.Id) {
		return
	}
	e.mu.Lock()
	accept := !e.closed && e.started
	e.mu.Unlock()
	if !accept {
		return
	}
	at := frame.At
	if at.IsZero() {
		at = e.deps.Clock.Now()
	}
	select {
	case e.queue <- frameJob{id: frame.Id, samples: samplesFromFrame(frame), at: at}:
	default:
		if e.deps.Dropped != nil {
			e.deps.Dropped()
		}
	}
}

// samplesFromFrame converts axes to millimetres and keeps only active slots with both axes.
// The frame contract already validated units; an unexpected unit drops the slot rather than
// guessing a position.
func samplesFromFrame(frame *devices.SpatialFrame) []Sample {
	samples := make([]Sample, 0, len(frame.Targets))
	for _, target := range frame.Targets {
		if !target.Active || target.X == nil || target.Y == nil {
			continue
		}
		x, okX := millimetres(target.X.Value, strings.ToLower(strings.TrimSpace(target.X.Unit)))
		y, okY := millimetres(target.Y.Value, strings.ToLower(strings.TrimSpace(target.Y.Unit)))
		if !okX || !okY {
			continue
		}
		samples = append(samples, Sample{Slot: target.ID, X: x, Y: y})
	}
	return samples
}

// Sync reconciles the derived device with the saved settings: it registers, rebuilds or
// retires it. Safe to call from configuration and device callbacks.
func (e *Engine) Sync(sourceID string) {
	if sourceID == "" || isDerivedDeviceID(sourceID) {
		return
	}
	e.mu.Lock()
	closed := e.closed
	e.mu.Unlock()
	if closed {
		return
	}

	source, ok := e.deps.Source(sourceID)
	mapped := ok && e.deps.Config.Mapped(sourceID)
	zones := e.deps.Config.Zones(sourceID)
	if !mapped || len(zones) == 0 {
		e.Remove(sourceID)
		return
	}

	deviceKey := source.FriendlyName + "\x00" + zoneSignature(zones)
	geometryKey := zoneSignature(zones)
	available := source.GetAvailability() != devices.OfflineAvailability

	e.mu.Lock()
	state := e.states[sourceID]
	if state == nil {
		state = &sourceState{tracker: NewTracker()}
		e.states[sourceID] = state
	}
	geometryChanged := state.geometryKey != "" && state.geometryKey != geometryKey
	needsRegister := !state.registered || state.deviceKey != deviceKey
	wasAvailable := state.available
	availabilityChanged := wasAvailable != available
	if geometryChanged || availabilityChanged {
		state.tracker = NewTracker()
		state.lastImmediate = nil
		state.lastSignals = Signals{}
		state.lastSignalsAt = time.Time{}
		state.pending = nil
	}
	state.name = source.FriendlyName
	state.zones = zones
	state.deviceKey = deviceKey
	state.geometryKey = geometryKey
	state.available = available
	restoreAt, restoreSignals := state.lastSignalsAt, state.lastSignals
	e.mu.Unlock()

	if needsRegister {
		// Hold the sink lock across the whole decision and registration, so a concurrent Sync
		// cannot retire the generation this one is about to create.
		e.sinkMu.Lock()
		e.mu.Lock()
		stillNeeded := !state.registered || state.deviceKey != deviceKey
		closed := e.closed
		e.mu.Unlock()
		if !stillNeeded || closed {
			e.sinkMu.Unlock()
			needsRegister = false
		}
	}
	if needsRegister {
		err := e.deps.Sink.Register(BuildDevice(source, zones))
		if err == nil {
			e.mu.Lock()
			state.registered = true
			e.mu.Unlock()
		}
		e.sinkMu.Unlock()
		if err != nil {
			utils.LogErrorf("presence: register %s: %v", DeviceID(sourceID), err)
			return
		}
		// A re-registration starts from unknown values. When the geometry is unchanged, restore
		// the last computed values immediately instead of waiting for a change-driven frame.
		if available && !geometryChanged && !restoreAt.IsZero() {
			elapsed := e.deps.Clock.Now().Sub(restoreAt)
			e.publish(sourceID, state, buildPayload(zones, advanceSignals(restoreSignals, elapsed)))
		}
	}
	if !available && (wasAvailable || needsRegister) {
		e.sinkMu.Lock()
		e.mu.Lock()
		closed := e.closed
		e.mu.Unlock()
		if !closed {
			if err := e.deps.Sink.Offline(DeviceID(sourceID)); err != nil {
				utils.LogErrorf("presence: offline %s: %v", DeviceID(sourceID), err)
			}
		}
		e.sinkMu.Unlock()
	}
}

// Remove retires the derived device for one source and forgets its state. A derived id is
// accepted defensively and resolved back to its source.
func (e *Engine) Remove(id string) {
	if source, ok := SourceDeviceID(id); ok {
		id = source
	}
	e.mu.Lock()
	state, known := e.states[id]
	delete(e.states, id)
	wasRegistered := state != nil && state.registered
	if state != nil {
		state.registered = false
		state.available = false
	}
	e.mu.Unlock()
	if !known || !wasRegistered {
		return
	}
	e.sinkMu.Lock()
	defer e.sinkMu.Unlock()
	e.mu.Lock()
	closed := e.closed
	e.mu.Unlock()
	if closed {
		return
	}
	if err := e.deps.Sink.Remove(DeviceID(id)); err != nil {
		utils.LogErrorf("presence: remove %s: %v", DeviceID(id), err)
	}
}

// Close stops the worker and waits for it. Later Observe and Sync calls are ignored.
func (e *Engine) Close() {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return
	}
	e.closed = true
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}
	e.mu.Unlock()
	e.sinkMu.Lock()
	e.sinkMu.Unlock()
	e.workers.Wait()
}

func (e *Engine) run() {
	defer e.workers.Done()
	ticker := time.NewTicker(workerTick)
	defer ticker.Stop()
	for {
		select {
		case <-e.ctx.Done():
			return
		case job := <-e.queue:
			e.handleFrame(job)
		case now := <-ticker.C:
			e.flushDue(now)
		}
	}
}

func (e *Engine) handleFrame(job frameJob) {
	e.mu.Lock()
	state := e.states[job.id]
	e.mu.Unlock()
	if state == nil {
		// A frame can arrive before the first configuration sync: reconcile now.
		e.Sync(job.id)
		e.mu.Lock()
		state = e.states[job.id]
		e.mu.Unlock()
		if state == nil {
			return
		}
	}

	e.mu.Lock()
	if !state.available || !state.registered {
		e.mu.Unlock()
		return
	}
	zones := state.zones
	tracker := state.tracker
	e.mu.Unlock()
	if tracker == nil {
		return
	}

	signals := tracker.Update(job.samples, zones, job.at)
	report := buildPayload(zones, signals)
	immediate := immediateValues(report, zones)
	now := e.deps.Clock.Now()

	e.mu.Lock()
	if reflect.DeepEqual(immediate, state.lastImmediate) {
		state.lastSignals = signals
		state.lastSignalsAt = now
		state.lastImmediate = immediate
		e.mu.Unlock()
		return
	}
	if now.Sub(state.lastPublish) < publishGap {
		state.pending = report
		state.lastSignals = signals
		state.lastSignalsAt = now
		state.lastImmediate = immediate
		e.mu.Unlock()
		return
	}
	state.lastImmediate = immediate
	state.lastSignals = signals
	state.lastSignalsAt = now
	state.lastPublish = now
	state.lastCounters = now
	state.pending = nil
	e.mu.Unlock()

	e.publish(job.id, state, report)
}

// publish delivers one report while holding the sink lock and re-checking state, so a report
// in flight can never land after the derived device was retired.
func (e *Engine) publish(sourceID string, state *sourceState, payload map[string]interface{}) {
	e.sinkMu.Lock()
	defer e.sinkMu.Unlock()
	e.mu.Lock()
	deliver := !e.closed && e.states[sourceID] == state && state.registered && state.available
	e.mu.Unlock()
	if !deliver {
		return
	}
	if err := e.deps.Sink.Update(DeviceID(sourceID), payload); err != nil {
		utils.LogErrorf("presence: update %s: %v", DeviceID(sourceID), err)
	}
}

// immediateValues selects the values that publish on change. Slow counters are excluded so a
// continuously occupied zone does not publish at frame rate.
func immediateValues(payload map[string]interface{}, zones []settings.SpatialZone) map[string]interface{} {
	immediate := map[string]interface{}{
		EntityOccupied:    payload[EntityOccupied],
		EntityTargetCount: payload[EntityTargetCount],
		EntityLastZone:    payload[EntityLastZone],
	}
	for _, zone := range zones {
		occupied, count, _ := ZoneEntityNames(zone.ID)
		immediate[occupied] = payload[occupied]
		immediate[count] = payload[count]
	}
	return immediate
}

// advanceSignals ages counters without touching tracker state, so a periodic flush never
// rewrites sample history or invents a position.
func advanceSignals(signals Signals, elapsed time.Duration) Signals {
	if elapsed <= 0 {
		return signals
	}
	next := Signals{
		Tracks:      signals.Tracks,
		Occupied:    signals.Occupied,
		TargetCount: signals.TargetCount,
		LastZone:    signals.LastZone,
		Zones:       make(map[string]ZoneSignal, len(signals.Zones)),
	}
	step := elapsed
	if step > maxCounterStep {
		step = maxCounterStep
	}
	if len(signals.Tracks) == 0 {
		next.EmptyFor = signals.EmptyFor + step
	}
	for id, zone := range signals.Zones {
		if zone.Occupied {
			zone.Dwell += step
		}
		next.Zones[id] = zone
	}
	return next
}

// flushDue publishes deferred transitions and the slow counters on their cadence.
func (e *Engine) flushDue(now time.Time) {
	type due struct {
		id      string
		state   *sourceState
		payload map[string]interface{}
	}
	var publishes []due
	e.mu.Lock()
	for id, state := range e.states {
		if !state.available || !state.registered {
			continue
		}
		counters := now.Sub(state.lastCounters) >= counterInterval
		if state.pending == nil && !counters {
			continue
		}
		payload := state.pending
		if payload == nil {
			elapsed := now.Sub(state.lastSignalsAt)
			signals := advanceSignals(state.lastSignals, elapsed)
			payload = buildPayload(state.zones, signals)
			state.lastSignals = signals
			state.lastSignalsAt = now
		}
		state.lastImmediate = immediateValues(payload, state.zones)
		state.pending = nil
		state.lastPublish = now
		state.lastCounters = now
		publishes = append(publishes, due{id: id, state: state, payload: payload})
	}
	e.mu.Unlock()
	for _, item := range publishes {
		e.publish(item.id, item.state, item.payload)
	}
}

// zoneSignature identifies the geometry that changes derived values. Names are excluded so a
// rename re-registers labels without discarding tracking history.
func zoneSignature(zones []settings.SpatialZone) string {
	var builder strings.Builder
	for _, zone := range zones {
		fmt.Fprintf(&builder, "%s|%s|", zone.ID, zone.Kind)
		for _, point := range zone.Points {
			fmt.Fprintf(&builder, "%g:%g;", point.X, point.Y)
		}
		builder.WriteByte('#')
	}
	return builder.String()
}
