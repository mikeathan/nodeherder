package storage_test

import (
	"node-herder/utils/storage"
	"testing"
)

func TestStore(t *testing.T) {
	disk := storage.DiskStorage[int]{}
	disk.rootDir = ""
	var item = 1
	err := disk.Store("1", item)
	if err != nil {
		t.Errorf("store failed %s", err.Error())
	}
	result, err := disk.Load("1")
	if err != nil {
		t.Errorf("load failed %s", err.Error())
	}

	if result != item {
		t.Errorf("item mismatch want %v got %d", item, result)

	}
}
