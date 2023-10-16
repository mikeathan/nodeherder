package storage_test

import (
	"node-herder/utils/storage"
	"testing"
)

type testItem struct {
	Value int `json:"value"`
}

func TestStorage(t *testing.T) {
	disk := storage.NewJsonDiskStorage[testItem]("temp")
	var item = &testItem{Value: 1}
	err := disk.Store("1", item)
	if err != nil {
		t.Errorf("store failed %s", err.Error())
	}
	result, err := disk.Load("1")
	if err != nil {
		t.Errorf("load failed %s", err.Error())
	}

	if result.Value != item.Value {
		t.Errorf("item mismatch want %v got %d", item.Value, result.Value)
	}
	var item2 = &testItem{Value: 2}
	err = disk.Store("2", item2)
	if err != nil {
		t.Errorf("store failed %s", err.Error())
	}

	allItems, err := disk.LoadAll()
	if err != nil {
		t.Errorf("LoadAll failed %s", err.Error())
	}

	for idx, a := range allItems {
		switch idx {
		case 0:
			if a.Value != item.Value {
				t.Errorf("item1 mismatch want %v got %d", item.Value, a.Value)
			}
		case 1:
			if a.Value != item2.Value {
				t.Errorf("item2 mismatch want %v got %d", item.Value, a.Value)
			}
		}

	}
	err = disk.Delete("1")
	if err != nil {
		t.Errorf("delete file failed %s", err.Error())
	}

	_, err = disk.Load("1")
	if err == nil {
		t.Errorf("file 1 still exists. error %s", err.Error())
	}

	err = disk.Delete("2")
	if err != nil {
		t.Errorf("delete file 2 failed %s", err.Error())
	}

	_, err = disk.Load("2")
	if err == nil {
		t.Errorf("file 2 still exists. error %s", err.Error())
	}
}

func TestFindAll(t *testing.T) {

}
