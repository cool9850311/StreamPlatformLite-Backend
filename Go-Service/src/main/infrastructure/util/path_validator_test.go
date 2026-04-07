package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		// Valid UUIDs
		{"Valid UUID v4", "83636040-7f54-49f2-ae40-9a1213614729", false},
		{"Valid UUID v1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", false},
		{"Valid UUID v4 uppercase", "83636040-7F54-49F2-AE40-9A1213614729", false},
		{"Valid UUID v4 lowercase", "a3bb189e-8bf9-3888-9912-ace4e6543002", false},
		{"Valid UUID hex format", "836360407f5449f2ae409a1213614729", false}, // UUID parser accepts this

		// Invalid UUIDs
		{"Path traversal", "../../etc/passwd", true},
		{"Invalid format", "abc-def-ghi", true},
		{"Empty string", "", true},
		{"Special chars", "<script>alert(1)</script>", true},
		{"Too short", "83636040-7f54", true},
		{"Too long", "83636040-7f54-49f2-ae40-9a1213614729-extra", true},
		{"Invalid characters", "83636040-7f54-49f2-ae40-9a121361472z", true},
		{"Null byte", "83636040-7f54-49f2-ae40-9a1213614729\x00", true},
		{"URL encoded", "%2e%2e%2f%2e%2e%2fetc%2fpasswd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHLSFilename(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		allowedExts []string
		wantErr     bool
	}{
		// Valid filenames
		{"Valid m3u8", "playlist.m3u8", []string{".m3u8", ".ts"}, false},
		{"Valid ts", "-1739708094904-1.ts", []string{".m3u8", ".ts"}, false},
		{"Valid mp4", "record.mp4", []string{".mp4"}, false},
		{"Valid ts with timestamp", "-1234567890-5.ts", []string{".ts"}, false},
		{"Valid long filename", "very-long-filename-with-many-characters-123456789.m3u8", []string{".m3u8"}, false},

		// Path traversal attempts
		{"Path traversal double dot", "../../etc/passwd", []string{".m3u8"}, true},
		{"Path traversal single dot", "../config.yaml", []string{".yaml"}, true},
		{"Absolute path Unix", "/etc/passwd", []string{".txt"}, true},
		{"Absolute path Windows", "C:\\Windows\\System32\\config.sys", []string{".sys"}, true},
		{"Contains slash", "subdir/file.m3u8", []string{".m3u8"}, true},
		{"Contains backslash", "subdir\\file.m3u8", []string{".m3u8"}, true},
		{"Hidden path traversal", "file..m3u8", []string{".m3u8"}, true},
		{"Double slash", "//etc/passwd", []string{".txt"}, true},

		// Invalid extensions
		{"Invalid extension php", "malicious.php", []string{".m3u8"}, true},
		{"Invalid extension sh", "exploit.sh", []string{".m3u8"}, true},
		{"Invalid extension exe", "virus.exe", []string{".m3u8"}, true},
		{"No extension", "noextension", []string{".m3u8"}, true},
		{"Wrong extension", "file.txt", []string{".m3u8"}, true},

		// Edge cases
		{"Empty filename", "", []string{".m3u8"}, true},
		{"Only extension", ".m3u8", []string{".m3u8"}, false}, // This is actually a valid filename in Unix
		{"Dot at start", ".hidden.m3u8", []string{".m3u8"}, false},
		{"Multiple dots", "file.backup.m3u8", []string{".m3u8"}, false},
		{"Null byte", "file\x00.m3u8", []string{".m3u8"}, true},

		// Special characters
		{"Space in filename", "file name.m3u8", []string{".m3u8"}, false},
		{"Dash in filename", "file-name.m3u8", []string{".m3u8"}, false},
		{"Underscore in filename", "file_name.m3u8", []string{".m3u8"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHLSFilename(tt.filename, tt.allowedExts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHLSFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureJoinPath(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()

	// Create hls directory structure
	hlsDir := filepath.Join(tmpDir, "hls")
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		t.Fatalf("Failed to create hls directory: %v", err)
	}

	// Create a test UUID directory
	testUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	uuidDir := filepath.Join(hlsDir, testUUID)
	if err := os.MkdirAll(uuidDir, 0755); err != nil {
		t.Fatalf("Failed to create UUID directory: %v", err)
	}

	// Create test files
	testM3u8 := filepath.Join(uuidDir, "playlist.m3u8")
	if err := os.WriteFile(testM3u8, []byte("#EXTM3U\n"), 0644); err != nil {
		t.Fatalf("Failed to create test m3u8 file: %v", err)
	}

	testTS := filepath.Join(uuidDir, "-1739708094904-1.ts")
	if err := os.WriteFile(testTS, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test ts file: %v", err)
	}

	tests := []struct {
		name     string
		rootPath string
		uuid     string
		filename string
		wantErr  bool
	}{
		// Valid paths
		{"Valid m3u8 file", tmpDir, testUUID, "playlist.m3u8", false},
		{"Valid ts file", tmpDir, testUUID, "-1739708094904-1.ts", false},

		// Invalid paths - path traversal
		{"Path traversal with double dot", tmpDir, testUUID, "../../../etc/passwd", true},
		{"Path traversal with parent", tmpDir, testUUID, "../config.yaml", true},
		{"Absolute path", tmpDir, testUUID, "/etc/passwd", true},

		// Invalid paths - non-existent files
		{"Non-existent file", tmpDir, testUUID, "nonexistent.m3u8", true},
		{"Non-existent UUID", tmpDir, "00000000-0000-0000-0000-000000000000", "playlist.m3u8", true},

		// Invalid root
		{"Invalid root path", "/nonexistent/path", testUUID, "playlist.m3u8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SecureJoinPath(tt.rootPath, tt.uuid, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureJoinPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error expected, verify the returned path
			if !tt.wantErr {
				expectedPath := filepath.Join(tt.rootPath, "hls", tt.uuid, tt.filename)
				if result != expectedPath {
					t.Errorf("SecureJoinPath() = %v, want %v", result, expectedPath)
				}

				// Verify the file actually exists
				if _, err := os.Stat(result); err != nil {
					t.Errorf("SecureJoinPath() returned path that doesn't exist: %v", result)
				}
			}
		})
	}
}

func TestSecureJoinPath_SymlinkProtection(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()

	// Create hls directory
	hlsDir := filepath.Join(tmpDir, "hls")
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		t.Fatalf("Failed to create hls directory: %v", err)
	}

	// Create a test UUID directory
	testUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	uuidDir := filepath.Join(hlsDir, testUUID)
	if err := os.MkdirAll(uuidDir, 0755); err != nil {
		t.Fatalf("Failed to create UUID directory: %v", err)
	}

	// Create a target directory outside hls
	outsideDir := filepath.Join(tmpDir, "outside")
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("Failed to create outside directory: %v", err)
	}

	// Create a file outside
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret data"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}

	// Create a symlink inside UUID directory pointing outside
	symlinkPath := filepath.Join(uuidDir, "symlink.m3u8")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	// Test accessing through symlink - should be blocked
	_, err := SecureJoinPath(tmpDir, testUUID, "symlink.m3u8")
	if err == nil {
		t.Error("SecureJoinPath() should block symlink escape, but didn't return error")
	}
}

