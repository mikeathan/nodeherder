package repository

import (
	"fmt"
	"sort"
	"sync"
)

type MemoryRepo[T any] struct {
	store   map[string]*T
	mutex   sync.RWMutex
	updated map[string]bool
}

func NewMemoryRepo[T any]() *MemoryRepo[T] {
	return &MemoryRepo[T]{
		store:   map[string]*T{},
		mutex:   sync.RWMutex{},
		updated: map[string]bool{},
	}
}

func (s *MemoryRepo[T]) Close() error {
	return nil
}

func (s *MemoryRepo[T]) Store(key string, value *T) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ok := s.updated[key]
	s.store[key] = value
	s.updated[key] = true

	return !ok, nil
}

func (s *MemoryRepo[T]) Find(key string) (*T, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if val, ok := s.store[key]; ok {
		return val, nil
	}
	return new(T), fmt.Errorf("key %v not found", key)
}

func (s *MemoryRepo[T]) FindAll() ([]*T, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]*T, 0, len(s.store))
	for _, key := range keys {
		values = append(values, s.store[key])
	}
	return values, nil
}

func (s *MemoryRepo[T]) Remove(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.store[key]; !ok {
		return fmt.Errorf("key %v not found", key)
	}

	delete(s.store, key)
	return nil
}
