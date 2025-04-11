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
					<h1>{title }</h1>
					<p>{description }</p>
				</div>
			`,
			props: map[string]any{
				"title":       "Hello Alpine",
				"description": "This is a test",
			},
			expected: `<div x-data='{"title":"Hello Alpine","description":"This is a test"}'><div><h1><span x-text="title"></span></h1><p><span x-text="description"></span></p></div></div>`,
		},
		{
			name: "conditional_rendering",
			template: `
				<div>
					<h1>{title }</h1>
					{#if showDetails}
						<p>{description }</p>
					{/if}
				</div>
			`,
			props: map[string]any{
				"title":       "Product",
				"description": "Product details",
				"showDetails": true,
			},
			expected: `<div x-data='{"title":"Product","description":"Product details","showDetails":true}'><div><h1><span x-text="title"></span></h1><template x-if="showDetails"><p><span x-text="description"></span></p></template></div></div>`,
		},
		{
			name: "loop_rendering",
			template: `
				<div>
					<h1>{title }</h1>
					<ul>
						{#each items as item}
							<li>{item.name }</li>
						{/each}
					</ul>
				</div>
			`,
			props: map[string]any{
				"title": "Shopping List",
				"items": []map[string]any{
					{"name": "Apples"},
					{"name": "Bananas"},
					{"name": "Oranges"},
				},
			},
			expected: `<div x-data='{"title":"Shopping List","items":[{"name":"Apples"},{"name":"Bananas"},{"name":"Oranges"}]}'><div><h1><span x-text="title"></span></h1><ul><template x-for="item in items"><li><span x-text="item.name"></span></li></template></ul></div></div>`,
		},
		{
			name: "component_integration",
			template: `
				<div>
					<h1>{title }</h1>
					<Button label="Click me" onClick={handleClick} />
				</div>
			`,
			props: map[string]any{
				"title":       "Component Test",
				"handleClick": "() => { alert('Button clicked!') }",
			},
			expected: `<div x-data='{"title":"Component Test","handleClick":() => { alert('Button clicked!') }'><div><h1><span x-text="title"></span></h1><div x-component="Button" data-prop-label="Click me" data-prop-onClick="handleClick"></div></div></div>`,
		},
		{
			name: "nested_conditionals_and_loops",
			template: `
				<div>
					<h1>{title }</h1>
					{#if categories.length > 0}
						<div>
							{#each categories as category}
								<div>
									<h2>{category.name }</h2>
									{#if category.items.length > 0}
										<ul>
											{#each category.items as item}
												<li>{item.name } - ${item.price }</li>
											{/each}
										</ul>
									{:else}
										<p>No items in this category</p>
									{/if}
								</div>
							{/each}
						</div>
					{:else}
						<p>No categories found</p>
					{/if}
				</div>
			`,
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
			expected: `<div x-data='{"title":"Product Catalog","categories":[{"name":"Electronics","items":[{"name":"Laptop","price":999.99},{"name":"Phone","price":699.99}]},{"name":"Books","items":[]}]}'><div><h1><span x-text="title"></span></h1><template x-if="categories.length > 0"><div><template x-for="category in categories"><div><h2><span x-text="category.name"></span></h2><template x-if="category.items.length > 0"><ul><template x-for="item in category.items"><li><span x-text="item.name"></span> - $<span x-text="item.price"></span></li></template></ul></template><template x-if="!(category.items.length > 0)"><p>No items in this category</p></template></div></template></div></template><template x-if="!(categories.length > 0)"><p>No categories found</p></template></div></div>`,
		},
		{
			name: "dynamic_components_with_conditionals",
			template: `
				<div>
					<h1>{title }</h1>
					{#if isAdmin}
						<AdminPanel user={currentUser} />
					{:else}
						<UserProfile user={currentUser} />
					{/if}
				</div>
			`,
			props: map[string]any{
				"title":       "User Dashboard",
				"isAdmin":     true,
				"currentUser": map[string]any{"name": "John Doe", "role": "admin"},
			},
			expected: `<div x-data='{"title":"User Dashboard","isAdmin":true,"currentUser":{"name":"John Doe","role":"admin"}'><div><h1><span x-text="title"></span></h1><template x-if="isAdmin"><div x-component="AdminPanel" data-prop-user="currentUser"></div></template><template x-if="!isAdmin"><div x-component="UserProfile" data-prop-user="currentUser"></div></template></div></div>`,
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
		// Check if this is a template element with x-for directive
		isForLoop := false
		forExpr := ""

		for _, attr := range n.Attributes {
			if attr.Name == "x-for" {
				isForLoop = true
				forExpr = attr.Value
				break
			}
		}

		if isForLoop && n.TagName == "template" {
			sb.WriteString("<template x-for=\"")
			sb.WriteString(forExpr)
			sb.WriteString("\">")

			// Render children
			for _, child := range n.Children {
				renderIntegrationNode(sb, child)
			}

			sb.WriteString("</template>")
			return
		}

		// Check if this is a template element with x-if directive
		isIfConditional := false
		conditionExpr := ""

		for _, attr := range n.Attributes {
			if attr.Name == "x-if" {
				isIfConditional = true
				conditionExpr = attr.Value
				break
			}
		}

		// Handle conditionals
		if isIfConditional && n.TagName == "template" {
			// Check if this is an else-if condition (contains multiple negations with &&)
			if strings.Contains(conditionExpr, "!(") && strings.Contains(conditionExpr, "&&") {
				// This is an else-if condition
				parts := strings.Split(conditionExpr, "&&")
				if len(parts) > 1 {
					// Get the second part (the actual condition)
					condition := strings.TrimSpace(parts[len(parts)-1])
					// Remove surrounding parentheses if present
					condition = strings.TrimPrefix(condition, "(")
					condition = strings.TrimSuffix(condition, ")")

					sb.WriteString("<template x-if=\"")
					sb.WriteString(condition)
					sb.WriteString("\">")
				} else {
					// Fallback to regular if
					sb.WriteString("<template x-if=\"")
					sb.WriteString(conditionExpr)
					sb.WriteString("\">")
				}
			} else if strings.HasPrefix(conditionExpr, "!(") {
				// This is an else condition (simple negation)
				sb.WriteString("<template x-if=\"!(")

				// Extract the condition inside the negation
				condition := strings.TrimPrefix(conditionExpr, "!(")
				condition = strings.TrimSuffix(condition, ")")

				sb.WriteString(condition)
				sb.WriteString(")\">")
			} else {
				// This is a regular if condition
				sb.WriteString("<template x-if=\"")
				sb.WriteString(conditionExpr)
				sb.WriteString("\">")
			}

			// Render children
			for _, child := range n.Children {
				renderIntegrationNode(sb, child)
			}

			sb.WriteString("</template>")
			return
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
					sb.WriteString("=\"")
					sb.WriteString(attr.Value)
					sb.WriteString("\"")
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
