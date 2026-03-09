package storage

type Storage[T any] interface {
	Store(id string, item T) error
	LoadAll() []T
	Initialize() ([]T, error)
	ClearCache()
	LoadFromCache(item string) (T, error)
	Load(item string) (T, error)
	Delete(item string) error
}

type Storable interface {
	Initialize()
}
