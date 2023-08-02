package utils

import (
	"errors"
	"fmt"
)

var nodeIds = []string{"node_id", "nodeid", "id", "nickname", "label", "name"}

func FindId(payload map[string]interface{}) (string, error) {
	for key, value := range payload {
		if contains(nodeIds, key) {
			return fmt.Sprint(value), nil
		}
	}
	return "", errors.New("invalid data: data structure is missing node id")
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
