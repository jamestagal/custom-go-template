package transformer

import (
	"reflect"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestAddComponentDataWrapper tests the addComponentDataWrapper helper function that wraps
// component nodes with an x-data attribute containing the component's data scope.
//
// Function signature: func addComponentDataWrapper(nodes []ast.Node, dataScope map[string]any) []ast.Node
//
// Requirements:
// 1. Single Root Element: If len(nodes) == 1 and it's an *ast.Element, add x-data attribute directly to it
// 2. Multiple Root Nodes: Wrap all nodes in a new <div> with x-data attribute
// 3. Format Data Scope: Use alpineDataFormatter(dataScope) to format the x-data value
// 4. Handle Empty Scope: Still add wrapper for consistency even if dataScope is empty
// 5. Handle Empty Nodes: Return empty slice if nodes is empty
// 6. Return Slice: Always return []ast.Node (single element or wrapper)
func TestAddComponentDataWrapper(t *testing.T) {
	tests := []struct {
		name           string
		nodes          []ast.Node
		dataScope      map[string]any
		expectedLength int
		checkFunc      func(*testing.T, []ast.Node, map[string]any)
		description    string
	}{
		// ========== Single Root Element Tests ==========
		{
			name: "single element with no attributes - add x-data",
			nodes: []ast.Node{
				&ast.Element{
					TagName:    "div",
					Attributes: []ast.Attribute{},
					Children:   []ast.Node{},
				},
			},
			dataScope: map[string]any{
				"count": 0,
				"name":  "test",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				if elem.TagName != "div" {
					t.Errorf("Expected tag 'div', got %q", elem.TagName)
				}
				// Check that x-data attribute was added
				hasXData := false
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
						if attr.Value == "" {
							t.Error("x-data attribute value should not be empty")
						}
						if !attr.IsAlpine {
							t.Error("x-data attribute should have IsAlpine=true")
						}
						if attr.AlpineType != "data" {
							t.Errorf("x-data attribute should have AlpineType='data', got %q", attr.AlpineType)
						}
						if !attr.Dynamic {
							t.Error("x-data attribute should have Dynamic=true")
						}
					}
				}
				if !hasXData {
					t.Error("x-data attribute was not added to single root element")
				}
			},
			description: "Single root element should have x-data added directly",
		},
		{
			name: "single element with existing attributes - preserve them",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "section",
					Attributes: []ast.Attribute{
						{Name: "id", Value: "main"},
						{Name: "class", Value: "container"},
					},
					Children: []ast.Node{},
				},
			},
			dataScope: map[string]any{
				"title": "Test",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				// Should have 3 attributes: id, class, x-data
				if len(elem.Attributes) != 3 {
					t.Errorf("Expected 3 attributes, got %d: %v", len(elem.Attributes), elem.Attributes)
				}
				// Verify original attributes preserved
				hasId := false
				hasClass := false
				hasXData := false
				for _, attr := range elem.Attributes {
					if attr.Name == "id" && attr.Value == "main" {
						hasId = true
					}
					if attr.Name == "class" && attr.Value == "container" {
						hasClass = true
					}
					if attr.Name == "x-data" {
						hasXData = true
					}
				}
				if !hasId || !hasClass || !hasXData {
					t.Errorf("Missing expected attributes: hasId=%v, hasClass=%v, hasXData=%v",
						hasId, hasClass, hasXData)
				}
			},
			description: "Existing attributes should be preserved when adding x-data",
		},
		{
			name: "single element with children - preserve children",
			nodes: []ast.Node{
				&ast.Element{
					TagName:    "div",
					Attributes: []ast.Attribute{},
					Children: []ast.Node{
						&ast.Element{TagName: "h1", Children: []ast.Node{}},
						&ast.TextNode{Content: "Hello"},
						&ast.Element{TagName: "p", Children: []ast.Node{}},
					},
				},
			},
			dataScope: map[string]any{
				"message": "test",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				// Children should be preserved
				if len(elem.Children) != 3 {
					t.Errorf("Expected 3 children, got %d", len(elem.Children))
				}
				// Verify children types
				if _, ok := elem.Children[0].(*ast.Element); !ok {
					t.Error("First child should be Element")
				}
				if _, ok := elem.Children[1].(*ast.TextNode); !ok {
					t.Error("Second child should be TextNode")
				}
				if _, ok := elem.Children[2].(*ast.Element); !ok {
					t.Error("Third child should be Element")
				}
			},
			description: "Children should be preserved when adding x-data to single element",
		},
		{
			name: "single element - different tag names",
			nodes: []ast.Node{
				&ast.Element{
					TagName:    "article",
					Attributes: []ast.Attribute{},
					Children:   []ast.Node{},
				},
			},
			dataScope: map[string]any{
				"data": "value",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				if elem.TagName != "article" {
					t.Errorf("Tag name should be preserved, got %q", elem.TagName)
				}
				// Check x-data added
				hasXData := false
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
					}
				}
				if !hasXData {
					t.Error("x-data should be added regardless of tag name")
				}
			},
			description: "x-data should be added to any tag type",
		},
		{
			name: "single element - already has x-data attribute",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:       "x-data",
							Value:      `{ existing: true }`,
							IsAlpine:   true,
							AlpineType: "data",
						},
					},
					Children: []ast.Node{},
				},
			},
			dataScope: map[string]any{
				"count": 0,
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				// Count x-data attributes
				xDataCount := 0
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						xDataCount++
					}
				}
				// Should not add duplicate x-data (behavior depends on implementation)
				// This test documents the actual behavior
				t.Logf("Element with existing x-data has %d x-data attributes", xDataCount)
			},
			description: "Behavior when element already has x-data (documents implementation)",
		},

		// ========== Multiple Root Nodes Tests ==========
		{
			name: "two elements - wrap in div",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
				&ast.Element{TagName: "section", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"count": 0,
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				// Should return a single wrapper element
				wrapper, ok := nodes[0].(*ast.Element)
				if !ok {
					t.Fatal("Expected wrapper to be an Element")
				}
				if wrapper.TagName != "div" {
					t.Errorf("Wrapper should be 'div', got %q", wrapper.TagName)
				}
				// Check wrapper has x-data
				hasXData := false
				for _, attr := range wrapper.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
					}
				}
				if !hasXData {
					t.Error("Wrapper should have x-data attribute")
				}
				// Check wrapper contains original nodes
				if len(wrapper.Children) != 2 {
					t.Errorf("Wrapper should have 2 children, got %d", len(wrapper.Children))
				}
			},
			description: "Multiple elements should be wrapped in div with x-data",
		},
		{
			name: "mix of elements and text nodes - wrap all",
			nodes: []ast.Node{
				&ast.Element{TagName: "h1", Children: []ast.Node{}},
				&ast.TextNode{Content: "Hello"},
				&ast.Element{TagName: "p", Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"message": "test",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				wrapper, ok := nodes[0].(*ast.Element)
				if !ok {
					t.Fatal("Expected wrapper to be an Element")
				}
				// Should wrap all 3 nodes
				if len(wrapper.Children) != 3 {
					t.Errorf("Wrapper should have 3 children, got %d", len(wrapper.Children))
				}
				// Verify child types preserved
				if _, ok := wrapper.Children[0].(*ast.Element); !ok {
					t.Error("First child should be Element")
				}
				if _, ok := wrapper.Children[1].(*ast.TextNode); !ok {
					t.Error("Second child should be TextNode")
				}
				if _, ok := wrapper.Children[2].(*ast.Element); !ok {
					t.Error("Third child should be Element")
				}
			},
			description: "Mix of node types should all be wrapped",
		},
		{
			name: "multiple elements with children - wrap all with children preserved",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.TextNode{Content: "First"},
					},
				},
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.TextNode{Content: "Second"},
					},
				},
			},
			dataScope: map[string]any{
				"items": []any{1, 2, 3},
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				wrapper := nodes[0].(*ast.Element)
				if len(wrapper.Children) != 2 {
					t.Errorf("Wrapper should have 2 children, got %d", len(wrapper.Children))
				}
				// Verify nested children preserved
				firstChild, ok := wrapper.Children[0].(*ast.Element)
				if !ok {
					t.Fatal("First child should be Element")
				}
				if len(firstChild.Children) != 1 {
					t.Error("First child should have its children preserved")
				}
			},
			description: "Nested children should be preserved when wrapping",
		},
		{
			name: "three or more nodes - all wrapped",
			nodes: []ast.Node{
				&ast.Element{TagName: "header", Children: []ast.Node{}},
				&ast.Element{TagName: "main", Children: []ast.Node{}},
				&ast.Element{TagName: "footer", Children: []ast.Node{}},
				&ast.TextNode{Content: "Extra"},
			},
			dataScope: map[string]any{
				"layout": "default",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				wrapper := nodes[0].(*ast.Element)
				if len(wrapper.Children) != 4 {
					t.Errorf("Wrapper should have 4 children, got %d", len(wrapper.Children))
				}
			},
			description: "Any number of multiple nodes should be wrapped",
		},

		// ========== Data Scope Variations ==========
		{
			name: "simple data types",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"title":    "Hello",
				"count":    42,
				"price":    19.99,
				"active":   true,
				"disabled": false,
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				// Find x-data attribute
				var xDataValue string
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						xDataValue = attr.Value
					}
				}
				if xDataValue == "" {
					t.Error("x-data attribute value should not be empty")
				}
				// Note: We don't test alpineDataFormatter output format here,
				// just that a value was set
				t.Logf("x-data value: %s", xDataValue)
			},
			description: "Simple data types should be formatted in x-data",
		},
		{
			name: "complex data - objects and arrays",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"user": map[string]any{
					"name": "John",
					"role": "admin",
				},
				"items": []any{1, 2, 3},
				"config": map[string]any{
					"theme": "dark",
				},
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				var xDataValue string
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						xDataValue = attr.Value
					}
				}
				if xDataValue == "" {
					t.Error("x-data should be set for complex data")
				}
				t.Logf("Complex data x-data value: %s", xDataValue)
			},
			description: "Complex data structures should be formatted",
		},
		{
			name: "data with functions",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"count":       0,
				"increment":   "function() { this.count++; }",
				"formatPrice": "price => `$${price}`",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				var xDataValue string
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						xDataValue = attr.Value
					}
				}
				if xDataValue == "" {
					t.Error("x-data should be set even with functions")
				}
				t.Logf("Data with functions x-data value: %s", xDataValue)
			},
			description: "Data scope with functions should be handled",
		},
		{
			name: "empty data scope",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope:      map[string]any{},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				// Should still add x-data for consistency
				hasXData := false
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
					}
				}
				if !hasXData {
					t.Error("x-data should be added even for empty scope (for consistency)")
				}
			},
			description: "Empty data scope should still add x-data wrapper",
		},
		{
			name: "nil data scope",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope:      nil,
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				// Should handle nil scope gracefully
				hasXData := false
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
					}
				}
				// Behavior with nil scope - this test documents actual behavior
				t.Logf("With nil scope, hasXData=%v", hasXData)
			},
			description: "Nil data scope should be handled gracefully",
		},
		{
			name: "large data scope",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"var1":  "value1",
				"var2":  2,
				"var3":  3.14,
				"var4":  true,
				"var5":  false,
				"var6":  map[string]any{"nested": "data"},
				"var7":  []any{1, 2, 3},
				"var8":  "value8",
				"var9":  9,
				"var10": "value10",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				var xDataValue string
				for _, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						xDataValue = attr.Value
					}
				}
				if xDataValue == "" {
					t.Error("x-data should handle large scope")
				}
				t.Logf("Large scope x-data length: %d", len(xDataValue))
			},
			description: "Large data scope with many keys should be handled",
		},

		// ========== Edge Cases ==========
		{
			name:           "empty nodes slice",
			nodes:          []ast.Node{},
			dataScope:      map[string]any{"count": 0},
			expectedLength: 0,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				if len(nodes) != 0 {
					t.Errorf("Empty input should return empty output, got %d nodes", len(nodes))
				}
			},
			description: "Empty nodes slice should return empty slice",
		},
		{
			name:           "nil nodes slice",
			nodes:          nil,
			dataScope:      map[string]any{"count": 0},
			expectedLength: 0,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				if nodes == nil {
					// If implementation returns nil, that's okay
					return
				}
				if len(nodes) != 0 {
					t.Errorf("Nil input should return nil or empty, got %d nodes", len(nodes))
				}
			},
			description: "Nil nodes slice should be handled gracefully",
		},
		{
			name: "single text node - not an element",
			nodes: []ast.Node{
				&ast.TextNode{Content: "Hello World"},
			},
			dataScope: map[string]any{
				"message": "test",
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				// Single non-element node should be wrapped
				wrapper, ok := nodes[0].(*ast.Element)
				if !ok {
					t.Fatal("Single text node should be wrapped in an Element")
				}
				if wrapper.TagName != "div" {
					t.Errorf("Wrapper should be 'div', got %q", wrapper.TagName)
				}
				// Check wrapper has x-data
				hasXData := false
				for _, attr := range wrapper.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
					}
				}
				if !hasXData {
					t.Error("Wrapper should have x-data attribute")
				}
				// Check text node is preserved
				if len(wrapper.Children) != 1 {
					t.Errorf("Wrapper should have 1 child, got %d", len(wrapper.Children))
				}
				if textNode, ok := wrapper.Children[0].(*ast.TextNode); !ok {
					t.Error("Child should be TextNode")
				} else if textNode.Content != "Hello World" {
					t.Errorf("Text content should be preserved, got %q", textNode.Content)
				}
			},
			description: "Single text node should be wrapped in div",
		},
		{
			name: "single conditional node - not an element",
			nodes: []ast.Node{
				&ast.Conditional{
					IfCondition: "isActive",
					IfContent:   []ast.Node{&ast.TextNode{Content: "Active"}},
				},
			},
			dataScope: map[string]any{
				"isActive": true,
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				// Single non-element node should be wrapped
				wrapper, ok := nodes[0].(*ast.Element)
				if !ok {
					t.Fatal("Single conditional should be wrapped in an Element")
				}
				// Check wrapper contains conditional
				if len(wrapper.Children) != 1 {
					t.Errorf("Wrapper should have 1 child, got %d", len(wrapper.Children))
				}
				if _, ok := wrapper.Children[0].(*ast.Conditional); !ok {
					t.Error("Child should be Conditional")
				}
			},
			description: "Single conditional node should be wrapped",
		},
		{
			name: "single loop node - not an element",
			nodes: []ast.Node{
				&ast.Loop{
					Iterator:   "item",
					Collection: "items",
					Content:    []ast.Node{&ast.TextNode{Content: "Item"}},
				},
			},
			dataScope: map[string]any{
				"items": []any{1, 2, 3},
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				// Single non-element node should be wrapped
				wrapper, ok := nodes[0].(*ast.Element)
				if !ok {
					t.Fatal("Single loop should be wrapped in an Element")
				}
				// Check wrapper contains loop
				if len(wrapper.Children) != 1 {
					t.Errorf("Wrapper should have 1 child, got %d", len(wrapper.Children))
				}
				if _, ok := wrapper.Children[0].(*ast.Loop); !ok {
					t.Error("Child should be Loop")
				}
			},
			description: "Single loop node should be wrapped",
		},

		// ========== Attribute Handling ==========
		{
			name: "verify x-data attribute properties",
			nodes: []ast.Node{
				&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"count": 0,
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				elem := nodes[0].(*ast.Element)
				var xDataAttr *ast.Attribute
				for i, attr := range elem.Attributes {
					if attr.Name == "x-data" {
						xDataAttr = &elem.Attributes[i]
						break
					}
				}
				if xDataAttr == nil {
					t.Fatal("x-data attribute not found")
				}
				// Verify attribute properties
				if xDataAttr.Name != "x-data" {
					t.Errorf("Attribute name should be 'x-data', got %q", xDataAttr.Name)
				}
				if xDataAttr.Value == "" {
					t.Error("Attribute value should not be empty")
				}
				if !xDataAttr.Dynamic {
					t.Error("Attribute Dynamic should be true")
				}
				if !xDataAttr.IsAlpine {
					t.Error("Attribute IsAlpine should be true")
				}
				if xDataAttr.AlpineType != "data" {
					t.Errorf("Attribute AlpineType should be 'data', got %q", xDataAttr.AlpineType)
				}
			},
			description: "x-data attribute should have correct properties set",
		},
		{
			name: "x-data added to wrapper div for multiple nodes",
			nodes: []ast.Node{
				&ast.Element{TagName: "span", Children: []ast.Node{}},
				&ast.Element{TagName: "span", Children: []ast.Node{}},
			},
			dataScope: map[string]any{
				"value": 123,
			},
			expectedLength: 1,
			checkFunc: func(t *testing.T, nodes []ast.Node, dataScope map[string]any) {
				wrapper := nodes[0].(*ast.Element)
				// Verify wrapper attributes
				if len(wrapper.Attributes) != 1 {
					t.Errorf("Wrapper should have 1 attribute (x-data), got %d", len(wrapper.Attributes))
				}
				attr := wrapper.Attributes[0]
				if attr.Name != "x-data" {
					t.Errorf("Wrapper attribute should be x-data, got %q", attr.Name)
				}
				if !attr.IsAlpine || attr.AlpineType != "data" {
					t.Error("Wrapper x-data should have correct Alpine properties")
				}
			},
			description: "Wrapper div for multiple nodes should only have x-data attribute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call function under test
			result := addComponentDataWrapper(tt.nodes, tt.dataScope)

			// Check result length
			if len(result) != tt.expectedLength {
				t.Errorf("%s: expected %d nodes, got %d", tt.description, tt.expectedLength, len(result))
			}

			// Run custom check function
			if tt.checkFunc != nil {
				tt.checkFunc(t, result, tt.dataScope)
			}
		})
	}
}

