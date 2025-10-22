package renderer

import (
	"strings"
	"sync"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestAggregateComponentStyles_SingleComponent tests style aggregation from a single component
func TestAggregateComponentStyles_SingleComponent(t *testing.T) {
	// Create a simple component with one style block
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".header { background-color: #f8f9fa; }",
			},
			&ast.Element{
				TagName: "header",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "header"},
				},
			},
		},
	}

	// Register the component
	transformer.RegisterComponent("Header", componentTemplate, []string{})

	// Aggregate styles
	result := AggregateComponentStyles(componentTemplate, "Header")

	// Verify result contains the style
	if !strings.Contains(result, ".header { background-color: #f8f9fa; }") {
		t.Errorf("Expected style content in result, got: %s", result)
	}

	// Verify result contains source comment
	if !strings.Contains(result, "/* Styles from: Header */") {
		t.Errorf("Expected source comment in result, got: %s", result)
	}
}

// TestAggregateComponentStyles_NestedComponents tests style aggregation with nested component imports
func TestAggregateComponentStyles_NestedComponents(t *testing.T) {
	// Create child component with style
	childTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".button { padding: 0.5rem; }",
			},
			&ast.Element{
				TagName: "button",
			},
		},
	}

	// Create parent component that imports child
	parentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Button", Path: "./components/Button.html"},
				},
			},
			&ast.StyleSection{
				Content: ".card { border: 1px solid #ddd; }",
			},
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "card"},
				},
			},
		},
	}

	// Register components
	transformer.RegisterComponent("Button", childTemplate, []string{})
	transformer.RegisterComponent("Card", parentTemplate, []string{})

	// Aggregate styles from parent
	result := AggregateComponentStyles(parentTemplate, "Card")

	// Verify child styles come BEFORE parent styles
	buttonIndex := strings.Index(result, ".button")
	cardIndex := strings.Index(result, ".card")

	if buttonIndex == -1 {
		t.Errorf("Expected Button styles in result")
	}
	if cardIndex == -1 {
		t.Errorf("Expected Card styles in result")
	}
	if buttonIndex > cardIndex {
		t.Errorf("Expected Button styles before Card styles, got Button at %d, Card at %d", buttonIndex, cardIndex)
	}

	// Verify source comments
	if !strings.Contains(result, "/* Styles from: Button */") {
		t.Errorf("Expected Button source comment")
	}
	if !strings.Contains(result, "/* Styles from: Card */") {
		t.Errorf("Expected Card source comment")
	}
}

// TestAggregateComponentStyles_MultipleStyleBlocks tests component with multiple <style> sections
func TestAggregateComponentStyles_MultipleStyleBlocks(t *testing.T) {
	// Create component with two style blocks
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".header { color: red; }",
			},
			&ast.Element{
				TagName: "header",
			},
			&ast.StyleSection{
				Content: ".footer { color: blue; }",
			},
		},
	}

	transformer.RegisterComponent("Layout", componentTemplate, []string{})

	result := AggregateComponentStyles(componentTemplate, "Layout")

	// Verify both style blocks are included
	if !strings.Contains(result, ".header { color: red; }") {
		t.Errorf("Expected first style block in result")
	}
	if !strings.Contains(result, ".footer { color: blue; }") {
		t.Errorf("Expected second style block in result")
	}

	// Count occurrences of source comment (should be 2, one for each block)
	commentCount := strings.Count(result, "/* Styles from: Layout */")
	if commentCount != 2 {
		t.Errorf("Expected 2 source comments (one per style block), got %d", commentCount)
	}
}

