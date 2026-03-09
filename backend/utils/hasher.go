package utils

import (
	"fmt"
	"hash/crc32"
	"strings"
)

var crc32q *crc32.Table = crc32.MakeTable(0xD5828281)

func HashData(data []byte) string {
	return fmt.Sprintf("%08x", crc32.Checksum(data, crc32q))
}

func HashName(data string) string {
	sanitizedValue := strings.ReplaceAll(data, " ", "_")
	return fmt.Sprintf("%08x", crc32.Checksum([]byte(sanitizedValue), crc32q))
}
