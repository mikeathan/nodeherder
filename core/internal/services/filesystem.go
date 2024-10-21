package services

import (
	"node-herder/utils"
	"time"

	"github.com/patrickmn/go-cache"
)

type FileSystemServiceImpl struct {
	loader utils.Loader
	walker utils.Walker
	cache  *cache.Cache
}

func NewFileSystemService(loader utils.Loader, walker utils.Walker) *FileSystemServiceImpl {
	return &FileSystemServiceImpl{
		loader: loader,
		walker: walker,
		cache:  cache.New(5*time.Minute, 10*time.Minute),
	}
}
