package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
)

var integrationFileID = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var ErrDefinitionConflict = errors.New("definition changed on disk")

// IntegrationDefinitions owns only shareable host YAML. Runtime activation/journaling belongs to
// the service transaction. Concurrent external edits and UI saves require operational coordination.
type IntegrationDefinitions struct {
	root string
	mu   sync.Mutex
}

func OpenIntegrationDefinitions(root string) (*IntegrationDefinitions, error) {
	if root == "" {
		return nil, errors.New("integration configuration directory is required")
	}
	if e := os.MkdirAll(root, 0755); e != nil {
		return nil, e
	}
	st, e := os.Lstat(root)
	if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("integration configuration root must be a directory, not a symlink")
	}
	return &IntegrationDefinitions{root: root}, nil
}
func definitionDigest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (r *IntegrationDefinitions) path(key string, create bool) (string, error) {
	parts := strings.Split(key, "/")
	if len(parts) != 2 || !integrationFileID.MatchString(parts[0]) || !integrationFileID.MatchString(parts[1]) {
		return "", errors.New("invalid definition identifier")
	}
	directory := filepath.Join(r.root, parts[0])
	if create {
		if e := os.Mkdir(directory, 0755); e != nil && !os.IsExist(e) {
			return "", e
		}
	}
	st, e := os.Lstat(directory)
	if e != nil {
		return "", e
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("definition directory cannot be a symlink")
	}
	p := filepath.Join(directory, parts[1]+".yaml")
	if st, e = os.Lstat(p); e == nil {
		if !st.Mode().IsRegular() || st.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("definition must be a regular file")
		}
		if st.Size() > 64<<10 {
			return "", errors.New("definition exceeds 64 KiB")
		}
	} else if !os.IsNotExist(e) {
		return "", e
	}
	return p, nil
}
func (r *IntegrationDefinitions) Read(key string) (string, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, e := r.path(key, false)
	if e != nil {
		return "", "", e
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return "", "", e
	}
	if len(b) > 64<<10 {
		return "", "", errors.New("definition exceeds 64 KiB")
	}
	return string(b), definitionDigest(b), nil
}
func (r *IntegrationDefinitions) List() ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, e := os.ReadDir(r.root)
	if e != nil {
		return nil, e
	}
	keys := []string{}
	for _, entry := range entries {
		if !entry.IsDir() || !integrationFileID.MatchString(entry.Name()) {
			continue
		}
		files, e := os.ReadDir(filepath.Join(r.root, entry.Name()))
		if e != nil {
			return nil, e
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".yaml") {
				key := entry.Name() + "/" + strings.TrimSuffix(file.Name(), ".yaml")
				if _, e = r.path(key, false); e != nil {
					return nil, e
				}
				keys = append(keys, key)
				if len(keys) > 128 {
					return nil, errors.New("at most 128 provider definitions are supported")
				}
			}
		}
	}
	sort.Strings(keys)
	return keys, nil
}

// Write atomically replaces an existing hash, or exclusively creates a missing seed/new definition
// when expected is empty. The caller must have durably journaled pending activation first.
func (r *IntegrationDefinitions) Write(key, source, expected string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(source) > 64<<10 {
		return errors.New("definition exceeds 64 KiB")
	}
	p, e := r.path(key, true)
	if e != nil {
		return e
	}
	current, e := os.ReadFile(p)
	exists := e == nil
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if exists && (expected == "" || definitionDigest(current) != expected) || !exists && expected != "" {
		return ErrDefinitionConflict
	}
	mode := fs.FileMode(0644)
	uid, gid := -1, -1
	if exists {
		st, e := os.Stat(p)
		if e != nil {
			return e
		}
		mode = st.Mode().Perm()
		if owner, ok := st.Sys().(*syscall.Stat_t); ok {
			uid = int(owner.Uid)
			gid = int(owner.Gid)
		}
	}
	tmp, e := os.CreateTemp(filepath.Dir(p), ".definition-*")
	if e != nil {
		return e
	}
	defer os.Remove(tmp.Name())
	closed := false
	defer func() {
		if !closed {
			_ = tmp.Close()
		}
	}()
	if uid >= 0 && os.Geteuid() == 0 {
		if e = tmp.Chown(uid, gid); e != nil {
			return e
		}
	}
	if e = tmp.Chmod(mode); e != nil {
		return e
	}
	if _, e = tmp.WriteString(source); e != nil {
		return e
	}
	if e = tmp.Sync(); e != nil {
		return e
	}
	if e = tmp.Close(); e != nil {
		return e
	}
	closed = true
	if !exists {
		e = os.Link(tmp.Name(), p)
		if os.IsExist(e) {
			return ErrDefinitionConflict
		}
	} else {
		check, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		if definitionDigest(check) != expected {
			return ErrDefinitionConflict
		}
		e = os.Rename(tmp.Name(), p)
	}
	if e != nil {
		return e
	}
	directory, e := os.Open(filepath.Dir(p))
	if e != nil {
		return e
	}
	defer directory.Close()
	if e = directory.Sync(); e != nil {
		return fmt.Errorf("flush definition directory: %w", e)
	}
	return nil
}
func (r *IntegrationDefinitions) Delete(key, expected string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, e := r.path(key, false)
	if e != nil {
		return e
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	if definitionDigest(b) != expected {
		return ErrDefinitionConflict
	}
	if e = os.Remove(p); e != nil {
		return e
	}
	dir, e := os.Open(filepath.Dir(p))
	if e != nil {
		return e
	}
	defer dir.Close()
	return dir.Sync()
}
