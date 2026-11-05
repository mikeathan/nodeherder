package spatial

import (
	"sync"
	"testing"
	"time"

	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
)

type fakeClock struct {
	utils.RealClock
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

type fakeConfig struct {
	mu     sync.Mutex
	mapped map[string]bool
	zones  map[string][]settings.SpatialZone
}

func (c *fakeConfig) Mapped(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mapped[id]
}

func (c *fakeConfig) Zones(id string) []settings.SpatialZone {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.zones[id]
}

func (c *fakeConfig) set(id string, mapped bool, zones []settings.SpatialZone) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mapped[id] = mapped
	c.zones[id] = zones
}

type sinkCall struct {
	kind    string
	id      string
	payload map[string]interface{}
}

type fakeSink struct {
	mu      sync.Mutex
	calls   []sinkCall
	block   chan struct{}
	entered chan struct{}
}

func (s *fakeSink) record(kind, id string, payload map[string]interface{}) {
	s.mu.Lock()
	s.calls = append(s.calls, sinkCall{kind: kind, id: id, payload: payload})
	s.mu.Unlock()
}

func (s *fakeSink) Register(device *devices.Device) error {
	s.record("register", device.Id, nil)
	return nil
}

func (s *fakeSink) Update(id string, payload map[string]interface{}) error {
	if s.block != nil {
		select {
		case s.entered <- struct{}{}:
		default:
		}
		<-s.block
	}
	s.record("update", id, payload)
	return nil
}

func (s *fakeSink) Offline(id string) error {
	s.record("offline", id, nil)
	return nil
}

func (s *fakeSink) Remove(id string) error {
	s.record("remove", id, nil)
	return nil
}

func (s *fakeSink) of(kind string) []sinkCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []sinkCall
	for _, call := range s.calls {
		if call.kind == kind {
			out = append(out, call)
		}
	}
	return out
}

func (s *fakeSink) updates() []sinkCall { return s.of("update") }

func onlineSource(id string) *devices.Device {
	device := devices.NewDevice(id)
	device.FriendlyName = "Hall radar"
	device.PowerSource = "mains"
	device.SetAvailability(devices.OnlineAvailability)
	return device
}

func newTestEngine(t *testing.T, sink *fakeSink, config *fakeConfig, clock *fakeClock, source *devices.Device) *Engine {
	t.Helper()
	engine := NewEngine(Deps{
		Config: config,
		Source: func(id string) (*devices.Device, bool) {
			if source != nil && source.Id == id {
				return source, true
			}
			return nil, false
		},
		Sink:  sink,
		Clock: clock,
	})
	t.Cleanup(engine.Close)
	return engine
}

func frameAt(ms int, active bool, x, y float64) *devices.SpatialFrame {
	return &devices.SpatialFrame{
		Id: "0xabc", At: base.Add(time.Duration(ms) * time.Millisecond),
		Targets: []devices.SpatialTarget{{
			ID: 1, Active: active,
			X: spatialAxis(x), Y: spatialAxis(y),
		}},
	}
}

func spatialAxis(value float64) *devices.SpatialAxis {
	return &devices.SpatialAxis{Value: value, Unit: "mm"}
}

func TestSyncRegistersAndRemovesDerivedDevice(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	engine := newTestEngine(t, sink, config, clock, onlineSource("0xabc"))

	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")
	registrations := sink.of("register")
	if len(registrations) != 1 || registrations[0].id != "presence-0xabc" {
		t.Fatalf("registrations = %+v", registrations)
	}

	config.set("0xabc", false, nil)
	engine.Sync("0xabc")
	if removals := sink.of("remove"); len(removals) != 1 || removals[0].id != "presence-0xabc" {
		t.Fatalf("removals = %+v", removals)
	}
}

func TestDerivedDeviceIsIgnoredAsSource(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	engine := newTestEngine(t, sink, config, clock, onlineSource("0xabc"))
	engine.Start()
	engine.Sync("presence-0xabc")
	engine.Observe(frameAt(0, true, 0, 1500))
	time.Sleep(20 * time.Millisecond)
	if calls := sink.of("register"); len(calls) != 0 {
		t.Fatalf("derived device registered as its own source: %+v", calls)
	}
}

