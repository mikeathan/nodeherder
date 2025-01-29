package ws_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"node-herder/internal/api"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubNewClientEventsAreReceived(t *testing.T) {

	var expectedPayload = []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}")
	var expectedMessage = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"

	wsHub := ws.NewWsHub()
	wsHub.Start()

	h := api.NewWsHandler(wsHub)

	for i := 0; i < 4; i++ {
		s, wsConn := NewTestWsServer(t, h)
		time.Sleep(100 * time.Millisecond)

		er := wsHub.Broadcast(ws.DeviceUpdated, expectedPayload)
		if er != nil {
			t.Fatalf("hub broadcast failed  %v", er)
		}
		reply := receiveWSMessage(t, wsConn)
		gotType := reply["type"]

		if gotType != ws.DeviceUpdated {
			t.Fatalf("Expected type %+v', got '%+v'", ws.DeviceUpdated, gotType)
		}
		gotData := reply["payload"]
		if gotData != expectedMessage {
			t.Fatalf("Expected message %+v', got '%+v'", expectedMessage, gotData)
		}

		defer s.Close()
		defer wsConn.Close()
	}

	defer wsHub.Close()

}

func TestHandlingLoadAutomationsMessage(t *testing.T) {

	wsHub := ws.NewWsHub()
	wsHub.Start()

	// input data
	inputTriggers := createTestAutomation()

	wsHub.OnLoadAutomations(func() interface{} {
		return inputTriggers
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer wsHub.Close()
	defer s.Close()
	//defer wsConn.Close()

	wsData := &ws.EventMessage{Type: ws.LoadAutomations, Payload: nil}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.Automations {
		t.Fatalf("Expected type %v', got '%+v'", ws.LoadAutomations, event.Type)
	}

	// output data
	var triggers []*automations.Device

	bytes, _ := json.Marshal(event.Payload)
	err = json.Unmarshal(bytes, &triggers)
	if err != nil {
		t.Fatal(err)
	}
	//

	for idx, trigger := range triggers {

		inputTrigger := inputTriggers[idx]
		if trigger.Enabled != inputTrigger.Enabled {
			t.Fatalf("unexpected trigger.Enabled value")
		}
		if trigger.FriendlyName != inputTrigger.FriendlyName {
			t.Fatalf("unexpected trigger.Name value")
		}
		if trigger.Description != inputTrigger.Description {
			t.Fatalf("unexpected trigger.Description value")
		}
		for sidx, sensorTrigger := range trigger.Triggers {
			inputSensorTrigger := inputTrigger.Triggers[sidx]

			if sensorTrigger.Name != inputSensorTrigger.Name {
				t.Fatalf("unexpected sensorTrigger.Name value")
			}

			for aidx, action := range sensorTrigger.Actions {
				sensorTriggerAction := action.(*automations.MqttTriggerAction)
				inputAction := inputSensorTrigger.Actions[aidx].(*automations.MqttTriggerAction)

				for eidx, expose := range inputAction.Exposes {
					if sensorTriggerAction.Exposes[eidx].Name != expose.Name {
						t.Fatalf("unexpected property.Name value")
					}
					if sensorTriggerAction.Exposes[eidx].Data != expose.Data {
						t.Fatalf("unexpected property.Data value")
					}
				}

				if sensorTriggerAction.Delay.Unit != inputAction.Delay.Unit {
					t.Fatalf("unexpected action.Delay.Unit value")
				}
				if sensorTriggerAction.Delay.Value != inputAction.Delay.Value {
					t.Fatalf("unexpected action.Delay.Value value")
				}
			}

			for cidx, condition := range sensorTrigger.Conditions {
				inputCondition := inputSensorTrigger.Conditions[cidx]
				if condition.EqualityOperator != inputCondition.EqualityOperator {
					t.Fatalf("unexpected condition.EqualityOperator  value")
				}
				if condition.Value != inputCondition.Value {
					t.Fatalf("unexpected condition.Value  value")
				}
				if condition.Name != inputCondition.Name {
					t.Fatalf("unexpected condition.Name  value")
				}
			}
		}
	}
}

func TestHandlingLoadHubStatesMessage(t *testing.T) {
	wsHub := ws.NewWsHub()
	wsHub.Start()

	inputDevices := createTestDevices()
	wsHub.OnLoadDevices(func() interface{} {
		return inputDevices
	})

	inputAppConfig := createAppconfig()
	wsHub.OnLoadAppConfig(func() (interface{}, error) {
		return inputAppConfig, nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	wsData := &ws.EventMessage{Type: ws.LoadHubSate, Payload: nil}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.HubState {
		t.Fatalf("Expected type %v', got '%+v'", ws.HubState, event.Type)
	}

	var hubState *settings.HubState

	bytes, _ := json.Marshal(event.Payload)
	err = json.Unmarshal(bytes, &hubState)
	if err != nil {
		t.Fatal(err)
	}

	// assert devices
	for idx, device := range hubState.Devices {

		inputDevice := inputDevices[idx]
		if device.Id != inputDevice.Id {
			t.Fatalf("unexpected device.Id value")
		}
		if device.FriendlyName != inputDevice.FriendlyName {
			t.Fatalf("unexpected device.FriendlyName value")
		}
		if device.Description != inputDevice.Description {
			t.Fatalf("unexpected device.Description value")
		}
		if device.ConnectionType != inputDevice.ConnectionType {
			t.Fatalf("unexpected device.ConnectionType value")
		}
		if device.PowerSource != inputDevice.PowerSource {
			t.Fatalf("unexpected device.PowerSource value")
		}
		for eidx, expose := range device.Exposes {
			inputExpose := inputDevice.Exposes[eidx]

			if expose.Name != inputExpose.Name {
				t.Fatalf("unexpected expose.Name value")
			}
			if expose.Description != inputExpose.Description {
				t.Fatalf("unexpected expose.Description value")
			}
			if expose.Data != inputExpose.Data {
				t.Fatalf("unexpected expose.Data value")
			}
			if expose.Unit != inputExpose.Unit {
				t.Fatalf("unexpected expose.Unit value")
			}
			for pidx, property := range expose.Properties {
				inputproperty := inputExpose.Properties[pidx]
				if property != inputproperty {
					t.Fatalf("unexpected property value")
				}
			}
		}
	}

	// assert app config
	if len(hubState.Config.Hub.Devices) != len(inputAppConfig.Hub.Devices) {
		t.Fatalf("Expected numer of appconfig devices. want %v', got '%v'", len(inputAppConfig.Hub.Devices), len(hubState.Config.Hub.Devices))
	}
	for id, d := range inputAppConfig.Hub.Devices {
		gotDeviceConfig := hubState.Config.Hub.Devices[id]
		if d.Id != gotDeviceConfig.Id {
			t.Fatalf("Expected device id %v', got '%v'", d.Id, gotDeviceConfig.Id)
		}
		if d.Disabled != gotDeviceConfig.Disabled {
			t.Fatalf("Expected Disabled %v', got '%v'", d.Disabled, gotDeviceConfig.Disabled)
		}
		if d.MetricsEnabled != gotDeviceConfig.MetricsEnabled {
			t.Fatalf("Expected MetricsEnabled %v', got '%v'", d.MetricsEnabled, gotDeviceConfig.MetricsEnabled)
		}

		if d.RateLimit.Value != gotDeviceConfig.RateLimit.Value {
			t.Fatalf("Expected RateLimit.Value %v', got '%v'", d.RateLimit.Value, gotDeviceConfig.RateLimit.Value)
		}
		if d.RateLimit.Unit != gotDeviceConfig.RateLimit.Unit {
			t.Fatalf("Expected RateLimit.Unit %v', got '%v'", d.RateLimit.Unit, gotDeviceConfig.RateLimit.Unit)
		}
	}
	defer s.Close()
	defer wsConn.Close()
	defer wsHub.Close()

}

// TODO:
// func TestHandlingLoadDeviceMessage(t *testing.T) {

// 	wsHub := ws.NewWsHub()

// 	// input data
// 	inputDevice := createDevice1()

// 	deviceName := "test Device 1"
// 	wsHub.OnLoadDevice(func(string) (interface{}, error) {
// 		return inputDevice, nil
// 	})

// 	h := api.NewWsHandler(wsHub)
// 	s, wsConn := NewTestWsServer(t, h)

// 	wsData := &ws.EventMessage{Type: ws.LoadDevice, Payload: nil}
// 	msg, err := wsData.MarshalJSON()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}

// 	SendMessage(t, wsConn, msg)

// 	_, m, err := wsConn.ReadMessage()
// 	if err != nil {
// 		t.Fatalf("%v", err)
// 	}

// 	var event ws.EventMessage
// 	err = json.Unmarshal(m, &event)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if event.Type != ws.Devices {
// 		t.Fatalf("Expected type %v', got '%+v'", ws.Devices, event.Type)
// 	}

// 	// output data
// 	var resultDevices []*devices.Device

// 	bytes, _ := json.Marshal(event.Payload)
// 	err = json.Unmarshal(bytes, &resultDevices)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	//

// 	for idx, device := range resultDevices {

// 		inputDevice := inputDevices[idx]
// 		if device.Id != inputDevice.Id {
// 			t.Fatalf("unexpected device.Id value")
// 		}
// 		if device.FriendlyName != inputDevice.FriendlyName {
// 			t.Fatalf("unexpected device.FriendlyName value")
// 		}
// 		if device.Description != inputDevice.Description {
// 			t.Fatalf("unexpected device.Description value")
// 		}
// 		if device.ConnectionType != inputDevice.ConnectionType {
// 			t.Fatalf("unexpected device.ConnectionType value")
// 		}
// 		if device.PowerSource != inputDevice.PowerSource {
// 			t.Fatalf("unexpected device.PowerSource value")
// 		}
// 		for eidx, expose := range device.Exposes {
// 			inputExpose := inputDevice.Exposes[eidx]

// 			if expose.Name != inputExpose.Name {
// 				t.Fatalf("unexpected expose.Name value")
// 			}
// 			if expose.Description != inputExpose.Description {
// 				t.Fatalf("unexpected expose.Description value")
// 			}
// 			if expose.Data != inputExpose.Data {
// 				t.Fatalf("unexpected expose.Data value")
// 			}
// 			if expose.Unit != inputExpose.Unit {
// 				t.Fatalf("unexpected expose.Unit value")
// 			}
// 			for pidx, property := range expose.Properties {
// 				inputproperty := inputExpose.Properties[pidx]
// 				if property != inputproperty {
// 					t.Fatalf("unexpected property value")
// 				}
// 			}
// 		}

// 	}

// 	defer s.Close()
// 	defer wsConn.Close()
// }

func TestSaveAutomation(t *testing.T) {
	testCases := []struct {
		err     error
		message string
	}{
		{err: nil, message: ws.OperationSuccess},
		{err: errors.New("error occurd"), message: ws.OperationFailed},
	}

	wsHub := ws.NewWsHub()
	wsHub.Start()

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	// input data
	inputAutomations := createTestAutomation()
	newItem := inputAutomations[0]

	for _, testCase := range testCases {
		wsHub.OnSaveAutomation(func(p interface{}) error {
			return testCase.err
		})

		wsData := &ws.EventMessage{Type: ws.SaveAutomation, Payload: newItem}
		msg, err := wsData.MarshalJSON()
		if err != nil {
			t.Fatal(err.Error())
		}

		SendMessage(t, wsConn, msg)

		_, m, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}

		var event ws.EventMessage
		err = json.Unmarshal(m, &event)
		if err != nil {
			t.Fatal(err)
		}

		if testCase.err != nil {
			response := event.Payload.(string)
			if response != testCase.err.Error() {
				t.Fatalf("Expected error %v', got '%+v'", testCase.err.Error(), response)
			}
		}
		if event.Type != testCase.message {
			t.Fatalf("Expected type %v', got '%+v'", testCase.message, event.Type)
		}
	}

	defer s.Close()
	defer wsConn.Close()
	wsHub.Close()
}

func TestDeleteAutomation(t *testing.T) {

	// input data
	inputAutomations := createTestAutomation()
	newItem := inputAutomations[0]

	testCases := []struct {
		err     error
		message string
		payload interface{}
	}{
		{err: nil, message: ws.Automations, payload: newItem},
		{err: errors.New("error occured"), message: ws.OperationFailed, payload: newItem},
		{err: errors.New("payload is empty"), message: ws.OperationFailed, payload: nil},
	}

	wsHub := ws.NewWsHub()
	wsHub.Start()

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	for _, testCase := range testCases {
		wsHub.OnDeleteAutomation(func(p interface{}) (interface{}, error) {

			return inputAutomations, testCase.err
		})

		wsData := &ws.EventMessage{Type: ws.DeleteAutomation, Payload: testCase.payload}
		msg, err := wsData.MarshalJSON()
		if err != nil {
			t.Fatal(err.Error())
		}

		SendMessage(t, wsConn, msg)

		_, m, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}

		var event ws.EventMessage
		err = json.Unmarshal(m, &event)
		if err != nil {
			t.Fatal(err)
		}

		if testCase.err != nil {
			response := event.Payload.(string)
			if response != testCase.err.Error() {
				t.Fatalf("Expected error %v', got '%+v'", testCase.err.Error(), response)
			}
		}
		if event.Type != testCase.message {
			t.Fatalf("Expected type %v', got '%+v'", testCase.message, event.Type)
		}
	}

	defer s.Close()
	defer wsConn.Close()
}

func TestDeleteAutomationTrigger(t *testing.T) {

	// input data
	inputAutomations := createTestAutomation()
	deviceTrigger := inputAutomations[0]

	payload := `{"automationId":"` + deviceTrigger.Id + `", "triggerId":1}`
	testCases := []struct {
		err     error
		message string
		payload interface{}
	}{
		{err: nil, message: ws.AutomationUpdated, payload: payload},
	}

	wsHub := ws.NewWsHub()
	wsHub.Start()

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	for _, testCase := range testCases {
		wsHub.OnDeleteAutomationTrigger(func(p interface{}) (interface{}, error) {

			bytes := []byte(p.(string))
			payload := make(map[string]interface{})

			err := json.Unmarshal(bytes, &payload)

			if err != nil {
				fmt.Println(err.Error())
				return nil, errors.New("delete automation trigger failed. Invalid payload type")
			}

			automationId := payload["automationId"].(string)
			triggerId, err := strconv.Atoi(fmt.Sprint(payload["triggerId"]))
			if err != nil {
				fmt.Println(err.Error())
				return nil, errors.New("delete automation trigger failed. Invalid triggerId type")
			}

			fmt.Println("automationId:", automationId, "triggerId:", triggerId)
			for _, inputAutomation := range inputAutomations {
				if inputAutomation.Id == automationId {

					origSize := len(inputAutomation.Triggers)
					inputAutomation.Triggers = append(inputAutomation.Triggers[:triggerId], inputAutomation.Triggers[triggerId+1:]...)

					if len(inputAutomation.Triggers) == origSize {
						return nil, fmt.Errorf("delete automation trigger failed. original size is equal to new size")
					}

					return inputAutomation, nil

				}
			}

			return nil, testCase.err
		})

		wsData := &ws.EventMessage{Type: ws.DeleteAutomationTrigger, Payload: testCase.payload}
		msg, err := wsData.MarshalJSON()
		if err != nil {
			t.Fatal(err.Error())
		}

		SendMessage(t, wsConn, msg)

		_, m, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}

		var event ws.EventMessage
		err = json.Unmarshal(m, &event)
		if err != nil {
			t.Fatal(err)
		}

		if testCase.err != nil {
			response := event.Payload.(string)
			if response != testCase.err.Error() {
				t.Fatalf("Expected error %v', got '%+v'", testCase.err.Error(), response)
			}
		}
		if event.Type != testCase.message {
			t.Fatalf("Expected type %v', got '%+v'", testCase.message, event.Type)
		}
	}

	defer s.Close()
	defer wsConn.Close()
}

func TestLoadAppConfigMessage(t *testing.T) {

	inputAppConfig := createAppconfig()
	wsHub := ws.NewWsHub()
	wsHub.Start()

	wsHub.OnLoadAppConfig(func() (interface{}, error) {
		return inputAppConfig, nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer wsHub.Close()
	defer s.Close()
	defer wsConn.Close()

	wsData := &ws.EventMessage{Type: ws.LoadAppconfig, Payload: nil}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.AppConfig {
		t.Fatalf("Expected type %v', got '%v'", ws.AppConfig, event.Type)
	}

	// output data
	var resultAppConfig *settings.AppConfig

	bytes, _ := json.Marshal(event.Payload)
	err = json.Unmarshal(bytes, &resultAppConfig)
	if err != nil {
		t.Fatal(err)
	}

	if len(resultAppConfig.Hub.Devices) != len(inputAppConfig.Hub.Devices) {
		t.Fatalf("Expected numer of appconfig devices. want %v', got '%v'", len(inputAppConfig.Hub.Devices), len(resultAppConfig.Hub.Devices))
	}
	for id, d := range inputAppConfig.Hub.Devices {
		gotDeviceConfig := resultAppConfig.Hub.Devices[id]
		if d.Id != gotDeviceConfig.Id {
			t.Fatalf("Expected device id %v', got '%v'", d.Id, gotDeviceConfig.Id)
		}
		if d.Disabled != gotDeviceConfig.Disabled {
			t.Fatalf("Expected Disabled %v', got '%v'", d.Disabled, gotDeviceConfig.Disabled)
		}
		if d.MetricsEnabled != gotDeviceConfig.MetricsEnabled {
			t.Fatalf("Expected MetricsEnabled %v', got '%v'", d.MetricsEnabled, gotDeviceConfig.MetricsEnabled)
		}

		if d.RateLimit.Unit != gotDeviceConfig.RateLimit.Unit {
			t.Fatalf("Expected RateLimit.Unit %v', got '%v'", d.RateLimit.Unit, gotDeviceConfig.RateLimit.Unit)
		}
		if d.RateLimit.Value != gotDeviceConfig.RateLimit.Value {
			t.Fatalf("Expected RateLimit.Value %v', got '%v'", d.RateLimit.Value, gotDeviceConfig.RateLimit.Value)
		}
	}
}

func TestSaveConfigMessage(t *testing.T) {

	inputAppConfig := createAppconfig()

	modifiedHistory := inputAppConfig.Hub.History
	modifiedHistory.ExpireAt = utils.IntervalFromDays(7891)
	modifiedHistory.SleepTimeout = utils.IntervalFromHours(123)

	wsHub := ws.NewWsHub()
	wsHub.Start()

	wsHub.OnSaveHistoryConfig(func(p interface{}) error {

		bytes := []byte(p.(string))
		payload := &settings.HistoryConfig{}

		err := json.Unmarshal(bytes, &payload)
		if err != nil {
			fmt.Println(err.Error())
			return errors.New("save device config failed. Invalid payload type")
		}

		if payload.ExpireAt.Value != modifiedHistory.ExpireAt.Value {
			t.Fatalf("Expected history ExpireAt Value %v', got '%v'", modifiedHistory.ExpireAt.Value, payload.ExpireAt.Value)
		}

		if payload.ExpireAt.Unit != modifiedHistory.ExpireAt.Unit {
			t.Fatalf("Expected history ExpireAt unit %v', got '%v'", modifiedHistory.ExpireAt.Unit, payload.ExpireAt.Unit)
		}
		if payload.SleepTimeout.Value != modifiedHistory.SleepTimeout.Value {
			t.Fatalf("Expected history SleepTimeout value %v', got '%v'", modifiedHistory.SleepTimeout.Value, payload.SleepTimeout.Value)
		}

		if payload.SleepTimeout.Unit != modifiedHistory.SleepTimeout.Unit {
			t.Fatalf("Expected history SleepTimeout unit %v', got '%v'", modifiedHistory.SleepTimeout.Unit, payload.SleepTimeout.Unit)
		}
		return nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer s.Close()
	defer wsConn.Close()

	reqBytes, _ := json.Marshal(modifiedHistory)
	wsData := &ws.EventMessage{Type: ws.SaveHistoryConfig, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.OperationSuccess {
		t.Fatalf("Expected type %v', got '%v'", ws.OperationSuccess, event.Type)
	}
}

func TestSaveDeviceConfigMessage(t *testing.T) {

	inputAppConfig := createAppconfig()

	modifiedDevConfig := inputAppConfig.Hub.Devices["x0333444"]
	wsHub := ws.NewWsHub()
	wsHub.Start()

	wsHub.OnSaveDeviceConfig(func(p interface{}) error {

		bytes := []byte(p.(string))
		payload := &settings.DeviceConfig{}

		err := json.Unmarshal(bytes, &payload)
		if err != nil {
			fmt.Println(err.Error())
			return errors.New("save device config failed. Invalid payload type")
		}

		if payload.Id != modifiedDevConfig.Id {
			t.Fatalf("Expected device id %v', got '%v'", modifiedDevConfig.Id, payload.Id)
		}
		if payload.Disabled != modifiedDevConfig.Disabled {
			t.Fatalf("Expected disabled%v', got '%v'", modifiedDevConfig.Disabled, payload.Disabled)
		}

		if payload.MetricsEnabled != modifiedDevConfig.MetricsEnabled {
			t.Fatalf("Expected MetricsEnabled %v', got '%v'", modifiedDevConfig.MetricsEnabled, payload.MetricsEnabled)
		}

		if payload.RateLimit.Unit != modifiedDevConfig.RateLimit.Unit {
			t.Fatalf("Expected RateLimit.Unit %v', got '%v'", modifiedDevConfig.RateLimit.Unit, payload.RateLimit.Unit)
		}
		if payload.RateLimit.Value != modifiedDevConfig.RateLimit.Value {
			t.Fatalf("Expected RateLimit.Value %v', got '%v'", modifiedDevConfig.RateLimit.Value, payload.RateLimit.Value)
		}
		return nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer s.Close()
	defer wsConn.Close()

	reqBytes, _ := json.Marshal(modifiedDevConfig)
	wsData := &ws.EventMessage{Type: ws.SaveDeviceConfig, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.OperationSuccess {
		t.Fatalf("Expected type %v', got '%v'", ws.OperationSuccess, event.Type)
	}
}

func TestHandleBridgeDeviceRemoveMessage(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	wsHub := ws.NewWsHub()
	wsHub.Start()

	deviceId := "x0123456"
	forceRemove := true

	wsHub.OnDeviceRemove(func(p interface{}) error {

		bytes := []byte(p.(string))

		req := devices.DeviceRemoveRequest{}
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return errors.New("device remove failed. Invalid payload type")
		}

		if req.ID != deviceId {
			t.Fatalf("Expected device id %v', got '%v'", deviceId, req.ID)
		}

		if req.Force != forceRemove {
			t.Fatalf("Expected forceRemove %v', got '%v'", forceRemove, req.Force)
		}

		wg.Done()
		return nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer s.Close()
	defer wsConn.Close()

	reqBytes, _ := json.Marshal(devices.NewDeviceRemoveRequest(deviceId, forceRemove))

	wsData := &ws.EventMessage{Type: ws.DeviceRemove, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	// we dont send back response so just assert the logic in the hanlder
	wg.Wait()
}

func TestHandleBridgeDeviceInterviewMessage(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	wsHub := ws.NewWsHub()
	wsHub.Start()

	deviceId := "x0123456"
	wsHub.OnDeviceInterview(func(p interface{}) error {

		bytes := []byte(p.(string))
		payload := make(map[string]interface{})

		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			fmt.Println(err.Error())
			return errors.New("delete automation trigger failed. Invalid payload type")
		}

		if payload["id"] != deviceId {
			t.Fatalf("Expected device id %v', got '%v'", deviceId, payload["id"])

		}
		wg.Done()
		return nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer s.Close()
	defer wsConn.Close()

	reqBytes, _ := json.Marshal(map[string]interface{}{"id": deviceId})
	wsData := &ws.EventMessage{Type: ws.DeviceInterview, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	// we dont send back response so just asset the logic in the hanlder
	wg.Wait()
}

func TestHandleBridgePermitJoin(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	wsHub := ws.NewWsHub()
	wsHub.Start()

	req := make(map[string]interface{})
	req["value"] = true
	req["time"] = 10

	wsHub.OnBridgePermitJoin(func(p interface{}) error {

		bytes := []byte(p.(string))
		payload := make(map[string]interface{})

		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			fmt.Println(err.Error())
			return errors.New("delete automation trigger failed. Invalid payload type")
		}

		if payload["value"] != true {
			t.Fatalf("Expected value %v', got '%v'", true, payload["value"])
		}

		if payload["time"] != float64(10) {
			t.Fatalf("Expected time %v', got '%v'", 10, payload["time"])
		}

		wg.Done()
		return nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer s.Close()
	defer wsConn.Close()

	reqBytes, _ := json.Marshal(req)
	wsData := &ws.EventMessage{Type: ws.BridgePermitJoin, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	// we dont send back response so just asset the logic in the hanlder
	wg.Wait()
}

func TestHandlingEnableRemoteLoggerMessage(t *testing.T) {
	wsHub := ws.NewWsHub()
	wsHub.Start()
	req := &settings.LoggerConfig{EnableRemoteLogger: true}

	wsHub.OnSaveLoggerConfig(func(p interface{}) error {
		bytes := []byte(p.(string))
		payload := &settings.LoggerConfig{}
		err := json.Unmarshal(bytes, &payload)
		if err != nil {
			t.Fatal("enable remote logger failed. Invalid payload type")
		}

		if payload.EnableRemoteLogger != req.EnableRemoteLogger {
			t.Fatalf("enable remote logger failed. want %v got %v", req.EnableRemoteLogger, payload.EnableRemoteLogger)
		}
		return nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	defer s.Close()
	defer wsConn.Close()

	reqBytes, _ := json.Marshal(req)
	wsData := &ws.EventMessage{Type: ws.SaveLoggerConfig, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.OperationSuccess {
		t.Fatalf("Expected type %v', got '%v'", ws.OperationSuccess, event.Type)
	}
}

func TestHandlingLoadMetricsMessage(t *testing.T) {

	wsHub := ws.NewWsHub()
	wsHub.Start()

	now := time.Now()

	from := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, time.UTC)

	// input data
	expose1 := metrics.NewExposeNumericMetricResult("temperature", from, to)
	timestamps := utils_test.CreateDateTimeTimestamps(1, 24, 1)
	values := utils_test.CreateFloatValues(24)
	for idx, value := range values {
		expose1.Add(value, timestamps[idx])
	}

	expose2 := metrics.NewExposeBinaryMetricResult("presence", from, to).(*metrics.ExposeTimeRangeMetricsResult)
	timestamps2 := utils_test.CreateDateTimeTimestamps(1, 10, 1)

	values2 := utils_test.CreateBinaryValues(10)
	expose2 = utils_test.AddBinaryDataToExposeMetricsResult(expose2, values2, timestamps2)

	expose3 := metrics.NewExposeEnumMetricResult("color_temp", from, to).(*metrics.ExposeTimeRangeMetricsResult)
	timestamps3 := utils_test.CreateDateTimeTimestamps(1, 5, 1)
	values3 := utils_test.CreateEnumValues(5)
	expose3 = utils_test.AddBinaryDataToExposeMetricsResult(expose3, values3, timestamps3)

	viewMetrics := metrics.NewDeviceMetricsResult("x01234")
	viewMetrics.Add(expose1)
	viewMetrics.Add(expose2)
	viewMetrics.Add(expose3)

	wsHub.OnLoadMetrics(func(p interface{}) (interface{}, error) {
		return viewMetrics, nil
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	req := metrics.LoadDeviceMetricsRequest{}
	req.Id = "x01234"
	req.From = from.UnixMilli()
	req.To = to.UnixMilli()
	reqBytes, err := utils_test.StructToBytes(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	wsData := &ws.EventMessage{Type: ws.LoadMetrics, Payload: reqBytes}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatal(err.Error())
	}

	defer s.Close()
	defer wsConn.Close()

	SendMessage(t, wsConn, msg)

	_, m, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var event ws.EventMessage
	err = json.Unmarshal(m, &event)
	if err != nil {
		t.Fatal(err)
	}

	if event.Type != ws.Metrics {
		t.Fatalf("Expected type %v', got '%v'", ws.Metrics, event.Type)
	}

	// output data
	var resultMetrics *metrics.DeviceMetricsResult

	bytes, _ := json.Marshal(event.Payload)
	err = json.Unmarshal(bytes, &resultMetrics)
	if err != nil {
		t.Fatal(err)
	}
	if resultMetrics.DeviceId != viewMetrics.DeviceId {
		t.Fatalf("Expected result id %v', got '%v'", viewMetrics.DeviceId, resultMetrics.DeviceId)
	}
	if len(resultMetrics.Exposes) != len(viewMetrics.Exposes) {
		t.Fatalf("Expected numer of exposes %v', got '%v'", len(viewMetrics.Exposes), len(resultMetrics.Exposes))
	}
	for idx, gotExpose := range resultMetrics.Exposes {
		wantExpose := viewMetrics.Exposes[idx]

		if gotExpose.GetType() == "numeric" {
			utils_test.AssertNumericExposeMetricResults(wantExpose, gotExpose, t)
		} else if gotExpose.GetType() == "binary" {
			utils_test.AssertTimeRangeExposeMetricResults(wantExpose, gotExpose, t)
		} else if gotExpose.GetType() == "enum" {
			utils_test.AssertTimeRangeExposeMetricResults(wantExpose, gotExpose, t)
		} else {
			t.Errorf("invalid expose type %v: ", gotExpose.GetType())
		}
	}
}

func SendMessage(t *testing.T, ws *websocket.Conn, msg []byte) {
	t.Helper()

	// m, err := json.Marshal(msg)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("%v", err)
	}
}

func receiveWSMessage(t *testing.T, ws *websocket.Conn) map[string]interface{} {
	t.Helper()

	_, m, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var payload interface{}
	err = json.Unmarshal(m, &payload)
	if err != nil {
		t.Fatal(err)
	}

	data, ok := payload.(map[string]interface{})
	if !ok {
		t.Errorf("error converting payload. want type map[string]interface{};  got %T", payload)
		return nil
	}

	return data
}

func NewTestWsServer(t *testing.T, h http.Handler) (*httptest.Server, *websocket.Conn) {
	t.Helper()

	s := httptest.NewServer(h)
	wsURL := httpToWs(t, s.URL)

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	return s, ws
}

func httpToWs(t *testing.T, s string) string {
	t.Helper()

	wsURL, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	switch wsURL.Scheme {
	case "http":
		wsURL.Scheme = "ws"
	case "https":
		wsURL.Scheme = "wss"
	}

	return wsURL.String()
}

func createAppconfig() *settings.AppConfig {
	appConfig := settings.NewAppConfig()
	deviceConfig := settings.NewDeviceConfig("x01234")
	deviceConfig.MetricsEnabled = true
	deviceConfig.RateLimit = utils.IntervalFromMilliseconds(100)

	deviceConfig2 := settings.NewDeviceConfig("x0111222")
	deviceConfig2.MetricsEnabled = true
	deviceConfig2.RateLimit = utils.IntervalFromMinutes(1)
	appConfig.AddDeviceConfig(deviceConfig)

	deviceConfig3 := settings.NewDeviceConfig("x0333444")
	deviceConfig3.MetricsEnabled = false
	deviceConfig3.RateLimit = utils.IntervalFromMilliseconds(1111) // rate limit at 60000 ms
	appConfig.AddDeviceConfig(deviceConfig3)
	return appConfig
}
func createTestDevices() []*devices.Device {
	all := []*devices.Device{}
	all = append(all, createDevice1())
	all = append(all, createDevice2())

	return all
}

func createDevice1() *devices.Device {
	device1 := devices.NewDevice("x01234")
	device1.FriendlyName = "test Device 1"
	device1.ConnectionType = "mqtt"
	device1.Description = "some test dev 1 description"
	device1.PowerSource = "mains"
	device1.Properties = map[string]any{}
	device1.Properties["last_seen"] = time.Now().Format(time.RFC3339)
	device1.Properties["link_quality"] = 45.0
	device1.Exposes = make(map[string]*devices.Entity)

	ent1 := &devices.Entity{}
	ent1.Description = "temperature readings"
	ent1.Name = "temperature"
	ent1.Unit = "*c"
	ent1.Data = 50.0

	ent2 := &devices.Entity{}
	ent2.Description = "humidity readings"
	ent2.Name = "humidity"
	ent2.Unit = "%"
	ent2.Data = 64.1
	device1.Exposes["1"] = ent1
	device1.Exposes["2"] = ent2

	return device1
}

func createDevice2() *devices.Device {
	device := &devices.Device{}
	device.Id = "x34567"
	device.FriendlyName = "test Device 2"
	device.ConnectionType = "http"
	device.Description = "some test dev 2 description"
	device.PowerSource = "power"
	device.Properties = map[string]any{}
	device.Properties["last_seen"] = time.Now().Format(time.RFC3339)
	device.Properties["link_quality"] = 89.0
	device.Exposes = make(map[string]*devices.Entity)

	ent1 := &devices.Entity{}
	ent1.Description = "smart light livining room"
	ent1.Name = "brightness"
	ent1.Data = 78.0
	ent1.Properties = make(map[string]any)
	ent1.Properties["type"] = "numeric"
	ent1.Properties["max"] = 255.0
	ent1.Properties["min"] = 0.0

	device.Exposes["1"] = ent1

	return device
}

func createTestAutomation() []*automations.Device {
	mqtt := &mocks.MockMqttClient{}

	// create device trigger 1

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, utils.IntervalFromMilliseconds(100))
	turnOnTriggerWithLux := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)

	deviceTrigger := automations.NewDevice("human sensor")
	deviceTrigger.Description = "test human sensor automation"
	deviceTrigger.Triggers = []*automations.Trigger{}
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTriggerWithLux)

	// create device trigger 2
	turnOffTrigger2 := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, utils.IntervalFromMinutes(5))
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOn(mqtt)

	deviceTrigger2 := automations.NewDevice("Motion Sensor 2")
	deviceTrigger2.Description = "test outdoor motion sensor 2 automation"

	deviceTrigger2.Triggers = append(deviceTrigger2.Triggers, turnOffTrigger2)
	deviceTrigger2.Triggers = append(deviceTrigger2.Triggers, turnOnTrigger)

	var triggers []*automations.Device

	triggers = append(triggers, deviceTrigger)
	triggers = append(triggers, deviceTrigger2)
	//bytes, _ := json.Marshal(triggers)

	return triggers
}

func createTriggerTurnOnLightWithPresenceOnAndLux(mqtt mqtt.MqttClient, lux any) *automations.Trigger {
	// action = turn off light
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "state",
			Data: true,
		},
	}

	turnOnAction.Delay = utils.IntervalFromMilliseconds(0)
	turnOnAction.Client = mqtt

	// Turn on sensor trigger
	turnOnTrigger := &automations.Trigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// condition = presence = off && lux <= 30
	turnOnCondition := &automations.Condition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	luxCondition := &automations.Condition{}
	luxCondition.Name = "lux"
	luxCondition.EqualityOperator = "<="
	luxCondition.Value = lux

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createTriggerTurnOnLightWithPresenceOn(mqtt mqtt.MqttClient) *automations.Trigger {
	// action = turn off light
	turnOnAction :=automations.NewTriggerAction()
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "state",
			Data: true,
		},
	}
	turnOnAction.Delay = utils.IntervalFromMilliseconds(0)
	turnOnAction.Client = mqtt

	// Turn on sensor trigger
	turnOnTrigger := &automations.Trigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// condition = presence = off
	turnOnCondition := &automations.Condition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(mqtt mqtt.MqttClient, delay *utils.TimeInterval) *automations.Trigger {

	// action = turn off light
	turnOffAction := automations.NewTriggerAction()
	turnOffAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "state",
			Data: false,
		},
	}

	turnOffAction.Delay = delay
	turnOffAction.Client = mqtt

	// Turn off sensor trigger
	turnOffTrigger := &automations.Trigger{}
	turnOffTrigger.Name = "presence"
	turnOffTrigger.Actions = []automations.MqttAction{turnOffAction}

	// condition = presence == false
	turnOffCondition := &automations.Condition{}
	turnOffCondition.Name = "presence"
	turnOffCondition.EqualityOperator = "="
	turnOffCondition.Value = false

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}
