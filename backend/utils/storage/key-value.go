package storage

import (
	"bytes"
	"errors"
	"fmt"
	"node-herder/utils"
	"path/filepath"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

type KeyValueDatabase interface {
	Close() error

	Set(key, value []byte) error

	Get(key []byte) ([]byte, error)

	SetBatch(bucketName string, data map[string]any, callback func(key string, value any) ([]byte, []byte, error)) error

	ViewInRange(bucketName string, from []byte, to []byte, callback func(key, value []byte) error) error

	HasDataInRange(bucketName string, from, to []byte) (bool, error)


	DeleteKey(key []byte) error

	GetAll(callback func(key, value []byte) error) error

	Prune(callback func(key []byte) (bool, error)) error
}

type BoltKeyValueDatabase struct {
	db         *bolt.DB
	rootBucket string
	mutex      *sync.RWMutex
}

func NewBoltKeyValueDatabase(filename string, bucketName string) (KeyValueDatabase, error) {

	dir := filepath.Dir(filename)
	finalDir := utils.EnsureDir(dir)
	if finalDir != dir {
		filename = filepath.Join(finalDir, filepath.Base(filename))
	}

	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		if err == bolt.ErrTimeout {
			utils.LogErrorf("TIMEOUT: Could not acquire lock on database %s. Another instance of nodeherder is likely running.", filename)
			return nil, err
		}
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

	b.mutex.Lock()
	defer b.mutex.Unlock()

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

func (b *BoltKeyValueDatabase) Set(key, value []byte) error {

	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(b.rootBucket))
		if err != nil {
			return bolt.ErrBucketNotFound
		}

		return bucket.Put(key, value)
	})
}

func (b *BoltKeyValueDatabase) SetBatch(bucketName string, data map[string]any, callback func(key string, value any) ([]byte, []byte, error)) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.db.Update(func(tx *bolt.Tx) error {

		bucket, err := b.createChildBucket(tx, bucketName)
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

func (b *BoltKeyValueDatabase) Get(key []byte) ([]byte, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(b.rootBucket))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		value = bucket.Get(key)
		return nil
	})

	return value, err
}

func (b *BoltKeyValueDatabase) DeleteKey(key []byte) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.rootBucket))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}
		return bucket.Delete(key)
	})
}

func (b *BoltKeyValueDatabase) GetAll(callback func(key, value []byte) error) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.rootBucket))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(key, value []byte) error {
			if value == nil {
				return nil // skip nested buckets
			}
			return callback(key, value)
		})
	})
}

func (b *BoltKeyValueDatabase) ViewInRange(bucketName string, from, to []byte, callback func(key, value []byte) error) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.db.View(func(tx *bolt.Tx) error {
		bucket, err := b.openChildBucket(tx, bucketName)
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

func (b *BoltKeyValueDatabase) HasDataInRange(bucketName string, from, to []byte) (bool, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var result bool
	err := b.db.View(func(tx *bolt.Tx) error {

		bucket, err := b.openChildBucket(tx, bucketName)
		if err != nil {
			return err
		}

		cursor := bucket.Cursor()
		for key, _ := cursor.Seek(from); key != nil && bytes.Compare(key, to) <= 0; cursor.Next() {
			result = true

			// err = callback(key, data)
			// if err != nil {
			// 	return err
			// }

			break
		}
		return nil
	})

	return result, err
}

func (b *BoltKeyValueDatabase) Prune(callback func(key []byte) (bool, error)) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		rootBucket := tx.Bucket([]byte(b.rootBucket))
		if rootBucket == nil {
			return errors.New("root bucket not found")
		}

		return rootBucket.ForEach(func(bucketName, _ []byte) error {
			bucket := rootBucket.Bucket(bucketName)
			if bucket == nil {
				return bolt.ErrBucketNotFound
			}

			return bucket.ForEach(func(key, value []byte) error {
				b.mutex.Lock()

				canDelete, err := callback(key)

				b.mutex.Unlock()

				if err != nil {
					return err
				}

				if canDelete {
					err = bucket.Delete(key)
					if err != nil {
						utils.LogErrorf("Error deleting key: %v", err)
					}
				}
				return nil
			})
		})
	})
	return err
	// return b.db.Update(func(tx *bolt.Tx) error {
	// 	rootBucket := tx.Bucket([]byte(b.rootBucket))
	// 	if rootBucket == nil {
	// 		return bolt.ErrBucketNotFound
	// 	}

	// 	c := rootBucket.Cursor()
	// 	for bucketName, _ := c.First(); bucketName != nil; bucketName, _ = c.Next() {

	// 		rootBucket.Bucket(bucketName).ForEach(func(key, _ []byte) error {
	// 			canDelete, err := callback(key)
	// 			if err != nil {
	// 				return err
	// 			}

	// 			if canDelete {
	// 				if err := rootBucket.Delete(key); err != nil {
	// 					return err
	// 				}
	// 			}
	// 			return nil
	// 		})
	// 	}

	// 	return nil
	// })
}

func (b *BoltKeyValueDatabase) openChildBucket(tx *bolt.Tx, bucketName string) (*bolt.Bucket, error) {

	bucket := tx.Bucket([]byte(b.rootBucket)).Bucket([]byte(bucketName))
	if bucket == nil {
		return nil, bolt.ErrBucketNotFound
	}

	return bucket, nil
}

func (b *BoltKeyValueDatabase) createChildBucket(tx *bolt.Tx, bucketName string) (*bolt.Bucket, error) {
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
