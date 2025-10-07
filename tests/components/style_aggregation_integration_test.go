package components

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// Helper function to render template and extract styles
func renderAndExtractStyles(t *testing.T, template *ast.Template) string {
	t.Helper()

	// IMPORTANT: Call AggregateComponentStyles BEFORE transformation
	// because it needs to see the imports in the FenceSection
	aggregatedStyles := renderer.AggregateComponentStyles(template, "page")

	return aggregatedStyles
}

// extractStylesFromTemplate calls the style aggregation directly
// This simulates what the renderer should do
func extractStylesFromTemplate(template *ast.Template) string {
	// Call AggregateComponentStyles on the original (pre-transform) template
	return renderer.AggregateComponentStyles(template, "page")
}

// TestRenderTemplate_HeaderSimpleStylesIncluded tests that HeaderSimple component
// styles are automatically included in rendered output
func TestRenderTemplate_HeaderSimpleStylesIncluded(t *testing.T) {
	// Create HeaderSimple component with styles
	headerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.header { background-color: #f8f9fa; }
.brand svg { height: 32px; }`,
			},
			&ast.Element{
				TagName: "header",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "header"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Header Content"},
				},
			},
		},
	}

	// Register HeaderSimple component
	transformer.RegisterComponent("HeaderSimple", headerTemplate, []string{})
	defer transformer.UnregisterComponent("HeaderSimple")

	// Create page template that imports HeaderSimple
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "HeaderSimple", Path: "./components/HeaderSimple.html"},
				},
			},
			&ast.ComponentNode{
				Name:    "HeaderSimple",
				Props:   []ast.ComponentProp{},
				Dynamic: false,
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify styles are included
	if !strings.Contains(style, "/* Styles from: HeaderSimple */") {
		t.Errorf("Expected source comment for HeaderSimple, got:\n%s", style)
	}

	if !strings.Contains(style, ".header { background-color: #f8f9fa; }") {
		t.Errorf("Expected .header styles, got:\n%s", style)
	}

	if !strings.Contains(style, ".brand svg { height: 32px; }") {
		t.Errorf("Expected .brand svg styles, got:\n%s", style)
	}
}

// TestRenderTemplate_NestedComponentStyles tests that styles from nested
// component hierarchies are aggregated correctly (dependencies first)
func TestRenderTemplate_NestedComponentStyles(t *testing.T) {
	// Create Button component (leaf component)
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.button { padding: 0.5rem; }`,
			},
			&ast.Element{
				TagName:  "button",
				Children: []ast.Node{&ast.TextNode{Content: "Click"}},
			},
		},
	}
	transformer.RegisterComponent("Button", buttonTemplate, []string{})
	defer transformer.UnregisterComponent("Button")

	// Create Card component (imports Button)
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Button", Path: "./components/Button.html"},
				},
			},
			&ast.StyleSection{
				Content: `.card { border: 1px solid #ccc; }`,
			},
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "card"},
				},
				Children: []ast.Node{
					&ast.ComponentNode{Name: "Button"},
				},
			},
		},
	}
	transformer.RegisterComponent("Card", cardTemplate, []string{})
	defer transformer.UnregisterComponent("Card")

	// Create page that imports Card
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Card", Path: "./components/Card.html"},
				},
			},
			&ast.ComponentNode{Name: "Card"},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify both components' styles are included
	if !strings.Contains(style, "/* Styles from: Button */") {
		t.Errorf("Expected Button styles comment")
	}
	if !strings.Contains(style, "/* Styles from: Card */") {
		t.Errorf("Expected Card styles comment")
	}

	if !strings.Contains(style, ".button { padding: 0.5rem; }") {
		t.Errorf("Expected Button styles")
	}
	if !strings.Contains(style, ".card { border: 1px solid #ccc; }") {
		t.Errorf("Expected Card styles")
	}

	// Verify correct order: Button (dependency) comes before Card (parent)
	buttonIndex := strings.Index(style, "/* Styles from: Button */")
	cardIndex := strings.Index(style, "/* Styles from: Card */")

	if buttonIndex == -1 || cardIndex == -1 {
		t.Fatalf("Could not find style comments in output")
	}

	if buttonIndex >= cardIndex {
		t.Errorf("Expected Button styles before Card styles (dependencies first), got:\n%s", style)
	}
}

