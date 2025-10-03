package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestComponentPropsTransformation(t *testing.T) {
	// Register component templates for testing
	registerTestPropComponents()
	
	tests := []struct {
		name     string
		input    ast.Node
		props    map[string]any
		expected string
	}{
		{
			name: "component_with_expression_props",
			input: &ast.ComponentNode{
				Name: "DataCard",
				Props: []ast.ComponentProp{
					{
						Name:        "title",
						Value:       "getTitle()",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "count",
						Value:       "items.length",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"items": []any{"item1", "item2", "item3"},
			},
			expected: `<div x-data="{ &quot;count&quot;: &quot;items.length&quot;, &quot;title&quot;: &quot;getTitle()&quot; }">DataCard Component</div>`,
		},
		{
			name: "component_with_mixed_props",
			input: &ast.ComponentNode{
				Name: "UserProfile",
				Props: []ast.ComponentProp{
					{
						Name:        "username",
						Value:       "currentUser.name",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "isAdmin",
						Value:       "true",
						IsShorthand: false,
						IsDynamic:   false,
					},
					{
						Name:        "avatar",
						Value:       "'/images/default.png'",
						IsShorthand: false,
						IsDynamic:   false,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"currentUser": map[string]any{
					"name": "John Doe",
					"id":   "123",
				},
			},
			expected: `<div x-data="{ &quot;avatar&quot;: &quot;/images/default.png&quot;, &quot;isAdmin&quot;: true, &quot;username&quot;: &quot;currentUser.name&quot; }">UserProfile Component</div>`,
		},
		{
			name: "component_with_complex_expression_props",
			input: &ast.ComponentNode{
				Name: "ProductCard",
				Props: []ast.ComponentProp{
					{
						Name:        "price",
						Value:       "formatCurrency(product.price)",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "discount",
						Value:       "product.price > 100 ? '10%' : '5%'",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"product": map[string]any{
					"price": 149.99,
				},
			},
			expected: `<div x-data="{ &quot;discount&quot;: &quot;product.price > 100 ? '10%' : '5%'&quot;, &quot;price&quot;: &quot;formatCurrency(product.price)&quot; }">ProductCard Component</div>`,
		},
		{
			name: "component_with_shorthand_and_spread_props",
			input: &ast.ComponentNode{
				Name: "Form",
				Props: []ast.ComponentProp{
					{
						Name:        "formData",
						Value:       "formData",
						IsShorthand: true,
						IsDynamic:   true,
					},
					{
						Name:        "errors",
						Value:       "validationErrors",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"formData": map[string]any{
					"name":  "",
					"email": "",
				},
				"validationErrors": map[string]any{
					"name": "Name is required",
				},
			},
			expected: `<div x-data="{ &quot;errors&quot;: &quot;validationErrors&quot;, &quot;formData&quot;: {&quot;email&quot;: &quot;&quot;, &quot;name&quot;: &quot;&quot;} }">Form Component</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a template with the component node as the only root node
			template := &ast.Template{
				RootNodes: []ast.Node{tt.input},
			}
			
			// Transform the template
			result := transformer.TransformAST(template, tt.props)
			
			// Check if we have any root nodes in the result
			if len(result.RootNodes) == 0 {
				t.Fatalf("Expected at least one node in the result, but got none")
			}
			
			// Get the root node
			rootNode := result.RootNodes[0]
			
			// Convert the root node to a string for comparison
			var sb strings.Builder
			renderComponentPropsNode(&sb, rootNode)
			output := sb.String()

			if output != tt.expected {
				t.Errorf("Expected output to be %q, but got %q", tt.expected, output)
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderComponentPropsNode(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		sb.WriteString("<")
		sb.WriteString(n.TagName)
		
		// Render attributes
		for _, attr := range n.Attributes {
			sb.WriteString(" ")
			sb.WriteString(attr.Name)
			if attr.Value != "" {
				sb.WriteString("=\"")
				sb.WriteString(attr.Value)
				sb.WriteString("\"")
			}
		}
		
		if n.SelfClosing {
			sb.WriteString(" />")
			return
		}
		
		sb.WriteString(">")
		
		// Render children
		for _, child := range n.Children {
			renderComponentPropsNode(sb, child)
		}
		
		sb.WriteString("</")
		sb.WriteString(n.TagName)
		sb.WriteString(">")
		
	case *ast.TextNode:
		sb.WriteString(n.Content)
		
	case *ast.ExpressionNode:
		sb.WriteString("<span x-text=\"")
		sb.WriteString(n.Expression)
		sb.WriteString("\"></span>")
	}
}

// registerTestPropComponents registers test component templates for prop tests
func registerTestPropComponents() {
	// DataCard component
	dataCardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "DataCard Component"},
		},
	}
	transformer.RegisterComponent("DataCard", dataCardTemplate, []string{"title", "count"})
	
	// UserProfile component
	userProfileTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "UserProfile Component"},
		},
	}
	transformer.RegisterComponent("UserProfile", userProfileTemplate, []string{"username", "isAdmin", "avatar"})
	
	// ProductCard component
	productCardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "ProductCard Component"},
		},
	}
	transformer.RegisterComponent("ProductCard", productCardTemplate, []string{"price", "discount"})
	
	// Form component
	formTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Form Component"},
		},
	}
	transformer.RegisterComponent("Form", formTemplate, []string{"formData", "errors"})
}
