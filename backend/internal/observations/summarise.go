package observations

// The summarising recorder turns live observations into stored daily evidence.
//
// It lives in this package rather than beside the store because it needs the accumulator defined in
// evidence.go, and the dependency runs one way only: this package depends on the summaries store,
// never the reverse. The store stays a dumb persistence layer.
//
// Same shape as internal/spatialhistory on purpose: one worker, a bounded queue that drops rather
// than blocking, and all writes serialised through that worker. Two background recorders with the
// same lifecycle should not be built two ways.

import (
	"context"
	"sync"
	"time"

	summaries "node-herder/internal/observations/summaries"
	model "node-herder/models/observations"
	"node-herder/utils"
)

const (
	// summaryQueueSize bounds pending work. Observations arrive far faster than one database write
	// per day requires, so the queue absorbs a burst and drops beyond it (NH-06).
	summaryQueueSize = 256

	// summaryTick is how often the worker looks for a day boundary.
	summaryTick = time.Minute

	// pruneEvery bounds retention work; the window is 30 days, so daily is generous.
	pruneEvery = 24 * time.Hour

	// dayLayout is the storage day format.
	dayLayout = "2006-01-02"
)

// Now is the one clock capability the recorder needs.
//
// It is deliberately narrower than utils.Clock: the recorder only ever reads the time, and asking
// for the whole interface would force every test fake to implement sleeping and timers it will
// never use.
type Now interface {
	Now() time.Time
}

// SummaryRepository is the persistence the recorder needs.
//
// It is an interface so the recorder can be exercised without a database, and so a different backend
// would not require changing this code.
type SummaryRepository interface {
	SaveSummary(summary model.DailySummary) error
	SaveTransition(transition model.Transition) error
	Prune(before time.Time) error
	Close() error
}

// SummarisingRecorder folds observations into daily summaries and stores completed days.
//
// It never blocks the observation pipeline: Observe enqueues and returns, and a full queue drops
// with a count. History is a convenience, and a slow disk must never stall the radio.
type SummarisingRecorder struct {
	repo  SummaryRepository
	clock Now

	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	queue   chan model.Observation
	started sync.Once
	stopped sync.Once

	// mutex guards the accumulator map and the drop counter. The worker owns the map in practice,
	// but Observe runs on the caller's goroutine, so the map needs a lock rather than an assumption.
	mutex      sync.Mutex
	accumulate map[string]*dayAccumulator
	dropped    uint64
	lastPrune  time.Time
}

// dayAccumulator is one identity's in-progress day.
type dayAccumulator struct {
	day    string
	family string
	state  evidence
}

// newSummarisingRecorder builds a recorder. The caller must call Start and Close.
func newSummarisingRecorder(repo summaries.Repository, clock Now) *SummarisingRecorder {
	ctx, cancel := context.WithCancel(context.Background())
	return &SummarisingRecorder{
		repo:       repo,
		clock:      clock,
		ctx:        ctx,
		cancel:     cancel,
		queue:      make(chan model.Observation, summaryQueueSize),
		accumulate: make(map[string]*dayAccumulator),
	}
}

// Start begins the worker. Repeated calls are no-ops.
func (r *SummarisingRecorder) Start() {
	r.started.Do(func() {
		r.wg.Add(1)
		go r.worker()
	})
}

// Observe queues one observation without blocking.
func (r *SummarisingRecorder) Observe(observation model.Observation) {
	if r.ctx.Err() != nil || observation.Identity.Key == "" {
		return
	}
	select {
	case r.queue <- observation:
	default:
		r.mutex.Lock()
		r.dropped++
		r.mutex.Unlock()
	}
}

// Dropped reports how many observations were discarded because the queue was full, so a silent
// pipeline is diagnosable rather than mysterious.
func (r *SummarisingRecorder) Dropped() uint64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.dropped
}

// Close stops the worker, stores completed evidence, and closes the repository. Safe to call once.
func (r *SummarisingRecorder) Close() {
	r.stopped.Do(func() {
		r.cancel()
		r.wg.Wait()
	})
}

