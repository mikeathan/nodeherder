package store_test

import (
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"testing"
)

func TestStoreLoadAllDevices(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()
	for _, wd := range wantDevices {
		store.StoreDevice(wd.FriendlyName, wd)
	}

	gotDevices, err := store.AllDevices()
	if err != nil {
		t.Fatalf("error loading devices %v:", err.Error())
	}

	if len(gotDevices) != len(wantDevices) {
		t.Fatalf("wrong number of devices. want %v got %v ", len(wantDevices), len(gotDevices))
	}

	for _, wd := range gotDevices {
		gt := wantDevices[wd.FriendlyName]
		utils_test.ValidateDevice(t, wd, gt)
	}
}

func TestStoreLoadFindsDeviceById(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()
	for _, wd := range wantDevices {
		store.StoreDevice(wd.FriendlyName, wd)
	}

	TODO
	// problem , it needs to call configure id mapper else is hashing the friendlyName
	//  which means i need to mock bridgeInfoList and call store.StoreBridge
	for _, wd := range wantDevices {
		gotDevice, err := store.FindDeviceById(wd.Id)
		if err != nil {
			t.Fatalf("error loading device %v error: %v", wd.Id, err.Error())
		}

		wantDevice := wantDevices[wd.FriendlyName]
		utils_test.ValidateDevice(t, wantDevice, gotDevice)
	}
}

func TestStoreLoadFindsDeviceByFriendlyName(t *testing.T) {

	wantDevices := createMockLivingRoomButtonDevices(255.0, 0.0)
	store := utils_test.CreateStore()
	for _, wd := range wantDevices {
		store.StoreDevice(wd.FriendlyName, wd)
	}

	testCases := []struct {
		id   string
		name string
	}{
		{id: "x1234", name: "livingroom"},
		{id: "x5678", name: "button"},
	}

	for _, testCase := range testCases {
		gotDevice, err := store.FindDeviceByFriendlyName(testCase.name)
		if err != nil {
			t.Fatalf("error loading device %v error: %v", testCase.name, err.Error())
		}

		wantDevice := wantDevices[testCase.name]
		utils_test.ValidateDevice(t, wantDevice, gotDevice)
	}
}

func createMockLivingRoomButtonDevices(brightnessValue float64, actionTimeValue float64) map[string]*devices.Device {
	var devices map[string]*devices.Device = make(map[string]*devices.Device)

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
		devices[d.name] = utils_test.CreateMockDevice(d.id, d.name, d.property, d.data, 0.0, 255.0)
	}

	return devices
}
