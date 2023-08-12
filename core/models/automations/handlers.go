package automations

type Handler interface {
	ProcessMessage(payload []byte) error
}

type BridgeHandler struct {
}

func (h *BridgeHandler) ProcessMessage(payload []byte) error {

	return nil
}
