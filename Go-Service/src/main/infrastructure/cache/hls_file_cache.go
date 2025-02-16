package cache

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type FileCache struct {
	cache sync.Map
}

func NewFileCache() *FileCache {
	return &FileCache{
		cache: sync.Map{},
	}
}

func (fc *FileCache) GetSingleFileName(filePath string) (string, error) {

	// Find all matching files
	matches, err := filepath.Glob(filePath)
	if err != nil {
		return "", err
	}

	// Check if we found any matches
	if len(matches) == 0 {
		return "", errors.New("no matching files found")
	}

	// Return the first match
	return matches[0], nil
}

func (fc *FileCache) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (fc *FileCache) StoreCache(filePath string, data []byte) {
	fc.cache.Store(filePath, data)
}

func (fc *FileCache) LoadCache(filePath string) ([]byte, bool) {
	data, ok := fc.cache.Load(filePath)
	if !ok {
		return nil, false
	}
	return data.([]byte), true
}

func (fc *FileCache) DeleteFile(filePath string) {
	fc.cache.Delete(filePath)
}

func (fc *FileCache) Range(f func(key, value interface{}) bool) {
	fc.cache.Range(f)
}
