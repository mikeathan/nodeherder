package storage

type Storage[T Storable] interface {
	Store(id string, item *T) error
	LoadAll() []*T
	Initialize() ([]*T, error)
	ClearCache()
	Load(item string) (*T, error)
	Delete(item string) error
}

type StorablePointer[T any] interface {
	*T
	Storable
}

type Storable interface {
	Initialize()
}
