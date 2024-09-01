package repository

import (
	"bytes"
	"fmt"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/utils"
	"node-herder/utils/storage"
	"sort"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const metricsBaseFilename = "metrics.db"
const metricsBucketName = "metrics"

type MetricsRepo struct {
	mutex        *sync.RWMutex
	db           *bolt.DB
	keyGenerator metrics.TimestampedKeyGenerator
	kvdb         storage.KeyValueDatabase
}

func NewMetricsRepoTEST(kvdb storage.KeyValueDatabase, keyGenerator metrics.TimestampedKeyGenerator) (metrics.Repository, error) {
	repo := &MetricsRepo{
		kvdb:         kvdb,
		keyGenerator: keyGenerator,
	}
	return repo, nil
}

func NewMetricsRepo(keyGenerator metrics.TimestampedKeyGenerator) (metrics.Repository, error) {

	return NewMetricsRepoFromFile(metricsBaseFilename, keyGenerator)
}

func NewMetricsRepoFromFile(filename string, keyGenerator metrics.TimestampedKeyGenerator) (metrics.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}
	// TODO: ideally pass device configs to configure the rate limiter timeout
	repo := &MetricsRepo{
		mutex:        &sync.RWMutex{},
		db:           db,
		keyGenerator: keyGenerator,
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

	callback := func(key string, value any) ([]byte, []byte, error) {
		buffer, err := utils.AnyToByteArray(value)
		if err != nil {
			return nil, nil, err
		}

		return s.keyGenerator.CreateKey(key), buffer, nil
	}

	return s.kvdb.SetBatch(id, data, callback)
}

func (s *MetricsRepo) StoreORIG(id string, data map[string]any) error {

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

			key := s.keyGenerator.CreateKey(name)
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
		event, err := s.collectEvents(cursor, expose, from, to)
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

	// sort exposekeys
	exposekeys := make([]string, 0, len(device.Exposes))
	for k := range device.Exposes {
		exposekeys = append(exposekeys, k)
	}

	sort.Strings(exposekeys)

	result := metrics.NewDeviceMetricsResult(device.Id)

	for _, key := range exposekeys {
		expose := device.Exposes[key]

		fromkey := s.keyGenerator.CreateKeyFromTimestamp(expose.Name, from)
		toKey := s.keyGenerator.CreateKeyFromTimestamp(expose.Name, to)

		event, err := metrics.NewExposeResult(expose.Name, expose.Type, from, to)
		if err != nil {
			return nil, err
		}

		callback := func(key, value []byte) error {
			if !bytes.HasSuffix(key, []byte(expose.Name)) {
				return nil
			}

			timestamp, err := s.keyGenerator.GetTimestampFromkey(key)
			if err != nil {
				return err
			}

			err = event.Collect(timestamp, value)
			if err != nil {
				return err
			}

			return nil
		}

		// NOTE:
		// For now we do a db call for each expose. Needs to be optimized

		err = s.kvdb.ViewInRange(device.Id, fromkey, toKey, callback)
		if err != nil {
			return nil, err
		}

		if event.Size() == 0 {
			continue
		}

		event.Flush()
		result.Add(event)
	}

	return result, nil
}

func (s *MetricsRepo) ViewDeviceTimeRangeORIG(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {

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
			event, err := s.collectEvents(cursor, expose, from, to)
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

func (s *MetricsRepo) collectEvents(cursor *bolt.Cursor, expose *devices.Entity, from time.Time, to time.Time) (metrics.ExposeResult, error) {

	result, err := metrics.NewExposeResult(expose.Name, expose.Type, from, to)
	if err != nil {
		return nil, err
	}

	fromKey := s.keyGenerator.CreateKeyFromTimestamp(expose.Name, from)
	tokey := s.keyGenerator.CreateKeyFromTimestamp(expose.Name, to)

	for key, data := cursor.Seek(fromKey); key != nil && bytes.Compare(key, tokey) <= 0; key, data = cursor.Next() {
		if !bytes.HasSuffix(key, []byte(expose.Name)) {
			continue
		}

		timestamp, err := s.keyGenerator.GetTimestampFromkey(key)
		if err != nil {
			return nil, err
		}

		err = result.Collect(timestamp, data)
		if err != nil {
			return nil, err
		}
	}

	result.Flush()
	return result, nil
}

// TEST WIP ===================
// ////////////////////////////////
// type PruningService struct {
// 	keyGenerator metrics.TimestampedKeyGenerator
// }

// func NewPruningService(keyGenerator metrics.TimestampedKeyGenerator) *PruningService {
// 	return &PruningService{
// 		keyGenerator: keyGenerator,
// 	}
// }

// func (p *PruningService) Run(db storage.KeyValueDatabase, bucketName string, duration time.Duration) error {

// 	return db.Update(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(bucketName))
// 		if bucket == nil {
// 			return fmt.Errorf("bucket not found: %s", bucketName)
// 		}

// 		c := bucket.Cursor()
// 		for key, _ := c.First(); key != nil; key, _ = c.Next() {
// 			timestamp, err := p.keyGenerator.GetTimestampFromkey(key)

// 			//expiresAt, err := strconv.ParseInt(strings.Split(k, "-")[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("error parsing expiration timestamp: %w", err)
// 			}

// 			if time.Now().Unix() > timestamp.Unix() {
// 				if err := bucket.Delete(key); err != nil {
// 					return fmt.Errorf("error deleting expired entry: %w", err)
// 				}
// 			}
// 		}

// 		return nil
// 	})

// }

// func (s *MetricsRepo) runPruningTask() {
// 	go func() {
// 		for {

// 			// TODO:
// 			// maybe pass duration in configuration
// 			s.clock.Sleep(time.Minute) // Change it hour or day !!!!!!
// 			utils.LogInfof("Start pruning bucket %v", metricsBucketName)

// 			err := s.pruneEntries(s.db, metricsBucketName)
// 			if err != nil {
// 				utils.LogErrorf("Error pruning entries: %v", err)
// 			}
// 			utils.LogInfof("End pruning bucket %v", metricsBucketName)

// 		}
// 	}()
// }
// func (s *MetricsRepo) pruneEntries(db *bolt.DB, bucketName string) error {
// 	return db.Update(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(bucketName))
// 		if bucket == nil {
// 			return fmt.Errorf("bucket not found: %s", bucketName)
// 		}

// 		c := bucket.Cursor()
// 		for key, _ := c.First(); key != nil; key, _ = c.Next() {

// 			fmt.Println("device id", string(key))

// 			bucket.Bucket(key).ForEach(func(k, _ []byte) error {
// 				fmt.Println("device key", string(k))
// 				// voc2024-08-25T18:02:52.455198804Z
// 				return nil
// 			})

// 			// timestamp, err := s.readTimestampFromKey(expose.Name, key)
// 			// timestamp, err := s.readTimestampFromKey(expose.Name, key)

// 			// expiresAt, err := strconv.ParseInt(strings.Split(k, "-")[0], 10, 64)
// 			// if err != nil {
// 			// 	return fmt.Errorf("error parsing expiration timestamp: %w", err)
// 			// }

// 			// if time.Now().Unix() > expiresAt {
// 			// 	if err := bucket.Delete(k); err != nil {
// 			// 		return fmt.Errorf("error deleting expired entry: %w", err)
// 			// 	}
// 			// }
// 		}

// 		return nil
// 	})
// }

// func (s *MetricsRepo) readTimestampFromKey(data []byte) (time.Time, error) {

// 	timestamp, err := time.Parse("2006-01-02T15:04:05.000000000Z", string(data[:30]))
// 	if err != nil {
// 		return s.clock.Now(), err
// 	}
// 	return timestamp, nil
// }

// func createKeyWithTimestamp(id string, timestamp time.Time) []byte {

// 	customFormat := "2006-01-02T15:04:05.000000000Z"
// 	timestampStr := timestamp.Format(customFormat)
// 	key := fmt.Sprintf("%s_%s", timestampStr, id)

// 	return []byte(key)
// }

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
