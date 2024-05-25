package settings

type Repository interface {
	Store(key string, value any) error
	Get(key string) (any, error)
	Close() error
}
