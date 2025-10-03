package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestStaticComponentTransformation(t *testing.T) {
	// Register component templates for testing
	registerTestComponents()
	
	tests := []struct {
		name     string
		input    ast.Node
		props    map[string]any
		expected string
	}{
		{
			name: "basic_component_no_props",
			input: &ast.ComponentNode{
				Name:    "Button",
				Props:   []ast.ComponentProp{},
				Dynamic: false,
			},
			props:    map[string]any{},
			expected: `<div x-data="{}">Button Component</div>`,
		},
		{
			name: "component_with_static_props",
			input: &ast.ComponentNode{
				Name: "Card",
				Props: []ast.ComponentProp{
					{
						Name:        "title",
						Value:       "Welcome",
						IsShorthand: false,
						IsDynamic:   false,
					},
					{
						Name:        "subtitle",
						Value:       "Hello World",
						IsShorthand: false,
						IsDynamic:   false,
					},
				},
				Dynamic: false,
			},
			props:    map[string]any{},
			expected: `<div x-data="{ &quot;subtitle&quot;: &quot;Hello World&quot;, &quot;title&quot;: &quot;Welcome&quot; }">Card Component</div>`,
		},
		{
			name: "component_with_dynamic_props",
			input: &ast.ComponentNode{
				Name: "UserProfile",
				Props: []ast.ComponentProp{
					{
						Name:        "user",
						Value:       "currentUser",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "showDetails",
						Value:       "isAdmin",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"currentUser": map[string]any{
					"name": "John Doe",
					"role": "Admin",
				},
				"isAdmin": true,
			},
			expected: `<div x-data="{ &quot;showDetails&quot;: true, &quot;user&quot;: {&quot;name&quot;: &quot;John Doe&quot;, &quot;role&quot;: &quot;Admin&quot;} }">UserProfile Component</div>`,
		},
		{
			name: "component_with_shorthand_props",
			input: &ast.ComponentNode{
				Name: "ProductCard",
				Props: []ast.ComponentProp{
					{
						Name:        "product",
						Value:       "product",
						IsShorthand: true,
						IsDynamic:   true,
					},
					{
						Name:        "inStock",
						Value:       "inStock",
						IsShorthand: true,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"product": map[string]any{
					"id":    "123",
					"name":  "Laptop",
					"price": 999.99,
				},
				"inStock": true,
			},
			expected: `<div x-data="{ &quot;inStock&quot;: true, &quot;product&quot;: {&quot;id&quot;: &quot;123&quot;, &quot;name&quot;: &quot;Laptop&quot;, &quot;price&quot;: 999.99} }">ProductCard Component</div>`,
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
			renderComponentNode(&sb, rootNode)
			output := sb.String()

			if output != tt.expected {
				t.Errorf("Expected output to be %q, but got %q", tt.expected, output)
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderComponentNode(sb *strings.Builder, node ast.Node) {
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
			renderComponentNode(sb, child)
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

// registerTestComponents registers test component templates
func registerTestComponents() {
	// Button component
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Button Component"},
		},
	}
	transformer.RegisterComponent("Button", buttonTemplate, []string{})
	
	// Card component
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Card Component"},
		},
	}
	transformer.RegisterComponent("Card", cardTemplate, []string{"title", "subtitle"})
	
	// UserProfile component
	userProfileTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "UserProfile Component"},
		},
	}
	transformer.RegisterComponent("UserProfile", userProfileTemplate, []string{"user", "showDetails"})
	
	// ProductCard component
	productCardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "ProductCard Component"},
		},
	}
	transformer.RegisterComponent("ProductCard", productCardTemplate, []string{"product", "inStock"})
}
