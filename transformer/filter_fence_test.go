package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestFilterOutFence verifies that filterOutFence() correctly removes FenceSection nodes
// while preserving all other node types in their original order.
func TestFilterOutFence(t *testing.T) {
	tests := []struct {
		name     string
		input    []ast.Node
		expected []ast.Node
	}{
		{
			name:     "empty slice",
			input:    []ast.Node{},
			expected: []ast.Node{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []ast.Node{},
		},
		{
			name: "only fence section",
			input: []ast.Node{
				&ast.FenceSection{
					RawContent: "let x = 1;",
				},
			},
			expected: []ast.Node{},
		},
		{
			name: "multiple fence sections only",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.FenceSection{RawContent: "let y = 2;"},
				&ast.FenceSection{RawContent: "const z = 3;"},
			},
			expected: []ast.Node{},
		},
		{
			name: "no fence sections",
			input: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
				&ast.ExpressionNode{Expression: "name"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
				&ast.ExpressionNode{Expression: "name"},
			},
		},
		{
			name: "fence at beginning",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
			},
		},
		{
			name: "fence at end",
			input: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
				&ast.FenceSection{RawContent: "let x = 1;"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
			},
		},
		{
			name: "fence in middle",
			input: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.TextNode{Content: "Hello"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
			},
		},
		{
			name: "multiple fences scattered",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.Element{TagName: "div"},
				&ast.FenceSection{RawContent: "let y = 2;"},
				&ast.TextNode{Content: "Hello"},
				&ast.FenceSection{RawContent: "const z = 3;"},
				&ast.ExpressionNode{Expression: "name"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Hello"},
				&ast.ExpressionNode{Expression: "name"},
			},
		},
		{
			name: "preserves element nodes",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{Name: "class", Value: "container"},
					},
				},
				&ast.Element{
					TagName: "span",
					Children: []ast.Node{
						&ast.TextNode{Content: "nested"},
					},
				},
			},
			expected: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{Name: "class", Value: "container"},
					},
				},
				&ast.Element{
					TagName: "span",
					Children: []ast.Node{
						&ast.TextNode{Content: "nested"},
					},
				},
			},
		},
		{
			name: "preserves text nodes",
			input: []ast.Node{
				&ast.TextNode{Content: "First"},
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.TextNode{Content: "Second"},
				&ast.TextNode{Content: "Third"},
			},
			expected: []ast.Node{
				&ast.TextNode{Content: "First"},
				&ast.TextNode{Content: "Second"},
				&ast.TextNode{Content: "Third"},
			},
		},
		{
			name: "preserves conditional nodes",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.Conditional{
					IfCondition: "user.isActive",
					IfContent: []ast.Node{
						&ast.TextNode{Content: "Active"},
					},
				},
			},
			expected: []ast.Node{
				&ast.Conditional{
					IfCondition: "user.isActive",
					IfContent: []ast.Node{
						&ast.TextNode{Content: "Active"},
					},
				},
			},
		},
		{
			name: "preserves loop nodes",
			input: []ast.Node{
				&ast.Loop{
					Iterator:   "item",
					Collection: "items",
					Content: []ast.Node{
						&ast.Element{TagName: "li"},
					},
				},
				&ast.FenceSection{RawContent: "let x = 1;"},
			},
			expected: []ast.Node{
				&ast.Loop{
					Iterator:   "item",
					Collection: "items",
					Content: []ast.Node{
						&ast.Element{TagName: "li"},
					},
				},
			},
		},
		{
			name: "preserves expression nodes",
			input: []ast.Node{
				&ast.ExpressionNode{Expression: "user.name"},
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.ExpressionNode{Expression: "user.email"},
			},
			expected: []ast.Node{
				&ast.ExpressionNode{Expression: "user.name"},
				&ast.ExpressionNode{Expression: "user.email"},
			},
		},
		{
			name: "preserves component nodes",
			input: []ast.Node{
				&ast.ComponentNode{
					Name: "Header",
					Props: []ast.ComponentProp{
						{Name: "title", Value: "Welcome"},
					},
				},
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.ComponentNode{
					Name: "Footer",
				},
			},
			expected: []ast.Node{
				&ast.ComponentNode{
					Name: "Header",
					Props: []ast.ComponentProp{
						{Name: "title", Value: "Welcome"},
					},
				},
				&ast.ComponentNode{
					Name: "Footer",
				},
			},
		},
		{
			name: "mixed node types with complex structure",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "let count = 0;"},
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.TextNode{Content: "Container"},
					},
				},
				&ast.FenceSection{RawContent: "const name = 'Test';"},
				&ast.Conditional{
					IfCondition: "showContent",
					IfContent: []ast.Node{
						&ast.ExpressionNode{Expression: "content"},
					},
					ElseContent: []ast.Node{
						&ast.TextNode{Content: "No content"},
					},
				},
				&ast.Loop{
					Iterator:   "item",
					Collection: "items",
					Content: []ast.Node{
						&ast.ComponentNode{Name: "Item"},
					},
				},
				&ast.FenceSection{RawContent: "function handler() {}"},
				&ast.TextNode{Content: "End"},
			},
			expected: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.TextNode{Content: "Container"},
					},
				},
				&ast.Conditional{
					IfCondition: "showContent",
					IfContent: []ast.Node{
						&ast.ExpressionNode{Expression: "content"},
					},
					ElseContent: []ast.Node{
						&ast.TextNode{Content: "No content"},
					},
				},
				&ast.Loop{
					Iterator:   "item",
					Collection: "items",
					Content: []ast.Node{
						&ast.ComponentNode{Name: "Item"},
					},
				},
				&ast.TextNode{Content: "End"},
			},
		},
		{
			name: "preserves order with identical node types",
			input: []ast.Node{
				&ast.TextNode{Content: "First"},
				&ast.FenceSection{RawContent: "let x = 1;"},
				&ast.TextNode{Content: "Second"},
				&ast.TextNode{Content: "Third"},
				&ast.FenceSection{RawContent: "let y = 2;"},
				&ast.TextNode{Content: "Fourth"},
			},
			expected: []ast.Node{
				&ast.TextNode{Content: "First"},
				&ast.TextNode{Content: "Second"},
				&ast.TextNode{Content: "Third"},
				&ast.TextNode{Content: "Fourth"},
			},
		},
		{
			name: "empty fence section",
			input: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.FenceSection{},
				&ast.TextNode{Content: "Text"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "div"},
				&ast.TextNode{Content: "Text"},
			},
		},
		{
			name: "fence with complex content",
			input: []ast.Node{
				&ast.FenceSection{
					Imports: []ast.ImportNode{
						{Name: "Header", Path: "./Header.html"},
					},
					Props: []ast.PropNode{
						{Name: "title", DefaultValue: "Default"},
					},
					Variables: []ast.VariableNode{
						{Keyword: "let", Name: "count", Value: "0"},
					},
					RawContent: "function handler() { console.log('test'); }",
				},
				&ast.Element{TagName: "main"},
			},
			expected: []ast.Node{
				&ast.Element{TagName: "main"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of input to verify it's not modified
			originalInput := make([]ast.Node, len(tt.input))
			copy(originalInput, tt.input)

			// Call function
			result := filterOutFence(tt.input)

			// Verify result length
			if len(result) != len(tt.expected) {
				t.Errorf("filterOutFence() returned %d nodes, expected %d", len(result), len(tt.expected))
			}

			// Verify each node matches expected
			for i := 0; i < len(result) && i < len(tt.expected); i++ {
				if result[i].NodeType() != tt.expected[i].NodeType() {
					t.Errorf("filterOutFence() node[%d] type = %s, expected %s",
						i, result[i].NodeType(), tt.expected[i].NodeType())
				}

				// Type-specific comparisons
				switch expected := tt.expected[i].(type) {
				case *ast.Element:
					if actual, ok := result[i].(*ast.Element); ok {
						if actual.TagName != expected.TagName {
							t.Errorf("filterOutFence() node[%d] Element.TagName = %s, expected %s",
								i, actual.TagName, expected.TagName)
						}
					}
				case *ast.TextNode:
					if actual, ok := result[i].(*ast.TextNode); ok {
						if actual.Content != expected.Content {
							t.Errorf("filterOutFence() node[%d] TextNode.Content = %q, expected %q",
								i, actual.Content, expected.Content)
						}
					}
				case *ast.ExpressionNode:
					if actual, ok := result[i].(*ast.ExpressionNode); ok {
						if actual.Expression != expected.Expression {
							t.Errorf("filterOutFence() node[%d] ExpressionNode.Expression = %s, expected %s",
								i, actual.Expression, expected.Expression)
						}
					}
				case *ast.ComponentNode:
					if actual, ok := result[i].(*ast.ComponentNode); ok {
						if actual.Name != expected.Name {
							t.Errorf("filterOutFence() node[%d] ComponentNode.Name = %s, expected %s",
								i, actual.Name, expected.Name)
						}
					}
				case *ast.Conditional:
					if actual, ok := result[i].(*ast.Conditional); ok {
						if actual.IfCondition != expected.IfCondition {
							t.Errorf("filterOutFence() node[%d] Conditional.IfCondition = %s, expected %s",
								i, actual.IfCondition, expected.IfCondition)
						}
					}
				case *ast.Loop:
					if actual, ok := result[i].(*ast.Loop); ok {
						if actual.Iterator != expected.Iterator {
							t.Errorf("filterOutFence() node[%d] Loop.Iterator = %s, expected %s",
								i, actual.Iterator, expected.Iterator)
						}
					}
				}
			}

			// Verify original input was not modified
			if len(tt.input) > 0 {
				inputModified := false
				if len(originalInput) != len(tt.input) {
					inputModified = true
				} else {
					for i := 0; i < len(originalInput); i++ {
						if originalInput[i] != tt.input[i] {
							inputModified = true
							break
						}
					}
				}
				if inputModified {
					t.Errorf("filterOutFence() modified the original input slice")
				}
			}

			// Verify no FenceSection nodes in result
			for i, node := range result {
				if _, isFence := node.(*ast.FenceSection); isFence {
					t.Errorf("filterOutFence() result[%d] is FenceSection, should be filtered out", i)
				}
			}
		})
	}
}