// TestAggregateComponentStyles_Deduplication tests that identical style blocks are deduplicated
func TestAggregateComponentStyles_Deduplication(t *testing.T) {
	// Create two components with identical styles
	sharedStyle := ".common { margin: 1rem; }"

	comp1Template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: sharedStyle,
			},
		},
	}

	comp2Template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Component1", Path: "./Component1.html"},
				},
			},
			&ast.StyleSection{
				Content: sharedStyle, // Same style as Component1
			},
		},
	}

	transformer.RegisterComponent("Component1", comp1Template, []string{})
	transformer.RegisterComponent("Component2", comp2Template, []string{})

	result := AggregateComponentStyles(comp2Template, "Component2")

	// Count occurrences of the shared style
	styleCount := strings.Count(result, sharedStyle)
	if styleCount != 1 {
		t.Errorf("Expected style to appear once (deduplicated), but appeared %d times", styleCount)
	}

	// Should only have one source comment for the deduplicated style
	// (from the first component that contributed it)
	commentCount := strings.Count(result, "/* Styles from:")
	if commentCount != 1 {
		t.Errorf("Expected 1 source comment for deduplicated style, got %d", commentCount)
	}
}

// TestAggregateComponentStyles_CircularDependency tests that circular dependencies don't cause infinite loops
func TestAggregateComponentStyles_CircularDependency(t *testing.T) {
	// Create Component A that imports Component B
	compATemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "ComponentB", Path: "./ComponentB.html"},
				},
			},
			&ast.StyleSection{
				Content: ".comp-a { color: red; }",
			},
		},
	}

	// Create Component B that imports Component A (circular)
	compBTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "ComponentA", Path: "./ComponentA.html"},
				},
			},
			&ast.StyleSection{
				Content: ".comp-b { color: blue; }",
			},
		},
	}

	transformer.RegisterComponent("ComponentA", compATemplate, []string{})
	transformer.RegisterComponent("ComponentB", compBTemplate, []string{})

	// This should NOT cause infinite loop
	result := AggregateComponentStyles(compATemplate, "ComponentA")

	// Verify both styles are present
	if !strings.Contains(result, ".comp-a") {
		t.Errorf("Expected ComponentA styles in result")
	}
	if !strings.Contains(result, ".comp-b") {
		t.Errorf("Expected ComponentB styles in result")
	}

	// Each style should appear exactly once
	if strings.Count(result, ".comp-a") > 1 {
		t.Errorf("ComponentA styles should appear only once")
	}
	if strings.Count(result, ".comp-b") > 1 {
		t.Errorf("ComponentB styles should appear only once")
	}
}

// TestAggregateComponentStyles_EmptyStyles tests handling of empty style blocks
func TestAggregateComponentStyles_EmptyStyles(t *testing.T) {
	// Create component with empty and whitespace-only style blocks
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: "", // Empty
			},
			&ast.StyleSection{
				Content: "   \n\t  ", // Whitespace only
			},
			&ast.StyleSection{
				Content: ".valid { color: red; }", // Valid style
			},
		},
	}

	transformer.RegisterComponent("TestComp", componentTemplate, []string{})

	result := AggregateComponentStyles(componentTemplate, "TestComp")

	// Result should only contain the valid style
	if !strings.Contains(result, ".valid { color: red; }") {
		t.Errorf("Expected valid style in result")
	}

	// Should only have ONE source comment (for the valid style)
	commentCount := strings.Count(result, "/* Styles from:")
	if commentCount != 1 {
		t.Errorf("Expected 1 source comment (empty styles skipped), got %d", commentCount)
	}
}

// TestAggregateComponentStyles_NoStyles tests component without any style blocks
func TestAggregateComponentStyles_NoStyles(t *testing.T) {
	// Create component with no styles
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
			},
		},
	}

	transformer.RegisterComponent("NoStyles", componentTemplate, []string{})

	result := AggregateComponentStyles(componentTemplate, "NoStyles")

	// Result should be empty or whitespace only
	if strings.TrimSpace(result) != "" {
		t.Errorf("Expected empty result for component without styles, got: %q", result)
	}
}

