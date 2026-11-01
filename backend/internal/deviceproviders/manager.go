package deviceproviders

// This file owns configured provider workers. It serializes enrollment changes,
// keeps identity bindings, and routes commands to the active instance without
// creating a second authoritative device state.
import (
	"context"
	"errors"
	"maps"
	"sync"

	"github.com/google/uuid"
)

// Instance owns the lifecycle and command transport for one enrollment. Run must
// stop when its context is canceled; Command must never automatically retry.
type Instance interface {
	Run(context.Context) error
	Command(context.Context, string, any) error
}

// Driver validates one enrollment and constructs an instance without network I/O.
type Driver interface {
	NewInstance(Enrollment) (Instance, error)
}

// DeviceRemover removes a stopped provider device from the shared registry.
type DeviceRemover interface {
	RemoveDevice(context.Context, string) error
}

// DeviceOffliner marks a stopped provider device unavailable before replacement.
type DeviceOffliner interface {
	SetDeviceOffline(context.Context, string) error
}

// Health describes a connection without exposing addresses, credentials or raw errors.
type Health struct {
	Status, Diagnostic string
	Reconnects         uint64
}

// HealthReporter is optional for providers that expose connection diagnostics.
type HealthReporter interface{ Health() Health }

// IdentityBinder is optional for drivers that can verify a stable physical identity.
type IdentityBinder interface {
	SetIdentityValidator(func(context.Context, string) error)
}

// InventoryBinder prevents previously assigned entity IDs changing their kind.
type InventoryBinder interface {
	SetInventoryValidator(func(context.Context, map[string]string) error)
}

// ErrPhysicalIdentityChanged indicates that an endpoint now identifies another device.
var ErrPhysicalIdentityChanged = errors.New("physical device identity changed")

// ErrEntityIdentityChanged indicates a previously used entity ID changed kind.
var ErrEntityIdentityChanged = errors.New("device entity identity changed")

type enrollmentStore interface {
	List() ([]Enrollment, error)
	Save(Enrollment) error
	Delete(string) error
	Close() error
}

type managedInstance struct {
	instance Instance
	done     chan struct{}
	cancel   context.CancelFunc
	item     Enrollment
	stopping bool
	removing bool
}

// Manager owns configured provider workers and their shared enrollment store.
// One manager starts at most one instance per enrollment, independent of UI clients.
type Manager struct {
	mu             sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	store          enrollmentStore
	drivers        map[string]Driver
	workers        map[string]*managedInstance
	closed         bool
	storeCloseOnce sync.Once
	storeCloseErr  error
}

// NewManager validates all persisted enrollments before starting workers. It
// owns the store on success; the caller must close it after a constructor error.
func NewManager(ctx context.Context, store enrollmentStore, drivers map[string]Driver) (*Manager, error) {
	if ctx == nil || store == nil {
		return nil, errors.New("provider manager configuration is required")
	}
	items, err := store.List()
	if err != nil {
		return nil, err
	}
	copied := make(map[string]Driver, len(drivers))
	for name, driver := range drivers {
		if name == "" || driver == nil {
			return nil, errors.New("invalid provider driver")
		}
		copied[name] = driver
	}
	instances := make(map[string]Instance, len(items))
	for _, item := range items {
		driver := copied[item.Provider]
		if driver == nil {
			return nil, errors.New("enrolled provider is unavailable")
		}
		instance, err := driver.NewInstance(item)
		if err != nil {
			return nil, err
		}
		instances[item.ID] = instance
	}
	child, cancel := context.WithCancel(ctx)
	m := &Manager{ctx: child, cancel: cancel, store: store, drivers: copied, workers: make(map[string]*managedInstance)}
	m.mu.Lock()
	for _, item := range items {
		m.start(item, instances[item.ID])
	}
	m.mu.Unlock()
	return m, nil
}

// Add validates and persists a new enrollment before starting its worker. It
// returns a copy without credentials and never performs device network I/O inline.
func (m *Manager) Add(item Enrollment) (Enrollment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.ctx.Err() != nil {
		return Enrollment{}, errors.New("provider manager is stopped")
	}
	if item.ID != "" || item.PhysicalID != "" || len(item.EntityKinds) != 0 {
		return Enrollment{}, errors.New("enrollment ID is assigned by the server")
	}
	item.ID = uuid.NewString()
	if err := validateEnrollment(&item); err != nil {
		return Enrollment{}, err
	}
	driver := m.drivers[item.Provider]
	if driver == nil {
		return Enrollment{}, errors.New("provider is unavailable")
	}
	instance, err := driver.NewInstance(item)
	if err != nil {
		return Enrollment{}, err
	}
	if err := m.store.Save(item); err != nil {
		return Enrollment{}, err
	}
	m.start(item, instance)
	item.Credential = ""
	return item, nil
}

