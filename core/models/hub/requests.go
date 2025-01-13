package hub

import (
	"node-herder/utils"

	"github.com/google/uuid"
)

const (
	BridgePermitJoin = "bridgePermitJoin"
)

type Request interface {
	ID() string
	Payload() interface{}
	Type() string
	Action() func(bool) error
}

type BridgePermitJoinRequest struct {
	Value         bool   `json:"value"`
	Time          int    `json:"time"`
	TransactionId uint32 `json:"transaction"`
	callback      func(bool) error
}

func NewBridgePermitJoinRequest(value bool, time int, handler func(bool) error) *BridgePermitJoinRequest {
	return &BridgePermitJoinRequest{
		Value:         value,
		Time:          time,
		TransactionId: uuid.New().ID(),
		callback:      handler,
	}
}

func (r *BridgePermitJoinRequest) Type() interface{} {
	return BridgePermitJoin
}

func (r *BridgePermitJoinRequest) Action() func(bool) error {
	return r.callback
}

func (r *BridgePermitJoinRequest) Payload() interface{} {
	return r.Value
}

func (r *BridgePermitJoinRequest) ID() string {
	return utils.ConvertInt32(r.TransactionId)
}
