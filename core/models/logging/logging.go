package logging

type RemoteHookEmitter interface {
	Broadcast(eventName string, data interface{}) error
}

// type RemoteHook struct {
// 	levels  []uint32
// 	emitter RemoteHookEmitter
// 	enabled bool
// }
