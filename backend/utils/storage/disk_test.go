package storage_test

import (
	"encoding/json"
	"fmt"
	"node-herder/utils/storage"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type testItem struct {
	Id           string `json:"id"`
	Value        int    `json:"value"`
	internalData map[string]int
}

func ctr() testItem {
	return testItem{internalData: make(map[string]int)}
}

func TestDiskReloadFailurePreservesCache(t *testing.T) {
	for _, missingRoot := range []bool{false, true} {
		t.Run(fmt.Sprint(missingRoot), func(t *testing.T) {
			root := t.TempDir()
			disk := storage.NewJsonDiskStorage(root, ctr, nil)
			if err := disk.Store("known", newTestitem("known", 42)); err != nil {
				t.Fatal(err)
			}
			if missingRoot {
				if err := os.Rename(root, root+"-moved"); err != nil {
					t.Fatal(err)
				}
				defer os.Rename(root+"-moved", root)
			} else {
				if err := os.WriteFile(filepath.Join(root, "known.json"), []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := disk.Initialize(); err == nil {
				t.Fatal("failed disk scan reported success")
			}
			got, err := disk.LoadFromCache("known")
			if err != nil || got.Value != 42 {
				t.Fatalf("failed reload discarded old cache: %v %v", got, err)
			}
		})
	}
}

func TestDiskSaveNeverPublishesPartialJSON(t *testing.T) {
	root := t.TempDir()
	disk := storage.NewJsonDiskStorage[map[string]interface{}](root, nil, nil)
	item := map[string]interface{}{"data": strings.Repeat("x", 65536)}
	if err := disk.Store("known", item); err != nil {
		t.Fatal(err)
	}
	started, stop, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		var once sync.Once
		for {
			select {
			case <-stop:
				return
			default:
			}
			data, err := os.ReadFile(filepath.Join(root, "known.json"))
			if err != nil {
				t.Error(err)
				once.Do(func() { close(started) })
				return
			}
			var value map[string]interface{}
			if err := json.Unmarshal(data, &value); err != nil {
				t.Errorf("partially written JSON visible: %v", err)
				once.Do(func() { close(started) })
				return
			}
			once.Do(func() { close(started) })
		}
	}()
	<-started
	defer func() { close(stop); <-done }()
	for i := 0; i < 100; i++ {
		item["version"] = i
		if err := disk.Store("known", item); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDiskSavePreservesPermissionsAndFailedEncoding(t *testing.T) {
	root := t.TempDir()
	disk := storage.NewJsonDiskStorage[map[string]interface{}](root, nil, nil)
	if err := disk.Store("known", map[string]interface{}{"value": 1}); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "known.json")
	if err := os.Chmod(file, 0600); err != nil {
		t.Fatal(err)
	}
	if err := disk.Store("known", map[string]interface{}{"value": 2}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := disk.Store("known", map[string]interface{}{"bad": make(chan struct{})}); err == nil {
		t.Fatal("unsupported encoding accepted")
	}
	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("failed encoding overwrote recipe")
	}
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions changed: %v", info.Mode())
	}
	items, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("temporary files leaked: %v", items)
	}
	got, err := disk.LoadFromCache("known")
	if err != nil || got["value"] != 2 {
		t.Fatalf("failed encoding changed cache: %v %v", got, err)
	}
}

func TestDiskSavePreservesSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	disk := storage.NewJsonDiskStorage(root, ctr, nil)
	if err := disk.Store("target", newTestitem("target", 1)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target.json", filepath.Join(root, "linked.json")); err != nil {
		t.Fatal(err)
	}
	if err := disk.Store("linked", newTestitem("target", 2)); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filepath.Join(root, "linked.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("save replaced symlink instead of target")
	}
	data, err := os.ReadFile(filepath.Join(root, "target.json"))
	if err != nil {
		t.Fatal(err)
	}
	var item testItem
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatal(err)
	}
	if item.Value != 2 {
		t.Fatal("symlink target not updated")
	}
}

func TestDiskFailedReplacementCleansTemporaryFiles(t *testing.T) {
	root := t.TempDir()
	disk := storage.NewJsonDiskStorage(root, ctr, nil)
	if err := os.Mkdir(filepath.Join(root, "blocked.json"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := disk.Store("blocked", newTestitem("blocked", 1)); err == nil {
		t.Fatal("failed rename reported success")
	}
	if _, err := disk.LoadFromCache("blocked"); err == nil {
		t.Fatal("failed rename updated cache")
	}
	items, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name() != "blocked.json" || !items[0].IsDir() {
		t.Fatalf("failed rename leaked staged file: %v", items)
	}
}

func TestDiskDirectoryCreationFailureReturnsError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "unavailable")
	// A dangling directory symlink makes MkdirAll fail without depending on UID
	// or read-only permission behavior (CI containers may run as root).
	if err := os.Symlink("missing-target", root); err != nil {
		t.Fatal(err)
	}
	disk := storage.NewJsonDiskStorage(root, ctr, nil)
	if err := disk.Store("known", newTestitem("known", 1)); err == nil {
		t.Fatal("directory creation failure accepted")
	}
	if _, err := disk.LoadFromCache("known"); err == nil {
		t.Fatal("failed directory creation changed cache")
	}
}

func newTestitem(id string, value int) testItem {
	return testItem{Id: id, Value: value, internalData: make(map[string]int)}
}

func createDiskStorage() storage.Storage[testItem] {
	ctr := func() testItem {
		return ctr()
	}
	return storage.NewJsonDiskStorage("temp", ctr, nil)
}

func TestInitializeFromDisk(t *testing.T) {

	// add some files in root dir
	items := []testItem{}
	items = append(items, newTestitem("some file 1", 1))
	items = append(items, newTestitem("some file 2", 2))

	disk := createDiskStorage()

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
		result.internalData[fmt.Sprint(idx)] = 1
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
	items := []testItem{}
	items = append(items, newTestitem("test1", 1))
	items = append(items, newTestitem("test2", 2))
	disk := createDiskStorage()

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
	items := []testItem{}
	items = append(items, newTestitem("test5", 5))
	items = append(items, newTestitem("test6", 6))
	disk := createDiskStorage()

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
