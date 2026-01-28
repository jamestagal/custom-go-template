package ast

import (
	"testing"
)

// TestDynamicComponentByNameNode_NodeType tests the NodeType method
func TestDynamicComponentByNameNode_NodeType(t *testing.T) {
	tests := []struct {
		name     string
		node     *DynamicComponentByNameNode
		expected string
	}{
		{
			name: "basic node",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props:          []ComponentProp{},
				SelfClosing:    true,
			},
			expected: "DynamicComponentByName",
		},
		{
			name: "with props",
			node: &DynamicComponentByNameNode{
				NameExpression: "comp",
				Props: []ComponentProp{
					{Name: "title", Value: "Hello", IsDynamic: false},
				},
				SelfClosing: true,
			},
			expected: "DynamicComponentByName",
		},
		{
			name: "with spread props",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"component.fields"},
				SelfClosing:    true,
			},
			expected: "DynamicComponentByName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.NodeType()
			if result != tt.expected {
				t.Errorf("NodeType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestDynamicComponentByNameNode_Creation tests creating various node configurations
func TestDynamicComponentByNameNode_Creation(t *testing.T) {
	tests := []struct {
		name              string
		nameExpression    string
		props             []ComponentProp
		spreadProps       []string
		selfClosing       bool
		expectedNameExpr  string
		expectedPropsLen  int
		expectedSpreadLen int
	}{
		{
			name:              "name expression only",
			nameExpression:    "component.name",
			props:             []ComponentProp{},
			spreadProps:       []string{},
			selfClosing:       true,
			expectedNameExpr:  "component.name",
			expectedPropsLen:  0,
			expectedSpreadLen: 0,
		},
		{
			name:           "with regular props",
			nameExpression: "compName",
			props: []ComponentProp{
				{Name: "title", Value: "Welcome", IsDynamic: false},
				{Name: "link", Value: "/about", IsDynamic: false},
			},
			spreadProps:       []string{},
			selfClosing:       true,
			expectedNameExpr:  "compName",
			expectedPropsLen:  2,
			expectedSpreadLen: 0,
		},
		{
			name:              "with spread props",
			nameExpression:    "component.name",
			props:             []ComponentProp{},
			spreadProps:       []string{"component.fields", "extraProps"},
			selfClosing:       true,
			expectedNameExpr:  "component.name",
			expectedPropsLen:  0,
			expectedSpreadLen: 2,
		},
		{
			name:           "mixed props and spread",
			nameExpression: "component.name",
			props: []ComponentProp{
				{Name: "allContent", Value: "allContent", IsDynamic: true},
				{Name: "theme", Value: "dark", IsDynamic: false},
			},
			spreadProps:       []string{"component.fields"},
			selfClosing:       true,
			expectedNameExpr:  "component.name",
			expectedPropsLen:  2,
			expectedSpreadLen: 1,
		},
		{
			name:              "non-self-closing",
			nameExpression:    "MyComponent",
			props:             []ComponentProp{},
			spreadProps:       []string{},
			selfClosing:       false,
			expectedNameExpr:  "MyComponent",
			expectedPropsLen:  0,
			expectedSpreadLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &DynamicComponentByNameNode{
				NameExpression: tt.nameExpression,
				Props:          tt.props,
				SpreadProps:    tt.spreadProps,
				SelfClosing:    tt.selfClosing,
			}

			if node.NameExpression != tt.expectedNameExpr {
				t.Errorf("NameExpression = %q, want %q", node.NameExpression, tt.expectedNameExpr)
			}
			if len(node.Props) != tt.expectedPropsLen {
				t.Errorf("Props length = %d, want %d", len(node.Props), tt.expectedPropsLen)
			}
			if len(node.SpreadProps) != tt.expectedSpreadLen {
				t.Errorf("SpreadProps length = %d, want %d", len(node.SpreadProps), tt.expectedSpreadLen)
			}
			if node.SelfClosing != tt.selfClosing {
				t.Errorf("SelfClosing = %v, want %v", node.SelfClosing, tt.selfClosing)
			}
		})
	}
}

// TestDynamicComponentByNameNode_WithDynamicProps tests nodes with dynamic prop values
func TestDynamicComponentByNameNode_WithDynamicProps(t *testing.T) {
	tests := []struct {
		name       string
		node       *DynamicComponentByNameNode
		expectFunc func(t *testing.T, node *DynamicComponentByNameNode)
	}{
		{
			name: "dynamic prop expression",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props: []ComponentProp{
					{Name: "title", Value: "content.title", IsDynamic: true},
					{Name: "link", Value: "/static", IsDynamic: false},
				},
				SelfClosing: true,
			},
			expectFunc: func(t *testing.T, node *DynamicComponentByNameNode) {
				if len(node.Props) != 2 {
					t.Errorf("Expected 2 props, got %d", len(node.Props))
				}
				if !node.Props[0].IsDynamic {
					t.Error("First prop should be dynamic")
				}
				if node.Props[1].IsDynamic {
					t.Error("Second prop should not be dynamic")
				}
			},
		},
		{
			name: "all dynamic props",
			node: &DynamicComponentByNameNode{
				NameExpression: "comp",
				Props: []ComponentProp{
					{Name: "title", Value: "data.title", IsDynamic: true},
					{Name: "desc", Value: "data.description", IsDynamic: true},
					{Name: "link", Value: "data.url", IsDynamic: true},
				},
				SelfClosing: true,
			},
			expectFunc: func(t *testing.T, node *DynamicComponentByNameNode) {
				for i, prop := range node.Props {
					if !prop.IsDynamic {
						t.Errorf("Prop %d should be dynamic", i)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expectFunc(t, tt.node)
		})
	}
}

// TestDynamicComponentByNameNode_WithSpreadProps tests spread prop variations
func TestDynamicComponentByNameNode_WithSpreadProps(t *testing.T) {
	tests := []struct {
		name          string
		node          *DynamicComponentByNameNode
		expectedLen   int
		expectedFirst string
	}{
		{
			name: "single spread",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"component.fields"},
				SelfClosing:    true,
			},
			expectedLen:   1,
			expectedFirst: "component.fields",
		},
		{
			name: "multiple spreads",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"defaults", "component.fields", "overrides"},
				SelfClosing:    true,
			},
			expectedLen:   3,
			expectedFirst: "defaults",
		},
		{
			name: "nested property spread",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"data.component.props"},
				SelfClosing:    true,
			},
			expectedLen:   1,
			expectedFirst: "data.component.props",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.node.SpreadProps) != tt.expectedLen {
				t.Errorf("SpreadProps length = %d, want %d", len(tt.node.SpreadProps), tt.expectedLen)
			}
			if tt.expectedLen > 0 && tt.node.SpreadProps[0] != tt.expectedFirst {
				t.Errorf("First spread = %q, want %q", tt.node.SpreadProps[0], tt.expectedFirst)
			}
		})
	}
}