// TestRenderTemplate_StylesInjectedOnce tests that styles appear only once
// even when component is used multiple times on the page
func TestRenderTemplate_StylesInjectedOnce(t *testing.T) {
	// Create component with styles
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.unique-style { color: blue; }`,
			},
			&ast.Element{
				TagName:  "div",
				Children: []ast.Node{&ast.TextNode{Content: "Content"}},
			},
		},
	}
	transformer.RegisterComponent("ReusableComponent", componentTemplate, []string{})
	defer transformer.UnregisterComponent("ReusableComponent")

	// Create page that uses component multiple times
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "ReusableComponent", Path: "./components/ReusableComponent.html"},
				},
			},
			&ast.ComponentNode{Name: "ReusableComponent"},
			&ast.ComponentNode{Name: "ReusableComponent"},
			&ast.ComponentNode{Name: "ReusableComponent"},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Count occurrences of the style
	count := strings.Count(style, ".unique-style { color: blue; }")

	if count != 1 {
		t.Errorf("Expected styles to appear exactly once, but found %d occurrences:\n%s", count, style)
	}

	// Count source comments
	commentCount := strings.Count(style, "/* Styles from: ReusableComponent */")
	if commentCount != 1 {
		t.Errorf("Expected source comment exactly once, but found %d occurrences", commentCount)
	}
}

// TestRenderTemplate_MultipleComponentStyles tests that styles from multiple
// different components are all included
func TestRenderTemplate_MultipleComponentStyles(t *testing.T) {
	// Create Header component
	headerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.header { background: white; }`,
			},
			&ast.Element{TagName: "header"},
		},
	}
	transformer.RegisterComponent("Header", headerTemplate, []string{})
	defer transformer.UnregisterComponent("Header")

	// Create Footer component
	footerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.footer { background: black; }`,
			},
			&ast.Element{TagName: "footer"},
		},
	}
	transformer.RegisterComponent("Footer", footerTemplate, []string{})
	defer transformer.UnregisterComponent("Footer")

	// Create Sidebar component
	sidebarTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.sidebar { width: 200px; }`,
			},
			&ast.Element{TagName: "aside"},
		},
	}
	transformer.RegisterComponent("Sidebar", sidebarTemplate, []string{})
	defer transformer.UnregisterComponent("Sidebar")

	// Create page using all three components
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Header", Path: "./components/Header.html"},
					{Name: "Footer", Path: "./components/Footer.html"},
					{Name: "Sidebar", Path: "./components/Sidebar.html"},
				},
			},
			&ast.ComponentNode{Name: "Header"},
			&ast.ComponentNode{Name: "Sidebar"},
			&ast.ComponentNode{Name: "Footer"},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify all components' styles are included
	expectedStyles := []string{
		"/* Styles from: Header */",
		".header { background: white; }",
		"/* Styles from: Footer */",
		".footer { background: black; }",
		"/* Styles from: Sidebar */",
		".sidebar { width: 200px; }",
	}

	for _, expected := range expectedStyles {
		if !strings.Contains(style, expected) {
			t.Errorf("Expected to find %q in output:\n%s", expected, style)
		}
	}
}

