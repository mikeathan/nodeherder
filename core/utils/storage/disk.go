package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"node-herder/utils"
	"os"
	"path/filepath"
	"strings"
)

type DiskStorage[T any] struct {
	rootDir string
}

// func NewDiskStorage(baseDir string) Storage[T] {
// return &DiskStorage{}
// }
func (d *DiskStorage[T]) LoadAll() []T {
	items := []T{}
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
			utils.LogErrorf("Error loading item %s %s", path, err.Error())
			return err
		}

		items = append(items, *item)
		return nil
	})

	if err != nil {
		utils.LogErrorf("Error loading items %s", err.Error())
	}
	return items
}

func (d *DiskStorage[T]) Delete(id string) error {

	return d.deleteFile(id)
}

func (d *DiskStorage[T]) Store(id string, item T) error {

	d.saveFile(item, id, true)
	return nil
}

func (d *DiskStorage[T]) Load(id string) (T, error) {

	filePath := d.getFilePath(id)

	item, err := d.loadFile(filePath)
	if err != nil {
		utils.LogErrorf("Error loading item %s %s", id, err.Error())
		var empty T
		return empty, err
	}

	return *item, nil
}

func (d *DiskStorage[T]) saveFile(item T, name string, pretty bool) error {

	//sanitize
	name = strings.Replace(name, " ", "_", -1)
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

func (d *DiskStorage[T]) getFilePath(name string) string {

	createDirIfNotExists(d.rootDir)
	return filepath.Join(d.rootDir, fmt.Sprintf("%s%s", name, d.rootDir))
}

func prettyJson(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func (d *DiskStorage[T]) loadFile(filePath string) (*T, error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {

		return nil, err
	}

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	item := new(T)
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

func (d *DiskStorage[T]) deleteFile(name string) error {
	// sanitize
	name = strings.Replace(name, " ", "_", -1)
	filePath := d.getFilePath(name)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
