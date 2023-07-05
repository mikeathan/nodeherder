package device

import (
	"bytes"
	"fmt"
	"hash/crc32"
)

type Crc32Hasher struct {
	buf bytes.Buffer
}

func NewCrc32Hasher() *Crc32Hasher {
	return &Crc32Hasher{
		buf: bytes.Buffer{},
	}
}
func (h *Crc32Hasher) Write(value any) {
	fmt.Fprintf(&h.buf, "%v", value)
}

func (h *Crc32Hasher) CalculateHash() string {
	b := h.buf.Bytes()
	if len(b) != 0 {
		crc32q := crc32.MakeTable(0xD5828281)
		return fmt.Sprintf("%08x", crc32.Checksum(b, crc32q))
	}

	return ""
}

func (h *Crc32Hasher) Reset() {
	h.buf.Reset()
}

// func xor(a []byte, b []byte) []byte {
// 	c := make([]byte, len(a))
// 	for i := range a {
// 		c[i] = a[i] ^ b[i]
// 	}
// 	return c
// }