// List returns persisted enrollment metadata with credentials removed.
func (m *Manager) List() ([]Enrollment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, errors.New("provider manager is stopped")
	}
	items, err := m.store.List()
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Credential = ""
		worker := m.workers[items[i].ID]
		if worker == nil || worker.stopping {
			items[i].Status = StatusOffline
			continue
		}
		select {
		case <-worker.done:
			items[i].Status, items[i].Diagnostic = StatusOffline, DiagnosticProviderStopped
		default:
			if reporter, ok := worker.instance.(HealthReporter); ok {
				health := reporter.Health()
				items[i].Status, items[i].Diagnostic = health.Status, health.Diagnostic
				items[i].Reconnects = health.Reconnects
			} else {
				items[i].Status = StatusConnecting
			}
		}
	}
	return items, nil
}

// Command routes a single request to the enrollment's active instance; its
// outcome can be ambiguous if a connection fails during transmission.
func (m *Manager) Command(ctx context.Context, id, entityID string, value any) error {
	m.mu.Lock()
	worker := m.workers[id]
	closed := m.closed
	stopping := worker != nil && worker.stopping
	m.mu.Unlock()
	if closed || worker == nil || stopping {
		return errors.New("provider device is offline")
	}
	select {
	case <-worker.done:
		return errors.New("provider device is offline")
	default:
	}
	return worker.instance.Command(ctx, entityID, value)
}

func (m *Manager) start(item Enrollment, instance Instance) {
	ctx, cancel := context.WithCancel(m.ctx)
	worker := &managedInstance{instance: instance, done: make(chan struct{}), cancel: cancel, item: item}
	m.workers[item.ID] = worker
	if binder, ok := instance.(IdentityBinder); ok {
		binder.SetIdentityValidator(func(ctx context.Context, identity string) error {
			return m.bindIdentity(ctx, worker, identity)
		})
	}
	if binder, ok := instance.(InventoryBinder); ok {
		binder.SetInventoryValidator(func(ctx context.Context, kinds map[string]string) error {
			return m.bindInventory(ctx, worker, kinds)
		})
	}
	go func() { defer close(worker.done); defer cancel(); _ = instance.Run(ctx) }()
}

