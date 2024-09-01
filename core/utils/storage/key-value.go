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

	ViewInRange(bucketName string, from []byte, to []byte, callback func(key, value []byte) error) error

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

	defer b.mutex.Unlock()
	b.mutex.Lock()

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

	defer b.mutex.Unlock()
	b.mutex.Lock()

	return b.db.Update(func(tx *bolt.Tx) error {

		bucket, err := b.createBucket(tx, bucketName)
		if err != nil {
			return err
		}

		return bucket.Put(key, value)
	})
}

func (b *BoltKeyValueDatabase) SetBatch(bucketName string, data map[string]any, callback func(key string, value any) ([]byte, []byte, error)) error {
	defer b.mutex.Unlock()
	b.mutex.Lock()

	return b.db.Update(func(tx *bolt.Tx) error {

		bucket, err := b.createBucket(tx, bucketName)
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
	defer b.mutex.RUnlock()
	b.mutex.RLock()

	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket, err := b.openBucket(tx, bucketName)
		if err != nil {
			return err
		}
		value = bucket.Get(key)
		return nil
	})
	return value, err
}

func (b *BoltKeyValueDatabase) Delete(bucketName string, key []byte) error {
	defer b.mutex.Unlock()
	b.mutex.Lock()

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := b.openBucket(tx, bucketName)
		if err != nil {
			return err
		}
		return bucket.Delete(key)
	})
}

func (b *BoltKeyValueDatabase) ViewInRange(bucketName string, from, to []byte, callback func(key, value []byte) error) error {
	defer b.mutex.RUnlock()
	b.mutex.RLock()

	return b.db.View(func(tx *bolt.Tx) error {
		bucket, err := b.openBucket(tx, bucketName)
		if err != nil {
			return err
		}

		cursor := bucket.Cursor()
		if cursor == nil {
			return fmt.Errorf("bucket cursor not found")
		}

		for key, data := cursor.Seek(from); key != nil && bytes.Compare(key, to) <= 0; key, data = cursor.Next() {

			err = callback(key, data)
			if err != nil {
				return err
			}

		}

		return nil
	})
}

func (b *BoltKeyValueDatabase) Prune(bucketName string, before time.Time) error {
	defer b.mutex.Unlock()
	b.mutex.Lock()

	beforeTimestamp := before.Unix()

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := b.openBucket(tx, bucketName)
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

func (b *BoltKeyValueDatabase) openBucket(tx *bolt.Tx, bucketName string) (*bolt.Bucket, error) {

	bucket := tx.Bucket([]byte(b.rootBucket)).Bucket([]byte(bucketName))
	if bucket == nil {
		return nil, bolt.ErrBucketNotFound
	}

	return bucket, nil
}

func (b *BoltKeyValueDatabase) createBucket(tx *bolt.Tx, bucketName string) (*bolt.Bucket, error) {
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
