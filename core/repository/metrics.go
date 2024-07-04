package repository

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/utils"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const metricsBaseFilename = "metrics.db"
const metricsBucketName = "metrics"

type MetricsRepo struct {
	mutex *sync.RWMutex
	db    *bolt.DB
	clock utils.Clock
}

func NewMetricsRepo() (metrics.Repository, error) {
	return NewMetricsRepoFromFile(metricsBaseFilename, &utils.RealClock{})
}

func NewMetricsRepoFromFile(filename string, clock utils.Clock) (metrics.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}
	// TODO: ideally pass device configs to configure the rate limiter timeout
	repo := &MetricsRepo{
		mutex: &sync.RWMutex{},
		db:    db,
		clock: clock,
	}
	err = repo.init()
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	return repo, nil
}

func (s *MetricsRepo) init() error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists([]byte(metricsBucketName)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *MetricsRepo) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *MetricsRepo) Store(id string, data map[string]any) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(metricsBucketName))
		if err != nil {
			return err
		}

		bucket, err = bucket.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return err
		}

		for name, value := range data {
			buf, err := json.Marshal(value)
			if err != nil {
				return err
			}

			key := createKeyWithTimestamp(name, s.clock.Now())
			err = bucket.Put(key, buf)

			if err != nil {
				fmt.Printf("DEBUG -  metrics: Expose=%v, Data=%v, Key=%v ERROR=%v\n", name, string(buf), string(key), err.Error())
				return err
			}

			fmt.Printf("DEBUG -  metrics: Expose=%v, Data=%v, Key=%v \n", name, string(buf), string(key))
		}
		return nil
		// // ??????
		// buf, err := json.Marshal(device)
		// if err != nil {
		// 	return err
		// }

		// key := createKeyFromDevice(device.Id, device)
		// //fmt.Printf("DEBUG store device metrics for: %v with key: %v \n", device.Id, string(key))
		// return bucket.Put(key, buf)
	})

	return nil
}

func (s *MetricsRepo) ViewExposeTimeRange(device *devices.Device, exposeName string, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {

	result := metrics.NewDeviceMetricsResult(device.Id)

	err := s.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte(metricsBucketName)).Bucket([]byte(device.Id)).Cursor()
		if cursor == nil {
			return bolt.ErrBucketNotFound
		}

		expose := device.Exposes[exposeName]
		event, err := s.findExposeTimeRangeEvent(cursor, expose, from, to)
		if err != nil {
			return err
		}
		result.Add(event)
		return nil
	})

	return result, err
}

func (s *MetricsRepo) ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {

	result := metrics.NewDeviceMetricsResult(device.Id)

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(metricsBucketName)).Bucket([]byte(device.Id))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}
		cursor := bucket.Cursor()
		if cursor == nil {
			return fmt.Errorf("bucket cursor not found")
		}

		// sort exposekeys
		exposekeys := make([]string, 0, len(device.Exposes))
		for k := range device.Exposes {
			exposekeys = append(exposekeys, k)
		}

		sort.Strings(exposekeys)

		for _, key := range exposekeys {
			expose := device.Exposes[key]
			event, err := s.findExposeTimeRangeEvent(cursor, expose, from, to)
			if err != nil {
				return err
			}
			result.Add(event)
		}

		return nil
	})

	return result, err
}

func (s *MetricsRepo) findExposeTimeRangeEvent(cursor *bolt.Cursor, expose *devices.Entity, from time.Time, to time.Time) (*metrics.ExposeMetricsResult, error) {

	exposeType := kindFromExposeType(expose)
	fromKey := createKeyWithTimestamp(expose.Name, from)
	tokey := createKeyWithTimestamp(expose.Name, to)

	event := metrics.NewExposeMetricsResult(expose.Name, exposeType.String())
	for key, value := cursor.Seek(fromKey); key != nil && bytes.Compare(key, tokey) <= 0; key, value = cursor.Next() {
		timestamp, err := s.readTimestampFromKey(expose.Name, key)
		if err != nil {
			return nil, err
		}

		data, err := utils.Unmarshal(value, exposeType)
		if err != nil {
			return nil, err
		}

		event.Add(data, &timestamp)
	}

	return event, nil
}

func kindFromExposeType(expose *devices.Entity) reflect.Kind {
	switch expose.Type {
	case "numeric":
		return reflect.Float32
	case "binary":
		return reflect.String
	case "enum":
		return reflect.Int
	}

	return reflect.Interface
}

func (s *MetricsRepo) readTimestampFromKey(id string, data []byte) (time.Time, error) {
	timestamp, err := time.Parse(time.RFC3339Nano, string(data[len(id):]))
	if err != nil {
		return s.clock.Now(), err

	}
	return timestamp, nil
}

func createKeyWithTimestamp(id string, timestamp time.Time) []byte {
	buffer := bytes.NewBuffer(nil)
	binary.Write(buffer, binary.BigEndian, []byte(id))

	timestampStr := timestamp.Format(time.RFC3339Nano)
	binary.Write(buffer, binary.BigEndian, []byte(timestampStr))

	return buffer.Bytes()
}

// // Paginate entries
// pageSize := 10
// pageNumber := 1

// err = db.View(func(tx *bolt.Tx) error {
// 	bucket := tx.Bucket([]byte("entries"))
// 	if bucket == nil {
// 		return fmt.Errorf("Bucket not found")
// 	}

// 	// Start pagination from the specified page number
// 	cursor := bucket.Cursor()

// 	// Calculate the offset to start pagination from
// 	offset := (pageNumber - 1) * pageSize

// 	// Iterate over keys
// 	count := 0
// 	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
// 		// Skip until the offset is reached
// 		if count < offset {
// 			count++
// 			continue
// 		}

// 		// Print key and value
// 		fmt.Printf("Key: %s, Value: %s\n", k, v)

// 		// Break the loop when pageSize entries are printed
// 		if count >= offset+pageSize {
// 			break
// 		}

// 		count++
// 	}

// 	return nil
// })
// if err != nil {
// 	log.Fatal(err)
// }
// }
