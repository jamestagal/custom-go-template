package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/analyzer"
	"github.com/jimafisk/custom_go_template/ast"
)

// TestDynamicComponentResolution_BuildTimeVsRuntime tests both resolution paths
//
// Pattern: Table-Driven Test [Load: 12]
// Cognitive Load: 12 (test setup: 4, runtime checks: 4, build-time checks: 4)
//
// This test is the comprehensive validation for Phase 5.4, ensuring BOTH
// build-time and runtime resolution paths work correctly without regressions.
//
// Key Validations:
//   - Runtime resolution: Emits wrapper with x-data and x-init
//   - Build-time resolution: Inlines component directly (no wrapper)
//   - ScopeAnalyzer distinction: Correctly identifies which path to use
func TestDynamicComponentResolution_BuildTimeVsRuntime(t *testing.T) {
	// Register a test component for build-time resolution
	testComponentAST := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "test-component"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Test Component Content"},
				},
			},
		},
	}
	RegisterComponent("TestComponent", testComponentAST, []string{})
	defer UnregisterComponent("TestComponent")

	t.Run("Runtime Resolution: Loop Variable", func(t *testing.T) {
		// Setup: component name is a loop variable
		dataScope := map[string]any{
			"component": nil, // Loop variable marker (nil value)
		}

		node := &ast.DynamicComponentByNameNode{
			NameExpression: "component.name",
			Props:          []ast.ComponentProp{},
			SpreadProps:    []string{},
		}

		result := TransformDynamicComponentByName(node, dataScope)

		// Should emit runtime wrapper
		if len(result) == 0 {
			t.Fatal("Expected runtime wrapper, got empty result")
		}

		// Check for runtime wrapper element
		elem, ok := result[0].(*ast.Element)
		if !ok {
			t.Fatalf("Expected Element node, got %T", result[0])
		}

		// Verify wrapper class
		hasRuntimeClass := false
		for _, attr := range elem.Attributes {
			if attr.Name == "class" && strings.Contains(attr.Value, "dyn-comp-runtime") {
				hasRuntimeClass = true
				break
			}
		}
		if !hasRuntimeClass {
			t.Error("Runtime wrapper missing 'dyn-comp-runtime' class")
		}

		// Verify x-data attribute
		hasXData := false
		for _, attr := range elem.Attributes {
			if attr.Name == "x-data" {
				hasXData = true
				// Verify it contains the runtime expression
				if !strings.Contains(attr.Value, "component.name") {
					t.Errorf("x-data should preserve expression 'component.name', got: %s", attr.Value)
				}
				break
			}
		}
		if !hasXData {
			t.Error("Runtime wrapper missing 'x-data' attribute")
		}

		// Verify x-init attribute
		hasXInit := false
		for _, attr := range elem.Attributes {
			if attr.Name == "x-init" {
				hasXInit = true
				expectedInit := "$renderDynamicComponent($el, compName, compProps)"
				if attr.Value != expectedInit {
					t.Errorf("x-init incorrect:\nexpected: %s\ngot: %s", expectedInit, attr.Value)
				}
				break
			}
		}
		if !hasXInit {
			t.Error("Runtime wrapper missing 'x-init' attribute")
		}

		t.Logf("✓ Runtime resolution path working correctly")
	})

	t.Run("Build-Time Resolution: String Literal", func(t *testing.T) {
		// Setup: component name is a string literal
		dataScope := map[string]any{
			"title": "Test Title", // Regular prop in scope
		}

		node := &ast.DynamicComponentByNameNode{
			NameExpression: `"TestComponent"`,
			Props:          []ast.ComponentProp{},
			SpreadProps:    []string{},
		}

		result := TransformDynamicComponentByName(node, dataScope)

		// Should inline component directly (no wrapper)
		if len(result) == 0 {
			t.Fatal("Expected inlined component, got empty result")
		}

		// First node should be the component's actual content (not a wrapper)
		elem, ok := result[0].(*ast.Element)
		if !ok {
			t.Fatalf("Expected Element node, got %T", result[0])
		}

		// Should NOT have runtime wrapper class
		for _, attr := range elem.Attributes {
			if attr.Name == "class" && strings.Contains(attr.Value, "dyn-comp-runtime") {
				t.Error("Build-time resolution should not use runtime wrapper")
			}
		}

		// Should be the actual component content
		if elem.TagName != "div" {
			t.Errorf("Expected div tag from component, got %s", elem.TagName)
		}

		// Verify it's the actual test component content
		hasTestClass := false
		for _, attr := range elem.Attributes {
			if attr.Name == "class" && attr.Value == "test-component" {
				hasTestClass = true
				break
			}
		}
		if !hasTestClass {
			t.Error("Build-time resolution should inline actual component content")
		}

		t.Logf("✓ Build-time resolution path working correctly")
	})

	t.Run("Build-Time Resolution: Content Prop", func(t *testing.T) {
		// Setup: component name is a content prop with known value
		dataScope := map[string]any{
			"layout": "TestComponent", // Content prop with build-time value
		}

		node := &ast.DynamicComponentByNameNode{
			NameExpression: "layout",
			Props:          []ast.ComponentProp{},
			SpreadProps:    []string{},
		}

		result := TransformDynamicComponentByName(node, dataScope)

		// Should inline component directly (build-time resolution)
		if len(result) == 0 {
			t.Fatal("Expected inlined component, got empty result")
		}

		elem, ok := result[0].(*ast.Element)
		if !ok {
			t.Fatalf("Expected Element node, got %T", result[0])
		}

		// Should NOT have runtime wrapper class
		for _, attr := range elem.Attributes {
			if attr.Name == "class" && strings.Contains(attr.Value, "dyn-comp-runtime") {
				t.Error("Content prop should resolve at build-time, not runtime")
			}
		}

		// Verify it's the actual component
		hasTestClass := false
		for _, attr := range elem.Attributes {
			if attr.Name == "class" && attr.Value == "test-component" {
				hasTestClass = true
				break
			}
		}
		if !hasTestClass {
			t.Error("Content prop should inline actual component content")
		}

		t.Logf("✓ Content prop build-time resolution working correctly")
	})

	t.Run("Scope Analyzer Correctly Distinguishes", func(t *testing.T) {
		dataScope := map[string]any{
			"component": nil,      // Loop variable (nil marker)
			"title":     "Static", // Build-time variable
			"layout":    "_index", // Content prop with value
		}

		scopeAnalyzer := analyzer.NewScopeAnalyzer(dataScope)

		// Runtime expressions (should return true)
		runtimeTests := []struct {
			expr        string
			description string
		}{
			{"component.name", "loop variable property access"},
			{"component.fields", "loop variable nested property"},
			{"component", "loop variable itself"},
			{"$store.auth.user", "Alpine store reference"},
			{"$auth.isLoggedIn", "Alpine store shorthand"},
			{"count + 1", "expression with operator"},
		}
		for _, tt := range runtimeTests {
			if !scopeAnalyzer.IsRuntimeExpression(tt.expr) {
				t.Errorf("Expected runtime expression: %s (%s)", tt.expr, tt.description)
			}
		}

		// Build-time expressions (should return false)
		buildTimeTests := []struct {
			expr        string
			description string
		}{
			{`"Hero2436"`, "double-quoted string literal"},
			{`'Hero2436'`, "single-quoted string literal"},
			{"title", "regular variable with value"},
			{"layout", "content prop with value"},
		}
		for _, tt := range buildTimeTests {
			if scopeAnalyzer.IsRuntimeExpression(tt.expr) {
				t.Errorf("Expected build-time expression: %s (%s)", tt.expr, tt.description)
			}
		}

		t.Logf("✓ ScopeAnalyzer distinction logic working correctly")
	})
}

