package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestComponentRegistryNormalization tests that components can be registered and found
// with multiple path variants
func TestComponentRegistryNormalization(t *testing.T) {
	// Clear registry before test
	componentTemplateRegistry = make(map[string]*ComponentTemplate)

	// Create a test component template
	testTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Children: []ast.Node{
					&ast.TextNode{Content: "Test Component"},
				},
			},
		},
	}

	tests := []struct {
		name           string
		registerAs     string
		lookupKeys     []string
		shouldFindAll  bool
	}{
		{
			name:       "Component registered by name",
			registerAs: "UserProfile",
			lookupKeys: []string{
				"UserProfile",
				"./components/UserProfile.html",
				"UserProfile.html",
			},
			shouldFindAll: true,
		},
		{
			name:       "Component registered by path",
			registerAs: "./components/Header.html",
			lookupKeys: []string{
				"Header",
				"./components/Header.html",
				"Header.html",
			},
			shouldFindAll: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear registry
			componentTemplateRegistry = make(map[string]*ComponentTemplate)

			// Register component with all normalized keys
			keys := normalizeComponentPath(tt.registerAs)
			for _, key := range keys {
				RegisterComponent(key, testTemplate, []string{})
			}

			// Try to look up with all expected keys
			for _, lookupKey := range tt.lookupKeys {
				_, exists := GetComponentTemplate(lookupKey)
				if tt.shouldFindAll && !exists {
					t.Errorf("Component registered as %q should be found with key %q, but wasn't",
						tt.registerAs, lookupKey)
				}
			}
		})
	}
}

// TestNormalizeComponentPath tests the path normalization function
func TestNormalizeComponentPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:  "Full relative path",
			input: "./components/UserProfile.html",
			wantKeys: []string{
				"./components/UserProfile.html", // original
				"UserProfile.html",              // filename
				"UserProfile",                   // name without extension
				"./components/UserProfile",      // path without extension
			},
		},
		{
			name:  "Component name only",
			input: "Header",
			wantKeys: []string{
				"Header", // original
			},
		},
		{
			name:  "Filename with extension",
			input: "Footer.html",
			wantKeys: []string{
				"Footer.html", // original
				"Footer",      // name without extension
			},
		},
		{
			name:  "Absolute path",
			input: "examples/components/Card.html",
			wantKeys: []string{
				"examples/components/Card.html", // original
				"Card.html",                     // filename
				"Card",                          // name without extension
				"examples/components/Card",      // path without extension
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeComponentPath(tt.input)

			// Check that all expected keys are present
			for _, wantKey := range tt.wantKeys {
				found := false
				for _, gotKey := range got {
					if gotKey == wantKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("normalizeComponentPath(%q) missing expected key %q\nGot: %v",
						tt.input, wantKey, got)
				}
			}
		})
	}
}
