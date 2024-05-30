package repository

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/utils"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const metricsBaseFilename = "metrics.db"

type RateLimiter struct {
	mutex     sync.Mutex
	rateLimit time.Duration
	lastWrite time.Time
	appConfig *settings.AppConfig
	store     map[string]time.Time // In-memory store for device IDs and last write times
}

func NewRateLimiter(appConfig *settings.AppConfig) *RateLimiter {
	return &RateLimiter{
		mutex:     sync.Mutex{},
		rateLimit: time.Minute,
		lastWrite: time.Time{},
		appConfig: appConfig,
		store:     map[string]time.Time{},
	}
}

func (rl *RateLimiter) AllowWrite(id string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	currentTime := time.Now()

	if rateLimit, ok := rl.appConfig.Devices[id].RateLimit; ok {

	}

	// Check if device exists in the in-memory store
	if lastWrite, ok := rl.store[id]; ok {
		if currentTime.Sub(lastWrite) < rl.rateLimit {
			return false // Rate limit exceeded
		}
	}

	// Update lastWrite time and store in map
	rl.lastWrite = currentTime
	rl.store[id] = currentTime

	return true
}

type MetricsRepo struct {
	mutex       *sync.RWMutex
	db          *bolt.DB
	rateLimiter *RateLimiter
}

func NewMetricsRepo(appConfig *settings.AppConfig) (metrics.Repository, error) {
	return NewMetricsRepoFromFile(metricsBaseFilename, appConfig)
}

func NewMetricsRepoFromFile(filename string, appConfig *settings.AppConfig) (metrics.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	// TODO: ideally pass device configs to configure the rate limiter timeout
	return &MetricsRepo{
		mutex:       &sync.RWMutex{},
		db:          db,
		rateLimiter: NewRateLimiter(appConfig),
	}, nil
}

func (s *MetricsRepo) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *MetricsRepo) Store(device *devices.Device) error {

	if !s.rateLimiter.AllowWrite(device.Id) {
		return nil
	}

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

func (s *MetricsRepo) ViewExposeTimeRange(device *devices.Device, exposeName string, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {

	result := metrics.NewDeviceMetricsResult(device.Id)

	err := s.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("metrics")).Bucket([]byte(device.Id)).Cursor()
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
		cursor := tx.Bucket([]byte("metrics")).Bucket([]byte(device.Id)).Cursor()
		if cursor == nil {
			return bolt.ErrBucketNotFound
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
		timestamp, err := readTimestampFromKey(expose.Name, key)
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
