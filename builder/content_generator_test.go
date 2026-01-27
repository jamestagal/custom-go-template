package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateContentJS_BasicStructure tests the basic output format.
func TestGenerateContentJS_BasicStructure(t *testing.T) {
	tempDir := t.TempDir()

	// Create content directory structure
	pagesDir := filepath.Join(tempDir, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		t.Fatalf("Failed to create pages dir: %v", err)
	}

	// Create a test JSON file
	aboutJSON := `{"title": "About Us", "description": "Test page"}`
	if err := os.WriteFile(filepath.Join(pagesDir, "about.json"), []byte(aboutJSON), 0644); err != nil {
		t.Fatalf("Failed to create about.json: %v", err)
	}

	// Generate content.js
	js, err := GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify structure
	if !strings.Contains(js, "const allContent = [") {
		t.Error("Missing 'const allContent = [' declaration")
	}
	if !strings.Contains(js, "export default allContent;") {
		t.Error("Missing 'export default allContent;'")
	}
	if !strings.Contains(js, `type: "pages"`) {
		t.Error("Missing type field")
	}
	if !strings.Contains(js, `path: "about"`) {
		t.Error("Missing path field")
	}
	if !strings.Contains(js, `filename: "about.json"`) {
		t.Error("Missing filename field")
	}
}

// TestGenerateContentJS_MultipleTypes tests content with multiple types.
func TestGenerateContentJS_MultipleTypes(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple content types
	types := []string{"pages", "news", "blog"}
	for _, typ := range types {
		dir := filepath.Join(tempDir, typ)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create %s dir: %v", typ, err)
		}

		content := map[string]interface{}{"title": typ + " content"}
		data, _ := json.Marshal(content)
		if err := os.WriteFile(filepath.Join(dir, "test.json"), data, 0644); err != nil {
			t.Fatalf("Failed to create test.json in %s: %v", typ, err)
		}
	}

	// Generate content.js
	js, err := GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify all types present
	for _, typ := range types {
		if !strings.Contains(js, `type: "`+typ+`"`) {
			t.Errorf("Missing type: %s", typ)
		}
	}
}

// TestGenerateContentJS_IndexFile tests special _index.json handling.
func TestGenerateContentJS_IndexFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create pages directory with _index.json
	pagesDir := filepath.Join(tempDir, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		t.Fatalf("Failed to create pages dir: %v", err)
	}

	indexJSON := `{"components": [{"name": "hero", "fields": {}}]}`
	if err := os.WriteFile(filepath.Join(pagesDir, "_index.json"), []byte(indexJSON), 0644); err != nil {
		t.Fatalf("Failed to create _index.json: %v", err)
	}

	// Generate content.js
	js, err := GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify _index.json has empty path (root)
	if !strings.Contains(js, `path: ""`) {
		t.Error("_index.json should have empty path")
	}
	if !strings.Contains(js, `filename: "_index.json"`) {
		t.Error("Missing filename for _index.json")
	}
}

// TestGenerateContentJS_NestedDirectories tests nested content directories.
func TestGenerateContentJS_NestedDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested structure: blog/2024/article.json
	nestedDir := filepath.Join(tempDir, "blog", "2024")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	articleJSON := `{"title": "Nested Article"}`
	if err := os.WriteFile(filepath.Join(nestedDir, "article.json"), []byte(articleJSON), 0644); err != nil {
		t.Fatalf("Failed to create article.json: %v", err)
	}

	// Generate content.js
	js, err := GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify type is still "blog" (first level)
	if !strings.Contains(js, `type: "blog"`) {
		t.Error("Type should be 'blog' from first-level directory")
	}
	// Verify path includes nested path
	if !strings.Contains(js, `path: "2024/article"`) {
		t.Error("Path should include nested directories")
	}
}

// TestGenerateContentJS_EmptyDirectory tests empty content directory.
func TestGenerateContentJS_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Generate content.js from empty directory
	js, err := GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Should still produce valid JS
	if !strings.Contains(js, "const allContent = [") {
		t.Error("Missing 'const allContent = ['")
	}
	if !strings.Contains(js, "];") {
		t.Error("Missing closing bracket")
	}
}

// TestGenerateContentJS_MissingDirectory tests non-existent directory handling.
func TestGenerateContentJS_MissingDirectory(t *testing.T) {
	// Generate content.js from non-existent directory
	js, err := GenerateContentJS("/non/existent/path")
	if err != nil {
		t.Fatalf("GenerateContentJS should handle missing directory: %v", err)
	}

	// Should produce valid empty JS
	if !strings.Contains(js, "const allContent = [") {
		t.Error("Should produce valid JS even for missing directory")
	}
}

