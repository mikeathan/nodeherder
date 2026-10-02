package automations_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"node-herder/utils/storage"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type reloadRegistrar struct {
	services.DeviceRegistrar
	lookup func(string) error
}

type sourceStateRegistrar struct {
	reloadRegistrar
	disabled  bool
	bridgeErr error
}

func (r *sourceStateRegistrar) FindBridgeInfo(id string) (*devices.BridgeInfo, error) {
	if id == "source" && r.bridgeErr != nil {
		return nil, r.bridgeErr
	}
	b, err := r.reloadRegistrar.FindBridgeInfo(id)
	b.Disabled = id == "source" && r.disabled
	return b, err
}

func TestEngineExplicitSourceDisable(t *testing.T) {
	for _, path := range []string{"reload", "add"} {
		for _, cleanupFailure := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cleanupFailure=%v", path, cleanupFailure), func(t *testing.T) {
				r := &sourceStateRegistrar{}
				probe := &handlerProbe{}
				commands := 0
				var lastCommand map[string]bool
				engine := automations.NewEngine([]automations.AutomationHandler{probe}, r, &delayedLifecycleClient{publish: func(_ string, payload interface{}) {
					commands++
					if err := json.Unmarshal(payload.([]byte), &lastCommand); err != nil {
						t.Error(err)
					}
				}})
				disk := reloadDisk(t)
				engine.WithStorage(disk)
				old := reloadRecipe(false)
				if err := engine.Add(old); err != nil {
					t.Fatal(err)
				}
				other := reloadRecipe(true)
				other.Id = "other"
				if err := engine.Add(other); err != nil {
					t.Fatal(err)
				}
				r.disabled = true
				if cleanupFailure {
					probe.prepareErr = errors.New("cleanup unavailable")
				}
				if path == "reload" {
					engine.Initialize()
				} else if err := engine.Add(reloadRecipe(true)); err == nil {
					t.Fatal("disabled source accepted")
				}
				if _, err := engine.Load("source"); err == nil {
					t.Fatal("disabled source remains executable")
				}
				if engine.IsAutomationEnabled("source") {
					t.Fatal("disabled source enabled")
				}
				if err := engine.HandleManual("source", "action"); err == nil {
					t.Fatal("disabled manual accepted")
				}
				device, _ := r.LookupById("source")
				engine.HandleDevice(device, map[string]interface{}{"action": "pressed"})
				if commands != 0 {
					t.Fatal("disabled source published")
				}
				if _, err := disk.Load("source"); err != nil {
					t.Fatal("recipe lost", err)
				}
				probe.prepareErr = nil
				before := probe.completed
				engine.Initialize()
				if probe.completed <= before {
					t.Fatal("retirement not retried")
				}
				if err := engine.HandleManual("other", "action"); err != nil {
					t.Fatal(err)
				}
				if commands != 1 {
					t.Fatal("other source affected")
				}
				r.disabled = false
				r.lookup = func(id string) error {
					if id == "target" {
						return errors.New("target unavailable")
					}
					return nil
				}
				engine.Initialize()
				if _, err := engine.Load("source"); err == nil {
					t.Fatal("failed re-enable published")
				}
				r.lookup = nil
				engine.Initialize()
				if err := engine.HandleManual("source", "action"); err != nil {
					t.Fatal(err)
				}
				if commands != 2 {
					t.Fatal("re-enable did not execute")
				}
				if lastCommand["presence"] {
					t.Fatal("re-enable used rejected replacement instead of persisted recipe")
				}
			})
		}
	}
}

func TestEngineSourceLookupFailureRetainsGeneration(t *testing.T) {
	r := &sourceStateRegistrar{}
	engine := automations.NewEngine(nil, r, &mocks.MockMqttClient{})
	engine.WithStorage(reloadDisk(t))
	old := reloadRecipe(false)
	if err := engine.Add(old); err != nil {
		t.Fatal(err)
	}
	r.bridgeErr = errors.New("bridge unavailable")
	engine.Initialize()
	got, err := engine.Load("source")
	if err != nil || got != old {
		t.Fatal("ordinary bridge failure discarded working generation", err)
	}
}

func TestEngineInitiallyDisabledSource(t *testing.T) {
	r := &sourceStateRegistrar{disabled: true}
	engine := automations.NewEngine(nil, r, &mocks.MockMqttClient{})
	disk := reloadDisk(t)
	engine.WithStorage(disk)
	if err := disk.Store("source", reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	engine.Initialize()
	if len(engine.GetAllTriggers()) != 0 {
		t.Fatal("initial disabled source published")
	}
	if err := engine.Add(reloadRecipe(true)); !errors.Is(err, automations.ErrAutomationSourceDisabled) {
		t.Fatal("disable cause lost", err)
	}
	r.disabled = false
	engine.Initialize()
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatal(err)
	}
}

func TestEngineSourceDisableDoesNotJoinCommittedCommand(t *testing.T) {
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	r := &sourceStateRegistrar{}
	engine := automations.NewEngine(nil, r, &delayedLifecycleClient{publish: func(string, interface{}) { close(entered); <-release }})
	engine.WithStorage(reloadDisk(t))
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	var once sync.Once
	defer once.Do(func() { close(release) })
	go func() {
		defer close(done)
		if err := engine.HandleManual("source", "action"); err != nil {
			t.Error(err)
		}
	}()
	awaitReload(t, entered)
	r.disabled = true
	disabled := make(chan struct{})
	go func() { defer close(disabled); engine.Initialize() }()
	awaitReload(t, disabled)
	if err := engine.HandleManual("source", "action"); err == nil {
		t.Fatal("new execution allowed after disable")
	}
	once.Do(func() { close(release) })
	awaitReload(t, done)
}