// TestRenderTemplate_StylesOrderedCorrectly tests that styles appear in
// dependency-first order (child components before parents)
func TestRenderTemplate_StylesOrderedCorrectly(t *testing.T) {
	// Create Icon component (deepest dependency)
	iconTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.icon { width: 16px; }`,
			},
			&ast.Element{TagName: "svg"},
		},
	}
	transformer.RegisterComponent("Icon", iconTemplate, []string{})
	defer transformer.UnregisterComponent("Icon")

	// Create Button component (imports Icon)
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Icon", Path: "./components/Icon.html"},
				},
			},
			&ast.StyleSection{
				Content: `.button { padding: 8px; }`,
			},
			&ast.Element{
				TagName: "button",
				Children: []ast.Node{
					&ast.ComponentNode{Name: "Icon"},
				},
			},
		},
	}
	transformer.RegisterComponent("Button", buttonTemplate, []string{})
	defer transformer.UnregisterComponent("Button")

	// Create Toolbar component (imports Button)
	toolbarTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Button", Path: "./components/Button.html"},
				},
			},
			&ast.StyleSection{
				Content: `.toolbar { display: flex; }`,
			},
			&ast.Element{
				TagName: "div",
				Children: []ast.Node{
					&ast.ComponentNode{Name: "Button"},
				},
			},
		},
	}
	transformer.RegisterComponent("Toolbar", toolbarTemplate, []string{})
	defer transformer.UnregisterComponent("Toolbar")

	// Create page (imports Toolbar)
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Toolbar", Path: "./components/Toolbar.html"},
				},
			},
			&ast.ComponentNode{Name: "Toolbar"},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Find positions of each style block
	iconPos := strings.Index(style, "/* Styles from: Icon */")
	buttonPos := strings.Index(style, "/* Styles from: Button */")
	toolbarPos := strings.Index(style, "/* Styles from: Toolbar */")

	// Verify all are present
	if iconPos == -1 || buttonPos == -1 || toolbarPos == -1 {
		t.Fatalf("Not all style comments found in output:\n%s", style)
	}

	// Verify correct order: Icon < Button < Toolbar (dependencies first)
	if !(iconPos < buttonPos && buttonPos < toolbarPos) {
		t.Errorf("Styles not in dependency order. Icon: %d, Button: %d, Toolbar: %d\nOutput:\n%s",
			iconPos, buttonPos, toolbarPos, style)
	}
}

// TestRenderTemplate_RealWorldHomePage tests with actual HeaderSimple component
// to validate the real-world use case
func TestRenderTemplate_RealWorldHomePage(t *testing.T) {
	// Parse real HeaderSimple component
	headerSimpleContent := `---
---

<style>
  .header {
    background-color: #f8f9fa;
    padding: 1rem 0;
    border-bottom: 1px solid #e9ecef;
    margin-bottom: 2rem;
  }

  .brand svg {
    height: 32px;
    width: auto;
  }
</style>

<header class="header">
  <a href="/" class="brand">
    <svg>...</svg>
  </a>
</header>`

	headerTemplate, err := parser.ParseTemplate(headerSimpleContent)
	if err != nil {
		t.Fatalf("Failed to parse HeaderSimple: %v", err)
	}

	transformer.RegisterComponent("HeaderSimple", headerTemplate, []string{})
	defer transformer.UnregisterComponent("HeaderSimple")

	// Create simplified home page that imports HeaderSimple
	homeTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "HeaderSimple", Path: "./components/HeaderSimple.html"},
				},
			},
			&ast.Element{
				TagName: "html",
				Children: []ast.Node{
					&ast.ComponentNode{Name: "HeaderSimple"},
					&ast.Element{
						TagName:  "main",
						Children: []ast.Node{&ast.TextNode{Content: "Home content"}},
					},
				},
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, homeTemplate)

	// Verify HeaderSimple styles are present
	if !strings.Contains(style, "/* Styles from: HeaderSimple */") {
		t.Errorf("Expected HeaderSimple source comment in output")
	}

	if !strings.Contains(style, ".header {") {
		t.Errorf("Expected .header styles in output")
	}

	if !strings.Contains(style, "background-color: #f8f9fa;") {
		t.Errorf("Expected background-color in output")
	}

	if !strings.Contains(style, ".brand svg {") {
		t.Errorf("Expected .brand svg styles in output")
	}

	if !strings.Contains(style, "height: 32px;") {
		t.Errorf("Expected height: 32px in output")
	}

	// Verify styles appear only once
	count := strings.Count(style, "/* Styles from: HeaderSimple */")
	if count != 1 {
		t.Errorf("Expected HeaderSimple styles exactly once, found %d times", count)
	}
}

// TestRenderTemplate_EmptyStylesNotInjected tests that components without
// styles don't create empty style sections
func TestRenderTemplate_EmptyStylesNotInjected(t *testing.T) {
	// Create component without any styles
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName:  "div",
				Children: []ast.Node{&ast.TextNode{Content: "No styles"}},
			},
		},
	}
	transformer.RegisterComponent("NoStyleComponent", componentTemplate, []string{})
	defer transformer.UnregisterComponent("NoStyleComponent")

	// Create page using component
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "NoStyleComponent", Path: "./components/NoStyleComponent.html"},
				},
			},
			&ast.ComponentNode{Name: "NoStyleComponent"},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify no style content is generated
	if strings.TrimSpace(style) != "" {
		t.Errorf("Expected empty style output for component without styles, got:\n%s", style)
	}
}

// TestRenderTemplate_PageStylesCombinedWithComponentStyles tests that page-level
// styles are preserved and combined with component styles
func TestRenderTemplate_PageStylesCombinedWithComponentStyles(t *testing.T) {
	// Create component with styles
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.component { color: blue; }`,
			},
			&ast.Element{TagName: "div"},
		},
	}
	transformer.RegisterComponent("StyledComponent", componentTemplate, []string{})
	defer transformer.UnregisterComponent("StyledComponent")

	// Create page with its own styles AND imported component
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "StyledComponent", Path: "./components/StyledComponent.html"},
				},
			},
			&ast.StyleSection{
				Content: `body { margin: 0; }
h1 { color: red; }`,
			},
			&ast.ComponentNode{Name: "StyledComponent"},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify component styles are included
	if !strings.Contains(style, "/* Styles from: StyledComponent */") {
		t.Errorf("Expected component styles")
	}
	if !strings.Contains(style, ".component { color: blue; }") {
		t.Errorf("Expected .component style")
	}

	// Verify page styles are also included
	if !strings.Contains(style, "body { margin: 0; }") {
		t.Errorf("Expected page body style")
	}
	if !strings.Contains(style, "h1 { color: red; }") {
		t.Errorf("Expected page h1 style")
	}

	// Component styles should come before page styles (dependencies first)
	componentPos := strings.Index(style, "/* Styles from: StyledComponent */")
	pageBodyPos := strings.Index(style, "body { margin: 0; }")

	if componentPos == -1 || pageBodyPos == -1 {
		t.Fatalf("Could not find styles in output:\n%s", style)
	}

	if componentPos >= pageBodyPos {
		t.Errorf("Expected component styles before page styles, got:\n%s", style)
	}
}

