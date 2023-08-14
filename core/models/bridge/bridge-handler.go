package bridge

type Handler interface {
	ProcessMessage(payload []byte) error
}

type BridgeHandler struct {
	// map of registered triggers
}

func (h *BridgeHandler) ProcessMessage(payload []byte) error {

	return nil
}
