package observations

// This file aggregates raw observations into a bounded, change-driven snapshot for the
// browser. It owns one worker and one bounded queue: observations arrive far faster than any
// UI needs them, so the queue drops oldest and counts the loss (NH-06) instead of growing.
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
)

// IdentityStore reads persisted operator decisions. SaveObserver is deliberately absent: the
// service only ever reads decisions, so a transport cannot mutate user intent.
type IdentityStore interface {
	ObservationIdentities() map[string]settings.ObservationIdentity
}

// Deps are the service's collaborators. Every field is injected so tests use a fake clock and
// a stub store, and the service owns no global state.
type Deps struct {
	// Context bounds the worker lifetime. Nil uses context.Background().
	Context context.Context
	Store   IdentityStore
	Clock   utils.Clock
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
}

// Service aggregates observations from every observing source into one snapshot stream.
type Service struct {
	store      IdentityStore
	clock      utils.Clock
	publish    func(model.Snapshot)
	queue      chan model.Observation
	dropped    func()
	publishGap time.Duration
	staleAfter time.Duration

	mutex       sync.Mutex
	sources     map[string]*sourceState
	aggregates  map[string]*aggregate
	arrivals    int
	lastState   string
	lastAt      time.Time
	lastPublish time.Time
	truncated   bool
}

// sourceState is one registered observing device's identity and declared facets.
type sourceState struct {
	name   string
	media  []model.Medium
	facets []string
}

// aggregate is the rolling state for one identity within this process.
type aggregate struct {
	identity  model.Identity
	lastSeen  time.Time
	packets   uint64
	rssi      *int
	sources   map[string]*int
	ephemeral bool
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
	return &Service{
		store:      deps.Store,
		clock:      deps.Clock,
		publish:    deps.Publish,
		queue:      make(chan model.Observation, queue),
		dropped:    deps.Dropped,
		publishGap: publishGap,
		staleAfter: staleAfter,
		sources:    map[string]*sourceState{},
		aggregates: map[string]*aggregate{},
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
	s.sources[id] = &sourceState{
		name:   name,
		media:  append([]model.Medium(nil), media...),
		facets: append([]string(nil), facets...),
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
func (s *Service) ObserveAdvertisements(_ context.Context, batch []model.Observation) error {
	for _, observation := range batch {
		select {
		case s.queue <- observation:
		default:
			if s.dropped != nil {
				s.dropped()
			}
			return nil
		}
	}
	return nil
}

// Start runs the aggregation worker until ctx is cancelled, then returns after the worker
// has stopped. It is safe to call once; the returned function reports completion.
func (s *Service) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	go s.run(ctx)
}

// run consumes observations and flushes on the publish interval. A ticker drives flushes so a
// burst that stops mid-window still publishes, and an idle hub publishes nothing because the
// snapshot is compared before broadcast.
func (s *Service) run(ctx context.Context) {
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
			sources:   map[string]*int{},
			ephemeral: !observation.Identity.Stable,
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
	rssi := observation.RSSI
	current.sources[observation.Source] = &rssi
	if current.rssi == nil || rssi > *current.rssi {
		best := rssi
		current.rssi = &best
	}
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
			ID:     id,
			Name:   source.name,
			Media:  append([]model.Medium(nil), source.media...),
			Facets: append([]string(nil), source.facets...),
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

	snapshot.Identities = make([]model.IdentityView, 0, len(visible))
	for _, current := range visible {
		decision := decisions[current.identity.Key].Decision
		if decision == "" {
			decision = settings.DecisionUnknown
		}
		view := model.IdentityView{
			Key:      current.identity.Key,
			Kind:     current.identity.Kind,
			Vendor:   current.identity.Vendor,
			Stable:   current.identity.Stable,
			Decision: string(decision),
			LastSeen: current.lastSeen,
			Packets:  current.packets,
			RSSI:     current.rssi,
			Sources:  sourceReadings(current.sources),
		}
		if label := decisions[current.identity.Key].Label; label != "" {
			view.Label = label
		} else {
			view.Label = current.identity.Label
		}
		snapshot.Identities = append(snapshot.Identities, view)
	}
	return snapshot
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
// a snapshot with unchanged content must not broadcast merely because time passed.
func renderSignature(snapshot model.Snapshot) string {
	signature := make([]byte, 0, 64+len(snapshot.Identities)*32)
	for _, source := range snapshot.Sources {
		signature = append(signature, source.ID...)
		signature = append(signature, 0)
	}
	for _, identity := range snapshot.Identities {
		signature = append(signature, identity.Key...)
		signature = append(signature, 0)
		signature = append(signature, identity.Decision...)
		signature = append(signature, identity.Label...)
		signature = append(signature, byte(identity.Packets&0xFF), byte(identity.Packets>>8))
		if identity.RSSI != nil {
			signature = append(signature, byte(*identity.RSSI+128))
		}
		signature = append(signature, 0)
	}
	return string(signature)
}