// TestFilterOutFence_PreservesNonFenceNodePointers verifies that filtering
// doesn't create new instances of non-fence nodes, but preserves the original pointers.
func TestFilterOutFence_PreservesNonFenceNodePointers(t *testing.T) {
	elem := &ast.Element{TagName: "div"}
	text := &ast.TextNode{Content: "Hello"}
	expr := &ast.ExpressionNode{Expression: "name"}
	fence := &ast.FenceSection{RawContent: "let x = 1;"}

	input := []ast.Node{elem, fence, text, expr}
	result := filterOutFence(input)

	// Verify exact pointer equality for non-fence nodes
	if len(result) != 3 {
		t.Fatalf("Expected 3 nodes after filtering, got %d", len(result))
	}

	if result[0] != elem {
		t.Error("Element pointer not preserved")
	}
	if result[1] != text {
		t.Error("TextNode pointer not preserved")
	}
	if result[2] != expr {
		t.Error("ExpressionNode pointer not preserved")
	}
}

// TestFilterOutFence_EmptySliceAllocation verifies that the function
// handles edge cases around slice allocation correctly.
func TestFilterOutFence_EmptySliceAllocation(t *testing.T) {
	tests := []struct {
		name  string
		input []ast.Node
	}{
		{
			name:  "nil input",
			input: nil,
		},
		{
			name:  "empty slice",
			input: []ast.Node{},
		},
		{
			name: "single fence",
			input: []ast.Node{
				&ast.FenceSection{RawContent: "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterOutFence(tt.input)

			// Result should not be nil
			if result == nil {
				t.Error("filterOutFence() returned nil, expected empty slice")
			}

			// Result should be empty
			if len(result) != 0 {
				t.Errorf("filterOutFence() returned %d nodes, expected 0", len(result))
			}
		})
	}
}