func TestOccupancyPublishesOnChangeOnly(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	source := onlineSource("0xabc")
	engine := newTestEngine(t, sink, config, clock, source)
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")

	// The first computation publishes the initial state; an unoccupied zone reports no count.
	engine.handleFrame(frameJob{id: "0xabc", samples: []Sample{{Slot: 1, X: 0, Y: 1500}}, at: base})
	initial := sink.updates()
	if len(initial) != 1 || initial[0].payload[EntityOccupied] != false {
		t.Fatalf("initial publish = %+v", initial)
	}
	clock.advance(time.Second)
	engine.handleFrame(frameJob{id: "0xabc", samples: []Sample{{Slot: 1, X: 0, Y: 1500}}, at: base.Add(time.Second)})
	updates := sink.updates()
	if len(updates) != 2 || updates[1].payload[EntityOccupied] != true {
		t.Fatalf("updates = %+v", updates)
	}
	occupied, count, _ := ZoneEntityNames("desk")
	if updates[1].payload[count] != 1 || updates[1].payload[EntityLastZone] != "Desk" {
		t.Fatalf("occupancy payload = %+v", updates[1].payload)
	}

	clock.advance(2 * time.Second)
	engine.handleFrame(frameJob{id: "0xabc", samples: []Sample{{Slot: 1, X: 0, Y: 1500}}, at: base.Add(3 * time.Second)})
	if updates := sink.updates(); len(updates) != 2 {
		t.Fatalf("unchanged occupancy published again: %+v", updates)
	}
	_ = occupied
}

func TestPublishIsRateLimitedAndFlushed(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	engine := newTestEngine(t, sink, config, clock, onlineSource("0xabc"))
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")

	for step := 0; step < 6; step++ {
		engine.handleFrame(frameJob{id: "0xabc", samples: []Sample{{Slot: 1, X: 0, Y: 1500}}, at: base})
		clock.advance(10 * time.Millisecond)
	}
	if updates := sink.updates(); len(updates) > 1 {
		t.Fatalf("burst published %d updates", len(updates))
	}
	clock.advance(time.Second)
	engine.flushDue(clock.Now())
	if updates := sink.updates(); len(updates) == 0 {
		t.Fatal("deferred transition never flushed")
	}
}

func TestCountersFlushAdvancesDwellAndEmpty(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	engine := newTestEngine(t, sink, config, clock, onlineSource("0xabc"))
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")
	for step := 0; step < 4; step++ {
		engine.handleFrame(frameJob{id: "0xabc", samples: []Sample{{Slot: 1, X: 0, Y: 1500}}, at: base.Add(time.Duration(step) * time.Second)})
		clock.advance(500 * time.Millisecond)
	}
	clock.advance(counterInterval)
	engine.flushDue(clock.Now())
	updates := sink.updates()
	if len(updates) == 0 {
		t.Fatal("no counter flush")
	}
	_, _, dwell := ZoneEntityNames("desk")
	last := updates[len(updates)-1].payload
	if value, ok := last[dwell].(int); !ok || value < 1 {
		t.Fatalf("dwell = %v", last[dwell])
	}
}

func TestOfflineSourceDoesNotPublish(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	source := onlineSource("0xabc")
	source.SetAvailability(devices.OfflineAvailability)
	engine := newTestEngine(t, sink, config, clock, source)
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")
	if calls := sink.of("offline"); len(calls) != 1 {
		t.Fatalf("offline calls = %+v", calls)
	}
	engine.handleFrame(frameJob{id: "0xabc", samples: []Sample{{Slot: 1, X: 0, Y: 1500}}, at: base})
	if updates := sink.updates(); len(updates) != 0 {
		t.Fatalf("offline source published %+v", updates)
	}
}

