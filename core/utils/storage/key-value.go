package storage

import (
	"bytes"
	"node-herder/utils"
	"time"

	"github.com/boltdb/bolt"
)

type KeyValueDatabase interface {
	//Update(fn func(tx *bolt.Tx) error) error
	//View(fn func(*bolt.Tx) error) error
	Close() error

	Set(bucketName string, key, value []byte) error

	SetBatch(bucketName string, key, data []interface{}) error

	Get(bucketName string, key []byte) ([]byte, error)

	ViewInRange(bucketName string, startTime, endTime time.Time, callback func(key, value []byte) error) ([]byte, error)

	Delete(bucketName string, key []byte) error

	Prune(bucketName string, before time.Time) error
}

type BoltKeyValueDatabase struct {
	db *bolt.DB
}

func NewBoltKeyValueDatabase(filename string, bucketName string) (KeyValueDatabase, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	kv := &BoltKeyValueDatabase{
		db: db,
	}
	err = kv.init(bucketName)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	return kv, nil
}

func (b *BoltKeyValueDatabase) init(bucketName string) error {
	tx, err := b.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltKeyValueDatabase) Close() error {
	err := b.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (b *BoltKeyValueDatabase) Set(bucketName string, key, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			var err error
			bucket, err = tx.CreateBucket([]byte(bucketName))
			if err != nil {
				return err

			}
		}

		return bucket.Put(key, value)
	})
}

func (b *BoltKeyValueDatabase) Get(bucketName string, key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}
		value = bucket.Get(key)
		return nil
	})
	return value, err
}

func (b *BoltKeyValueDatabase) Delete(bucketName string, key []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}
		return bucket.Delete(key)
	})
}

// we can pass callback as argument eg callback func(key, value []byte) error

// w have problem with searching in differnt exposes - check metrics example
func (b *BoltKeyValueDatabase) ViewInRange(bucketName string, startTime, endTime time.Time, callback func(key, value []byte) error) ([]byte, error) {

	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	var results []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName)) // cursor := tx.Bucket([]byte(metricsBucketName)).Bucket([]byte(device.Id)).Cursor()
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		c := bucket.Cursor()
		for k, v := c.Seek([]byte{byte(startTimestamp)}); k != nil && bytes.Compare(k, []byte{byte(endTimestamp)}) <= 0; k, v = c.Next() {
			// Append the value to the results (adjust as needed based on your value format)
			results = append(results, v...)
		}
		return nil
	})
	return results, err
}

func (db *BoltKeyValueDatabase) Prune(bucketName string, before time.Time) error {
	// Assuming timestamps are stored as integers within the keys
	beforeTimestamp := before.Unix()

	return db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		c := bucket.Cursor()
		for k, _ := c.Seek([]byte{byte(beforeTimestamp)}); k != nil; k, _ = c.Next() {
			// Delete the key if the timestamp is before the specified time
			if err := bucket.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *BoltKeyValueDatabase) View(fn func(*bolt.Tx) error) error {
	return b.db.View(fn)
}

func (b *BoltKeyValueDatabase) Update(fn func(tx *bolt.Tx) error) error {
	return b.db.Update(fn)
}
