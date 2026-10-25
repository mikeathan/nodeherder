package repository

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

const integrationSchema = "1"

var integrationBuckets = []string{"definitions", "instances", "setups", "operations", "limits", "identities"}

// IntegrationKeyring is supplied separately from the database and JWT secrets.
// Keep retiring keys until rotation and all separately retained backups have been verified.
type IntegrationKeyring struct {
	ActiveKeyID string            `json:"active_key_id"`
	Keys        map[string]string `json:"keys"`
}

// ReadIntegrationKeyring requires a bounded private regular file; it never creates missing keys.
func ReadIntegrationKeyring(path string) (IntegrationKeyring, error) {
	var k IntegrationKeyring
	info, err := os.Lstat(path)
	if err != nil {
		return k, fmt.Errorf("read integration keyring: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > 16<<10 {
		return k, errors.New("integration keyring must be a private regular file of at most 16 KiB")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return k, fmt.Errorf("read integration keyring: %w", err)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(&k); err != nil {
		return k, errors.New("invalid integration keyring")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return k, errors.New("invalid integration keyring")
	}
	return k, nil
}

type encryptedIntegrationRecord struct {
	Version    int    `json:"version"`
	KeyID      string `json:"key_id"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

// IntegrationRepository owns the single handle for private connection state.
// It never falls back to temporary storage or generates replacement keys.
type IntegrationRepository struct {
	db           *bolt.DB
	installation string
	active       string
	keys         map[string]cipher.AEAD
	closeOnce    sync.Once
	closeErr     error
}

// OpenIntegrationRepository opens one strict persistent database and verifies all decryption keys.
func OpenIntegrationRepository(directory string, keyring IntegrationKeyring) (repo *IntegrationRepository, err error) {
	if directory == "" {
		return nil, errors.New("integration data directory is required")
	}
	keys := map[string]cipher.AEAD{}
	for id, s := range keyring.Keys {
		if id == "" || len(id) > 64 {
			return nil, errors.New("invalid integration key identifier")
		}
		raw, e := base64.StdEncoding.DecodeString(s)
		if e != nil || len(raw) != 32 {
			return nil, errors.New("integration keys must be base64 encoded 32-byte values")
		}
		block, e := aes.NewCipher(raw)
		if e != nil {
			return nil, e
		}
		keys[id], e = cipher.NewGCM(block)
		if e != nil {
			return nil, e
		}
	}
	if keys[keyring.ActiveKeyID] == nil {
		return nil, errors.New("active integration key is missing")
	}
	if err = os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("integration data directory: %w", err)
	}
	path := filepath.Join(directory, "integrations.db")
	if st, e := os.Lstat(path); e == nil && (!st.Mode().IsRegular() || st.Mode().Perm()&0077 != 0) {
		return nil, errors.New("integration database must be a private regular file")
	}
	db, e := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if e != nil {
		return nil, fmt.Errorf("open integration database: %w", e)
	}
	r := &IntegrationRepository{db: db, active: keyring.ActiveKeyID, keys: keys}
	defer func() {
		if err != nil {
			_ = db.Close()
		}
	}()

	fresh := false
	err = db.View(func(tx *bolt.Tx) error {
		meta := tx.Bucket([]byte("metadata"))
		if meta == nil {
			populated := false
			_ = tx.ForEach(func([]byte, *bolt.Bucket) error { populated = true; return nil })
			if populated {
				return errors.New("missing integration schema")
			}
			fresh = true
			return nil
		}
		return r.readMetadata(tx)
	})
	if err != nil {
		return nil, err
	}
	if fresh {
		err = db.Update(func(tx *bolt.Tx) error {
			meta, e := tx.CreateBucket([]byte("metadata"))
			if e != nil {
				return e
			}
			installation, e := uuid.NewRandom()
			if e != nil {
				return e
			}
			if e = meta.Put([]byte("schema"), []byte(integrationSchema)); e != nil {
				return e
			}
			if e = meta.Put([]byte("installation"), []byte(installation.String())); e != nil {
				return e
			}
			for _, name := range integrationBuckets {
				if _, e = tx.CreateBucket([]byte(name)); e != nil {
					return e
				}
			}
			return r.readMetadata(tx)
		})
		if err != nil {
			return nil, err
		}
	}

	// Verify every record before admitting work. A missing old key cannot look like an empty install.
	err = r.View(func(tx *IntegrationTx) error {
		for _, b := range integrationBuckets {
			if e := tx.Each(b, func(_ string, _ []byte) error { return nil }); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
func (r *IntegrationRepository) readMetadata(tx *bolt.Tx) error {
	meta := tx.Bucket([]byte("metadata"))
	if meta == nil || string(meta.Get([]byte("schema"))) != integrationSchema {
		return errors.New("unsupported integration database schema; restore or upgrade with a compatible binary")
	}
	r.installation = string(meta.Get([]byte("installation")))
	if _, e := uuid.Parse(r.installation); e != nil {
		return errors.New("invalid integration installation identity")
	}
	for _, name := range integrationBuckets {
		if tx.Bucket([]byte(name)) == nil {
			return errors.New("missing integration aggregate")
		}
	}
	return nil
}

// Close releases the owned bbolt handle exactly once.
func (r *IntegrationRepository) Close() error {
	r.closeOnce.Do(func() { r.closeErr = r.db.Close() })
	return r.closeErr
}

// IntegrationTx exposes record transactions; callers never handle ciphertext or raw bbolt buckets.
type IntegrationTx struct {
	tx   *bolt.Tx
	repo *IntegrationRepository
}

// View reads a consistent decrypted aggregate snapshot.
func (r *IntegrationRepository) View(fn func(*IntegrationTx) error) error {
	return r.db.View(func(tx *bolt.Tx) error { return fn(&IntegrationTx{tx, r}) })
}

// Update commits all aggregate/identity/checkpoint changes together or rolls them back.
func (r *IntegrationRepository) Update(fn func(*IntegrationTx) error) error {
	return r.db.Update(func(tx *bolt.Tx) error { return fn(&IntegrationTx{tx, r}) })
}
func (tx *IntegrationTx) bucket(name string) (*bolt.Bucket, error) {
	for _, s := range integrationBuckets {
		if s == name {
			bucket := tx.tx.Bucket([]byte(name))
			if bucket == nil {
				return nil, errors.New("missing integration aggregate")
			}
			return bucket, nil
		}
	}
	return nil, errors.New("unknown integration aggregate")
}
func (r *IntegrationRepository) aad(bucket, id string) []byte {
	return []byte("nodeherder.integrations/" + integrationSchema + "\x00" + r.installation + "\x00" + bucket + "\x00" + id)
}
func (r *IntegrationRepository) decrypt(bucket, id string, data []byte) ([]byte, error) {
	var record encryptedIntegrationRecord
	if e := json.Unmarshal(data, &record); e != nil || record.Version != 1 {
		return nil, errors.New("invalid encrypted integration record")
	}
	key := r.keys[record.KeyID]
	if key == nil {
		return nil, errors.New("integration decryption key is missing")
	}
	if len(record.Nonce) != key.NonceSize() {
		return nil, errors.New("invalid integration nonce")
	}
	plain, e := key.Open(nil, record.Nonce, record.Ciphertext, r.aad(bucket, id))
	if e != nil {
		return nil, errors.New("integration record integrity check failed")
	}
	return plain, nil
}
func (tx *IntegrationTx) Get(bucket, id string, out any) (bool, error) {
	b, e := tx.bucket(bucket)
	if e != nil {
		return false, e
	}
	v := b.Get([]byte(id))
	if v == nil {
		return false, nil
	}
	p, e := tx.repo.decrypt(bucket, id, v)
	if e != nil {
		return false, e
	}
	if e = DecodeIntegrationRecord(p, out); e != nil {
		return false, errors.New("invalid integration record content")
	}
	return true, nil
}
func (tx *IntegrationTx) Put(bucket, id string, value any) error {
	b, e := tx.bucket(bucket)
	if e != nil {
		return e
	}
	if id == "" || len(id) > 256 {
		return errors.New("invalid integration record identifier")
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if e := encoder.Encode(value); e != nil {
		return e
	}
	plain := buffer.Bytes()
	if len(plain) > 4<<20 {
		return errors.New("integration record limit exceeded")
	}
	key := tx.repo.keys[tx.repo.active]
	nonce := make([]byte, key.NonceSize())
	if _, e = io.ReadFull(rand.Reader, nonce); e != nil {
		return e
	}
	record := encryptedIntegrationRecord{1, tx.repo.active, nonce, key.Seal(nil, nonce, plain, tx.repo.aad(bucket, id))}
	data, e := json.Marshal(record)
	if e != nil {
		return e
	}
	return b.Put([]byte(id), data)
}
func (tx *IntegrationTx) Delete(bucket, id string) error {
	b, e := tx.bucket(bucket)
	if e != nil {
		return e
	}
	return b.Delete([]byte(id))
}

// Each supplies decrypted JSON only within the transaction; do not retain its buffers.
func (tx *IntegrationTx) Each(bucket string, fn func(string, []byte) error) error {
	b, e := tx.bucket(bucket)
	if e != nil {
		return e
	}
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			return errors.New("invalid nested integration bucket")
		}
		plain, e := tx.repo.decrypt(bucket, string(k), v)
		if e != nil {
			return e
		}
		return fn(string(k), plain)
	})
}

// Rotate rewrites at most limit records atomically, leaving remaining old-key records readable.
func (r *IntegrationRepository) Rotate(limit int) (int, error) {
	if limit < 1 || limit > 32 {
		return 0, errors.New("rotation batch must be 1–32 records")
	}
	count := 0
	err := r.Update(func(tx *IntegrationTx) error {
		for _, name := range integrationBuckets {
			b, _ := tx.bucket(name)
			c := b.Cursor()
			for k, v := c.First(); k != nil && count < limit; k, v = c.Next() {
				var envelope encryptedIntegrationRecord
				if err := json.Unmarshal(v, &envelope); err != nil {
					return err
				}
				if envelope.KeyID == r.active {
					continue
				}
				plain, e := r.decrypt(name, string(k), v)
				if e != nil {
					return e
				}
				if e = tx.Put(name, string(k), json.RawMessage(plain)); e != nil {
					return e
				}
				count++
			}
		}
		return nil
	})
	return count, err
}

// Backup writes a consistent encrypted database snapshot. Back up YAML and keyring separately.
func (r *IntegrationRepository) Backup(w io.Writer) error {
	return r.db.View(func(tx *bolt.Tx) error { _, e := tx.WriteTo(w); return e })
}

// DecodeIntegrationRecord preserves exact opaque numeric captures in heterogeneous private state.
func DecodeIntegrationRecord(data []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if e := decoder.Decode(out); e != nil {
		return e
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return errors.New("invalid integration record content")
	}
	return nil
}
