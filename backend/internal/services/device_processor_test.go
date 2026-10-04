package services_test

import (
	"encoding/json"
	"fmt"
	"math"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type rapidActionClient struct {
	mocks.MockMqttClient
	publish func(string, interface{})
}

func TestDeviceProcessorCreationOwnership(t *testing.T) {
	for _, registered := range []bool{false, true} {
		for _, disable := range []string{"", "defaults", "override"} {
			t.Run(fmt.Sprintf("registered=%t/disable=%s", registered, disable), func(t *testing.T) {
				store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
				registrar := services.NewHubRegisterService(store, &mocks.MockEventHub{}, 30000)
				if registered {
					registrar.RegisterBridge(utils_test.CreateBridgeInfoList([]*devices.Device{utils_test.CreateDevice("creation", "creation", "brightness", float64(0), 0, 255)}))
				}
				cfg := settings.DefaultDeviceConfig()
				cfg.DefaultDebounceByCategory = nil
				if err := store.AppConfig().SetDeviceConfigDefaults(cfg); err != nil {
					t.Fatal(err)
				}
				entered, release := make(chan *devices.Device, 1), make(chan struct{})
				var seeds, updates atomic.Int32
				var delivered []interface{}
				var processor *services.DeviceProcessor
				events := &devices.DeviceRequestEvents{
					AvailabilityTimeout: time.Hour,
					OnNewDevice: func(d *devices.Device) {
						if seeds.Add(1) != 1 {
							return
						}
						// Same-device update re-entry must enqueue, not wait for Seed.
						if err := processor.CreateOrUpdateDevice("creation", "mqtt", map[string]interface{}{"brightness": float64(11)}); err != nil {
							t.Error(err)
						}
						entered <- d
						<-release
					},
					OnDeviceUpdated: func(d *devices.Device, _ *devices.UpdatePackage) {
						updates.Add(1)
						entity, _ := d.GetExpose("brightness")
						value := entity.Data.Value()
						delivered = append(delivered, value)
						if disable == "" && value == float64(11) {
							// Re-entry while draining must not block or skip queued input.
							processor.OnDeviceConfigUpdated(settings.NewDeviceConfig(d.Id))
							if err := processor.CreateOrUpdateDevice("creation", "mqtt", map[string]interface{}{"brightness": float64(12)}); err != nil {
								t.Error(err)
							}
						}
					},
					OnDeviceAvailabilityChanged: func(*devices.UpdatePackage) {},
					OnDeviceMeasurementsUpdated: func(*devices.Device, map[string]interface{}) {},
				}
				processor = services.NewDeviceProcessorBuilder().WithRegistrar(registrar).WithStore(store).WithEvents(events).WithAutomationQuerier(mocks.NewMockAutomationDeviceQuerier()).Build()
				store.AppConfig().RegisterDeviceConfigUpdateListener(processor.OnDeviceConfigUpdated)
				defer func() { c := settings.DefaultDeviceConfig(); c.Disabled = true; processor.OnDeviceConfigUpdated(c) }()
				var once sync.Once
				defer once.Do(func() { close(release) })
				done := make(chan error, 1)
				go func() {
					done <- processor.CreateOrUpdateDevice("creation", "mqtt", map[string]interface{}{"brightness": float64(10)})
				}()
				var device *devices.Device
				select {
				case device = <-entered:
				case <-time.After(5 * time.Second):
					t.Fatal("Seed re-entry blocked")
				}
				if disable == "defaults" {
					c := settings.DefaultDeviceConfig()
					c.Disabled = true
					c.DefaultDebounceByCategory = nil
					if err := store.AppConfig().SetDeviceConfigDefaults(c); err != nil {
						t.Fatal(err)
					}
				}
				if disable == "override" {
					c := settings.NewDeviceConfig(device.Id)
					c.Disabled = true
					processor.OnDeviceConfigUpdated(c)
				}
				payload := map[string]interface{}{"brightness": float64(20)}
				if err := processor.CreateOrUpdateDevice("creation", "mqtt", payload); err != nil {
					t.Fatal(err)
				}
				// Queued input must not borrow the caller's mutable map.
				payload["brightness"] = float64(99)
				once.Do(func() { close(release) })
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("creation did not finish")
				}
				if seeds.Load() != 1 {
					t.Fatalf("got %d Seed callbacks, want one", seeds.Load())
				}
				want := float64(12)
				if disable != "" {
					want = 11
				}
				entity, _ := device.GetExpose("brightness")
				if got := entity.Data.Value(); got != want {
					t.Fatalf("brightness = %v, want %v", got, want)
				}
				wantUpdates := int32(3)
				if disable != "" {
					wantUpdates = 1
				}
				if updates.Load() != wantUpdates {
					t.Fatalf("updates = %d, want %d", updates.Load(), wantUpdates)
				}
				if disable == "" && (delivered[0] != float64(11) || delivered[1] != float64(20) || delivered[2] != float64(12)) {
					t.Fatalf("queued delivery order/copy = %v, want [11 20 12]", delivered)
				}
			})
		}
	}
}

