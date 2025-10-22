package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

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

	// Should have :class attribute
	found := false
	for _, attr := range transformed {
		if attr.Name == ":class" && attr.Value == "className" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected :class=\"className\", got: %+v", transformed)
	}

	t.Logf("Transformed attributes: %+v", transformed)
}

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

	// Should have :class attribute with combined expression
	found := false
	for _, attr := range transformed {
		if attr.Name == ":class" {
			found = true
			t.Logf("Found :class with value: %s", attr.Value)
			// Should contain both static parts and the type variable
			// Example: 'notification ' + 'notification-' + type
		}
	}

	if !found {
		t.Errorf("Expected :class attribute, got: %+v", transformed)
	}

	t.Logf("Transformed attributes: %+v", transformed)
}
