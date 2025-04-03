package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"node-herder/utils"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const ext = ".json"

type JsonDiskStorage[T any] struct {
	rootDir string
	cache   map[string]*T
	mutex   sync.RWMutex
	creator func() *T
}

func NewJsonDiskStorage[T any](baseDir string, ctor func() *T) Storage[T] {
	d := new(JsonDiskStorage[T])
	d.rootDir = baseDir
	d.cache = map[string]*T{}
	d.mutex = sync.RWMutex{}
	d.creator = ctor

	return d
}

func (d *JsonDiskStorage[T]) Initialize() ([]*T, error) {
	defer d.mutex.Unlock()
	d.mutex.Lock()

	d.deleteCache()

	err := filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			utils.LogErrorf("Error loading item %s", err.Error())
			return err
		}
		if info.IsDir() {
			return nil
		}

		item, err := d.loadFile(path)
		if err != nil {
			// we dont want to return error as it will stop loading next item
			utils.LogErrorf("Error loading item %s %s", path, err.Error())
			return nil
		}

		name := filenameWithoutExtension(path)
		d.addToCache(name, item)
		return nil
	})

	if err != nil {
		utils.LogErrorf("Error loading items %s", err.Error())
	}

	return d.findAll(), nil
}

func (d *JsonDiskStorage[T]) LoadAll() []*T {
	defer d.mutex.RUnlock()
	d.mutex.RLock()

	return d.findAll()
}

func (d *JsonDiskStorage[T]) findAll() []*T {
	keys := make([]string, 0, len(d.cache))
	values := make([]*T, 0, len(d.cache))

	for k, _ := range d.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values = append(values, d.cache[k])
	}

	return values
}
func (d *JsonDiskStorage[T]) Delete(name string) error {
	defer d.mutex.Unlock()
	d.mutex.Lock()

	err := d.deleteFile(name)
	if err != nil {
		return err
	}

	d.deleteFromCache(name)
	return nil
}

func (d *JsonDiskStorage[T]) ClearCache() {
	defer d.mutex.RUnlock()
	d.mutex.RLock()

	d.deleteCache()
}

func (d *JsonDiskStorage[T]) deleteCache() {

	for k := range d.cache {
		delete(d.cache, k)
	}
}
func (d *JsonDiskStorage[T]) Store(name string, item *T) error {

	defer d.mutex.Unlock()
	d.mutex.Lock()

	err := d.saveFile(item, name, true)
	if err != nil {
		return err
	}

	d.addToCache(name, item)
	return nil
}

func (d *JsonDiskStorage[T]) LoadFromCache(name string) (*T, error) {

	defer d.mutex.RUnlock()
	d.mutex.RLock()

	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}
	//utils.LogDebugf("item %s not in cache", name)
	return nil, fmt.Errorf("item %s not in cache", name)
}

func (d *JsonDiskStorage[T]) Load(name string) (*T, error) {

	defer d.mutex.RUnlock()
	d.mutex.RLock()

	item := d.loadFromCache(name)
	if item != nil {
		return item, nil
	}

	filePath := d.getFilePath(name)
	item, err := d.loadFile(filePath)
	if err != nil {
		utils.LogInfof("Error loading item %s %s", name, err.Error())
		return nil, err
	}

	return item, nil
}

func (d *JsonDiskStorage[T]) addToCache(name string, item *T) {
	name = sanitize(name)
	d.cache[name] = item
}

func (d *JsonDiskStorage[T]) loadFromCache(name string) *T {
	name = sanitize(name)
	if item, ok := d.cache[name]; ok {
		return item
	}

	return nil
}

func (d *JsonDiskStorage[T]) deleteFromCache(name string) {
	name = sanitize(name)
	delete(d.cache, name)
}

func (d *JsonDiskStorage[T]) saveFile(item *T, name string, pretty bool) error {

	//sanitize
	name = sanitize(name)
	filePath := d.getFilePath(name)

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	if pretty {
		data, err = prettyJson(data)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (d *JsonDiskStorage[T]) getFilePath(name string) string {

	createDirIfNotExists(d.rootDir)
	return filepath.Join(d.rootDir, fmt.Sprintf("%s%s", name, ext))
}

func prettyJson(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func filenameWithoutExtension(fullPath string) string {
	fileName := filepath.Base(fullPath)
	return strings.TrimSuffix(fileName, path.Ext(fileName))
}

func (d *JsonDiskStorage[T]) loadFile(filePath string) (*T, error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {

		return nil, err
	}

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	item := d.creator()
	err = json.Unmarshal(data, &item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func createDirIfNotExists(name string) {
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(name, os.ModePerm)
		if err != nil {
			utils.LogErrorf(fmt.Sprintf("Failed to create automations directory %s Error: %v", name, err))
			panic(err)
		}
	}
}

// func sanitize(name string) string {
// 	return strings.Replace(name, " ", "_", -1)
// }

func (d *JsonDiskStorage[T]) deleteFile(name string) error {
	// sanitize
	name = sanitize(name)
	filePath := d.getFilePath(name)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
