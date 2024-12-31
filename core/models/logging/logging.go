package logging

const (
	LoadAction = "load"
)

const (
	LogLevelDebug   string = "debug"
	LogLevelInfo    string = "info"
	LogLevelWarning string = "warning"
	LogLevelError   string = "error"
)

type FileLogRequest struct {
	File   string `json:"file"`
	Action string `json:"action"`
}

func NewFileLogRequest(file string, action string) *FileLogRequest {
	return &FileLogRequest{
		File:   file,
		Action: action,
	}
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
