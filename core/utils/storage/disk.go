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

const ext = ".json"

type JsonDiskStorage[T any] struct {
	rootDir string
}

func NewJsonDiskStorage[T any](baseDir string) Storage[T] {
	d := new(JsonDiskStorage[T])
	d.rootDir = baseDir
	return d
}

func (d *JsonDiskStorage[T]) LoadAll() ([]*T, error) {
	items := []*T{}
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

		items = append(items, item)
		return nil
	})

	if err != nil {
		utils.LogErrorf("Error loading items %s", err.Error())
	}
	return items, nil
}

func (d *JsonDiskStorage[T]) Delete(id string) error {

	return d.deleteFile(id)
}

func (d *JsonDiskStorage[T]) Store(name string, item *T) error {

	d.saveFile(item, name, true)
	return nil
}

func (d *JsonDiskStorage[T]) Load(name string) (*T, error) {

	filePath := d.getFilePath(name)

	item, err := d.loadFile(filePath)
	if err != nil {
		utils.LogErrorf("Error loading item %s %s", name, err.Error())
		return nil, err
	}

	return item, nil
}

func (d *JsonDiskStorage[T]) saveFile(item *T, name string, pretty bool) error {

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

func (d *JsonDiskStorage[T]) getFilePath(name string) string {

	createDirIfNotExists(d.rootDir)
	return filepath.Join(d.rootDir, fmt.Sprintf("%s%s", name, ext))
}

func prettyJson(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
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

func (d *JsonDiskStorage[T]) deleteFile(name string) error {
	// sanitize
	name = strings.Replace(name, " ", "_", -1)
	filePath := d.getFilePath(name)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
