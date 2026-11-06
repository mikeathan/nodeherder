package spatialhistory

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"node-herder/utils"
	"node-herder/utils/storage"

	bolt "go.etcd.io/bbolt"
)

const (
	filename   = "spatial.db"
	bucketName = "spatial"
	timeLayout = "2006-01-02T15:04:05.000000000Z"
)

// Repository owns chunk persistence. Implementations must be safe for one writer and many
// readers; the recorder serializes writes through its worker.
type Repository interface {
	SaveChunk(device string, chunk Chunk) error
	Chunks(device string, from, to time.Time) ([]Chunk, error)
	Prune(before time.Time) error
	Close() error
}

// Store is a bbolt-backed repository in its own file, sharing the data directory with metrics.
type Store struct {
	kv    storage.KeyValueDatabase
	clock utils.Clock
}

// NewStore opens data/spatial.db with the spatial bucket.
func NewStore() (*Store, error) {
	return NewStoreAt(filepath.Join(utils.GetDataDir(), filename))
}

// NewStoreAt opens a store at an explicit path; tests use this, production uses the data dir.
func NewStoreAt(path string) (*Store, error) {
	kv, err := storage.NewBoltKeyValueDatabase(path, bucketName)
	if err != nil {
		return nil, err
	}
	return &Store{kv: kv, clock: utils.NewRealClock()}, nil
}

func chunkKey(at time.Time, session string) []byte {
	return []byte(fmt.Sprintf("%s_%s", at.UTC().Format(timeLayout), session))
}

func keyTimestamp(key []byte) (time.Time, error) {
	if len(key) < len(timeLayout) {
		return time.Time{}, fmt.Errorf("spatial history key too short")
	}
	return time.Parse(timeLayout, string(key[:len(timeLayout)]))
}

// SaveChunk writes one bounded value; keys sort by chunk start time.
func (s *Store) SaveChunk(device string, chunk Chunk) error {
	if device == "" || chunk.Session == "" || len(chunk.Points) == 0 {
		return nil
	}
	if err := s.kv.SetBatch(device, map[string]any{chunk.Session: chunk}, func(_ string, data any) ([]byte, []byte, error) {
		stored := data.(Chunk)
		encoded, err := json.Marshal(stored)
		if err != nil {
			return nil, nil, err
		}
		return chunkKey(time.UnixMilli(stored.From), stored.Session), encoded, nil
	}); err != nil {
		return err
	}
	return nil
}

// Chunks returns chunks whose start falls in the inclusive range.
func (s *Store) Chunks(device string, from, to time.Time) ([]Chunk, error) {
	if to.Before(from) || device == "" {
		return nil, nil
	}
	fromKey := chunkKey(from, "")
	toKey := append([]byte(to.UTC().Format(timeLayout)+"_"), 0xff)
	var chunks []Chunk
	if err := s.kv.ViewInRange(device, fromKey, toKey, func(_, value []byte) error {
		var chunk Chunk
		if err := json.Unmarshal(value, &chunk); err != nil {
			return err
		}
		if len(chunk.Points) > 0 {
			chunks = append(chunks, chunk)
		}
		return nil
	}); err != nil {
		if errors.Is(err, bolt.ErrBucketNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return chunks, nil
}

// Prune deletes chunks older than the cutoff, oldest first.
func (s *Store) Prune(before time.Time) error {
	return s.kv.Prune(func(key []byte) (bool, error) {
		timestamp, err := keyTimestamp(key)
		if err != nil {
			return false, nil
		}
		return timestamp.Before(before), nil
	})
}

// Close releases the database file.
func (s *Store) Close() error { return s.kv.Close() }