// ============================================================================
// DYNAMIC COMPONENT TESTS (NEW - Bug Fix for <= syntax)
// ============================================================================

// TestRenderTemplate_DynamicComponentStyles tests that dynamic components
// using the <= syntax have their styles aggregated correctly
func TestRenderTemplate_DynamicComponentStyles(t *testing.T) {
	// Create UserProfile component with extensive styles (like the real one)
	userProfileTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.profile-card {
    border-radius: 0.5rem;
    background-color: white;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }

  .profile-header {
    padding: 1.5rem;
    background-color: #f8f9fa;
  }

  .profile-name {
    font-size: 1.25rem;
    font-weight: bold;
  }`,
			},
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "profile-card"},
				},
			},
		},
	}
	transformer.RegisterComponent("UserProfile", userProfileTemplate, []string{})
	defer transformer.UnregisterComponent("UserProfile")

	// Create page template with DynamicComponentNode (static path)
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.DynamicComponentNode{
				PathExpression: "./components/UserProfile.html",
				Props:          []ast.ComponentProp{},
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify UserProfile styles included
	if !strings.Contains(style, "/* Styles from: UserProfile */") {
		t.Errorf("Expected UserProfile source comment in result, got:\n%s", style)
	}

	// Check specific CSS classes
	expectedClasses := []string{
		".profile-card",
		".profile-header",
		".profile-name",
	}

	for _, class := range expectedClasses {
		if !strings.Contains(style, class) {
			t.Errorf("Expected UserProfile style %q in result, got:\n%s", class, style)
		}
	}

	// Check specific properties
	if !strings.Contains(style, "border-radius: 0.5rem") {
		t.Errorf("Expected border-radius property")
	}

	if !strings.Contains(style, "box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1)") {
		t.Errorf("Expected box-shadow property")
	}
}

// TestRenderTemplate_MixedImportsAndDynamicComponents tests that both
// FenceSection imports and dynamic component usage work together
func TestRenderTemplate_MixedImportsAndDynamicComponents(t *testing.T) {
	// Create Header component
	headerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.header { background: #f8f9fa; }`,
			},
			&ast.Element{TagName: "header"},
		},
	}
	transformer.RegisterComponent("Header", headerTemplate, []string{})
	defer transformer.UnregisterComponent("Header")

	// Create Footer component
	footerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.footer { background: #343a40; }`,
			},
			&ast.Element{TagName: "footer"},
		},
	}
	transformer.RegisterComponent("Footer", footerTemplate, []string{})
	defer transformer.UnregisterComponent("Footer")

	// Create page with Header imported but Footer as dynamic component
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Header", Path: "./components/Header.html"},
				},
			},
			&ast.ComponentNode{Name: "Header"},
			&ast.DynamicComponentNode{
				PathExpression: "./components/Footer.html",
				Props:          []ast.ComponentProp{},
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Both Header (imported) and Footer (dynamic) should be included
	if !strings.Contains(style, ".header") {
		t.Errorf("Expected Header styles (from import) to be aggregated")
	}

	if !strings.Contains(style, ".footer") {
		t.Errorf("Expected Footer styles (from dynamic component) to be aggregated")
	}

	// Verify source comments for both
	if !strings.Contains(style, "/* Styles from: Header */") {
		t.Errorf("Expected Header source comment")
	}

	if !strings.Contains(style, "/* Styles from: Footer */") {
		t.Errorf("Expected Footer source comment")
	}
}

