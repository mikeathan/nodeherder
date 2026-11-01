package deviceproviders

// This file owns the separate encrypted enrollment store. It persists provider
// configuration, credentials and identity bindings, never live entity state.
import (
	"errors"
	"io"
	"net"
	"node-herder/repository"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ErrEnrollmentExists indicates that a provider endpoint is already enrolled.
var ErrEnrollmentExists = errors.New("device endpoint is already enrolled")

// ErrPhysicalIdentityConflict indicates that a physical device is already bound.
var ErrPhysicalIdentityConflict = errors.New("physical device is already enrolled")

// Enrollment contains connection configuration and optional health metadata, never
// entity state. Credentials and private identity bindings are excluded from JSON.
type Enrollment struct {
	ID          string            `json:"id"`
	Provider    string            `json:"provider"`
	Name        string            `json:"name,omitempty"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Status      string            `json:"status,omitempty"`
	Diagnostic  string            `json:"diagnostic,omitempty"`
	Reconnects  uint64            `json:"reconnections,omitempty"`
	Credential  string            `json:"-"`
	PhysicalID  string            `json:"-"`
	EntityKinds map[string]string `json:"-"`
}

type enrollmentRecord struct {
	Version     int               `json:"version"`
	ID          string            `json:"id"`
	Provider    string            `json:"provider"`
	Name        string            `json:"name,omitempty"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Credential  string            `json:"credential"`
	PhysicalID  string            `json:"physical_id,omitempty"`
	EntityKinds map[string]string `json:"entity_kinds,omitempty"`
}

// EnrollmentRepository owns a separate encrypted database for provider settings.
// It has no dependency on the optional HTTP integration service's lifecycle.
type EnrollmentRepository struct {
	storage   *repository.IntegrationRepository
	closeOnce sync.Once
	closeErr  error
}

// OpenEnrollments opens the device-providers child of dataDirectory using its
// own keyring. The caller owns the result and must close it.
func OpenEnrollments(dataDirectory string, keyring repository.IntegrationKeyring) (*EnrollmentRepository, error) {
	if dataDirectory == "" {
		return nil, errors.New("device data directory is required")
	}
	storage, err := repository.OpenIntegrationRepository(filepath.Join(dataDirectory, "device-providers"), keyring)
	if err != nil {
		return nil, err
	}
	r := &EnrollmentRepository{storage: storage}
	if _, err := r.List(); err != nil {
		_ = storage.Close()
		return nil, err
	}
	return r, nil
}

// Close releases the database handle and is safe to call repeatedly.
func (r *EnrollmentRepository) Close() error {
	r.closeOnce.Do(func() { r.closeErr = r.storage.Close() })
	return r.closeErr
}

// List returns a snapshot of all enrollments including credentials for internal
// provider use. Serialization of Enrollment excludes credentials.
func (r *EnrollmentRepository) List() ([]Enrollment, error) {
	var result []Enrollment
	err := r.storage.View(func(tx *repository.IntegrationTx) error {
		var err error
		result, err = readEnrollments(tx)
		return err
	})
	return result, err
}

// Save atomically creates or replaces one enrollment. It rejects duplicate
// provider/address/port bindings and enforces a maximum of 32 enrollments.
// Provider-specific credential validation remains the caller's responsibility.
func (r *EnrollmentRepository) Save(item Enrollment) error {
	if err := validateEnrollment(&item); err != nil {
		return err
	}
	return r.storage.Update(func(tx *repository.IntegrationTx) error {
		all, err := readEnrollments(tx)
		if err != nil {
			return err
		}
		exists := false
		for _, other := range all {
			if other.ID == item.ID {
				exists = true
				continue
			}
			if other.Provider == item.Provider && other.Address == item.Address && other.Port == item.Port {
				return ErrEnrollmentExists
			}
			if item.PhysicalID != "" && other.Provider == item.Provider && other.PhysicalID == item.PhysicalID {
				return ErrPhysicalIdentityConflict
			}
		}
		if !exists && len(all) >= maxEnrollments {
			return errors.New("device enrollment limit reached")
		}
		return tx.Put("instances", item.ID, enrollmentRecord{Version: enrollmentRecordVersion, ID: item.ID, Provider: item.Provider, Name: item.Name, Address: item.Address, Port: item.Port, Credential: item.Credential, PhysicalID: item.PhysicalID, EntityKinds: item.EntityKinds})
	})
}

// Delete removes the active record. The caller must first stop and fence the
// session; old database pages and backups are not securely erased.
func (r *EnrollmentRepository) Delete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid enrollment identifier")
	}
	return r.storage.Update(func(tx *repository.IntegrationTx) error { return tx.Delete("instances", id) })
}

