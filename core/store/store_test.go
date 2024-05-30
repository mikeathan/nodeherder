package store_test

import (
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"testing"
	"time"
)

func TestStoreLoadAllDevices(t *testing.T) {

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

func TestStoreUpdateStoresMetricsIfEnabled(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)

	//create device repo
	deviceRepo := repository.NewMemoryDeviceRepo()

	//create metrics repo
	metricsTempFile := utils_test.Tempfile()
	defer os.Remove(metricsTempFile)
	metricsRepo, err := repository.NewMetricsRepoFromFile(metricsTempFile)
	if err != nil {
		t.Fatalf("metrics repo failed. error: %v ", err.Error())
	}
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
	appConfig.Add(deviceConfig)
	settingsRepo.Save(appConfig)

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
				we.Data = (i + 1) + id*2
			}

			err := store.UpdateDevice(wd.FriendlyName, wd)
			if err != nil {
				t.Fatalf("error updating device %v error: %v:", wd.FriendlyName, err.Error())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// assert
}

func TestStoreUpdateDevice(t *testing.T) {

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

	// update source devices and store them
	for id, wd := range wantDevices {
		wd.ConnectionType = "test connection type"
		wd.PowerSource = "test power source"
		for _, we := range wd.Exposes {
			we.Data = id * 2
		}

		err := store.UpdateDevice(wd.FriendlyName, wd)
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
