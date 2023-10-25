package storage_test

import (
	"fmt"
	"node-herder/utils/storage"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testItem struct {
	Id           string `json:"id"`
	Value        int    `json:"value"`
	internalData map[string]int
}

func (t *testItem) Initialize() {
	t.internalData = make(map[string]int)
}

func TestInitializeFromDisk(t *testing.T) {

	// add some files in root dir
	items := []*testItem{}
	items = append(items, &testItem{Id: "some file 1", Value: 1})
	items = append(items, &testItem{Id: "some file 2", Value: 2})
	disk := storage.NewJsonDiskStorage[testItem]("temp")

	for _, item := range items {
		err := disk.Store(item.Id, item)
		if err != nil {
			t.Errorf("store file %s failed %s", item.Id, err.Error())
		}
	}

	time.Sleep(100 * time.Millisecond)

	// will clear cache and load all files in root folder from disk
	allItems, err := disk.Initialize()
	if err != nil {
		t.Errorf("initialize failed %s", err.Error())
	}
	if len(allItems) != len(items) {
		t.Errorf("initialize did not find any files. want: %d files. got: %d files", len(items), len(allItems))
	}

	for idx, result := range allItems {

		item := items[idx]
		// do some caching to make sure internal datastructures have been initialized
		item.internalData[fmt.Sprint(idx)] = 1
		//
		if result.Id != item.Id {
			t.Errorf("item %d mismatch want %v got %v", idx, item.Id, result.Id)
		}

		if result.Value != item.Value {
			t.Errorf("item %d mismatch want %v got %v", idx, item.Value, result.Value)
		}
	}

	// delete all
	for _, item := range items {
		err = disk.Delete(item.Id)
		if err != nil {
			t.Errorf("delete file %s failed %s", item.Id, err.Error())
		}
	}

	// make sure root dir is clear
	allItems, err = disk.Initialize()
	if err != nil {
		t.Errorf("initialize failed. Error %s", err.Error())
	}
	if len(allItems) != 0 {
		t.Errorf("initialize found files. want: 0 files. got: %d files", len(allItems))
	}
}

func TestLoadingFromDisk(t *testing.T) {
	items := []*testItem{}
	items = append(items, &testItem{Id: "test1", Value: 1})
	items = append(items, &testItem{Id: "test2", Value: 2})
	disk := storage.NewJsonDiskStorage[testItem]("temp")

	for _, item := range items {
		err := disk.Store(item.Id, item)
		if err != nil {
			t.Errorf("store failed %s", err.Error())
		}
	}

	time.Sleep(100 * time.Millisecond)

	disk.ClearCache()

	for _, item := range items {
		result, err := disk.Load(item.Id)
		if err != nil {
			t.Errorf("load file %s failed %s", item.Id, err.Error())
		}

		if result.Id != item.Id {
			t.Errorf("item.Id  mismatch want %v got %v", item.Id, result.Id)
		}

		if result.Value != item.Value {
			t.Errorf("item.Value mismatch want %v got %d", item.Value, result.Value)
		}
	}

	// delete all
	for _, item := range items {
		err := disk.Delete(item.Id)
		if err != nil {
			t.Errorf("delete file %s failed %s", item.Id, err.Error())
		}
	}

	// make sure root dir is clear
	allItems, err := disk.Initialize()
	if err != nil {
		t.Errorf("initialize failed. Error %s", err.Error())
	}
	if len(allItems) != 0 {
		t.Errorf("initialize found files. want: 0 files. got: %d files", len(allItems))
	}

}

func TestLoadingFromCache(t *testing.T) {
	items := []*testItem{}
	items = append(items, &testItem{Id: "test5", Value: 5})
	items = append(items, &testItem{Id: "test6", Value: 6})
	disk := storage.NewJsonDiskStorage[testItem]("temp")

	for _, item := range items {
		err := disk.Store(item.Id, item)
		if err != nil {
			t.Errorf("store failed %s", err.Error())
		}
	}

	time.Sleep(100 * time.Millisecond)

	// manually delete files
	for _, item := range items {
		name := strings.Replace(item.Id, " ", "_", -1)

		filePath := filepath.Join("temp", fmt.Sprintf("%s%s", name, ".json"))

		err := os.Remove(filePath)
		if err != nil {
			t.Errorf("delete file %s failed %s", item.Id, err.Error())
		}
	}

	// load from cache
	for _, item := range items {
		result, err := disk.Load(item.Id)
		if err != nil {
			t.Errorf("load file %s failed %s", item.Id, err.Error())
		}

		if result.Id != item.Id {
			t.Errorf("item.Id  mismatch want %v got %v", item.Id, result.Id)
		}

		if result.Value != item.Value {
			t.Errorf("item.Value mismatch want %v got %d", item.Value, result.Value)
		}
	}

	// make sure root dir is clear, clears cache so we need to do this last
	allItems, err := disk.Initialize()
	if err != nil {
		t.Errorf("initialize failed. Error %s", err.Error())
	}
	if len(allItems) != 0 {
		t.Errorf("initialize found files. want: 0 files. got: %d files", len(allItems))
	}

}
