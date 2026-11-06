package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"node-herder/internal/spatialhistory"
	"node-herder/utils"
)

type storageEntry struct {
	Label         string `json:"label"`
	Bytes         int64  `json:"bytes"`
	RetentionDays int    `json:"retentionDays,omitempty"`
}

type storageUsage struct {
	Entries        []storageEntry `json:"entries"`
	DataDirBytes   int64          `json:"dataDirBytes"`
	DiskFreeBytes  int64          `json:"diskFreeBytes"`
	DiskTotalBytes int64          `json:"diskTotalBytes"`
	MetricsDays    int            `json:"metricsRetentionDays"`
	PositionsDays  int            `json:"positionsRetentionDays"`
}

// NewSystemStorageHandler reports read-only data-directory usage. It never exposes file contents.
func NewSystemStorageHandler(metricsRetentionDays int) http.Handler {
	var mutex sync.Mutex
	var cached storageUsage
	var cachedAt time.Time
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		if time.Since(cachedAt) < 15*time.Second {
			usage := cached
			mutex.Unlock()
			writeStorageJSON(w, usage)
			return
		}
		mutex.Unlock()
		dataDir := utils.GetDataDir()
		usage := storageUsage{
			MetricsDays:   metricsRetentionDays,
			PositionsDays: int(spatialhistory.Retention.Hours() / 24),
			Entries: []storageEntry{
				{Label: "Metrics database", Bytes: fileBytes(filepath.Join(dataDir, "metrics.db")), RetentionDays: metricsRetentionDays},
				{Label: "Spatial history", Bytes: fileBytes(filepath.Join(dataDir, "spatial.db")), RetentionDays: int(spatialhistory.Retention.Hours() / 24)},
				{Label: "Logs", Bytes: dirBytes(utils.GetLogsDir())},
			},
		}
		usage.DataDirBytes = dirBytes(dataDir)
		var stat syscall.Statfs_t
		if err := syscall.Statfs(dataDir, &stat); err == nil {
			usage.DiskFreeBytes = int64(stat.Bavail) * int64(stat.Bsize)
			usage.DiskTotalBytes = int64(stat.Blocks) * int64(stat.Bsize)
		}
		mutex.Lock()
		cached, cachedAt = usage, time.Now()
		mutex.Unlock()
		writeStorageJSON(w, usage)
	})
}

func writeStorageJSON(w http.ResponseWriter, usage storageUsage) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(usage)
}

func fileBytes(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func dirBytes(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if info, infoErr := entry.Info(); infoErr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}
