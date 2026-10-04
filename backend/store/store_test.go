package store_test

import (
	"context"
	"fmt"
	metrics "node-herder/internal/metrics/domain"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type cleanupLifecycleRepo struct {
	mocks.NopMetricsRepo
	prune func(time.Duration) error
}

func (r *cleanupLifecycleRepo) Prune(expiry time.Duration) error { return r.prune(expiry) }

func stopCleanupTask(t *testing.T, task settings.Task) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- task.Stop() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup Stop did not finish")
	}
}

func TestMetricsCleanupLifecycleIdleRestartAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	repo := &cleanupLifecycleRepo{prune: func(time.Duration) error { calls.Add(1); return nil }}
	task := store.NewMetricsCleanupTask(ctx, repo)
	config := settings.NewAppConfig()
	config.Hub.History.SleepTimeout = utils.IntervalFromHours(12)
	stopCleanupTask(t, task) // before Start
	t.Cleanup(func() { stopCleanupTask(t, task) })
	for i := 0; i < 10; i++ {
		if err := task.Start(config); err != nil {
			t.Fatal(err)
		}
	}
	stopCleanupTask(t, task) // must interrupt the twelve-hour wait
	stopCleanupTask(t, task)
	if calls.Load() != 0 {
		t.Fatal("cleanup pruned before its sleep elapsed")
	}
	if err := task.Start(config); err != nil {
		t.Fatal(err)
	}
	cancel()
	stopCleanupTask(t, task)
	if err := task.Start(config); err != context.Canceled {
		t.Fatalf("Start after parent cancellation: %v", err)
	}
}

func TestMetricsCleanupLifecycleConcurrentStartStop(t *testing.T) {
	task := store.NewMetricsCleanupTask(context.Background(), &cleanupLifecycleRepo{
		prune: func(time.Duration) error { t.Error("unexpected prune during long wait"); return nil },
	})
	t.Cleanup(func() { stopCleanupTask(t, task) })
	config := settings.NewAppConfig()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if err := task.Start(config); err != nil {
					t.Error(err)
				}
				if err := task.Stop(); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	stopCleanupTask(t, task)
}

func TestMetricsCleanupLifecycleStopJoinsPrune(t *testing.T) {
	entered := make(chan time.Duration, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	var calls atomic.Int32
	repo := &cleanupLifecycleRepo{prune: func(expiry time.Duration) error {
		calls.Add(1)
		entered <- expiry
		<-release
		return nil
	}}
	task := store.NewMetricsCleanupTask(context.Background(), repo)
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); stopCleanupTask(t, task) })
	config := settings.NewAppConfig()
	config.Hub.History = settings.NewHistoryConfig(utils.IntervalFromMilliseconds(1), utils.IntervalFromHours(3))
	if err := task.Start(config); err != nil {
		t.Fatal(err)
	}
	select {
	case expiry := <-entered:
		if expiry != 3*time.Hour {
			t.Fatalf("unexpected retention: %v", expiry)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("prune did not start")
	}
	stopped := make(chan error, 1)
	go func() { stopped <- task.Stop() }()
	select {
	case <-stopped:
		t.Fatal("Stop returned while Prune was still running")
	case <-time.After(20 * time.Millisecond): // bounded negative assertion, prune is gated
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not join released Prune")
	}
	if calls.Load() != 1 {
		t.Fatalf("expected exactly one prune, got %d", calls.Load())
	}
}

