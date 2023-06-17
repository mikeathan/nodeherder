package hub_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"node-herder/hub"
	"strconv"
	"strings"
	"testing"
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


func TestHub(t *testing.T) {
	
	// /https://ieftimov.com/posts/testing-in-go-websockets/
	//2023/06/16 20:43:41 ws upgrade error:websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	//url:= makeWsProto("/ws")
	w := httptest.NewRecorder()
	h := hub.NewWsHandler("/ws")
	h.ServeHTTP(w, req)


	message := "test message 1"

	hub :=hub.NewEventHub()
	go hub.Run()
	hub.Broadcast([]byte(message))
}
func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

// func newWSServer(t *testing.T, h http.Handler) (*httptest.Server, *websocket.Conn) {
// 	t.Helper()

// 	s := httptest.NewServer(h)
// 	wsURL := httpToWs(t, s.URL)

// 	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	return s, ws
// }