// TestAggregateComponentStyles_RealWorldHeaderSimple tests the actual HeaderSimple use case
func TestAggregateComponentStyles_RealWorldHeaderSimple(t *testing.T) {
	// Create a realistic HeaderSimple component
	headerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{},
			&ast.StyleSection{
				Content: `
  .header {
    background-color: #f8f9fa;
    padding: 1rem 0;
    border-bottom: 1px solid #e9ecef;
  }

  .brand {
    display: flex;
    align-items: center;
  }

  .brand svg {
    height: 32px;
    width: auto;
  }

  .nav-item {
    color: #495057;
    text-decoration: none;
  }

  .nav-item:hover {
    color: #228be6;
  }
`,
			},
			&ast.Element{
				TagName: "header",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "header"},
				},
			},
		},
	}

	transformer.RegisterComponent("HeaderSimple", headerTemplate, []string{})

	result := AggregateComponentStyles(headerTemplate, "HeaderSimple")

	// Verify key CSS classes are present
	expectedClasses := []string{".header", ".brand", ".nav-item"}
	for _, class := range expectedClasses {
		if !strings.Contains(result, class) {
			t.Errorf("Expected style to contain %s", class)
		}
	}

	// Verify source comment
	if !strings.Contains(result, "/* Styles from: HeaderSimple */") {
		t.Errorf("Expected HeaderSimple source comment")
	}

	// Verify it's a substantial style block
	if len(result) < 100 {
		t.Errorf("Expected substantial style output, got only %d characters", len(result))
	}
}

// TestAggregateComponentStyles_NilTemplate tests handling of nil template
func TestAggregateComponentStyles_NilTemplate(t *testing.T) {
	// This should not panic
	result := AggregateComponentStyles(nil, "NilComponent")

	// Result should be empty
	if strings.TrimSpace(result) != "" {
		t.Errorf("Expected empty result for nil template, got: %q", result)
	}
}

// TestAggregateComponentStyles_NilRootNodes tests handling of template with nil RootNodes
func TestAggregateComponentStyles_NilRootNodes(t *testing.T) {
	template := &ast.Template{
		RootNodes: nil,
	}

	result := AggregateComponentStyles(template, "EmptyTemplate")

	// Result should be empty
	if strings.TrimSpace(result) != "" {
		t.Errorf("Expected empty result for template with nil RootNodes, got: %q", result)
	}
}

// TestAggregateComponentStyles_ThreeLevelNesting tests deeply nested component imports
func TestAggregateComponentStyles_ThreeLevelNesting(t *testing.T) {
	// Level 1: Icon (deepest dependency)
	iconTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".icon { width: 16px; height: 16px; }",
			},
		},
	}

	// Level 2: Button imports Icon
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Icon", Path: "./Icon.html"},
				},
			},
			&ast.StyleSection{
				Content: ".button { padding: 0.5rem; }",
			},
		},
	}

	// Level 3: Card imports Button
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Button", Path: "./Button.html"},
				},
			},
			&ast.StyleSection{
				Content: ".card { border: 1px solid #ddd; }",
			},
		},
	}

	transformer.RegisterComponent("Icon", iconTemplate, []string{})
	transformer.RegisterComponent("Button", buttonTemplate, []string{})
	transformer.RegisterComponent("Card", cardTemplate, []string{})

	result := AggregateComponentStyles(cardTemplate, "Card")

	// Verify all three styles are present
	if !strings.Contains(result, ".icon") {
		t.Errorf("Expected Icon styles")
	}
	if !strings.Contains(result, ".button") {
		t.Errorf("Expected Button styles")
	}
	if !strings.Contains(result, ".card") {
		t.Errorf("Expected Card styles")
	}

	// Verify dependency order (Icon -> Button -> Card)
	iconIndex := strings.Index(result, ".icon")
	buttonIndex := strings.Index(result, ".button")
	cardIndex := strings.Index(result, ".card")

	if iconIndex > buttonIndex || buttonIndex > cardIndex {
		t.Errorf("Expected dependency order: Icon < Button < Card, got Icon=%d, Button=%d, Card=%d",
			iconIndex, buttonIndex, cardIndex)
	}
}