func TestEngineDisabledSourceRetiresSchedules(t *testing.T) {
	for _, failure := range []string{"none", "prepare", "activate"} {
		t.Run(failure, func(t *testing.T) {
			clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
			handler := automations.NewAutomationScheduler(automations.WithAutomationsFuncs(), automations.WithSchedulerClock(clock))
			probe := &handlerProbe{}
			r := &sourceStateRegistrar{}
			commands := 0
			engine := automations.NewEngine([]automations.AutomationHandler{handler, probe}, r, &delayedLifecycleClient{publish: func(string, interface{}) { commands++ }})
			engine.WithStorage(reloadDisk(t))
			old := reloadRecipe(false)
			old.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}}
			if err := engine.Add(old); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := handler.Process(automations.NewDevice("source")); err != nil {
					t.Error(err)
				}
			}()
			r.disabled = true
			cleanupErr := errors.New("retirement failed")
			if failure == "prepare" {
				probe.prepareErr = cleanupErr
			}
			if failure == "activate" {
				probe.activateErr = cleanupErr
			}
			err := engine.Add(reloadRecipe(true))
			if !errors.Is(err, automations.ErrAutomationSourceDisabled) {
				t.Fatal("disable cause lost", err)
			}
			if failure != "none" && !errors.Is(err, cleanupErr) {
				t.Fatal("cleanup cause lost", err)
			}
			if failure == "none" && handler.IsRunning(old) {
				t.Fatal("disabled schedule still running")
			}
			clock.Advance(time.Second)
			if failure == "none" && old.IsEnabled() {
				t.Fatal("retired timer fired")
			}
			if err := engine.HandleManual("source", "action"); err == nil {
				t.Fatal("timer restored disabled execution")
			}
			device, _ := r.LookupById("source")
			engine.HandleDevice(device, map[string]interface{}{"action": "pressed"})
			if commands != 0 {
				t.Fatal("disabled source published")
			}
			probe.prepareErr, probe.activateErr = nil, nil
			engine.Initialize()
			if handler.IsRunning(old) {
				t.Fatal("cleanup retry left schedule active")
			}
			r.disabled = false
			engine.Initialize()
			current, err := engine.Load("source")
			if err != nil || !handler.IsRunning(old) {
				t.Fatal("re-enable did not restore schedules", err)
			}
			if current.IsEnabled() {
				t.Fatal("re-enable bypassed schedule window")
			}
			clock.Advance(24 * time.Hour)
			if err := engine.HandleManual("source", "action"); err != nil {
				t.Fatal("restored schedule did not enable execution", err)
			}
			if commands != 1 {
				t.Fatal("restored schedule execution count", commands)
			}
		})
	}
}

func (*reloadRegistrar) FindBridgeInfo(string) (*devices.BridgeInfo, error) {
	b := &devices.BridgeInfo{FriendlyName: "source"}
	b.Definition.Exposes = []devices.BridgeExpose{{Property: "presence", Type: "binary", ValueOff: false, ValueOn: true}}
	return b, nil
}
func (r *reloadRegistrar) LookupById(id string) (*devices.Device, error) {
	if r.lookup != nil {
		if err := r.lookup(id); err != nil {
			return nil, err
		}
	}
	d := devices.NewDevice(id)
	d.FriendlyName = id
	e := devices.NewEntity("action")
	e.Data.SetValue("pressed")
	d.Exposes["action"] = e
	return d, nil
}

func reloadRecipe(value bool) *automations.Device {
	a := automations.NewDevice("source")
	a.SetEnabled(true)
	t := automations.NewDeviceTrigger("action")
	t.Conditions = []automations.Condition{&automations.ExposeCondition{BaseCondition: *automations.NewBaseCondition(automations.ExposeConditionType, utils.Equals), Name: "action", Value: "pressed"}}
	command := automations.NewTriggerAction()
	command.Id = "target"
	command.Exposes = []*automations.MqttTriggerActionExpose{{Name: "presence", Data: value}}
	t.Actions = []automations.MqttAction{command}
	a.Triggers = append(a.Triggers, t)
	return a
}

func reloadDisk(t *testing.T) storage.Storage[automations.Automation] {
	t.Helper()
	s := automations.NewAutomationSerialiser()
	return storage.NewJsonDiskStorage(t.TempDir(), nil, s.Unmarshal)
}

func awaitReload(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reload phase blocked")
	}
}

func TestEngineReadyReload(t *testing.T) {
	for _, initial := range []bool{false, true} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("initial=%t/failure=%t", initial, fail), func(t *testing.T) {
				disk := reloadDisk(t)
				r := &reloadRegistrar{}
				var commands []bool
				client := &delayedLifecycleClient{publish: func(_ string, p interface{}) {
					var data map[string]bool
					if err := json.Unmarshal(p.([]byte), &data); err != nil {
						t.Error(err)
					}
					commands = append(commands, data["presence"])
				}}
				engine := automations.NewEngine(nil, r, client)
				engine.WithStorage(disk)
				if !initial {
					if err := engine.Add(reloadRecipe(false)); err != nil {
						t.Fatal(err)
					}
				}
				if err := disk.Store("source", reloadRecipe(true)); err != nil {
					t.Fatal(err)
				}
				entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
				var once sync.Once
				r.lookup = func(id string) error {
					if id == "target" {
						once.Do(func() { close(entered) })
						<-release
						if fail {
							return errors.New("target unavailable")
						}
					}
					return nil
				}
				var releaseOnce sync.Once
				defer releaseOnce.Do(func() { close(release) })
				go func() { defer close(done); engine.Initialize() }()
				awaitReload(t, entered)
				err := engine.HandleManual("source", "action")
				if initial {
					if err == nil || engine.IsAutomationEnabled("source") {
						t.Error("unready initial generation exposed")
					}
				} else {
					if err != nil {
						t.Error(err)
					}
					if len(commands) != 1 || commands[0] {
						t.Fatalf("old generation not used: %v", commands)
					}
					device, _ := r.LookupById("source")
					engine.HandleDevice(device, map[string]interface{}{"action": "pressed"})
					if len(commands) != 2 || commands[1] {
						t.Fatalf("physical event lost/misrouted during reload: %v", commands)
					}
				}
				releaseOnce.Do(func() { close(release) })
				awaitReload(t, done)
				before := len(commands)
				err = engine.HandleManual("source", "action")
				if initial && fail {
					if err == nil || len(commands) != before {
						t.Fatal("failed initial generation became executable")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if len(commands) != before+1 || commands[before] != !fail {
					t.Fatalf("wrong committed generation: %v", commands)
				}
			})
		}
	}
}

type failingRecipeStorage struct {
	storage.Storage[automations.Automation]
	storeErr, deleteErr error
}

func (s *failingRecipeStorage) Store(id string, a automations.Automation) error {
	if s.storeErr != nil {
		return s.storeErr
	}
	return s.Storage.Store(id, a)
}
func (s *failingRecipeStorage) Delete(id string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	return s.Storage.Delete(id)
}