func (m *Manager) bindInventory(ctx context.Context, worker *managedInstance, kinds map[string]string) error {
	if ctx == nil {
		return errors.New("inventory context is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.closed || worker.stopping || m.workers[worker.item.ID] != worker {
		return errors.New("provider manager is stopped")
	}
	bindings := maps.Clone(worker.item.EntityKinds)
	if bindings == nil {
		bindings = make(map[string]string)
	}
	changed := false
	for id, kind := range kinds {
		if previous, ok := bindings[id]; ok && previous != kind {
			return ErrEntityIdentityChanged
		}
		if _, ok := bindings[id]; !ok {
			bindings[id] = kind
			changed = true
		}
	}
	if !changed {
		return nil
	}
	item := worker.item
	item.EntityKinds = bindings
	if err := m.store.Save(item); err != nil {
		return err
	}
	worker.item = item
	return nil
}

func (m *Manager) bindIdentity(ctx context.Context, worker *managedInstance, identity string) error {
	if ctx == nil || identity == "" {
		return errors.New("physical device identity is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.closed || worker.stopping || m.workers[worker.item.ID] != worker {
		return errors.New("provider manager is stopped")
	}
	if worker.item.PhysicalID != "" {
		if worker.item.PhysicalID != identity {
			return ErrPhysicalIdentityChanged
		}
		return nil
	}
	item := worker.item
	item.PhysicalID = identity
	if err := m.store.Save(item); err != nil {
		return err
	}
	worker.item = item
	return nil
}

// ErrEnrollmentNotFound indicates no active or saved enrollment has the requested ID.
var ErrEnrollmentNotFound = errors.New("provider enrollment not found")

// EnrollmentUpdate replaces public connection fields. A nil Credential keeps
// the existing secret; a non-nil empty value selects an unencrypted connection.
type EnrollmentUpdate struct {
	Name, Address string
	Port          int
	Credential    *string
}

// Update validates a replacement before stopping the current session. It joins
// the old worker before saving and starting the replacement, so sessions never overlap.
func (m *Manager) Update(ctx context.Context, id string, update EnrollmentUpdate) (Enrollment, error) {
	if ctx == nil {
		return Enrollment{}, errors.New("update context is required")
	}
	m.mu.Lock()
	worker := m.workers[id]
	if m.closed || worker == nil {
		m.mu.Unlock()
		return Enrollment{}, ErrEnrollmentNotFound
	}
	if worker.removing {
		m.mu.Unlock()
		return Enrollment{}, errors.New("device change already in progress")
	}
	item := worker.item
	item.Name, item.Address, item.Port = update.Name, update.Address, update.Port
	if update.Credential != nil {
		item.Credential = *update.Credential
	}
	if err := validateEnrollment(&item); err != nil {
		m.mu.Unlock()
		return Enrollment{}, err
	}
	instance, err := m.drivers[item.Provider].NewInstance(item)
	if err != nil {
		m.mu.Unlock()
		return Enrollment{}, err
	}
	worker.stopping, worker.removing = true, true
	worker.cancel()
	m.mu.Unlock()
	recoverCanceled := false
	defer func() { m.finishChange(worker, recoverCanceled) }()
	select {
	case <-worker.done:
	case <-ctx.Done():
		recoverCanceled = true
		return Enrollment{}, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		m.restart(worker)
		return Enrollment{}, err
	}
	if offliner, ok := m.drivers[item.Provider].(DeviceOffliner); ok {
		if err := offliner.SetDeviceOffline(ctx, id); err != nil {
			m.restart(worker)
			return Enrollment{}, err
		}
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return Enrollment{}, errors.New("provider manager is stopped")
	}
	err = m.store.Save(item)
	if err == nil {
		m.start(item, instance)
	}
	m.mu.Unlock()
	if err != nil {
		m.restart(worker)
		return Enrollment{}, err
	}
	item.Credential = ""
	return item, nil
}

// Remove stops a worker, removes its shared device and then deletes its encrypted
// configuration. It retains historical metrics and never overlaps old/new sessions.
func (m *Manager) Remove(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("removal context is required")
	}
	m.mu.Lock()
	worker := m.workers[id]
	if m.closed || worker == nil {
		m.mu.Unlock()
		return ErrEnrollmentNotFound
	}
	if worker.removing {
		m.mu.Unlock()
		return errors.New("device removal already in progress")
	}
	worker.stopping, worker.removing = true, true
	worker.cancel()
	m.mu.Unlock()
	recoverCanceled := false
	defer func() { m.finishChange(worker, recoverCanceled) }()
	select {
	case <-worker.done:
	case <-ctx.Done():
		recoverCanceled = true
		return ctx.Err()
	}
	m.mu.Lock()
	closed := m.closed
	driver := m.drivers[worker.item.Provider]
	m.mu.Unlock()
	if closed {
		return errors.New("provider manager is stopped")
	}
	remover, ok := driver.(DeviceRemover)
	if !ok {
		m.restart(worker)
		return errors.New("provider removal is unavailable")
	}
	if err := remover.RemoveDevice(ctx, id); err != nil {
		m.restart(worker)
		return err
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errors.New("provider manager is stopped")
	}
	err := m.store.Delete(id)
	if err == nil {
		delete(m.workers, id)
	}
	m.mu.Unlock()
	if err != nil {
		m.restart(worker)
		return err
	}
	return nil
}

func (m *Manager) finishChange(worker *managedInstance, recoverCanceled bool) {
	m.mu.Lock()
	worker.removing = false
	m.mu.Unlock()
	if recoverCanceled {
		go m.restartWhenStopped(worker)
	}
}

// A canceled request cannot restart until the old session has actually exited.
// Shutdown cancels the wait; a newer operation or generation owns its own recovery.
func (m *Manager) restartWhenStopped(worker *managedInstance) {
	select {
	case <-worker.done:
	case <-m.ctx.Done():
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.ctx.Err() != nil || m.workers[worker.item.ID] != worker || worker.removing {
		return
	}
	instance, err := m.drivers[worker.item.Provider].NewInstance(worker.item)
	if err == nil {
		m.start(worker.item, instance)
	}
}

// restart is called only after the previous worker has exited. A failed delete
// keeps the enrollment and recreates the provider so it can rehydrate the device.
func (m *Manager) restart(worker *managedInstance) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.ctx.Err() != nil || m.workers[worker.item.ID] != worker {
		return
	}
	instance, err := m.drivers[worker.item.Provider].NewInstance(worker.item)
	if err == nil {
		m.start(worker.item, instance)
	}
}

// Close cancels and joins all workers before closing the store. If ctx expires,
// a later call can finish the join and close; it never silently abandons workers.
func (m *Manager) Close(ctx context.Context) error {
	if ctx == nil {
		return errors.New("shutdown context is required")
	}
	m.mu.Lock()
	m.closed = true
	m.cancel()
	workers := make([]*managedInstance, 0, len(m.workers))
	for _, worker := range m.workers {
		workers = append(workers, worker)
	}
	m.mu.Unlock()
	for _, worker := range workers {
		select {
		case <-worker.done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.storeCloseOnce.Do(func() { m.storeCloseErr = m.store.Close() })
	return m.storeCloseErr
}
