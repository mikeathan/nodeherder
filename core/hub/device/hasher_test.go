package device_test

import (
	"node-herder/hub/device"
	"testing"
)

func getTestData() []any {

	return []any{1, 2.1234, "test 1", "@test sth * data", ""}
}

func TestHasher(t *testing.T) {

	h1 := device.NewCrc32Hasher()
	for _, v := range getTestData() {
		h1.Write(v)
	}

	got1 := h1.Hash()
	if got1 == "" {
		t.Fatalf("got %s", got1)
	}

	h2 := device.NewCrc32Hasher()
	for _, v := range getTestData() {
		h2.Write(v)
	}

	got2 := h1.Hash()
	if got2 == "" {
		t.Fatalf("got %s", got2)
	}

	if got1 != got2 {
		t.Fatalf("got %s want %s", got1, got2)
	}
}
