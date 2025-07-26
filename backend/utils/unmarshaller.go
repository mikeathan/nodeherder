package utils

import (
	"encoding/json"
	"reflect"
)

var unmarshaller = map[reflect.Kind]func(buffer []byte) (any, error){
	reflect.Float32:   unmarshalFloat32,
	reflect.String:    unmarshalString,
	reflect.Int:       unmarshalInt,
	reflect.Interface: unmarshalInterface,
}

func Unmarshal(buffer []byte, dataType reflect.Kind) (any, error) {
	return unmarshaller[dataType](buffer)
}

func unmarshalInterface(buffer []byte) (any, error) {
	var value interface{}
	err := json.Unmarshal(buffer, &value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func unmarshalInt(buffer []byte) (any, error) {
	var value int
	err := json.Unmarshal(buffer, &value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func unmarshalFloat32(buffer []byte) (any, error) {
	var value float32
	err := json.Unmarshal(buffer, &value)
	if err != nil {
		return 0, err
	}

	return value, nil
}
func unmarshalString(buffer []byte) (any, error) {
	var value string
	err := json.Unmarshal(buffer, &value)
	if err != nil {
		return "", err
	}

	return value, nil
}
