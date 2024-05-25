package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/settings"
	"sync"

	"github.com/boltdb/bolt"
)

const baseFilename = "settings.db"
const settingsBucketName = "settings"

type FileSettingsRepo struct {
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewFileSettingsRepo() (settings.Repository, error) {
	return NewFileSettingsRepoFromFile(baseFilename)
}

func NewFileSettingsRepoFromFile(filename string) (settings.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &FileSettingsRepo{
		db:    db,
		mutex: sync.RWMutex{},
	}, nil
}

func (s *FileSettingsRepo) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *FileSettingsRepo) Store(key string, value any) error {
	err := s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(settingsBucketName))
		if err != nil {
			return err
		}

		buf, err := json.Marshal(value)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), buf)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *FileSettingsRepo) Get(key string) (any, error) {

	var value any
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(settingsBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(key))
		if buffer == nil {
			return fmt.Errorf("key %v not found", key)
		}

		err := json.Unmarshal(buffer, &value)
		if err != nil {
			return err
		}

		return nil
	})

	return value, err
}
