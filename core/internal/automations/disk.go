package automations

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
	"sync"
)

const (
	automationDir = "configs/automations"
	automationExt = ".config"
)

type DiskRepository struct {
	automations map[string]*Device
	mutex       sync.RWMutex
}

func NewDiskRepository() Repository {

	return &DiskRepository{
		automations: map[string]*Device{},
		mutex:       sync.RWMutex{},
	}
}

func (d *DiskRepository) Store(id string, automation *Device) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.automations[automation.Id] = automation

	// store to file

	saveFile(automation, automation.Id, true)
}

func (d *DiskRepository) Find(id string) (*Device, error) {

	return nil, nil
}

func (a *DiskRepository) clear() {
	for k := range a.automations {
		delete(a.automations, k)
	}
}

func (d *DiskRepository) Load() []*Device {
	d.clear()
	triggers := []*Device{}
	err := filepath.Walk(automationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			utils.LogErrorf("Error loading automations %s", err.Error())
			return err
		}
		if info.IsDir() {
			return nil
		}

		trigger, err := load(path)
		if err != nil {
			utils.LogErrorf("Error loading automation %s %s", path, err.Error())
			return err
		}

		d.automations[trigger.Id] = trigger

		triggers = append(triggers, trigger)
		return nil
	})

	if err != nil {
		utils.LogErrorf("Error loading automations %s", err.Error())
	}
	return triggers
}

func (d *DiskRepository) Delete(id string) error {
	if _, ok := d.automations[id]; !ok {

		utils.LogErrorf("delete automation id %s failed. Error=automation not found", id)
		return fmt.Errorf("automation %s not found", id)
	}

	delete(d.automations, id)

	// delete file
	deleteFile(id)
	return nil
}

func saveFile(automation *Device, name string, pretty bool) error {

	//sanitize
	name = strings.Replace(name, " ", "_", -1)
	filePath := getFilePath(name)

	data, err := json.Marshal(automation)
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

func getFilePath(name string) string {

	createDirIfNotExists(automationDir)
	return filepath.Join(automationDir, fmt.Sprintf("%s%s", name, automationExt))
}

func prettyJson(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func SaveAutomations(triggers []*Device) {
	for _, trigger := range triggers {
		saveFile(trigger, trigger.FriendlyName, true)
	}
}

func LoadTrigger(name string) (*Device, error) {
	filePath := getFilePath(name)

	return load(filePath)
}

func load(filePath string) (*Device, error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	t := &Device{}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
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

func deleteFile(name string) error {
	// sanitize
	name = strings.Replace(name, " ", "_", -1)
	filePath := getFilePath(name)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