// Backup writes a consistent encrypted snapshot. Back up the matching keyring separately.
func (r *EnrollmentRepository) Backup(w io.Writer) error { return r.storage.Backup(w) }

// Rotate rewrites at most limit records using the active key, retaining old keys for backups.
func (r *EnrollmentRepository) Rotate(limit int) (int, error) { return r.storage.Rotate(limit) }

func readEnrollments(tx *repository.IntegrationTx) ([]Enrollment, error) {
	var result []Enrollment
	seen := make(map[string]bool)
	physical := make(map[string]bool)
	err := tx.Each("instances", func(id string, plain []byte) error {
		var record enrollmentRecord
		if err := repository.DecodeIntegrationRecord(plain, &record); err != nil || record.Version != enrollmentRecordVersion || record.ID != id {
			return errors.New("invalid device enrollment record")
		}
		item := Enrollment{ID: record.ID, Provider: record.Provider, Name: record.Name, Address: record.Address, Port: record.Port, Credential: record.Credential, PhysicalID: record.PhysicalID, EntityKinds: record.EntityKinds}
		if err := validateEnrollment(&item); err != nil {
			return errors.New("invalid device enrollment record")
		}
		endpoint, _ := NewEndpoint(item.Address, item.Port)
		binding := item.Provider + "\x00" + endpoint.NetworkAddress()
		if seen[binding] {
			return errors.New("duplicate device enrollment")
		}
		seen[binding] = true
		if item.PhysicalID != "" {
			binding = item.Provider + "\x00" + item.PhysicalID
			if physical[binding] {
				return errors.New("duplicate physical device enrollment")
			}
			physical[binding] = true
		}
		result = append(result, item)
		if len(result) > maxEnrollments {
			return errors.New("device enrollment limit exceeded")
		}
		return nil
	})
	return result, err
}

func validateEnrollment(item *Enrollment) error {
	if id, err := uuid.Parse(item.ID); err != nil || id.String() != item.ID {
		return errors.New("invalid enrollment identifier")
	}
	if item.Provider == "" || len(item.Provider) > maxProviderNameBytes || strings.Trim(item.Provider, providerNameCharacters) != "" || len(item.Credential) > maxCredentialBytes || len(item.Name) > maxEnrollmentNameBytes || !utf8.ValidString(item.Name) {
		return errors.New("invalid enrollment configuration")
	}
	if len(item.PhysicalID) > maxPhysicalIDBytes || strings.ContainsRune(item.PhysicalID, '\x00') || !utf8.ValidString(item.PhysicalID) {
		return errors.New("invalid physical device identity")
	}
	if len(item.EntityKinds) > maxEntityBindings {
		return errors.New("device entity history limit reached")
	}
	for id, kind := range item.EntityKinds {
		if id == "" || kind == "" || len(id) > maxEntityIDBytes || len(kind) > maxEntityKindBytes || !utf8.ValidString(id) || !utf8.ValidString(kind) || strings.ContainsRune(id, '\x00') || strings.ContainsRune(kind, '\x00') {
			return errors.New("invalid device entity binding")
		}
	}
	endpoint, err := NewEndpoint(item.Address, item.Port)
	if err != nil {
		return err
	}
	item.Address, _, err = net.SplitHostPort(endpoint.NetworkAddress())
	return err
}
