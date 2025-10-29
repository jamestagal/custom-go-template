package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// TestClassExpressionInterpolation tests that expressions in class attributes are handled
func TestClassExpressionInterpolation(t *testing.T) {
	template := `<div class="static {dynamicClass}">
  <p>Test</p>
</div>`

	// Parse template
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with props
	props := map[string]any{
		"dynamicClass": "highlight",
	}
	transformedAST := TransformAST(templateAST, props)

	// Find the div element in transformed AST
	var divElement *ast.Element
	for _, node := range transformedAST.RootNodes {
		if el, ok := node.(*ast.Element); ok && el.TagName == "div" {
			divElement = el
			break
		}
	}

	if divElement == nil {
		t.Fatal("Expected to find div element")
	}

	// With mixed static and dynamic content, this becomes a dynamic expression
	// The class attribute should be dynamic (x-bind:class)
	var found bool
	for _, attr := range divElement.Attributes {
		if attr.Name == "class" && attr.Dynamic {
			found = true
			// The value should be an expression like: 'static ' + dynamicClass
			if !strings.Contains(attr.Value, "dynamicClass") {
				t.Errorf("Expected dynamic expression to reference dynamicClass, got: %s", attr.Value)
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected dynamic class attribute, found: %+v", divElement.Attributes)
	}
}

// TestMultipleExpressionInClass tests multiple expressions in single class attribute
func TestMultipleExpressionInClass(t *testing.T) {
	template := `<div class="{baseClass} {modifierClass}">
  <p>Test</p>
</div>`

	// Parse template
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with props
	props := map[string]any{
		"baseClass":     "btn",
		"modifierClass": "btn-primary",
	}
	transformedAST := TransformAST(templateAST, props)

	// Find the div element
	var divElement *ast.Element
	for _, node := range transformedAST.RootNodes {
		if el, ok := node.(*ast.Element); ok && el.TagName == "div" {
			divElement = el
			break
		}
	}

	if divElement == nil {
		t.Fatal("Expected to find div element")
	}

	// Multiple expressions with space between creates a concatenated expression
	// The class attribute should be dynamic
	var found bool
	for _, attr := range divElement.Attributes {
		if attr.Name == "class" {
			found = true
			// Either fully interpolated or dynamic expression with references
			if attr.Dynamic {
				// Should reference both variables
				if !strings.Contains(attr.Value, "baseClass") || !strings.Contains(attr.Value, "modifierClass") {
					t.Errorf("Expected dynamic expression to reference both variables, got: %s", attr.Value)
				}
			} else {
				// Fully interpolated - should contain both values
				if !strings.Contains(attr.Value, "btn") || !strings.Contains(attr.Value, "primary") {
					t.Errorf("Expected interpolated class to contain both values, got: %s", attr.Value)
				}
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected class attribute, found: %+v", divElement.Attributes)
	}
}

// TestPureExpressionInClassAttribute tests build-time interpolation when data is available
func TestPureExpressionInClassAttribute(t *testing.T) {
	// Disable optimization to get wrapper behavior for testing
	originalFlag := OptimizeXData
	OptimizeXData = false
	defer func() { OptimizeXData = originalFlag }()

	template := `---
prop className = "active"
---

<div class="{className}">
  <p>Test</p>
</div>`

	// Parse template
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with props - this should do build-time interpolation
	props := map[string]any{
		"className": "active",
	}
	transformedAST := TransformAST(templateAST, props)

	// Find the wrapper div (has x-data) first
	var wrapperDiv *ast.Element
	for _, node := range transformedAST.RootNodes {
		if el, ok := node.(*ast.Element); ok && el.TagName == "div" {
			// Check if it has x-data
			for _, attr := range el.Attributes {
				if attr.Name == "x-data" {
					wrapperDiv = el
					break
				}
			}
			if wrapperDiv != nil {
				break
			}
		}
	}

	if wrapperDiv == nil {
		t.Fatal("Expected to find wrapper div with x-data")
	}

	// Find the actual div with class attribute (child of wrapper)
	var divElement *ast.Element
	for _, child := range wrapperDiv.Children {
		if el, ok := child.(*ast.Element); ok && el.TagName == "div" {
			divElement = el
			break
		}
	}

	if divElement == nil {
		t.Fatal("Expected to find inner div element with class attribute")
	}

	// UPDATED: With build-time interpolation, when data is available at build time,
	// the expression should be interpolated to the actual value, not kept as a dynamic binding
	var found bool
	for _, attr := range divElement.Attributes {
		if attr.Name == "class" && attr.Value == "active" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected build-time interpolated class=\"active\", found attributes: %+v", divElement.Attributes)
	}

	t.Logf("Transformed attributes: %+v", divElement.Attributes)
}

// TestStaticClassAttribute tests that static class attributes are not changed
func TestStaticClassAttribute(t *testing.T) {
	template := `<div class="static-class">
  <p>Test</p>
</div>`

	// Parse template
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform without props
	transformedAST := TransformAST(templateAST, map[string]any{})

	// Find the div element in transformed AST
	var divElement *ast.Element
	for _, node := range transformedAST.RootNodes {
		if el, ok := node.(*ast.Element); ok && el.TagName == "div" {
			divElement = el
			break
		}
	}

	if divElement == nil {
		t.Fatal("Expected to find div element")
	}

	// Should have static class attribute unchanged
	var found bool
	for _, attr := range divElement.Attributes {
		if attr.Name == "class" && attr.Value == "static-class" && !attr.Dynamic {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected static class=\"static-class\", found: %+v", divElement.Attributes)
	}
}
