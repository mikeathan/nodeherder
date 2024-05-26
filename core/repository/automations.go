package repository

import "node-herder/models/devices"

type Repository interface {
	Store(id string, automation *devices.Device)
	Find(id string) (*devices.Device, error)
	Delete(id string) error
	Load() []*devices.Device
}