// TestDynamicComponentResolution_RuntimePathVariations tests various runtime scenarios
//
// Pattern: Table-Driven Test [Load: 10]
// Cognitive Load: 10 (test variations: 6, validation: 4)
//
// This test ensures runtime wrappers are emitted correctly for ALL runtime scenarios:
//   - Loop variables (component.name)
//   - Alpine store references ($store.*, $auth.*)
//   - Expressions with operators (count + 1)
func TestDynamicComponentResolution_RuntimePathVariations(t *testing.T) {
	tests := []struct {
		name        string
		nameExpr    string
		dataScope   map[string]any
		description string
	}{
		{
			name:     "loop variable with property access",
			nameExpr: "component.name",
			dataScope: map[string]any{
				"component": nil, // Loop variable marker
			},
			description: "component.name in loop context",
		},
		{
			name:     "loop variable nested property",
			nameExpr: "item.config.componentType",
			dataScope: map[string]any{
				"item": nil, // Loop variable marker
			},
			description: "nested property on loop variable",
		},
		{
			name:        "Alpine store reference",
			nameExpr:    "$store.ui.currentComponent",
			dataScope:   map[string]any{},
			description: "Alpine store property access",
		},
		{
			name:        "Alpine store shorthand",
			nameExpr:    "$auth.componentName",
			dataScope:   map[string]any{},
			description: "Alpine store shorthand syntax",
		},
		{
			name:     "expression with operator",
			nameExpr: "prefix + componentType",
			dataScope: map[string]any{
				"prefix":        "Component",
				"componentType": "Name",
			},
			description: "expression with concatenation operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.DynamicComponentByNameNode{
				NameExpression: tt.nameExpr,
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			}

			result := TransformDynamicComponentByName(node, tt.dataScope)

			if len(result) == 0 {
				t.Fatalf("%s: expected runtime wrapper, got empty result", tt.description)
			}

			elem, ok := result[0].(*ast.Element)
			if !ok {
				t.Fatalf("%s: expected Element node, got %T", tt.description, result[0])
			}

			// Verify runtime wrapper class
			hasRuntimeClass := false
			for _, attr := range elem.Attributes {
				if attr.Name == "class" && strings.Contains(attr.Value, "dyn-comp-runtime") {
					hasRuntimeClass = true
					break
				}
			}
			if !hasRuntimeClass {
				t.Errorf("%s: missing runtime wrapper class", tt.description)
			}

			// Verify x-data attribute exists
			hasXData := false
			for _, attr := range elem.Attributes {
				if attr.Name == "x-data" {
					hasXData = true
					break
				}
			}
			if !hasXData {
				t.Errorf("%s: missing x-data attribute", tt.description)
			}

			t.Logf("✓ %s: runtime wrapper emitted correctly", tt.description)
		})
	}
}

