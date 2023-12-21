package api

import (
	"net/http"
	"os"
	"path"
	"strings"
)

type FileServer struct {
	Handle  http.Handler
	rootDir string
}

func NewFileServer(rootDir string) *FileServer {

	fs := new(FileServer)
	fs.rootDir = rootDir
	fs.Handle = http.FileServer(http.Dir(rootDir))

	return fs
}

func (fs *FileServer) ResolveRoot() http.Handler {
	return fs.Handle
}

func (fs *FileServer) Resolve(root bool) http.Handler {
	if root {
		return fs.Handle
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If the requested file exists then return if; otherwise return index.html (fileserver default page)
		if r.URL.Path != "/" {
			fullPath := fs.rootDir + strings.TrimPrefix(path.Clean(r.URL.Path), "/")
			_, err := os.Stat(fullPath)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				// Requested file does not exist so we return the default (resolves to index.html)
				r.URL.Path = "/"
			}
		}
		fs.Handle.ServeHTTP(w, r)
	})
}
