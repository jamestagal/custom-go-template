package components

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/tests/testutils"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestComponentTransformation(t *testing.T) {
	// Register a test component
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:  "class",
						Value: "component-content",
					},
				},
				Children: []ast.Node{
					&ast.TextNode{
						Content: "Component Content: ",
					},
					&ast.ExpressionNode{
						Expression: "message",
					},
				},
				SelfClosing: false,
			},
		},
	}

	// Register the component
	transformer.RegisterComponent("TestComponent", componentTemplate, []string{"message"})

	// Create a test template that uses the component
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.ComponentNode{
				Name: "TestComponent",
				Props: []ast.ComponentProp{
					{
						Name:      "message",
						Value:     "Hello World",
						IsDynamic: false,
					},
				},
				Dynamic: false,
			},
		},
	}

	// Transform the template
	transformedTemplate := transformer.TransformAST(template, map[string]any{})

	// Render the transformed template to HTML
	html := testutils.RenderNode(transformedTemplate.RootNodes[0])

	// Updated expectation: The transformer now outputs JavaScript object literal format
	// instead of JSON-escaped format. This is actually better for Alpine.js compatibility
	// as it matches Alpine's native data format (e.g., x-data="{ count: 0 }").
	// Old format: x-data="{&quot;message&quot;:&quot;Hello World&quot;}"
	// New format: x-data="{ message: 'Hello World' }" (JavaScript object literal)
	expected := `<div class="component-content" x-data="{ message: 'Hello World' }">Component Content: <span x-text="message"></span></div>`

	// Normalize whitespace for comparison
	normalizedHTML := testutils.NormalizeWhitespace(html)
	normalizedExpected := testutils.NormalizeWhitespace(expected)

	if normalizedHTML != normalizedExpected {
		t.Errorf("Component transformation failed.\nExpected: %s\nGot: %s", normalizedExpected, normalizedHTML)
	}
}

func TestDynamicPropsComponentTransformation(t *testing.T) {
	// Register a test component
	componentTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:  "class",
						Value: "dynamic-component",
					},
				},
				Children: []ast.Node{
					&ast.TextNode{
						Content: "Count: ",
					},
					&ast.ExpressionNode{
						Expression: "count",
					},
				},
				SelfClosing: false,
			},
		},
	}

	// Register the component
	transformer.RegisterComponent("DynamicComponent", componentTemplate, []string{"count"})

	// Create a test template that uses the component with dynamic props
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.ComponentNode{
				Name: "DynamicComponent",
				Props: []ast.ComponentProp{
					{
						Name:      "count",
						Value:     "{parentCount}",
						IsDynamic: true,
					},
				},
				Dynamic: false,
			},
		},
	}

	// Transform the template with parent scope
	parentScope := map[string]any{
		"parentCount": 42,
	}
	transformedTemplate := transformer.TransformAST(template, parentScope)

	// Render the transformed template to HTML
	html := testutils.RenderNode(transformedTemplate.RootNodes[0])

	// Updated expectation: When dynamic props are used, the transformer now creates a wrapper div
	// with x-data containing both the parent scope variable (for reactivity) and the component scope.
	// This ensures that dynamic props remain reactive to parent changes.
	// The wrapper pattern: <div x-data="{count:null,parentCount:42}"><component x-data="{ count: 42 }">...</component></div>
	// This is the correct behavior for reactive dynamic props in Alpine.js.
	expected := `<div x-data="{count:null,parentCount:42}"><div class="dynamic-component" x-data="{ count: 42 }">Count: <span x-text="count"></span></div></div>`

	// Normalize whitespace for comparison
	normalizedHTML := testutils.NormalizeWhitespace(html)
	normalizedExpected := testutils.NormalizeWhitespace(expected)

	if normalizedHTML != normalizedExpected {
		t.Errorf("Dynamic props component transformation failed.\nExpected: %s\nGot: %s", normalizedExpected, normalizedHTML)
	}
}
