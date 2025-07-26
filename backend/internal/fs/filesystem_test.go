package fs_test

import (
	"node-herder/internal/fs"
	"node-herder/mocks"
	"node-herder/utils"
	"path/filepath"
	"testing"
	"time"
)

func TestListFileLogsShouldMatchExtension(t *testing.T) {

	mockeFiles := []string{"nodeherder.log", "nodeherder2.log", "nodeherder3.log", "nodeherder4.log"}

	walker := mocks.NewMockWalker(mockeFiles)
	files, err := fs.ListFilesWithExtension(walker, utils.LogsPath, filepath.Ext(utils.LogName))

	if err != nil {
		t.Errorf("Failed to list logs: %v", err)
	}

	if len(files) == 0 {
		t.Errorf("Expected %d logs, got %d", len(mockeFiles), len(files))
	}

	for id, file := range files {
		if file != mockeFiles[id] {
			t.Errorf("Expected %s, got %s", mockeFiles[id], file)
		}
	}
}

func TestListFileLogsShouldNotMatchExtension(t *testing.T) {

	mockeFiles := []string{"nodeherder.txt", "nodeherder2.t", "nodeherder3.p", "nodeherder4.z"}

	walker := mocks.NewMockWalker(mockeFiles)
	files, err := fs.ListFilesWithExtension(walker, utils.LogsPath, filepath.Ext(utils.LogName))

	if err != nil {
		t.Errorf("Failed to list logs: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(files))
	}
}

func TestFileSystemExpiration(t *testing.T) {

	mockFiles := []string{"nodeherder.txt"}
	mockFile := []byte("test 1")

	fileCallback := func() []byte {
		return mockFile
	}

	walker := mocks.NewMockWalker(mockFiles)
	loader := mocks.NewMockFileLoaderWithCallback(fileCallback)
	fileSystem := fs.NewFileSystem(
		fs.WithFileWalker(walker),
		fs.WithFileLoader(loader),
		fs.WithExpiration(100*time.Millisecond))

	fsBuffer, err := fileSystem.Load(mockFiles[0])

	if err != nil {
		t.Errorf("Failed to load file: %v", err)
	}

	if string(fsBuffer) != string([]byte("test 1")) {
		t.Errorf("Expected %s, got %s", string(mockFile), string(fsBuffer))
	}

	mockFile = []byte("test 2")

	time.Sleep(150 * time.Millisecond)
	fsBuffer, _ = fileSystem.Load(mockFiles[0])

	if string(fsBuffer) != string([]byte("test 2")) {
		t.Errorf("Expected %s, got %s", string(mockFile), string(fsBuffer))
	}
}
