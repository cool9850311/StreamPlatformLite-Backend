package util

import (
	"os"
	"path/filepath"
	"strings"
)

func GetGoServiceRootPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Find the index of "src"
	index := strings.LastIndex(dir, "src")
	if index == -1 {
		return "/app/Go-Service", nil
	}
	// Return the parent directory of the folder containing "src"
	return filepath.Dir(dir[:index]), nil
}

func GetProjectRootPath() (string, error) {
	goServiceRoot, err := GetGoServiceRootPath()
	if err != nil {
		return "/app", nil
	}
	// Return the parent directory of Go-Service/
	return filepath.Join(goServiceRoot, ".."), nil
}
