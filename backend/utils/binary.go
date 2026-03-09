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