func TestEnginePersistenceAndTriggerErrors(t *testing.T) {
	s := &failingRecipeStorage{Storage: reloadDisk(t)}
	r := &reloadRegistrar{}
	var commands int
	engine := automations.NewEngine(nil, r, &delayedLifecycleClient{publish: func(string, interface{}) { commands++ }})
	engine.WithStorage(s)
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	old, _ := engine.Load("source")
	s.storeErr = errors.New("save failed")
	if err := engine.Add(reloadRecipe(true)); !errors.Is(err, s.storeErr) {
		t.Fatalf("save failure lost: %v", err)
	}
	if got, _ := engine.Load("source"); got != old {
		t.Fatal("failed save replaced active generation")
	}
	if err := engine.DeleteTrigger("source", 0); !errors.Is(err, s.storeErr) {
		t.Fatalf("trigger delete save failure lost: %v", err)
	}
	if len(old.GetTriggers()) != 1 {
		t.Fatal("failed trigger deletion mutated live recipe")
	}
	for _, index := range []int{-1, 1} {
		if err := engine.DeleteTrigger("source", index); err == nil {
			t.Fatal("invalid trigger index accepted")
		}
	}
	s.deleteErr = errors.New("delete failed")
	if err := engine.Delete("source"); !errors.Is(err, s.deleteErr) {
		t.Fatalf("delete failure lost: %v", err)
	}
	if err := engine.HandleManual("source", "missing"); err == nil {
		t.Fatal("unknown trigger reported success")
	}
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatal(err)
	}
	if commands != 1 {
		t.Fatalf("working command lost after failures: %d", commands)
	}
	// An unconfigured action must fail without skipping the later valid action.
	a := reloadRecipe(true)
	good := a.Triggers[0].GetActions()[0]
	if err := good.Configure(r, &delayedLifecycleClient{publish: func(string, interface{}) { commands++ }}); err != nil {
		t.Fatal(err)
	}
	a.Triggers[0].(*automations.DeviceTrigger).Actions = []automations.MqttAction{automations.NewTriggerAction(), good}
	device, _ := r.LookupById("source")
	if err := a.EvaluateTrigger(automations.NewDeviceEvent(device, nil), "action"); err == nil {
		t.Fatal("action error reported success")
	}
	if commands != 2 {
		t.Fatal("action failure skipped later working action")
	}
}

func TestSchedulerGenerationReplacement(t *testing.T) {
	for _, scenario := range []string{"identical", "type change", "remove job", "invalid time", "unknown type"} {
		t.Run(scenario, func(t *testing.T) {
			clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
			handler := automations.NewAutomationScheduler(automations.WithAutomationsFuncs(), automations.WithSchedulerClock(clock))
			old := reloadRecipe(false)
			old.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}, {StartAt: "12:00:02", Type: "disable"}}
			if err := handler.Process(old); err != nil {
				t.Fatal(err)
			}
			defer func() {
				empty := automations.NewDevice("source")
				if err := handler.Process(empty); err != nil {
					t.Error(err)
				}
			}()
			next := reloadRecipe(true)
			next.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}, {StartAt: "12:00:02", Type: "disable"}}
			switch scenario {
			case "type change":
				next.Schedules[0].Type = "disable"
			case "remove job":
				next.Schedules = next.Schedules[:1]
			case "invalid time":
				next.Schedules[0].StartAt = "invalid"
			case "unknown type":
				next.Schedules[0].Type = "unknown"
			}
			err := handler.Process(next)
			if scenario == "invalid time" || scenario == "unknown type" {
				if err == nil {
					t.Fatal("invalid replacement reported success")
				}
				clock.Advance(time.Second)
				if !old.IsEnabled() {
					t.Fatal("failed replacement stopped old schedule")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if next.IsEnabled() {
				t.Fatal("schedule state not preserved/initialized")
			}
			clock.Advance(time.Second)
			if old.IsEnabled() {
				t.Fatal("old generation still owns schedule")
			}
			if next.IsEnabled() != (scenario != "type change") {
				t.Fatal("replacement schedule did not use new time/type binding")
			}
			clock.Advance(time.Second)
			if next.IsEnabled() != (scenario == "remove job") {
				t.Fatal("removed/changed job still firing")
			}
		})
	}
}

type handlerProbe struct {
	prepareErr, activateErr          error
	prepare                          func()
	activated, rolledBack, completed int
}

func (*handlerProbe) Type() string                          { return "probe" }
func (*handlerProbe) IsRunning(automations.Automation) bool { return false }
func (h *handlerProbe) Process(a automations.Automation) error {
	c, err := h.Prepare(a)
	if err != nil {
		return err
	}
	if err = c.Activate(); err != nil {
		c.Rollback()
		return err
	}
	c.Complete()
	return nil
}
func (h *handlerProbe) Prepare(automations.Automation) (automations.AutomationHandlerChange, error) {
	if h.prepare != nil {
		h.prepare()
	}
	if h.prepareErr != nil {
		return nil, h.prepareErr
	}
	return &handlerProbeChange{h}, nil
}

type handlerProbeChange struct{ h *handlerProbe }

func (c *handlerProbeChange) Activate() error { c.h.activated++; return c.h.activateErr }
func (c *handlerProbeChange) Rollback()       { c.h.rolledBack++ }
func (c *handlerProbeChange) Complete()       { c.h.completed++ }

func TestEngineHandlerRollback(t *testing.T) {
	for _, failure := range []string{"prepare", "activate", "store"} {
		t.Run(failure, func(t *testing.T) {
			first, second := &handlerProbe{}, &handlerProbe{}
			s := &failingRecipeStorage{Storage: reloadDisk(t)}
			engine := automations.NewEngine([]automations.AutomationHandler{first, second}, &reloadRegistrar{}, &mocks.MockMqttClient{})
			engine.WithStorage(s)
			if err := engine.Add(reloadRecipe(false)); err != nil {
				t.Fatal(err)
			}
			old, _ := engine.Load("source")
			sentinel := errors.New(failure)
			switch failure {
			case "prepare":
				second.prepareErr = sentinel
			case "activate":
				second.activateErr = sentinel
			case "store":
				s.storeErr = sentinel
			}
			if err := engine.Add(reloadRecipe(true)); !errors.Is(err, sentinel) {
				t.Fatalf("failure not propagated: %v", err)
			}
			if got, _ := engine.Load("source"); got != old {
				t.Fatal("handler/save failure published candidate")
			}
			if first.rolledBack != 1 || first.completed != 1 {
				t.Fatal("previously prepared handler not rolled back")
			}
			wantRollback := 1
			if failure == "prepare" {
				wantRollback = 0
			}
			if second.rolledBack != wantRollback || second.completed != 1 {
				t.Fatal("failing handler lifecycle incorrect")
			}
			second.prepareErr, second.activateErr, s.storeErr = nil, nil, nil
			if err := engine.Add(reloadRecipe(true)); err != nil {
				t.Fatalf("rollback left management ownership locked: %v", err)
			}
		})
	}
}

