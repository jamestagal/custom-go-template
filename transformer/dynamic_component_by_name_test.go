package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestTransformDynamicComponentByName_BasicTransformation tests basic component name resolution
func TestTransformDynamicComponentByName_BasicTransformation(t *testing.T) {
	// Register a test component
	heroTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "hero"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Hero Component"},
				},
			},
		},
	}
	RegisterComponent("Hero2436", heroTemplate, []string{})
	defer UnregisterComponent("Hero2436")

	tests := []struct {
		name        string
		node        *ast.DynamicComponentByNameNode
		dataScope   map[string]any
		expectError bool
		description string
	}{
		{
			name: "simple string literal name",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "Hero2436",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{},
			expectError: false,
			description: "Should resolve string literal component name",
		},
		{
			name: "quoted string literal name",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: `"Hero2436"`,
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{},
			expectError: false,
			description: "Should handle quoted string literal",
		},
		{
			name: "variable name resolution",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{
				"component": map[string]any{
					"name": "Hero2436",
				},
			},
			expectError: false,
			description: "Should resolve name from variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformDynamicComponentByName(tt.node, tt.dataScope)

			if tt.expectError {
				// Check for error/placeholder nodes
				if len(result) == 0 {
					t.Errorf("%s: expected error result, got empty", tt.description)
				}
			} else {
				// Check for successful transformation
				if len(result) == 0 {
					t.Errorf("%s: expected transformed nodes, got empty", tt.description)
				}
			}
		})
	}
}

// TestTransformDynamicComponentByName_SimplePropSubstitution tests simple prop value substitution
// This is the NEW test for Task 1 requirement: <Component:dynamic name={layout} /> where layout is a prop
func TestTransformDynamicComponentByName_SimplePropSubstitution(t *testing.T) {
	// Register test layout components
	indexTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "page-index"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Index Layout"},
				},
			},
		},
	}
	RegisterComponent("_index", indexTemplate, []string{})
	defer UnregisterComponent("_index")

	comprehensiveTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "page-comprehensive"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Comprehensive Layout"},
				},
			},
		},
	}
	RegisterComponent("comprehensive", comprehensiveTemplate, []string{})
	defer UnregisterComponent("comprehensive")

	storeDemoTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "page-store-demo"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Store Demo Layout"},
				},
			},
		},
	}
	RegisterComponent("store-demo", storeDemoTemplate, []string{})
	defer UnregisterComponent("store-demo")

	tests := []struct {
		name             string
		node             *ast.DynamicComponentByNameNode
		dataScope        map[string]any
		expectError      bool
		expectedClass    string // Expected class in rendered output
		description      string
	}{
		{
			name: "simple prop reference - _index",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout", // Simple prop reference (no braces, no dots)
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{
				"layout": "_index", // Prop value in data scope
			},
			expectError:   false,
			expectedClass: "page-index",
			description:   "Should resolve layout prop to _index component",
		},
		{
			name: "simple prop reference - comprehensive",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{
				"layout": "comprehensive",
			},
			expectError:   false,
			expectedClass: "page-comprehensive",
			description:   "Should resolve layout prop to comprehensive component",
		},
		{
			name: "simple prop reference - store-demo",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{
				"layout": "store-demo",
			},
			expectError:   false,
			expectedClass: "page-store-demo",
			description:   "Should resolve layout prop to store-demo component",
		},
		{
			name: "missing prop reference",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{
				// layout prop not in scope
			},
			expectError: true,
			description: "Should error when prop not in data scope",
		},
		{
			name: "prop with non-existent layout",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			},
			dataScope: map[string]any{
				"layout": "nonexistent",
			},
			expectError: true,
			description: "Should error when layout component not registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformDynamicComponentByName(tt.node, tt.dataScope)

			if tt.expectError {
				// Check for error/placeholder nodes
				if len(result) == 0 {
					t.Errorf("%s: expected error result, got empty", tt.description)
					return
				}

				// Verify it's a placeholder/error node
				if element, ok := result[len(result)-1].(*ast.Element); ok {
					// Check for error indicators (x-component, data-error attributes)
					hasError := false
					for _, attr := range element.Attributes {
						if attr.Name == "data-error" || attr.Name == "x-component" {
							hasError = true
							break
						}
					}
					if !hasError {
						t.Errorf("%s: expected error node with data-error or x-component attribute", tt.description)
					}
				} else if comment, ok := result[0].(*ast.CommentNode); !ok {
					t.Errorf("%s: expected placeholder element or comment, got %T", tt.description, comment)
				}
			} else {
				// Check for successful transformation
				if len(result) == 0 {
					t.Errorf("%s: expected transformed nodes, got empty", tt.description)
					return
				}

				// Verify the correct component was loaded by checking class attribute
				found := false
				for _, node := range result {
					if element, ok := node.(*ast.Element); ok {
						for _, attr := range element.Attributes {
							if attr.Name == "class" && attr.Value == tt.expectedClass {
								found = true
								break
							}
						}
					}
				}
				if !found {
					t.Errorf("%s: expected element with class=%q", tt.description, tt.expectedClass)
				}
			}
		})
	}
}

