package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"node-herder/models/devices"
	"node-herder/utils"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const filename = "metrics.db"

type MetricsRepo struct {
	store map[string]*devices.Device
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewMetricsRepo() (devices.MetricsRepository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	return &MetricsRepo{
		store: map[string]*devices.Device{},
		mutex: sync.RWMutex{},
		db:    db,
	}, nil
}

func (s *MetricsRepo) Close() {
	err := s.db.Close()
	if err != nil {
		utils.LogError(err)
	}
}

// todo
// need new repository interface or match current ?
// new struct for metrics
// timestamp
// key
// data
func (s *MetricsRepo) Store(key string, device *devices.Device) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte("metrics"))
		if err != nil {
			return err
		}

		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := bucket.NextSequence()

		// Marshal user data into bytes.
		buf, err := json.Marshal(device)
		if err != nil {
			return err
		}

	 store using device.Id  and timestamp
		// WIP - refactor
		lastSeenStr, _ := device.Properties["last_seen"].(string)
		lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
		if err != nil {
			return err
		}
		//
		return bucket.Put([]byte(lastSeen.Format(time.RFC3339)), buf)
	})

	return nil
}

func (s *MetricsRepo) ViewRange(device *devices.Device, from time.Time, to time.Time) error {

	device.Id 
	return s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("metrics")).Cursor()
		min := []byte(from.Format(time.RFC3339))
		max := []byte(to.AddDate(0, 0, 0).Format(time.RFC3339))

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			fmt.Println(string(k), string(v))
		}

		return nil
	})
}

func (s *MetricsRepo) ViewDevice(from time.Time, to time.Time) error {
	return nil
}