func TestEngineScheduleSaveRollbackAndRemoval(t *testing.T) {
	for _, removal := range []string{"delete", "disk reload", "remove schedules"} {
		t.Run(removal, func(t *testing.T) {
			clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
			handler := automations.NewAutomationScheduler(automations.WithAutomationsFuncs(), automations.WithSchedulerClock(clock))
			s := &failingRecipeStorage{Storage: reloadDisk(t)}
			engine := automations.NewEngine([]automations.AutomationHandler{handler}, &reloadRegistrar{}, &mocks.MockMqttClient{})
			engine.WithStorage(s)
			old := reloadRecipe(false)
			old.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}}
			if err := engine.Add(old); err != nil {
				t.Fatal(err)
			}
			defer func() { empty := automations.NewDevice("source"); _ = handler.Process(empty) }()
			next := reloadRecipe(true)
			next.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:02", Type: "enable"}}
			s.storeErr = errors.New("save failed")
			if err := engine.Add(next); !errors.Is(err, s.storeErr) {
				t.Fatal("save error lost")
			}
			clock.Advance(time.Second)
			if !old.IsEnabled() || next.IsEnabled() || !handler.IsRunning(old) {
				t.Fatal("save failure changed schedule ownership")
			}
			s.storeErr = nil
			s.deleteErr = errors.New("delete failed")
			if err := engine.Delete("source"); !errors.Is(err, s.deleteErr) {
				t.Fatal("delete error lost")
			}
			if !handler.IsRunning(old) {
				t.Fatal("failed delete stopped active jobs")
			}
			s.deleteErr = nil
			switch removal {
			case "delete":
				if err := engine.Delete("source"); err != nil {
					t.Fatal(err)
				}
			case "disk reload":
				if err := s.Storage.Delete("source"); err != nil {
					t.Fatal(err)
				}
				engine.Initialize()
			case "remove schedules":
				if err := engine.Add(reloadRecipe(true)); err != nil {
					t.Fatal(err)
				}
			}
			if handler.IsRunning(old) {
				t.Fatal("removed jobs remain active")
			}
			old.SetEnabled(false)
			clock.Advance(48 * time.Hour)
			if old.IsEnabled() || next.IsEnabled() {
				t.Fatal("retired/discarded schedule fired")
			}
			if removal != "remove schedules" {
				if _, err := engine.Load("source"); err == nil {
					t.Fatal("deleted generation still visible")
				}
			}
		})
	}
}

func TestEngineUpdateCallbackReentry(t *testing.T) {
	r := &reloadRegistrar{}
	engine := automations.NewEngine(nil, r, &mocks.MockMqttClient{})
	engine.WithStorage(reloadDisk(t))
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	var probed atomic.Bool
	r.lookup = func(id string) error {
		// Probe once: the deferred reload requested below reconfigures and would
		// otherwise request yet another reload from inside every reload.
		if id != "target" || probed.Swap(true) {
			return nil
		}
		if _, err := engine.Load("source"); err != nil {
			t.Error(err)
		}
		if len(engine.GetAllTriggers()) != 1 || !engine.IsAutomationEnabled("source") {
			t.Error("callback cannot read active generation")
		}
		if err := engine.HandleManual("source", "action"); err != nil {
			t.Error(err)
		}
		if err := engine.Add(reloadRecipe(true)); !errors.Is(err, automations.ErrAutomationUpdateInProgress) {
			t.Errorf("re-entry should fail busy: %v", err)
		}
		if err := engine.Delete("source"); !errors.Is(err, automations.ErrAutomationUpdateInProgress) {
			t.Error("delete re-entry not rejected")
		}
		if err := engine.DeleteTrigger("source", 0); !errors.Is(err, automations.ErrAutomationUpdateInProgress) {
			t.Error("trigger-delete re-entry not rejected")
		}
		engine.Initialize() // Re-entrant reload is deferred, never blocks.
		return nil
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := engine.Add(reloadRecipe(true)); err != nil {
			t.Error(err)
		}
	}()
	awaitReload(t, done)
}

func TestEngineReloadDuringUpdateRunsAfterRelease(t *testing.T) {
	r := &reloadRegistrar{}
	disk := reloadDisk(t)
	engine := automations.NewEngine(nil, r, &mocks.MockMqttClient{})
	engine.WithStorage(disk)
	entered, release := make(chan struct{}), make(chan struct{})
	var blockOnce sync.Once
	r.lookup = func(id string) error {
		blockOnce.Do(func() { close(entered); <-release })
		return nil
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := engine.Add(reloadRecipe(true)); err != nil {
			t.Error(err)
		}
	}()
	awaitReload(t, entered)

	// A bridge reload arrives while the save owns the gate: it must not block,
	// and the new recipe it would discover must still be published afterwards.
	second := reloadRecipe(true)
	second.Id = "second"
	if err := disk.Store("second", second); err != nil {
		t.Fatal(err)
	}
	reloaded := make(chan struct{})
	go func() { defer close(reloaded); engine.Initialize() }()
	awaitReload(t, reloaded)
	if _, err := engine.Load("second"); err == nil {
		t.Fatal("reload ran while the gate was owned")
	}

	close(release)
	awaitReload(t, done)
	if _, err := engine.Load("second"); err != nil {
		t.Fatalf("deferred reload was dropped: %v", err)
	}
	if _, err := engine.Load("source"); err != nil {
		t.Fatalf("saved automation lost by deferred reload: %v", err)
	}
}