// TestAddComponentDataWrapper_ReturnType verifies that the function always returns []ast.Node
func TestAddComponentDataWrapper_ReturnType(t *testing.T) {
	tests := []struct {
		name      string
		nodes     []ast.Node
		dataScope map[string]any
	}{
		{
			name:      "single element",
			nodes:     []ast.Node{&ast.Element{TagName: "div"}},
			dataScope: map[string]any{"count": 0},
		},
		{
			name:      "multiple elements",
			nodes:     []ast.Node{&ast.Element{TagName: "div"}, &ast.Element{TagName: "span"}},
			dataScope: map[string]any{"count": 0},
		},
		{
			name:      "empty nodes",
			nodes:     []ast.Node{},
			dataScope: map[string]any{"count": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addComponentDataWrapper(tt.nodes, tt.dataScope)

			// Verify return type is []ast.Node
			resultType := reflect.TypeOf(result).String()
			expectedType := "[]ast.Node"
			if resultType != expectedType {
				t.Errorf("Return type should be %s, got %s", expectedType, resultType)
			}

			// Verify result is not nil (even for empty input)
			if result == nil && tt.nodes != nil {
				t.Error("Result should not be nil when input is not nil")
			}
		})
	}
}

// TestAddComponentDataWrapper_DataScopeFormatting verifies that alpineDataFormatter is used
func TestAddComponentDataWrapper_DataScopeFormatting(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{TagName: "div", Attributes: []ast.Attribute{}, Children: []ast.Node{}},
	}

	dataScope := map[string]any{
		"firstName": "John",
		"lastName":  "Doe",
		"age":       30,
		"active":    true,
	}

	result := addComponentDataWrapper(nodes, dataScope)

	// Get the x-data value
	elem := result[0].(*ast.Element)
	var xDataValue string
	for _, attr := range elem.Attributes {
		if attr.Name == "x-data" {
			xDataValue = attr.Value
			break
		}
	}

	// Verify x-data value is not empty
	if xDataValue == "" {
		t.Error("x-data value should not be empty")
	}

	// Log the formatted value to document the format
	t.Logf("alpineDataFormatter produced: %s", xDataValue)

	// The actual format depends on alpineDataFormatter implementation
	// We just verify it's called and produces output
}

