package repository_test

import (
	"compress/gzip"
	"encoding/base64"
	"io"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"path/filepath"
	"strings"
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

func TestFileRepositoryCanRemoveDevice(t *testing.T) {

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

	err = repo.Remove(name)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = repo.FindDevice(name)
	if err == nil {
		t.Fatal("device should not exist")
	}
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
	b2f1 := devices.BridgeExpose{}
	b2f1.Access = devices.WriteBridgeAccessMode
	b2f1.Description = "On/off state of this light"
	b2f1.Name = "state"
	b2f1.Property = "state"
	b2f1.Type = "binary"
	b2f1.ValueOff = "OFF"
	b2f1.ValueOn = "ON"

	b2f2 := devices.BridgeExpose{}
	b2f2.Access = devices.WriteBridgeAccessMode
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

// Synthetic database created with Bolt v1.3.1 before the driver replacement.
func copyLegacyBoltFixture(t *testing.T) string {
	t.Helper()
	const data = "H4sIAAAAAAAA/+zdvWoUURQH8LOTL4yJpBRswhRWEVSsAqKgCD6DSNjduQlDvmRnNySELXwSCx/AJ7Cx9wUsLK2MlViL2Rs/AgEDxhX392v+98K9M2fZ+pyJbDrn0buFo+Lb4v5oPxO/ms05l7PI+frz+y+XPnTeBAAAAAAAAAAAAAAAAAAAAHBurZxn9f9PnTpfnDp/cv/F3bdXP1759OqCywUAAAAAAAAAAAAAAAAAAID/0kk//9KY6wAAAAAAAAAAAAAAAAAAAIBJdvJ9/2LMdQAAAAAAAAAAAAAAAAAAAMAkm85Z/DQBYDki5iNiISJaEXE9n7sXEVtpo909uLGZDvJqr701SM2gc+usNxTHzxhZiojFPG0g399O/V7dvXP7Yn8mAAAAAAAAAAAAAAAAAAAA/NNmchYxdZytiLgZEbMRMcz7lxExFxHP8/5aK+Jy7ubv9OpqI5319NO9//MR8eD7rcc767tPDss6pbTWrqpeappytcxzAaq0V3dTuVKu9+q0U20drO20t1O5Wj6q9/uDXlrOB4ZPR4vmd2tYjIiHP+YPjG4flnV1/nevlGn/2W6TmnL1cDjsp6a/1hl0N1N/+g/9OwAAAAAAAAAAAAAAAAAAAEyK2ZxL+fv/Rd5Pja0iAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAPj7vgYAAP//83mKIQAAAgA="
	r, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(data)))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	decoded, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "legacy.db")
	if err := os.WriteFile(filename, decoded, 0600); err != nil {
		t.Fatal(err)
	}
	return filename
}

func TestFileRepositoryLegacyReadWriteReopen(t *testing.T) {
	filename := copyLegacyBoltFixture(t)
	for pass := 0; pass < 2; pass++ {
		repo, err := repository.NewFileDeviceRepoFromFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = repo.Close() })
		device, err := repo.FindDevice("legacy-device")
		if err != nil || device == nil {
			t.Fatalf("legacy device: %v", err)
		}
		want := "Fixture device"
		if pass == 1 {
			want = "Updated fixture"
		}
		if device.Id != "legacy-device" || device.FriendlyName != want || len(device.Exposes) != 0 {
			t.Fatalf("legacy device changed: %+v", device)
		}
		bridge, err := repo.FindBridgeInfo("legacy-device")
		if err != nil || bridge == nil || bridge.FriendlyName != "Fixture device" {
			t.Fatalf("legacy bridge: %v, %v", bridge, err)
		}
		if pass == 0 {
			device.FriendlyName = "Updated fixture"
			if _, err := repo.Store(device.Id, device); err != nil {
				t.Fatal(err)
			}
			newDevice := devices.NewDevice("new-device")
			newDevice.FriendlyName = "New fixture"
			if _, err := repo.Store(newDevice.Id, newDevice); err != nil {
				t.Fatal(err)
			}
		} else {
			added, err := repo.FindDevice("new-device")
			if err != nil || added == nil || added.FriendlyName != "New fixture" {
				t.Fatalf("new device after reopen: %v, %v", added, err)
			}
		}
		if err := repo.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
