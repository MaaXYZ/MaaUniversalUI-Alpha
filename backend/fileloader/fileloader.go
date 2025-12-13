package fileloader

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type FileLoader struct {
	basePath string
}

func New(basePath string) *FileLoader {
	return &FileLoader{basePath: basePath}
}

func (fl *FileLoader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("FileLoader ServeHTTP", r.URL.Path)

	relativePath := r.URL.Path

	safePath := filepath.Clean(relativePath)

	fullPath := filepath.Join(fl.basePath, safePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, fullPath)
}