// TestAddComponentDataWrapper_PreservesNodeStructure verifies that original node
// structure is preserved (not modified) during wrapping
func TestAddComponentDataWrapper_PreservesNodeStructure(t *testing.T) {
	originalChild := &ast.TextNode{Content: "Original"}
	originalElement := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{Name: "id", Value: "original"},
		},
		Children: []ast.Node{originalChild},
	}

	nodes := []ast.Node{originalElement}
	dataScope := map[string]any{"count": 0}

	result := addComponentDataWrapper(nodes, dataScope)

	// Verify result contains our element
	resultElem := result[0].(*ast.Element)

	// Check that original element properties are preserved
	if resultElem.TagName != "div" {
		t.Errorf("TagName should be preserved, got %q", resultElem.TagName)
	}

	// Check original id attribute is still there
	hasId := false
	for _, attr := range resultElem.Attributes {
		if attr.Name == "id" && attr.Value == "original" {
			hasId = true
		}
	}
	if !hasId {
		t.Error("Original 'id' attribute should be preserved")
	}

	// Check children are preserved
	if len(resultElem.Children) != 1 {
		t.Errorf("Children should be preserved, got %d children", len(resultElem.Children))
	}

	// Verify child content
	if textNode, ok := resultElem.Children[0].(*ast.TextNode); ok {
		if textNode.Content != "Original" {
			t.Errorf("Child content should be preserved, got %q", textNode.Content)
		}
	} else {
		t.Error("Child type should be preserved as TextNode")
	}
}

