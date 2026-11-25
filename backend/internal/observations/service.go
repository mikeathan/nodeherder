// Package observations turns raw radio observations into a bounded, change-driven snapshot.
//
// Layout:
//
//	ble.go       identity resolution from a Bluetooth advertisement (medium logic, transport-free)
//	trend.go     direction of change and the smoothed signal median (pure functions)
//	service.go   aggregation, the bounded queue, and the snapshot contract
//
// Design constraints that shaped it:
//
//   - Observations arrive far faster than a UI can use them, so the queue is bounded and
//     rejects at enqueue rather than growing. Drops are counted per source, so a saturated
//     queue is visible instead of silent (NH-06).
//   - A rejected batch is the *newest*, not the oldest: the channel is full of data that
//     already arrived, and discarding it to admit newer packets would make the viewer stutter
//     backwards in time.
//   - State is memory-only. A restart loses the live view and nothing else; nothing here
//     persists an observation, and identity triage lives in settings.
package observations

import (
	"context"
	"sort"
	"sync"
	"time"

	model "node-herder/models/observations"
	"node-herder/models/settings"
	"node-herder/utils"
)

// Aggregation limits. The publish gap bounds broadcast rate independently of inbound rate;
// the identity cap bounds one snapshot; the source cap bounds descriptors.
const (
	defaultQueueSize  = 64
	defaultPublishGap = 250 * time.Millisecond
	// defaultTelemetryGap bounds snapshots caused only by per-packet telemetry (RSSI,
	// signal windows, packet and diagnostic counters). Those change on every BLE batch,
	// so publishing them at defaultPublishGap broadcasts a full snapshot four times a
	// second to every client forever. Structural changes still use defaultPublishGap.
	defaultTelemetryGap  = time.Second
	defaultStaleIdentity = 10 * time.Minute
	maxSources           = 32
	// defaultClassifyInterval bounds how often stored evidence is re-read. A device's class can
	// only change when a day of evidence is written, so re-reading more than once a minute would
	// be wasted work; a minute is fast enough that a completed day reaches the viewer promptly.
	defaultClassifyInterval = time.Minute
)

// Proximity trend limits.
const (
	// rssiWindowSize bounds recent readings per identity. At typical advertisement rates this
	// covers roughly a minute: long enough to see an approach, short enough to react to a turn.
	rssiWindowSize = 20
	// rssiTrendHorizon is how old the newest reading may be before trend becomes unknown,
	// so a stalled device never reports a direction that stopped being true.
	rssiTrendHorizon = 45 * time.Second
	// rssiTrendDeadbandDB is the smoothed change required to call a trend. Ordinary RSSI jitter
	// is several dB, so without a deadband steady motion would read as movement.
	rssiTrendDeadbandDB = 3.0
	// rssiTrendMinReadings is the shortest window that can describe a direction.
	rssiTrendMinReadings = 3
)

// IdentityStore reads persisted operator decisions. SaveObserver is deliberately absent: the
// service only ever reads decisions, so a transport cannot mutate user intent.
type IdentityStore interface {
	ObservationIdentities() map[string]settings.ObservationIdentity
}

// HistoryReader supplies the stored evidence classification reads.
//
// It is separate from IdentityStore because the two have different lifetimes and different
// owners: decisions live in settings and are authoritative, while summaries live in their own
// database and are advisory evidence. A nil reader means classification is unavailable, not that
// it is broken, so the service degrades to the pre-filter rather than refusing to start.
type HistoryReader interface {
	Summaries(key string, limit int) ([]model.DailySummary, error)
}

// Deps are the service's collaborators. Every field is injected so tests use a fake clock and
// a stub store, and the service owns no global state.
type Deps struct {
	Store IdentityStore
	Clock utils.Clock
	// Publish receives each changed snapshot. Nil discards snapshots, which lets a hub
	// without a browser still aggregate for other consumers.
	Publish func(model.Snapshot)
	// Queue overrides the queue depth; 0 uses defaultQueueSize.
	Queue int
	// PublishGap overrides the structural broadcast interval; 0 uses defaultPublishGap.
	PublishGap time.Duration
	// TelemetryGap overrides the interval for snapshots that changed only by telemetry;
	// 0 uses defaultTelemetryGap. Values below PublishGap are clamped to PublishGap.
	TelemetryGap time.Duration
	// StaleAfter drops identities not seen for this long; 0 uses defaultStaleIdentity.
	StaleAfter time.Duration
	// Dropped observes discarded batches, so a saturated queue is visible rather than silent.
	Dropped func()
	// History reads stored daily evidence for classification; nil disables classification.
	History HistoryReader
	// Recorder receives every observation so evidence accumulates for later classification.
	// Nil disables accumulation.
	Recorder ObservationRecorder
	// ClassifyInterval bounds how often stored evidence is re-read; 0 uses
	// defaultClassifyInterval.
	ClassifyInterval time.Duration
	// Active reports whether any integration wants this data. Nil means always active, which is the
	// behaviour for a hub with no observation integration to ask.
	//
	// Gating here rather than at the device is deliberate: the device keeps its session and its
	// entity state, and only the advertisement stream stops. That keeps disabling an integration a
	// data decision instead of a device-lifecycle one.
	Active func() bool
}

