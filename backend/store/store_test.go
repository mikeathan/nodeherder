package store_test

import (
	"fmt"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
	"sort"
	"sync"
	"testing"
	"time"
)

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

	mockLoggerConfig := settings.NewLoggerConfig(true)
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
		deviceConfig.RateLimit = utils.IntervalFromMilliseconds(1)
		cfg.SetDeviceConfigOverrides(deviceConfig)
		if err != nil {
			t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
		}
	}

	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	// set new history config
	sleepTimeout := utils.IntervalFromSeconds(2) // start the cleanup after we finished ading and asserting the data. 2 seconds should be enough
	expireAt := utils.IntervalFromHours(1)

	cfg.SaveHistoryConfig(settings.NewHistoryConfig(sleepTimeout, expireAt))

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

	wg.Add(1)
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

func TestStoreMetricsLimitsDataWithDefaultRateLimiter(t *testing.T) {
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

	wg.Add(1)
	metircsHits := 0
	// create metrics repo
	var metricsStoreHandler = func(id string, data map[string]any) {
		fmt.Printf("invoked with %v \n", id)

		if metircsHits > 0 {
			t.Fatalf("rate limiter failed. we only expect 1 hit.")
		}

		if id != dev1.Id {
			t.Fatalf("invalid device invoked for metrics want %v got %v ", dev1.Id, id)
		}

		metircsHits++
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
			err := store.StoreMetrics(wd.FriendlyName, payload)
			if err != nil {
				t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func TestStoreMetricsLimitsDataWithConfiguredRateLimiter(t *testing.T) {
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
	deviceConfig.RateLimit = utils.IntervalFromMilliseconds(100)
	appConfig.AddDeviceConfig(deviceConfig)
	settingsRepo.SaveAppConfig(appConfig)

	wg.Add(10)
	metircsHits := 0
	// create metrics repo
	var metricsStoreHandler = func(id string, data map[string]any) {
		fmt.Printf("invoked with %v \n", id)

		if metircsHits >= 10 {
			t.Fatalf("invalid metrics hits  want 10 got %v.", metircsHits)
		}

		if id != dev1.Id {
			t.Fatalf("invalid device invoked for metrics want %v got %v ", dev1.Id, id)
		}

		metircsHits++
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
			err := store.StoreMetrics(wd.FriendlyName, payload)
			if err != nil {
				t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()
}

func TestStoreMetricsLimitsDataWithMultipleDevicesConfiguredRateLimiter(t *testing.T) {
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

	// configure 1st device
	dev1 := wantDevices[0]
	appConfig := settings.NewAppConfig()
	deviceConfig := settings.NewDeviceConfig(dev1.Id)
	deviceConfig.MetricsEnabled = true
	deviceConfig.RateLimit = utils.IntervalFromMilliseconds(100)

	// configure 2st device
	dev2 := wantDevices[1]
	deviceConfig2 := settings.NewDeviceConfig(dev2.Id)
	deviceConfig2.MetricsEnabled = true
	deviceConfig2.RateLimit = utils.IntervalFromMinutes(1)
	appConfig.AddDeviceConfig(deviceConfig)
	settingsRepo.SaveAppConfig(appConfig)

	expectedHits := map[string]int{}
	expectedHits[dev1.Id] = 24 // TODO: numbers are wrong - needs redoing
	expectedHits[dev2.Id] = 1  // TODO: numbers are wrong - needs redoing

	wg.Add(10)
	hitsCounter := map[string]int{}
	// create metrics repo
	var metricsStoreHandler = func(id string, data map[string]any) {
		fmt.Printf("invoked with %v \n", id)

		if hitsCounter[id] >= expectedHits[id] {
			t.Fatalf("invalid metrics hits want %v got %v.", expectedHits[id], hitsCounter[id])
		}

		if id != dev1.Id {
			t.Fatalf("invalid device invoked for metrics want %v got %v ", dev1.Id, id)
		}

		hitsCounter[id]++
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
			err := store.StoreMetrics(wd.FriendlyName, payload)
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