func TestQueueDropsWhenWorkerIsBusy(t *testing.T) {
	block := make(chan struct{})
	sink := &fakeSink{block: block, entered: make(chan struct{}, 1)}
	config := &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	dropped := 0
	engine := NewEngine(Deps{
		Config:  config,
		Source:  func(id string) (*devices.Device, bool) { return onlineSource(id), true },
		Sink:    sink,
		Clock:   clock,
		Queue:   1,
		Dropped: func() { dropped++ },
	})
	t.Cleanup(engine.Close)
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")
	engine.Start()

	// The first frame occupies the worker inside Update; the queue takes one more and drops
	// everything after it.
	go engine.Observe(frameAt(0, true, 0, 1500))
	<-sink.entered
	engine.Observe(frameAt(100, true, 0, 1500))
	engine.Observe(frameAt(200, true, 0, 1500))
	engine.Observe(frameAt(300, true, 0, 1500))
	close(block)
	if dropped == 0 {
		t.Fatalf("queue never dropped a frame (dropped=%d)", dropped)
	}
}

func TestCloseStopsWorkerAndIgnoresFrames(t *testing.T) {
	sink, config := &fakeSink{}, &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	engine := newTestEngine(t, sink, config, clock, onlineSource("0xabc"))
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Sync("0xabc")
	engine.Start()
	engine.Close()
	engine.Observe(frameAt(0, true, 0, 1500))
	engine.Sync("0xabc")
	if updates := sink.updates(); len(updates) != 0 {
		t.Fatalf("closed engine published %+v", updates)
	}
}

func TestSamplesFromFrameSkipsInactiveAndUnknownUnits(t *testing.T) {
	frame := &devices.SpatialFrame{Targets: []devices.SpatialTarget{
		{ID: 1, Active: true, X: spatialAxis(1000), Y: spatialAxis(2000)},
		{ID: 2, Active: false, X: spatialAxis(1), Y: spatialAxis(2)},
		{ID: 3, Active: true, X: &devices.SpatialAxis{Value: 1, Unit: "ft"}, Y: spatialAxis(2)},
	}}
	samples := samplesFromFrame(frame)
	if len(samples) != 1 || samples[0].Slot != 1 || samples[0].X != 1000 {
		t.Fatalf("samples = %+v", samples)
	}
}

func TestCounterAgeingSkipsUnoccupiedZonesAndCapsResume(t *testing.T) {
	idle := Signals{Zones: map[string]ZoneSignal{"desk": {Occupied: false, Dwell: 5 * time.Second}}, EmptyFor: 2 * time.Second}
	aged := advanceSignals(idle, 10*time.Second)
	if aged.Zones["desk"].Dwell != 5*time.Second {
		t.Fatalf("unoccupied dwell grew to %v", aged.Zones["desk"].Dwell)
	}
	if aged.EmptyFor != 12*time.Second {
		t.Fatalf("empty = %v, want 12s", aged.EmptyFor)
	}

	occupied := Signals{Zones: map[string]ZoneSignal{"desk": {Occupied: true, Dwell: 5 * time.Second}}}
	capped := advanceSignals(occupied, time.Hour)
	if capped.Zones["desk"].Dwell != 5*time.Second+maxCounterStep {
		t.Fatalf("capped dwell = %v, want %v", capped.Zones["desk"].Dwell, 5*time.Second+maxCounterStep)
	}
}

func TestConcurrentSyncRemoveAndClose(t *testing.T) {
	sink := &fakeSink{}
	config := &fakeConfig{mapped: map[string]bool{}, zones: map[string][]settings.SpatialZone{}}
	clock := &fakeClock{now: base}
	engine := NewEngine(Deps{
		Config: config,
		Source: func(id string) (*devices.Device, bool) { return onlineSource(id), true },
		Sink:   sink,
		Clock:  clock,
	})
	config.set("0xabc", true, []settings.SpatialZone{deskZone()})
	engine.Start()
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(3)
		go func() { defer group.Done(); engine.Sync("0xabc") }()
		go func() { defer group.Done(); engine.Observe(frameAt(i*100, true, 0, 1500)) }()
		go func() { defer group.Done(); engine.Remove("0xabc") }()
	}
	group.Wait()
	engine.Close()
}
