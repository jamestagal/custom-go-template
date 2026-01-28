package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/types"
)

// TestGenerateBundleHash tests content hash generation.
func TestGenerateBundleHash(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		wantLen  int
		wantSame bool // If true, same content should produce same hash
	}{
		{
			name:     "simple content",
			content:  []byte("hello world"),
			wantLen:  8,
			wantSame: true,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			wantLen:  8,
			wantSame: true,
		},
		{
			name:     "complex content",
			content:  []byte(`const component = (props) => '<div>${props.title}</div>'`),
			wantLen:  8,
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := GenerateBundleHash(tt.content)
			hash2 := GenerateBundleHash(tt.content)

			if len(hash1) != tt.wantLen {
				t.Errorf("GenerateBundleHash() len = %d, want %d", len(hash1), tt.wantLen)
			}

			if tt.wantSame && hash1 != hash2 {
				t.Errorf("GenerateBundleHash() not deterministic: %s != %s", hash1, hash2)
			}
		})
	}

	// Test different content produces different hashes
	t.Run("different content different hash", func(t *testing.T) {
		hash1 := GenerateBundleHash([]byte("content1"))
		hash2 := GenerateBundleHash([]byte("content2"))
		if hash1 == hash2 {
			t.Error("Different content should produce different hashes")
		}
	})
}

// TestGenerateBundleHashFromComponents tests deterministic component hashing.
func TestGenerateBundleHashFromComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []string
		wantSame   bool
	}{
		{
			name:       "single component",
			components: []string{"hero"},
			wantSame:   true,
		},
		{
			name:       "multiple components",
			components: []string{"hero", "footer", "nav"},
			wantSame:   true,
		},
		{
			name:       "empty list",
			components: []string{},
			wantSame:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := GenerateBundleHashFromComponents(tt.components)
			hash2 := GenerateBundleHashFromComponents(tt.components)

			if tt.wantSame && hash1 != hash2 {
				t.Errorf("Hash not deterministic: %s != %s", hash1, hash2)
			}
		})
	}

	// Test order independence
	t.Run("order independent", func(t *testing.T) {
		hash1 := GenerateBundleHashFromComponents([]string{"a", "b", "c"})
		hash2 := GenerateBundleHashFromComponents([]string{"c", "a", "b"})
		if hash1 != hash2 {
			t.Error("Component order should not affect hash")
		}
	})
}

// TestUsageAnalyzer tests component usage analysis.
func TestUsageAnalyzer(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	layoutsDir := filepath.Join(tempDir, "layouts")

	// Create directories
	if err := os.MkdirAll(filepath.Join(contentDir, "pages"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(layoutsDir, "global"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(layoutsDir, "content"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create test content files
	indexContent := map[string]any{
		"components": []map[string]any{
			{"name": "hero", "fields": map[string]any{}},
			{"name": "services", "fields": map[string]any{}},
		},
	}
	indexJSON, _ := json.Marshal(indexContent)
	if err := os.WriteFile(filepath.Join(contentDir, "pages", "_index.json"), indexJSON, 0644); err != nil {
		t.Fatal(err)
	}

	aboutContent := map[string]any{
		"components": []map[string]any{
			{"name": "hero", "fields": map[string]any{}},
			{"name": "team", "fields": map[string]any{}},
		},
	}
	aboutJSON, _ := json.Marshal(aboutContent)
	if err := os.WriteFile(filepath.Join(contentDir, "pages", "about.json"), aboutJSON, 0644); err != nil {
		t.Fatal(err)
	}

	// Create global component file
	if err := os.WriteFile(filepath.Join(layoutsDir, "global", "nav.html"), []byte("<nav></nav>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run analyzer
	analyzer := NewUsageAnalyzer(contentDir, layoutsDir)
	usage, err := analyzer.Analyze()
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}

	// Verify results
	t.Run("pages found", func(t *testing.T) {
		if len(usage) != 2 {
			t.Errorf("Expected 2 pages, got %d", len(usage))
		}
	})

	t.Run("index components", func(t *testing.T) {
		indexUsage := usage["pages/_index"]
		if indexUsage == nil {
			t.Fatal("pages/_index not found")
		}
		// Should have hero, services, and nav (global)
		if len(indexUsage.Components) < 2 {
			t.Errorf("Expected at least 2 components, got %d", len(indexUsage.Components))
		}
	})

	t.Run("common components", func(t *testing.T) {
		common := analyzer.GetCommonComponents(2)
		// hero is used on both pages
		found := false
		for _, c := range common {
			if c == "hero" {
				found = true
				break
			}
		}
		if !found {
			t.Error("hero should be in common components")
		}
	})

	t.Run("page unique components", func(t *testing.T) {
		common := analyzer.GetCommonComponents(2)
		unique := analyzer.GetPageComponents("pages/_index", common)
		// services should be unique to _index
		found := false
		for _, c := range unique {
			if c == "services" {
				found = true
				break
			}
		}
		if !found {
			t.Error("services should be in unique components for _index")
		}
	})
}

// TestCommonChunkGeneration tests common chunk generation.
func TestCommonChunkGeneration(t *testing.T) {
	// Create mock registry
	registry := map[string]*types.ComponentTemplate{
		"hero": {
			Name: "hero",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{Name: "class", Value: "hero"},
						},
					},
				},
			},
		},
		"footer": {
			Name: "footer",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "footer",
					},
				},
			},
		},
	}

	t.Run("empty common components", func(t *testing.T) {
		result, err := GenerateCommonChunk([]string{}, registry)
		if err != nil {
			t.Fatalf("GenerateCommonChunk() error: %v", err)
		}
		if len(result.Components) != 0 {
			t.Error("Expected empty components list")
		}
	})

	t.Run("with common components", func(t *testing.T) {
		result, err := GenerateCommonChunk([]string{"hero", "footer"}, registry)
		if err != nil {
			t.Fatalf("GenerateCommonChunk() error: %v", err)
		}

		if len(result.Components) != 2 {
			t.Errorf("Expected 2 components, got %d", len(result.Components))
		}

		if len(result.Hash) != 8 {
			t.Errorf("Expected 8-char hash, got %d chars", len(result.Hash))
		}

		// Content should include component functions
		content := string(result.Content)
		if !strings.Contains(content, "'hero':") {
			t.Error("Content should include hero component")
		}
		if !strings.Contains(content, "'footer':") {
			t.Error("Content should include footer component")
		}
		if !strings.Contains(content, "export default common") {
			t.Error("Content should export common")
		}
	})
}

