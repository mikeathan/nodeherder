package storage

type Storage[T any] interface {
	Store(id string, item *T) error
	LoadAll() []*T
	Initialize() ([]*T, error)
	ClearCache()
	Load(item string) (*T, error)
	Delete(item string) error
}