func TestMetricsCleanupOwnsConfigSnapshot(t *testing.T) {
	for _, factory := range []struct {
		name   string
		create func(context.Context, metrics.Repository) settings.Task
	}{
		{"constructor", store.NewMetricsCleanupTask}, {"default", store.DefaultMetricsCleanupTask},
	} {
		t.Run(factory.name, func(t *testing.T) {
			entered, release := make(chan time.Duration, 1), make(chan struct{})
			var first, released sync.Once
			task := factory.create(context.Background(), &cleanupLifecycleRepo{prune: func(expiry time.Duration) error {
				first.Do(func() { entered <- expiry; <-release })
				return nil
			}})
			t.Cleanup(func() { released.Do(func() { close(release) }); stopCleanupTask(t, task) })
			config := settings.NewAppConfig()
			config.Hub.History = settings.NewHistoryConfig(utils.IntervalFromMilliseconds(1), utils.IntervalFromHours(3))
			if err := task.Start(config); err != nil {
				t.Fatal(err)
			}
			// Start must capture values, not retain mutable configuration pointers.
			config.Hub.History.SleepTimeout.Value = 12
			config.Hub.History.SleepTimeout.Unit = utils.UnitHours
			config.Hub.History.ExpireAt.Value = 9
			select {
			case expiry := <-entered:
				if expiry != 3*time.Hour {
					t.Fatalf("retention changed after Start: %v", expiry)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("worker used modified sleep instead of its snapshot")
			}
		})
	}
}

func TestStoreLoadAllDevices(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	sort.Slice(wantDevices, func(i, j int) bool {
		return wantDevices[i].FriendlyName < wantDevices[j].FriendlyName
	})
	store := utils_test.CreateStore()

	// NOTE:
	// need to add bridgeinfo so the new devices can be registered withthe mapper
	// else if not found in bridge it will use the friendlyname to has the id for mapping
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	// store devices in store
	for _, wd := range wantDevices {
		err := store.StoreDevice(wd.FriendlyName, wd)
		if err != nil {
			t.Fatalf("error storing device %v, %v", wd.Id, err.Error())
		}
	}

	gotDevices, err := store.AllDevices()
	if err != nil {
		t.Fatalf("error loading devices %v:", err.Error())
	}

	if len(gotDevices) != len(wantDevices) {
		t.Fatalf("wrong number of devices. want %v got %v ", len(wantDevices), len(gotDevices))
	}

	for id, wd := range wantDevices {
		gt := gotDevices[id]
		utils_test.ValidateDevice(t, wd, gt)
	}
}

func TestStoreSavesLoggerConfig(t *testing.T) {
	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})

	appConfig := settings.NewAppConfig()
	appConfig.Hub.History = settings.DefaultHistoryConfig()
	appStore, cleanup, err := utils_test.CreateFileStoreWithAppConfig(appConfig, mockClock)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := appStore.AppConfig()
	l, err := cfg.LoadLoggerConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig failed. err %v ", err)
	}

	if l.EnableRemoteLogger != false {
		t.Fatalf("Logger config EnableRemoteLogger is set")
	}

	mockLoggerConfig := settings.NewLoggerConfig(true, "debug")
	cfg.SaveLoggerConfig(mockLoggerConfig)

	l, err = cfg.LoadLoggerConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig failed. err %v ", err)
	}
	if l.EnableRemoteLogger != true {
		t.Fatalf("Logger config EnableRemoteLogger is not set")
	}
}

