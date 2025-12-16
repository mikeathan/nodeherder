package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"strings"
)

func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func ByteArrayToAny(data []byte, v any) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

func AnyToByteArray(v any) ([]byte, error) {

	if b, ok := v.(bool); ok {
		v = fmt.Sprintf("%v", b)
	}

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, fmt.Errorf("error encoding value: %w", err)
	}
	return buf.Bytes(), nil
}

func ParseBool(v any) (bool, bool) {
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