// TestDynamicComponentByNameNode_MixedPropsAndSpread tests combinations
func TestDynamicComponentByNameNode_MixedPropsAndSpread(t *testing.T) {
	tests := []struct {
		name                string
		node                *DynamicComponentByNameNode
		expectedPropsCount  int
		expectedSpreadCount int
		description         string
	}{
		{
			name: "spread before regular props",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"component.fields"},
				Props: []ComponentProp{
					{Name: "allContent", Value: "allContent", IsDynamic: true},
					{Name: "content", Value: "content", IsDynamic: true},
				},
				SelfClosing: true,
			},
			expectedPropsCount:  2,
			expectedSpreadCount: 1,
			description:         "Spread props should be processed before regular props",
		},
		{
			name: "multiple spreads with regular props",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"defaults", "component.fields"},
				Props: []ComponentProp{
					{Name: "theme", Value: "dark", IsDynamic: false},
					{Name: "allContent", Value: "allContent", IsDynamic: true},
				},
				SelfClosing: true,
			},
			expectedPropsCount:  2,
			expectedSpreadCount: 2,
			description:         "Multiple spreads with override props",
		},
		{
			name: "interleaved ordering test",
			node: &DynamicComponentByNameNode{
				NameExpression: "MyComp",
				SpreadProps:    []string{"props1", "props2", "props3"},
				Props: []ComponentProp{
					{Name: "a", Value: "1", IsDynamic: false},
					{Name: "b", Value: "2", IsDynamic: false},
					{Name: "c", Value: "var", IsDynamic: true},
				},
				SelfClosing: true,
			},
			expectedPropsCount:  3,
			expectedSpreadCount: 3,
			description:         "Mix of multiple spreads and multiple props",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.node.Props) != tt.expectedPropsCount {
				t.Errorf("Props count = %d, want %d", len(tt.node.Props), tt.expectedPropsCount)
			}
			if len(tt.node.SpreadProps) != tt.expectedSpreadCount {
				t.Errorf("SpreadProps count = %d, want %d", len(tt.node.SpreadProps), tt.expectedSpreadCount)
			}
		})
	}
}

// TestDynamicComponentByNameNode_SelfClosing tests self-closing variations
func TestDynamicComponentByNameNode_SelfClosing(t *testing.T) {
	tests := []struct {
		name        string
		selfClosing bool
	}{
		{
			name:        "self-closing",
			selfClosing: true,
		},
		{
			name:        "non-self-closing",
			selfClosing: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &DynamicComponentByNameNode{
				NameExpression: "TestComponent",
				SelfClosing:    tt.selfClosing,
			}

			if node.SelfClosing != tt.selfClosing {
				t.Errorf("SelfClosing = %v, want %v", node.SelfClosing, tt.selfClosing)
			}
		})
	}
}

