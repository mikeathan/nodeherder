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
	clock        utils.Clock
	tasks        []metrics.Task
}

func NewMetricsRepo() (metrics.Repository, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(metricsBaseFilename, metricsBucketName)
	if err != nil {
		return nil, err
	}
	return NewMetricsRepoFromDatabase(kvdb, utils.NewRealClock())
}

func NewMetricsRepoFromDatabase(kvdb storage.KeyValueDatabase, clock utils.Clock) (metrics.Repository, error) {
	repo := &MetricsRepo{
		keyGenerator: metrics.NewTimestampedKeyGenerator(clock),
		kvdb:         kvdb,
		clock:        clock,
		tasks:        []metrics.Task{},
	}
	return repo, nil
}

func (s *MetricsRepo) AddTask(task metrics.Task) {
	s.tasks = append(s.tasks, task)
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

func (s *MetricsRepo) Prune(expireAt time.Duration) error {

	currentTime := s.clock.Now()
	callback := func(key []byte) (bool, error) {

		timestamp, err := s.keyGenerator.GetTimestampFromkey(key)
		if err != nil {
			return false, err
		}

		return currentTime.Sub(timestamp) > expireAt, nil
	}

	return s.kvdb.Prune(callback)
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
