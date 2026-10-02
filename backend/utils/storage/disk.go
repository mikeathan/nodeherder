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
	cache   map[string]T
	mutex   sync.RWMutex
	creator func() T
	loader  func(data []byte) (T, error)
}

func NewJsonDiskStorage[T any](baseDir string, ctor func() T, loader func(data []byte) (T, error)) Storage[T] {
	d := new(JsonDiskStorage[T])
	d.rootDir = baseDir
	d.cache = map[string]T{}
	d.mutex = sync.RWMutex{}
	d.creator = ctor
	d.loader = loader

	return d
}

func (d *JsonDiskStorage[T]) Initialize() ([]T, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	next := make(map[string]T)
	err := filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			utils.LogErrorf("Error loading item %s", err.Error())
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ext {
			return nil
		}

		item, err := d.loadFile(path)
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}

		name := filenameWithoutExtension(path)
		next[name] = item
		return nil
	})

	if err != nil {
		return nil, err
	}
	d.cache = next
	return d.findAll(), nil
}

func (d *JsonDiskStorage[T]) LoadAll() []T {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.findAll()
}

func (d *JsonDiskStorage[T]) findAll() []T {
	keys := make([]string, 0, len(d.cache))
	values := make([]T, 0, len(d.cache))

	for k := range d.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values = append(values, d.cache[k])
	}

	return values
}
func (d *JsonDiskStorage[T]) Delete(name string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := d.deleteFile(name)
	if err != nil {
		return err
	}

	d.deleteFromCache(name)
	return nil
}

func (d *JsonDiskStorage[T]) ClearCache() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.deleteCache()
}

func (d *JsonDiskStorage[T]) deleteCache() {

	for k := range d.cache {
		delete(d.cache, k)
	}
}
func (d *JsonDiskStorage[T]) Store(name string, item T) error {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := d.saveFile(item, name, true)
	if err != nil {
		return err
	}

	d.addToCache(name, item)
	return nil
}

func (d *JsonDiskStorage[T]) LoadFromCache(name string) (T, error) {

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	item, ok := d.loadFromCache(name)
	if ok {
		return item, nil
	}
	//utils.LogDebugf("item %s not in cache", name)
	return zeroValue[T](), fmt.Errorf("item %s not in cache", name)
}

func (d *JsonDiskStorage[T]) Load(name string) (T, error) {

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	item, ok := d.loadFromCache(name)
	if ok {
		return item, nil
	}

	filePath := d.getFilePath(name)
	item, err := d.loadFile(filePath)
	if err != nil {
		utils.LogInfof("Error loading item %s %s", name, err.Error())
		return item, err
	}

	return item, nil
}

func (d *JsonDiskStorage[T]) addToCache(name string, item T) {
	d.cache[name] = item
}

func (d *JsonDiskStorage[T]) loadFromCache(name string) (T, bool) {
	if item, ok := d.cache[name]; ok {
		return item, true
	}

	return zeroValue[T](), false
}

func (d *JsonDiskStorage[T]) deleteFromCache(name string) {
	delete(d.cache, name)
}

func (d *JsonDiskStorage[T]) saveFile(item T, name string, pretty bool) error {
	if err := os.MkdirAll(d.rootDir, os.ModePerm); err != nil {
		return fmt.Errorf("create recipe directory: %w", err)
	}
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

	return replaceJSONFile(filePath, data)
}

// Stage beside the destination so rename is atomic on the same filesystem.
// Preserve existing permissions and valid symlink targets, as direct writes did.
func replaceJSONFile(filePath string, data []byte) error {
	if info, err := os.Lstat(filePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(filePath)
		if err != nil {
			return err
		}
		filePath = resolved
	}
	mode := os.FileMode(0644)
	if info, err := os.Stat(filePath); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(filePath), ".automation-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if err := temp.Chmod(mode); err != nil {
		return err
	}
	if _, err := temp.Write(data); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), filePath)
}

func (d *JsonDiskStorage[T]) getFilePath(name string) string {
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

func (d *JsonDiskStorage[T]) loadFile(filePath string) (T, error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {

		return zeroValue[T](), err
	}
	defer jsonFile.Close()

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return zeroValue[T](), err
	}
	if d.loader != nil {
		return d.loader(data)
	}

	item := d.creator()
	if err := json.Unmarshal(data, &item); err != nil {
		return zeroValue[T](), err
	}
	return item, nil
}

func zeroValue[T any]() T {
	var zero T
	return zero
}

func (d *JsonDiskStorage[T]) deleteFile(name string) error {
	// sanitize
	filePath := d.getFilePath(name)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
