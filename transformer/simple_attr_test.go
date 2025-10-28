package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestTransformAttributesWithStores_RegularExpression tests that simple prop expressions
// are interpolated at BUILD-TIME when the value is available in dataScope
func TestTransformAttributesWithStores_RegularExpression(t *testing.T) {
	// Create an attribute with a regular expression
	attrs := []ast.Attribute{
		{
			Name:  "class",
			Value: "{className}",
		},
	}

	dataScope := map[string]any{
		"className": "active",
	}

	// Transform
	transformed := transformAttributeExpressions(attrs, dataScope)

	// Should have class attribute with BUILD-TIME interpolated value (not :class)
	// This is the expected behavior for Plenti where content props are resolved at build time
	found := false
	for _, attr := range transformed {
		if attr.Name == "class" && attr.Value == "active" && !attr.Dynamic {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected class=\"active\" (BUILD-TIME interpolation), got: %+v", transformed)
	}

	t.Logf("Transformed attributes: %+v", transformed)
}

// TestTransformAttributesWithStores_RuntimeBinding tests that expressions
// create runtime bindings when value is NOT available in dataScope
func TestTransformAttributesWithStores_RuntimeBinding(t *testing.T) {
	// Create an attribute with an expression NOT in dataScope
	attrs := []ast.Attribute{
		{
			Name:  "class",
			Value: "{dynamicClass}",
		},
	}

	// Empty dataScope - variable not available
	dataScope := map[string]any{}

	// Transform
	transformed := transformAttributeExpressions(attrs, dataScope)

	// Should have :class attribute for RUNTIME binding
	found := false
	for _, attr := range transformed {
		if attr.Name == "class" && attr.Value == "dynamicClass" && attr.Dynamic {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected :class=\"dynamicClass\" (RUNTIME binding), got: %+v", transformed)
	}

	t.Logf("Transformed attributes: %+v", transformed)
}

// TestTransformAttributesWithStores_MixedExpression tests mixed static/dynamic content
func TestTransformAttributesWithStores_MixedExpression(t *testing.T) {
	// Create an attribute with mixed static and expression content
	attrs := []ast.Attribute{
		{
			Name:  "class",
			Value: "notification notification-{type}",
		},
	}

	dataScope := map[string]any{
		"type": "success",
	}

	// Transform
	transformed := transformAttributeExpressions(attrs, dataScope)

	// Should have :class attribute with combined expression for RUNTIME evaluation
	// Mixed expressions always need runtime binding because we can't partially interpolate
	found := false
	for _, attr := range transformed {
		if attr.Name == "class" && attr.Dynamic {
			found = true
			t.Logf("Found :class with value: %s (Dynamic: %v)", attr.Value, attr.Dynamic)
			// Should contain the concatenated expression for runtime evaluation
		}
	}

	if !found {
		t.Errorf("Expected :class attribute with Dynamic=true (mixed expression), got: %+v", transformed)
	}

	t.Logf("Transformed attributes: %+v", transformed)
}
