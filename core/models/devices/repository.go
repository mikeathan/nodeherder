package devices

// this is going to be DeviceRepository
type Repository interface {
	AllDevices() []*Device
	Store(key string, device *Device)
	FindDevice(key string) (*Device, error)
	FindDevices(ids []string) []*Device
}

type MetricsRepository interface {
	Store(key string, device *Device) error
	Close()
}
