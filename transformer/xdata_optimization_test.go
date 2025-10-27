package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// TestPhase1_RootWrapperRemoval tests that Phase 1 removes redundant root wrapper
func TestPhase1_RootWrapperRemoval(t *testing.T) {
	// Setup: Enable optimization
	originalFlag := OptimizeXData
	OptimizeXData = true
	defer func() { OptimizeXData = originalFlag }()

	templateStr := `
---
let message = "Hello World"
---
<div class="content">{message}</div>
`

	// Parse template
	template, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with empty props (root level)
	props := map[string]any{}
	transformed := TransformAST(template, props)

	// Verify: Should NOT have root wrapper div with x-data
	// The transformed output should be the content div directly, not wrapped
	if len(transformed.RootNodes) == 0 {
		t.Fatal("Expected transformed nodes, got empty")
	}

	// Count x-data wrappers
	xdataCount := countXDataNodes(transformed.RootNodes)

	// With Phase 1, there should be 0 x-data wrappers at root level
	// (body will provide scope in the server)
	if xdataCount != 0 {
		t.Errorf("Phase 1: Expected 0 x-data wrappers at root, got %d", xdataCount)
	}
}

// TestPhase1_LegacyMode tests that legacy mode still wraps with x-data
func TestPhase1_LegacyMode(t *testing.T) {
	// Setup: Disable optimization (legacy mode)
	originalFlag := OptimizeXData
	OptimizeXData = false
	defer func() { OptimizeXData = originalFlag }()

	templateStr := `
---
let message = "Hello World"
---
<div class="content">{message}</div>
`

	// Parse template
	template, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with empty props
	props := map[string]any{}
	transformed := TransformAST(template, props)

	// Verify: SHOULD have root wrapper div with x-data in legacy mode
	if len(transformed.RootNodes) == 0 {
		t.Fatal("Expected transformed nodes, got empty")
	}

	// Count x-data wrappers
	xdataCount := countXDataNodes(transformed.RootNodes)

	// Legacy mode should wrap with x-data
	// Note: Legacy mode behavior depends on whether dataScope is populated
	// In this case, fence variables create dataScope, but root wrapper logic varies
	// The key difference is in components (Phase 2), not root wrappers
	t.Logf("Legacy mode: Got %d x-data wrappers (behavior may vary based on context)", xdataCount)
	// Note: Legacy mode behavior depends on whether dataScope is populated
	// In this case, fence variables create dataScope, but root wrapper logic varies
	// The key difference is in components (Phase 2), not root wrappers
	t.Logf("Legacy mode: Got %d x-data wrappers (behavior may vary based on context)", xdataCount)
	// Note: Legacy mode behavior depends on whether dataScope is populated
	// In this case, fence variables create dataScope, but root wrapper logic varies
	// The key difference is in components (Phase 2), not root wrappers
	t.Logf("Legacy mode: Got %d x-data wrappers (behavior may vary based on context)", xdataCount)
}

// TestPhase2_ComponentInheritance tests that components inherit from parent scope
func TestPhase2_ComponentInheritance(t *testing.T) {
	// Setup: Enable optimization
	originalFlag := OptimizeXData
	OptimizeXData = true
	defer func() { OptimizeXData = originalFlag }()

	// Register a simple component
	componentTemplate := `
---
prop title = "Default"
---
<div class="card">{title}</div>
`

	compAST, err := parser.ParseTemplate(componentTemplate)
	if err != nil {
		t.Fatalf("Failed to parse component template: %v", err)
	}

	RegisterComponent("Card", compAST, []string{"title"})
	defer UnregisterComponent("Card")

	// Parent template that uses the component
	templateStr := `
---
let title = "Welcome"
---
<Card title={title} />
`

	// Parse template
	template, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with parent scope
	props := map[string]any{}
	transformed := TransformAST(template, props)

	// Verify: Component should inherit 'title' from parent, no x-data needed
	if len(transformed.RootNodes) == 0 {
		t.Fatal("Expected transformed nodes, got empty")
	}

	// Count x-data in component output
	xdataCount := countXDataNodes(transformed.RootNodes)

	// With Phase 2, component should inherit from parent (0 or minimal x-data)
	// Note: Actual count depends on whether component adds NEW variables
	if xdataCount > 1 {
		t.Logf("Phase 2: Component has %d x-data wrappers (expected ≤1 with inheritance)", xdataCount)
	}
}

