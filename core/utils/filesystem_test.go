package utils_test

import (
	"node-herder/mocks"
	"node-herder/utils"
	"path/filepath"
	"testing"
)

func TestListFileLogsShouldMatchExtension(t *testing.T) {

	mockeFiles := []string{"nodeherder.log", "nodeherder2.log", "nodeherder3.log", "nodeherder4.log"}

	walker := mocks.NewMockWalker(mockeFiles)
	files, err := utils.ListFilesWithExtension(walker, utils.LogsPath, filepath.Ext(utils.LogName))

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
	files, err := utils.ListFilesWithExtension(walker, utils.LogsPath, filepath.Ext(utils.LogName))

	if err != nil {
		t.Errorf("Failed to list logs: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(files))
	}

}
