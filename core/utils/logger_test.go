package utils_test

import (
	"encoding/json"
	"fmt"
	"node-herder/utils"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestJsonHook(t *testing.T) {

	//todo create testCases wit messages and log levels to assert on
	testIndex := 0
	testCases := []struct {
		message string
		level   string
	}{
		{message: "message-debug-1", level: "debug"},
		{message: "message-info-1", level: "info"},
		{message: "message-warning-1", level: "warning"},
		{message: "message-error-1", level: "error"},
		{message: "message-error-2", level: "error"},
		{message: "message-debug-2", level: "debug"},
		{message: "message-debug-3", level: "debug"},
		{message: "message-info-2", level: "info"},
		{message: "message-warning-2", level: "warning"},
	}

	handler := func(message []byte) error {
		fmt.Println("handler called with ", string(message))

		testCase := testCases[testIndex]
		var logMessage utils.LogMessage
		err := json.Unmarshal(message, &logMessage)

		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err.Error())
		}

		if logMessage.Level == "debug" {
			t.Errorf("log level is debug and is unsupported	")
		}

		if logMessage.Message != testCase.message {
			t.Errorf("log message is not correct want: %s got: %s", testCase.message, logMessage.Message)
		}

		if logMessage.Level != testCase.level {
			t.Errorf("log level is not correct want: %s got: %s", testCase.level, logMessage.Level)
		}

		return nil
	}

	hook := utils.NewJsonHook(handler, true)
	utils.AddHook(hook)

	for _, testCase := range testCases {
		if testCase.level == "info" {
			utils.LogInfo(testCase.message)
		} else if testCase.level == "warning" {
			utils.LogWarn(testCase.message)
		} else if testCase.level == "error" {
			utils.LogError(testCase.message)
		} else if testCase.level == "debug" {
			utils.LogDebug(testCase.message)
		} else {
			t.Errorf("Unsupported log level")
		}

		testIndex++
	}

}

func TestJsonHookLevels(t *testing.T) {

	expectedLevels := []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.PanicLevel, logrus.FatalLevel}
	hook := utils.NewJsonHook(func(message []byte) error { return nil }, true)
	if !reflect.DeepEqual(hook.Levels(), expectedLevels) {
		t.Errorf("hook levels are not correct want: %v got: %v", expectedLevels, hook.Levels())
	}

}

func TestEnableDisableJsonHook(t *testing.T) {

	testIndex := 0
	testCases := []struct {
		message string
		level   string
		enabled bool
	}{
		{message: "message-info-1", level: "info", enabled: true},
		{message: "message-warning-1", level: "warning", enabled: true},
		{message: "message-error-1", level: "error", enabled: false},
		{message: "message-error-2", level: "error", enabled: true},
		{message: "message-info-2", level: "info", enabled: true},
		{message: "message-warning-2", level: "warning", enabled: false},
		{message: "message-debug-1", level: "debug", enabled: true},
		{message: "message-debug-2", level: "debug", enabled: false},

	}

	handler := func(message []byte) error {
		fmt.Println("handler called with ", string(message))

		testCase := testCases[testIndex]
		if !testCase.enabled {
			t.Errorf("hook should be disabled")
		}

		var logMessage utils.LogMessage
		err := json.Unmarshal(message, &logMessage)

		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err.Error())
		}

		if logMessage.Level == "debug" {
			t.Errorf("log level is debug and is unsupported	")
		}

		if logMessage.Message != testCase.message {
			t.Errorf("log message is not correct want: %s got: %s", testCase.message, logMessage.Message)
		}

		if logMessage.Level != testCase.level {
			t.Errorf("log level is not correct want: %s got: %s", testCase.level, logMessage.Level)
		}

		return nil
	}

	hook := utils.NewJsonHook(handler, true)
	utils.AddHook(hook)

	for _, testCase := range testCases {
		hook.Enabled(testCase.enabled)

		if testCase.level == "info" {
			utils.LogInfo(testCase.message)
		} else if testCase.level == "warning" {
			utils.LogWarn(testCase.message)
		} else if testCase.level == "error" {
			utils.LogError(testCase.message)
		} else if testCase.level == "debug" {
			utils.LogDebug(testCase.message)
		} else {
			t.Errorf("Unsupported log level")
		}

		testIndex++
	}

}
