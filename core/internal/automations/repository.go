package automations

type Repository interface {
	Store(id string, automation *Device)
	Find(id string) (*Device, error)
	Delete(id string) error
	Load() []*Device
}