func TestValidatePathBoundary(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create subdirectories
	allowedDir := filepath.Join(tmpDir, "allowed")
	if err := os.MkdirAll(allowedDir, 0755); err != nil {
		t.Fatalf("Failed to create allowed directory: %v", err)
	}

	forbiddenDir := filepath.Join(tmpDir, "forbidden")
	if err := os.MkdirAll(forbiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create forbidden directory: %v", err)
	}

	tests := []struct {
		name           string
		targetPath     string
		expectedPrefix string
		wantErr        bool
	}{
		// Valid paths
		{"Path within boundary", filepath.Join(allowedDir, "file.txt"), allowedDir, false},
		{"Path at boundary root", allowedDir, allowedDir, false},
		{"Path in subdirectory", filepath.Join(allowedDir, "subdir", "file.txt"), allowedDir, false},

		// Invalid paths
		{"Path outside boundary", filepath.Join(forbiddenDir, "file.txt"), allowedDir, true},
		{"Path in parent directory", tmpDir, allowedDir, true},
		{"Completely different path", "/etc/passwd", allowedDir, true},

		// Edge cases
		{"Empty target path", "", allowedDir, true},
		{"Empty prefix path", allowedDir, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathBoundary(tt.targetPath, tt.expectedPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathBoundary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathBoundary_RelativePaths(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test structure
	hlsDir := filepath.Join(tmpDir, "hls")
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		t.Fatalf("Failed to create hls directory: %v", err)
	}

	testUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	uuidDir := filepath.Join(hlsDir, testUUID)
	if err := os.MkdirAll(uuidDir, 0755); err != nil {
		t.Fatalf("Failed to create UUID directory: %v", err)
	}

	// Test with path containing .. (should be resolved to absolute and checked)
	relativePath := filepath.Join(uuidDir, "..", "..", "forbidden")
	err := ValidatePathBoundary(relativePath, hlsDir)
	if err == nil {
		t.Error("ValidatePathBoundary() should detect path escape through .., but didn't return error")
	}
}

func TestValidateHLSFilename_ExtensionEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		allowedExts []string
		wantErr     bool
	}{
		// Test with empty allowed extensions
		{"Empty allowed list", "file.m3u8", []string{}, true},
		{"Multiple allowed extensions", "file.ts", []string{".m3u8", ".ts", ".mp4"}, false},
		{"Case sensitive extension", "file.M3U8", []string{".m3u8"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHLSFilename(tt.filename, tt.allowedExts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHLSFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateUUID(b *testing.B) {
	validUUID := "83636040-7f54-49f2-ae40-9a1213614729"
	for i := 0; i < b.N; i++ {
		ValidateUUID(validUUID)
	}
}

func BenchmarkValidateHLSFilename(b *testing.B) {
	filename := "playlist.m3u8"
	allowedExts := []string{".m3u8", ".ts"}
	for i := 0; i < b.N; i++ {
		ValidateHLSFilename(filename, allowedExts)
	}
}

func BenchmarkValidatePathBoundary(b *testing.B) {
	tmpDir := b.TempDir()
	targetPath := filepath.Join(tmpDir, "file.txt")
	for i := 0; i < b.N; i++ {
		ValidatePathBoundary(targetPath, tmpDir)
	}
}
