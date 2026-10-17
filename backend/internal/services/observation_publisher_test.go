package services_test

import (
	"context"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type publisherFixture struct {
	publisher *services.ObservationPublisher
	clock     *mocks.MockClock
	device    *devices.Device
	updates   []*devices.UpdatePackage
}

func newPublisherFixture(t *testing.T, interval time.Duration, threshold float64) *publisherFixture {
	t.Helper()
	f := &publisherFixture{device: devices.NewDevice("sensor"), clock: mocks.NewMockClock(func() time.Time { return time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC) })}
	f.device.SetAvailable(true)
	for _, name := range []string{"distance", "link"} {
		e := devices.NewEntity(name)
		e.Type, e.AccessMode = bridge.NumericDataType, bridge.ReadAccessMode
		e.Data.SetValue(float64(100))
		f.device.Exposes[name] = e
	}
	f.publisher = services.NewObservationPublisher(context.Background(), f.device, f.clock, func(e *devices.Entity) settings.ReportingPolicy {
		return settings.ReportingPolicy{LiveInterval: interval, MinimumChange: threshold}
	}, func(p *devices.UpdatePackage) { f.updates = append(f.updates, p) })
	t.Cleanup(f.publisher.Close)
	return f
}
func (f *publisherFixture) report(name string, value any) {
	p := devices.NewUpdatePackage(f.device.Id)
	p.Data[name] = value
	f.publisher.Publish(p)
}

func TestObservationPublisherLatestReadingWithoutStarvation(t *testing.T) {
	f := newPublisherFixture(t, 30*time.Second, 0)
	for i := 1; i < 30; i++ {
		f.clock.Advance(time.Second)
		f.report("link", float64(100+i))
	}
	if len(f.updates) != 0 {
		t.Fatal("interval emitted early")
	}
	f.clock.Advance(time.Second)
	if len(f.updates) != 1 || f.updates[0].Data["link"] != float64(129) {
		t.Fatalf("final reading not flushed: %+v", f.updates)
	}
	f.clock.Advance(time.Minute)
	if len(f.updates) != 1 {
		t.Fatal("idle timer repeated output")
	}
}
func TestObservationPublisherUsesPublishedBaseline(t *testing.T) {
	f := newPublisherFixture(t, 10*time.Second, 10)
	f.report("distance", float64(120))
	f.report("distance", float64(105))
	f.clock.Advance(10 * time.Second)
	if len(f.updates) != 0 {
		t.Fatal("published an intermediate spike instead of latest value")
	}
	f.report("distance", float64(110))
	if len(f.updates) != 1 {
		t.Fatal("exact threshold not immediate after interval")
	}
	f.report("distance", float64(122))
	f.report("distance", float64(110))
	f.clock.Advance(10 * time.Second)
	if len(f.updates) != 1 {
		t.Fatal("return to baseline must cancel pending noise")
	}
}
func TestObservationPublisherIndependentFieldsAndBypasses(t *testing.T) {
	f := newPublisherFixture(t, 10*time.Second, 0)
	f.report("distance", float64(120))
	f.report("link", float64(130))
	f.clock.Advance(10 * time.Second)
	if len(f.updates) != 1 || len(f.updates[0].Data) != 2 {
		t.Fatal("due fields were not coalesced")
	}
	f.device.Exposes["link"].AccessMode = bridge.ReadWriteAccessMode
	f.report("link", float64(130))
	f.report("link", float64(130))
	f.report("distance", nil)
	f.report("distance", float64(130))
	f.clock.Advance(10 * time.Second)
	if len(f.updates) != 5 {
		t.Fatalf("writable/unknown/recovery evidence lost: %d", len(f.updates))
	}
}
func TestObservationPublisherCancellationAndOffline(t *testing.T) {
	for _, mode := range []string{"clear", "close", "offline"} {
		t.Run(mode, func(t *testing.T) {
			f := newPublisherFixture(t, time.Second, 0)
			f.report("link", float64(120))
			switch mode {
			case "clear":
				f.publisher.Clear()
			case "close":
				f.publisher.Close()
			case "offline":
				f.device.SetAvailable(false)
				f.publisher.Clear()
			}
			f.clock.Advance(time.Minute)
			if len(f.updates) != 0 {
				t.Fatal("cancelled observation published")
			}
		})
	}
}
func TestObservationPublisherContextAndConcurrentClose(t *testing.T) {
	f := newPublisherFixture(t, time.Second, 0)
	ctx, cancel := context.WithCancel(context.Background())
	var mu sync.Mutex
	emitted := 0
	p := services.NewObservationPublisher(ctx, f.device, f.clock, func(*devices.Entity) settings.ReportingPolicy {
		return settings.ReportingPolicy{LiveInterval: time.Second}
	}, func(*devices.UpdatePackage) { mu.Lock(); emitted++; mu.Unlock() })
	cancel()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				p.Clear()
				p.Close()
			}
		}()
	}
	wg.Wait()
	f.clock.Advance(time.Minute)
	update := devices.NewUpdatePackage(f.device.Id)
	update.Data["link"] = float64(150)
	p.Publish(update)
	mu.Lock()
	defer mu.Unlock()
	if emitted != 0 {
		t.Fatal("cancelled context emitted")
	}
}
func TestObservationPublisherImmediateMeasurements(t *testing.T) {
	f := newPublisherFixture(t, 0, 5)
	for _, value := range []float64{101, 105, 106, 110} {
		f.report("distance", value)
	}
	got := []any{}
	for _, p := range f.updates {
		got = append(got, p.Data["distance"])
	}
	if !reflect.DeepEqual(got, []any{float64(105), float64(110)}) {
		t.Fatalf("unexpected immediate values %v", got)
	}
}

