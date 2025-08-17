package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid relative path",
			path:     "folder/subfolder",
			expected: "folder/subfolder",
			wantErr:  false,
		},
		{
			name:     "valid absolute path",
			path:     "/usr/local/bin",
			expected: "/usr/local/bin",
			wantErr:  false,
		},
		{
			name:    "directory traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
			errMsg:  "directory traversal",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "path cannot be empty",
		},
		{
			name:     "path with dots cleaned",
			path:     "./folder/../subfolder",
			expected: "subfolder",
			wantErr:  false,
		},
		{
			name:     "current directory",
			path:     ".",
			expected: ".",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizePath(tt.path)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizePath() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("SanitizePath() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizePath() unexpected error = %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("SanitizePath() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestValidateConfigKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid api-key",
			key:     "api-key",
			wantErr: false,
		},
		{
			name:    "valid api-base",
			key:     "api-base",
			wantErr: false,
		},
		{
			name:    "valid model",
			key:     "model",
			wantErr: false,
		},
		{
			name:    "valid tree-path",
			key:     "tree-path",
			wantErr: false,
		},
		{
			name:    "valid log-level",
			key:     "log-level",
			wantErr: false,
		},
		{
			name:    "invalid key",
			key:     "invalid-key",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigKey(tt.key)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateConfigKey() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateConfigKey() unexpected error = %v", err)
			}
		})
	}
}

func TestSanitizeConfigValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid api-key",
			key:      "api-key",
			value:    "sk-1234567890abcdef",
			expected: "sk-1234567890abcdef",
			wantErr:  false,
		},
		{
			name:    "api-key with newline",
			key:     "api-key",
			value:   "sk-123\n456",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:     "valid model name",
			key:      "model",
			value:    "gpt-3.5-turbo",
			expected: "gpt-3.5-turbo",
			wantErr:  false,
		},
		{
			name:    "invalid model name",
			key:     "model",
			value:   "model@name",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:     "valid tree-path",
			key:      "tree-path",
			value:    "/home/user/documents",
			expected: "/home/user/documents",
			wantErr:  false,
		},
		{
			name:    "tree-path with traversal",
			key:     "tree-path",
			value:   "../../../etc",
			wantErr: true,
			errMsg:  "directory traversal",
		},
		{
			name:     "log-level normalization",
			key:      "log-level",
			value:    "DEBUG",
			expected: "debug",
			wantErr:  false,
		},
		{
			name:    "unknown key",
			key:     "unknown",
			value:   "value",
			wantErr: true,
			errMsg:  "unknown config key",
		},
		{
			name:     "whitespace trimming",
			key:      "model",
			value:    "  gpt-4  ",
			expected: "gpt-4",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeConfigValue(tt.key, tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizeConfigValue() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("SanitizeConfigValue() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizeConfigValue() unexpected error = %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("SanitizeConfigValue() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestRedactSensitiveValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "redact long api-key",
			key:      "api-key",
			value:    "sk-1234567890abcdef",
			expected: "sk-1...cdef",
		},
		{
			name:     "redact short api-key",
			key:      "api-key",
			value:    "short",
			expected: "***",
		},
		{
			name:     "non-sensitive value",
			key:      "model",
			value:    "gpt-3.5-turbo",
			expected: "gpt-3.5-turbo",
		},
		{
			name:     "empty api-key",
			key:      "api-key",
			value:    "",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSensitiveValue(tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("RedactSensitiveValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSecureFileOperations_CreateSecureFile(t *testing.T) {
	tmpDir := t.TempDir()
	ops := &SecureFileOperations{}
	
	testPath := filepath.Join(tmpDir, "subdir", "test.txt")
	
	file, err := ops.CreateSecureFile(testPath)
	if err != nil {
		t.Fatalf("CreateSecureFile() error = %v", err)
	}
	defer file.Close()
	
	// Check that file exists
	if _, err := os.Stat(testPath); err != nil {
		t.Errorf("File was not created: %v", err)
	}
	
	// Check permissions
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestSecureFileOperations_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	ops := &SecureFileOperations{}
	
	testPath := filepath.Join(tmpDir, "atomic_test.txt")
	testData := []byte("test data for atomic write")
	
	err := ops.AtomicWrite(testPath, testData)
	if err != nil {
		t.Fatalf("AtomicWrite() error = %v", err)
	}
	
	// Check that file exists and has correct content
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	
	if string(content) != string(testData) {
		t.Errorf("Expected content %q, got %q", testData, content)
	}
	
	// Check permissions
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestSecureFileOperations_ValidateFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	ops := &SecureFileOperations{}
	
	tests := []struct {
		name        string
		permissions os.FileMode
		wantErr     bool
	}{
		{
			name:        "secure permissions",
			permissions: 0600,
			wantErr:     false,
		},
		{
			name:        "insecure permissions",
			permissions: 0644,
			wantErr:     true,
		},
		{
			name:        "world readable",
			permissions: 0604,
			wantErr:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := filepath.Join(tmpDir, tt.name+".txt")
			
			// Create file with specific permissions
			file, err := os.Create(testPath)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			file.Close()
			
			if err := os.Chmod(testPath, tt.permissions); err != nil {
				t.Fatalf("Failed to set permissions: %v", err)
			}
			
			err = ops.ValidateFilePermissions(testPath)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateFilePermissions() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateFilePermissions() unexpected error = %v", err)
			}
		})
	}
}

func TestSecureFileOperations_EnsureSecurePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	ops := &SecureFileOperations{}
	
	testPath := filepath.Join(tmpDir, "permissions_test.txt")
	
	// Create file with insecure permissions
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()
	
	if err := os.Chmod(testPath, 0644); err != nil {
		t.Fatalf("Failed to set initial permissions: %v", err)
	}
	
	// Fix permissions
	err = ops.EnsureSecurePermissions(testPath)
	if err != nil {
		t.Fatalf("EnsureSecurePermissions() error = %v", err)
	}
	
	// Verify permissions were fixed
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}
}