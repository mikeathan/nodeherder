package storage

import (
	"bytes"
	"fmt"
	"node-herder/utils"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

type KeyValueDatabase interface {
	Close() error

	Set(bucketName string, key, value []byte) error

	SetBatch(bucketName string, data map[string]any, callback func(key string, value any) ([]byte, []byte, error)) error

	Get(bucketName string, key []byte) ([]byte, error)

	ViewBatchInRange(bucketName string, keys map[string]TimeRangeKey, callback func(key, value []byte) error) error

	ViewInRange(bucketName string, startTime, endTime time.Time, callback func(key, value []byte) error) ([]byte, error)

	Delete(bucketName string, key []byte) error

	Prune(bucketName string, before time.Time) error
}

type BoltKeyValueDatabase struct {
	db         *bolt.DB
	rootBucket string
	mutex      *sync.RWMutex // TODO: use this !!!!
}

func NewBoltKeyValueDatabase(filename string, bucketName string) (KeyValueDatabase, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogErrorf("opening keyvalue database %s failed. error %v", filename, err.Error())
		return nil, err
	}

	kv := &BoltKeyValueDatabase{
		db:         db,
		rootBucket: bucketName,
		mutex:      &sync.RWMutex{},
	}

	err = kv.init(bucketName)
	if err != nil {
		utils.LogErrorf("error initialising keyvalue database %v", err.Error())
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

		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}

		return bucket.Put(key, value)
	})
}

func (b *BoltKeyValueDatabase) SetBatch(bucketName string, data map[string]any, callback func(key string, value any) ([]byte, []byte, error)) error {
	return b.db.Update(func(tx *bolt.Tx) error {

		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}

		for name, value := range data {
			key, buffer, err := callback(name, value)
			if err != nil {
				return err
			}

			err = bucket.Put(key, buffer)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (b *BoltKeyValueDatabase) Get(bucketName string, key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}
		value = bucket.Get(key)
		return nil
	})
	return value, err
}

func (b *BoltKeyValueDatabase) Delete(bucketName string, key []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}
		return bucket.Delete(key)
	})
}

type TimeRangeKey struct {
	From []byte
	To   []byte
}

func (b *BoltKeyValueDatabase) ViewBatchInRange(bucketName string, keys map[string]TimeRangeKey, callback func(key, value []byte) error) error {

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}

		cursor := bucket.Cursor()
		if cursor == nil {
			return fmt.Errorf("bucket cursor not found")
		}

		for name, timeRangeKey := range keys {

			for key, data := cursor.Seek(timeRangeKey.From); key != nil && bytes.Compare(key, timeRangeKey.To) <= 0; key, data = cursor.Next() {
				if !bytes.HasSuffix(key, []byte(name)) {
					continue
				}

				err = callback(key, data)
				if err != nil {
					return err
				}

			}
		}

		return nil
	})

	return nil

}

func (b *BoltKeyValueDatabase) ViewInRange(bucketName string, startTime, endTime time.Time, callback func(key, value []byte) error) ([]byte, error) {

	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	var results []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}

		c := bucket.Cursor()
		for key, v := c.Seek([]byte{byte(startTimestamp)}); key != nil && bytes.Compare(key, []byte{byte(endTimestamp)}) <= 0; key, v = c.Next() {

			// TODO invoke calllback
			results = append(results, v...)
		}
		return nil
	})
	return results, err
}

func (b *BoltKeyValueDatabase) Prune(bucketName string, before time.Time) error {
	beforeTimestamp := before.Unix()

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := b.bucket(tx, bucketName)
		if err != nil {
			return err
		}

		c := bucket.Cursor()
		for k, _ := c.Seek([]byte{byte(beforeTimestamp)}); k != nil; k, _ = c.Next() {
			if err := bucket.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *BoltKeyValueDatabase) bucket(tx *bolt.Tx, bucketName string) (*bolt.Bucket, error) {
	bucket, err := tx.CreateBucketIfNotExists([]byte(b.rootBucket))
	if err != nil {
		return nil, bolt.ErrBucketNotFound
	}

	bucket, err = bucket.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return nil, fmt.Errorf("error creating bucket: %s", bucketName)
	}

	return bucket, nil
}
