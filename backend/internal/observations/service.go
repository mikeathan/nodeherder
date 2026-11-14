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
	defaultQueueSize     = 64
	defaultPublishGap    = 250 * time.Millisecond
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
	// PublishGap overrides the broadcast interval; 0 uses defaultPublishGap.
	PublishGap time.Duration
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
	staleAfter    time.Duration
	history       HistoryReader
	recorder      ObservationRecorder
	classifyEvery time.Duration

	wg sync.WaitGroup

	mutex       sync.Mutex
	sources     map[string]*sourceState
	aggregates  map[string]*aggregate
	arrivals    int
	lastState   string
	lastPublish time.Time
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
	// readings is a bounded, oldest-first window of recent signal strengths used to describe
	// direction of change. It is deliberately memory-only: trend is live state, not history.
	readings []rssiReading
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
		staleAfter:    staleAfter,
		history:       deps.History,
		recorder:      deps.Recorder,
		classifyEvery: classifyEvery,
		sources:       map[string]*sourceState{},
		aggregates:    map[string]*aggregate{},
		classified:    map[string]model.ClassificationView{},
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
func (s *Service) RecordSourceActivity(id string, advertisements, decoded uint64, at time.Time) {
	if id == "" {
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
}

// ObserveAdvertisements enqueues a batch of observations without blocking. When the queue is
// full the batch is dropped and counted rather than stalling the source's session. The name
// avoids colliding with the ESPHome sink's Observe, which carries a different payload.
//
// Drops are counted once per rejected batch rather than once per packet: the purpose is to make
// a saturated queue visible, and taking the service mutex per packet on a hot path would make
// saturation worse. The loop stops at the first rejection, so a burst cannot thrash it.
func (s *Service) ObserveAdvertisements(_ context.Context, batch []model.Observation) error {
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
	current.sources[observation.Source] = &rssi
	// RSSI tracks the most recent reading, not a lifetime maximum: a device that moves away
	// must appear further, and a high-water mark would freeze its distance at the closest
	// point it ever reached.
	current.rssi = &rssi
	// The trend window is fed here, on the aggregation worker, so the browser never has to
	// reconstruct history and a reload cannot change the answer.
	current.readings = appendReading(current.readings, rssi, seenAt)
}

// flush builds a snapshot, drops stale identities, and publishes only when the rendered state
// changed. Comparing a serialised form keeps the check honest without a deep-equal walk.
func (s *Service) flush() {
	s.mutex.Lock()
	now := s.clock.Now()
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
	changed := rendered != s.lastState && now.Sub(s.lastPublish) >= s.publishGap
	if changed {
		s.lastState = rendered
		s.lastPublish = now
	}
	s.mutex.Unlock()

	if changed && s.publish != nil {
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
		At:       now,
		Arrivals: s.arrivals,
		Sources:  make([]model.SourceDescriptor, 0, len(s.sources)),
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
		trend := computeTrend(current.readings, now)
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
// Infrastructure and neighbour are hidden because they are the noise the operator complained
// about; everything else is shown. A provisional result keeps its own decision, because the
// pre-filter is more conservative than this rule and its reasoning is about position rather
// than about class.
func showByDefaultFor(class string, stable bool) bool {
	switch class {
	case ClassInfrastructure:
		// A rotating address cannot be household equipment even if the evidence looks like it, so
		// trust movement over presence when the two disagree and keep it visible for inspection.
		return !stable
	case ClassNeighbour:
		return false
	default:
		// Personal, transient and learning are all shown: they are either the operator's own
		// devices or things they may want to triage.
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

// renderSignature is the change-detection key for a snapshot. Timestamps are excluded because
// a snapshot with unchanged content must not broadcast merely because time passed. Diagnostics
// counters are included: a source receiving advertisements that decode to nothing changes no
// identity, yet that is precisely the state the operator needs to see.
func renderSignature(snapshot model.Snapshot) string {
	signature := make([]byte, 0, 64+len(snapshot.Identities)*32)
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
		signature = appendUint64(signature, source.Diagnostics.Advertisements)
		signature = appendUint64(signature, source.Diagnostics.Decoded)
		signature = appendUint64(signature, source.Diagnostics.Dropped)
	}
	for _, identity := range snapshot.Identities {
		signature = append(signature, identity.Key...)
		signature = append(signature, 0)
		signature = append(signature, identity.Decision...)
		signature = append(signature, identity.Label...)
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