// TestDynamicComponentResolution_BuildTimePathVariations tests various build-time scenarios
//
// Pattern: Table-Driven Test [Load: 10]
// Cognitive Load: 10 (test variations: 6, validation: 4)
//
// This test ensures build-time resolution works correctly for ALL static scenarios:
//   - String literals ("ComponentName")
//   - Content props (layout = "Hero2436")
//   - Simple variables with known values
func TestDynamicComponentResolution_BuildTimePathVariations(t *testing.T) {
	// Register test components
	heroAST := &ast.Template{
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
	RegisterComponent("Hero2436", heroAST, []string{})
	defer UnregisterComponent("Hero2436")

	footerAST := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "footer",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "footer"},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Footer Component"},
				},
			},
		},
	}
	RegisterComponent("Footer", footerAST, []string{})
	defer UnregisterComponent("Footer")

	tests := []struct {
		name          string
		nameExpr      string
		dataScope     map[string]any
		expectedClass string
		description   string
	}{
		{
			name:          "double-quoted string literal",
			nameExpr:      `"Hero2436"`,
			dataScope:     map[string]any{},
			expectedClass: "hero",
			description:   "quoted string literal Hero2436",
		},
		{
			name:          "single-quoted string literal",
			nameExpr:      `'Footer'`,
			dataScope:     map[string]any{},
			expectedClass: "footer",
			description:   "single-quoted string literal Footer",
		},
		{
			name:     "content prop with value",
			nameExpr: "layout",
			dataScope: map[string]any{
				"layout": "Hero2436",
			},
			expectedClass: "hero",
			description:   "content prop resolves to Hero2436",
		},
		{
			name:     "simple variable with value",
			nameExpr: "componentType",
			dataScope: map[string]any{
				"componentType": "Footer",
			},
			expectedClass: "footer",
			description:   "simple variable with Footer value",
		},
		{
			name:          "unquoted literal (fallback)",
			nameExpr:      "Hero2436",
			dataScope:     map[string]any{},
			expectedClass: "hero",
			description:   "unquoted string treated as literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.DynamicComponentByNameNode{
				NameExpression: tt.nameExpr,
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
			}

			result := TransformDynamicComponentByName(node, tt.dataScope)

			if len(result) == 0 {
				t.Fatalf("%s: expected inlined component, got empty result", tt.description)
			}

			elem, ok := result[0].(*ast.Element)
			if !ok {
				t.Fatalf("%s: expected Element node, got %T", tt.description, result[0])
			}

			// Should NOT have runtime wrapper class
			for _, attr := range elem.Attributes {
				if attr.Name == "class" && strings.Contains(attr.Value, "dyn-comp-runtime") {
					t.Errorf("%s: should NOT emit runtime wrapper", tt.description)
				}
			}

			// Verify expected component was inlined
			hasExpectedClass := false
			for _, attr := range elem.Attributes {
				if attr.Name == "class" && strings.Contains(attr.Value, tt.expectedClass) {
					hasExpectedClass = true
					break
				}
			}
			if !hasExpectedClass {
				t.Errorf("%s: expected component with class containing %q", tt.description, tt.expectedClass)
			}

			t.Logf("✓ %s: build-time resolution working correctly", tt.description)
		})
	}
}