// ObservationRecorder accepts observations for evidence accumulation.
type ObservationRecorder interface {
	Observe(observation model.Observation)
}

// Service aggregates observations from every observing source into one snapshot stream.
type Service struct {
	store         IdentityStore
	clock         utils.Clock
	publish       func(model.Snapshot)
	queue         chan model.Observation
	dropped       func()
	publishGap    time.Duration
	telemetryGap  time.Duration
	staleAfter    time.Duration
	history       HistoryReader
	recorder      ObservationRecorder
	classifyEvery time.Duration
	active        func() bool

	wg sync.WaitGroup

	mutex          sync.Mutex
	sources        map[string]*sourceState
	aggregates     map[string]*aggregate
	arrivals       int
	lastState      string
	lastStructural string
	lastPublish    time.Time
	// observing caches the gate's answer so the snapshot can carry it without asking again.
	//
	// The authoritative answer lives behind the integration runtime's lock, and Snapshot is called
	// from inside that lock when the runtime resolves a local provider's census. Calling the gate
	// from there would take the same lock twice and deadlock. Refreshed on the flush tick, which
	// runs with no lock held, so it is at most one publish gap behind.
	observing bool
	// classified caches each identity's class so the store is read on a slow interval rather than
	// once per snapshot. Classification depends on stored evidence, which changes at most once a
	// day, so re-reading it every 250 ms would be pure waste.
	classified   map[string]model.ClassificationView
	lastClassify time.Time
}

// sourceState is one registered observing device's identity and declared facets.
type sourceState struct {
	name        string
	media       []model.Medium
	facets      []string
	diagnostics model.SourceDiagnostics
}

// aggregate is the rolling state for one identity within this process.
type aggregate struct {
	identity  model.Identity
	firstSeen time.Time
	lastSeen  time.Time
	packets   uint64
	rssi      *int
	sources   map[string]*int
	// readings is a bounded, oldest-first window per source. A single shared window would interleave
	// two antennas' readings, so a device heard strongly by one proxy and weakly by another would
	// average to a distance true of neither — and "strongest reading wins" could not be applied.
	readings map[string][]rssiReading
	// primary is the source whose readings describe this identity: the one heard most strongly and
	// still reporting recently. With one source it is always that source.
	primary string
}

// rssiReading is one signal strength sample in the trend window.
type rssiReading struct {
	RSSI int
	At   time.Time
}

// NewService builds a service. Call Start to begin consuming observations.
func NewService(deps Deps) *Service {
	queue := deps.Queue
	if queue <= 0 {
		queue = defaultQueueSize
	}
	publishGap := deps.PublishGap
	if publishGap <= 0 {
		publishGap = defaultPublishGap
	}
	telemetryGap := deps.TelemetryGap
	if telemetryGap <= 0 {
		telemetryGap = defaultTelemetryGap
	}
	if telemetryGap < publishGap {
		telemetryGap = publishGap
	}
	staleAfter := deps.StaleAfter
	if staleAfter <= 0 {
		staleAfter = defaultStaleIdentity
	}
	classifyEvery := deps.ClassifyInterval
	if classifyEvery <= 0 {
		classifyEvery = defaultClassifyInterval
	}
	return &Service{
		store:         deps.Store,
		clock:         deps.Clock,
		publish:       deps.Publish,
		queue:         make(chan model.Observation, queue),
		dropped:       deps.Dropped,
		publishGap:    publishGap,
		telemetryGap:  telemetryGap,
		staleAfter:    staleAfter,
		history:       deps.History,
		recorder:      deps.Recorder,
		classifyEvery: classifyEvery,
		active:        deps.Active,
		// Seeded from the gate so the first snapshot is honest before a flush has run. A nil gate
		// means "always observing", matching isActive.
		observing: deps.Active == nil || deps.Active(),

		sources:    map[string]*sourceState{},
		aggregates: map[string]*aggregate{},
		classified: map[string]model.ClassificationView{},
	}
}

// RegisterSource declares an observing device and the facets it can populate. Re-registering
// the same id replaces its descriptor, so a reconnect does not duplicate it.
func (s *Service) RegisterSource(id, name string, media []model.Medium, facets []string) {
	if id == "" {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.sources) >= maxSources {
		if _, exists := s.sources[id]; !exists {
			return
		}
	}
	// Preserve counters across a reconnect so a restart of the session does not make the
	// viewer look like it has never received anything.
	existing := s.sources[id]
	state := &sourceState{
		name:   name,
		media:  append([]model.Medium(nil), media...),
		facets: append([]string(nil), facets...),
	}
	if existing != nil {
		state.diagnostics = existing.diagnostics
	}
	s.sources[id] = state
}