type countingObservationClock struct {
	*mocks.MockClock
	callbacks []func()
}

func (c *countingObservationClock) AfterFunc(duration time.Duration, callback func()) utils.Timer {
	c.callbacks = append(c.callbacks, callback)
	return c.MockClock.AfterFunc(duration, callback)
}
func TestObservationPublisherBoundsTimersAndRejectsStaleCallback(t *testing.T) {
	device := devices.NewDevice("bounded")
	device.SetAvailable(true)
	for _, name := range []string{"fast", "slow"} {
		e := devices.NewEntity(name)
		e.Type, e.AccessMode = bridge.NumericDataType, bridge.ReadAccessMode
		e.Data.SetValue(float64(0))
		device.Exposes[name] = e
	}
	clock := &countingObservationClock{MockClock: mocks.NewMockClock(func() time.Time { return time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC) })}
	updates := []*devices.UpdatePackage{}
	p := services.NewObservationPublisher(context.Background(), device, clock, func(e *devices.Entity) settings.ReportingPolicy {
		interval := 10 * time.Second
		if e.Name == "fast" {
			interval = time.Second
		}
		return settings.ReportingPolicy{LiveInterval: interval}
	}, func(update *devices.UpdatePackage) { updates = append(updates, update) })
	defer p.Close()
	report := func(name string, value float64) {
		update := devices.NewUpdatePackage(device.Id)
		update.Data[name] = value
		p.Publish(update)
	}
	for i := 1; i <= 1000; i++ {
		report("slow", float64(i))
	}
	if len(clock.callbacks) != 1 {
		t.Fatalf("allocated %d timers for one pending field", len(clock.callbacks))
	}
	oldCallback := clock.callbacks[0]
	report("fast", 1)
	oldCallback() // A stopped timer can already have entered its callback.
	clock.Advance(time.Second)
	if len(updates) != 1 || updates[0].Data["fast"] != float64(1) {
		t.Fatal("stale callback disturbed earlier deadline")
	}
	clock.Advance(9 * time.Second)
	if len(updates) != 2 || updates[1].Data["slow"] != float64(1000) {
		t.Fatal("later field did not flush latest value")
	}
}

