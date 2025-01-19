package hub

import (
	"node-herder/models/settings"

	"github.com/google/uuid"
)

const (
	BridgePermitJoin = "bridgePermitJoin"
)

type Context interface {
	Enqueue(value Request)
	Dequeue(key string) (Request, error)
	Process(key string) error
}

type Request interface {
	ID() string
	Payload() interface{}
	Type() string
	Action() func(bool) error
}

type BridgePermitJoinRequest struct {
	Value         bool   `json:"value"`
	Time          int    `json:"time"`
	TransactionId string `json:"transaction"`
	callback      func(bool) error
	payload       *settings.BridgeConfig
}

func NewBridgePermitJoinRequest(config *settings.BridgeConfig, handler func(bool) error) *BridgePermitJoinRequest {
	return &BridgePermitJoinRequest{
		payload:       config,
		Value:         config.PermitJoin,
		Time:          config.TimeExpireAt.Value,
		TransactionId: uuid.New().String(),
		callback:      handler,
	}
}
func (r *BridgePermitJoinRequest) Payload() interface{} {
	return r.payload
}

func (r *BridgePermitJoinRequest) Type() string {
	return BridgePermitJoin
}

func (r *BridgePermitJoinRequest) Action() func(bool) error {
	return r.callback
}

func (r *BridgePermitJoinRequest) ID() string {
	return r.TransactionId
}
