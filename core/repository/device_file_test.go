package repository_test

import (
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"testing"
)

func TestFileRepositoryCanAddAndFindBridgeInfo(t *testing.T) {

	tempfile := utils_test.Tempfile()

	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	bridgeInfo := createMockBridgeInfo()

	err = repo.StoreBridge(bridgeInfo)
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, device := range bridgeInfo {
		bridge, err := repo.FindBridgeInfo(device.IeeeAddress)
		if err != nil {
			t.Fatal(err.Error())
		}

		validateBridge(t, device, bridge)
	}
}

func TestFileRepositoryCanFindAllFindBridgeInfo(t *testing.T) {

	tempfile := utils_test.Tempfile()

	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	inputBridgeInfo := createMockBridgeInfo()

	err = repo.StoreBridge(inputBridgeInfo)
	if err != nil {
		t.Fatal(err.Error())
	}
	bridgeList, err := repo.AllBridgeInfo()
	if err != nil {
		t.Fatal(err.Error())
	}

	for idx, inputBridge := range inputBridgeInfo {
		outputBridge := bridgeList[idx]
		validateBridge(t, inputBridge, outputBridge)
	}
}
func TestFileRepositoryCanAddAndFindDevice(t *testing.T) {

	tempfile := utils_test.Tempfile()

	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	name := "device 1"
	device, _ := devices.CreateNewDevice("1", name, "mqtt", nil, createMockPayload(name, 50, 60.1, 23.5, 120.0))

	_, err = repo.Store(name, device)
	if err != nil {
		t.Fatal(err.Error())
	}

	res, err := repo.FindDevice(name)
	if err != nil {
		t.Fatal(err.Error())
	}

	validateDevice(t, device, res)
}

func TestFileRepositoryStoreReturnsStatusOfData(t *testing.T) {

	tempfile := utils_test.Tempfile()
	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}
	defer repo.Close()
	defer os.Remove(tempfile)

	name := "device 1"
	device, _ := devices.CreateNewDevice("1", name, "mqtt", nil, createMockPayload(name, 50, 60.1, 23.5, 120.0))

	isNew, _ := repo.Store(name, device)
	if !isNew {
		t.Fatalf("device does not exists")
	}

	device, _ = devices.CreateNewDevice("1", name, "mqtt", nil, createMockPayload(name, 100, 1.1, 12.5, 10.0))
	isNew, _ = repo.Store(name, device)
	if isNew {
		t.Fatalf("device does exists")
	}
}

func TestFileRepositoryCanFindDevices(t *testing.T) {

	tempfile := utils_test.Tempfile()

	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}
	defer repo.Close()
	defer os.Remove(tempfile)

	device, _ := devices.CreateNewDevice("1", "device 1", "mqtt", nil, createMockPayload("device 1", 50, 60.1, 23.5, 120.0))
	device2, _ := devices.CreateNewDevice("2", "device 2", "mqtt", nil, createMockPayload("device 2", 23, 12.1, 33.5, 111.0))
	devices := []*devices.Device{device, device2}

	for _, device := range devices {
		_, err = repo.Store(device.Id, device)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	res, err := repo.FindDevices([]string{"1", "2"})
	if err != nil {
		t.Fatal(err.Error())
	}

	for idx, resDevice := range res {
		sourceDevice := devices[idx]
		validateDevice(t, sourceDevice, resDevice)
	}
}

func TestFileRepositoryCanFindAllDevices(t *testing.T) {

	tempfile := utils_test.Tempfile()

	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}
	defer repo.Close()
	defer os.Remove(tempfile)

	device, _ := devices.CreateNewDevice("1", "device 1", "mqtt", nil, createMockPayload("device 1", 50, 60.1, 23.5, 120.0))
	device2, _ := devices.CreateNewDevice("2", "device 2", "mqtt", nil, createMockPayload("device 2", 23, 12.1, 33.5, 111.0))
	device3, _ := devices.CreateNewDevice("3", "device 3", "mqtt", nil, createMockPayload("device 3", 1, 12.1, 33.25, 1111.10))
	device4, _ := devices.CreateNewDevice("4", "device 4", "mqtt", nil, createMockPayload("device 4", 2, 22.1, 33.5, 321.0))

	devices := []*devices.Device{device, device2, device3, device4}

	for _, device := range devices {
		_, err = repo.Store(device.Id, device)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	res, err := repo.AllDevices()
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(res) != len(devices) {
		t.Errorf("devices found mismatch want %v got %v", len(devices), len(res))
	}

	for idx, resDevice := range res {
		sourceDevice := devices[idx]
		validateDevice(t, sourceDevice, resDevice)
	}
}

func createMockBridgeInfo() []*devices.BridgeInfo {

	dev1 := &devices.BridgeInfo{}
	dev1.IeeeAddress = "0x123456"
	dev1.Definition.Description = "Mocking human sensor"
	dev1.FriendlyName = "human sensor"
	dev1.Type = "EndDevice"
	dev1.PowerSource = "battery"
	dev1.Disabled = false
	dev1.InterviewCompleted = true

	e := devices.BridgeExpose{}
	e.Name = "presence"
	e.Property = "presence"
	e.Type = "binary"
	e.Description = "determines if presence has been detected"

	dev1.Definition.Exposes = append(dev1.Definition.Exposes, e)

	//
	dev2 := &devices.BridgeInfo{}
	dev2.IeeeAddress = "0x56789"
	dev2.Definition.Description = "Mocking Attic light"
	dev2.FriendlyName = "Attic light"
	dev2.Type = "EndDevice"
	dev2.PowerSource = "mains"
	dev2.Disabled = false
	dev2.InterviewCompleted = true

	e2 := devices.BridgeExpose{}
	e2.Type = "light"
	b2f1 := devices.BridgeInfoFeature{}
	b2f1.Description = "On/off state of this light"
	b2f1.Name = "state"
	b2f1.Property = "state"
	b2f1.Type = "binary"
	b2f1.ValueOff = "OFF"
	b2f1.ValueOn = "ON"

	b2f2 := devices.BridgeInfoFeature{}
	b2f2.Description = "Brightness of this light"
	b2f2.Name = "brightness"
	b2f2.Property = "brightness"
	b2f2.Type = "numeric"
	b2f2.ValueMax = 255
	b2f2.ValueMin = 0

	e2.Features = append(e2.Features, b2f1)
	e2.Features = append(e2.Features, b2f2)

	dev2.Definition.Exposes = append(dev2.Definition.Exposes, e2)

	return []*devices.BridgeInfo{dev1, dev2}
}