func TestEngineConcurrentGenerationDelivery(t *testing.T) {
	var commands atomic.Int32
	client := &delayedLifecycleClient{publish: func(_ string, p interface{}) {
		var data map[string]bool
		if err := json.Unmarshal(p.([]byte), &data); err != nil {
			t.Error(err)
		}
		commands.Add(1)
	}}
	r := &reloadRegistrar{}
	engine := automations.NewEngine(nil, r, client)
	engine.WithStorage(reloadDisk(t))
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(5)
	for worker := 0; worker < 4; worker++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if err := engine.HandleManual("source", "action"); err != nil {
					t.Error(err)
				}
				if !engine.IsAutomationEnabled("source") {
					t.Error("enabled generation disappeared during Add")
				}
				if len(engine.GetAllTriggers()) != 1 {
					t.Error("generation disappeared")
				}
			}
		}()
	}
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			if err := engine.Add(reloadRecipe(i%2 == 0)); err != nil {
				t.Error(err)
			}
		}
	}()
	wg.Wait()
	if commands.Load() != 400 {
		t.Fatalf("command loss/duplication during replacement: %d", commands.Load())
	}
}

func TestEngineLaterHandlerFailurePreservesSchedule(t *testing.T) {
	for _, phase := range []string{"prepare", "activate"} {
		t.Run(phase, func(t *testing.T) {
			clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
			handler := automations.NewAutomationScheduler(automations.WithSchedulerClock(clock), automations.WithAutomationsFuncs())
			probe := &handlerProbe{}
			engine := automations.NewEngine([]automations.AutomationHandler{handler, probe}, &reloadRegistrar{}, &mocks.MockMqttClient{})
			engine.WithStorage(reloadDisk(t))
			old := reloadRecipe(false)
			old.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}}
			if err := engine.Add(old); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = handler.Process(automations.NewDevice("source")) }()
			next := reloadRecipe(true)
			next.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:02", Type: "enable"}}
			failure := errors.New(phase)
			if phase == "prepare" {
				probe.prepareErr = failure
			} else {
				probe.activateErr = failure
			}
			if err := engine.Add(next); !errors.Is(err, failure) {
				t.Fatal("handler failure hidden")
			}
			if got, _ := engine.Load("source"); got != old {
				t.Fatal("failed handler replaced generation")
			}
			if !handler.IsRunning(old) {
				t.Fatal("failed handler stopped previous schedule")
			}
			clock.Advance(time.Second)
			if !old.IsEnabled() {
				t.Fatal("old schedule lost during rollback")
			}
			clock.Advance(time.Second)
			// The candidate may retain its input Enabled value when never activated;
			// it must never have a running scheduler after rollback.
			if !handler.IsRunning(old) {
				t.Fatal("rollback discarded old registry entry")
			}
		})
	}
}

func TestEngineMalformedReplacement(t *testing.T) {
	for _, malformed := range []string{"nil automation", "nil trigger", "nil condition", "nil action", "nil expose", "nil step"} {
		t.Run(malformed, func(t *testing.T) {
			engine := automations.NewEngine(nil, &reloadRegistrar{}, &mocks.MockMqttClient{})
			engine.WithStorage(reloadDisk(t))
			old := reloadRecipe(false)
			if err := engine.Add(old); err != nil {
				t.Fatal(err)
			}
			a := reloadRecipe(true)
			trigger := a.Triggers[0].(*automations.DeviceTrigger)
			var candidate automations.Automation = a
			switch malformed {
			case "nil automation":
				candidate = nil
			case "nil trigger":
				a.Triggers[0] = nil
			case "nil condition":
				trigger.Conditions[0] = nil
			case "nil action":
				trigger.Actions[0] = nil
			case "nil expose":
				trigger.Actions[0].(*automations.MqttTriggerAction).Exposes = []*automations.MqttTriggerActionExpose{nil}
			case "nil step":
				action := automations.NewStepAction()
				action.Id = "target"
				action.Steps = []*automations.Step{nil}
				trigger.Actions[0] = action
			}
			if err := engine.Add(candidate); err == nil {
				t.Fatal("malformed candidate accepted")
			}
			if got, _ := engine.Load("source"); got != old {
				t.Fatal("malformed candidate replaced runtime")
			}
		})
	}
}

type failedExecutionAction struct {
	*automations.MqttTriggerAction
	failure error
}

func (a *failedExecutionAction) Execute(automations.AutomationContext) error { return a.failure }

func TestEngineManualExecutionFailureAndNoop(t *testing.T) {
	r := &reloadRegistrar{}
	var commands int
	engine := automations.NewEngine(nil, r, &delayedLifecycleClient{publish: func(string, interface{}) { commands++ }})
	engine.WithStorage(reloadDisk(t))
	a := reloadRecipe(false)
	good := a.Triggers[0].GetActions()[0]
	sentinel := errors.New("execution failed")
	bad := &failedExecutionAction{MqttTriggerAction: automations.NewTriggerAction(), failure: sentinel}
	bad.Id = "target"
	a.Triggers[0].(*automations.DeviceTrigger).Actions = []automations.MqttAction{bad, good}
	if err := engine.Add(a); err != nil {
		t.Fatal(err)
	}
	if err := engine.HandleManual("source", "action"); !errors.Is(err, sentinel) {
		t.Fatalf("manual action cause lost: %v", err)
	}
	if commands != 1 {
		t.Fatal("failed first action skipped valid second action")
	}
	a.SetEnabled(false)
	if err := engine.HandleManual("source", "action"); err == nil {
		t.Fatal("disabled automation executed")
	}
	a.SetEnabled(true)
	r.lookup = func(string) error { return sentinel }
	if err := engine.HandleManual("source", "action"); !errors.Is(err, sentinel) {
		t.Fatal("source lookup failure lost")
	}
	r.lookup = nil
	noop := reloadRecipe(true)
	condition := noop.Triggers[0].GetConditions()[0].(*automations.ExposeCondition)
	condition.Value = "not pressed"
	if err := condition.InitHandlers(utils.NewRealClock()); err != nil {
		t.Fatal(err)
	}
	if err := engine.Add(noop); err != nil {
		t.Fatal(err)
	}
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatalf("condition no-op changed semantics: %v", err)
	}
	if commands != 1 {
		t.Fatal("condition mismatch emitted command")
	}
}