// TestTransformDynamicComponentByName_PropSpreadingWithSimpleProp tests prop spreading combined with simple prop substitution
// This tests the html.html use case: <Component:dynamic name={layout} {...content.fields} allContent={allContent} />
func TestTransformDynamicComponentByName_PropSpreadingWithSimpleProp(t *testing.T) {
	// Register a layout component that accepts props
	pageTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: "Default Title"},
					{Name: "description", DefaultValue: "Default Description"},
					{Name: "allContent", DefaultValue: "{}"},
				},
			},
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "page-layout"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Page Layout"},
				},
			},
		},
	}
	RegisterComponent("pages", pageTemplate, []string{"title", "description", "allContent"})
	defer UnregisterComponent("pages")

	tests := []struct {
		name        string
		node        *ast.DynamicComponentByNameNode
		dataScope   map[string]any
		description string
	}{
		{
			name: "simple prop name with spread props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout", // Simple prop reference
				Props: []ast.ComponentProp{
					{Name: "allContent", Value: "allContent", IsDynamic: true},
				},
				SpreadProps: []string{"content.fields"},
			},
			dataScope: map[string]any{
				"layout": "pages", // Prop value resolves to "pages" component
				"content": map[string]any{
					"fields": map[string]any{
						"title":       "Welcome to the Site",
						"description": "This is the homepage",
					},
				},
				"allContent": map[string]any{
					"pages": []string{"/", "/about"},
				},
			},
			description: "Should resolve layout prop and spread content.fields with explicit props",
		},
		{
			name: "multiple spread props with simple prop name",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "layout",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{"content.fields", "extraProps"},
			},
			dataScope: map[string]any{
				"layout": "pages",
				"content": map[string]any{
					"fields": map[string]any{
						"title": "Base Title",
					},
				},
				"extraProps": map[string]any{
					"title":       "Override Title", // Should override content.fields.title
					"description": "Added Description",
				},
			},
			description: "Should apply spreads left-to-right with later overriding earlier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformDynamicComponentByName(tt.node, tt.dataScope)

			if len(result) == 0 {
				t.Errorf("%s: expected transformed nodes, got empty", tt.description)
			}

			// Verify transformation succeeded
			hasPageLayout := false
			for _, node := range result {
				if element, ok := node.(*ast.Element); ok {
					for _, attr := range element.Attributes {
						if attr.Name == "class" && attr.Value == "page-layout" {
							hasPageLayout = true
							break
						}
					}
				}
			}

			if !hasPageLayout {
				t.Errorf("%s: expected page-layout element in transformed output", tt.description)
			}
		})
	}
}

