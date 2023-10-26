package devices

type Repository interface {
	AllDevices() []*Device
	Store(key string, device *Device)
	FindDevice(key string) (*Device, error)
}