func (c *rapidActionClient) Publish(topic string, payload interface{}) { c.publish(topic, payload) }

func TestDeviceProcessorRapidPhysicalActions(t *testing.T) {
	for _, tc := range []struct {
		name            string
		rotary, confirm bool
	}{
		{"identical button presses", false, true},
		{"repeated and alternating rotations with confirmed feedback", true, true},
		{"rotations with feedback withheld", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := devices.NewDevice("dial-id")
			source.FriendlyName = "dial"
			actionEntity := devices.NewEntity("action")
			actionEntity.Type = bridge.EnumDataType
			actionEntity.Data.SetValue("idle")
			source.Exposes["action"] = actionEntity
			timeEntity := devices.NewEntity("action_time")
			timeEntity.Type = bridge.NumericDataType
			timeEntity.Data.SetValue(float64(0))
			source.Exposes["action_time"] = timeEntity
			light := utils_test.CreateDevice("light-id", "light", "brightness", float64(200), 0, 255)
			store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
			for _, d := range []*devices.Device{source, light} {
				cfg := settings.NewDeviceConfig(d.Id)
				cfg.DebounceOverrides["brightness"] = utils.IntervalFromHours(1)
				cfg.DebounceOverrides["action_time"] = utils.IntervalFromHours(1)
				if err := store.AppConfig().SetDeviceConfigOverrides(cfg); err != nil {
					t.Fatal(err)
				}
			}
			registrar := services.NewHubRegisterService(store, &mocks.MockEventHub{}, 30000)
			registrar.RegisterBridge(utils_test.CreateBridgeInfoList([]*devices.Device{source, light}))
			var err error
			source, err = registrar.LookupById(source.Id)
			if err != nil {
				t.Fatal(err)
			}
			light, err = registrar.LookupById(light.Id)
			if err != nil {
				t.Fatal(err)
			}
			var processor *services.DeviceProcessor
			var commands []float64
			client := &rapidActionClient{publish: func(topic string, payload interface{}) {
				if topic != "light/set" {
					t.Fatalf("unexpected topic %s", topic)
				}
				var data map[string]interface{}
				if err := json.Unmarshal(payload.([]byte), &data); err != nil {
					t.Fatal(err)
				}
				value, ok := data["brightness"].(float64)
				if !ok || len(data) != 1 {
					t.Fatalf("unexpected command %s", payload)
				}
				commands = append(commands, value)
				// Explicit simulated device confirmation, not automatic optimistic state.
				if tc.confirm {
					if err := processor.CreateOrUpdateDevice("light", "mqtt", data); err != nil {
						t.Fatal(err)
					}
				}
			}}
			engine := automations.NewEngine(nil, registrar, client)
			engine.WithStorage(mocks.NewMockAutomationStorage[*automations.Device](nil))
			triggers, eventUpdates, numericUpdates := 0, 0, 0
			var expectedAction string
			var expectedTime float64
			events := &devices.DeviceRequestEvents{
				AvailabilityTimeout:         time.Hour,
				OnNewDevice:                 func(*devices.Device) {},
				OnDeviceAvailabilityChanged: func(*devices.UpdatePackage) { t.Error("unexpected expiry") },
				OnDeviceMeasurementsUpdated: func(*devices.Device, map[string]interface{}) {},
				OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
					if _, ok := p.Data["action"]; ok && d.Id == source.Id {
						eventUpdates++
					}
					if _, ok := p.Data["action_time"]; ok {
						numericUpdates++
					}
					if _, ok := p.Data["brightness"]; ok {
						numericUpdates++
					}
				},
				OnDeviceAutomationTriggered: func(d *devices.Device, delta map[string]interface{}) {
					triggers++
					if delta["action"] != expectedAction || d.Exposes["action"].Data.Value() != expectedAction || d.Exposes["action_time"].Data.Value() != expectedTime {
						t.Fatal("automation did not receive current physical event/state")
					}
					engine.HandleDevice(d, delta)
				},
			}
			processor = services.NewDeviceProcessorBuilder().WithRegistrar(registrar).WithStore(store).
				WithEvents(events).WithAutomationQuerier(engine).Build()
			t.Cleanup(func() {
				cfg := settings.DefaultDeviceConfig()
				cfg.Disabled = true
				processor.OnDeviceConfigUpdated(cfg)
			})
			// Prime service/debounce ownership before enabling automations.
			if err := processor.CreateOrUpdateDevice("dial", "mqtt", map[string]interface{}{"action": "idle", "action_time": float64(0)}); err != nil {
				t.Fatal(err)
			}
			if err := processor.CreateOrUpdateDevice("light", "mqtt", map[string]interface{}{"brightness": float64(200)}); err != nil {
				t.Fatal(err)
			}
			automation := automations.NewDevice(source.Id)
			automation.SetEnabled(true)
			labels := []string{"single"}
			if tc.rotary {
				labels = []string{"dial_rotate_right_slow", "dial_rotate_left_slow"}
			}
			for i, label := range labels {
				trigger := automations.NewDeviceTrigger("action")
				trigger.Conditions = []automations.Condition{utils_test.NewExposeCondition("action", label, utils.Equals)}
				if tc.rotary {
					operator := "+"
					if i == 1 {
						operator = "-"
					}
					action := automations.NewStepAction()
					action.Id, action.Property, action.Data = light.Id, "brightness", float64(0.5)
					action.Steps = []*automations.Step{{Id: light.Id, Property: "brightness", Operator: operator}, {Id: source.Id, Property: "action_time", Operator: "*"}}
					trigger.Actions = []automations.MqttAction{action}
				} else {
					action := automations.NewTriggerAction()
					action.Id = light.Id
					action.Exposes = []*automations.MqttTriggerActionExpose{{Name: "brightness", Data: float64(42)}}
					trigger.Actions = []automations.MqttAction{action}
				}
				automation.Triggers = append(automation.Triggers, trigger)
			}
			if err := engine.Add(automation); err != nil {
				t.Fatal(err)
			}
			previous := float64(200)
			for i := 0; i < 120; i++ {
				expectedAction, expectedTime = "single", float64(2)
				want := float64(42)
				if tc.rotary {
					pattern := []int{0, 0, 1, 0, 1, 1}
					direction := pattern[i%len(pattern)]
					expectedAction, expectedTime = labels[direction], float64(2*(i%6+1))
					base := float64(200)
					if tc.confirm {
						base = previous
					}
					want = math.Min(255, base+expectedTime*0.5)
					if direction == 1 {
						want = math.Max(0, base-expectedTime*0.5)
					}
				}
				// Match decoded MQTT payload types; broker transport itself is out of scope.
				raw := fmt.Sprintf(`{"action":%q,"action_time":%v}`, expectedAction, expectedTime)
				var payload map[string]interface{}
				if err := json.Unmarshal([]byte(raw), &payload); err != nil {
					t.Fatal(err)
				}
				if err := processor.CreateOrUpdateDevice("dial", "mqtt", payload); err != nil {
					t.Fatal(err)
				}
				if len(commands) != i+1 || commands[i] != want {
					t.Fatalf("event %d: commands %v, want new brightness %v", i, commands, want)
				}
				previous = want
			}
			if triggers != 120 || eventUpdates != 120 || numericUpdates != 0 {
				t.Fatalf("automation/event/numeric updates = %d/%d/%d, want 120/120/0", triggers, eventUpdates, numericUpdates)
			}
			wantReported := float64(200)
			if tc.confirm {
				wantReported = previous
			}
			if got := light.Exposes["brightness"].Data.Value(); got != wantReported {
				t.Fatalf("reported brightness = %v, want %v", got, wantReported)
			}
		})
	}
}