// TestAggregateComponentStyles_MissingImportedComponent tests graceful handling when imported component not registered
func TestAggregateComponentStyles_MissingImportedComponent(t *testing.T) {
	// Create component that imports a non-existent component
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "NonExistent", Path: "./NonExistent.html"},
				},
			},
			&ast.StyleSection{
				Content: ".parent { color: blue; }",
			},
		},
	}

	transformer.RegisterComponent("Parent", template, []string{})

	// This should not panic
	result := AggregateComponentStyles(template, "Parent")

	// Should still include parent styles
	if !strings.Contains(result, ".parent { color: blue; }") {
		t.Errorf("Expected parent styles even when import is missing")
	}
}

// ============================================================================
// DYNAMIC COMPONENT TESTS (Bug Fix)
// ============================================================================

// TestAggregateComponentStyles_DynamicComponentNode tests style aggregation for DynamicComponentNode
func TestAggregateComponentStyles_DynamicComponentNode(t *testing.T) {
	// Create UserProfile component with styles
	userProfileTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".profile-card { border-radius: 0.5rem; }",
			},
		},
	}
	transformer.RegisterComponent("UserProfile", userProfileTemplate, []string{})

	// Create page template with DynamicComponentNode (static path)
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.DynamicComponentNode{
				PathExpression: "./components/UserProfile.html",
				Props:          []ast.ComponentProp{},
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Verify UserProfile styles included
	if !strings.Contains(result, "/* Styles from: UserProfile */") {
		t.Errorf("Expected UserProfile source comment in result")
	}
	if !strings.Contains(result, ".profile-card") {
		t.Errorf("Expected UserProfile styles in result, got: %s", result)
	}
}

// TestAggregateComponentStyles_DynamicComponentNode_RuntimePath tests skipping dynamic runtime paths
func TestAggregateComponentStyles_DynamicComponentNode_RuntimePath(t *testing.T) {
	// Create page template with DynamicComponentNode (runtime variable path)
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.DynamicComponentNode{
				PathExpression: "{path}", // Runtime variable - should be skipped
				Props:          []ast.ComponentProp{},
			},
			&ast.StyleSection{
				Content: ".page { color: red; }",
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Should only contain page styles, not any dynamic component
	if !strings.Contains(result, ".page { color: red; }") {
		t.Errorf("Expected page styles in result")
	}

	// Should not panic and should skip the runtime path gracefully
	if strings.Contains(result, "/* Styles from: path */") {
		t.Errorf("Should not include runtime variable paths")
	}
}

// TestAggregateComponentStyles_DynamicComponentNode_VariableInPath tests paths with variables in them
func TestAggregateComponentStyles_DynamicComponentNode_VariableInPath(t *testing.T) {
	// Create page template with DynamicComponentNode (path with variable)
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.DynamicComponentNode{
				PathExpression: "./components/{comp}.html", // Variable in path - should be skipped
				Props:          []ast.ComponentProp{},
			},
			&ast.StyleSection{
				Content: ".page { color: blue; }",
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Should only contain page styles
	if !strings.Contains(result, ".page { color: blue; }") {
		t.Errorf("Expected page styles in result")
	}

	// Should skip the variable path
	if strings.Contains(result, "/* Styles from: comp */") || strings.Contains(result, "/* Styles from: {comp} */") {
		t.Errorf("Should not include paths with runtime variables")
	}
}

// TestAggregateComponentStyles_RegularComponentNode tests ComponentNode style aggregation
func TestAggregateComponentStyles_RegularComponentNode(t *testing.T) {
	// Create Button component with styles
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".btn { padding: 0.5rem 1rem; }",
			},
		},
	}
	transformer.RegisterComponent("Button", buttonTemplate, []string{})

	// Create page template with ComponentNode
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Children: []ast.Node{
					&ast.ComponentNode{
						Name:  "Button",
						Props: []ast.ComponentProp{},
					},
				},
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Verify Button styles included
	if !strings.Contains(result, "/* Styles from: Button */") {
		t.Errorf("Expected Button source comment in result")
	}
	if !strings.Contains(result, ".btn") {
		t.Errorf("Expected Button styles in result, got: %s", result)
	}
}

