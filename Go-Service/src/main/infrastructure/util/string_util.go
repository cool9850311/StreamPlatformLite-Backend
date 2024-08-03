package util

import (
	"strings"
)

func TrimPathToBase(path, base string) string {
	index := strings.Index(path, base)
	if index == -1 {
		return ""
	}

	trimmedPath := path[:index+len(base)]
	return trimmedPath
}