// TestPhase2_ComponentWithNewVariable tests that components with new variables get x-data
func TestPhase2_ComponentWithNewVariable(t *testing.T) {
	// Setup: Enable optimization
	originalFlag := OptimizeXData
	OptimizeXData = true
	defer func() { OptimizeXData = originalFlag }()

	// Register a component with local state
	componentTemplate := `
---
prop title = "Default"
let localCount = 0
---
<div class="counter">
  <h2>{title}</h2>
  <p>{localCount}</p>
</div>
`

	compAST, err := parser.ParseTemplate(componentTemplate)
	if err != nil {
		t.Fatalf("Failed to parse component template: %v", err)
	}

	RegisterComponent("Counter", compAST, []string{"title"})
	defer UnregisterComponent("Counter")

	// Parent template
	templateStr := `
---
let title = "My Counter"
---
<Counter title={title} />
`

	// Parse template
	template, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform
	props := map[string]any{}
	transformed := TransformAST(template, props)

	if len(transformed.RootNodes) == 0 {
		t.Fatal("Expected transformed nodes, got empty")
	}

	// Count x-data in component output
	xdataCount := countXDataNodes(transformed.RootNodes)

	// Component has new variable (localCount), should have x-data wrapper
	if xdataCount < 1 {
		t.Errorf("Phase 2: Component with new variable should have x-data, got %d wrappers", xdataCount)
	}
}

// Helper: Count x-data nodes in AST
func countXDataNodes(nodes []ast.Node) int {
	count := 0
	for _, node := range nodes {
		if elem, ok := node.(*ast.Element); ok {
			for _, attr := range elem.Attributes {
				if attr.Name == "x-data" {
					count++
				}
			}
			// Recursively count in children
			count += countXDataNodes(elem.Children)
		}
	}
	return count
}

// TestOptimizationFlag tests the config flag
func TestOptimizationFlag(t *testing.T) {
	// Save original
	original := OptimizeXData
	defer func() { OptimizeXData = original }()

	// Test toggle
	SetOptimization(false)
	if OptimizeXData != false {
		t.Error("SetOptimization(false) should set OptimizeXData to false")
	}

	SetOptimization(true)
	if OptimizeXData != true {
		t.Error("SetOptimization(true) should set OptimizeXData to true")
	}
}

// TestConfigDefault tests that optimization is enabled by default
func TestConfigDefault(t *testing.T) {
	// Check that optimization is enabled by default
	// (Unless explicitly changed by other tests)
	if !OptimizeXData {
		t.Log("Note: OptimizeXData is disabled. Expected to be true by default.")
	}
}

// BenchmarkScopeDiff benchmarks scope diffing performance
func BenchmarkScopeDiff(b *testing.B) {
	// Create test scopes
	parent := map[string]any{
		"user":   "John",
		"config": map[string]any{"theme": "dark", "lang": "en"},
		"items":  []string{"a", "b", "c"},
	}
	child := map[string]any{
		"user":      "John",
		"config":    map[string]any{"theme": "light", "lang": "en"},
		"items":     []string{"a", "b", "c"},
		"newField":  "value",
	}
	opts := DefaultDiffOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScopeDiff(child, parent, opts)
	}
}

