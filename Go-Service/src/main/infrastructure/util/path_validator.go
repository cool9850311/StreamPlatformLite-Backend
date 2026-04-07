package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ValidateUUID validates UUID format using standard UUID parser
func ValidateUUID(uuidStr string) error {
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		return errors.New("invalid UUID format")
	}
	return nil
}

// ValidateHLSFilename validates filename safety with whitelist approach
func ValidateHLSFilename(filename string, allowedExts []string) error {
	// 1. Check for empty filename
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	// 2. Check for null bytes
	if strings.Contains(filename, "\x00") {
		return errors.New("invalid filename: null byte detected")
	}

	// 3. Check for path separators
	if strings.Contains(filename, "..") ||
		strings.Contains(filename, "/") ||
		strings.Contains(filename, "\\") {
		return errors.New("invalid filename: path traversal detected")
	}

	// 4. Validate extension whitelist
	ext := filepath.Ext(filename)
	validExt := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			validExt = true
			break
		}
	}
	if !validExt {
		return errors.New("invalid file extension")
	}

	// 5. Use filepath.IsLocal for additional validation
	if !filepath.IsLocal(filename) {
		return errors.New("filename is not local")
	}

	return nil
}

// SecureJoinPath safely joins paths using os.Root API (Go 1.24+)
// This automatically prevents path traversal and symlink escape attacks
func SecureJoinPath(rootPath, uuidStr, filename string) (string, error) {
	// Open hls directory as root
	hlsRootPath := filepath.Join(rootPath, "hls")
	hlsRoot, err := os.OpenRoot(hlsRootPath)
	if err != nil {
		return "", errors.New("failed to open HLS root directory")
	}
	defer hlsRoot.Close()

	// Build relative path: {uuid}/{filename}
	relativePath := filepath.Join(uuidStr, filename)

	// os.Root automatically prevents path traversal and symlink escape
	file, err := hlsRoot.Open(relativePath)
	if err != nil {
		return "", errors.New("file access denied or not found")
	}
	defer file.Close()

	// Return the safe path
	return filepath.Join(hlsRootPath, relativePath), nil
}

// ValidatePathBoundary validates that a path is within expected boundaries
// This is a fallback defense-in-depth measure
func ValidatePathBoundary(targetPath, expectedPrefix string) error {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return errors.New("failed to resolve absolute path")
	}

	absPrefix, err := filepath.Abs(expectedPrefix)
	if err != nil {
		return errors.New("failed to resolve prefix path")
	}

	// Ensure target path starts with the expected prefix
	if !strings.HasPrefix(absTarget, absPrefix+string(filepath.Separator)) &&
		absTarget != absPrefix {
		return errors.New("path outside allowed boundary")
	}

	return nil
}
