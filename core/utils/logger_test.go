package utils_test

import (
	"fmt"
	"node-herder/utils"
	"testing"
)

func TestRemoteLogger(t *testing.T) {

	utils.LogDebug("message")

	emit := func(message []byte) error {
		fmt.Println(string(message))
		return nil
	}
	hook := utils.NewRemoteLogger(emit)
	utils.AddHook(hook)

}