func TestEngineTriggerDeleteAndRepeatedReload(t *testing.T) {
	r := &reloadRegistrar{}
	var commands int
	engine := automations.NewEngine(nil, r, &delayedLifecycleClient{publish: func(string, interface{}) { commands++ }})
	engine.WithStorage(reloadDisk(t))
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	old, _ := engine.Load("source")
	if err := engine.Add(old); err != nil {
		t.Fatal(err)
	}
	if current, _ := engine.Load("source"); current == old {
		t.Fatal("same-instance Add reused published object")
	}
	if err := engine.DeleteTrigger("source", 0); err != nil {
		t.Fatal(err)
	}
	if len(old.GetTriggers()) != 1 {
		t.Fatal("successful deletion mutated old generation")
	}
	if err := engine.HandleManual("source", "action"); err == nil || commands != 0 {
		t.Fatal("deleted trigger still callable")
	}
	if err := engine.Add(reloadRecipe(true)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		engine.Initialize()
		if err := engine.HandleManual("source", "action"); err != nil {
			t.Fatal(err)
		}
	}
	if commands != 5 {
		t.Fatal("reload accumulated duplicate triggers")
	}
}

func TestSchedulerPreservesEnabledWindowOnReorder(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
	handler := automations.NewAutomationScheduler(automations.WithSchedulerClock(clock), automations.WithAutomationsFuncs())
	old := reloadRecipe(false)
	old.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}, {StartAt: "12:00:02", Type: "disable"}}
	if err := handler.Process(old); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = handler.Process(automations.NewDevice("source")) }()
	clock.Advance(time.Second)
	if !old.IsEnabled() {
		t.Fatal("initial window did not start")
	}
	next := reloadRecipe(false)
	next.SetEnabled(false)
	next.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:02", Type: "disable"}, {StartAt: "12:00:01", Type: "enable"}}
	if err := handler.Process(next); err != nil {
		t.Fatal(err)
	}
	if !next.IsEnabled() {
		t.Fatal("reordered same schedules reset enabled window")
	}
	clock.Advance(time.Second)
	if next.IsEnabled() || !old.IsEnabled() {
		t.Fatal("schedule still controls old generation")
	}
}

func TestEngineFirstSaveFailureDiscardsSchedule(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
	handler := automations.NewAutomationScheduler(automations.WithSchedulerClock(clock), automations.WithAutomationsFuncs())
	s := &failingRecipeStorage{Storage: reloadDisk(t), storeErr: errors.New("save failed")}
	engine := automations.NewEngine([]automations.AutomationHandler{handler}, &reloadRegistrar{}, &mocks.MockMqttClient{})
	engine.WithStorage(s)
	a := reloadRecipe(false)
	a.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}}
	if err := engine.Add(a); !errors.Is(err, s.storeErr) {
		t.Fatal("first save failure lost")
	}
	clock.Advance(48 * time.Hour)
	if a.IsEnabled() || handler.IsRunning(a) {
		t.Fatal("discarded first generation retained active jobs")
	}
	if _, err := engine.Load("source"); err == nil {
		t.Fatal("failed first save exposed runtime")
	}
	s.storeErr = nil
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal("first save failure retained handler ownership")
	}
}

func TestActionLookupFailureKeepsErrorCause(t *testing.T) {
	trigger, step, preset := automations.NewTriggerAction(), automations.NewStepAction(), automations.NewPresetCyclingAction()
	trigger.Id, step.Id, preset.Id = "target", "target", "target"
	failure := errors.New("target missing")
	r := &reloadRegistrar{lookup: func(string) error { return failure }}
	for _, action := range []automations.MqttAction{trigger, step, preset} {
		t.Run(string(action.GetType()), func(t *testing.T) {
			if err := action.Configure(r, &mocks.MockMqttClient{}); !errors.Is(err, failure) {
				t.Fatalf("contextual error lost cause: %v", err)
			}
		})
	}
}

func TestEngineDiskFailureDoesNotRemoveRuntime(t *testing.T) {
	disk := reloadDisk(t)
	engine := automations.NewEngine(nil, &reloadRegistrar{}, &mocks.MockMqttClient{})
	engine.WithStorage(disk)
	old := reloadRecipe(false)
	if err := engine.Add(old); err != nil {
		t.Fatal(err)
	}
	// A null/invalid recipe aborts the scan. Even an absent other file must not
	// become a runtime deletion until a complete successful scan establishes it.
	if err := disk.Store("broken", nil); err != nil {
		t.Fatal(err)
	}
	if err := disk.Delete("source"); err != nil {
		t.Fatal(err)
	}
	engine.Initialize()
	if got, _ := engine.Load("source"); got != old {
		t.Fatal("failed scan removed working automation")
	}
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatal(err)
	}
	if err := disk.Delete("broken"); err != nil {
		t.Fatal(err)
	}
	engine.Initialize()
	if _, err := engine.Load("source"); err == nil {
		t.Fatal("successful empty scan did not remove old generation")
	}
}

func TestEnginePublishCallbackCanReplaceGeneration(t *testing.T) {
	var engine *automations.AutomationEngine
	var commands []bool
	client := &delayedLifecycleClient{publish: func(_ string, p interface{}) {
		var data map[string]bool
		if err := json.Unmarshal(p.([]byte), &data); err != nil {
			t.Error(err)
		}
		commands = append(commands, data["presence"])
		if len(commands) == 1 {
			if err := engine.Add(reloadRecipe(true)); err != nil {
				t.Error(err)
			}
		}
	}}
	engine = automations.NewEngine(nil, &reloadRegistrar{}, client)
	engine.WithStorage(reloadDisk(t))
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := engine.HandleManual("source", "action"); err != nil {
			t.Error(err)
		}
	}()
	awaitReload(t, done)
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 2 || commands[0] || !commands[1] {
		t.Fatalf("bad publication re-entry binding: %v", commands)
	}
}

func TestEngineDeleteDoesNotJoinCommittedCommand(t *testing.T) {
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	engine := automations.NewEngine(nil, &reloadRegistrar{}, &delayedLifecycleClient{publish: func(string, interface{}) { close(entered); <-release }})
	engine.WithStorage(reloadDisk(t))
	if err := engine.Add(reloadRecipe(false)); err != nil {
		t.Fatal(err)
	}
	var once sync.Once
	defer once.Do(func() { close(release) })
	go func() {
		defer close(done)
		if err := engine.HandleManual("source", "action"); err != nil {
			t.Error(err)
		}
	}()
	awaitReload(t, entered)
	deleted := make(chan struct{})
	go func() {
		defer close(deleted)
		if err := engine.Delete("source"); err != nil {
			t.Error(err)
		}
	}()
	awaitReload(t, deleted)
	if err := engine.HandleManual("source", "action"); err == nil {
		t.Fatal("new execution allowed after deletion")
	}
	once.Do(func() { close(release) })
	awaitReload(t, done)
}

