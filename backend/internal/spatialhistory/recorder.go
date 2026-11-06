package spatialhistory

import (
	"context"
	"sync"
	"time"

	"node-herder/models/devices"
	"node-herder/utils"
)

const (
	sourceIdleClose = 30 * time.Minute
	workerTick      = time.Minute
	pruneEvery      = 24 * time.Hour
)

type job struct {
	device   string
	at       time.Time
	targets  []rawTarget
	disabled bool
}

type sourceState struct {
	session    string
	lastSeen   time.Time
	chunkStart time.Time
	lastPoint  map[int]Point
	buffer     []Point
	capped     bool
}

type dayKey struct {
	device string
	day    string
}

// Recorder turns accepted frames into movement-gated session chunks. It never blocks the
// device lane: frames enter a bounded queue and one worker performs all writes.
type Recorder struct {
	repo    Repository
	service *Service
	enabled func(string) bool
	clock   utils.Clock
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	queue   chan job
	started sync.Once
	stopped sync.Once

	sources  map[string]*sourceState
	daily    map[dayKey]int
	lastPrun time.Time
}

// NewRecorder builds a recorder. enabled is the source device's history policy.
func NewRecorder(repo Repository, enabled func(string) bool, clock utils.Clock) *Recorder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Recorder{
		repo: repo, service: NewService(repo), enabled: enabled, clock: clock,
		ctx: ctx, cancel: cancel, queue: make(chan job, 256),
		sources: make(map[string]*sourceState), daily: make(map[dayKey]int),
	}
}

// Service exposes the read models backed by the same store.
func (r *Recorder) Service() *Service { return r.service }

// Start begins the worker; repeated calls are no-ops.
func (r *Recorder) Start() {
	r.started.Do(func() {
		r.wg.Add(1)
		go r.worker()
	})
}

// Observe queues one accepted frame without blocking. A full queue drops the frame; history is
// best-effort and must never back-pressure device processing.
func (r *Recorder) Observe(frame *devices.SpatialFrame) {
	if frame == nil || frame.Id == "" || r.ctx.Err() != nil {
		return
	}
	disabled := r.enabled == nil || !r.enabled(frame.Id)
	targets := make([]rawTarget, 0, len(frame.Targets))
	for _, target := range frame.Targets {
		if !target.Active || target.X == nil || target.Y == nil {
			continue
		}
		x, okX := toMillimetres(target.X.Value, target.X.Unit)
		y, okY := toMillimetres(target.Y.Value, target.Y.Unit)
		if !okX || !okY {
			continue
		}
		targets = append(targets, rawTarget{slot: target.ID, x: x, y: y, active: true})
	}
	select {
	case r.queue <- job{device: frame.Id, at: frame.At, targets: targets, disabled: disabled}:
	default:
	}
}

// Close flushes buffers and closes the store. It is safe to call once; later calls are no-ops.
func (r *Recorder) Close() {
	r.stopped.Do(func() {
		r.cancel()
		r.wg.Wait()
	})
}

func (r *Recorder) worker() {
	defer r.wg.Done()
	ticker := time.NewTicker(workerTick)
	defer ticker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			r.flushAll()
			_ = r.repo.Close()
			return
		case frame := <-r.queue:
			r.handle(frame)
		case <-ticker.C:
			r.tick()
		}
	}
}

func (r *Recorder) handle(frame job) {
	state := r.sources[frame.device]
	if state != nil {
		state.lastSeen = frame.at
	}
	if frame.disabled {
		r.closeSession(frame.device)
		return
	}
	if len(frame.targets) == 0 {
		// A complete frame with no active target is the firmware clearing the slots.
		r.closeSession(frame.device)
		return
	}
	if state == nil || state.session == "" {
		state = &sourceState{
			session: newSessionID(), lastPoint: make(map[int]Point), chunkStart: frame.at,
			lastSeen: frame.at,
		}
		r.sources[frame.device] = state
	}
	for _, target := range mergeTargets(frame.targets) {
		point := Point{At: frame.at.UnixMilli(), Slot: target.slot, X: quantize(target.x), Y: quantize(target.y)}
		last, ok := state.lastPoint[target.slot]
		if ok {
			if distanceMM(last, point) < movementGateMM || frame.at.Sub(time.UnixMilli(last.At)) < minSampleGap {
				continue
			}
		}
		if r.dailyCount(frame.device, frame.at) >= MaxPointsPerDay {
			state.capped = true
			continue
		}
		state.lastPoint[target.slot] = point
		state.buffer = append(state.buffer, point)
		r.daily[dayKey{device: frame.device, day: frame.at.UTC().Format("2006-01-02")}]++
		if len(state.buffer) >= chunkPoints || frame.at.Sub(state.chunkStart) >= chunkInterval {
			r.flush(frame.device, state)
		}
	}
}

func (r *Recorder) dailyCount(device string, at time.Time) int {
	return r.daily[dayKey{device: device, day: at.UTC().Format("2006-01-02")}]
}

func (r *Recorder) flush(device string, state *sourceState) {
	if len(state.buffer) == 0 {
		return
	}
	chunk := Chunk{
		Session: state.session,
		From:    state.buffer[0].At,
		To:      state.buffer[len(state.buffer)-1].At,
		Points:  append([]Point(nil), state.buffer...),
		Capped:  state.capped,
	}
	if err := r.repo.SaveChunk(device, chunk); err != nil {
		utils.LogErrorf("spatial history: save %s: %v", device, err)
	}
	state.buffer = nil
	state.chunkStart = time.Now().UTC()
}

func (r *Recorder) closeSession(device string) {
	state := r.sources[device]
	if state == nil {
		return
	}
	r.flush(device, state)
	delete(r.sources, device)
}

func (r *Recorder) flushAll() {
	for device, state := range r.sources {
		r.flush(device, state)
	}
}

func (r *Recorder) tick() {
	now := r.clock.Now()
	for device, state := range r.sources {
		if now.Sub(state.lastSeen) >= sourceIdleClose {
			r.closeSession(device)
		}
	}
	if r.lastPrun.IsZero() || now.Sub(r.lastPrun) >= pruneEvery {
		r.lastPrun = now
		if err := r.repo.Prune(now.Add(-Retention)); err != nil {
			utils.LogErrorf("spatial history: prune: %v", err)
		}
	}
}

func toMillimetres(value float64, unit string) (float64, bool) {
	switch unit {
	case "mm":
		return value, true
	case "cm":
		return value * 10, true
	case "m":
		return value * 1000, true
	default:
		return 0, false
	}
}