// TestAggregateComponentStyles_ComponentsInConditionals tests components inside conditionals
func TestAggregateComponentStyles_ComponentsInConditionals(t *testing.T) {
	// Create Alert component with styles
	alertTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".alert { border: 1px solid; }",
			},
		},
	}
	transformer.RegisterComponent("Alert", alertTemplate, []string{})

	// Create page template with component in conditional
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Conditional{
				IfCondition: "showAlert",
				IfContent: []ast.Node{
					&ast.ComponentNode{
						Name:  "Alert",
						Props: []ast.ComponentProp{},
					},
				},
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Verify Alert styles included
	if !strings.Contains(result, "/* Styles from: Alert */") {
		t.Errorf("Expected Alert source comment in result")
	}
	if !strings.Contains(result, ".alert") {
		t.Errorf("Expected Alert styles in result")
	}
}

// TestAggregateComponentStyles_ComponentsInLoops tests components inside loops
func TestAggregateComponentStyles_ComponentsInLoops(t *testing.T) {
	// Create Card component with styles
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".card { box-shadow: 0 2px 4px rgba(0,0,0,0.1); }",
			},
		},
	}
	transformer.RegisterComponent("Card", cardTemplate, []string{})

	// Create page template with component in loop
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Loop{
				Iterator:   "item",
				Collection: "items",
				Content: []ast.Node{
					&ast.ComponentNode{
						Name:  "Card",
						Props: []ast.ComponentProp{},
					},
				},
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Verify Card styles included
	if !strings.Contains(result, "/* Styles from: Card */") {
		t.Errorf("Expected Card source comment in result")
	}
	if !strings.Contains(result, ".card") {
		t.Errorf("Expected Card styles in result")
	}
}

// TestAggregateComponentStyles_MixedComponentUsage tests mix of imports and inline usage
func TestAggregateComponentStyles_MixedComponentUsage(t *testing.T) {
	// Create Header component
	headerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".header { background: white; }",
			},
		},
	}
	transformer.RegisterComponent("Header", headerTemplate, []string{})

	// Create Footer component
	footerTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".footer { background: black; }",
			},
		},
	}
	transformer.RegisterComponent("Footer", footerTemplate, []string{})

	// Create page with Header imported but Footer used inline
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Header", Path: "./Header.html"},
				},
			},
			&ast.Element{
				TagName: "body",
				Children: []ast.Node{
					&ast.ComponentNode{
						Name: "Footer",
					},
				},
			},
			&ast.StyleSection{
				Content: ".page { margin: 0; }",
			},
		},
	}

	// Aggregate
	result := AggregateComponentStyles(pageTemplate, "home")

	// Verify both Header and Footer styles included
	if !strings.Contains(result, ".header") {
		t.Errorf("Expected Header styles (imported) in result")
	}
	if !strings.Contains(result, ".footer") {
		t.Errorf("Expected Footer styles (inline usage) in result")
	}
	if !strings.Contains(result, ".page") {
		t.Errorf("Expected page styles in result")
	}
}