func TestDeviceProcessorConcurrentRegistryConfiguration(t *testing.T) {
	store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
	defaults := settings.DefaultDeviceConfig()
	defaults.Disabled = true
	if err := store.AppConfig().SetDeviceConfigDefaults(defaults); err != nil {
		t.Fatal(err)
	}
	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(services.NewHubRegisterService(store, &mocks.MockEventHub{}, 30000)).
		WithStore(store).WithEvents(&devices.DeviceRequestEvents{}).
		WithAutomationQuerier(mocks.NewMockAutomationDeviceQuerier()).Build()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			for i := 0; i < 100; i++ {
				name := fmt.Sprintf("registry-%d-%d", worker, i)
				if worker == 0 {
					if err := processor.CreateOrUpdateDevice(name, "wifi", map[string]interface{}{"state": "ON"}); err != nil {
						t.Errorf("create: %v", err)
					}
				} else {
					processor.OnDeviceConfigUpdated(defaults)
					override := settings.NewDeviceConfig(store.ResolveFriendlyName(fmt.Sprintf("registry-0-%d", i)))
					override.Disabled = true
					processor.OnDeviceConfigUpdated(override)
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()
}

func TestDeviceProcessorSeedCallbackCanConfigureRegistry(t *testing.T) {
	store := utils_test.CreateStoreFromDeviceRepo(repository.NewMemoryDeviceRepo())
	var processor *services.DeviceProcessor
	callbackDone := make(chan struct{})
	creationDone := make(chan error, 1)
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: time.Hour,
		OnNewDevice: func(d *devices.Device) {
			processor.OnDeviceConfigUpdated(settings.DefaultDeviceConfig())
			processor.OnDeviceConfigUpdated(settings.NewDeviceConfig(d.Id))
			close(callbackDone)
		},
		OnDeviceAvailabilityChanged: func(*devices.UpdatePackage) {},
	}
	processor = services.NewDeviceProcessorBuilder().
		WithRegistrar(services.NewHubRegisterService(store, &mocks.MockEventHub{}, 30000)).
		WithStore(store).WithEvents(events).
		WithAutomationQuerier(mocks.NewMockAutomationDeviceQuerier()).Build()
	go func() {
		creationDone <- processor.CreateOrUpdateDevice("callback-reentry", "wifi", map[string]interface{}{"state": "ON"})
	}()
	select {
	case err := <-creationDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Seed callback blocked on registry configuration")
	}
	select {
	case <-callbackDone:
	default:
		t.Fatal("Seed callback was not invoked")
	}
	disabled := settings.DefaultDeviceConfig()
	disabled.Disabled = true
	processor.OnDeviceConfigUpdated(disabled)
}

func TestDeviceProcessor_CreateOrUpdateDevice_NewDevice(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error: OnDeviceUpdated called for new device")
		},
		OnNewDevice: func(d *devices.Device) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device add: %s", err)
			}
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		AvailabilityTimeout: 1,
	}
	deviceName := "Attic room Light"

	lastSeen := time.Now().Format(time.RFC3339)
	payload := map[string]interface{}{}
	payload["brightness"] = 120.1
	payload["color_temp"] = 100.1
	payload["state"] = "on"
	payload["last_seen"] = lastSeen
	payload["battery"] = 100

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	err := processor.CreateOrUpdateDevice(deviceName, "wifi", payload)

	if err != nil {
		t.Errorf("Error creating new device: %s", err)
	}
	wg.Wait()
	d, err := store.FindDeviceByFriendlyName(deviceName)
	if err != nil {
		t.Errorf("Error finding device: %s", err)
	}

	if d.FriendlyName != deviceName {
		t.Errorf("Device FriendlyName mismatch want: %s got: %s", deviceName, d.FriendlyName)
	}
	if d.LastSeen != lastSeen {
		t.Errorf("Device LastSeen mismatch want: %s got: %s", lastSeen, d.LastSeen)
	}
	if d.ConnectionType != "wifi" {
		t.Errorf("Device ConnectionType mismatch want: %s got: %s", "wifi", d.ConnectionType)
	}
	if d.PowerSource != "battery" {
		t.Errorf("Device PowerSource mismatch want: %s got: %s", "battery", d.PowerSource)
	}
	if d.Exposes["brightness"].Data.Value() != 120.1 {
		t.Errorf("Device Expose brightness mismatch want: %f got: %f", 120.1, d.Exposes["brightness"].Data.Value())
	}
	if d.Exposes["color_temp"].Data.Value() != 100.1 {
		t.Errorf("Device Expose color_temp mismatch want: %f got: %f", 100.1, d.Exposes["color_temp"].Data.Value())
	}
	if d.Exposes["state"].Data.Value() != "on" {
		t.Errorf("Device Expose state mismatch want: %s got: %s", "on", d.Exposes["state"].Data.Value())
	}
}