// TestPageBundleGeneration tests page bundle generation.
func TestPageBundleGeneration(t *testing.T) {
	registry := map[string]*types.ComponentTemplate{
		"services": {
			Name: "services",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{TagName: "section"},
				},
			},
		},
	}

	t.Run("with common chunk", func(t *testing.T) {
		result, err := GeneratePageBundle(
			"pages/_index",
			[]string{"services"},
			[]string{"hero"},
			registry,
			"../common.abc123.js",
		)
		if err != nil {
			t.Fatalf("GeneratePageBundle() error: %v", err)
		}

		if !result.IncludesCommon {
			t.Error("Should include common chunk")
		}

		content := string(result.Content)
		if !strings.Contains(content, "import common from") {
			t.Error("Should import common chunk")
		}
		if !strings.Contains(content, "export default { ...common, ...unique }") {
			t.Error("Should export merged registry")
		}
	})

	t.Run("without common chunk", func(t *testing.T) {
		result, err := GeneratePageBundle(
			"pages/about",
			[]string{"services"},
			[]string{},
			registry,
			"",
		)
		if err != nil {
			t.Fatalf("GeneratePageBundle() error: %v", err)
		}

		if result.IncludesCommon {
			t.Error("Should not include common chunk")
		}

		content := string(result.Content)
		if strings.Contains(content, "import common") {
			t.Error("Should not import common chunk")
		}
	})
}

// TestBundleManifest tests manifest generation.
func TestBundleManifest(t *testing.T) {
	commonResult := &CommonChunkResult{
		Content:    []byte("common content"),
		Components: []string{"hero"},
		Hash:       "abc12345",
		Size:       1000,
	}

	pageBundles := map[string]*PageBundleResult{
		"pages/_index": {
			Page:             "pages/_index",
			Content:          []byte("index content"),
			UniqueComponents: []string{"services"},
			IncludesCommon:   true,
			Hash:             "def67890",
			TotalSize:        2000,
			UniqueSize:       1000,
		},
	}

	manifest, err := GenerateBundleManifest(
		"fingerprint123",
		commonResult,
		pageBundles,
		"/generated/layouts.js",
		10000,
		10,
		3,
	)
	if err != nil {
		t.Fatalf("GenerateBundleManifest() error: %v", err)
	}

	t.Run("version generated", func(t *testing.T) {
		if manifest.Version == "" {
			t.Error("Version should be set")
		}
	})

	t.Run("common chunk included", func(t *testing.T) {
		if manifest.Common == nil {
			t.Fatal("Common should be set")
		}
		if manifest.Common.Path == "" {
			t.Error("Common path should be set")
		}
	})

	t.Run("page bundles included", func(t *testing.T) {
		if len(manifest.Bundles) != 1 {
			t.Errorf("Expected 1 bundle, got %d", len(manifest.Bundles))
		}
		bundle := manifest.Bundles["pages/_index"]
		if bundle == nil {
			t.Fatal("pages/_index bundle not found")
		}
		if !bundle.IncludesCommon {
			t.Error("Bundle should include common")
		}
	})

	t.Run("stats calculated", func(t *testing.T) {
		if manifest.Stats == nil {
			t.Fatal("Stats should be set")
		}
		if manifest.Stats.TotalComponents != 10 {
			t.Errorf("Expected 10 total components, got %d", manifest.Stats.TotalComponents)
		}
		if manifest.Stats.OrphanComponents != 3 {
			t.Errorf("Expected 3 orphan components, got %d", manifest.Stats.OrphanComponents)
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		jsonBytes, err := manifest.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON() error: %v", err)
		}
		if len(jsonBytes) == 0 {
			t.Error("JSON should not be empty")
		}
	})

	t.Run("script tag attributes", func(t *testing.T) {
		attrs := manifest.GetScriptTagAttributes("pages/_index")
		if attrs["data-bundle"] == "" {
			t.Error("data-bundle should be set")
		}
		if attrs["data-common"] == "" {
			t.Error("data-common should be set for page with common")
		}
		if attrs["data-fallback"] == "" {
			t.Error("data-fallback should be set")
		}
	})
}
