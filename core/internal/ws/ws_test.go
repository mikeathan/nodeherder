package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"node-herder/transport/api"
	"node-herder/transport/ws"
	"testing"

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

func SendMessage(t *testing.T, ws *websocket.Conn, msg []byte) {
	t.Helper()

	m, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	if err := ws.WriteMessage(websocket.BinaryMessage, m); err != nil {
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