// TestExtractComponentNameFromPath tests the helper function
func TestExtractComponentNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Simple path",
			path:     "./components/UserProfile.html",
			expected: "UserProfile",
		},
		{
			name:     "Path with quotes",
			path:     "\"./components/Button.html\"",
			expected: "Button",
		},
		{
			name:     "Path with single quotes",
			path:     "'./components/Card.html'",
			expected: "Card",
		},
		{
			name:     "Runtime variable - should skip",
			path:     "{path}",
			expected: "",
		},
		{
			name:     "Variable in path - should skip",
			path:     "./components/{comp}.html",
			expected: "",
		},
		{
			name:     "Path without extension",
			path:     "./components/Header",
			expected: "Header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentNameFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("extractComponentNameFromPath(%q) = %q, expected %q",
					tt.path, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// CACHE TESTS (Task 4)
// ============================================================================

// TestGetAggregatedStyles_Cache_MissOnFirstCall tests that first call performs full aggregation
func TestGetAggregatedStyles_Cache_MissOnFirstCall(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create a component with styles
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".cache-test { color: blue; }",
			},
		},
	}

	transformer.RegisterComponent("CacheTestComponent", template, []string{})

	// First call should be a cache miss and perform full aggregation
	result := GetAggregatedStyles(template, nil, "CacheTestComponent", "")

	// Verify result contains expected style
	if !strings.Contains(result, ".cache-test { color: blue; }") {
		t.Errorf("Expected style content in result, got: %s", result)
	}

	// Verify source comment exists
	if !strings.Contains(result, "/* Styles from: CacheTestComponent */") {
		t.Errorf("Expected source comment in result")
	}
}

// TestGetAggregatedStyles_Cache_HitOnSecondCall tests that second call returns cached result
func TestGetAggregatedStyles_Cache_HitOnSecondCall(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create a component
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".cache-hit-test { margin: 1rem; }",
			},
		},
	}

	transformer.RegisterComponent("CacheHitComponent", template, []string{})

	// First call - cache miss
	firstResult := GetAggregatedStyles(template, nil, "CacheHitComponent", "")

	// Second call - should be cache hit
	secondResult := GetAggregatedStyles(template, nil, "CacheHitComponent", "")

	// Both results should be identical
	if firstResult != secondResult {
		t.Errorf("Expected identical results from cache, got different results:\nFirst: %s\nSecond: %s",
			firstResult, secondResult)
	}

	// Verify content is correct
	if !strings.Contains(secondResult, ".cache-hit-test { margin: 1rem; }") {
		t.Errorf("Expected cached result to contain style content")
	}
}

// TestGetAggregatedStyles_Cache_DifferentComponents tests that different components cache separately
func TestGetAggregatedStyles_Cache_DifferentComponents(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create two different components
	templateA := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".component-a { color: red; }",
			},
		},
	}

	templateB := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".component-b { color: green; }",
			},
		},
	}

	transformer.RegisterComponent("ComponentA", templateA, []string{})
	transformer.RegisterComponent("ComponentB", templateB, []string{})

	// Call for both components
	resultA := GetAggregatedStyles(templateA, nil, "ComponentA", "")
	resultB := GetAggregatedStyles(templateB, nil, "ComponentB", "")

	// Verify each component has its own cached result
	if !strings.Contains(resultA, ".component-a { color: red; }") {
		t.Errorf("Expected ComponentA styles in resultA")
	}
	if strings.Contains(resultA, ".component-b") {
		t.Errorf("ComponentA result should not contain ComponentB styles")
	}

	if !strings.Contains(resultB, ".component-b { color: green; }") {
		t.Errorf("Expected ComponentB styles in resultB")
	}
	if strings.Contains(resultB, ".component-a") {
		t.Errorf("ComponentB result should not contain ComponentA styles")
	}

	// Call again to verify cache works for both
	resultA2 := GetAggregatedStyles(templateA, nil, "ComponentA", "")
	resultB2 := GetAggregatedStyles(templateB, nil, "ComponentB", "")

	if resultA != resultA2 {
		t.Errorf("ComponentA cache not working correctly")
	}
	if resultB != resultB2 {
		t.Errorf("ComponentB cache not working correctly")
	}
}

