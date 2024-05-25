package repository_test

import (
	"io/ioutil"
	"node-herder/models/settings"
	repository "node-herder/repository/settings"
	"os"
	"reflect"
	"testing"
)

func TestFileSettingsRepositoryCanAddAndFindValue(t *testing.T) {

	tempfile := tempfile()

	repo, err := repository.NewFileSettingsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise device file repo", err.Error())
	}

	defer repo.Close()
	defer os.Remove(tempfile)

	testCases := []struct {
		key   string
		value any
	}{
		{key: "test 1", value: createMockSettingsJson()},
		{key: "test 2", value: 10.56},
		{key: "test 3", value: false},
		{key: "test 4", value: 5.0},
	}

	for _, testCase := range testCases {
		err := repo.Store(testCase.key, testCase.value)
		if err != nil {
			t.Errorf("failed storing key %v, error %v", testCase.key, err.Error())
		}
	}

	for _, testCase := range testCases {

		val, err := repo.Get(testCase.key)
		if err != nil {
			t.Errorf("failed getting key %v, error %v", testCase.key, err.Error())
		}

		if !reflect.DeepEqual(val, testCase.value) {
			t.Errorf("value mismatch for key %v, error %v", testCase.key, err.Error())

		}
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
	cfg2.Id = "x01234567"
	cfg2.Disabled = false
	cfg2.History = true

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
