package logging

const (
	LoadAction = "load"
)

type FileLogRequest struct {
	File   string `json:"file"`
	Action string `json:"action"`
}

func NewFileLogRequest() *FileLogRequest {
	return &FileLogRequest{}
}

type LogMessage struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type RemoteHookEmitter interface {
	Broadcast(eventName string, data interface{}) error
}

// type RemoteHook struct {
// 	levels  []uint32
// 	emitter RemoteHookEmitter
// 	enabled bool
// }
