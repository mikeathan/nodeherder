package repository

import (
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/utils"
	"node-herder/utils/storage"
	"path/filepath"
	"sort"
	"time"
)

const metricsBaseFilename = "metrics.db"
const metricsBucketName = "metrics"

type MetricsRepo struct {
	keyGenerator metrics.TimestampedKeyGenerator
	kvdb         storage.KeyValueDatabase
	clock        utils.Clock
}

func NewMetricsRepo() (metrics.Repository, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filepath.Join("data", metricsBaseFilename), metricsBucketName)
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

func GetDayRange(now time.Time, duration time.Duration) (time.Time, time.Time) {
	from := now.Truncate(24 * time.Hour)
	to := from.Add(24 * time.Hour)
	return from, to
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


todo caching 
func (s *MetricsRepo) ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {

	// sort exposekeys for result ordering
	exposekeys := make([]string, 0, len(device.Exposes))
	for k := range device.Exposes {
		exposekeys = append(exposekeys, k)
	}
	sort.Strings(exposekeys)

	result := metrics.NewDeviceMetricsResult(device.Id)

	// Prepare result collectors per expose
	collectors := make(map[string]metrics.ExposeResult, len(exposekeys))
	for _, k := range exposekeys {
		ex := device.Exposes[k]
		r, err := metrics.NewExposeResult(ex.Name, ex.Type, from, to)
		if err != nil {
			return nil, err
		}
		collectors[ex.Name] = r
	}

	fromKey := s.keyGenerator.CreateKeyPrefixFromTimestamp(from)
	toKey := s.keyGenerator.CreateMaxKeyFromTimestamp(to)

	callback := func(key, value []byte) error {
		// Route by expose name (id suffix)
		id, err := s.keyGenerator.GetIdFromKey(key)
		if err != nil {
			return err
		}
		collector, ok := collectors[id]
		if !ok {
			// not an expose we care about for this device
			return nil
		}

		timestamp, err := s.keyGenerator.GetTimestampFromkey(key)
		if err != nil {
			return err
		}
		if err := collector.Collect(timestamp, value); err != nil {
			return err
		}
		return nil
	}

	if err := s.kvdb.ViewInRange(device.Id, fromKey, toKey, callback); err != nil {
		return nil, err
	}

	for _, k := range exposekeys {
		c := collectors[k]
		if c.Size() == 0 {
			continue
		}
		c.Flush()
		result.Add(c)
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