// RecordSourceActivity accumulates delivery counters for one source so the viewer can
// distinguish "connected but silent" from "not configured". The transport reports raw counts
// because only it can measure them. An unknown source is ignored: registration happens before
// subscription, so this ordering means a stray report cannot create a phantom descriptor.
//
// It honours the same gate as ObserveAdvertisements. Without that, a sensor streaming before the
// integration was added — or while it is disabled — advanced these counters while every batch was
// discarded, so the page could report "571 heard" beside "0 devices heard". The counters exist to
// answer "is the radio working for us", and "us" is exactly what the gate decides.
func (s *Service) RecordSourceActivity(id string, advertisements, decoded uint64, at time.Time) {
	if id == "" {
		return
	}
	if !s.isActive() {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	state, ok := s.sources[id]
	if !ok {
		return
	}
	state.diagnostics.Advertisements += advertisements
	state.diagnostics.Decoded += decoded
	if advertisements > 0 {
		seenAt := at
		state.diagnostics.LastAdvertisement = &seenAt
	}
}

// UnregisterSource forgets a source descriptor. Its observed identities age out normally, so
// a source restart does not erase what the operator was looking at.
func (s *Service) UnregisterSource(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sources, id)

	// Drop this source's reading windows from every identity.
	//
	// The descriptors map was freed but the per-identity `readings[source]` and `sources[source]`
	// entries were not, so a source that came and went repeatedly — a proxy restarting every few
	// minutes — accumulated a window and a reading for every identity it had ever heard. At ten
	// minutes of stale identities and a handful of proxies that is bounded in practice, but it is a
	// leak with no upper bound in principle, and clearing it costs one pass.
	for _, current := range s.aggregates {
		delete(current.readings, id)
		delete(current.sources, id)
		// Re-decide who describes this identity now that one anchor is gone; otherwise a device that
		// only a departed source ever heard keeps reporting that source's last distance.
		current.primary = s.strongestSource(current.sources, current.readings, "", s.clock.Now())
		if current.primary == "" {
			// No readings left, so the last distance is no longer meaningful. The aggregate itself
			// stays until it ages out, because another source may hear it within the window.
			current.rssi = nil
		}
	}
}

// ObserveAdvertisements enqueues a batch of observations without blocking. When the queue is
// full the batch is dropped and counted rather than stalling the source's session. The name
// avoids colliding with the ESPHome sink's Observe, which carries a different payload.
//
// Drops are counted once per rejected batch rather than once per packet: the purpose is to make
// a saturated queue visible, and taking the service mutex per packet on a hot path would make
// saturation worse. The loop stops at the first rejection, so a burst cannot thrash it.
func (s *Service) ObserveAdvertisements(_ context.Context, batch []model.Observation) error {
	// Read once per batch rather than per observation: a batch is one device's delivery, so the
	// answer cannot change meaningfully within it, and the check stays cheap on the hot path.
	if !s.isActive() {
		return nil
	}
	for _, observation := range batch {
		select {
		case s.queue <- observation:
		default:
			s.noteDroppedBatch(observation.Source)
			if s.dropped != nil {
				s.dropped()
			}
			return nil
		}
	}
	return nil
}

// Snapshot returns the current live view.
//
// Exposed so an integration can publish what the hub is already aggregating instead of
// re-deriving it. The snapshot is a copy, so a caller cannot mutate aggregation state.
func (s *Service) Snapshot() model.Snapshot {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.snapshotLocked(s.clock.Now())
}

// isActive reports whether observation is wanted. A nil gate means always, so an installation with
// no integration configured behaves exactly as it did before this gate existed.
//
// The gate consults the integration runtime, which reads a BoltDB transaction. That sounds expensive
// for a path called once per advertisement batch, but a read transaction on an indexed key is
// microseconds while a batch arrives a few times a second — three orders of magnitude of headroom.
// A cache here would trade that for lag on an operator's toggle, so it is deliberately not cached:
// disabling observation must take effect on the next batch, not on the next publish tick.
func (s *Service) isActive() bool {
	return s.active == nil || s.active()
}

// noteDroppedBatch increments the drop counter for one source once. Attribution keeps a
// saturated queue visible per device instead of as an unattributed global number.
func (s *Service) noteDroppedBatch(source string) {
	if source == "" {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if state, ok := s.sources[source]; ok {
		state.diagnostics.Dropped++
	}
}

// Start launches the aggregation worker. It returns immediately; call Wait to observe shutdown.
//
// It is safe to call once. A second call would start a second consumer on the same queue, which
// would split observations between two aggregates rather than rejecting the call — hence "once"
// rather than "idempotently".
func (s *Service) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	s.wg.Add(1)
	go s.run(ctx)
}

// Wait blocks until the aggregation worker has stopped, which happens when the context passed to
// Start is cancelled.
//
// This exists so shutdown can order itself. The service reads the summary store, and the recorder
// closes it: without a way to wait, the store could be closed while a classification pass was
// still reading, which turns a clean shutdown into a stream of errors from a closing database.
func (s *Service) Wait() { s.wg.Wait() }

