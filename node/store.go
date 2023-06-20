package node

type Store interface {
	Add(deviceName string, payload interface{})
	ListAll() interface{}
}

type Z2MStore struct {
	store map[string]interface{}
}

func NewZ2MStore() Store {
	return &Z2MStore{store: map[string]interface{}{}}
}

func (s *Z2MStore) Add(deviceName string, payload interface{}) {

	s.store[deviceName] = payload
}

func (s *Z2MStore) ListAll() interface{} {
	return nil
}