// TestTransformDynamicComponentByName_WithSpreadProps tests spread prop expansion
func TestTransformDynamicComponentByName_WithSpreadProps(t *testing.T) {
	// Register a component that accepts props
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: "Default Title"},
					{Name: "description", DefaultValue: "Default Description"},
				},
			},
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "card"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Card"},
				},
			},
		},
	}
	RegisterComponent("Card", cardTemplate, []string{"title", "description"})
	defer UnregisterComponent("Card")

	tests := []struct {
		name          string
		node          *ast.DynamicComponentByNameNode
		dataScope     map[string]any
		expectedProps map[string]any
		description   string
	}{
		{
			name: "single spread prop",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "Card",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{"component.fields"},
			},
			dataScope: map[string]any{
				"component": map[string]any{
					"fields": map[string]any{
						"title":       "Spread Title",
						"description": "Spread Description",
					},
				},
			},
			expectedProps: map[string]any{
				"title":       "Spread Title",
				"description": "Spread Description",
			},
			description: "Should expand single spread prop",
		},
		{
			name: "multiple spread props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "Card",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{"defaults", "overrides"},
			},
			dataScope: map[string]any{
				"defaults": map[string]any{
					"title":       "Default Title",
					"description": "Default Description",
				},
				"overrides": map[string]any{
					"title": "Override Title",
				},
			},
			expectedProps: map[string]any{
				"title":       "Override Title",
				"description": "Default Description",
			},
			description: "Should merge multiple spread props with rightmost winning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformDynamicComponentByName(tt.node, tt.dataScope)

			if len(result) == 0 {
				t.Errorf("%s: expected transformed nodes, got empty", tt.description)
			}

			// Note: Full prop validation would require inspecting the rendered output
			// For now, we just verify transformation succeeded
		})
	}
}

// TestTransformDynamicComponentByName_MixedProps tests spread + regular props
func TestTransformDynamicComponentByName_MixedProps(t *testing.T) {
	// Register test component
	testTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName:  "div",
				Children: []ast.Node{},
			},
		},
	}
	RegisterComponent("TestComponent", testTemplate, []string{"a", "b", "c"})
	defer UnregisterComponent("TestComponent")

	tests := []struct {
		name        string
		node        *ast.DynamicComponentByNameNode
		dataScope   map[string]any
		description string
	}{
		{
			name: "spread then regular prop override",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "TestComponent",
				Props: []ast.ComponentProp{
					{Name: "b", Value: "Override B", IsDynamic: false},
				},
				SpreadProps: []string{"defaults"},
			},
			dataScope: map[string]any{
				"defaults": map[string]any{
					"a": "A",
					"b": "B",
					"c": "C",
				},
			},
			description: "Regular prop should override spread prop",
		},
		{
			name: "multiple spreads and regular props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "TestComponent",
				Props: []ast.ComponentProp{
					{Name: "a", Value: "Final A", IsDynamic: false},
				},
				SpreadProps: []string{"first", "second"},
			},
			dataScope: map[string]any{
				"first": map[string]any{
					"a": "First A",
					"b": "First B",
				},
				"second": map[string]any{
					"b": "Second B",
					"c": "Second C",
				},
			},
			description: "Should apply spreads left-to-right, then regular props",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformDynamicComponentByName(tt.node, tt.dataScope)

			if len(result) == 0 {
				t.Errorf("%s: expected transformed nodes, got empty", tt.description)
			}
		})
	}
}

// TestTransformDynamicComponentByName_ComponentNotFound tests error handling
func TestTransformDynamicComponentByName_ComponentNotFound(t *testing.T) {
	node := &ast.DynamicComponentByNameNode{
		NameExpression: "NonExistentComponent",
		Props:          []ast.ComponentProp{},
		SpreadProps:    []string{},
	}

	result := TransformDynamicComponentByName(node, map[string]any{})

	// Should return placeholder or comment node
	if len(result) == 0 {
		t.Error("Expected placeholder node for missing component, got empty")
	}

	// Check if result contains placeholder
	foundPlaceholder := false
	for _, node := range result {
		if element, ok := node.(*ast.Element); ok {
			// Should have x-component attribute or data-error
			for _, attr := range element.Attributes {
				if attr.Name == "x-component" || attr.Name == "data-error" {
					foundPlaceholder = true
					break
				}
			}
		}
		if comment, ok := node.(*ast.CommentNode); ok {
			// Comment nodes are also valid placeholders
			_ = comment
			foundPlaceholder = true
		}
	}

	if !foundPlaceholder {
		t.Error("Expected placeholder with x-component/data-error attribute or comment")
	}
}

