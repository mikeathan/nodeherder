package repository_test

import (
	"io/ioutil"
	"node-herder/models/settings"
	repository "node-herder/repository/settings"
	"os"
	"reflect"
	"testing"
)

func TestFileSettingsRepositoryCanAddAndLoad(t *testing.T) {

	tempfile := tempfile()

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	appConfig := createMockAppConfig()
	err = repo.Save(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	res, err := repo.Load()
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	for key, d := range res.Devices {
		inputDev := appConfig.Devices[key]
		if !reflect.DeepEqual(d, inputDev) {
			t.Error("device config mismatch")
		}
	}
}

func TestFileSettingsRepositoryCanAddAndFindValue(t *testing.T) {

	tempfile := tempfile()

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	appConfig := createMockAppConfig()
	err = repo.Save(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	dataKeys := make([]string, 0, len(appConfig.Devices))
	for k := range appConfig.Devices {
		dataKeys = append(dataKeys, k)
	}
	id := dataKeys[0]

	found, err := repo.FindDeviceConfig(id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	input := appConfig.Devices[id]

	if !reflect.DeepEqual(input, found) {
		t.Error("device config mismatch")
	}
}

func TestFileSettingsRepositoryCanAddNewDeviceConfig(t *testing.T) {

	tempfile := tempfile()

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	appConfig := createMockAppConfig()
	err = repo.Save(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	newCfg := &settings.DeviceConfig{}
	newCfg.Id = "x055555555"
	newCfg.Disabled = true
	newCfg.History = false

	repo.SaveDeviceConfig(newCfg)

	found, err := repo.FindDeviceConfig(newCfg.Id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	if !reflect.DeepEqual(newCfg, found) {
		t.Error("device config mismatch")
	}
}

func TestFileSettingsRepositoryCanUpdateExistingDeviceConfig(t *testing.T) {

	tempfile := tempfile()

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	appConfig := createMockAppConfig()
	err = repo.Save(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	dataKeys := make([]string, 0, len(appConfig.Devices))
	for k := range appConfig.Devices {
		dataKeys = append(dataKeys, k)
	}
	id := dataKeys[0]

	found, err := repo.FindDeviceConfig(id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	if (!found.History){
		t.Error("found.history value invalid. want true got false")

	}
	found.History = false

	err = repo.SaveDeviceConfig(found)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}
	updated, err := repo.FindDeviceConfig(id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}
	if (updated.History){
		t.Error("updated.history value invalid. want false got true")

	}
}

func createMockAppConfig() *settings.AppConfig {
	appconfig := settings.NewAppConfig()
	cfg := settings.DeviceConfig{}
	cfg.Id = "x01234567"
	cfg.Disabled = false
	cfg.History = true

	appconfig.Add(&cfg)

	cfg2 := settings.DeviceConfig{}
	cfg2.Id = "x0erp09876"
	cfg2.Disabled = true
	cfg2.History = false

	appconfig.Add(&cfg2)

	cfg3 := settings.DeviceConfig{}
	cfg3.Id = "x0lip1245h"
	cfg3.Disabled = false
	cfg3.History = true

	appconfig.Add(&cfg3)

	cfg4 := settings.DeviceConfig{}
	cfg4.Id = "x9lo0124hggfs"
	cfg4.Disabled = false
	cfg4.History = true

	appconfig.Add(&cfg4)

	return appconfig

}
func createMockSettingsJson() string {
	return `
	{
	  "id": "test_1",
	  "name": "settings file 1",
	  "items": [
		{ 
		 	"deviceId": "x01234",
		  	"metrics": false
		},
		{ 
			"deviceId": "x45567",
			"metrics": true
		},
		{ 
			"deviceId": "x78910",
			"metrics": false
		}
	  ]
	}
	`
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
