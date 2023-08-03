package utils

import (
	"errors"
	"fmt"
)

var nodeIds = []string{"node_id", "nodeid", "id", "nickname", "label", "name"}
var batchPayloadIds = []string{"readings", "items", "data", "items"}

func FindPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	for key, value := range payload {
		payloadMap, ok := value.(map[string]interface{})
		if ok {
			if !contains(batchPayloadIds, key) {
				return nil, errors.New("invalid data: data structure not containign valid root payload")
			}
			timestamp, ok := payload["timestamp"]
			if ok {
				payloadMap["timestamp"] = timestamp
			}
			return payloadMap, nil
		}

	}
	return payload, nil
}

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