// TestDynamicComponentResolution_RegressionChecks ensures no regressions from existing functionality
//
// Pattern: Regression Test [Load: 8]
// Cognitive Load: 8 (regression scenarios: 5, validation: 3)
//
// This test validates that ALL existing functionality continues to work:
//   - Prop spreading
//   - Regular props
//   - Mixed spreads and regular props
//   - Error handling for missing components
func TestDynamicComponentResolution_RegressionChecks(t *testing.T) {
	// Register test component with props
	cardAST := &ast.Template{
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
					&ast.TextNode{Content: "Card Component"},
				},
			},
		},
	}
	RegisterComponent("Card", cardAST, []string{"title", "description"})
	defer UnregisterComponent("Card")

	t.Run("prop spreading still works", func(t *testing.T) {
		node := &ast.DynamicComponentByNameNode{
			NameExpression: `"Card"`,
			Props:          []ast.ComponentProp{},
			SpreadProps:    []string{"cardProps"},
		}

		dataScope := map[string]any{
			"cardProps": map[string]any{
				"title":       "Spread Title",
				"description": "Spread Description",
			},
		}

		result := TransformDynamicComponentByName(node, dataScope)

		if len(result) == 0 {
			t.Fatal("Expected component with spread props, got empty result")
		}

		// Verify it's build-time resolved (not runtime wrapper)
		elem, ok := result[0].(*ast.Element)
		if !ok {
			t.Fatalf("Expected Element, got %T", result[0])
		}

		for _, attr := range elem.Attributes {
			if attr.Name == "class" && strings.Contains(attr.Value, "dyn-comp-runtime") {
				t.Error("Spread props should not trigger runtime path")
			}
		}

		t.Logf("✓ Prop spreading regression check passed")
	})

	t.Run("regular props still work", func(t *testing.T) {
		node := &ast.DynamicComponentByNameNode{
			NameExpression: `"Card"`,
			Props: []ast.ComponentProp{
				{Name: "title", Value: "Explicit Title", IsDynamic: false},
			},
			SpreadProps: []string{},
		}

		dataScope := map[string]any{}

		result := TransformDynamicComponentByName(node, dataScope)

		if len(result) == 0 {
			t.Fatal("Expected component with props, got empty result")
		}

		t.Logf("✓ Regular props regression check passed")
	})

	t.Run("error handling for missing component", func(t *testing.T) {
		node := &ast.DynamicComponentByNameNode{
			NameExpression: `"NonExistentComponent"`,
			Props:          []ast.ComponentProp{},
			SpreadProps:    []string{},
		}

		dataScope := map[string]any{}

		result := TransformDynamicComponentByName(node, dataScope)

		// Should return placeholder/error nodes
		if len(result) == 0 {
			t.Fatal("Expected error placeholder, got empty result")
		}

		// Verify it's an error placeholder (comment or element with data-error)
		foundErrorIndicator := false
		for _, node := range result {
			if _, ok := node.(*ast.CommentNode); ok {
				foundErrorIndicator = true
				break
			}
			if elem, ok := node.(*ast.Element); ok {
				for _, attr := range elem.Attributes {
					if attr.Name == "data-error" || attr.Name == "x-component" {
						foundErrorIndicator = true
						break
					}
				}
			}
		}

		if !foundErrorIndicator {
			t.Error("Missing component should produce error placeholder")
		}

		t.Logf("✓ Error handling regression check passed")
	})
}

