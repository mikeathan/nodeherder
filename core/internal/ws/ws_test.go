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
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TODO: needs more work to store all connections and check if each clinets receives the message
func TestHubNewClientConnectedEventsTypesOfPayloads(t *testing.T) {

	testCases := []struct {
		Event   string
		Payload []byte
		Message string
	}{
		{
			Event:   ws.Devices,
			Payload: []byte("{\"battery\":100,\"humidity\":66.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":41,\"temperature\":36,\"voltage\":2900}"),
			Message: "{\"battery\":100,\"humidity\":66.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":41,\"temperature\":36,\"voltage\":2900}",
		},
		{
			Event:   ws.Automations,
			Payload: []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"),
			Message: "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}",
		},
	}

	for _, testCase := range testCases {

		wsHub := ws.NewWsHub()
		wsHub.OnLoadDevices(func() interface{} {
			return testCase.Payload
		})
		h := api.NewWsHandler(wsHub)
		s, wsConn := NewTestWsServer(t, h)
		wsHub.Broadcast(testCase.Event, testCase.Payload)

		reply := receiveWSMessage(t, wsConn)
		gotType := reply["type"]

		if gotType != testCase.Event {
			t.Fatalf("Expected type %+v', got '%+v'", testCase.Event, gotType)
		}
		gotData := reply["payload"]
		wantData := testCase.Message
		if gotData != wantData {
			t.Fatalf("Expected message %+v', got '%+v'", wantData, gotData)
		}

		defer s.Close()
		defer wsConn.Close()
		wsConn.Close()
	}
}

func TestHubNewClientEventsAreReceived(t *testing.T) {

	var expectedPayload = []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}")
	var expectedMessage = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"

	wsHub := ws.NewWsHub()
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
}

