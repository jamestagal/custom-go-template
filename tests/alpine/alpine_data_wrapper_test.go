package alpine

import (
	"strconv"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/tests/testutils"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestAlpineDataWrapper(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []ast.Node
		fence    string
		props    map[string]any
		expected string
	}{
		{
			name: "Basic Data Wrapper",
			fence: `
let count = 0
let name = "John"
`,
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Element{
							TagName: "h1",
							Children: []ast.Node{
								&ast.TextNode{Content: "Hello, "},
								&ast.ExpressionNode{Expression: "name"},
								&ast.TextNode{Content: "!"},
							},
						},
						&ast.Element{
							TagName: "button",
							Children: []ast.Node{
								&ast.TextNode{Content: "Count: "},
								&ast.ExpressionNode{Expression: "count"},
							},
						},
					},
				},
			},
			props: map[string]any{},
			expected: `<div x-data="{&quot;count&quot;:0,&quot;name&quot;:&quot;John&quot;}">
  <h1>Hello, <span x-text="name"></span>!</h1>
  <button>Count: <span x-text="count"></span></button>
</div>`,
		},
		{
			name: "Data Wrapper with Props",
			fence: `
let count = 0
`,
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Element{
							TagName: "h1",
							Children: []ast.Node{
								&ast.TextNode{Content: "Hello, "},
								&ast.ExpressionNode{Expression: "name"},
								&ast.TextNode{Content: "!"},
							},
						},
						&ast.Element{
							TagName: "button",
							Children: []ast.Node{
								&ast.TextNode{Content: "Count: "},
								&ast.ExpressionNode{Expression: "count"},
							},
						},
					},
				},
			},
			props: map[string]any{
				"name": "Alice",
			},
			expected: `<div x-data="{&quot;count&quot;:0,&quot;name&quot;:&quot;Alice&quot;}">
  <h1>Hello, <span x-text="name"></span>!</h1>
  <button>Count: <span x-text="count"></span></button>
</div>`,
		},
		{
			name: "Complex Data Structure",
			fence: `
let user = { name: "John", age: 30 }
let items = ["apple", "banana", "orange"]
`,
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Element{
							TagName: "h1",
							Children: []ast.Node{
								&ast.TextNode{Content: "User: "},
								&ast.ExpressionNode{Expression: "user.name"},
								&ast.TextNode{Content: ", Age: "},
								&ast.ExpressionNode{Expression: "user.age"},
							},
						},
						&ast.Element{
							TagName: "ul",
							Children: []ast.Node{
								&ast.Element{
									TagName: "li",
									Children: []ast.Node{
										&ast.TextNode{Content: "First item: "},
										&ast.ExpressionNode{Expression: "items[0]"},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{},
			expected: `<div x-data="{&quot;items&quot;:[&quot;apple&quot;,&quot;banana&quot;,&quot;orange&quot;],&quot;user&quot;:{&quot;age&quot;:30,&quot;name&quot;:&quot;John&quot;}}"> <h1>User: <span x-text="user.name"></span>, Age: <span x-text="user.age"></span></h1> <ul><li>First item: <span x-text="items[0]"></span></li></ul> </div>`,
		},
		{
			name: "Function Expressions",
			fence: `
let increment = function() { return count++ }
let count = 0
`,
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Element{
							TagName: "button",
							Children: []ast.Node{
								&ast.TextNode{Content: "Increment"},
							},
						},
						&ast.Element{
							TagName: "p",
							Children: []ast.Node{
								&ast.TextNode{Content: "Count: "},
								&ast.ExpressionNode{Expression: "count"},
							},
						},
					},
				},
			},
			props: map[string]any{},
			expected: `<div x-data="{&quot;count&quot;:0,&quot;increment&quot;:function() { return count++ }}"> <button>Increment</button> <p>Count: <span x-text="count"></span></p> </div>`,
		},
		{
			name: "Nested Variables Detection",
			fence: `
let count = 0
`,
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Conditional{
							IfCondition: "count > 0",
							IfContent: []ast.Node{
								&ast.Element{
									TagName: "p",
									Children: []ast.Node{
										&ast.TextNode{Content: "Count is positive: "},
										&ast.ExpressionNode{Expression: "count"},
									},
								},
							},
							ElseContent: []ast.Node{
								&ast.Element{
									TagName: "p",
									Children: []ast.Node{
										&ast.TextNode{Content: "Count is zero: "},
										&ast.ExpressionNode{Expression: "count"},
									},
								},
								&ast.Conditional{
									IfCondition: "showReset",
									IfContent: []ast.Node{
										&ast.Element{
											TagName: "button",
											Children: []ast.Node{
												&ast.TextNode{Content: "Reset"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"showReset": true,
			},
			expected: `<div x-data="{&quot;count&quot;:0,&quot;showReset&quot;:true}"> <template x-if="count > 0"><p>Count is positive: <span x-text="count"></span></p></template><template x-else><p>Count is zero: <span x-text="count"></span></p><template x-if="showReset"><button>Reset</button></template></template> </div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a data scope from the fence content
			dataScope := extractDataFromFence(tt.fence)

			// Add props to the data scope
			for k, v := range tt.props {
				dataScope[k] = v
			}

			// Transform the nodes with the Alpine.js data wrapper
			result := transformer.TransformWithAlpineData(tt.nodes, dataScope)

			// Render the transformed nodes to HTML
			var html string
			if len(result) == 1 {
				// If there's only one node, render it directly
				html = testutils.RenderNode(result[0])
			} else {
				// If there are multiple nodes, create a temporary wrapper
				wrapper := &ast.Element{
					TagName:     "div",
					Attributes:  []ast.Attribute{},
					Children:    result,
					SelfClosing: false,
				}
				html = testutils.RenderNode(wrapper)
				// Remove the wrapper div tags
				html = strings.TrimPrefix(html, "<div>")
				html = strings.TrimSuffix(html, "</div>")
			}

			// Normalize both expected and actual outputs before comparison
			expectedNormalized := testutils.NormalizeWhitespace(tt.expected)
			htmlNormalized := testutils.NormalizeWhitespace(html)

			if htmlNormalized != expectedNormalized {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", expectedNormalized, htmlNormalized)
			}
		})
	}
}

// extractDataFromFence parses the fence content to extract variable declarations
func extractDataFromFence(fence string) map[string]any {
	dataScope := make(map[string]any)

	lines := strings.Split(fence, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "let ") {
			// Extract variable name and value
			declaration := strings.TrimPrefix(line, "let ")
			eqIndex := strings.Index(declaration, "=")

			if eqIndex > 0 {
				varName := strings.TrimSpace(declaration[:eqIndex])
				varValue := strings.TrimSpace(declaration[eqIndex+1:])

				// Handle function expressions
				if strings.Contains(varValue, "function") {
					dataScope[varName] = varValue
				} else {
					// Handle primitive values
					if varValue == "0" {
						dataScope[varName] = 0
					} else if varValue == "true" {
						dataScope[varName] = true
					} else if varValue == "false" {
						dataScope[varName] = false
					} else if strings.HasPrefix(varValue, "\"") && strings.HasSuffix(varValue, "\"") {
						// String value
						dataScope[varName] = strings.Trim(varValue, "\"")
					} else if strings.HasPrefix(varValue, "{") && strings.HasSuffix(varValue, "}") {
						// Object value - simple parsing
						dataScope[varName] = parseSimpleObject(varValue)
					} else if strings.HasPrefix(varValue, "[") && strings.HasSuffix(varValue, "]") {
						// Array value - simple parsing
						dataScope[varName] = parseSimpleArray(varValue)
					} else {
						// Default to string value
						dataScope[varName] = varValue
					}
				}
			}
		}
	}

	return dataScope
}

// parseSimpleObject parses a simple JavaScript object into a map
func parseSimpleObject(objStr string) map[string]any {
	result := make(map[string]any)

	// Remove the outer braces
	content := strings.TrimSpace(objStr[1 : len(objStr)-1])

	// Split by commas, but respect nested objects
	pairs := strings.Split(content, ",")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		colonIndex := strings.Index(pair, ":")

		if colonIndex > 0 {
			key := strings.TrimSpace(pair[:colonIndex])
			// Remove quotes from key if present
			key = strings.Trim(key, "\"'")

			value := strings.TrimSpace(pair[colonIndex+1:])

			// Parse the value based on its type
			if value == "true" {
				result[key] = true
			} else if value == "false" {
				result[key] = false
			} else if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
				// Try parsing as int
				result[key] = intVal
			} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				// Try parsing as float
				result[key] = floatVal
			} else if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				// String with double quotes
				result[key] = strings.Trim(value, "\"")
			} else if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				// String with single quotes
				result[key] = strings.Trim(value, "'")
			} else if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
				// Nested object
				result[key] = parseSimpleObject(value)
			} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				// Nested array
				result[key] = parseSimpleArray(value)
			} else {
				// Default to string
				result[key] = value
			}
		}
	}

	return result
}

// parseSimpleArray parses a simple JavaScript array into a slice
func parseSimpleArray(arrStr string) []any {
	var result []any

	// Remove the outer brackets
	content := strings.TrimSpace(arrStr[1 : len(arrStr)-1])

	// Split by commas
	elements := strings.Split(content, ",")

	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}

		// Parse the element based on its type
		if elem == "true" {
			result = append(result, true)
		} else if elem == "false" {
			result = append(result, false)
		} else if intVal, err := strconv.ParseInt(elem, 10, 64); err == nil {
			// Try parsing as int
			result = append(result, intVal)
		} else if floatVal, err := strconv.ParseFloat(elem, 64); err == nil {
			// Try parsing as float
			result = append(result, floatVal)
		} else if strings.HasPrefix(elem, "\"") && strings.HasSuffix(elem, "\"") {
			// String with double quotes
			result = append(result, strings.Trim(elem, "\""))
		} else if strings.HasPrefix(elem, "'") && strings.HasSuffix(elem, "'") {
			// String with single quotes
			result = append(result, strings.Trim(elem, "'"))
		} else if strings.HasPrefix(elem, "{") && strings.HasSuffix(elem, "}") {
			// Nested object
			result = append(result, parseSimpleObject(elem))
		} else if strings.HasPrefix(elem, "[") && strings.HasSuffix(elem, "]") {
			// Nested array
			result = append(result, parseSimpleArray(elem))
		} else {
			// Default to string
			result = append(result, elem)
		}
	}

	return result
}
