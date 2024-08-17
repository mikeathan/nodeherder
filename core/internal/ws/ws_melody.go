package ws

import (
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

type EventHubMelogy interface {
	Broadcast(eventName string, data interface{}) error
	RegisterNewClient(conn *websocket.Conn)
	EmitDevices()
	EmitDeviceList(names []string)
	EmitDevice(name string) error
	OnLoadAutomations(action func() interface{})
	OnLoadDevices(action func() interface{})
	OnLoadDevice(action func(id string) (interface{}, error))
	OnLoadDeviceList(action func(ids []string) interface{})
	OnDeviceSetValue(func(payload interface{}) error)
	OnDeviceRename(func(payload interface{}) error)
	OnSaveAutomation(func(payload interface{}) error)
	OnDeleteAutomation(func(payload interface{}) (interface{}, error))
	OnDeleteAutomationTrigger(func(payload interface{}) (interface{}, error))
	OnLoadMetrics(action func(interface{}) (interface{}, error))
	OnLoadAppConfig(action func() (interface{}, error))
	OnSaveDeviceConfig(func(payload interface{}) error)
}

type wsMelodyServer struct {
	server                    *melody.Melody
	onLoadAutomations         func() interface{}
	onLoadDevices             func() interface{}
	onLoadDeviceList          (func(ids []string) interface{})
	onLoadDevice              func(id string) (interface{}, error)
	onLoadMetrics             func(interface{}) (interface{}, error)
	onSaveAutomation          func(interface{}) error
	onDeviceSetValue          func(interface{}) error
	onDeviceRename            func(interface{}) error
	onDeleteAutomation        func(interface{}) (interface{}, error)
	onDeleteAutomationTrigger func(interface{}) (interface{}, error)
	onLoadAppConfig           func() (interface{}, error)
	onSaveDeviceConfig        func(interface{}) error
}

func NewWsHubMelody() wsMelodyServer {

	wsHub := &wsMelodyServer{
		server:                    melody.New(),
		onSaveAutomation:          func(payload interface{}) error { return nil },
		onLoadMetrics:             func(interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomation:        func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeleteAutomationTrigger: func(payload interface{}) (interface{}, error) { return nil, nil },
		onDeviceSetValue:          func(payload interface{}) error { return nil },
		onDeviceRename:            func(payload interface{}) error { return nil },
		onLoadDevice:              func(id string) (interface{}, error) { return nil, nil },
		onLoadDeviceList:          func(ids []string) interface{} { return nil },
		onLoadDevices:             func() interface{} { return nil },
		onLoadAutomations:         func() interface{} { return nil },
		onLoadAppConfig:           func() (interface{}, error) { return nil, nil },
		onSaveDeviceConfig:        func(payload interface{}) error { return nil },
	}

	return wsHub
}

// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		http.ServeFile(w, r, "index.html")
// 	})

// 	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
// 		m.HandleRequest(w, r)
// 	})

// 	m.HandleMessage(func(s *melody.Session, msg []byte) {
// 		m.Broadcast(msg)
// 	})

// 	http.ListenAndServe(":5000", nil)