// TestGenerateContentJS_FieldsPreservation tests that all JSON fields are preserved.
func TestGenerateContentJS_FieldsPreservation(t *testing.T) {
	tempDir := t.TempDir()

	// Create content with complex fields
	pagesDir := filepath.Join(tempDir, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		t.Fatalf("Failed to create pages dir: %v", err)
	}

	complexJSON := `{
		"title": "Complex Page",
		"published": true,
		"count": 42,
		"tags": ["go", "alpine"],
		"author": {
			"name": "John",
			"email": "john@example.com"
		},
		"components": [
			{"name": "hero", "fields": {"title": "Hello"}}
		]
	}`
	if err := os.WriteFile(filepath.Join(pagesDir, "complex.json"), []byte(complexJSON), 0644); err != nil {
		t.Fatalf("Failed to create complex.json: %v", err)
	}

	// Generate content.js
	js, err := GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify complex fields are present
	checks := []string{
		`"title": "Complex Page"`,
		`"published": true`,
		`"count": 42`,
		`"tags"`,
		`"author"`,
		`"components"`,
	}
	for _, check := range checks {
		if !strings.Contains(js, check) {
			t.Errorf("Missing field in output: %s", check)
		}
	}
}

// TestScanContentDirectory_Sorting tests that entries are sorted deterministically.
func TestScanContentDirectory_Sorting(t *testing.T) {
	tempDir := t.TempDir()

	// Create content in non-alphabetical order
	files := []struct {
		typ  string
		name string
	}{
		{"pages", "zebra.json"},
		{"blog", "article.json"},
		{"pages", "about.json"},
		{"news", "update.json"},
	}

	for _, f := range files {
		dir := filepath.Join(tempDir, f.typ)
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, f.name), []byte("{}"), 0644)
	}

	entries, err := ScanContentDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanContentDirectory failed: %v", err)
	}

	// Verify sorted by type, then path
	if len(entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(entries))
	}

	expectedOrder := []string{
		"blog:article",
		"news:update",
		"pages:about",
		"pages:zebra",
	}

	for i, expected := range expectedOrder {
		actual := entries[i].Type + ":" + entries[i].Path
		if actual != expected {
			t.Errorf("Entry %d: expected %s, got %s", i, expected, actual)
		}
	}
}

// TestParseContentFile_PlentiFormat tests Plenti-compatible format.
func TestParseContentFile_PlentiFormat(t *testing.T) {
	tempDir := t.TempDir()

	// Create a Plenti-style content file
	pagesDir := filepath.Join(tempDir, "pages")
	os.MkdirAll(pagesDir, 0755)

	plentiJSON := `{
		"components": [
			{
				"name": "hero2436",
				"fields": {
					"title": "Welcome",
					"description": "A test page"
				}
			}
		]
	}`
	filePath := filepath.Join(pagesDir, "home.json")
	os.WriteFile(filePath, []byte(plentiJSON), 0644)

	entry, err := ParseContentFile(tempDir, filePath)
	if err != nil {
		t.Fatalf("ParseContentFile failed: %v", err)
	}

	// Verify Plenti format
	if entry.Type != "pages" {
		t.Errorf("Type: expected 'pages', got %q", entry.Type)
	}
	if entry.Path != "home" {
		t.Errorf("Path: expected 'home', got %q", entry.Path)
	}
	if entry.Filename != "home.json" {
		t.Errorf("Filename: expected 'home.json', got %q", entry.Filename)
	}

	// Verify components array is preserved
	components, ok := entry.Fields["components"].([]interface{})
	if !ok {
		t.Fatal("Components field should be an array")
	}
	if len(components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(components))
	}
}

// TestExtractContentType tests content type extraction from paths.
func TestExtractContentType(t *testing.T) {
	tests := []struct {
		relPath  string
		expected string
	}{
		{"pages/about.json", "pages"},
		{"news/article.json", "news"},
		{"blog/2024/post.json", "blog"},
		{"portfolio/projects/web/site.json", "portfolio"},
		{"index.json", ""},          // Root level
		{"single.json", ""},         // Root level
	}

	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			got := extractContentType(tt.relPath)
			if got != tt.expected {
				t.Errorf("extractContentType(%q) = %q, want %q", tt.relPath, got, tt.expected)
			}
		})
	}
}

// TestExtractURLPath tests URL path extraction from file paths.
func TestExtractURLPath(t *testing.T) {
	tests := []struct {
		relPath     string
		contentType string
		expected    string
	}{
		{"pages/about.json", "pages", "about"},
		{"pages/_index.json", "pages", ""},              // _index becomes empty
		{"news/2024/article.json", "news", "2024/article"},
		{"blog/_defaults.json", "blog", "_defaults"},    // _defaults preserved
		{"index.json", "", "index"},                      // Root level
	}

	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			got := extractURLPath(tt.relPath, tt.contentType)
			if got != tt.expected {
				t.Errorf("extractURLPath(%q, %q) = %q, want %q",
					tt.relPath, tt.contentType, got, tt.expected)
			}
		})
	}
}

// TestWriteContentJS tests file writing functionality.
func TestWriteContentJS(t *testing.T) {
	contentDir := t.TempDir()
	outputDir := t.TempDir()

	// Create some content
	pagesDir := filepath.Join(contentDir, "pages")
	os.MkdirAll(pagesDir, 0755)
	os.WriteFile(filepath.Join(pagesDir, "test.json"), []byte(`{"title":"Test"}`), 0644)

	// Write content.js
	outputPath := filepath.Join(outputDir, "generated", "content.js")
	err := WriteContentJS(contentDir, outputPath)
	if err != nil {
		t.Fatalf("WriteContentJS failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Verify content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(data), "const allContent") {
		t.Error("Output file missing content")
	}
}
