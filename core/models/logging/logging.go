package logging

import "time"

type EnableRemoteLoggerRequest struct {
	Enable bool `json:"enable"`
}

type LogMessage struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type RemoteHookEmitter interface {
	Broadcast(eventName string, data interface{}) error
}

// type RemoteHook struct {
// 	levels  []uint32
// 	emitter RemoteHookEmitter
// 	enabled bool
// }
