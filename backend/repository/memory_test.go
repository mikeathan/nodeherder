package repository_test

import (
	"fmt"
	"node-herder/repository"
	"reflect"
	"sort"
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestMemoryRepo_Store(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()
	key := "user1"
	value := TestStruct{Name: "John Doe", Age: 30}

	isUpdated, err := repo.Store(key, value)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
	if !isUpdated {
		t.Errorf("Expected isUpdated to be true for a new record")
	}

	retrievedValue, err := repo.Find(key)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, value) { // Use DeepEqual for structs
		t.Errorf("Expected retrieved value to be %v, got %v", value, retrievedValue)
	}
}

func TestMemoryRepo_Update(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()
	key := "user1"
	value := TestStruct{Name: "John Doe", Age: 30}

	_, err := repo.Store(key, value)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}

	updatedValue := TestStruct{Name: "John Doe", Age: 35}
	isUpdated, err := repo.Store(key, updatedValue)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
	if isUpdated {
		t.Errorf("Expected isUpdated to be false for an existing record")
	}

	retrievedValue, err := repo.Find(key)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, updatedValue) {
		t.Errorf("Expected retrieved value to be %v, got %v", updatedValue, retrievedValue)
	}
}

func TestMemoryRepo_Dequeue(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()
	key := "user1"
	value := TestStruct{Name: "John Doe", Age: 30}

	_, err := repo.Store(key, value)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}

	dequeuedValue, err := repo.Dequeue(key)
	if err != nil {
		t.Errorf("Dequeue failed: %v", err)
	}
	if !reflect.DeepEqual(dequeuedValue, value) {
		t.Errorf("Expected dequeued value to be %v, got %v", value, dequeuedValue)
	}

	_, err = repo.Find(key)
	if err == nil {
		t.Errorf("Expected key %v to be removed after Dequeue, but found", key)
	} else if err.Error() != fmt.Sprintf("key %v not found", key) {
		t.Errorf("Unexpected error on Find: %v", err)
	}
}

func TestMemoryRepo_Find_NotFound(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()
	key := "non-existent"

	_, err := repo.Find(key)
	if err == nil {
		t.Errorf("Expected error for non-existent key, but got nil")
	} else if err.Error() != fmt.Sprintf("key %v not found", key) {
		t.Errorf("Unexpected error on Find: %v", err)
	}
}

func TestMemoryRepo_FindAll(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()

	data := map[string]TestStruct{
		"user1": {Name: "John Doe", Age: 30},
		"user3": {Name: "Alice Smith", Age: 25},
		"user2": {Name: "Bob Johnson", Age: 40},
	}

	var keys []string
	for k, v := range data {
		_, err := repo.Store(k, v)
		if err != nil {
			t.Fatalf("Store failed: %v", err) // Fatal error in setup
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	expectedValues := make([]TestStruct, 0, len(data))
	for _, key := range keys {
		expectedValues = append(expectedValues, data[key])
	}
	actualValues, err := repo.FindAll()
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}

	if !reflect.DeepEqual(actualValues, expectedValues) {
		t.Errorf("Expected FindAll to return %v, got %v", expectedValues, actualValues)
	}
}

func TestMemoryRepo_Remove(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()
	key := "user1"
	value := TestStruct{Name: "John Doe", Age: 30}

	_, err := repo.Store(key, value)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	err = repo.Remove(key)
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	_, err = repo.Find(key)
	if err == nil {
		t.Errorf("Expected error after removal, but found value")
	} else if err.Error() != fmt.Sprintf("key %v not found", key) {
		t.Errorf("Unexpected error on Find after removal: %v", err)
	}

	// Removing a non-existent key
	err = repo.Remove("non-existent")
	if err == nil {
		t.Errorf("Expected error removing non-existent key, but got nil")
	} else if err.Error() != fmt.Sprintf("key %v not found", "non-existent") {
		t.Errorf("Unexpected error on Remove non-existent key: %v", err)
	}
}

func TestMemoryRepo_Close(t *testing.T) {
	repo := repository.NewMemoryRepo[TestStruct]()
	err := repo.Close()
	if err != nil {
		t.Errorf("Close returned an unexpected error: %v", err)
	}
}
