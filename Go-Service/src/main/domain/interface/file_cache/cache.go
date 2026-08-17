package file_cache

import "time"

type IFileCache interface {
	GetSingleFileName(filePath string) (string, error)
	ReadFile(filePath string) ([]byte, error)
	StoreCache(filePath string, data []byte)
	LoadCache(filePath string) ([]byte, bool)
	DeleteFile(filePath string)
	Range(f func(key, value interface{}) bool)
	Exists(path string) (bool, error)
	Stat(path string) (int64, time.Time, error)
}