func TestStoreMetricsCleanupTasks(t *testing.T) {

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})

	appConfig := settings.NewAppConfig()
	appConfig.Hub.History = settings.DefaultHistoryConfig()
	appStore, cleanup, err := utils_test.CreateFileStoreWithAppConfig(appConfig, mockClock)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)

	// store bridgeInfoList
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err = appStore.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	cfg := appStore.AppConfig()
	//Enable metrics for all devices
	for _, wd := range wantDevices {
		deviceConfig := settings.NewDeviceConfig(wd.Id)
		deviceConfig.MetricsEnabled = true
		cfg.SetDeviceConfigOverrides(deviceConfig)
		if err != nil {
			t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
		}
	}

	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	// configure cleanup to stay idle while we populate and verify the metrics.
	initialSleepTimeout := utils.IntervalFromMinutes(5)
	expireAt := utils.IntervalFromHours(1)

	if _, err := cfg.SaveHistoryConfig(settings.NewHistoryConfig(initialSleepTimeout, expireAt)); err != nil {
		t.Fatalf("failed to save history config: %v", err)
	}

	timestamps := utils_test.CreateDateTimeTimestamps(3, 24, 1)

	numOfEvents := 3 * 24
	// trigger multiple events for each device
	for i, timestamp := range timestamps {
		for id, wd := range wantDevices {
			for _, we := range wd.Exposes {
				we.Data.SetValue(float32((i + 1) + id*2))
			}

			payload := utils_test.Payload(wd)
			mockClock.SetMockTime(timestamp)

			err := appStore.StoreMetrics(wd.FriendlyName, payload)
			time.Sleep(time.Millisecond * 5)

			if err != nil {
				t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
			}
		}
	}

	// reset clock its used in pruning

	mockClock.SetMockTime(time.Now().UTC())

	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month(), now.Day()-5, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	// assert that metrics are stored
	for _, wd := range wantDevices {

		result, err := appStore.ViewMetrics(wd, from, to)
		time.Sleep(time.Second * 1)
		if err != nil {
			t.Fatalf("error retreiving metrics device %v error: %v:", wd.FriendlyName, err.Error())
		}
		for _, expose := range result.Exposes {

			event := metrics.ToNumericExposeResults(expose)
			if len(event.Data) != numOfEvents {
				t.Fatalf("error metrics results mismatch for %s. want %v got %v", event.Name, numOfEvents, len(event.Data))

			}
		}
	}

	// Speed up cleanup and give it enough time to prune old entries.
	cleanupSleepTimeout := utils.IntervalFromSeconds(1)
	if _, err := cfg.SaveHistoryConfig(settings.NewHistoryConfig(cleanupSleepTimeout, expireAt)); err != nil {
		t.Fatalf("failed to save cleanup history config: %v", err)
	}
	time.Sleep(cleanupSleepTimeout.Duration() + time.Second)

	// assert that metrics are removed
	for _, wd := range wantDevices {

		result, err := appStore.ViewMetrics(wd, from, to)
		if err != nil {
			t.Fatalf("error retreiving metrics device %v error: %v:", wd.FriendlyName, err.Error())
		}

		for _, expose := range result.Exposes {

			event := metrics.ToNumericExposeResults(expose)
			for _, event := range event.Data {
				timestamp := time.UnixMilli(event.X).UTC()

				// assert for any events that are older than 1 hour
				if now.Sub(timestamp) > expireAt.Duration() {

					t.Errorf("failed to prune event timestamp %v", timestamp)
				}
			}
		}
	}

}

