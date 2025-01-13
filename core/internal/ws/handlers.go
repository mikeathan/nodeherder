package ws

import (
	"node-herder/models/hub"
	"node-herder/repository"
	"node-herder/utils"
	"time"
)

// wIP
type Context struct {
	repo     *repository.MemoryRepo[hub.Request]
	handlers map[string]RequestHandler
}

func NewContext() *Context {
	ctx := &Context{
		repo:     repository.NewMemoryRepo[hub.Request](),
		handlers: map[string]RequestHandler{},
	}

	ctx.handlers["bridgepermitjoin"] = NewPermitJoinRequestHandler()

	return ctx
}

func (p *Context) Enqueue(value hub.Request) {
	p.repo.Store(value.ID(), value)
}

func (p *Context) Dequeue(key string) (hub.Request, error) {
	return p.repo.Dequeue(key)
}

func (p *Context) Process(key string) error {
	req, err := p.Dequeue(key)
	if err != nil {
		return err
	}
	handler, ok := p.handlers[req.Type()]
	if !ok {
		// we dont have a handler for this request type
		return nil
	}

	return handler.Process(req)
}

// hanlders
type RequestHandler interface {
	Process(hub.Request) error
}

type PermitJoinRequestHandler struct {
	activeStateTimer *utils.ActiveStateTimer
}

func NewPermitJoinRequestHandler() RequestHandler {
	return &PermitJoinRequestHandler{
		activeStateTimer: utils.NewActiveStateTimer(),
	}
}

func (p *PermitJoinRequestHandler) Process(value hub.Request) error {

	if active, ok := value.(bool); ok {
		if active {
			// TODO

			return p.activeStateTimer.Start(value.Action(), time.Duration(r.Time) * time.Second)
		}
		return p.activeStateTimer.Stop(true)
	}
	return nil
}
