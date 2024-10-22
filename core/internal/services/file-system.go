package services

import (
	"node-herder/utils"
	"time"

	"github.com/patrickmn/go-cache"
)

type FileSystemService interface {
	ListFiles(path string) ([]string, error)
	ListFilesWithExtension(path string, extension string) ([]string, error)
	Load(file string) ([]byte, error)
}

type FileSystemServiceImpl struct {
	loader          utils.Loader
	walker          utils.Walker
	cache           *cache.Cache
	cacheExpiration time.Duration
	cleanupInterval time.Duration
}

func WithExpiration(cacheExpiration time.Duration) func(s *FileSystemServiceImpl) {
	return func(s *FileSystemServiceImpl) { s.cacheExpiration = cacheExpiration }
}

func WithCleanupInterval(cleanupInterval time.Duration) func(s *FileSystemServiceImpl) {
	return func(s *FileSystemServiceImpl) { s.cleanupInterval = cleanupInterval }
}

func NewFileSystemService(loader utils.Loader, walker utils.Walker, opts ...func(s *FileSystemServiceImpl)) *FileSystemServiceImpl {
	fs := &FileSystemServiceImpl{
		loader:          loader,
		walker:          walker,
		cacheExpiration: time.Minute,      // default cache expiration
		cleanupInterval: 10 * time.Minute, // default cleanup interval
	}

	for _, opt := range opts {
		opt(fs)
	}

	fs.cache = cache.New(fs.cacheExpiration, fs.cleanupInterval)
	return fs
}

func (s *FileSystemServiceImpl) ListFiles(path string) ([]string, error) {
	if files, ok := s.cache.Get(path); ok {
		return files.([]string), nil
	}

	files, err := utils.ListFiles(s.walker, path)
	if err != nil {
		return nil, err
	}
	s.cache.Add(path, files, cache.DefaultExpiration)
	return files, nil
}

func (s *FileSystemServiceImpl) ListFilesWithExtension(path string, extension string) ([]string, error) {
	if files, ok := s.cache.Get(path); ok {
		return files.([]string), nil
	}

	files, err := utils.ListFilesWithExtension(s.walker, path, extension)
	if err != nil {
		return nil, err
	}
	s.cache.Add(path, files, cache.DefaultExpiration)
	return files, nil
}

func (s *FileSystemServiceImpl) Load(file string) ([]byte, error) {
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
