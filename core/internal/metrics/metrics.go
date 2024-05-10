package metrics

import (
	"bytes"
	"encoding/binary"
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
func (s *MetricsRepo) Store(device *devices.Device) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte("metrics"))
		if err != nil {
			return err
		}

		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		bucket.NextSequence()

		buf, err := json.Marshal(device)
		if err != nil {
			return err
		}

		return bucket.Put(createKeyFromDevice(device), buf)
	})

	return nil
}

func (s *MetricsRepo) ViewRange(device *devices.Device, from time.Time, to time.Time) error {

	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("metrics")).Cursor()
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		fromKey := createKeyWithTimestamp(device, from)
		tokey := createKeyWithTimestamp(device, to)

		for k, v := bucket.Seek(fromKey); k != nil && bytes.Compare(k, tokey) <= 0; k, v = bucket.Next() {
			fmt.Println(string(k), string(v))

			// var point SensorData
			// 	if err := json.Unmarshal(v, &point); err != nil {
			// 		return err
			// 	}
			// 	data = append(data, point)
		}

		return nil
	})
}

func createKeyWithTimestamp(device *devices.Device, timestamp time.Time) []byte {

	buffer := bytes.NewBuffer(nil)
	binary.Write(buffer, binary.BigEndian, []byte(device.Id))
	binary.Write(buffer, binary.BigEndian, timestamp.UnixMilli())

	// Add a separator byte
	buffer.WriteByte(0)

	return buffer.Bytes()
}

func createKeyFromDevice(device *devices.Device) []byte {

	lastSeenStr, _ := device.Properties["last_seen"].(string)
	lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
	if err != nil {
		lastSeen = time.Now()
	}

	return createKeyWithTimestamp(device, lastSeen)
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
