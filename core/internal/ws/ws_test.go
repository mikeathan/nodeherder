package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"node-herder/internal/api"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TODO: needs more work to store all connections and check if each clinets receives the message
func TestHubNewClientConnectedEventsTypesOfPayloads(t *testing.T) {

	testCases := []struct {
		Payload interface{}
		Message string
	}{
		{
			Payload: "{\"battery\":100,\"humidity\":66.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":41,\"temperature\":36,\"voltage\":2900}",
			Message: "{\"battery\":100,\"humidity\":66.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":41,\"temperature\":36,\"voltage\":2900}",
		},
		{
			Payload: []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"),
			Message: "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}",
		},
	}

	for _, testCase := range testCases {

		wsHub := ws.NewWsHub()
		wsHub.OnConnected(func() interface{} {
			return testCase.Payload
		})
		h := api.NewWsHandler(wsHub)
		s, wsConn := NewTestWsServer(t, h)

		reply := receiveWSMessage(t, wsConn)
		gotType := reply["type"]

		if gotType != ws.ClientConnected {
			t.Fatalf("Expected type %+v', got '%+v'", ws.ClientConnected, gotType)
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

func TestHubNewClientConnectedEvents(t *testing.T) {

	var expectedPayload = []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}")
	var expectedMessage = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"

	wsHub := ws.NewWsHub()
	wsHub.OnConnected(func() interface{} {
		return expectedPayload
	})
	h := api.NewWsHandler(wsHub)

	for i := 0; i < 4; i++ {
		s, wsConn := NewTestWsServer(t, h)

		reply := receiveWSMessage(t, wsConn)
		gotType := reply["type"]

		if gotType != ws.ClientConnected {
			t.Fatalf("Expected type %+v', got '%+v'", ws.ClientConnected, gotType)
		}
		gotData := reply["payload"]
		if gotData != expectedMessage {
			t.Fatalf("Expected message %+v', got '%+v'", expectedMessage, gotData)
		}

		defer s.Close()
		defer wsConn.Close()
	}
}

func TestHubNewClientEventsAreReceived(t *testing.T) {

	var expectedPayload = []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}")
	var expectedMessage = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"

	wsHub := ws.NewWsHub()
	h := api.NewWsHandler(wsHub)

	for i := 0; i < 4; i++ {
		s, wsConn := NewTestWsServer(t, h)

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
	inputTriggersBytes := createTestAutomation()
	var inputTriggers []*automations.Device
	err := json.Unmarshal(inputTriggersBytes, &inputTriggers)
	if err != nil {
		t.Fatal(err)
	}
	//
	wsHub.OnLoadAutomations(func() []byte {
		return inputTriggersBytes
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

	if event.Type != ws.LoadAutomations {
		t.Fatalf("Expected type %v', got '%+v'", ws.LoadAutomations, event.Type)
	}

	// output data
	var triggers []*automations.Device

	bytes := []byte(event.Payload.(string))
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
		if trigger.Name != inputTrigger.Name {
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
			if sensorTrigger.Action.Friendlyname != inputSensorTrigger.Action.Friendlyname {
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

func createTestAutomation() []byte {
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

	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger2)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	var triggers []*automations.Device

	triggers = append(triggers, deviceTrigger)
	triggers = append(triggers, deviceTrigger2)
	bytes, _ := json.Marshal(triggers)

	return bytes
}

func createTriggerTurnOnLightWithPresenceOnAndLux(mqtt mqtt.MqttClient, lux any) *automations.Trigger {
	// action = turn off light
	turnOnAction := &automations.MqttAction{}
	turnOnAction.Friendlyname = "Attic light"
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
	turnOnAction.Friendlyname = "Attic light"
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
	turnOffAction.Friendlyname = "Attic light"
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
