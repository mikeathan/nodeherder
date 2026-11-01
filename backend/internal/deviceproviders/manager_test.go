package deviceproviders

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"node-herder/repository"
)

type testDriver struct {
	started  chan string
	stopped  chan string
	removed  chan string
	exitGate <-chan struct{}
}

func TestManagerBindsPhysicalIdentityAcrossUpdates(t *testing.T) {
	keys := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{3}, 32))}}
	repo, err := OpenEnrollments(t.TempDir(), keys)
	if err != nil {
		t.Fatal(err)
	}
	driver := &testDriver{started: make(chan string, 4), stopped: make(chan string, 4)}
	manager, err := NewManager(context.Background(), repo, map[string]Driver{"fake": driver})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close(context.Background())
	first, err := manager.Add(Enrollment{Provider: "fake", Address: "192.168.1.3", Port: 6053})
	if err != nil {
		t.Fatal(err)
	}
	second, err := manager.Add(Enrollment{Provider: "fake", Address: "192.168.1.4", Port: 6053})
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.bindIdentity(context.Background(), manager.workers[first.ID], "physical-one"); err != nil {
		t.Fatal(err)
	}
	worker := manager.workers[first.ID]
	if err := manager.bindInventory(context.Background(), worker, map[string]string{"entity-1": "sensor"}); err != nil {
		t.Fatal(err)
	}
	if err := manager.bindInventory(context.Background(), worker, map[string]string{}); err != nil {
		t.Fatal(err)
	}
	if err := manager.bindInventory(context.Background(), worker, map[string]string{"entity-1": "switch"}); !errors.Is(err, ErrEntityIdentityChanged) {
		t.Fatalf("changed entity kind: %v", err)
	}
	if err := manager.bindIdentity(context.Background(), manager.workers[second.ID], "physical-one"); !errors.Is(err, ErrPhysicalIdentityConflict) {
		t.Fatalf("duplicate: %v", err)
	}
	if err := manager.bindIdentity(context.Background(), manager.workers[first.ID], "physical-two"); !errors.Is(err, ErrPhysicalIdentityChanged) {
		t.Fatalf("changed: %v", err)
	}
	if err := manager.bindInventory(context.Background(), manager.workers[first.ID], map[string]string{"entity-1": "switch"}); !errors.Is(err, ErrEntityIdentityChanged) {
		t.Fatalf("entity kind changed after update: %v", err)
	}
	if _, err := manager.Update(context.Background(), first.ID, EnrollmentUpdate{Address: "192.168.1.5", Port: 6053}); err != nil {
		t.Fatal(err)
	}
	if err := manager.bindIdentity(context.Background(), manager.workers[first.ID], "physical-two"); !errors.Is(err, ErrPhysicalIdentityChanged) {
		t.Fatalf("changed after update: %v", err)
	}
	items, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == first.ID && item.PhysicalID != "physical-one" {
			t.Fatal("identity was not persisted")
		}
	}
}

func (d *testDriver) RemoveDevice(_ context.Context, id string) error { d.removed <- id; return nil }

type testInstance struct {
	id     string
	driver *testDriver
}

func (d *testDriver) NewInstance(e Enrollment) (Instance, error) {
	return &testInstance{id: e.ID, driver: d}, nil
}
func (i *testInstance) Run(ctx context.Context) error {
	i.driver.started <- i.id
	<-ctx.Done()
	if i.driver.exitGate != nil {
		<-i.driver.exitGate
	}
	i.driver.stopped <- i.id
	return ctx.Err()
}
func (i *testInstance) Command(context.Context, string, any) error { return nil }

func TestManagerRecoversAfterCanceledChange(t *testing.T) {
	for _, operation := range []string{"update", "remove"} {
		t.Run(operation, func(t *testing.T) {
			keys := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{3}, 32))}}
			repo, err := OpenEnrollments(t.TempDir(), keys)
			if err != nil {
				t.Fatal(err)
			}
			gate := make(chan struct{})
			driver := &testDriver{started: make(chan string, 2), stopped: make(chan string, 2), exitGate: gate}
			manager, err := NewManager(context.Background(), repo, map[string]Driver{"fake": driver})
			if err != nil {
				t.Fatal(err)
			}
			defer manager.Close(context.Background())
			item, err := manager.Add(Enrollment{Provider: "fake", Address: "192.168.1.3", Port: 6053})
			if err != nil {
				t.Fatal(err)
			}
			<-driver.started
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if operation == "update" {
				_, err = manager.Update(ctx, item.ID, EnrollmentUpdate{Address: "192.168.1.4", Port: 6053})
			} else {
				err = manager.Remove(ctx, item.ID)
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("change error: %v", err)
			}
			close(gate)
			select {
			case id := <-driver.started:
				if id != item.ID {
					t.Fatalf("restarted %s", id)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("original worker not restarted")
			}
			items, err := manager.List()
			if err != nil || len(items) != 1 || items[0].Address != "192.168.1.3" {
				t.Fatalf("enrollment changed: %+v %v", items, err)
			}
		})
	}
}

func TestManagerPersistsStartsAndJoins(t *testing.T) {
	keys := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{3}, 32))}}
	dir := t.TempDir()
	repo, err := OpenEnrollments(dir, keys)
	if err != nil {
		t.Fatal(err)
	}
	driver := &testDriver{started: make(chan string, 2), stopped: make(chan string, 2), removed: make(chan string, 2)}
	manager, err := NewManager(context.Background(), repo, map[string]Driver{"fake": driver})
	if err != nil {
		t.Fatal(err)
	}
	item, err := manager.Add(Enrollment{Provider: "fake", Name: "sensor", Address: "192.168.1.3", Port: 6053, Credential: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if item.Credential != "" || item.ID == "" || <-driver.started != item.ID {
		t.Fatal("invalid created enrollment")
	}
	list, err := manager.List()
	if err != nil || len(list) != 1 || list[0].Credential != "" {
		t.Fatalf("list: %v %v", list, err)
	}
	if err := manager.Command(context.Background(), item.ID, "entity", true); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Add(Enrollment{Provider: "fake", Address: "192.168.1.3", Port: 6053}); err == nil {
		t.Fatal("duplicate accepted")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	updated, err := manager.Update(ctx, item.ID, EnrollmentUpdate{Name: "updated", Address: "192.168.1.4", Port: 6053})
	if err != nil || updated.Credential != "" || updated.Address != "192.168.1.4" {
		t.Fatalf("update: %+v %v", updated, err)
	}
	if <-driver.stopped != item.ID || <-driver.started != item.ID {
		t.Fatal("worker did not restart")
	}
	stored, err := repo.List()
	if err != nil || len(stored) != 1 || stored[0].Credential != "secret" {
		t.Fatal("update did not preserve credential")
	}
	if err := manager.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if <-driver.stopped != item.ID {
		t.Fatal("worker did not stop")
	}
	if err := manager.Command(ctx, item.ID, "entity", true); err == nil {
		t.Fatal("command accepted after close")
	}
	repo, err = OpenEnrollments(dir, keys)
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := NewManager(context.Background(), repo, map[string]Driver{"fake": driver})
	if err != nil {
		t.Fatal(err)
	}
	if <-driver.started != item.ID {
		t.Fatal("saved enrollment did not restart")
	}
	if err := restarted.Remove(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	if <-driver.stopped != item.ID || <-driver.removed != item.ID {
		t.Fatal("worker or device not removed")
	}
	if items, err := restarted.List(); err != nil || len(items) != 0 {
		t.Fatalf("enrollment retained: %v %v", items, err)
	}
	if err := restarted.Close(ctx); err != nil {
		t.Fatal(err)
	}
}
