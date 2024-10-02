package utils_test

import (
	"encoding/json"
	"fmt"
	"node-herder/mocks"
	"node-herder/utils"
	"testing"
)

func setup() {
	utils.RemoveRemoteLoggerHook()
}

func TestRemoteLoggerEmitter(t *testing.T) {

	setup()

	testIndex := 0
	expectedEventName := "logger"
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

	handler := func(eventName string, data interface{}) error {

		message, ok := data.([]byte)
		if !ok {
			t.Errorf("Failed to unmarshal message: %v", data)
		}

		fmt.Println("handler called with ", string(message))

		testCase := testCases[testIndex]

		if eventName != expectedEventName {
			t.Errorf("event name is not correct want: %s got: %s", expectedEventName, eventName)
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
	emitter := mocks.NewMockRemoteLoggerEmitter(handler)
	utils.RegisterRemoteLoggerHook(emitter)

	utils.EnableRemoteLoggerHook(true)

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

func TestRemoveRemoteLoggerHook(t *testing.T) {

	setup()

	id := 0
	expectedMessages := []string{"Test Info message"}
	handler := func(eventName string, data interface{}) error {

		message, ok := data.([]byte)
		if !ok {
			t.Errorf("Failed to unmarshal message: %v", data)
		}

		var logMessage utils.LogMessage
		err := json.Unmarshal(message, &logMessage)

		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err.Error())
		}

		if logMessage.Level == "debug" {
			t.Errorf("log level is debug and is unsupported	")
		}

		expecteMessage := expectedMessages[id]
		if logMessage.Message != expecteMessage {
			t.Errorf("log message is not correct want: %s got: %s", expecteMessage, logMessage.Message)
		}

		if logMessage.Level != "info" {
			t.Errorf("log level is not correct want: %s got: %s", "info", logMessage.Level)
		}

		id++
		return nil
	}

	emitter := mocks.NewMockRemoteLoggerEmitter(handler)
	utils.RegisterRemoteLoggerHook(emitter)
	utils.EnableRemoteLoggerHook(true)

	utils.LogInfo("Test Info message")

	// disable remote logger
	utils.RemoveRemoteLoggerHook()

	utils.LogInfo("Test Info message after disable remote logger")
	if id != 1 {
		t.Errorf("Remote logger should be disabled")
	}
}

func TestEnableRemoteLoggerHook(t *testing.T) {

	setup()

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

	handler := func(eventName string, data interface{}) error {

		message, ok := data.([]byte)
		if !ok {
			t.Errorf("Failed to unmarshal message: %v", data)
		}
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

	emitter := mocks.NewMockRemoteLoggerEmitter(handler)
	utils.RegisterRemoteLoggerHook(emitter)

	for _, testCase := range testCases {
		utils.EnableRemoteLoggerHook(testCase.enabled)

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
