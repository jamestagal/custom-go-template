package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// TestClassExpressionTransformation tests that {expr} in class attributes are transformed to :class bindings
func TestClassExpressionTransformation(t *testing.T) {
	template := `---
prop type = "success"
---

<div class="notification notification-{type}">
  <p>Test</p>
</div>`

	// Parse template
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with props
	props := map[string]any{
		"type": "success",
	}
	transformedAST := TransformAST(templateAST, props)

	// The wrapper div has x-data, the inner div should have :class
	// Find the wrapper div
	var wrapperDiv *ast.Element
	for _, node := range transformedAST.RootNodes {
		if el, ok := node.(*ast.Element); ok && el.TagName == "div" {
			// Check if it has x-data (wrapper div)
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

	t.Logf("Wrapper div has %d children", len(wrapperDiv.Children))

	// Find the inner div (first child should be the original div)
	var innerDiv *ast.Element
	for _, child := range wrapperDiv.Children {
		if el, ok := child.(*ast.Element); ok && el.TagName == "div" {
			innerDiv = el
			break
		}
	}

	if innerDiv == nil {
		t.Fatalf("Expected to find inner div as child of wrapper, found children: %+v", wrapperDiv.Children)
	}

	// Find the :class attribute on the inner div
	var classAttr *ast.Attribute
	for i, attr := range innerDiv.Attributes {
		if attr.Name == ":class" {
			classAttr = &innerDiv.Attributes[i]
			break
		}
	}

	if classAttr == nil {
		t.Fatalf("Expected :class attribute on inner div, found attributes: %+v", innerDiv.Attributes)
	}

	// Should contain mixed expression with static and dynamic parts
	attrValue := classAttr.Value
	if !strings.Contains(attrValue, `'notification `) && !strings.Contains(attrValue, `'notification-'`) {
		t.Errorf("Expected mixed static/dynamic class expression, got: %s", attrValue)
	}

	// Should contain the type variable reference
	if !strings.Contains(attrValue, `type`) {
		t.Errorf("Expected type variable in class expression, got: %s", attrValue)
	}

	t.Logf("Transformed :class value: %s", attrValue)
}

// TestPureExpressionInClassAttribute tests that class="{expr}" transforms to :class="expr"
func TestPureExpressionInClassAttribute(t *testing.T) {
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

	// Transform with props
	props := map[string]any{
		"className": "active",
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
		t.Fatal("Expected to find div element in transformed AST")
	}

	// Should have :class attribute with value "className"
	var found bool
	for _, attr := range divElement.Attributes {
		if attr.Name == ":class" && attr.Value == "className" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected :class=\"className\" binding, found attributes: %+v", divElement.Attributes)
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
		t.Fatal("Expected to find div element in transformed AST")
	}

	// Should have regular class attribute, not :class
	var hasStaticClass bool
	var hasDynamicClass bool
	for _, attr := range divElement.Attributes {
		if attr.Name == "class" && attr.Value == "static-class" {
			hasStaticClass = true
		}
		if attr.Name == ":class" {
			hasDynamicClass = true
		}
	}

	if !hasStaticClass {
		t.Errorf("Expected static class attribute, found attributes: %+v", divElement.Attributes)
	}

	if hasDynamicClass {
		t.Errorf("Expected no :class binding for static classes, found attributes: %+v", divElement.Attributes)
	}

	t.Logf("Transformed attributes: %+v", divElement.Attributes)
}

// TestAttributeWithStoreAndRegularExpression tests mixed store + regular expressions
func TestAttributeWithStoreAndRegularExpression(t *testing.T) {
	template := `---
prop variant = "large"
---

<div class="{variant} theme-{$theme.mode}">
  <p>Test</p>
</div>`

	// Parse template
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform with props
	props := map[string]any{
		"variant": "large",
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
		t.Fatal("Expected to find div element in transformed AST")
	}

	// Find :class attribute
	var classAttr *ast.Attribute
	for i, attr := range divElement.Attributes {
		if attr.Name == ":class" {
			classAttr = &divElement.Attributes[i]
			break
		}
	}

	if classAttr == nil {
		t.Fatalf("Expected :class attribute, found attributes: %+v", divElement.Attributes)
	}

	// Should contain both variant and $store.theme.mode
	attrValue := classAttr.Value
	if !strings.Contains(attrValue, `variant`) {
		t.Errorf("Expected variant in :class value, got: %s", attrValue)
	}

	if !strings.Contains(attrValue, `$store.theme.mode`) {
		t.Errorf("Expected $store.theme.mode in :class value, got: %s", attrValue)
	}

	t.Logf("Transformed :class value: %s", attrValue)
}
