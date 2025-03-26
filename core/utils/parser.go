package utils

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

var nodeIds = []string{"nickname", "label", "name"}
var batchPayloadIds = []string{"readings", "items", "data", "items"}
var whitelistNames = []string{"timestamp", "last_seen", "label", "name", "nickname"}

func ParsePayload(payload map[string]interface{}) (string, map[string]interface{}, error) {

	id, err := findId(payload)
	if err != nil {
		return ",", nil, err
	}

	for key, value := range payload {
		payloadMap, ok := value.(map[string]interface{})
		if ok {
			if !contains(batchPayloadIds, key) {
				return "", nil, errors.New("invalid data: data structure not containing valid payload section")
			}

			timestamp, ok := payload["timestamp"]
			if ok {
				timestamp, err := convertTimestamp(timestamp)
				if err == nil {
					payloadMap["last_seen"] = timestamp
				}
			}

			payloadMap = sanitizeLegacyPayload(payloadMap)
			return id, payloadMap, nil
		}
	}

	payload = sanitizeLegacyPayload(payload)
	return id, payload, nil
}

func sanitizeLegacyPayload(payload map[string]interface{}) map[string]interface{} {

	for key, value := range payload {
		if contains(whitelistNames, key) {
			continue
		}
		strVal, ok := value.(string)
		if !ok {
			continue
		}

		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		extracted := re.Find([]byte(strVal))
		sanitizedInt, err := strconv.Atoi(string(extracted))

		if err != nil {
			sanitizedFloat, err := strconv.ParseFloat(string(extracted), 8)
			if err != nil {
				LogDebugf("Error during conversion of %s in legacy payload", string(extracted))
			}

			payload[key] = sanitizedFloat
			continue
		}
		payload[key] = sanitizedInt
	}

	return payload
}

func convertTimestamp(timestamp interface{}) (string, error) {
	timestampStr, ok := timestamp.(string)
	if !ok {
		return "", errors.New("error - invalid timestamp format")
	}
	converted, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return "", err
	}
	return converted.Format(time.RFC3339), nil
}

func findId(payload map[string]interface{}) (string, error) {
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

func ComparePayloadValues(a interface{}, b interface{}) bool {
	if isPrimitive(a) && isPrimitive(b) {
		return a == b
	}

	return reflect.DeepEqual(a, b)
}

func isPrimitive(data interface{}) bool {
	switch data.(type) {
	case string, int, int64, int32, float64, float32, bool, uint, uint64, uint32:
		return true
	}
	return false
}
