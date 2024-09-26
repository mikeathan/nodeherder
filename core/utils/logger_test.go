package utils_test

import (
	"encoding/json"
	"fmt"
	"node-herder/utils"
	"testing"
	"time"
)

func TestRemoteLogger(t *testing.T) {

	//todo create testCases wit messages and log levels to assert on

	testCases := []struct {
		message string
		level   string
	}{
		{message: "message", level: "debug"},
		{message: "message", level: "info"},
		{message: "message", level: "warn"},
		{message: "message", level: "error"},
	}

	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	utils.LogDebug("message")

	emit := func(message []byte) error {
		fmt.Println(string(message))

		var logMessage utils.LogMessage
		err := json.Unmarshal(message, &logMessage)
		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err.Error())
		}

		if logMessage.Level == "debug" {
			t.Errorf("log level is debug and is unsupported	")
		}

		return nil
	}

	hook := utils.NewJsonHook(emit)
	utils.AddHook(hook)

	for _, testCase := range testCases {
		if testCase.level == "info" {
			utils.LogInfo(testCase.message)
		} else if testCase.level == "warn" {
			utils.LogWarn(testCase.message)
		} else if testCase.level == "error" {
			utils.LogError(testCase.message)
		} else if testCase.level == "debug" {
			utils.LogDebug(testCase.message)
		} else {
			t.Errorf("Unsupported log level")
		}
	}
	//wg.Wait()

	time.Sleep(10000 * time.Second)

}
