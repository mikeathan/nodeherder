package repository_test

import (
	"node-herder/models/settings"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"node-herder/utils"
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
	err = repo.SaveAppConfig(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	res, err := repo.Load()
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	assertAppConfig(t, res, appConfig)
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
	err = repo.SaveAppConfig(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	dataKeys := make([]string, 0, len(appConfig.Hub.Devices.Overrides))
	for k := range appConfig.Hub.Devices.Overrides {
		dataKeys = append(dataKeys, k)
	}
	id := dataKeys[0]

	found, err := repo.LoadOrDefaultDeviceConfig(id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	input := appConfig.Hub.Devices.Overrides[id]

	assertDeviceConfig(t, found, input)
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
	err = repo.SaveAppConfig(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	newCfg := settings.NewDeviceConfig("x055555555")
	newCfg.Disabled = true
	newCfg.MetricsEnabled = false

	repo.SaveDeviceConfig(newCfg)

	found, err := repo.LoadOrDefaultDeviceConfig(newCfg.Id)
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	assertDeviceConfig(t, found, newCfg)
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
	err = repo.SaveAppConfig(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}

	found, err := repo.LoadOrDefaultDeviceConfig("x01234567")
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
	updated, err := repo.LoadOrDefaultDeviceConfig("x01234567")
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}
	if updated.MetricsEnabled {
		t.Error("updated.history value invalid. want false got true")

	}
}

func TestFileSettingsRepositoryAppConfigContainsBridgeConfig(t *testing.T) {

	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()

	appConfig := createMockAppConfig()
	err = repo.SaveAppConfig(appConfig)
	if err != nil {
		t.Errorf("save failed with %v", err.Error())
	}
	res, err := repo.Load()
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}
	if !reflect.DeepEqual(settings.NewBridgeConfig(), res.Bridge) {
		t.Error("bridge config mismatch")
	}
	bridgeCfg := settings.NewBridgeConfig()
	bridgeCfg.PermitJoin = true
	bridgeCfg.TimeExpireAt = &utils.TimeInterval{
		Value: 10,
		Unit:  "minutes",
	}

	repo.SaveBridgeConfig(bridgeCfg)

	res, err = repo.Load()
	if err != nil {
		t.Errorf("load failed with %v", err.Error())
	}

	assertAppConfig(t, res, appConfig)
}

func assertAppConfig(t *testing.T, res *settings.AppConfig, inputAppconfig *settings.AppConfig) {
	for key, d := range res.Hub.Devices.Overrides {
		inputDev := inputAppconfig.Hub.Devices.Overrides[key]
		assertDeviceConfig(t, d, inputDev)
	}
}

func assertDeviceConfig(t *testing.T, d *settings.DeviceConfig, inputDev *settings.DeviceConfig) {

	if d.Id != inputDev.Id {
		t.Error("device id mismatch")
	}
	if d.Disabled != inputDev.Disabled {
		t.Error("device disabled mismatch")
	}
	if d.MetricsEnabled != inputDev.MetricsEnabled {
		t.Error("device metrics enabled mismatch")
	}

	if !reflect.DeepEqual(d.RateLimit, inputDev.RateLimit) {
		t.Error("device rate limit mismatch")
	}
	if !reflect.DeepEqual(d.DefaultDebounceByCategory, inputDev.DefaultDebounceByCategory) {
		t.Error("device default debounce mismatch")
	}
}

func createMockAppConfig() *settings.AppConfig {
	appconfig := settings.NewAppConfig()
	cfg := settings.NewDeviceConfig("x01234567")
	cfg.Disabled = false
	cfg.MetricsEnabled = true

	appconfig.AddDeviceConfig(cfg)

	cfg2 := settings.NewDeviceConfig("x0erp09876")
	cfg2.Disabled = true
	cfg2.MetricsEnabled = false

	appconfig.AddDeviceConfig(cfg2)

	cfg3 := settings.NewDeviceConfig("x0lip1245h")
	cfg3.Disabled = false
	cfg3.MetricsEnabled = true

	appconfig.AddDeviceConfig(cfg3)

	cfg4 := settings.NewDeviceConfig("x9lo0124hggfs")
	cfg4.Disabled = false
	cfg4.MetricsEnabled = true

	appconfig.AddDeviceConfig(cfg4)

	return appconfig
}