func TestHandlingLoadAutomationsMessage(t *testing.T) {

	wsHub := ws.NewWsHub()

	// input data
	inputTriggers := createTestAutomation()

	wsHub.OnLoadAutomations(func() interface{} {
		return inputTriggers
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	wsData := &ws.EventMessage{Type: ws.LoadAutomations, Payload: nil}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatalf(err.Error())
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
			if sensorTrigger.Action.FriendlyName != inputSensorTrigger.Action.FriendlyName {
				t.Fatalf("unexpected sensorTrigger.Action.Friendlyname  value")
			}
			if sensorTrigger.Action.Property != inputSensorTrigger.Action.Property {
				t.Fatalf("unexpected sensorTrigger.Action.Property  value")
			}
			if sensorTrigger.Action.Type != inputSensorTrigger.Action.Type {
				t.Fatalf("unexpected sensorTrigger.Action.Type  value")
			}
			if sensorTrigger.Action.Delay != inputSensorTrigger.Action.Delay {
				t.Fatalf("unexpected sensorTrigger.Action.Delay  value")
			}
			if sensorTrigger.Action.Data != inputSensorTrigger.Action.Data {
				t.Fatalf("unexpected sensorTrigger.Action.Value  value")
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

	defer s.Close()
	defer wsConn.Close()
}

func TestHandlingLoadBridgeFeaturesMessage(t *testing.T) {
	wsHub := ws.NewWsHub()

	// input data
	inputFeatures := createTestFeatures()

	wsHub.OnLoadBridgeFeatures(func() interface{} {
		return inputFeatures
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	wsData := &ws.EventMessage{Type: ws.LoadBridgeFeatures, Payload: nil}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatalf(err.Error())
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

	if event.Type != ws.BridgeFeatures {
		t.Fatalf("Expected type %v', got '%+v'", ws.BridgeFeatures, event.Type)
	}

	// output data
	var resultFeatures []*devices.BridgeFeature
	bytes, _ := json.Marshal(event.Payload)
	err = json.Unmarshal(bytes, &resultFeatures)
	if err != nil {
		t.Fatal(err)
	}
	//

	for idx, feature := range resultFeatures {

		inputFeature := inputFeatures[idx]
		if feature.Id != inputFeature.Id {
			t.Fatalf("unexpected feature.Id value")
		}

		for key, property := range feature.Properties {

			inputProperty := inputFeature.Properties[key]
			if property.Name != inputProperty.Name {
				t.Fatalf("unexpected property name")
			}
			if property.Type != inputProperty.Type {
				t.Fatalf("unexpected property type")
			}
			for aKey, value := range property.Attributes {

				if value != inputProperty.Attributes[aKey] {
					t.Fatalf("unexpected attribute value")
				}
			}
		}
	}

	defer s.Close()
	defer wsConn.Close()
}

func TestHandlingLoadDevicesMessage(t *testing.T) {

	wsHub := ws.NewWsHub()

	// input data
	inputDevices := createTestDevices()

	wsHub.OnLoadDevices(func() interface{} {
		return inputDevices
	})

	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	wsData := &ws.EventMessage{Type: ws.LoadDevices, Payload: nil}
	msg, err := wsData.MarshalJSON()
	if err != nil {
		t.Fatalf(err.Error())
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

	if event.Type != ws.Devices {
		t.Fatalf("Expected type %v', got '%+v'", ws.Devices, event.Type)
	}

	// output data
	var resultDevices []*devices.DeviceV2

	bytes, _ := json.Marshal(event.Payload)
	err = json.Unmarshal(bytes, &resultDevices)
	if err != nil {
		t.Fatal(err)
	}
	//

	for idx, device := range resultDevices {

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

	defer s.Close()
	defer wsConn.Close()
}

func TestSaveAutomation(t *testing.T) {
	testCases := []struct {
		err     error
		message string
	}{
		{err: nil, message: ws.OperationSuccess},
		{err: errors.New("error occurd"), message: ws.OperationFailed},
	}

	wsHub := ws.NewWsHub()
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
			t.Fatalf(err.Error())
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

func TestDeleteAutomation(t *testing.T) {

	// input data
	inputAutomations := createTestAutomation()
	newItem := inputAutomations[0]

	testCases := []struct {
		err     error
		message string
		payload interface{}
	}{
		{err: nil, message: ws.OperationSuccess, payload: newItem},
		{err: errors.New("error occured"), message: ws.OperationFailed, payload: newItem},
		{err: errors.New("payload is empty"), message: ws.OperationFailed, payload: nil},
	}

	wsHub := ws.NewWsHub()
	h := api.NewWsHandler(wsHub)
	s, wsConn := NewTestWsServer(t, h)

	for _, testCase := range testCases {
		wsHub.OnDeleteAutomation(func(p interface{}) error {

			return testCase.err
		})

		wsData := &ws.EventMessage{Type: ws.DeleteAutomation, Payload: testCase.payload}
		msg, err := wsData.MarshalJSON()
		if err != nil {
			t.Fatalf(err.Error())
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
			t.Fatalf(err.Error())
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

func createTestFeatures() []*devices.BridgeFeature {
	all := []*devices.BridgeFeature{}

	f1 := devices.NewBridgeFeature("x0123")
	p := devices.NewBridgeProperty("state", "binary")
	p.Attributes["on"] = "ON"
	p.Attributes["off"] = "OFF"
	p.Attributes["toggle"] = "TOGGLE"

	p2 := devices.NewBridgeProperty("state", "binary")

	p2.Attributes["type"] = "numeric"
	p2.Attributes["max"] = 255.0
	p2.Attributes["min"] = 50.0
	f1.Add(p)
	f1.Add(p2)
	all = append(all, f1)
	return all
}
func createTestDevices() []*devices.DeviceV2 {
	all := []*devices.DeviceV2{}
	all = append(all, createDevice1())
	all = append(all, createDevice2())

	return all
}

func createDevice1() *devices.DeviceV2 {
	device1 := &devices.DeviceV2{}
	device1.Id = "x01234"
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

func createDevice2() *devices.DeviceV2 {
	device := &devices.DeviceV2{}
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
	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOnTriggerWithLux := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)

	deviceTrigger := automations.NewDevice("human sensor")
	deviceTrigger.Description = "test human sensor automation"
	deviceTrigger.Triggers = []*automations.Trigger{}
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTriggerWithLux)

	// create device trigger 2
	turnOffTrigger2 := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 5*time.Minute)
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
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "Attic light"
	turnOnAction.Type = "light"
	turnOnAction.Property = "state"
	turnOnAction.Data = true
	turnOnAction.Delay = 0
	turnOnAction.Client = mqtt

	// Turn on sensor trigger
	turnOnTrigger := &automations.Trigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Action = turnOnAction

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
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "Attic light"
	turnOnAction.Type = "light"
	turnOnAction.Property = "state"
	turnOnAction.Data = true
	turnOnAction.Delay = 0
	turnOnAction.Client = mqtt

	// Turn on sensor trigger
	turnOnTrigger := &automations.Trigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Action = turnOnAction

	// condition = presence = off
	turnOnCondition := &automations.Condition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(mqtt mqtt.MqttClient, delay time.Duration) *automations.Trigger {
	// action = turn off light
	turnOffAction := &automations.MqttAction{}
	turnOffAction.FriendlyName = "Attic light"
	turnOffAction.Type = "light"
	turnOffAction.Property = "state"
	turnOffAction.Data = false
	turnOffAction.Delay = delay
	turnOffAction.Client = mqtt

	// Turn off sensor trigger
	turnOffTrigger := &automations.Trigger{}
	turnOffTrigger.Name = "presence"
	turnOffTrigger.Action = turnOffAction

	// condition = presence == false
	turnOffCondition := &automations.Condition{}
	turnOffCondition.Name = "presence"
	turnOffCondition.EqualityOperator = "="
	turnOffCondition.Value = false

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}