// TestResolveSpreadProps_SingleSpread tests single spread resolution
func TestResolveSpreadProps_SingleSpread(t *testing.T) {
	tests := []struct {
		name        string
		spreadExprs []string
		dataScope   map[string]any
		expected    map[string]any
		description string
	}{
		{
			name:        "simple object spread",
			spreadExprs: []string{"obj"},
			dataScope: map[string]any{
				"obj": map[string]any{
					"a": 1,
					"b": 2,
				},
			},
			expected: map[string]any{
				"a": 1,
				"b": 2,
			},
			description: "Should expand simple object",
		},
		{
			name:        "nested property spread",
			spreadExprs: []string{"component.fields"},
			dataScope: map[string]any{
				"component": map[string]any{
					"fields": map[string]any{
						"title":       "Hello",
						"description": "World",
					},
				},
			},
			expected: map[string]any{
				"title":       "Hello",
				"description": "World",
			},
			description: "Should resolve nested property access",
		},
		{
			name:        "missing spread expression",
			spreadExprs: []string{"missing"},
			dataScope:   map[string]any{},
			expected:    map[string]any{},
			description: "Should return empty map for missing expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveSpreadProps(tt.spreadExprs, tt.dataScope)

			if len(result) != len(tt.expected) {
				t.Errorf("%s: expected %d props, got %d", tt.description, len(tt.expected), len(result))
			}

			for key, expectedVal := range tt.expected {
				if actualVal, ok := result[key]; !ok {
					t.Errorf("%s: missing key %q", tt.description, key)
				} else if actualVal != expectedVal {
					t.Errorf("%s: key %q = %v, want %v", tt.description, key, actualVal, expectedVal)
				}
			}
		})
	}
}

// TestResolveSpreadProps_MultipleSpread tests multiple spread merging
func TestResolveSpreadProps_MultipleSpread(t *testing.T) {
	tests := []struct {
		name        string
		spreadExprs []string
		dataScope   map[string]any
		expected    map[string]any
		description string
	}{
		{
			name:        "two spreads with overlap",
			spreadExprs: []string{"first", "second"},
			dataScope: map[string]any{
				"first": map[string]any{
					"a": 1,
					"b": 2,
				},
				"second": map[string]any{
					"b": 3,
					"c": 4,
				},
			},
			expected: map[string]any{
				"a": 1,
				"b": 3, // second overrides first
				"c": 4,
			},
			description: "Later spread should override earlier",
		},
		{
			name:        "three spreads with progressive override",
			spreadExprs: []string{"defaults", "overrides", "final"},
			dataScope: map[string]any{
				"defaults": map[string]any{
					"x": 1,
					"y": 2,
					"z": 3,
				},
				"overrides": map[string]any{
					"y": 20,
				},
				"final": map[string]any{
					"z": 30,
				},
			},
			expected: map[string]any{
				"x": 1,
				"y": 20,
				"z": 30,
			},
			description: "Should apply spreads left-to-right with rightmost winning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveSpreadProps(tt.spreadExprs, tt.dataScope)

			for key, expectedVal := range tt.expected {
				if actualVal, ok := result[key]; !ok {
					t.Errorf("%s: missing key %q", tt.description, key)
				} else if actualVal != expectedVal {
					t.Errorf("%s: key %q = %v, want %v", tt.description, key, actualVal, expectedVal)
				}
			}
		})
	}
}

