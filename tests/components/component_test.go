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

	// Build-time expression resolution: {message} resolves to "Hello World" directly
	// since message is in the component's data scope with a known value.
	// No x-text span generated - the value is inlined at build-time.
	expected := `<div class="component-content" x-data="{ message: 'Hello World' }">Component Content: Hello World</div>`

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

	// Build-time expression resolution: {count} resolves to 42 directly
	// since count is in the component's data scope with a known value.
	// No x-text span generated - the value is inlined at build-time.
	expected := `<div class="dynamic-component" x-data="{ count: 42 }">Count: 42</div>`

	// Normalize whitespace for comparison
	normalizedHTML := testutils.NormalizeWhitespace(html)
	normalizedExpected := testutils.NormalizeWhitespace(expected)

	if normalizedHTML != normalizedExpected {
		t.Errorf("Dynamic props component transformation failed.\nExpected: %s\nGot: %s", normalizedExpected, normalizedHTML)
	}
}
