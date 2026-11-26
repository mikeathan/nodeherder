package fs

import (
	"os"
	"path/filepath"
	"time"

	"github.com/patrickmn/go-cache"
)

type Loader interface {
	Load(file string) ([]byte, error)
}

type FileSystemLoader struct{}

func NewFileSystemLoader() *FileSystemLoader {
	return &FileSystemLoader{}
}

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

func NewFileSystemWalker() *FileSystemWalker {
	return &FileSystemWalker{}
}

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

type FileSystem interface {
	ListFiles(path string) ([]string, error)
	ListFilesWithExtension(path string, extension string) ([]string, error)
	Load(file string) ([]byte, error)
	// Stat reports file metadata for a listed path. It does not follow a different data
	// source than Load; both operate on the log directory contract owned by the caller.
	Stat(file string) (os.FileInfo, error)
	// Delete removes exactly the named file. Callers own path validation and confirmation.
	Delete(file string) error
}

type FileSystemImpl struct {
	loader          Loader
	walker          Walker
	cache           *cache.Cache
	cacheExpiration time.Duration
	cleanupInterval time.Duration
}

func WithExpiration(cacheExpiration time.Duration) func(s *FileSystemImpl) {
	return func(s *FileSystemImpl) { s.cacheExpiration = cacheExpiration }
}

func WithCleanupInterval(cleanupInterval time.Duration) func(s *FileSystemImpl) {
	return func(s *FileSystemImpl) { s.cleanupInterval = cleanupInterval }
}

func WithFileLoader(loader Loader) func(s *FileSystemImpl) {
	return func(s *FileSystemImpl) { s.loader = loader }
}

func WithFileWalker(walker Walker) func(s *FileSystemImpl) {
	return func(s *FileSystemImpl) { s.walker = walker }
}

func NewFileSystem(opts ...func(s *FileSystemImpl)) *FileSystemImpl {
	fs := &FileSystemImpl{
		loader:          NewFileSystemLoader(),
		walker:          NewFileSystemWalker(),
		cacheExpiration: time.Minute,      // default cache expiration
		cleanupInterval: 10 * time.Minute, // default cleanup interval
	}

	for _, opt := range opts {
		opt(fs)
	}

	fs.cache = cache.New(fs.cacheExpiration, fs.cleanupInterval)
	return fs
}

func (s *FileSystemImpl) ListFiles(path string) ([]string, error) {
	if files, ok := s.cache.Get(path); ok {
		return files.([]string), nil
	}

	files, err := ListFiles(s.walker, path)
	if err != nil {
		return nil, err
	}
	s.cache.Add(path, files, cache.DefaultExpiration)
	return files, nil
}

func (s *FileSystemImpl) ListFilesWithExtension(path string, extension string) ([]string, error) {
	if files, ok := s.cache.Get(path); ok {
		return files.([]string), nil
	}

	files, err := ListFilesWithExtension(s.walker, path, extension)
	if err != nil {
		return nil, err
	}
	s.cache.Add(path, files, cache.DefaultExpiration)
	return files, nil
}

// Delete removes one file; no directory or wildcard operation is exposed here.
func (s *FileSystemImpl) Delete(file string) error {
	return os.Remove(file)
}

// Stat returns metadata for one file without caching it; callers only need this on an
// explicit log-list request.
func (s *FileSystemImpl) Stat(file string) (os.FileInfo, error) {
	return os.Stat(file)
}

func (s *FileSystemImpl) Load(file string) ([]byte, error) {
	if buf, ok := s.cache.Get(file); ok {
		return buf.([]byte), nil
	}

	buf, err := s.loader.Load(file)
	if err != nil {
		return nil, err
	}

	s.cache.Add(file, buf, cache.DefaultExpiration)
	return buf, nil

}