// TestDynamicComponentByNameNode_NodeInterface tests that DynamicComponentByNameNode implements Node interface
func TestDynamicComponentByNameNode_NodeInterface(t *testing.T) {
	// This test ensures DynamicComponentByNameNode implements the Node interface
	var _ Node = (*DynamicComponentByNameNode)(nil)
}

// TestDynamicComponentByNameNode_RealWorldScenarios tests practical use cases
func TestDynamicComponentByNameNode_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name     string
		node     *DynamicComponentByNameNode
		scenario string
	}{
		{
			name: "Plenti page component iteration",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"component.fields"},
				Props: []ComponentProp{
					{Name: "allContent", Value: "allContent", IsDynamic: true},
					{Name: "content", Value: "content", IsDynamic: true},
				},
				SelfClosing: true,
			},
			scenario: "Typical Plenti page layout with magic variables",
		},
		{
			name: "Blog post component with overrides",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				SpreadProps:    []string{"component.fields"},
				Props: []ComponentProp{
					{Name: "theme", Value: "dark", IsDynamic: false},
					{Name: "author", Value: "content.author", IsDynamic: true},
				},
				SelfClosing: true,
			},
			scenario: "Blog layout with theme override and author injection",
		},
		{
			name: "Hero component with defaults then overrides",
			node: &DynamicComponentByNameNode{
				NameExpression: "Hero2436",
				SpreadProps:    []string{"defaults", "component.fields"},
				Props: []ComponentProp{
					{Name: "featured", Value: "true", IsDynamic: false},
				},
				SelfClosing: true,
			},
			scenario: "Hero component with default props, content fields, then featured override",
		},
		{
			name: "Static component name with dynamic props",
			node: &DynamicComponentByNameNode{
				NameExpression: "ContactForm",
				SpreadProps:    []string{"component.fields"},
				Props: []ComponentProp{
					{Name: "submitHandler", Value: "handleSubmit", IsDynamic: true},
				},
				SelfClosing: true,
			},
			scenario: "Static component name with spread and custom handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify node is well-formed
			if tt.node.NameExpression == "" {
				t.Error("NameExpression should not be empty")
			}
			if tt.node.NodeType() != "DynamicComponentByName" {
				t.Errorf("NodeType() = %q, want %q", tt.node.NodeType(), "DynamicComponentByName")
			}

			t.Logf("Scenario: %s", tt.scenario)
			t.Logf("  Name: %s", tt.node.NameExpression)
			t.Logf("  Spreads: %d", len(tt.node.SpreadProps))
			t.Logf("  Props: %d", len(tt.node.Props))
		})
	}
}

// TestDynamicComponentByNameNode_EdgeCases tests edge cases and boundary conditions
func TestDynamicComponentByNameNode_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		node     *DynamicComponentByNameNode
		testFunc func(t *testing.T, node *DynamicComponentByNameNode)
	}{
		{
			name: "empty name expression",
			node: &DynamicComponentByNameNode{
				NameExpression: "",
				SelfClosing:    true,
			},
			testFunc: func(t *testing.T, node *DynamicComponentByNameNode) {
				// Should still be valid node, validation happens elsewhere
				if node.NodeType() != "DynamicComponentByName" {
					t.Error("Should return correct NodeType even with empty name")
				}
			},
		},
		{
			name: "no props or spreads",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props:          []ComponentProp{},
				SpreadProps:    []string{},
				SelfClosing:    true,
			},
			testFunc: func(t *testing.T, node *DynamicComponentByNameNode) {
				if len(node.Props) != 0 {
					t.Error("Props should be empty")
				}
				if len(node.SpreadProps) != 0 {
					t.Error("SpreadProps should be empty")
				}
			},
		},
		{
			name: "nil props slice",
			node: &DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props:          nil,
				SpreadProps:    nil,
				SelfClosing:    true,
			},
			testFunc: func(t *testing.T, node *DynamicComponentByNameNode) {
				// Nil slices are valid in Go, should work correctly
				if node.Props != nil && len(node.Props) != 0 {
					t.Error("Nil props should behave like empty slice")
				}
			},
		},
		{
			name: "complex nested expression",
			node: &DynamicComponentByNameNode{
				NameExpression: "data.components[0].metadata.componentName",
				SpreadProps:    []string{"data.components[0].props"},
				SelfClosing:    true,
			},
			testFunc: func(t *testing.T, node *DynamicComponentByNameNode) {
				if node.NameExpression == "" {
					t.Error("Complex expression should be preserved")
				}
				if len(node.SpreadProps) != 1 {
					t.Error("Spread with complex expression should work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t, tt.node)
		})
	}
}
