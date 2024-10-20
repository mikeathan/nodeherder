package utils

import (
	"os"
	"path/filepath"
)

type Loader interface {
	Load(file string) ([]byte, error)
}

type FileSystemLoader struct{}

func (FileSystemLoader) Load(file string) ([]byte, error) {
	return os.ReadFile(file)
}
func LoadFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

type Walker interface {
	Walk(root string, walkFn filepath.WalkFunc) error
}

type FileSystemWalker struct{}

func (FileSystemWalker) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func ListFilesWithExtension(walker Walker, root string, extension string) ([]string, error) {
	var files []string
	err := walker.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == extension {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}

func ListFiles(walker Walker, root string) ([]string, error) {
	var files []string
	err := walker.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
