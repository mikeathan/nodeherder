package mocks

import (
	"node-herder/internal/automations"
	"node-herder/utils"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAutomationCacheConcurrentAccess(t *testing.T) {
	cache := NewMockAutomationStorage[automations.Automation](nil)
	item := automations.NewDevice("fixture")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = cache.Store("fixture", item)
				_, _ = cache.LoadFromCache("fixture")
				_ = cache.LoadAll()
				_ = cache.Delete("fixture")
				cache.ClearCache()
			}
		}()
	}
	wg.Wait()
	_ = cache.Store("b", automations.NewDevice("b"))
	_ = cache.Store("a", automations.NewDevice("a"))
	items := cache.LoadAll()
	if len(items) != 2 || items[0].GetId() != "a" || items[1].GetId() != "b" {
		t.Fatal("cache snapshot must preserve sorted key order")
	}
}

func TestClockCallbackResetAndStop(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	clock := NewMockClock(func() time.Time { return now })
	calls := 0
	var timer utils.Timer
	timer = clock.AfterFunc(time.Second, func() {
		calls++
		_ = clock.Now()
		if calls == 1 {
			timer.Reset(time.Second)
		}
	})
	clock.Advance(time.Second)
	clock.Advance(time.Second)
	clock.Advance(time.Second)
	if calls != 2 {
		t.Fatalf("callback reset lost/duplicated timer: %d calls", calls)
	}
	timer.Reset(time.Second)
	timer.Reset(time.Second)
	clock.Advance(time.Second)
	if calls != 3 {
		t.Fatalf("repeated reset duplicated timer: %d calls", calls)
	}
	timer.Reset(time.Second)
	if !timer.Stop() {
		t.Fatal("active timer was not stopped")
	}
	clock.Advance(time.Second)
	if calls != 3 {
		t.Fatal("stopped callback ran")
	}
}

func TestClockConcurrentAccess(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	clock := NewMockClock(func() time.Time { return now })
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				timer := clock.AfterFunc(time.Millisecond, func() { _ = clock.Now() })
				_ = clock.Now()
				timer.Reset(time.Millisecond)
				clock.Advance(time.Millisecond)
				timer.Stop()
			}
		}()
	}
	wg.Wait()
}

func TestMQTTHandlerSnapshotAndResponseCompletion(t *testing.T) {
	client := &MockMqttClient{}
	var oldCalls, newCalls atomic.Int32
	client.OnMessageHandler(func(string, []byte) { oldCalls.Add(1) })
	client.Publish("fixture/set", []byte("command"))
	client.OnMessageHandler(func(string, []byte) { newCalls.Add(1) })
	client.WaitResponses()
	if oldCalls.Load() != 1 || newCalls.Load() != 0 {
		t.Fatal("delayed delivery must use the handler registered at publish time")
	}
}

func TestMQTTConcurrentConfigurationAndDelivery(t *testing.T) {
	client := &MockMqttClient{}
	var calls atomic.Int32
	handler := func(string, []byte) { calls.Add(1) }
	client.OnMessageHandler(handler)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				client.OnMessageHandler(handler)
				client.AddResponse("fixture", "response")
				client.SetResponseDelay(time.Millisecond)
				client.Publish("fixture/set", "command")
			}
		}()
	}
	wg.Wait()
	client.WaitResponses()
	if calls.Load() != 160 {
		t.Fatalf("expected one delivery per publish, got %d", calls.Load())
	}
}