func TestEngineReplacementAllowsAcceptedDelayedAction(t *testing.T) {
	commands := make(chan bool, 2)
	client := &delayedLifecycleClient{publish: func(_ string, p interface{}) {
		var data map[string]bool
		if err := json.Unmarshal(p.([]byte), &data); err != nil {
			t.Error(err)
		}
		commands <- data["presence"]
	}}
	engine := automations.NewEngine(nil, &reloadRegistrar{}, client)
	engine.WithStorage(reloadDisk(t))
	old := reloadRecipe(false)
	delayed := old.Triggers[0].GetActions()[0].(*automations.MqttTriggerAction)
	delayed.Delay = utils.IntervalFromMilliseconds(30)
	defer delayed.Stop()
	if err := engine.Add(old); err != nil {
		t.Fatal(err)
	}
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatal(err)
	}
	if err := engine.Add(reloadRecipe(true)); err != nil {
		t.Fatal(err)
	}
	if err := engine.HandleManual("source", "action"); err != nil {
		t.Fatal(err)
	}
	seen := map[bool]int{}
	for i := 0; i < 2; i++ {
		select {
		case value := <-commands:
			seen[value]++
		case <-time.After(5 * time.Second):
			t.Fatal("accepted action lost during replacement")
		}
	}
	if seen[false] != 1 || seen[true] != 1 {
		t.Fatalf("old/new commands lost or duplicated: %v", seen)
	}
}

func TestAutomationEnabledConcurrentJSON(t *testing.T) {
	d := automations.NewDevice("device-id")
	d.FriendlyName = "sensor"
	d.Description = "description"
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			d.SetEnabled(i%2 == 0)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = d.IsEnabled()
			data, err := json.Marshal(d)
			if err != nil {
				t.Error(err)
				return
			}
			var fields map[string]interface{}
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Error(err)
				return
			}
			if len(fields) != 7 || fields["id"] != "device-id" || fields["friendlyname"] != "sensor" || fields["description"] != "description" || fields["type"] != "device" {
				t.Errorf("changed wire fields: %s", data)
				return
			}
			if _, ok := fields["enabled"].(bool); !ok {
				t.Errorf("enabled is not boolean: %s", data)
				return
			}
		}
	}()
	wg.Wait()
	for _, enabled := range []bool{true, false} {
		d.SetEnabled(enabled)
		data, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := automations.CreateFromPayload(data)
		if err != nil {
			t.Fatal(err)
		}
		if decoded.IsEnabled() != enabled {
			t.Fatalf("enabled roundtrip: want %v", enabled)
		}
	}
}

// func TestExportAutomationsFromFile(t *testing.T) {

// 	mqtt := &mocks.MockMqttClient{}
// 	store := utils_test.CreateStore()
// 	eventHub := &mocks.MockEventHub{}
// 	registrar := services.NewHubRegisterService(store, eventHub, 30000)

// 	dev1 := createMockDevice("0x56789", "livingroom", "brightness", nil, 0.0, 255.0)
// 	dev2 := createMockDevice("0x56789", "humansensor", "left_click", nil, 0.0, 255.0)

// 	bridgeinfos := createBridgeInfoes()
// 	registrar.RegisterBridge(bridgeinfos, 60)
// 	registrar.Register("livingroom", dev1)
// 	registrar.Register("humansensor", dev2)

// 	storage := mocks.NewMockAutomationStorage([]*automations.Device{})

// 	engine := automations.NewEngine([]automations.AutomationHandler{}, registrar, mqtt)
// 	engine.WithStorage(storage)
// 	engine.Initialize()

// 	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
// 	turnOffTrigger.Actions[0].Id = "0x56789"

// 	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)
// 	turnOnTrigger.Actions[0].Id = "0x56789"

// 	// create device trigger
// 	inputDeviceTriggers := []*automations.Device{}
// 	deviceTrigger1 := automations.NewDevice("0x123456")
// 	deviceTrigger1.FriendlyName = "humansensor"
// 	deviceTrigger1.Description = "test human sensor automation"
// 	deviceTrigger1.Triggers = append(deviceTrigger1.Triggers, turnOffTrigger)
// 	deviceTrigger1.Triggers = append(deviceTrigger1.Triggers, turnOnTrigger)
// 	inputDeviceTriggers = append(inputDeviceTriggers, deviceTrigger1)

// 	for _, d := range inputDeviceTriggers {
// 		err := engine.Add(d)
// 		if err != nil {
// 			t.Fatalf("ERROR adding trigger %s", err.Error())
// 		}
// 	}

// 	outputDeviceTriggers := engine.GetAllTriggers()
// 	if len(outputDeviceTriggers) != len(inputDeviceTriggers) {
// 		t.Fatalf("ERROR size mismatch. want %v got %d", len(inputDeviceTriggers), len(outputDeviceTriggers))
// 	}
// 	for dIdx, outDeviceTrigger := range outputDeviceTriggers {
// 		inputDeviceTrigger := inputDeviceTriggers[dIdx]
// 		if outDeviceTrigger.Id != inputDeviceTrigger.Id {
// 			t.Fatalf("ERROR Id mismatch")
// 		}
// 		if outDeviceTrigger.FriendlyName != inputDeviceTrigger.FriendlyName {
// 			t.Fatalf("ERROR Name mismatch")
// 		}
// 		if outDeviceTrigger.Description != inputDeviceTrigger.Description {
// 			t.Fatalf("ERROR Description mismatch")
// 		}
// 		if outDeviceTrigger.Enabled != inputDeviceTrigger.Enabled {
// 			t.Fatalf("ERROR Description mismatch")
// 		}

// 		for tidx, outputTrigger := range outDeviceTrigger.Triggers {
// 			inputTrigger := inputDeviceTrigger.Triggers[tidx]

// 			if outputTrigger.Name != inputTrigger.Name {
// 				t.Fatalf("ERROR Trigger.Name mismatch")
// 			}

// 			// check actions
// 			for aidx, outputAction := range outputTrigger.Actions {
// 				inputAction := inputTrigger.Actions[aidx]

// 				if outputAction.Id != inputAction.Id {
// 					t.Fatalf("ERROR Action.Id mismatch")
// 				}
// 				if outputAction.FriendlyName != inputAction.FriendlyName {
// 					t.Fatalf("ERROR Action.Friendlyname mismatch")
// 				}

