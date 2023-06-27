package hub_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"node-herder/hub"
	"testing"

	"github.com/gorilla/websocket"
)

// TODO: needs more work to store all connections and check if each clinets receives the message
func TestHubNewClientConnectedEvents(t *testing.T) {

	// testCases := []struct {
	// 	Payload interface{}
	// 	Message string
	// }{
	// 	{
	// 		Payload: "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}",
	// 		Message: "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}",
	// 	},
	// 	{
	// 		Payload: []byte("{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"),
	// 		Message: "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}",
	// 	},
	// }

	// wsConfig := hub.WsConfig{
	// 	OnConnected: func() interface{} {
	// 		return nil
	// 	}}

	// ws := hub.NewWsHub(wsConfig)
	// h := hub.NewWsHandler(ws)

	// for _, testCase := range testCases {
	// 	wsConfig.OnConnected = func() interface{} {
	// 		return testCase.Payload
	// 	}

	// 	for i := 0; i < 1; i++ {
	// 		s, ws := newWSServer(t, h)

	// 		reply := receiveWSMessage(t, ws)
	// 		gotType := reply["type"]
	// 		if gotType != hub.ClientConnected {
	// 			t.Fatalf("Expected type %+v', got '%+v'", hub.ClientConnected, gotType)
	// 		}
	// 		gotData := reply["payload"]
	// 		if gotData != testCase.Message {
	// 			t.Fatalf("Expected message %+v', got '%+v'", testCase.Message, gotData)
	// 		}

	// 		defer s.Close()
	// 		defer ws.Close()
	// 	}
	// }

	var message = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"
	//var bytesMessage = []byte(message)

	wsConfig := hub.WsConfig{
		OnConnected: func() interface{} {
			bytes := message
			return bytes
		}}

	ws := hub.NewWsHub(wsConfig)
	h := hub.NewWsHandler(ws)

	for i := 0; i < 1; i++ {
		s, ws := newWSServer(t, h)

		reply := receiveWSMessage(t, ws)
		gotType := reply["type"]
		if gotType != hub.ClientConnected {
			t.Fatalf("Expected type %+v', got '%+v'", hub.ClientConnected, gotType)
		}
		gotData := reply["payload"]
		if gotData != message {
			t.Fatalf("Expected message %+v', got '%+v'", message, gotData)
		}

		defer s.Close()
		defer ws.Close()
	}
}

func TestHubNewClientEventsAreReceived(t *testing.T) {

	message := "device updated"
	wsConfig := hub.WsConfig{}

	ws := hub.NewWsHub(wsConfig)
	h := hub.NewWsHandler(ws)

	for i := 0; i < 4; i++ {
		s, con := newWSServer(t, h)

		er := ws.Broadcast(hub.DeviceUpdated, message)
		if er != nil {
			t.Fatalf("hub broadcast failed  %v", er)
		}
		reply := receiveWSMessage(t, con)
		gotType := reply["type"]
		if gotType != hub.DeviceUpdated {
			t.Fatalf("Expected type %+v', got '%+v'", hub.ClientConnected, gotType)
		}
		gotData := reply["payload"]
		if gotData != message {
			t.Fatalf("Expected message %+v', got '%+v'", message, gotData)
		}

		defer s.Close()
		defer con.Close()
	}

}

func sendMessage(t *testing.T, ws *websocket.Conn, msg []byte) {
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

func newWSServer(t *testing.T, h http.Handler) (*httptest.Server, *websocket.Conn) {
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