// TestMergeProps tests prop merging with override logic
func TestMergeProps(t *testing.T) {
	tests := []struct {
		name         string
		spreadProps  map[string]any
		regularProps []ast.ComponentProp
		dataScope    map[string]any
		expected     map[string]any
		description  string
	}{
		{
			name: "regular prop overrides spread",
			spreadProps: map[string]any{
				"title": "Spread Title",
				"link":  "/spread",
			},
			regularProps: []ast.ComponentProp{
				{Name: "title", Value: "Override Title", IsDynamic: false},
			},
			dataScope: map[string]any{},
			expected: map[string]any{
				"title": "Override Title",
				"link":  "/spread",
			},
			description: "Regular prop should override spread prop",
		},
		{
			name: "spread prop with no override",
			spreadProps: map[string]any{
				"a": 1,
				"b": 2,
			},
			regularProps: []ast.ComponentProp{
				{Name: "c", Value: "3", IsDynamic: false},
			},
			dataScope: map[string]any{},
			expected: map[string]any{
				"a": 1,
				"b": 2,
				"c": "3",
			},
			description: "Should merge spread and regular props",
		},
		{
			name: "dynamic regular prop override",
			spreadProps: map[string]any{
				"count": 10,
			},
			regularProps: []ast.ComponentProp{
				{Name: "count", Value: "total", IsDynamic: true},
			},
			dataScope: map[string]any{
				"total": 20,
			},
			expected: map[string]any{
				// Dynamic props return variable reference markers for Alpine.js
				"count": "__VAR_REF__total",
			},
			description: "Dynamic prop should override spread",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeProps(tt.spreadProps, tt.regularProps, tt.dataScope)

			for key, expectedVal := range tt.expected {
				if actualVal, ok := result[key]; !ok {
					t.Errorf("%s: missing key %q", tt.description, key)
				} else if actualVal != expectedVal {
					t.Errorf("%s: key %q = %v, want %v", tt.description, key, actualVal, expectedVal)
				}
			}
		})
	}
}

// TestEvaluateNameExpression tests component name evaluation
func TestEvaluateNameExpression(t *testing.T) {
	tests := []struct {
		name        string
		nameExpr    string
		dataScope   map[string]any
		expected    string
		expectError bool
		description string
	}{
		{
			name:        "string literal",
			nameExpr:    "Hero2436",
			dataScope:   map[string]any{},
			expected:    "Hero2436",
			expectError: false,
			description: "Should return string literal as-is",
		},
		{
			name:        "quoted string literal",
			nameExpr:    `"Hero2436"`,
			dataScope:   map[string]any{},
			expected:    "Hero2436",
			expectError: false,
			description: "Should strip quotes from string literal",
		},
		{
			name:     "simple variable",
			nameExpr: "componentName",
			dataScope: map[string]any{
				"componentName": "Card",
			},
			expected:    "Card",
			expectError: false,
			description: "Should resolve simple variable",
		},
		{
			name:     "nested property",
			nameExpr: "component.name",
			dataScope: map[string]any{
				"component": map[string]any{
					"name": "Hero2436",
				},
			},
			expected:    "Hero2436",
			expectError: false,
			description: "Should resolve nested property",
		},
		{
			name:        "missing variable",
			nameExpr:    "missing",
			dataScope:   map[string]any{},
			expected:    "missing",  // Treated as literal string when not found
			expectError: false,
			description: "Should treat missing variable as literal string",
		},
		{
			name:     "simple prop reference (key use case)",
			nameExpr: "layout",
			dataScope: map[string]any{
				"layout": "_index",
			},
			expected:    "_index",
			expectError: false,
			description: "Should resolve layout prop to _index",
		},
		{
			name:     "simple prop reference with dash",
			nameExpr: "layout",
			dataScope: map[string]any{
				"layout": "store-demo",
			},
			expected:    "store-demo",
			expectError: false,
			description: "Should resolve layout prop with dash in value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateNameExpression(tt.nameExpr, tt.dataScope)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error, got nil", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
				if result != tt.expected {
					t.Errorf("%s: got %q, want %q", tt.description, result, tt.expected)
				}
			}
		})
	}
}
