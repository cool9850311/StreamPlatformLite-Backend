package file_cache

type IFileCache interface {
	GetSingleFileName(filePath string) (string, error)
	ReadFile(filePath string) ([]byte, error)
	StoreCache(filePath string, data []byte)
	LoadCache(filePath string) ([]byte, bool)
	DeleteFile(filePath string)
	Range(f func(key, value interface{}) bool)
}