func TestStoreDeviceStoreDoesNotStoreMetricsIfDisabled(t *testing.T) {
	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)

	//create device repo
	deviceRepo := repository.NewMemoryDeviceRepo()

	//create settings repo
	settingsTempFile := utils_test.Tempfile()
	defer os.Remove(settingsTempFile)
	settingsRepo, err := repository.NewFileSettingsRepoFromFile(settingsTempFile)
	if err != nil {
		t.Fatalf("settings repo failed. error: %v ", err.Error())
	}

	// get first device and enable metrics
	dev1 := wantDevices[0]
	appConfig := settings.NewAppConfig()
	deviceConfig := settings.NewDeviceConfig(dev1.Id)
	deviceConfig.MetricsEnabled = false
	appConfig.AddDeviceConfig(deviceConfig)
	settingsRepo.SaveAppConfig(appConfig)

	// create metrics repo
	var metricsStoreHandler = func(id string, data map[string]any) {
		t.Fatalf("unexpected device %v invoked for metrics", id)
	}

	metricsRepo := &mocks.NopMetricsRepo{}
	metricsRepo.WithStoreHandler(metricsStoreHandler)

	//create store
	store := utils_test.CreateStoreFromRepos(deviceRepo, metricsRepo, settingsRepo)

	// store bridgeInfoList
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err = store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	// update all devices but assert that only metrics enabled device stores metrics

	// trigger multiple events for each device
	for i := 0; i < 10; i++ {
		for id, wd := range wantDevices {
			for _, we := range wd.Exposes {
				we.Data.SetValue(float32((i + 1) + id*2))
			}

			payload := utils_test.Payload(wd)
			err := store.StoreMetrics(wd.FriendlyName, payload)
			if err != nil {
				t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestStoreDeviceUpdateStoresMetricsIfEnabled(t *testing.T) {
	wg := &sync.WaitGroup{}

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)

	//create device repo
	deviceRepo := repository.NewMemoryDeviceRepo()

	//create settings repo
	settingsTempFile := utils_test.Tempfile()
	defer os.Remove(settingsTempFile)
	settingsRepo, err := repository.NewFileSettingsRepoFromFile(settingsTempFile)
	if err != nil {
		t.Fatalf("settings repo failed. error: %v ", err.Error())
	}

	// get first device and enable metrics
	dev1 := wantDevices[0]
	appConfig := settings.NewAppConfig()
	deviceConfig := settings.NewDeviceConfig(dev1.Id)
	deviceConfig.MetricsEnabled = true
	appConfig.AddDeviceConfig(deviceConfig)
	settingsRepo.SaveAppConfig(appConfig)

	wg.Add(10) // Since rate limit is gone, loop below triggers it exactly 10 times
	// create metrics repo
	var metricsStoreHandler = func(id string, data map[string]any) {
		fmt.Printf("invoked with %v \n", id)
		if id != dev1.Id {
			t.Fatalf("invalid device invoked for metrics want %v got %v ", dev1.Id, id)
		}
		wg.Done()
	}

	metricsRepo := &mocks.NopMetricsRepo{}
	metricsRepo.WithStoreHandler(metricsStoreHandler)

	//create store
	store := utils_test.CreateStoreFromRepos(deviceRepo, metricsRepo, settingsRepo)

	// store bridgeInfoList
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err = store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	// update all devices but assert that only metrics enabled device stores metrics

	// trigger multiple events for each device
	for i := 0; i < 10; i++ {
		for id, wd := range wantDevices {
			for _, we := range wd.Exposes {
				we.Data.SetValue(float32((i + 1) + id*2))
			}

			payload := utils_test.Payload(wd)
			store.StoreMetrics(wd.FriendlyName, payload)
			if err != nil {
				t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func TestStoreUpdateDevice(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	sort.Slice(wantDevices, func(i, j int) bool {
		return wantDevices[i].FriendlyName < wantDevices[j].FriendlyName
	})
	store := utils_test.CreateStore()

	// NOTE:
	// need to add bridgeinfo so the new devices can be registered withthe mapper
	// else if not found in bridge it will use the friendlyname to has the id for mapping
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	// store devices in store
	for _, wd := range wantDevices {
		err := store.StoreDevice(wd.FriendlyName, wd)
		if err != nil {
			t.Fatalf("error storing device %v, %v", wd.Id, err.Error())
		}
	}

	// update source devices and store them
	for id, wd := range wantDevices {
		wd.ConnectionType = "test connection type"
		wd.PowerSource = "test power source"
		for _, we := range wd.Exposes {
			we.Data.SetValue(id * 2)
		}

		err := store.StoreDevice(wd.FriendlyName, wd)
		if err != nil {
			t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
		}
	}
	gotDevices, err := store.AllDevices()
	if err != nil {
		t.Fatalf("error loading devices %v:", err.Error())
	}

	if len(gotDevices) != len(wantDevices) {
		t.Fatalf("wrong number of devices. want %v got %v ", len(wantDevices), len(gotDevices))
	}

	for id, wd := range wantDevices {
		gt := gotDevices[id]
		utils_test.ValidateDevice(t, wd, gt)
	}

	for id, wd := range wantDevices {
		gt := gotDevices[id]
		utils_test.ValidateDevice(t, wd, gt)
	}
}

func TestStoreLoadFindsDeviceById(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()

	// NOTE:
	// need to add bridgeinfo so the new devices can be registered withthe mapper
	// else if not found in bridge it will use the friendlyname to has the id for mapping
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	for _, wd := range wantDevices {
		err := store.StoreDevice(wd.FriendlyName, wd)
		if err != nil {
			t.Fatalf("error storing device %v, %v", wd.Id, err.Error())
		}
	}

	for _, wd := range wantDevices {
		gotDevice, err := store.FindDeviceById(wd.Id)
		if err != nil {
			t.Fatalf("error loading device %v error: %v", wd.Id, err.Error())
		}

		utils_test.ValidateDevice(t, wd, gotDevice)
	}
}

func TestStoreLoadFindsDeviceByFriendlyName(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()

	// NOTE:
	// need to add bridgeinfo so the new devices can be registered withthe mapper
	// else if not found in bridge it will use the friendlyname to has the id for mapping
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	// store devices in store
	for _, wd := range wantDevices {
		err := store.StoreDevice(wd.FriendlyName, wd)
		if err != nil {
			t.Fatalf("error storing device %v, %v", wd.Id, err.Error())
		}
	}
	for _, wd := range wantDevices {

		gotDevice, err := store.FindDeviceByFriendlyName(wd.FriendlyName)
		if err != nil {
			t.Fatalf("error loading device %v error: %v", wd.FriendlyName, err.Error())
		}

		utils_test.ValidateDevice(t, wd, gotDevice)
	}
}

func TestStoreFindDeviceByIds(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()

	// NOTE:
	// need to add bridgeinfo so the new devices can be registered withthe mapper
	// else if not found in bridge it will use the friendlyname to has the id for mapping
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	ids := []string{}

	// store devices in store
	for _, wd := range wantDevices {
		err := store.StoreDevice(wd.FriendlyName, wd)
		if err != nil {
			t.Fatalf("error storing device %v, %v", wd.Id, err.Error())
		}

		ids = append(ids, wd.Id)
	}

	gotDevices, err := store.FindDeviceByIds(ids)
	if err != nil {
		t.Fatalf("error loading devices by ids. error: %v", err.Error())
	}

	for id, wd := range wantDevices {
		gd := gotDevices[id]
		utils_test.ValidateDevice(t, wd, gd)
	}
}

func TestStoreFindBridgeInfoById(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()

	// NOTE:
	// need to add bridgeinfo so the new devices can be registered withthe mapper
	// else if not found in bridge it will use the friendlyname to has the id for mapping
	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	for _, wb := range bridgeList {
		gotBridge, err := store.FindBridgeInfoById(wb.IeeeAddress)
		if err != nil {
			t.Fatalf("error loading bridge %v error: %v", wb.IeeeAddress, err.Error())
		}

		utils_test.ValidateBridge(t, wb, gotBridge)
	}
}

func TestStoreFindBridgeInfoByFriendlyName(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()

	bridgeList := utils_test.CreateBridgeInfoList(wantDevices)
	err := store.StoreBridgeInfoList(bridgeList)
	if err != nil {
		t.Fatalf("error storing BridgeInfoList: %v", err.Error())
	}

	for _, wb := range bridgeList {
		gotBridge, err := store.FindBridgeInfoByFriendlyName(wb.FriendlyName)
		if err != nil {
			t.Fatalf("error loading bridge %v error: %v", wb.FriendlyName, err.Error())
		}

		utils_test.ValidateBridge(t, wb, gotBridge)
	}
}

func createMockLivingRoomButtonDevices(brightnessValue float64, actionTimeValue float64) []*devices.Device {
	//var devices map[string]*devices.Device = make(map[string]*devices.Device)
	var devices = []*devices.Device{}
	testDevices := []struct {
		id       string
		name     string
		property string
		data     any
	}{
		{id: "x1234", name: "livingroom", property: "brightness", data: brightnessValue},
		{id: "x5678", name: "button", property: "action_time", data: actionTimeValue},
	}

	for _, d := range testDevices {
		devices = append(devices, utils_test.CreateDevice(d.id, d.name, d.property, d.data, 0.0, 255.0))
	}

	return devices
}