func TestDeviceProcessor_CreateOrUpdateDevice_ExistingDevice(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device update: %s", err)
			}
			wg.Done()
		},
		OnNewDevice: func(d *devices.Device) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device add: %s", err)
			}
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
		AvailabilityTimeout: 1,
	}
	deviceName := "Attic room Light"

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	lastSeen := time.Now().Format(time.RFC3339)
	updatePayload := map[string]interface{}{}
	updatePayload["brightness"] = 10.1
	updatePayload["color_temp"] = 120.1
	updatePayload["state"] = "false"
	updatePayload["last_seen"] = lastSeen
	updatePayload["battery"] = 100
	err = processor.CreateOrUpdateDevice(deviceName, "wifi", updatePayload)

	wg.Wait()
	if err != nil {
		t.Errorf("Error updating device: %s", err)
	}
	d, err := store.FindDeviceByFriendlyName(deviceName)
	if err != nil {
		t.Errorf("Error finding device: %s", err)
	}

	if d.FriendlyName != deviceName {
		t.Errorf("Device FriendlyName mismatch want: %s got: %s", deviceName, d.FriendlyName)
	}
	if d.LastSeen != lastSeen {
		t.Errorf("Device LastSeen mismatch want: %s got: %s", lastSeen, d.LastSeen)
	}
	if d.ConnectionType != "mqtt" {
		t.Errorf("Device ConnectionType mismatch want: %s got: %s", "mqtt", d.ConnectionType)
	}
	if d.PowerSource != "mains (single phase)" {
		t.Errorf("Device PowerSource mismatch want: %s got: %s", "mains (single phase)", d.PowerSource)
	}
	if d.Exposes["brightness"].Data.Value() != 10.1 {
		t.Errorf("Device Expose brightness mismatch want: %f got: %f", 10.1, d.Exposes["brightness"].Data.Value())
	}
	if d.Exposes["color_temp"].Data.Value() != 120.1 {
		t.Errorf("Device Expose color_temp mismatch want: %f got: %f", 120.1, d.Exposes["color_temp"].Data.Value())
	}
	if d.Exposes["state"].Data.Value() != "false" {
		t.Errorf("Device Expose state mismatch want: %s got: %s", "false", d.Exposes["state"].Data.Value())
	}
}