func TestObservationPublisherReconfiguresPendingWithoutLosingFinalValue(t *testing.T) {
	f := newPublisherFixture(t, time.Second, 0)
	interval := 10 * time.Second
	published := []float64{}
	p := services.NewObservationPublisher(context.Background(), f.device, f.clock, func(*devices.Entity) settings.ReportingPolicy {
		return settings.ReportingPolicy{LiveInterval: interval}
	}, func(update *devices.UpdatePackage) { published = append(published, update.Data["link"].(float64)) })
	defer p.Close()
	update := devices.NewUpdatePackage(f.device.Id)
	update.Data["link"] = float64(120)
	p.Publish(update)
	f.clock.Advance(5 * time.Second)
	p.Reconfigure() // Unrelated setting edit retains the same deadline and final value.
	f.clock.Advance(5 * time.Second)
	if !reflect.DeepEqual(published, []float64{120}) {
		t.Fatalf("configuration discarded pending value: %v", published)
	}
	update = devices.NewUpdatePackage(f.device.Id)
	update.Data["link"] = float64(130)
	p.Publish(update)
	interval = 0
	p.Reconfigure()
	if !reflect.DeepEqual(published, []float64{120, 130}) {
		t.Fatal("shortened interval did not publish retained reading")
	}
	f.clock.Advance(time.Minute)
	if len(published) != 2 {
		t.Fatal("stale configuration timer published twice")
	}
}

// Expose context registrations without relying on finalizers or goroutine counts.
// Hiding the underlying cancelCtx value makes context use this AfterFunc adapter.
type trackedShutdownContext struct {
	context.Context
	active atomic.Int32
}

func (c *trackedShutdownContext) Value(any) any { return nil }
func (c *trackedShutdownContext) AfterFunc(callback func()) func() bool {
	c.active.Add(1)
	var once sync.Once
	release := func() { once.Do(func() { c.active.Add(-1) }) }
	stop := context.AfterFunc(c.Context, func() { release(); callback() })
	return func() bool {
		stopped := stop()
		if stopped {
			release()
		}
		return stopped
	}
}
func newTrackedShutdownContext(t *testing.T) *trackedShutdownContext {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &trackedShutdownContext{Context: ctx}
}

func TestObservationPublisherCloseUnregistersShutdownCallback(t *testing.T) {
	f := newPublisherFixture(t, time.Second, 0)
	ctx := newTrackedShutdownContext(t)
	published := 0
	p := services.NewObservationPublisher(ctx, f.device, f.clock, func(*devices.Entity) settings.ReportingPolicy {
		return settings.ReportingPolicy{LiveInterval: time.Second}
	}, func(*devices.UpdatePackage) { published++ })
	if ctx.active.Load() != 1 {
		t.Fatal("shutdown callback not registered")
	}
	update := devices.NewUpdatePackage(f.device.Id)
	update.Data["link"] = float64(120)
	p.Publish(update)
	p.Close()
	p.Close()
	if ctx.active.Load() != 0 {
		t.Fatal("closed publisher retained by live hub context")
	}
	f.clock.Advance(time.Minute)
	if published != 0 {
		t.Fatal("retired publisher emitted pending reading")
	}
}

func TestObservationPublisherAlreadyCancelledConstruction(t *testing.T) {
	f := newPublisherFixture(t, time.Second, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := services.NewObservationPublisher(ctx, f.device, f.clock, func(*devices.Entity) settings.ReportingPolicy { return settings.ReportingPolicy{} }, func(*devices.UpdatePackage) { t.Error("cancelled publisher emitted") })
	p.Close()
	update := devices.NewUpdatePackage(f.device.Id)
	update.Data["link"] = float64(120)
	p.Publish(update)
}
