package hub

type Repository interface {
	Add(deviceName string, payload interface{})
	ListAll() interface{}
}

type MemoryRepository struct {
	store map[string]interface{}
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{store: map[string]interface{}{}}
}

func (s *MemoryRepository) Add(deviceName string, payload interface{}) {

	s.store[deviceName] = payload
}

func (s *MemoryRepository) ListAll() interface{} {
	return nil
}
