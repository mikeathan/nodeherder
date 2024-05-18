package metrics

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"node-herder/models/devices"
	"node-herder/utils"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const baseFilename = "metrics.db"

type MetricsRepo struct {
	store map[string]*devices.Device
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewMetricsRepo() (devices.MetricsRepository, error) {
	return NewMetricsRepoFromFile(baseFilename)
}

func NewMetricsRepoFromFile(filename string) (devices.MetricsRepository, error) {

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

func (s *MetricsRepo) Store(device *devices.Device) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte("metrics"))
		if err != nil {
			return err
		}

		bucket, err = bucket.CreateBucketIfNotExists([]byte(device.Id))
		if err != nil {
			return err
		}

		for _, expose := range device.Exposes {
			buf, err := json.Marshal(expose.Data)
			if err != nil {
				return err
			}

			key := createKeyFromDevice(expose.Name, device)

			err = bucket.Put(key, buf)
			if err != nil {
				return err
			}
		}

		buf, err := json.Marshal(device)
		if err != nil {
			return err
		}

		key := createKeyFromDevice(device.Id, device)
		return bucket.Put(key, buf)
	})

	return nil
}

func (s *MetricsRepo) ViewExposeTimeRange(deviceId string, exposeName string, from time.Time, to time.Time) (*devices.DeviceMetricsResult, error) {

	result := devices.NewDeviceMetricsResult(deviceId)

	err := s.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("metrics")).Bucket([]byte(deviceId)).Cursor()
		if cursor == nil {
			return bolt.ErrBucketNotFound
		}
		event := devices.NewExposeMetricsResult(exposeName)

		fromKey := createKeyWithTimestamp(exposeName, from)
		tokey := createKeyWithTimestamp(exposeName, to)

		for k, value := cursor.Seek(fromKey); k != nil && bytes.Compare(k, tokey) <= 0; k, value = cursor.Next() {
			timestamp, err := readTimestampFromKey(exposeName, k)
			if err != nil {
				return err
			}

			wrong here
			event.Add(value, &timestamp)

			//fmt.Printf("propert %v data: %v  timestamp : %v \n", exposeName, string(value), timestamp)
		}
		result.Add(event)
		return nil
	})

	return result, err
}

func (s *MetricsRepo) ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*devices.DeviceMetricsResult, error) {

	result := devices.NewDeviceMetricsResult(device.Id)

	err := s.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("metrics")).Bucket([]byte(device.Id)).Cursor()
		if cursor == nil {
			return bolt.ErrBucketNotFound
		}

		for _, expose := range device.Exposes {

			fromKey := createKeyWithTimestamp(expose.Name, from)
			tokey := createKeyWithTimestamp(expose.Name, to)

			event := devices.NewExposeMetricsResult(expose.Name)

			for key, value := cursor.Seek(fromKey); key != nil && bytes.Compare(key, tokey) <= 0; key, value = cursor.Next() {
				timestamp, err := readTimestampFromKey(expose.Name, key)
				if err != nil {
					return err
				}

				event.Add(value, &timestamp)

				//fmt.Printf("propert %v data: %v  timestamp : %v \n", expose.Name, string(value), timestamp)
			}

			result.Add(event)
		}

		return nil
	})

	return result, err
}

func readTimestampFromKey(id string, data []byte) (time.Time, error) {
	timestamp, err := time.Parse(time.RFC3339, string(data[len(id):]))
	if err != nil {
		return time.Now(), err

	}
	return timestamp, nil
}

func createKeyWithTimestamp(id string, timestamp time.Time) []byte {
	buffer := bytes.NewBuffer(nil)
	binary.Write(buffer, binary.BigEndian, []byte(id))

	timestampStr := timestamp.Format(time.RFC3339)
	binary.Write(buffer, binary.BigEndian, []byte(timestampStr))

	return buffer.Bytes()
}

func createKeyFromDevice(id string, device *devices.Device) []byte {
	lastSeenStr, _ := device.Properties["last_seen"].(string)
	lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
	if err != nil {
		lastSeen = time.Now()
	}

	return createKeyWithTimestamp(id, lastSeen)
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
