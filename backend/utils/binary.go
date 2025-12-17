package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func AnyToByteArray(v any) ([]byte, error) {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("error encoding value: %w", err)
	}
	return buf.Bytes(), nil
}
