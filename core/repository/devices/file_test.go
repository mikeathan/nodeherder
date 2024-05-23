package repository_test

import (
	"io/ioutil"
	"node-herder/models/devices"
	repository "node-herder/repository/devices"
	"os"
	"testing"
)

func TestFileRepositoryCanAddOneDevice(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewFileDeviceRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	name := "device 1"
	device, _ := devices.CreateNewDevice("1", name, "mqtt", nil, createMockPayload(name, 50, 60.1, 23.5, 120.0))

	err = repo.Store(name, device)
	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := repo.FindDevice(name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	validateDevice(t, device, res)
}

func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
