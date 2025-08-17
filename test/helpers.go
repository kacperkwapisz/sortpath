package test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for testing and returns cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	
	dir, err := os.MkdirTemp("", "sortpath-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	
	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("failed to cleanup temp dir %s: %v", dir, err)
		}
	}
	
	return dir, cleanup
}

// WriteFile writes content to a file in the given directory
func WriteFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directories for %s: %v", path, err)
	}
	
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// CreateTestFS creates a test filesystem structure
func CreateTestFS(t *testing.T, dir string, structure map[string]string) {
	t.Helper()
	
	for path, content := range structure {
		WriteFile(t, dir, path, content)
	}
}

// LoadTestData loads test data from testdata directory
func LoadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", filename, err)
	}
	
	return data
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", path)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file %s to not exist", path)
	}
}

// AssertFileMode checks file permissions
func AssertFileMode(t *testing.T, path string, expectedMode fs.FileMode) {
	t.Helper()
	
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", path, err)
	}
	
	if info.Mode() != expectedMode {
		t.Errorf("expected file %s to have mode %v, got %v", path, expectedMode, info.Mode())
	}
}