// worker serialises all accumulation and persistence.
func (r *SummarisingRecorder) worker() {
	defer r.wg.Done()

	ticker := time.NewTicker(summaryTick)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			// Drain what is already queued before promoting, so a shutdown does not silently discard
			// the observations that arrived just before it. The queue is bounded, so this cannot spin.
			r.drain()
			r.promoteAll()
			_ = r.repo.Close()
			return
		case observation := <-r.queue:
			r.handle(observation)
		case <-ticker.C:
			r.tick()
		}
	}
}

// drain folds every queued observation into the accumulators.
func (r *SummarisingRecorder) drain() {
	for {
		select {
		case observation := <-r.queue:
			r.handle(observation)
		default:
			return
		}
	}
}

// handle folds one observation into its accumulator, promoting a previous day if midnight passed.
func (r *SummarisingRecorder) handle(observation model.Observation) {
	key := observation.Identity.Key
	day := observation.ObservedAt.UTC().Format(dayLayout)
	family := fingerprintFamily(observation.Identity)

	r.mutex.Lock()
	accumulator, found := r.accumulate[key]
	if !found {
		accumulator = &dayAccumulator{day: day, family: family}
		r.accumulate[key] = accumulator
	}

	var completed *model.DailySummary
	if accumulator.day != day {
		// The observation belongs to a new day, so the previous one is complete. Store it and start
		// the new day with this observation rather than discarding it.
		previous := accumulator.snapshot(key)
		accumulator.start(day, family)
		completed = &previous
	}

	accumulator.state.observe(observation)
	r.mutex.Unlock()

	if completed != nil {
		r.save(*completed)
	}
}

// tick promotes days that ended and runs retention.
func (r *SummarisingRecorder) tick() {
	now := r.clock.Now()
	today := now.UTC().Format(dayLayout)

	r.mutex.Lock()
	var pending []model.DailySummary
	for key, accumulator := range r.accumulate {
		if accumulator.day != today {
			pending = append(pending, accumulator.snapshot(key))
			delete(r.accumulate, key)
		}
	}
	r.mutex.Unlock()

	for _, summary := range pending {
		r.save(summary)
	}

	if r.lastPrune.IsZero() || now.Sub(r.lastPrune) >= pruneEvery {
		r.lastPrune = now
		if err := r.repo.Prune(now.AddDate(0, 0, -summaries.RetentionDays)); err != nil {
			utils.LogErrorf("observation summaries: prune: %v", err)
		}
	}
}

// promoteAll stores every in-progress day, so a restart does not discard the day so far.
func (r *SummarisingRecorder) promoteAll() {
	r.mutex.Lock()
	pending := make([]model.DailySummary, 0, len(r.accumulate))
	for key, accumulator := range r.accumulate {
		pending = append(pending, accumulator.snapshot(key))
	}
	r.accumulate = make(map[string]*dayAccumulator)
	r.mutex.Unlock()

	for _, summary := range pending {
		r.save(summary)
	}
}

// save writes one day, logging rather than returning because the worker has nowhere to return to.
//
// A day below the write floor is discarded by the store itself, so this does not need to re-check
// it: the rule lives in one place, at the write boundary.
func (r *SummarisingRecorder) save(summary model.DailySummary) {
	if summary.Key == "" || summary.Day == "" {
		return
	}
	if err := r.repo.SaveSummary(summary); err != nil {
		utils.LogErrorf("observation summaries: save %s: %v", summary.Key, err)
	}
}

// snapshot renders the accumulator's current day as a storable row.
func (a *dayAccumulator) snapshot(key string) model.DailySummary {
	summary := a.state.summary(key, a.day, a.family)
	return summary
}

// start resets the accumulator for a new day, keeping lifetime fields.
func (a *dayAccumulator) start(day, family string) {
	a.day = day
	a.family = family
	a.state.resetDaily()
}
