package services

import "sync"

type HubRequest interface {
	ID() string
	Process(value any) error
}

type HubRequestQueue struct {
	requests map[string]HubRequest
	mutex    sync.Mutex
}

func (q *HubRequestQueue) Add(req HubRequest) {
	q.mutex.Lock()
	q.requests[req.ID()] = req
	q.mutex.Unlock()
}

func (q *HubRequestQueue) Get(id string) HubRequest {
	q.mutex.Lock()
	req := q.requests[id]
	q.mutex.Unlock()
	return req
}

func (q *HubRequestQueue) Delete(id string) {
	q.mutex.Lock()
	delete(q.requests, id)
	q.mutex.Unlock()
}

func NewHubRequestQueue() *HubRequestQueue {
	return &HubRequestQueue{
		requests: make(map[string]HubRequest),
	}
}
