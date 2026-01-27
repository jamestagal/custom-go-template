package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/builder"
)

// TestPlentiContentFormat_RealContent tests content generation with real project content.
func TestPlentiContentFormat_RealContent(t *testing.T) {
	// Find project root (where content/ directory is)
	projectRoot := findProjectRoot(t)
	contentDir := filepath.Join(projectRoot, "content")

	// Skip if content directory doesn't exist
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		t.Skip("Content directory not found, skipping real content test")
	}

	// Generate content.js
	js, err := builder.GenerateContentJS(contentDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify Plenti-compatible structure
	verifyPlentiFormat(t, js)
}

// TestPlentiContentFormat_PageLookup tests that content can be looked up by filepath.
func TestPlentiContentFormat_PageLookup(t *testing.T) {
	tempDir := t.TempDir()

	// Create content matching real project structure
	pagesDir := filepath.Join(tempDir, "pages")
	os.MkdirAll(pagesDir, 0755)

	// Create about.json matching real format
	aboutJSON := `{
		"components": [
			{"name": "hero", "fields": {"title": "About Us"}}
		]
	}`
	aboutPath := filepath.Join(pagesDir, "about.json")
	os.WriteFile(aboutPath, []byte(aboutJSON), 0644)

	// Generate content.js
	js, err := builder.GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify filepath field matches Plenti pattern
	// Plenti uses: data-content-filepath="content/pages/about.json"
	expectedFilepath := "content/pages/about.json"
	if !strings.Contains(js, `filepath: "`+expectedFilepath+`"`) {
		t.Errorf("Expected filepath %q in output", expectedFilepath)
		t.Logf("Output:\n%s", js)
	}
}

// TestPlentiContentFormat_NewsContent tests news/blog content type.
func TestPlentiContentFormat_NewsContent(t *testing.T) {
	tempDir := t.TempDir()

	// Create news directory with flat JSON (not components)
	newsDir := filepath.Join(tempDir, "news")
	os.MkdirAll(newsDir, 0755)

	// Create news article with flat fields (Plenti pattern for blog/news)
	articleJSON := `{
		"title": "Product Launch",
		"published": true,
		"author": {
			"name": "John Doe",
			"image": {"src": "/img/john.jpg", "alt": "John"}
		},
		"publish": {
			"date": "2025-02-01"
		},
		"blogImage": {
			"src": "/img/product.jpg",
			"alt": "Product photo"
		}
	}`
	os.WriteFile(filepath.Join(newsDir, "product-launch.json"), []byte(articleJSON), 0644)

	// Generate content.js
	js, err := builder.GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify news type
	if !strings.Contains(js, `type: "news"`) {
		t.Error("Missing news type")
	}

	// Verify path format (no leading slash, no extension)
	if !strings.Contains(js, `path: "product-launch"`) {
		t.Error("Path should be 'product-launch' without extension")
	}

	// Verify nested fields preserved
	if !strings.Contains(js, `"author"`) {
		t.Error("Author field should be preserved")
	}
	if !strings.Contains(js, `"publish"`) {
		t.Error("Publish field should be preserved")
	}
}

// TestPlentiContentFormat_IndexPage tests _index.json handling.
func TestPlentiContentFormat_IndexPage(t *testing.T) {
	tempDir := t.TempDir()

	// Create _index.json (home page)
	pagesDir := filepath.Join(tempDir, "pages")
	os.MkdirAll(pagesDir, 0755)

	indexJSON := `{
		"components": [
			{"name": "hero2436", "fields": {"title": "Welcome"}}
		]
	}`
	os.WriteFile(filepath.Join(pagesDir, "_index.json"), []byte(indexJSON), 0644)

	// Generate content.js
	js, err := builder.GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify _index.json becomes empty path (root)
	// Plenti pattern: _index.json is the root of that type
	if !strings.Contains(js, `path: ""`) {
		t.Error("_index.json should have empty path")
	}
	if !strings.Contains(js, `filepath: "content/pages/_index.json"`) {
		t.Error("Filepath should preserve _index.json")
	}
}

// TestPlentiContentFormat_MultipleTypes tests multiple content types together.
func TestPlentiContentFormat_MultipleTypes(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple content types like Plenti projects
	types := map[string]string{
		"pages": `{"components": [{"name": "hero", "fields": {}}]}`,
		"news":  `{"title": "News Article", "published": true}`,
		"blog":  `{"title": "Blog Post", "date": "2025-01-01"}`,
	}

	for typ, content := range types {
		dir := filepath.Join(tempDir, typ)
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "test.json"), []byte(content), 0644)
	}

	// Generate content.js
	js, err := builder.GenerateContentJS(tempDir)
	if err != nil {
		t.Fatalf("GenerateContentJS failed: %v", err)
	}

	// Verify all types present
	for typ := range types {
		if !strings.Contains(js, `type: "`+typ+`"`) {
			t.Errorf("Missing type: %s", typ)
		}
	}

	// Verify sorted order (blog, news, pages)
	blogPos := strings.Index(js, `type: "blog"`)
	newsPos := strings.Index(js, `type: "news"`)
	pagesPos := strings.Index(js, `type: "pages"`)

	if blogPos > newsPos || newsPos > pagesPos {
		t.Error("Content types should be sorted alphabetically")
	}
}

// TestPlentiContentFormat_DataContentFilepath tests filepath matches HTML attribute.
func TestPlentiContentFormat_DataContentFilepath(t *testing.T) {
	tempDir := t.TempDir()

	// Create content files
	pagesDir := filepath.Join(tempDir, "pages")
	os.MkdirAll(pagesDir, 0755)

	files := []string{"about.json", "contact.json", "_index.json"}
	for _, f := range files {
		os.WriteFile(filepath.Join(pagesDir, f), []byte("{}"), 0644)
	}

	entries, err := builder.ScanContentDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanContentDirectory failed: %v", err)
	}

	// Verify each entry has correct filepath for HTML data-content-filepath
	expectedFilepaths := map[string]string{
		"about":   "content/pages/about.json",
		"contact": "content/pages/contact.json",
		"":        "content/pages/_index.json", // _index becomes empty path
	}

	for _, entry := range entries {
		expected, ok := expectedFilepaths[entry.Path]
		if ok && entry.Filepath != expected {
			t.Errorf("Path %q: filepath = %q, want %q", entry.Path, entry.Filepath, expected)
		}
	}
}

// verifyPlentiFormat checks that the JS output matches Plenti expectations.
func verifyPlentiFormat(t *testing.T, js string) {
	t.Helper()

	// Check structure
	checks := []struct {
		name    string
		pattern string
	}{
		{"const declaration", "const allContent = ["},
		{"export statement", "export default allContent;"},
		{"type field", "type:"},
		{"path field", "path:"},
		{"filepath field", "filepath:"},
		{"filename field", "filename:"},
		{"fields object", "fields:"},
	}

	for _, check := range checks {
		if !strings.Contains(js, check.pattern) {
			t.Errorf("Missing %s: %q", check.name, check.pattern)
		}
	}
}

// findProjectRoot finds the project root directory.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