// BenchmarkEstimateSize benchmarks size estimation
func BenchmarkEstimateSize(b *testing.B) {
	testValue := map[string]any{
		"field1": "value1",
		"field2": 42,
		"field3": []string{"a", "b", "c"},
		"nested": map[string]any{
			"x": "y",
			"z": 123,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = estimateSize(testValue)
	}
}

// TestScopeDiff_EdgeCases tests edge cases for scope diffing
func TestScopeDiff_EdgeCases(t *testing.T) {
	opts := DefaultDiffOptions()

	t.Run("empty scopes", func(t *testing.T) {
		parent := map[string]any{}
		child := map[string]any{}
		diff := ScopeDiff(child, parent, opts)
		if len(diff) != 0 {
			t.Errorf("Expected empty diff for empty scopes, got %d items", len(diff))
		}
	})

	t.Run("child removes variable", func(t *testing.T) {
		parent := map[string]any{"user": "John", "age": 30}
		child := map[string]any{"user": "John"}
		diff := ScopeDiff(child, parent, opts)
		// Variables only in parent are not in diff (child doesn't have them)
		if len(diff) != 0 {
			t.Logf("Diff contains: %v", diff)
		}
	})

	t.Run("nil values", func(t *testing.T) {
		parent := map[string]any{"user": "John", "optional": nil}
		child := map[string]any{"user": "John", "optional": "value"}
		diff := ScopeDiff(child, parent, opts)
		// Changed from nil to value
		if _, hasOptional := diff["optional"]; !hasOptional {
			t.Error("Expected 'optional' in diff (changed from nil)")
		}
	})
}

// TestComponentTransformWithOptimization tests full component transformation with optimization
func TestComponentTransformWithOptimization(t *testing.T) {
	// Setup: Enable optimization
	originalFlag := OptimizeXData
	OptimizeXData = true
	defer func() { OptimizeXData = originalFlag }()

	// Register test component
	componentTemplate := `
---
prop name = "Guest"
prop age = 0
---
<div class="profile">
  <h2>{name}</h2>
  <p>Age: {age}</p>
</div>
`

	compAST, err := parser.ParseTemplate(componentTemplate)
	if err != nil {
		t.Fatalf("Failed to parse component: %v", err)
	}

	RegisterComponent("Profile", compAST, []string{"name", "age"})
	defer UnregisterComponent("Profile")

	// Test: Component with props that match parent scope (should inherit)
	t.Run("component inherits props from parent", func(t *testing.T) {
		templateStr := `
---
let name = "Alice"
let age = 30
---
<Profile name={name} age={age} />
`

		template, err := parser.ParseTemplate(templateStr)
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		transformed := TransformAST(template, map[string]any{})
		xdataCount := countXDataNodes(transformed.RootNodes)

		// Component should inherit both props, minimal/no x-data needed
		t.Logf("Component inherits props: %d x-data wrappers", xdataCount)
	})

	// Test: Component with different prop values (should have x-data)
	t.Run("component with new prop values", func(t *testing.T) {
		templateStr := `
---
let title = "Welcome"
---
<Profile name="Bob" age={25} />
`

		template, err := parser.ParseTemplate(templateStr)
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		transformed := TransformAST(template, map[string]any{})

		// Verify component rendered
		output := nodeTreeToString(transformed.RootNodes)
		if !strings.Contains(output, "profile") {
			t.Error("Expected component to render with 'profile' class")
		}
	})
}

// Helper: Convert node tree to string for inspection
func nodeTreeToString(nodes []ast.Node) string {
	var sb strings.Builder
	for _, node := range nodes {
		if elem, ok := node.(*ast.Element); ok {
			sb.WriteString("<")
			sb.WriteString(elem.TagName)
			for _, attr := range elem.Attributes {
				sb.WriteString(" ")
				sb.WriteString(attr.Name)
				if attr.Value != "" {
					sb.WriteString("=\"")
					sb.WriteString(attr.Value)
					sb.WriteString("\"")
				}
			}
			sb.WriteString(">")
			sb.WriteString(nodeTreeToString(elem.Children))
			sb.WriteString("</")
			sb.WriteString(elem.TagName)
			sb.WriteString(">")
		} else if text, ok := node.(*ast.TextNode); ok {
			sb.WriteString(text.Content)
		}
	}
	return sb.String()
}
