package settings

type Task interface {
	Start(cfg *AppConfig) error
	Stop() error
}
