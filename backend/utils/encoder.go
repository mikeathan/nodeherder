package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
)

func DecodeGobValue(data []byte, v any) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

func ConvertToFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint64:
		return float64(x), true
	case uint32:
		return float64(x), true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func ConvertToBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true

	case string:
		switch strings.ToLower(strings.TrimSpace(x)) {
		case "1", "true", "on", "open", "yes":
			return true, true
		case "0", "false", "off", "closed", "no":
			return false, true
		default:
			return false, false
		}

	case int:
		return x != 0, true

	case int64:
		return x != 0, true

	case float64:
		return x != 0, true
	}

	return false, false
}

func ConvertToString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case []byte:
		return string(x), true
	case bool:
		if x {
			return "true", true
		}
		return "false", true
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", x), true
	}
	return "", false
}

func DecodeBinaryValue(raw []byte) (bool, error) {
	var boolVal bool
	if err := DecodeGobValue(raw, &boolVal); err == nil {
		return boolVal, nil
	}

	var strVal string
	if err := DecodeGobValue(raw, &strVal); err == nil {
		if v, ok := ConvertToBool(strVal); ok {
			return v, nil
		}
		return false, fmt.Errorf("failed to parse bool value: %v", strVal)
	}

	var floatVal float64
	if err := DecodeGobValue(raw, &floatVal); err == nil {
		return floatVal != 0, nil
	}

	var intVal int64
	if err := DecodeGobValue(raw, &intVal); err == nil {
		return intVal != 0, nil
	}

	return false, fmt.Errorf("unsupported binary metric encoding")
}

func DecodeNumericValue(raw []byte) (float64, error) {
	var floatVal float64
	if err := DecodeGobValue(raw, &floatVal); err == nil {
		return floatVal, nil
	}

	var float32Val float32
	if err := DecodeGobValue(raw, &float32Val); err == nil {
		return float64(float32Val), nil
	}

	var intVal int64
	if err := DecodeGobValue(raw, &intVal); err == nil {
		return float64(intVal), nil
	}

	var int32Val int32
	if err := DecodeGobValue(raw, &int32Val); err == nil {
		return float64(int32Val), nil
	}

	var uintVal uint64
	if err := DecodeGobValue(raw, &uintVal); err == nil {
		return float64(uintVal), nil
	}

	return 0, fmt.Errorf("unsupported numeric metric encoding")
}

func DecodeEnumMetricValue(raw []byte) (string, error) {
	var strVal string
	if err := DecodeGobValue(raw, &strVal); err == nil {
		return strVal, nil
	}

	return "", fmt.Errorf("unsupported enum metric encoding")
}