func TestOnDeviceConfigUpdated_WithDeviceOverride_ShouldDisableDevice(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(2)

	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error	should 	not call OnDeviceUpdated for disabled device")
		},
		OnNewDevice: func(d *devices.Device) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device add: %s", err)
			}
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
		AvailabilityTimeout: 1,
	}

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	//  Send payload 1
	// "friendly_name": "Attic room Light",
	// "ieee_address": "0x70ac08fffefafeca",
	deviceName := "Attic room Light"
	lastSeen := time.Now().Format(time.RFC3339)
	updatePayload := map[string]interface{}{}
	updatePayload["brightness"] = 10.1
	updatePayload["color_temp"] = 120.1
	updatePayload["state"] = "false"
	updatePayload["last_seen"] = lastSeen
	updatePayload["battery"] = 100
	processor.CreateOrUpdateDevice(deviceName, "wifi", updatePayload)

	//  Send payload 2
	// "friendly_name": "Living room presence sensor",
	// "ieee_address": "0xa4c13894070052fc",
	deviceName2 := "Living room presence sensor"
	lastSeen2 := time.Now().Format(time.RFC3339)
	updatePayload2 := map[string]interface{}{}
	updatePayload2["presence"] = true
	updatePayload2["target_distance"] = 102.1
	updatePayload2["last_seen"] = lastSeen2
	processor.CreateOrUpdateDevice(deviceName2, "mqtt", updatePayload2)

	cfg := settings.NewDeviceConfig("0xa4c13894070052fc")
	cfg.Disabled = true

	processor.OnDeviceConfigUpdated(cfg)

	time.Sleep(200 * time.Millisecond)
}

