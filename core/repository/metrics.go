package repository

import (
	"bytes"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/utils"
	"node-herder/utils/storage"
	"sort"
	"time"
)

const metricsBaseFilename = "metrics.db"
const metricsBucketName = "metrics"

type MetricsRepo struct {
	keyGenerator metrics.TimestampedKeyGenerator
	kvdb         storage.KeyValueDatabase
}

func NewMetricsRepo() (metrics.Repository, error) {
	keyGenerator := metrics.NewTimestampedKeyGenerator(utils.NewRealClock())
	kvdb, err := storage.NewBoltKeyValueDatabase(metricsBaseFilename, metricsBucketName)
	if err != nil {
		return nil, err
	}
	return NewMetricsRepoFromDatabase(kvdb, keyGenerator)
}

func NewMetricsRepoFromDatabase(kvdb storage.KeyValueDatabase, keyGenerator metrics.TimestampedKeyGenerator) (metrics.Repository, error) {
	repo := &MetricsRepo{
		kvdb:         kvdb,
		keyGenerator: keyGenerator,
	}
	return repo, nil
}

func (s *MetricsRepo) Close() error {
	err := s.kvdb.Close()
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

//			}
//		}()
//	}

func (s *MetricsRepo) Prune(expireAt time.Time) error {

	callback := func(key []byte) (bool, error) {
		timestamp, err := s.keyGenerator.GetTimestampFromkey(key)
		if err != nil {
			return false, err
		}

		return timestamp.Before(expireAt), nil
	}

	return s.kvdb.Prune(callback)
}

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
