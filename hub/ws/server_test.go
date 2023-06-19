package ws_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"node-herder/hub"
	"strconv"
	"testing"

	"github.com/gorilla/websocket"
)

type device struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

func TestXxx(t *testing.T) {
	return
	var topic = "zigbee2mqtt/TH1"

	var msg = "{'battery':100,'humidity':59.8,'last_seen':'2023-05-31T19:02:28+01:00','linkquality':51,'temperature':18.4,'voltage':3000}"
	// escape
	escapedmsg := strconv.Quote(msg)
	var payload = []byte(escapedmsg)
	payload1 := json.RawMessage(payload)
	var device = device{Name: topic, Payload: payload1}
	bytes, err := json.Marshal(device)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}

func TestHubReceiveMessagesFromMultipleConnections(t *testing.T) {

	// TODO: needs more work to store all connections and check if each clinets receives the message
	h := hub.Init("/")

	for i := 0; i < 4; i++ {
		s, ws := newWSServer(t, h)
		defer s.Close()
		defer ws.Close()
		message := fmt.Sprintf("test message %d", i)
		sendMessage(t, ws, []byte(message))

		reply := receiveWSMessage(t, ws)

		if string(reply) != message {
			t.Fatalf("Expected '%+v', got '%+v'", message, reply)
		}
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

func receiveWSMessage(t *testing.T, ws *websocket.Conn) []byte {
	t.Helper()

	_, m, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var reply []byte
	err = json.Unmarshal(m, &reply)
	if err != nil {
		t.Fatal(err)
	}

	return reply
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