func TestOnDeviceConfigUpdated_WithDeviceDefaults_ShouldDisableAllDevices(t *testing.T) {

	//wg := sync.WaitGroup{}

	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatal("Error creating file store:", err)
		return
	}
	defer cleanup()

	eventHub := &mocks.MockEventHub{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	newDeviceIndex := 0
	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error	should 	not call OnDeviceUpdated for disabled device")
		},
		OnNewDevice: func(d *devices.Device) {
			newDeviceIndex++
			if newDeviceIndex > 2 {
				t.Errorf("Error	should 	not call OnNewDevice for disabled device")
			}
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
		AvailabilityTimeout: 1,
	}

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	// send two new devices before setting defaults

	//  Send payload 1
	// "friendly_name": "Attic room Light",
	// "ieee_address": "0x70ac08fffefafeca",
	deviceName := "Attic room Light"
	lastSeen := time.Now().Format(time.RFC3339)
	updatePayload := map[string]interface{}{}
	updatePayload["brightness"] = 10.1
	updatePayload["color_temp"] = 120.1
	updatePayload["state"] = "false"
	updatePayload["last_seen"] = lastSeen
	updatePayload["battery"] = 100
	processor.CreateOrUpdateDevice(deviceName, "wifi", updatePayload)

	//  Send payload 2
	// "friendly_name": "Living room presence sensor",
	// "ieee_address": "0xa4c13894070052fc",
	deviceName2 := "Living room presence sensor"
	lastSeen2 := time.Now().Format(time.RFC3339)
	updatePayload2 := map[string]interface{}{}
	updatePayload2["presence"] = true
	updatePayload2["target_distance"] = 102.1
	updatePayload2["last_seen"] = lastSeen2
	processor.CreateOrUpdateDevice(deviceName2, "mqtt", updatePayload2)

	cfg := store.AppConfig()
	cfg.RegisterDeviceConfigUpdateListener(func(cfg *settings.DeviceConfig) {
		processor.OnDeviceConfigUpdated(cfg)
	})
	// create defaults and set devices disabled
	defaults := settings.DefaultDeviceConfig()
	defaults.Disabled = true
	cfg.SetDeviceConfigDefaults(defaults)

	// send again, this update should be ignored
	lastSeen2 = time.Now().Format(time.RFC3339)
	updatePayload2["presence"] = false
	updatePayload2["target_distance"] = 12.1
	updatePayload2["last_seen"] = lastSeen2
	processor.CreateOrUpdateDevice(deviceName2, "mqtt", updatePayload2)

	// send new device and it should be ignored

	//  Send payload 3
	// "friendly_name": "Attic alarm",
	// "ieee_address": "0xa4c1389b273366c3",
	deviceName3 := "Attic alarm"
	lastSeen3 := time.Now().Format(time.RFC3339)
	updatePayload3 := map[string]interface{}{}
	updatePayload3["alarm"] = true
	updatePayload3["last_seen"] = lastSeen3
	processor.CreateOrUpdateDevice(deviceName3, "mqtt", updatePayload3)
}
