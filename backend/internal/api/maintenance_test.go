package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/api"
	"node-herder/internal/fs"
	"node-herder/internal/ratelimiter"
	"node-herder/mocks"
	"node-herder/models/logging"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func purgeRequest(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/history/purge", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.NewPurgeMetricsHandler(ratelimiter.NewWindowRateLimiter(10, time.Minute), mocks.NewMockAppStore()).ServeHTTP(w, req)
	return w
}

func TestPurgeMetricsValidatesCutoff(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"missing", `{}`, http.StatusBadRequest},
		{"malformed", `{"before":"yesterday"}`, http.StatusBadRequest},
		{"future", `{"before":"2999-01-01T00:00:00Z"}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := purgeRequest(t, tc.body).Code; got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPurgeMetricsAcceptsExplicitPastCutoff(t *testing.T) {
	before := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	w := purgeRequest(t, `{"before":"`+before+`"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	var response api.PurgeMetricsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if response.PurgedBefore != before {
		t.Fatalf("purgedBefore = %q, want %q", response.PurgedBefore, before)
	}
}

func TestLogFileDeleteRemovesNamedArchive(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "nodeherder.log"), []byte("log"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	t.Chdir(dir)
	file := "nodeherder.log"
	fsys := fs.NewFileSystem(fs.WithFileWalker(mocks.NewMockWalker([]string{file})))
	body, err := json.Marshal(logging.NewFileLogRequest(file, logging.DeleteAction))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/logfile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.NewLogFileHandler(fsys).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("file still exists or stat failed: %v", err)
	}
}

func TestLogFileDeleteRejectsTraversal(t *testing.T) {
	fsys := fs.NewFileSystem()
	body, _ := json.Marshal(logging.NewFileLogRequest("../secret.log", logging.DeleteAction))
	req := httptest.NewRequest(http.MethodPost, "/api/logfile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.NewLogFileHandler(fsys).ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