// TestClearStyleCache tests that cache clearing forces re-aggregation
func TestClearStyleCache(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create a component
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".clear-cache-test { padding: 2rem; }",
			},
		},
	}

	transformer.RegisterComponent("ClearCacheComponent", template, []string{})

	// First call - cache miss
	firstResult := GetAggregatedStyles(template, nil, "ClearCacheComponent", "")

	// Verify cache hit works
	cachedResult := GetAggregatedStyles(template, nil, "ClearCacheComponent", "")
	if firstResult != cachedResult {
		t.Errorf("Cache not working before clear")
	}

	// Clear the cache
	ClearStyleCache()

	// Next call should re-aggregate (cache miss again)
	afterClearResult := GetAggregatedStyles(template, nil, "ClearCacheComponent", "")

	// Result should be identical (same aggregation logic) but freshly computed
	if firstResult != afterClearResult {
		t.Errorf("Expected same result after cache clear, content differs:\nFirst: %s\nAfter Clear: %s",
			firstResult, afterClearResult)
	}

	// Verify content is still correct
	if !strings.Contains(afterClearResult, ".clear-cache-test { padding: 2rem; }") {
		t.Errorf("Expected style content after cache clear")
	}
}

// TestGetAggregatedStyles_ConcurrentAccess tests thread safety of cache with concurrent access
func TestGetAggregatedStyles_ConcurrentAccess(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create a component
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".concurrent-test { border: 1px solid #ddd; }",
			},
		},
	}

	transformer.RegisterComponent("ConcurrentComponent", template, []string{})

	// Run multiple goroutines concurrently
	const goroutineCount = 50
	var wg sync.WaitGroup
	results := make([]string, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = GetAggregatedStyles(template, nil, "ConcurrentComponent", "")
		}(i)
	}

	wg.Wait()

	// Verify all results are identical
	firstResult := results[0]
	for i, result := range results {
		if result != firstResult {
			t.Errorf("Concurrent access produced different results at index %d:\nExpected: %s\nGot: %s",
				i, firstResult, result)
		}
	}

	// Verify content is correct
	if !strings.Contains(firstResult, ".concurrent-test { border: 1px solid #ddd; }") {
		t.Errorf("Expected style content in concurrent access results")
	}
}

// TestGetAggregatedStyles_Cache_NilTemplate tests cache behavior with nil template
func TestGetAggregatedStyles_Cache_NilTemplate(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Call with nil template
	result := GetAggregatedStyles(nil, nil, "NilComponent", "")

	// Should return empty string
	if strings.TrimSpace(result) != "" {
		t.Errorf("Expected empty result for nil template, got: %q", result)
	}

	// Second call should also return empty (should not panic)
	result2 := GetAggregatedStyles(nil, nil, "NilComponent", "")
	if strings.TrimSpace(result2) != "" {
		t.Errorf("Expected empty result on second call for nil template")
	}
}

// TestGetAggregatedStyles_Cache_EmptyComponentName tests cache with empty component name
func TestGetAggregatedStyles_Cache_EmptyComponentName(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create a template
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".empty-name { color: purple; }",
			},
		},
	}

	// Call with empty component name
	result := GetAggregatedStyles(template, nil, "", "")

	// Should still work (empty string is valid map key)
	if !strings.Contains(result, ".empty-name { color: purple; }") {
		t.Errorf("Expected style content even with empty component name")
	}

	// Second call should be cached
	result2 := GetAggregatedStyles(template, nil, "", "")
	if result != result2 {
		t.Errorf("Cache should work with empty component name")
	}
}

