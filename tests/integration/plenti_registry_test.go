package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/builder"
	"github.com/jimafisk/custom_go_template/transformer"
	"github.com/jimafisk/custom_go_template/types"
)

// TestPlentiSignatureLookup tests that components can be looked up by Plenti signature
func TestPlentiSignatureLookup(t *testing.T) {
	// Setup: Register a component with file path
	componentAST := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Children: []ast.Node{
					&ast.TextNode{Content: "Hero Component"},
				},
			},
		},
	}

	// Register with file path (this should create signature-based registration)
	transformer.RegisterComponent("layouts/components/Hero2436.html", componentAST, []string{"title"})
	defer transformer.UnregisterComponent("layouts/components/Hero2436.html")
	defer transformer.UnregisterComponent("Hero2436")
	defer transformer.UnregisterComponent("layouts_components_Hero2436_html")

	tests := []struct {
		name       string
		lookupKey  string
		wantExists bool
	}{
		{
			name:       "lookup by file path",
			lookupKey:  "layouts/components/Hero2436.html",
			wantExists: true,
		},
		{
			name:       "lookup by short name",
			lookupKey:  "Hero2436",
			wantExists: true,
		},
		{
			name:       "lookup by Plenti signature",
			lookupKey:  "layouts_components_Hero2436_html",
			wantExists: true,
		},
		{
			name:       "lookup by case-insensitive name",
			lookupKey:  "hero2436",
			wantExists: true,
		},
		{
			name:       "lookup non-existent component",
			lookupKey:  "NonExistent",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, exists := transformer.GetComponentTemplate(tt.lookupKey)
			if exists != tt.wantExists {
				t.Errorf("GetComponentTemplate(%q) exists = %v, want %v", tt.lookupKey, exists, tt.wantExists)
			}
			if tt.wantExists && tmpl == nil {
				t.Errorf("GetComponentTemplate(%q) returned nil template", tt.lookupKey)
			}
		})
	}
}

// TestRegisterComponentTemplate tests the new RegisterComponentTemplate function
func TestRegisterComponentTemplate(t *testing.T) {
	// Create a component template using types.NewComponentTemplate
	componentAST := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "section",
				Children: []ast.Node{
					&ast.TextNode{Content: "Services Component"},
				},
			},
		},
	}

	ct := types.NewComponentTemplate("layouts/components/Services2437.html", componentAST, []string{"items"})
	transformer.RegisterComponentTemplate(ct)

	// Cleanup
	defer transformer.UnregisterComponent("Services2437")
	defer transformer.UnregisterComponent("layouts_components_Services2437_html")
	defer transformer.UnregisterComponent("layouts/components/Services2437.html")

	// Verify registration
	if ct.Name != "Services2437" {
		t.Errorf("Name = %q, want %q", ct.Name, "Services2437")
	}
	if ct.Signature != "layouts_components_Services2437_html" {
		t.Errorf("Signature = %q, want %q", ct.Signature, "layouts_components_Services2437_html")
	}
	if ct.Category != "components" {
		t.Errorf("Category = %q, want %q", ct.Category, "components")
	}

	// Verify lookup works by all keys
	for _, key := range []string{"Services2437", "layouts_components_Services2437_html"} {
		tmpl, exists := transformer.GetComponentTemplate(key)
		if !exists {
			t.Errorf("GetComponentTemplate(%q) not found", key)
		}
		if tmpl.Name != "Services2437" {
			t.Errorf("GetComponentTemplate(%q).Name = %q, want %q", key, tmpl.Name, "Services2437")
		}
	}
}

