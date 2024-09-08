package util
import (
	"os"
	"path/filepath"
	"strings"
	"errors"
)

func GetGoServiceRootPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Find the index of "Go-Service/src"
	index := strings.Index(dir, "Go-Service/src")
	if index == -1 {
		return "", errors.New("Go-Service/src not found in the current directory path")
	}
	// Return the path up to "Go-Service/"
	return dir[:index+len("Go-Service")], nil
}

func GetProjectRootPath() (string, error) {
	goServiceRoot, err := GetGoServiceRootPath()
	if err != nil {
		return "", err
	}
	// Return the parent directory of Go-Service/
	return filepath.Join(goServiceRoot, ".."), nil
}