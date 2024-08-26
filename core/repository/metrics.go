package repository

import (
	"bytes"
	"fmt"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/utils"
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

func NewMetricsRepoFromFile(filename string, keyGenerator metrics.TimestampedKeyGenerator) (metrics.Repository, error) {

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
	err = tx.Commit()
	if err != nil {
		return err
	}

	//s.runPruningTask()
	return nil
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

			buffer, err := utils.AnyToByteArray(value)
			if err != nil {
				return err
			}

			key := createKeyWithTimestamp(name, s.clock.Now())
			err = bucket.Put(key, buffer)

			if err != nil {
				return err
			}
		}
		return nil
	})

	utils.LogDebugf("Store metrics for device %v", id)

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

		if event.Size() == 0 {
			return nil
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

			if event.Size() == 0 {
				continue
			}

			result.Add(event)
		}

		return nil
	})

	return result, err
}

type timeRangeValue struct {
	Value     string
	Timestamp time.Time
}

func (s *MetricsRepo) readTimeRangeValues(cursor *bolt.Cursor, expose *devices.Entity, from time.Time, to time.Time) (*metrics.ExposeTimeRangeMetricsResult, error) {

	fromKey := createKeyWithTimestamp(expose.Name, from)
	tokey := createKeyWithTimestamp(expose.Name, to)

	var prevValue *timeRangeValue = nil

	events := metrics.NewExposeTimeRangeMetricResult(expose, from, to)

	for key, data := cursor.Seek(fromKey); key != nil && bytes.Compare(key, tokey) <= 0; key, data = cursor.Next() {
		timestamp, err := s.readTimestampFromKey(key)
		if err != nil {
			return nil, err
		}

		var value string
		err = utils.ByteArrayToAny(data, &value)
		if err != nil {
			return nil, err
		}

		if prevValue == nil {
			prevValue = &timeRangeValue{value, timestamp}
			continue
		}

		if prevValue.Value != value {

			events.Add(prevValue.Value, prevValue.Timestamp, timestamp)
			prevValue = &timeRangeValue{value, timestamp}
		}
	}

	if len(events.Data)%2 != 0 {
		events.Add(prevValue.Value, prevValue.Timestamp, to)
	}

	return events, nil
}

func (s *MetricsRepo) readNumericValues(cursor *bolt.Cursor, expose *devices.Entity, from time.Time, to time.Time) (*metrics.ExposeNumericMetricsResult, error) {

	fromKey := createKeyWithTimestamp(expose.Name, from)
	tokey := createKeyWithTimestamp(expose.Name, to)

	event := metrics.NewExposeNumericMetricResult(expose.Name, from, to)

	for key, data := cursor.Seek(fromKey); key != nil && bytes.Compare(key, tokey) <= 0; key, data = cursor.Next() {
		timestamp, err := s.readTimestampFromKey(key)
		if err != nil {
			return nil, err
		}

		var value float32
		err = utils.ByteArrayToAny(data, &value)
		if err != nil {
			return nil, err
		}

		truncated := utils.TruncateFloat32(value, 1)
		event.Add(truncated, timestamp)
	}

	return event, nil
}

func (s *MetricsRepo) findExposeTimeRangeEvent(cursor *bolt.Cursor, expose *devices.Entity, from time.Time, to time.Time) (metrics.ExposeResult, error) {

	if expose.Type == "numeric" {
		return s.readNumericValues(cursor, expose, from, to)
	} else if expose.Type == "binary" || expose.Type == "enum" {
		return s.readTimeRangeValues(cursor, expose, from, to)
	} else {
		return nil, fmt.Errorf("expose type %v not supported", expose.Type)
	}
}

func (s *MetricsRepo) runPruningTask() {
	go func() {
		for {

			// TODO:
			// maybe pass duration in configuration
			s.clock.Sleep(time.Minute) // Change it hour or day !!!!!!
			utils.LogInfof("Start pruning bucket %v", metricsBucketName)

			err := s.pruneEntries(s.db, metricsBucketName)
			if err != nil {
				utils.LogErrorf("Error pruning entries: %v", err)
			}
			utils.LogInfof("End pruning bucket %v", metricsBucketName)

		}
	}()
}
func (s *MetricsRepo) pruneEntries(db *bolt.DB, bucketName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket not found: %s", bucketName)
		}

		c := bucket.Cursor()
		for key, _ := c.First(); key != nil; key, _ = c.Next() {

			fmt.Println("device id", string(key))

			bucket.Bucket(key).ForEach(func(k, _ []byte) error {
				fmt.Println("device key", string(k))
				// voc2024-08-25T18:02:52.455198804Z
				return nil
			})

			// timestamp, err := s.readTimestampFromKey(expose.Name, key)
			// timestamp, err := s.readTimestampFromKey(expose.Name, key)

			// expiresAt, err := strconv.ParseInt(strings.Split(k, "-")[0], 10, 64)
			// if err != nil {
			// 	return fmt.Errorf("error parsing expiration timestamp: %w", err)
			// }

			// if time.Now().Unix() > expiresAt {
			// 	if err := bucket.Delete(k); err != nil {
			// 		return fmt.Errorf("error deleting expired entry: %w", err)
			// 	}
			// }
		}

		return nil
	})
}

func (s *MetricsRepo) readTimestampFromKey(data []byte) (time.Time, error) {

	timestamp, err := time.Parse("2006-01-02T15:04:05.000000000Z", string(data[:30]))
	if err != nil {
		return s.clock.Now(), err
	}
	return timestamp, nil
}

func createKeyWithTimestamp(id string, timestamp time.Time) []byte {

	customFormat := "2006-01-02T15:04:05.000000000Z"
	timestampStr := timestamp.Format(customFormat)
	key := fmt.Sprintf("%s_%s", timestampStr, id)

	return []byte(key)
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