// WaitContext blocks until the worker has stopped or ctx expires, and reports whether it stopped.
//
// Shutdown must be bounded. The worker exits when its own context is cancelled, but that context
// is not guaranteed to be cancelled on every exit path — a failed Listen never cancels it — so an
// unbounded Wait there would hang the process instead of shutting it down. Returning a bool lets
// the caller log a timeout rather than pretend the ordering held.
func (s *Service) WaitContext(ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

// run consumes observations and flushes on the publish interval. A ticker drives flushes so a
// burst that stops mid-window still publishes, and an idle hub publishes nothing because the
// snapshot is compared before broadcast.
func (s *Service) run(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.publishGap)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case observation := <-s.queue:
			s.record(observation)
		case <-ticker.C:
			s.flush()
		}
	}
}

// record folds one observation into the aggregate state.
func (s *Service) record(observation model.Observation) {
	if observation.Identity.Key == "" {
		return
	}
	// Accumulate evidence for later classification. Done on this worker so accumulation is
	// serialised with aggregation, and before the mutex because the recorder has its own queue and
	// must never be reached while holding this lock.
	if s.recorder != nil {
		s.recorder.Observe(observation)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	seenAt := observation.ObservedAt
	if seenAt.IsZero() {
		seenAt = s.clock.Now()
	}
	current, ok := s.aggregates[observation.Identity.Key]
	if !ok {
		current = &aggregate{
			identity:  observation.Identity,
			firstSeen: seenAt,
			sources:   map[string]*int{},
			readings:  map[string][]rssiReading{},
		}
		s.aggregates[observation.Identity.Key] = current
		if observation.Identity.Stable {
			s.arrivals++
		}
	}
	current.packets++
	if seenAt.After(current.lastSeen) {
		current.lastSeen = seenAt
	}
	// Prefer the operator-supplied label, then the most recent advertised one.
	if current.identity.Label == "" && observation.Identity.Label != "" {
		current.identity.Label = observation.Identity.Label
	}
	// A distinct allocation per source reading. Assigning `&observation.RSSI` would alias the
	// observation, which the caller may reuse for its next entry — a genuine hazard with batch
	// decoders that decode into a reused struct.
	rssi := observation.RSSI

	// RSSI 0 means the source could not measure the signal, not "zero dBm". Every other consumer of
	// this value already treats it that way: evidence.observe skips it, the provisional classifier
	// rejects it, and the neighbour score ignores it. Accepting it here put a 0 dBm spike into the
	// 20-reading window, which can make the median 0 — rendered as "0 dBm" and mapped to the minimum
	// distance, so the device is drawn on top of the radar. The packet still counts; only the signal
	// reading is discarded.
	if observation.RSSI == 0 {
		current.primary = s.strongestSource(current.sources, current.readings, current.primary, s.clock.Now())
		return
	}
	current.sources[observation.Source] = &rssi

	// Feed this source's own window, then re-decide which source describes the device.
	//
	// "Strongest wins" rather than averaging: with two proxies at different distances the nearer one
	// hears the device better and gives the truer distance. Averaging would produce a number matching
	// neither antenna, and a device moving towards one proxy would appear to move half as much.
	current.readings[observation.Source] = appendReading(current.readings[observation.Source], rssi, seenAt)
	current.primary = s.strongestSource(current.sources, current.readings, current.primary, s.clock.Now())
	if current.primary == observation.Source {
		// RSSI tracks the most recent reading, not a lifetime maximum: a device that moves away must
		// appear further, and a high-water mark would freeze its distance at its closest point.
		current.rssi = &rssi
	}
}

// strongestSource returns the source whose reading currently describes the device.
//
// Strength alone is not enough: a source that stopped reporting keeps its last reading, so a proxy
// that heard the device once at close range would pin the distance there indefinitely while a
// further antenna kept reporting. A silent source is therefore not a candidate — that is what
// `sourceIsCurrent` is for, and it is why the horizon is shared with the trend.
func (s *Service) strongestSource(
	sources map[string]*int,
	readings map[string][]rssiReading,
	incumbent string,
	now time.Time,
) string {
	best, bestRSSI := "", 0
	for id, reading := range sources {
		if reading == nil || !sourceIsCurrent(readings[id], now) {
			continue
		}
		// Strictly greater. The incumbent is preferred below, so an equal-strength source cannot
		// displace it — which is what keeps the display stable between two proxies that hear a device
		// equally well. An id tiebreak here was tried and is unreachable: the incumbent compares
		// first and wins on `>=`.
		if best == "" || *reading > bestRSSI {
			best, bestRSSI = id, *reading
		}
	}
	if best == "" {
		// Nothing current: keep whatever was chosen, so the last known distance persists rather than
		// the row emptying itself the moment every source goes quiet.
		return incumbent
	}
	if incumbentReading, ok := sources[incumbent]; ok && incumbentReading != nil &&
		sourceIsCurrent(readings[incumbent], now) && *incumbentReading >= bestRSSI {
		return incumbent
	}
	return best
}

// sourceIsCurrent reports whether a source produced a reading within the trend horizon.
func sourceIsCurrent(readings []rssiReading, now time.Time) bool {
	if len(readings) == 0 {
		return false
	}
	return now.Sub(readings[len(readings)-1].At) <= rssiTrendHorizon
}

// flush builds a snapshot, drops stale identities, and publishes only when the rendered state
// changed. Comparing a serialised form keeps the check honest without a deep-equal walk.
func (s *Service) flush() {
	// Read the gate before taking the mutex: it reaches into the integration runtime, which takes
	// its own lock, and the runtime calls back into Snapshot while holding that lock. Asking here
	// keeps the order one-way.
	observing := s.isActive()

	s.mutex.Lock()
	now := s.clock.Now()
	closed := s.observing && !observing
	s.observing = observing
	if closed {
		// Nobody wants this data any more. Keeping the observation itself would let the last census
		// be republished — or re-read by a stale consumer — as if a radio that is no longer being
		// read were still describing the room. Clearing is the difference between "not observing"
		// and "still observing, and here is what it saw".
		//
		// Sources are deliberately kept. A descriptor names a device and the facets it can populate,
		// which is a fact about the device rather than about what was observed, and only the live
		// session can announce it — once, guarded by its own `announced` flag. Clearing it here
		// stranded the viewer: disabling and re-enabling a proxy left `sources` empty until the
		// ESPHome session happened to reconnect, so the filter facets disappeared, the counters
		// never resumed, and a silent source became indistinguishable from an absent one. Their
		// counters freeze on their own, because RecordSourceActivity honours the same gate.
		s.aggregates = map[string]*aggregate{}
		s.classified = map[string]model.ClassificationView{}
		s.arrivals = 0
	}
	for key, current := range s.aggregates {
		if now.Sub(current.lastSeen) > s.staleAfter {
			delete(s.aggregates, key)
		}
	}
	// Collect the identities whose evidence is stale in the cache, then release the lock before
	// touching the database. Reading the store under this lock would let a slow disk stall the
	// radio, which is the one thing this service must never do.
	due := s.lastClassify.IsZero() || now.Sub(s.lastClassify) >= s.classifyEvery
	var pending []string
	if due {
		pending = make([]string, 0, len(s.aggregates))
		for key := range s.aggregates {
			pending = append(pending, key)
		}
	}
	s.mutex.Unlock()

	if due {
		// Refresh first, then build the snapshot once. Refreshing after building would publish a
		// stale class and then immediately publish again, doubling the broadcast for no gain.
		s.refreshClassifications(pending)
		// Re-read the clock: the classification pass reads the database, so the `now` captured
		// before it can be hundreds of milliseconds stale by this point. A snapshot timestamped in
		// the past makes the viewer's "last seen 5 s ago" read wrong, and the publish-gap check
		// below would be computed against a moment that has already passed.
		now = s.clock.Now()
	}

	s.mutex.Lock()
	snapshot := s.snapshotLocked(now)
	rendered := renderSignature(snapshot)
	structural := structuralSignature(snapshot)
	changed := rendered != s.lastState
	publish := false
	if changed {
		// Telemetry-only churn uses the slower cadence; a structural change must reach
		// the viewer promptly. lastStructural is only updated on publish, so a structural
		// change that missed its gap stays structural on the next tick.
		gap := s.publishGap
		if structural == s.lastStructural {
			gap = s.telemetryGap
		}
		publish = now.Sub(s.lastPublish) >= gap
	}
	if publish {
		s.lastState = rendered
		s.lastStructural = structural
		s.lastPublish = now
	}
	s.mutex.Unlock()

	if publish && s.publish != nil {
		s.publish(snapshot)
	}
}

// rebuild re-renders a snapshot after the classification cache changed.
func (s *Service) rebuild(now time.Time) (model.Snapshot, string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	snapshot := s.snapshotLocked(now)
	return snapshot, renderSignature(snapshot)
}

// snapshotLocked builds the renderable snapshot. The caller holds the mutex.
func (s *Service) snapshotLocked(now time.Time) model.Snapshot {
	snapshot := model.Snapshot{
		At:        now,
		Arrivals:  s.arrivals,
		Observing: s.observing,
		Sources:   make([]model.SourceDescriptor, 0, len(s.sources)),
	}
	sourceIDs := make([]string, 0, len(s.sources))
	for id := range s.sources {
		sourceIDs = append(sourceIDs, id)
	}
	sort.Strings(sourceIDs)
	for _, id := range sourceIDs {
		source := s.sources[id]
		snapshot.Sources = append(snapshot.Sources, model.SourceDescriptor{
			ID:          id,
			Name:        source.name,
			Media:       append([]model.Medium(nil), source.media...),
			Facets:      append([]string(nil), source.facets...),
			Diagnostics: source.diagnostics,
		})
	}

	decisions := map[string]settings.ObservationIdentity{}
	if s.store != nil {
		decisions = s.store.ObservationIdentities()
	}

	visible := make([]*aggregate, 0, len(s.aggregates))
	for _, current := range s.aggregates {
		if decisions[current.identity.Key].Decision == settings.DecisionIgnored {
			continue
		}
		visible = append(visible, current)
	}
	// Most recently seen first, with the key as a stable tiebreak so identical timestamps
	// cannot make the signature flap.
	sort.Slice(visible, func(i, j int) bool {
		if visible[i].lastSeen.Equal(visible[j].lastSeen) {
			return visible[i].identity.Key < visible[j].identity.Key
		}
		return visible[i].lastSeen.After(visible[j].lastSeen)
	})
	if len(visible) > model.MaxSnapshotIdentities {
		visible = visible[:model.MaxSnapshotIdentities]
		snapshot.Truncated = true
	}

	// One scan for the reference reading, not one per identity.
	strongestRSSI := s.strongestRSSILocked()
	snapshot.Classification = s.classificationProgressLocked(now)

	snapshot.Identities = make([]model.IdentityView, 0, len(visible))
	for _, current := range visible {
		decision := decisions[current.identity.Key].Decision
		if decision == "" {
			decision = settings.DecisionUnknown
		}
		view := model.IdentityView{
			Key:       current.identity.Key,
			Kind:      current.identity.Kind,
			Vendor:    current.identity.Vendor,
			Stable:    current.identity.Stable,
			Decision:  string(decision),
			FirstSeen: current.firstSeen,
			LastSeen:  current.lastSeen,
			Packets:   current.packets,
			RSSI:      current.rssi,
			Sources:   sourceReadings(current.sources),
		}
		// The trend describes the primary source's window, so the arrow and the distance agree: both
		// answer "how is this device moving relative to the antenna that hears it best".
		trend := computeTrend(current.readings[current.primary], now)
		view.Trend, view.TrendDeltaDB, view.Signal = trend.Trend, trend.DeltaDB, trend.Signal
		view.RSSISmoothed = trend.Smoothed
		if label := decisions[current.identity.Key].Label; label != "" {
			view.Label = label
		} else {
			view.Label = current.identity.Label
		}
		// A stored classification is attached when available; otherwise the first-packet
		// pre-filter decides. Both paths carry ShowByDefault, so the viewer never has to guess
		// whether a missing class means "hidden" or "not yet known".
		if cached, ok := s.classified[current.identity.Key]; ok {
			attached := cached
			view.Classification = &attached
		} else {
			preliminary := s.provisionalLocked(current, strongestRSSI)
			view.Classification = &preliminary
		}
		snapshot.Identities = append(snapshot.Identities, view)
	}
	return snapshot
}

// provisionalLocked renders the first-packet pre-filter for one identity. Caller holds the mutex.
//
// This is what makes the first day usable: the pre-filter needs no history, so it decides
// visibility immediately instead of leaving the operator with the full census until tomorrow.
//
// strongestRSSI is computed once per snapshot and passed in. It used to be derived here by
// scanning every aggregate, which made one snapshot O(n^2) in identities — and the snapshot is
// built on every publish tick, with a cap of 256 identities, so that was 256 full scans every
// 250 ms for a value that is identical for all of them.
func (s *Service) provisionalLocked(current *aggregate, strongestRSSI int) model.ClassificationView {
	rssi := 0
	if current.rssi != nil {
		rssi = *current.rssi
	}
	guess := provisionalClass(current.identity, rssi, strongestRSSI, s.clock.Now())
	return model.ClassificationView{
		Class:         guess.Class,
		Provisional:   guess.Provisional,
		ShowByDefault: guess.ShowByDefault,
		Reason:        guess.Reason,
		Advisory:      true,
	}
}

// strongestRSSILocked returns the strongest single reading in the current snapshot. Caller holds
// the mutex.
//
// This is the reference for "far away" in the pre-filter. Using an identity's own history would
// be circular, and a fixed dBm threshold would be wrong for every installation whose antenna gain
// differs — so the room compares against itself.
func (s *Service) strongestRSSILocked() int {
	strongest := 0
	for _, current := range s.aggregates {
		if current.rssi == nil {
			continue
		}
		if strongest == 0 || *current.rssi > strongest {
			strongest = *current.rssi
		}
	}
	return strongest
}

// refreshClassifications re-reads stored evidence and updates the cache.
//
// It is called on the slow interval, outside the service mutex, because it touches the database.
// Reading the store while holding the aggregation lock would let a slow disk stall the radio,
// which is the one thing this service must never do.
func (s *Service) refreshClassifications(keys []string) {
	if s.history == nil {
		return
	}

	s.mutex.Lock()
	stable := make(map[string]bool, len(s.aggregates))
	for key, current := range s.aggregates {
		stable[key] = current.identity.Stable
	}
	s.mutex.Unlock()

	s.mutex.Lock()
	previous := s.classified
	s.mutex.Unlock()

	refreshed := make(map[string]model.ClassificationView, len(keys))
	for _, key := range keys {
		summaries, err := s.history.Summaries(key, retentionDays)
		if err != nil {
			// A read failure must not clear an existing class: showing nothing is a worse answer
			// than showing the last good evidence. Carrying the cached entry forward is what makes
			// that true — the cache is replaced wholesale below, so a bare `continue` would have
			// dropped the class the comment promises to keep. It is logged because the store
			// returns errors rather than logging them, and swallowing one would leave a broken
			// database looking exactly like an installation with no history yet.
			utils.LogErrorf("observation summaries: read %s: %v", key, err)
			if cached, ok := previous[key]; ok {
				refreshed[key] = cached
			}
			continue
		}
		if len(summaries) == 0 {
			// Nothing stored yet. Leaving the key absent lets the pre-filter answer, which is the
			// correct first-day behaviour rather than an empty class.
			continue
		}
		result := classify(summaries, s.clock.Now())
		refreshed[key] = model.ClassificationView{
			Class:          result.Class,
			Provisional:    result.Provisional,
			ShowByDefault:  showByDefaultFor(result.Class, stable[key]),
			Candidate:      result.Candidate,
			ChangePending:  result.ChangePending,
			Confidence:     result.Confidence,
			HoursObserved:  result.HoursObserved,
			DaysConsidered: result.DaysConsidered,
			Reason:         result.Evidence.Reason,
			Advisory:       true,
		}
	}

	s.mutex.Lock()
	// Rebuild rather than merge. Merging kept an entry for every identity ever seen, because
	// nothing removed one when its aggregate aged out — over a day of rotating addresses that map
	// grows without bound. Keeping only identities still present is correct as well as bounded: an
	// identity that has aged out has no cached class to show, and when it reappears it is a fresh
	// sighting that will be classified from stored evidence on the next pass.
	s.classified = refreshed
	s.lastClassify = s.clock.Now()
	s.mutex.Unlock()
}

// showByDefaultFor decides the default visibility of a settled class.
//
// Only installed equipment is hidden, and only when its address is stable.
//
// Hiding neighbours was a mistake. Rotating addresses are never hidden — they are the
// `mobile-unidentifiable` class — while a stable address is the only kind that can be identified
// and drawn as a blip. So hiding stable-but-weak devices removed precisely the ones the operator
// could see, and left a radar showing one dot out of forty. The class still labels it as probably
// outside; that is information, not a reason to hide it.
//
// Installed equipment stays hidden because a fridge genuinely is noise, and a rotating address
// cannot be installed equipment whatever the evidence says, so stability is still required.
func showByDefaultFor(class string, stable bool) bool {
	switch class {
	case ClassInfrastructure:
		return !stable
	default:
		return true
	}
}

// sourceReadings renders per-source RSSI in a stable order.
func sourceReadings(source map[string]*int) []model.SourceReading {
	ids := make([]string, 0, len(source))
	for id := range source {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	readings := make([]model.SourceReading, 0, len(ids))
	for _, id := range ids {
		readings = append(readings, model.SourceReading{ID: id, RSSI: source[id]})
	}
	return readings
}

// renderSignature is the full change-detection key for a snapshot. Timestamps are excluded
// because a snapshot with unchanged content must not broadcast merely because time passed.
// Diagnostics counters are included: a source receiving advertisements that decode to
// nothing changes no identity, yet that is precisely the state the operator needs to see.
func renderSignature(snapshot model.Snapshot) string {
	return snapshotSignature(snapshot, true)
}

// structuralSignature is the change key without per-packet telemetry. It decides whether a
// change is structural (publish at publishGap) or telemetry-only (publish at telemetryGap).
// The payload always keeps exact values: only the cadence depends on this key.
func structuralSignature(snapshot model.Snapshot) string {
	return snapshotSignature(snapshot, false)
}

// snapshotSignature is the single encoder for both keys. With telemetry disabled it omits
// the fields that change on every BLE batch: diagnostic counters, packet counts, RSSI,
// smoothed RSSI, trend, trend delta and signal windows. The gate, classification
// progress, source IDs, identity membership (keys), decisions, labels and classes remain,
// so a meaningful change still broadcasts promptly. Source names/media/facets and
// identity kind/vendor/stable were absent from the original key and remain absent;
// they reach the viewer with the next snapshot. Both source IDs and identity keys are
// NUL-terminated here, which relies on neither containing a NUL byte.
func snapshotSignature(snapshot model.Snapshot, telemetry bool) string {
	signature := make([]byte, 0, 64+len(snapshot.Identities)*32)
	// The gate itself must participate. Closing it publishes one clearing snapshot, and without this
	// the clearing snapshot would render identically to the last populated one — so the broadcast
	// that tells every consumer to stop showing a census would be suppressed as a duplicate.
	if snapshot.Observing {
		signature = append(signature, 1)
	} else {
		signature = append(signature, 0)
	}
	// Progress must participate: it changes while no identity field does, and it is the one thing
	// telling the operator whether the classifier is waiting or broken.
	signature = appendUint64(signature, uint64(snapshot.Classification.DaysStored))
	signature = appendUint64(signature, uint64(snapshot.Classification.HoursObserved))
	if snapshot.Classification.Ready {
		signature = append(signature, 1)
	} else {
		signature = append(signature, 0)
	}
	for _, source := range snapshot.Sources {
		signature = append(signature, source.ID...)
		signature = append(signature, 0)
		if telemetry {
			signature = appendUint64(signature, source.Diagnostics.Advertisements)
			signature = appendUint64(signature, source.Diagnostics.Decoded)
			signature = appendUint64(signature, source.Diagnostics.Dropped)
		}
	}
	if !telemetry {
		// Without counter bytes, separate the source IDs from the identity keys so two
		// different snapshots cannot encode to the same byte string.
		signature = append(signature, 0)
	}
	for _, identity := range snapshot.Identities {
		signature = append(signature, identity.Key...)
		signature = append(signature, 0)
		signature = append(signature, identity.Decision...)
		signature = append(signature, identity.Label...)
		if telemetry {
			signature = appendUint64(signature, identity.Packets)
			if identity.RSSI != nil {
				// Encode the full value so a reading moving outside one byte still registers.
				signature = appendUint64(signature, uint64(int64(*identity.RSSI)+128))
			}
			// The smoothed value drives the displayed distance, so it must participate or the
			// figure would freeze while the underlying window moved.
			if identity.RSSISmoothed != nil {
				signature = appendUint64(signature, uint64(int64(*identity.RSSISmoothed)+128))
			}
			// Trend and its sparkline change without any identity field changing, so they must
			// participate or the arrow and history would freeze on a live device.
			signature = append(signature, identity.Trend...)
			signature = appendUint64(signature, uint64(int64(identity.TrendDeltaDB*8)+4096))
			for _, sample := range identity.Signal {
				signature = appendUint64(signature, uint64(int64(sample)+128))
			}
		}
		// Classification must participate or a class that changes without any other field changing
		// would never be broadcast — the viewer would keep showing the old class indefinitely.
		// This is the same omission as D-4 (diagnostics), found the same way: by asking what can
		// change while the signature stays equal.
		signature = appendClassification(signature, identity.Classification)
		signature = append(signature, 0)
	}
	return string(signature)
}

// classificationProgressLocked reports how far classification has got. Caller holds the mutex.
//
// The distinction that matters is "no evidence yet" versus "not working". Both render every row
// as learning, so without this the operator has no way to tell them apart — which is exactly how
// a feature can be complete and still look broken.
func (s *Service) classificationProgressLocked(now time.Time) model.ClassificationProgress {
	progress := model.ClassificationProgress{Identities: len(s.aggregates)}

	for key := range s.aggregates {
		if _, ok := s.classified[key]; ok {
			progress.DaysStored++
			continue
		}
	}

	// The best single identity's evidence decides readiness, because classification is per
	// identity: one device with a day of history is classifiable even if forty others are not.
	for _, view := range s.classified {
		if view.HoursObserved > progress.HoursObserved {
			progress.HoursObserved = view.HoursObserved
		}
	}

	progress.Ready = progress.DaysStored > 0
	if !progress.Ready && progress.Identities > 0 {
		// No summaries exist yet, so no identity has been classified. Say why, and say it in terms
		// of the rule that caused it rather than in terms of a percentage.
		progress.Note = "Classification begins after the first full day. Until then every device " +
			"is shown as learning and installed equipment is not hidden."
	}
	return progress
}

// appendClassification adds a classification to a render signature.
//
// Every field that can differ between two snapshots is included, including the reason: when a
// day's evidence lands the class may stay the same while its justification and hours change, and
// that is exactly the moment the operator is watching for.
func appendClassification(signature []byte, classification *model.ClassificationView) []byte {
	if classification == nil {
		return append(signature, 0)
	}
	signature = append(signature, 1)
	signature = append(signature, classification.Class...)
	signature = append(signature, 0)
	if classification.Provisional {
		signature = append(signature, 1)
	} else {
		signature = append(signature, 0)
	}
	if classification.ShowByDefault {
		signature = append(signature, 1)
	} else {
		signature = append(signature, 0)
	}
	signature = append(signature, classification.Candidate...)
	signature = append(signature, 0)
	if classification.ChangePending {
		signature = append(signature, 1)
	} else {
		signature = append(signature, 0)
	}
	signature = appendUint64(signature, uint64(classification.Confidence*1000))
	signature = appendUint64(signature, uint64(classification.HoursObserved))
	signature = appendUint64(signature, uint64(classification.DaysConsidered))
	signature = append(signature, classification.Reason...)
	return signature
}

// appendUint64 encodes a counter in a fixed width so the signature cannot collide with an
// adjacent field's bytes.
func appendUint64(buffer []byte, value uint64) []byte {
	for shift := 0; shift < 64; shift += 8 {
		buffer = append(buffer, byte(value>>shift))
	}
	return buffer
}