// 				if outputAction.Property != inputAction.Property {
// 					t.Fatalf("ERROR Action.Property mismatch")
// 				}

// 				if outputAction.Delay != inputAction.Delay {
// 					t.Fatalf("ERROR Action.Delay mismatch")
// 				}
// 			}

// 			for cidx, outputCondition := range outputTrigger.Conditions {

// 				inputCondition := inputTrigger.Conditions[cidx]

// 				if outputCondition.EqualityOperator != inputCondition.EqualityOperator {
// 					t.Fatalf("ERROR Condition.EqualityOperator mismatch")
// 				}
// 				if outputCondition.Name != inputCondition.Name {
// 					t.Fatalf("ERROR Condition.Name mismatch")
// 				}

// 				if outputCondition.Value != inputCondition.Value {
// 					t.Fatalf("ERROR Condition.Value mismatch")
// 				}
// 			}
// 		}
// 	}

// 	for _, inputDeviceTrigger := range inputDeviceTriggers {
// 		err := engine.Delete(inputDeviceTrigger.Id)
// 		if err != nil {
// 			t.Fatalf("ERROR deleting file %v .Error %s", inputDeviceTrigger.Id, err.Error())
// 		}
// 	}

// 	inputDeviceTriggers = engine.GetAllTriggers()
// 	if len(inputDeviceTriggers) != 0 {
// 		t.Fatalf("ERROR triggers found. expecting empty")
// 	}
// }

func TestEngineAutomationUpdateShouldNotResetScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	wg.Add(1) // we expect only one event to be triggered

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation automations.Automation) error {
			automation.SetEnabled(true)
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation automations.Automation) error {
			automation.SetEnabled(false)
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(600 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation but keep the same schedule
	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is still running
	if !scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only one event to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()

	deviceAutomation.Schedules = []*automations.TimeSchedule{}
	if err := scheduleHandler.Process(deviceAutomation); err != nil {
		t.Fatalf("ERROR cleaning scheduler: %v", err)
	}
}

func TestEngineAutomationUpdateShouldResetAndTriggerAgainScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	// we expect only 2 event to be triggered
	// first event is on startup
	// second event is on automation schedule update
	wg.Add(2)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation automations.Automation) error {
			automation.SetEnabled(true)
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation automations.Automation) error {
			automation.SetEnabled(false)
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(600 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation and change schedule
	loc, _ = time.LoadLocation("Europe/London")
	now = time.Now().In(loc)
	start = now.Add(50 * time.Millisecond)
	deviceAutomation.Schedules[0] = utils_test.CreateTimeSchedule(start)

	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is still running
	if !scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only 2 events to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()

	deviceAutomation.Schedules = []*automations.TimeSchedule{}
	if err := scheduleHandler.Process(deviceAutomation); err != nil {
		t.Fatalf("ERROR cleaning scheduler: %v", err)
	}
}

func TestEngineAutomationUpdateShouldResetScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	// we expect only 1 event to be triggered
	// first event is on startup
	// after we update the automation schedule, the second event ashould not be trigger on time
	wg.Add(1)
	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation automations.Automation) error {
			automation.SetEnabled(true)
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation automations.Automation) error {
			automation.SetEnabled(false)
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(600 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation and change schedule
	loc, _ = time.LoadLocation("Europe/London")
	now = time.Now().In(loc)
	start = now.Add(1000 * time.Millisecond)
	deviceAutomation.Schedules[0] = utils_test.CreateTimeSchedule(start)

	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is still running
	if !scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only 1 event to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()

	deviceAutomation.Schedules = []*automations.TimeSchedule{}
	if err := scheduleHandler.Process(deviceAutomation); err != nil {
		t.Fatalf("ERROR cleaning scheduler: %v", err)
	}
}

func TestEngineAutomationUpdateShouldStopScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	// mqtt.OnMessageHandler(func(topic string, payload []byte) {
	// 	fmt.Printf("Received message on topic %s\n", topic)
	// })

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	// we expect only 1 event to be triggered
	// first event is on startup
	// after we update the automation schedule, the second event ashould not be trigger on time
	wg.Add(1)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation automations.Automation) error {
			automation.SetEnabled(true)
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation automations.Automation) error {
			automation.SetEnabled(false)
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(600 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.IsEnabled() {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation remove schedules
	deviceAutomation.Schedules = []*automations.TimeSchedule{}

	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is not running
	if scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only 1 event to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()

	deviceAutomation.Schedules = []*automations.TimeSchedule{}
	if err := scheduleHandler.Process(deviceAutomation); err != nil {
		t.Fatalf("ERROR cleaning scheduler: %v", err)
	}
}

func TestEngineSchedulerConfiguresAutomation(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})
	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1500 * time.Millisecond)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	wg.Add(1)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation automations.Automation) error {
			automation.SetEnabled(true)
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation automations.Automation) error {
			automation.SetEnabled(false)
			wg.Done()

			return nil
		}),
	)
	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	a, _ := engine.Load(deviceAutomation.Id)

	// make sure automation is disabled when we have scheduler enabled
	if a.IsEnabled() {
		t.Fatalf("ERROR automation is enabled (initial state)")
	}

	// trigger the automation
	doorSensorDevice.Exposes["contact"].Data.SetValue(true)
	engine.HandleDevice(doorSensorDevice, map[string]interface{}{"contact": true})
	time.Sleep(50 * time.Millisecond)

	// scheduler is enable so any event will enable automation
	wg.Wait()

	a, _ = engine.Load(deviceAutomation.Id)

	// scheduler should have enabled automation
	if !a.IsEnabled() {
		t.Fatalf("ERROR automation is not enabled")
	}

	// wait until end of schedule to disable automation
	wg.Add(1)
	wg.Wait()

	// trigger the automation
	doorSensorDevice.Exposes["contact"].Data.SetValue(false)
	engine.HandleDevice(doorSensorDevice, map[string]interface{}{"contact": false})
	time.Sleep(50 * time.Millisecond)

	a, _ = engine.Load(deviceAutomation.Id)

	// scheduler should have disabled automation
	if a.IsEnabled() {
		t.Fatalf("ERROR automation is enabled")
	}

	deviceAutomation.Schedules = []*automations.TimeSchedule{}
	if err := scheduleHandler.Process(deviceAutomation); err != nil {
		t.Fatalf("ERROR cleaning scheduler: %v", err)
	}
}