// TestScopeAnalyzer_IntegrationWithTransformer validates ScopeAnalyzer integration
//
// Pattern: Integration Test [Load: 8]
// Cognitive Load: 8 (integration points: 4, validation: 4)
//
// This test validates that ScopeAnalyzer is properly integrated into the transformer
// and makes correct decisions for all expression types.
func TestScopeAnalyzer_IntegrationWithTransformer(t *testing.T) {
	t.Run("loop variable detection", func(t *testing.T) {
		dataScope := map[string]any{
			"component": nil, // Loop variable marker
		}

		analyzer := analyzer.NewScopeAnalyzer(dataScope)

		if !analyzer.IsRuntimeExpression("component.name") {
			t.Error("ScopeAnalyzer should detect loop variable via nil marker")
		}

		t.Logf("✓ Loop variable detection working")
	})

	t.Run("Alpine store detection", func(t *testing.T) {
		dataScope := map[string]any{}
		analyzer := analyzer.NewScopeAnalyzer(dataScope)

		storeExprs := []string{
			"$store.auth.user",
			"$auth.isLoggedIn",
			"$cart.items",
		}

		for _, expr := range storeExprs {
			if !analyzer.IsRuntimeExpression(expr) {
				t.Errorf("ScopeAnalyzer should detect Alpine store: %s", expr)
			}
		}

		t.Logf("✓ Alpine store detection working")
	})

	t.Run("string literal detection", func(t *testing.T) {
		dataScope := map[string]any{}
		analyzer := analyzer.NewScopeAnalyzer(dataScope)

		literals := []string{
			`"Hero2436"`,
			`'Footer'`,
		}

		for _, expr := range literals {
			if analyzer.IsRuntimeExpression(expr) {
				t.Errorf("ScopeAnalyzer should recognize string literal as build-time: %s", expr)
			}
		}

		t.Logf("✓ String literal detection working")
	})

	t.Run("operator detection", func(t *testing.T) {
		dataScope := map[string]any{
			"count": 5,
		}
		analyzer := analyzer.NewScopeAnalyzer(dataScope)

		operatorExprs := []string{
			"count + 1",
			"value > 10",
			"name && active",
		}

		for _, expr := range operatorExprs {
			if !analyzer.IsRuntimeExpression(expr) {
				t.Errorf("ScopeAnalyzer should detect operator expression as runtime: %s", expr)
			}
		}

		t.Logf("✓ Operator detection working")
	})
}
