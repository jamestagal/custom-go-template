package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/tests/testutils"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestAlpineIntegration(t *testing.T) {
	tests := []struct {
		name     string
		template string
		props    map[string]any
		expected string
	}{
		{
			name: "basic_expressions",
			template: `
				<div>
					<h1>{title}</h1>
					<p>{description}</p>
				</div>
			`,
			props: map[string]any{
				"title":       "Hello Alpine",
				"description": "This is a test",
			},
			// Build-time resolution: expressions with values in props get resolved
			expected: `<div><h1>Hello Alpine</h1><p>This is a test</p></div>`,
		},
		{
			name: "conditional_rendering",
			template: `
				<div>
					<h1>{title}</h1>
					{if showDetails}
						<p>{description}</p>
					{/if}
				</div>
			`,
			// NOTE: Don't provide showDetails in props to test RUNTIME x-if
			props: map[string]any{
				"title":       "Product",
				"description": "Product details",
			},
			// Runtime x-if since showDetails not in props
			expected: `<div><h1>Product</h1><template x-if="showDetails"><p>Product details</p></template></div>`,
		},
		{
			name: "loop_rendering",
			template: `
				<div>
					<h1>{title}</h1>
					<ul>
						{for item in items}
							<li>{item.name}</li>
						{/for}
					</ul>
				</div>
			`,
			// NOTE: Don't provide items in props to test RUNTIME x-for
			props: map[string]any{
				"title": "Shopping List",
			},
			// Runtime x-for since items not in props
			// Format: "item in items" (no index) when parsed from template
			expected: `<div><h1>Shopping List</h1><ul><template x-for="item in items"><li><span x-text="item.name"></span></li></template></ul></div>`,
		},
		{
			name: "component_integration",
			template: `
				<div>
					<h1>{title}</h1>
					<Button label="Click me" onClick={handleClick} />
				</div>
			`,
			props: map[string]any{
				"title":       "Component Test",
				"handleClick": "() => { alert('Button clicked!') }",
			},
			// Build-time resolution for expressions
			expected: `<div><h1>Component Test</h1><div x-component="Button" data-prop-label="Click me" data-prop-onClick="handleClick"></div></div>`,
		},
		{
			name: "nested_conditionals_and_loops",
			template: `
				<div>
					<h1>{title}</h1>
					{if categories.length > 0}
						<div>
							{for category in categories}
								<div>
									<h2>{category.name}</h2>
									{if category.items.length > 0}
										<ul>
											{for item in category.items}
												<li>{item.name} - ${item.price}</li>
											{/for}
										</ul>
									{else}
										<p>No items in this category</p>
									{/if}
								</div>
							{/for}
						</div>
					{else}
						<p>No categories found</p>
					{/if}
				</div>
			`,
			// Provide categories to test build-time loop expansion
			// Note: This tests that loops expand correctly with actual data
			props: map[string]any{
				"title": "Product Catalog",
				"categories": []map[string]any{
					{
						"name": "Electronics",
						"items": []map[string]any{
							{"name": "Laptop", "price": 999.99},
							{"name": "Phone", "price": 699.99},
						},
					},
					{
						"name":  "Books",
						"items": []map[string]any{},
					},
				},
			},
			// Build-time loop expansion with runtime inner conditionals
			// Alpine.js 3.x: else uses negated x-if
			expected: `<div><h1>Product Catalog</h1><template x-if="categories.length > 0"><div><div><h2>Electronics</h2><template x-if="category.items.length > 0"><ul><li>Laptop - $999.99</li><li>Phone - $699.99</li></ul></template><template x-if="!(category.items.length > 0)"><p>No items in this category</p></template></div><div><h2>Books</h2><template x-if="category.items.length > 0"><ul></ul></template><template x-if="!(category.items.length > 0)"><p>No items in this category</p></template></div></div></template><template x-if="!(categories.length > 0)"><p>No categories found</p></template></div>`,
		},
		{
			name: "dynamic_components_with_conditionals",
			template: `
				<div>
					<h1>{title}</h1>
					{if isAdmin}
						<AdminPanel user={currentUser} />
					{else}
						<UserProfile user={currentUser} />
					{/if}
				</div>
			`,
			// NOTE: Don't provide isAdmin in props to test RUNTIME x-if
			props: map[string]any{
				"title":       "User Dashboard",
				"currentUser": map[string]any{"name": "John Doe", "role": "admin"},
			},
			// Runtime x-if/negated x-if since isAdmin not in props
			// Alpine.js 3.x: else uses negated x-if
			expected: `<div><h1>User Dashboard</h1><template x-if="isAdmin"><div x-component="AdminPanel" data-prop-user="currentUser"></div></template><template x-if="!(isAdmin)"><div x-component="UserProfile" data-prop-user="currentUser"></div></template></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the template
			parsedTemplate, err := parser.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			// Transform the AST
			transformedTemplate := transformer.TransformAST(parsedTemplate, tt.props)

			// Render the transformed template to a string
			var sb strings.Builder
			renderIntegrationTemplate(&sb, transformedTemplate)
			output := sb.String()

			// Normalize whitespace for comparison
			normalizedOutput := testutils.NormalizeWhitespace(output)
			normalizedExpected := testutils.NormalizeWhitespace(tt.expected)

			if normalizedOutput != normalizedExpected {
				t.Errorf("Expected output to be:\n%s\n\nBut got:\n%s", normalizedExpected, normalizedOutput)
			}
		})
	}
}

// Helper function to render a template to string for testing
func renderIntegrationTemplate(sb *strings.Builder, template *ast.Template) {
	for _, node := range template.RootNodes {
		renderIntegrationNode(sb, node)
	}
}

// Helper function to render a node to string for testing
func renderIntegrationNode(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Template:
		// Render template as Alpine.js x-data wrapper
		// The actual x-data wrapper should be added by the transformer
		// Just render the root nodes here
		for _, child := range n.RootNodes {
			renderIntegrationNode(sb, child)
		}

	case *ast.Element:
		// Special case for x-data wrapper
		if n.TagName == "div" && len(n.Attributes) > 0 {
			for _, attr := range n.Attributes {
				if attr.Name == "x-data" {
					// This is the root wrapper with x-data
					sb.WriteString("<div x-data='")
					sb.WriteString(attr.Value)
					sb.WriteString("'>")

					// Render children
					for _, child := range n.Children {
						renderIntegrationNode(sb, child)
					}

					sb.WriteString("</div>")
					return
				}
			}
		}

		// Special case for template with x-if, x-else-if, or x-else
		// Alpine.js 3.x: else branches use negated x-if pattern, NOT x-else
		if n.TagName == "template" && len(n.Attributes) > 0 {
			for _, attr := range n.Attributes {
				if attr.Name == "x-if" {
					// Render all x-if conditions (including negated ones for else branches)
					sb.WriteString("<template x-if=\"")
					sb.WriteString(attr.Value)
					sb.WriteString("\">")

					// Render children
					for _, child := range n.Children {
						renderIntegrationNode(sb, child)
					}

					sb.WriteString("</template>")
					return
				} else if attr.Name == "x-else-if" {
					// This is an else-if condition
					sb.WriteString("<template x-else-if=\"")
					sb.WriteString(attr.Value)
					sb.WriteString("\">")

					// Render children
					for _, child := range n.Children {
						renderIntegrationNode(sb, child)
					}

					sb.WriteString("</template>")
					return
				}
			}
		}

		// Check if this is a component (has x-component attribute)
		isComponent := false
		componentName := ""

		for _, attr := range n.Attributes {
			if attr.Name == "x-component" {
				isComponent = true
				componentName = attr.Value
				break
			}
		}

		if isComponent {
			sb.WriteString("<div x-component=\"")
			sb.WriteString(componentName)
			sb.WriteString("\"")

			// Render props as data-prop-* attributes
			for _, attr := range n.Attributes {
				if strings.HasPrefix(attr.Name, "data-prop-") {
					sb.WriteString(" ")
					sb.WriteString(attr.Name)

					// Use the specified quote style if available, otherwise default to double quotes
					openQuote := "\""
					closeQuote := "\""
					if attr.QuoteStyle == "'" {
						openQuote = "'"
						closeQuote = "'"
					}

					sb.WriteString("=")
					sb.WriteString(openQuote)
					sb.WriteString(attr.Value)
					sb.WriteString(closeQuote)
				}
			}

			sb.WriteString("></div>")
			return
		}

		// Regular element handling
		sb.WriteString("<")
		sb.WriteString(n.TagName)

		// Render attributes
		for _, attr := range n.Attributes {
			sb.WriteString(" ")
			sb.WriteString(attr.Name)
			if attr.Value != "" {
				// Use the specified quote style if available, otherwise default to double quotes
				openQuote := "\""
				closeQuote := "\""
				if attr.QuoteStyle == "'" {
					openQuote = "'"
					closeQuote = "'"
				}

				sb.WriteString("=")
				sb.WriteString(openQuote)
				sb.WriteString(attr.Value)
				sb.WriteString(closeQuote)
			}
		}

		if n.SelfClosing {
			sb.WriteString(" />")
			return
		}

		sb.WriteString(">")

		// Render children
		for _, child := range n.Children {
			renderIntegrationNode(sb, child)
		}

		sb.WriteString("</")
		sb.WriteString(n.TagName)
		sb.WriteString(">")

	case *ast.TextNode:
		// Trim whitespace-only text nodes to reduce differences in test output
		if strings.TrimSpace(n.Content) == "" {
			return
		}
		sb.WriteString(n.Content)

	case *ast.ExpressionNode:
		sb.WriteString("<span x-text=\"")
		sb.WriteString(n.Expression)
		sb.WriteString("\"></span>")

	case *ast.ComponentNode:
		// Render component as a div with x-component attribute
		sb.WriteString("<div x-component=\"")
		sb.WriteString(n.Name)
		sb.WriteString("\"")

		// Render props as data-prop-* attributes
		for _, prop := range n.Props {
			sb.WriteString(" data-prop-")
			sb.WriteString(prop.Name)
			sb.WriteString("=\"")

			if prop.IsDynamic {
				// For dynamic props, use the variable name
				propValue := prop.Value
				// Remove curly braces if present
				propValue = strings.TrimPrefix(propValue, "{")
				propValue = strings.TrimSuffix(propValue, "}")
				propValue = strings.TrimSpace(propValue)
				sb.WriteString(propValue)
			} else {
				// For static props, use the literal value
				sb.WriteString(prop.Value)
			}

			sb.WriteString("\"")
		}

		sb.WriteString("></div>")
	}
}