// TestRenderTemplate_DynamicComponentInConditional tests dynamic components
// inside conditional blocks
func TestRenderTemplate_DynamicComponentInConditional(t *testing.T) {
	// Create Alert component
	alertTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.alert { padding: 1rem; border-radius: 0.375rem; }`,
			},
			&ast.Element{TagName: "div"},
		},
	}
	transformer.RegisterComponent("Alert", alertTemplate, []string{})
	defer transformer.UnregisterComponent("Alert")

	// Create page with dynamic component in conditional
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Conditional{
				IfCondition: "showAlert",
				IfContent: []ast.Node{
					&ast.DynamicComponentNode{
						PathExpression: "./components/Alert.html",
						Props:          []ast.ComponentProp{},
					},
				},
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify Alert styles included even though in conditional
	if !strings.Contains(style, "/* Styles from: Alert */") {
		t.Errorf("Expected Alert source comment")
	}

	if !strings.Contains(style, ".alert") {
		t.Errorf("Expected Alert styles")
	}
}

// TestRenderTemplate_DynamicComponentInLoop tests dynamic components
// inside loop blocks
func TestRenderTemplate_DynamicComponentInLoop(t *testing.T) {
	// Create Card component
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `.card { border: 1px solid #ddd; margin: 1rem; }`,
			},
			&ast.Element{TagName: "div"},
		},
	}
	transformer.RegisterComponent("Card", cardTemplate, []string{})
	defer transformer.UnregisterComponent("Card")

	// Create page with dynamic component in loop
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Loop{
				Iterator:   "item",
				Collection: "items",
				Content: []ast.Node{
					&ast.DynamicComponentNode{
						PathExpression: "./components/Card.html",
						Props:          []ast.ComponentProp{},
					},
				},
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Verify Card styles included even though in loop
	if !strings.Contains(style, "/* Styles from: Card */") {
		t.Errorf("Expected Card source comment")
	}

	if !strings.Contains(style, ".card") {
		t.Errorf("Expected Card styles")
	}
}

// TestRenderTemplate_RuntimeDynamicComponentsSkipped tests that runtime
// variable paths are gracefully skipped (cannot be resolved at build time)
func TestRenderTemplate_RuntimeDynamicComponentsSkipped(t *testing.T) {
	// Create page with runtime variable path
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.DynamicComponentNode{
				PathExpression: "{path}", // Runtime variable - should be skipped
				Props:          []ast.ComponentProp{},
			},
			&ast.StyleSection{
				Content: `.page { color: red; }`,
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Should only contain page styles, not any dynamic component
	if !strings.Contains(style, ".page { color: red; }") {
		t.Errorf("Expected page styles in result")
	}

	// Should not panic and should skip the runtime path gracefully
	if strings.Contains(style, "/* Styles from: path */") {
		t.Errorf("Should not include runtime variable paths")
	}
}

// TestRenderTemplate_VariableInPathSkipped tests that paths with variables
// like {comp} are skipped
func TestRenderTemplate_VariableInPathSkipped(t *testing.T) {
	// Create page with variable in path
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.DynamicComponentNode{
				PathExpression: "./components/{comp}.html", // Variable in path
				Props:          []ast.ComponentProp{},
			},
			&ast.StyleSection{
				Content: `.page { color: blue; }`,
			},
		},
	}

	// Extract styles
	style := renderAndExtractStyles(t, pageTemplate)

	// Should only contain page styles
	if !strings.Contains(style, ".page { color: blue; }") {
		t.Errorf("Expected page styles in result")
	}

	// Should skip the variable path
	if strings.Contains(style, "/* Styles from: comp */") {
		t.Errorf("Should not include paths with runtime variables")
	}
}
