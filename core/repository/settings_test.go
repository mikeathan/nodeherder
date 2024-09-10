package repository_test

import (
	"node-herder/models/settings"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"reflect"
	"testing"
)

func TestFileSettingsRepositoryCanAddAndLoad(t *testing.T) {

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()

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

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()

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

	found, err := repo.FindOrAddDeviceConfigIfNotExists(id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	input := appConfig.Devices[id]

	if !reflect.DeepEqual(input, found) {
		t.Error("device config mismatch")
	}
}

func TestFileSettingsRepositoryCanAddNewDeviceConfig(t *testing.T) {

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()

	appConfig := createMockAppConfig()
	err = repo.Save(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	newCfg := &settings.DeviceConfig{}
	newCfg.Id = "x055555555"
	newCfg.Disabled = true
	newCfg.MetricsEnabled = false

	repo.SaveDeviceConfig(newCfg)

	found, err := repo.FindOrAddDeviceConfigIfNotExists(newCfg.Id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	if !reflect.DeepEqual(newCfg, found) {
		t.Error("device config mismatch")
	}
}

func TestFileSettingsRepositoryCanUpdateExistingDeviceConfig(t *testing.T) {

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()

	appConfig := createMockAppConfig()
	err = repo.Save(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	found, err := repo.FindOrAddDeviceConfigIfNotExists("x01234567")
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	if !found.MetricsEnabled {
		t.Error("found.history value invalid. want true got false")
	}
	found.MetricsEnabled = false

	err = repo.SaveDeviceConfig(found)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}
	updated, err := repo.FindOrAddDeviceConfigIfNotExists("x01234567")
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}
	if updated.MetricsEnabled {
		t.Error("updated.history value invalid. want false got true")

	}
}

func createMockAppConfig() *settings.AppConfig {
	appconfig := settings.NewAppConfig()
	cfg := &settings.DeviceConfig{}
	cfg.Id = "x01234567"
	cfg.Disabled = false
	cfg.MetricsEnabled = true

	appconfig.Add(cfg)

	cfg2 := &settings.DeviceConfig{}
	cfg2.Id = "x0erp09876"
	cfg2.Disabled = true
	cfg2.MetricsEnabled = false

	appconfig.Add(cfg2)

	cfg3 := &settings.DeviceConfig{}
	cfg3.Id = "x0lip1245h"
	cfg3.Disabled = false
	cfg3.MetricsEnabled = true

	appconfig.Add(cfg3)

	cfg4 := &settings.DeviceConfig{}
	cfg4.Id = "x9lo0124hggfs"
	cfg4.Disabled = false
	cfg4.MetricsEnabled = true

	appconfig.Add(cfg4)

	return appconfig
}
