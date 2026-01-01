package storage

import (
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/models/devices"
	"node-herder/utils"
	"node-herder/utils/storage"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const metricsBaseFilename = "metrics.db"
const metricsBucketName = "metrics"

type MetricsRepo struct {
	keyGenerator domain.TimestampedKeyGenerator
	kvdb         storage.KeyValueDatabase
	clock        utils.Clock

	// cache for tailing recent metrics
	tailWindow    time.Duration
	tailCache     map[string][]domain.CachedEntry
	tailCacheLock sync.RWMutex
}

func NewMetricsRepo() (domain.Repository, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filepath.Join("data", metricsBaseFilename), metricsBucketName)
	if err != nil {
		return nil, err
	}
	return NewMetricsRepoFromDatabase(kvdb, utils.NewRealClock(), 90*time.Second)
}

func NewMetricsRepoFromDatabase(kvdb storage.KeyValueDatabase, clock utils.Clock, tailWindow time.Duration) (domain.Repository, error) {
	repo := &MetricsRepo{
		keyGenerator:  domain.NewTimestampedKeyGenerator(clock),
		kvdb:          kvdb,
		clock:         clock,
		tailWindow:    tailWindow,
		tailCache:     make(map[string][]domain.CachedEntry),
		tailCacheLock: sync.RWMutex{},
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

	callback := func(exposeName string, value any) ([]byte, []byte, error) {
		buffer, err := utils.AnyToByteArray(value)
		if err != nil {
			return nil, nil, err
		}

		return s.keyGenerator.CreateKey(exposeName), buffer, nil
	}

	err := s.kvdb.SetBatch(id, data, callback)
	if err != nil {
		return err
	}

	// cache metrics in tail cache
	s.updateTailCache(id, data)
	return nil
}

func (s *MetricsRepo) ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*domain.DeviceMetricsResult, error) {

	collectors, err := s.prepareExposeCollectors(device, from, to)
	if err != nil {
		return nil, err
	}

	if err := s.execute(device.Id, from, to, nil, collectors); err != nil {
		return nil, err
	}

	result := s.finalizeCollectors(collectors, device.Id)
	return result, nil
}

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

func (s *MetricsRepo) QueryDevice(deviceID string,
	from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error) {

	if err := s.execute(deviceID, from, to, filters, collectors); err != nil {
		return nil, err
	}

	result := s.finalizeCollectors(collectors, deviceID)
	return result, nil
}

func (s *MetricsRepo) execute(deviceID string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) error {

	//  Decide the DB range
	var dbFrom, dbTo time.Time
	var tailFrom time.Time
	useTailCache := s.tailWindow > 0

	if useTailCache {
		dbFrom, dbTo, tailFrom = s.computeDatabaseRange(from, to)
	} else {
		dbFrom = from
		dbTo = to
	}

	// DB scan
	if err := s.queryDatabase(
		deviceID,
		dbFrom,
		dbTo, filters, collectors); err != nil {
		return err
	}

	// tail merge
	if useTailCache {
		if err := s.mergeTailCache(
			deviceID,
			tailFrom,
			to,
			filters,
			collectors,
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *MetricsRepo) updateTailCache(deviceID string, data map[string]any) {

	if s.tailWindow == 0 {
		return
	}

	now := s.clock.Now()

	s.tailCacheLock.Lock()
	defer s.tailCacheLock.Unlock()

	entries := s.tailCache[deviceID]

	// append new entries
	for exposeName, val := range data {
		buf, err := utils.AnyToByteArray(val)
		if err != nil {
			continue
		}

		entries = append(entries, domain.CachedEntry{
			ExposeName: exposeName,
			Timestamp:  now,
			Value:      buf,
		})
	}

	// prune old
	cutoff := now.Add(-s.tailWindow)
	idx := 0
	for ; idx < len(entries); idx++ {
		if entries[idx].Timestamp.After(cutoff) {
			break
		}
	}
	if idx > 0 {
		entries = entries[idx:]
	}

	s.tailCache[deviceID] = entries
}

func (s *MetricsRepo) prepareExposeCollectors(device *devices.Device, from, to time.Time) (map[string]domain.ExposeResult, error) {
	exposeNames := make([]string, 0, len(device.Exposes))
	for k := range device.Exposes {
		exposeNames = append(exposeNames, k)
	}
	sort.Strings(exposeNames)

	collectors := make(map[string]domain.ExposeResult, len(exposeNames))
	for _, name := range exposeNames {
		ex := device.Exposes[name]
		r, err := domain.NewExposeResult(ex.Name, ex.Type, domain.AggNone, from, to)
		if err != nil {
			// unknown expose type, skip
			continue
		}
		collectors[ex.Name] = r
	}

	return collectors, nil
}

func (s *MetricsRepo) finalizeCollectors(collectors map[string]domain.ExposeResult, deviceID string) *domain.DeviceMetricsResult {
	result := domain.NewDeviceMetricsResult(deviceID)

	for _, c := range collectors {
		if c.Size() == 0 {
			continue
		}
		c.Flush()
		result.Add(c)
	}

	return result
}

func (s *MetricsRepo) computeDatabaseRange(from, to time.Time) (time.Time, time.Time, time.Time) {
	now := s.clock.Now()
	tailStart := now.Add(-s.tailWindow)

	dbFrom := from
	dbTo := to

	if to.After(tailStart) {
		dbTo = tailStart
	}

	return dbFrom, dbTo, tailStart
}

func (s *MetricsRepo) queryDatabase(deviceID string,
	dbFrom, dbTo time.Time,
	filters []domain.MetricFilter,
	collectors map[string]domain.ExposeResult) error {

	if !dbTo.After(dbFrom) {
		return nil
	}

	fromKey := s.keyGenerator.CreateKeyPrefixFromTimestamp(dbFrom)
	toKey := s.keyGenerator.CreateMaxKeyFromTimestamp(dbTo)

	callback := func(key, value []byte) error {
		expose, err := s.keyGenerator.GetIdFromKey(key)
		if err != nil {
			return err
		}
		collector, ok := collectors[expose]
		if !ok {
			return nil
		}

		ts, err := s.keyGenerator.GetTimestampFromkey(key)
		if err != nil {
			return err
		}
		if !query.MatchesFilters(value, filters, collector.GetType()) {
			return nil
		}

		return collector.Collect(ts, value)
	}

	return s.kvdb.ViewInRange(deviceID, fromKey, toKey, callback)
}

func (s *MetricsRepo) mergeTailCache(deviceID string, from, to time.Time,
	filters []domain.MetricFilter,
	collectors map[string]domain.ExposeResult) error {

	s.tailCacheLock.RLock()
	entries := s.tailCache[deviceID]
	s.tailCacheLock.RUnlock()

	for _, e := range entries {
		// filter by time
		if e.Timestamp.Before(from) || e.Timestamp.After(to) {
			continue
		}

		collector, ok := collectors[e.ExposeName]
		if !ok {
			continue
		}

		if !query.MatchesFilters(e.Value, filters, collector.GetType()) {
			continue
		}

		// merge
		if err := collector.Collect(e.Timestamp, e.Value); err != nil {
			return err
		}
	}

	return nil
}