// TestAddComponentDataWrapper_WrapperStructure verifies the structure of wrapper div
// created for multiple nodes
func TestAddComponentDataWrapper_WrapperStructure(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{TagName: "h1", Children: []ast.Node{&ast.TextNode{Content: "Title"}}},
		&ast.Element{TagName: "p", Children: []ast.Node{&ast.TextNode{Content: "Content"}}},
	}

	dataScope := map[string]any{
		"title":   "Test",
		"content": "Hello",
	}

	result := addComponentDataWrapper(nodes, dataScope)

	// Should return single wrapper
	if len(result) != 1 {
		t.Fatalf("Should return 1 wrapper node, got %d", len(result))
	}

	wrapper, ok := result[0].(*ast.Element)
	if !ok {
		t.Fatal("Wrapper should be an Element")
	}

	// Verify wrapper structure
	if wrapper.TagName != "div" {
		t.Errorf("Wrapper tag should be 'div', got %q", wrapper.TagName)
	}

	if wrapper.SelfClosing {
		t.Error("Wrapper should not be self-closing")
	}

	// Verify wrapper has x-data attribute
	hasXData := false
	for _, attr := range wrapper.Attributes {
		if attr.Name == "x-data" {
			hasXData = true
		}
	}
	if !hasXData {
		t.Error("Wrapper should have x-data attribute")
	}

	// Verify wrapper contains original nodes as children
	if len(wrapper.Children) != 2 {
		t.Errorf("Wrapper should have 2 children, got %d", len(wrapper.Children))
	}

	// Verify first child is h1
	if h1, ok := wrapper.Children[0].(*ast.Element); ok {
		if h1.TagName != "h1" {
			t.Errorf("First child should be h1, got %q", h1.TagName)
		}
	} else {
		t.Error("First child should be Element")
	}

	// Verify second child is p
	if p, ok := wrapper.Children[1].(*ast.Element); ok {
		if p.TagName != "p" {
			t.Errorf("Second child should be p, got %q", p.TagName)
		}
	} else {
		t.Error("Second child should be Element")
	}
}
