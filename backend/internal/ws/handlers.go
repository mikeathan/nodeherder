package ws

import (
	"errors"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/utils"
	"reflect"
	"time"
)

type RequestContext struct {
	repo     *repository.MemoryRepo[hub.Request]
	handlers map[string]RequestHandler
}

func NewRequestContext() hub.Context {
	ctx := &RequestContext{
		repo:     repository.NewMemoryRepo[hub.Request](),
		handlers: map[string]RequestHandler{},
	}

	ctx.handlers[hub.BridgePermitJoin] = NewPermitJoinRequestHandler()

	return ctx
}

func (p *RequestContext) Enqueue(value hub.Request) {
	p.repo.Store(value.ID(), value)
}

func (p *RequestContext) Dequeue(key string) (hub.Request, error) {
	return p.repo.Dequeue(key)
}

func (p *RequestContext) Process(key string) error {
	req, err := p.Dequeue(key)
	if err != nil {
		return err
	}
	handler, ok := p.handlers[req.Type()]
	if !ok {
		// we dont have a handler for this request type. exit
		return nil
	}

	return handler.Process(req)
}

type RequestHandler interface {
	Process(hub.Request) error
}

// Handlers
type PermitJoinRequestHandler struct {
	activeStateTimer *utils.ActiveStateTimer
}

func NewPermitJoinRequestHandler() RequestHandler {
	return &PermitJoinRequestHandler{
		activeStateTimer: utils.NewActiveStateTimer(),
	}
}

func (p *PermitJoinRequestHandler) Process(request hub.Request) error {

	payload, ok := request.Payload().(*settings.BridgeConfig)
	if !ok {
		return errors.New("invalid payload type. expecting settings.BridgeConfig got " + reflect.TypeOf(request.Payload()).String())
	}

	if payload.PermitJoin {
		expireAt := time.Duration(payload.TimeExpireAt.Value) * time.Second
		return p.activeStateTimer.Start(request.Action(), expireAt)
	}

	return p.activeStateTimer.Stop(true)
}
