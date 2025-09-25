package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type TypeRegistry[K comparable, T any] map[K]reflect.Type

type Serializer[K comparable, T any] struct {
	TypeRegistry TypeRegistry[K, T]
	TypeField    string
}

func (s *Serializer[K, T]) Unmarshal(data []byte) (T, error) {
	var zero T

	temp := make(map[string]interface{})
	if err := json.Unmarshal(data, &temp); err != nil {
		return zero, err
	}

	rawType, ok := temp[s.TypeField]
	if !ok {
		return zero, fmt.Errorf("missing type field %q", s.TypeField)
	}

	// cast the key to K
	key, ok := rawType.(string)
	if !ok {
		return zero, fmt.Errorf("type field is not a string")
	}

	var k K
	switch any(k).(type) {
	case string:
		k = any(key).(K)
	default:
		return zero, fmt.Errorf("unsupported key type %T", k)
	}

	concreteType, ok := s.TypeRegistry[k]
	if !ok {
		return zero, fmt.Errorf("unknown type: %v", k)
	}

	instance := reflect.New(concreteType).Interface().(T)

	if err := json.Unmarshal(data, instance); err != nil {
		return zero, err
	}

	return instance, nil
}