// TestGetAggregatedStyles_Cache_NestedComponents tests caching with nested component imports
func TestGetAggregatedStyles_Cache_NestedComponents(t *testing.T) {
	// Clear cache before test
	ClearStyleCache()

	// Create child component
	childTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".child-cached { font-size: 14px; }",
			},
		},
	}

	// Create parent component that imports child
	parentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "CachedChild", Path: "./CachedChild.html"},
				},
			},
			&ast.StyleSection{
				Content: ".parent-cached { font-size: 16px; }",
			},
		},
	}

	transformer.RegisterComponent("CachedChild", childTemplate, []string{})
	transformer.RegisterComponent("CachedParent", parentTemplate, []string{})

	// First call - cache miss, should aggregate both parent and child
	firstResult := GetAggregatedStyles(parentTemplate, nil, "CachedParent", "")

	// Verify both styles are present
	if !strings.Contains(firstResult, ".child-cached") {
		t.Errorf("Expected child styles in first result")
	}
	if !strings.Contains(firstResult, ".parent-cached") {
		t.Errorf("Expected parent styles in first result")
	}

	// Second call - cache hit
	secondResult := GetAggregatedStyles(parentTemplate, nil, "CachedParent", "")

	// Should be identical
	if firstResult != secondResult {
		t.Errorf("Expected cache to work with nested components")
	}

	// Verify child appears before parent
	childIndex := strings.Index(secondResult, ".child-cached")
	parentIndex := strings.Index(secondResult, ".parent-cached")
	if childIndex > parentIndex {
		t.Errorf("Expected child styles before parent styles in cached result")
	}
}

// ============================================================================
// BENCHMARK TESTS (Task 4)
// ============================================================================

// createBenchmarkTemplate creates a template for benchmarking
func createBenchmarkTemplate() *ast.Template {
	return &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: `
.benchmark-header {
	background-color: #f8f9fa;
	padding: 1rem 0;
	border-bottom: 1px solid #e9ecef;
}

.benchmark-content {
	max-width: 1200px;
	margin: 0 auto;
	padding: 2rem;
}

.benchmark-footer {
	background-color: #343a40;
	color: white;
	padding: 2rem 0;
}
`,
			},
		},
	}
}

// BenchmarkGetAggregatedStyles_CacheHit benchmarks cache hit performance
func BenchmarkGetAggregatedStyles_CacheHit(b *testing.B) {
	// Setup: Create component and prime cache
	template := createBenchmarkTemplate()
	transformer.RegisterComponent("BenchmarkComponent", template, []string{})
	GetAggregatedStyles(template, nil, "BenchmarkComponent", "") // Prime cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAggregatedStyles(template, nil, "BenchmarkComponent", "")
	}
}

// BenchmarkGetAggregatedStyles_CacheMiss benchmarks cache miss performance
func BenchmarkGetAggregatedStyles_CacheMiss(b *testing.B) {
	template := createBenchmarkTemplate()
	transformer.RegisterComponent("BenchmarkMissComponent", template, []string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClearStyleCache() // Force miss
		GetAggregatedStyles(template, nil, "BenchmarkMissComponent", "")
	}
}

// BenchmarkAggregateComponentStyles_NoCache benchmarks uncached aggregation
func BenchmarkAggregateComponentStyles_NoCache(b *testing.B) {
	template := createBenchmarkTemplate()
	transformer.RegisterComponent("BenchmarkNoCacheComponent", template, []string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AggregateComponentStyles(template, "BenchmarkNoCacheComponent")
	}
}

// BenchmarkGetAggregatedStyles_NestedComponents_CacheHit benchmarks nested component cache hit
func BenchmarkGetAggregatedStyles_NestedComponents_CacheHit(b *testing.B) {
	// Create nested component structure
	childTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.StyleSection{
				Content: ".bench-child { padding: 0.5rem; }",
			},
		},
	}

	parentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "BenchChild", Path: "./BenchChild.html"},
				},
			},
			&ast.StyleSection{
				Content: ".bench-parent { margin: 1rem; }",
			},
		},
	}

	transformer.RegisterComponent("BenchChild", childTemplate, []string{})
	transformer.RegisterComponent("BenchParent", parentTemplate, []string{})

	// Prime cache
	GetAggregatedStyles(parentTemplate, nil, "BenchParent", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAggregatedStyles(parentTemplate, nil, "BenchParent", "")
	}
}
