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

	// Expected HTML - component is transformed inline with x-data on the component's root element
	expected := `<div class="component-content" x-data="{&quot;message&quot;:&quot;Hello World&quot;}">Component Content: <span x-text="message"></span></div>`

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

	// Expected HTML (with dynamic prop resolved to value)
	expected := `<div class="dynamic-component" x-data="{&quot;count&quot;:42}">Count: <span x-text="count"></span></div>`

	// Normalize whitespace for comparison
	normalizedHTML := testutils.NormalizeWhitespace(html)
	normalizedExpected := testutils.NormalizeWhitespace(expected)

	if normalizedHTML != normalizedExpected {
		t.Errorf("Dynamic props component transformation failed.\nExpected: %s\nGot: %s", normalizedExpected, normalizedHTML)
	}
}