// TestGenerateComponentRegistryWithSignatures tests that the registry generator
// produces Plenti-compatible output with signatures as primary keys
func TestGenerateComponentRegistryWithSignatures(t *testing.T) {
	// Create test components
	hero := types.NewComponentTemplate("layouts/components/Hero2436.html", &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{TagName: "div", Children: []ast.Node{&ast.TextNode{Content: "Hero"}}},
		},
	}, []string{"title"})

	nav := types.NewComponentTemplate("layouts/global/nav.html", &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{TagName: "nav", Children: []ast.Node{&ast.TextNode{Content: "Nav"}}},
		},
	}, nil)

	components := []builder.ComponentTemplate{*hero, *nav}
	registryJS := builder.GenerateComponentRegistry(components)

	// Verify signature-based keys
	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "has header comment",
			contains: "// Auto-generated Plenti-compatible component registry",
		},
		{
			name:     "has hero signature key",
			contains: "'layouts_components_Hero2436_html'",
		},
		{
			name:     "has nav signature key",
			contains: "'layouts_global_nav_html'",
		},
		{
			name:     "has hero short name alias",
			contains: "'Hero2436'",
		},
		{
			name:     "has nav short name alias",
			contains: "'nav'",
		},
		{
			name:     "exports registry",
			contains: "export default registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(registryJS, tt.contains) {
				t.Errorf("Registry output missing %q\n\nGot:\n%s", tt.contains, registryJS)
			}
		})
	}
}

// TestComponentTemplateFieldsPopulated ensures all fields are correctly populated
func TestComponentTemplateFieldsPopulated(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantName string
		wantSig  string
		wantCat  string
	}{
		{
			name:     "standard component",
			filePath: "layouts/components/Hero2436.html",
			wantName: "Hero2436",
			wantSig:  "layouts_components_Hero2436_html",
			wantCat:  "components",
		},
		{
			name:     "content template",
			filePath: "layouts/content/pages.html",
			wantName: "pages",
			wantSig:  "layouts_content_pages_html",
			wantCat:  "content",
		},
		{
			name:     "global layout",
			filePath: "layouts/global/html.html",
			wantName: "html",
			wantSig:  "layouts_global_html_html",
			wantCat:  "global",
		},
		{
			name:     "leading underscore",
			filePath: "layouts/content/_index.html",
			wantName: "_index",
			wantSig:  "layouts_content__index_html",
			wantCat:  "content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := types.NewComponentTemplate(tt.filePath, nil, nil)

			if ct.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", ct.Name, tt.wantName)
			}
			if ct.Signature != tt.wantSig {
				t.Errorf("Signature = %q, want %q", ct.Signature, tt.wantSig)
			}
			if ct.Category != tt.wantCat {
				t.Errorf("Category = %q, want %q", ct.Category, tt.wantCat)
			}
			if ct.FilePath != tt.filePath {
				t.Errorf("FilePath = %q, want %q", ct.FilePath, tt.filePath)
			}
		})
	}
}

// TestDualRegistrationBackwardCompatibility ensures both old and new registration methods work
func TestDualRegistrationBackwardCompatibility(t *testing.T) {
	componentAST := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{TagName: "div"},
		},
	}

	// Method 1: Old-style registration with just component name
	transformer.RegisterComponent("OldStyleComponent", componentAST, nil)
	defer transformer.UnregisterComponent("OldStyleComponent")

	// Method 2: New-style registration with file path
	transformer.RegisterComponent("layouts/components/NewStyleComponent.html", componentAST, nil)
	defer transformer.UnregisterComponent("NewStyleComponent")
	defer transformer.UnregisterComponent("layouts/components/NewStyleComponent.html")
	defer transformer.UnregisterComponent("layouts_components_NewStyleComponent_html")

	// Both should be findable
	if _, exists := transformer.GetComponentTemplate("OldStyleComponent"); !exists {
		t.Error("Old-style registered component not found")
	}

	if _, exists := transformer.GetComponentTemplate("NewStyleComponent"); !exists {
		t.Error("New-style registered component not found by short name")
	}

	if _, exists := transformer.GetComponentTemplate("layouts_components_NewStyleComponent_html"); !exists {
		t.Error("New-style registered component not found by signature")
	}
}
