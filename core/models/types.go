package models

type Device struct {
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

type EventEmitter interface {
	Broadcast(eventName string, data interface{}) error
}
