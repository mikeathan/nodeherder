package hub

import (
	"encoding/json"
	"errors"
	"node-herder/utils"
	"time"

	"github.com/google/uuid"
)

type Request interface {
	ID() string
	Process(value any) error
}

type BridgePermitJoinRequest struct {
	Value         bool   `json:"value"`
	Time          int    `json:"time"`
	TransactionId uint32 `json:"transaction"`
	timer         *utils.ActiveStateTimer
}

func NewBridgePermitJoinRequest(value bool, time int, handler func(bool) error) *BridgePermitJoinRequest {

	transaction := uuid.New()
	br := &BridgePermitJoinRequest{
		Value:         value,
		Time:          time,
		TransactionId: transaction.ID(),
		timer:         utils.NewActiveStateTimer(handler),
	}

	return br
}

func (r *BridgePermitJoinRequest) ToJson() ([]byte, error) {
	return json.Marshal(r)
}

func (r *BridgePermitJoinRequest) ID() string {
	return utils.ConvertInt32(r.TransactionId)
}

func (r *BridgePermitJoinRequest) Process(value any) error {

	if active, ok := value.(bool); ok {
		if active {
			return r.timer.Start(time.Duration(r.Time) * time.Second)
		}
		return r.timer.Stop(true)
	}

	return errors.New("invalid value type")
}
