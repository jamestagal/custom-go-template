package builder

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateBuildFingerprint_Consistency tests that the same input produces the same hash.
func TestGenerateBuildFingerprint_Consistency(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"index.html":      "<html><body>Hello</body></html>",
		"style.css":       "body { color: red; }",
		"script.js":       "console.log('hello');",
		"nested/page.html": "<html><body>Nested</body></html>",
	}

	for name, content := range testFiles {
		path := filepath.Join(tempDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Generate fingerprint multiple times
	fingerprint1 := GenerateBuildFingerprint(tempDir)
	fingerprint2 := GenerateBuildFingerprint(tempDir)
	fingerprint3 := GenerateBuildFingerprint(tempDir)

	// All should be identical
	if fingerprint1 != fingerprint2 || fingerprint2 != fingerprint3 {
		t.Errorf("Fingerprints not consistent: %q, %q, %q", fingerprint1, fingerprint2, fingerprint3)
	}

	// Should be exactly FingerprintLength characters
	if len(fingerprint1) != FingerprintLength {
		t.Errorf("Fingerprint length = %d, want %d", len(fingerprint1), FingerprintLength)
	}

	// Should be hex string
	if !isHexString(fingerprint1) {
		t.Errorf("Fingerprint %q is not a hex string", fingerprint1)
	}
}

// TestGenerateBuildFingerprint_ChangesWithContent tests that hash changes when files change.
func TestGenerateBuildFingerprint_ChangesWithContent(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial file
	testFile := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(testFile, []byte("<html>Version 1</html>"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Generate first fingerprint
	fingerprint1 := GenerateBuildFingerprint(tempDir)

	// Modify file content
	if err := os.WriteFile(testFile, []byte("<html>Version 2</html>"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Generate second fingerprint
	fingerprint2 := GenerateBuildFingerprint(tempDir)

	// Should be different
	if fingerprint1 == fingerprint2 {
		t.Errorf("Fingerprints should differ after content change: both are %q", fingerprint1)
	}
}

// TestGenerateBuildFingerprint_EmptyDirectory tests handling of empty directory.
func TestGenerateBuildFingerprint_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	fingerprint := GenerateBuildFingerprint(tempDir)

	// Should return a valid fingerprint (hash of empty input)
	if len(fingerprint) != FingerprintLength {
		t.Errorf("Fingerprint length = %d, want %d", len(fingerprint), FingerprintLength)
	}
}

// TestGenerateBuildFingerprint_MissingDirectory tests handling of non-existent directory.
func TestGenerateBuildFingerprint_MissingDirectory(t *testing.T) {
	fingerprint := GenerateBuildFingerprint("/non/existent/path")

	// Should return empty string (graceful degradation)
	if fingerprint != "" {
		t.Errorf("Fingerprint for missing directory = %q, want empty string", fingerprint)
	}
}

// TestGenerateBuildFingerprint_IgnoresOtherFiles tests that only .html, .js, .css are included.
func TestGenerateBuildFingerprint_IgnoresOtherFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create only valid files first
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte("<html>Test</html>"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	fingerprint1 := GenerateBuildFingerprint(tempDir)

	// Add files that should be ignored
	ignoredFiles := []string{
		"image.png",
		"data.json",
		"readme.md",
		"config.yaml",
	}
	for _, name := range ignoredFiles {
		if err := os.WriteFile(filepath.Join(tempDir, name), []byte("ignored content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	fingerprint2 := GenerateBuildFingerprint(tempDir)

	// Fingerprints should be identical (ignored files don't affect hash)
	if fingerprint1 != fingerprint2 {
		t.Errorf("Adding ignored files changed fingerprint: %q → %q", fingerprint1, fingerprint2)
	}
}

// TestCreateFingerprintedOutput_DirectoryStructure tests directory creation.
func TestCreateFingerprintedOutput_DirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	fingerprint := "abc123def0"

	output, err := CreateFingerprintedOutput(tempDir, fingerprint)
	if err != nil {
		t.Fatalf("CreateFingerprintedOutput failed: %v", err)
	}

	// Verify output struct
	if output.Fingerprint != fingerprint {
		t.Errorf("Fingerprint = %q, want %q", output.Fingerprint, fingerprint)
	}
	if output.CoreDir != filepath.Join(fingerprint, "core") {
		t.Errorf("CoreDir = %q, want %q", output.CoreDir, filepath.Join(fingerprint, "core"))
	}

	// Verify all directories exist
	expectedDirs := []string{
		output.CoreDir,
		output.GeneratedDir,
		output.ComponentsDir,
		output.ContentDir,
		output.GlobalDir,
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(tempDir, dir)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("Directory %q not created: %v", fullPath, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", fullPath)
		}
	}
}

// TestCreateFingerprintedOutput_EmptyFingerprint tests error handling for empty fingerprint.
func TestCreateFingerprintedOutput_EmptyFingerprint(t *testing.T) {
	tempDir := t.TempDir()

	_, err := CreateFingerprintedOutput(tempDir, "")
	if err == nil {
		t.Error("Expected error for empty fingerprint, got nil")
	}
}

// TestGetFingerprintedPath tests path generation.
func TestGetFingerprintedPath(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint string
		assetPath   string
		want        string
	}{
		{
			name:        "core main.js",
			fingerprint: "abc123def0",
			assetPath:   "core/main.js",
			want:        "abc123def0/core/main.js",
		},
		{
			name:        "generated layouts.js",
			fingerprint: "xyz789",
			assetPath:   "generated/layouts.js",
			want:        "xyz789/generated/layouts.js",
		},
		{
			name:        "empty fingerprint",
			fingerprint: "",
			assetPath:   "core/main.js",
			want:        "core/main.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFingerprintedPath(tt.fingerprint, tt.assetPath)
			if got != tt.want {
				t.Errorf("GetFingerprintedPath(%q, %q) = %q, want %q",
					tt.fingerprint, tt.assetPath, got, tt.want)
			}
		})
	}
}

// TestCleanOldFingerprints tests cleanup of old fingerprint directories.
func TestCleanOldFingerprints(t *testing.T) {
	tempDir := t.TempDir()

	// Create some fingerprint directories
	fingerprints := []string{"abc123def0", "xyz789abcd", "111222333a"}
	currentFingerprint := "xyz789abcd"

	for _, fp := range fingerprints {
		if err := os.MkdirAll(filepath.Join(tempDir, fp, "core"), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Create some non-fingerprint directories that should NOT be removed
	nonFingerprints := []string{"static", "images", "css"}
	for _, dir := range nonFingerprints {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Clean old fingerprints
	err := CleanOldFingerprints(tempDir, currentFingerprint)
	if err != nil {
		t.Fatalf("CleanOldFingerprints failed: %v", err)
	}

	// Current fingerprint should still exist
	if _, err := os.Stat(filepath.Join(tempDir, currentFingerprint)); os.IsNotExist(err) {
		t.Error("Current fingerprint was incorrectly removed")
	}

	// Old fingerprints should be removed
	for _, fp := range fingerprints {
		if fp == currentFingerprint {
			continue
		}
		if _, err := os.Stat(filepath.Join(tempDir, fp)); !os.IsNotExist(err) {
			t.Errorf("Old fingerprint %q was not removed", fp)
		}
	}

	// Non-fingerprint directories should still exist
	for _, dir := range nonFingerprints {
		if _, err := os.Stat(filepath.Join(tempDir, dir)); os.IsNotExist(err) {
			t.Errorf("Non-fingerprint directory %q was incorrectly removed", dir)
		}
	}
}

// TestIsHexString tests hex string validation.
func TestIsHexString(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"abc123", true},
		{"ABC123", true},
		{"abcdef0123456789", true},
		{"ABCDEF0123456789", true},
		{"abc123xyz", false}, // contains non-hex chars
		{"abc-123", false},   // contains dash
		{"", true},           // empty string is technically valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isHexString(tt.input)
			if got != tt.want {
				t.Errorf("isHexString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
