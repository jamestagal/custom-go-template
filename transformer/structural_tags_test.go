package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestIsStructuralTag verifies that structural HTML tags are correctly identified
func TestIsStructuralTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected bool
	}{
		// Structural tags (should return true)
		{name: "html tag", tagName: "html", expected: true},
		{name: "HTML uppercase", tagName: "HTML", expected: true},
		{name: "head tag", tagName: "head", expected: true},
		{name: "HEAD uppercase", tagName: "HEAD", expected: true},
		{name: "body tag", tagName: "body", expected: true},
		{name: "BODY uppercase", tagName: "BODY", expected: true},
		{name: "doctype tag", tagName: "!doctype", expected: true},
		{name: "DOCTYPE uppercase", tagName: "!DOCTYPE", expected: true},

		// Regular tags (should return false)
		{name: "div tag", tagName: "div", expected: false},
		{name: "header component", tagName: "header", expected: false},
		{name: "footer component", tagName: "footer", expected: false},
		{name: "section tag", tagName: "section", expected: false},
		{name: "main tag", tagName: "main", expected: false},
		{name: "nav tag", tagName: "nav", expected: false},
		{name: "article tag", tagName: "article", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStructuralTag(tt.tagName)
			if result != tt.expected {
				t.Errorf("isStructuralTag(%q) = %v, want %v", tt.tagName, result, tt.expected)
			}
		})
	}
}

// TestWrapWithXData_StructuralTags verifies that structural tags don't get x-data added
func TestWrapWithXData_StructuralTags(t *testing.T) {
	tests := []struct {
		name             string
		tagName          string
		shouldSkipXData  bool
		expectedXDataSet bool
	}{
		{
			name:             "head tag should skip x-data",
			tagName:          "head",
			shouldSkipXData:  true,
			expectedXDataSet: false,
		},
		{
			name:             "body tag should skip x-data",
			tagName:          "body",
			shouldSkipXData:  true,
			expectedXDataSet: false,
		},
		{
			name:             "html tag should skip x-data",
			tagName:          "html",
			shouldSkipXData:  true,
			expectedXDataSet: false,
		},
		{
			name:             "div tag should get x-data",
			tagName:          "div",
			shouldSkipXData:  false,
			expectedXDataSet: true,
		},
		{
			name:             "header component should get x-data",
			tagName:          "header",
			shouldSkipXData:  false,
			expectedXDataSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple element node
			element := &ast.Element{
				TagName: tt.tagName,
				Children: []ast.Node{
					&ast.TextNode{Content: "content"},
				},
			}

			// Create data scope with some props
			dataScope := map[string]any{
				"title":       "My Title",
				"description": "My Description",
			}

			// Wrap the element
			result := wrapWithXData([]ast.Node{element}, dataScope)

			// Verify result
			if len(result) != 1 {
				t.Fatalf("Expected 1 node, got %d", len(result))
			}

			resultElement, ok := result[0].(*ast.Element)
			if !ok {
				t.Fatalf("Expected element node, got %T", result[0])
			}

			// Check if x-data was added
			hasXData := false
			for _, attr := range resultElement.Attributes {
				if attr.Name == "x-data" {
					hasXData = true
					break
				}
			}

			if hasXData != tt.expectedXDataSet {
				if tt.shouldSkipXData {
					t.Errorf("Structural tag <%s> should NOT have x-data, but it does", tt.tagName)
				} else {
					t.Errorf("Regular tag <%s> should have x-data, but it doesn't", tt.tagName)
				}
			}

			// Verify the tag name wasn't changed
			if resultElement.TagName != tt.tagName {
				t.Errorf("Tag name changed from %q to %q", tt.tagName, resultElement.TagName)
			}
		})
	}
}

// TestHeadComponent_NoXData verifies that a Head component with props doesn't get x-data on the head tag
func TestHeadComponent_NoXData(t *testing.T) {
	// Register a Head component with props
	headTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: "My App"},
					{Name: "description", DefaultValue: "Description"},
				},
			},
			&ast.Element{
				TagName: "head",
				Children: []ast.Node{
					&ast.Element{
						TagName: "title",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "title"},
						},
					},
					&ast.Element{
						TagName: "meta",
						Attributes: []ast.Attribute{
							{Name: "name", Value: "description"},
							{Name: "content", Value: "{description}", Dynamic: true},
						},
					},
				},
			},
		},
	}

	RegisterComponent("Head", headTemplate, []string{"title", "description"})
	defer UnregisterComponent("Head")

	// Create a component node referencing the Head component with custom props
	componentNode := &ast.ComponentNode{
		Name: "Head",
		Props: []ast.ComponentProp{
			{Name: "title", Value: "Custom Go Template", IsDynamic: false},
			{Name: "description", Value: "A powerful template engine", IsDynamic: false},
		},
	}

	// Transform the component
	parentScope := make(map[string]any)
	result := transformComponent(componentNode, parentScope)

	// Verify result
	if len(result) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(result))
	}

	headElement, ok := result[0].(*ast.Element)
	if !ok {
		t.Fatalf("Expected element node, got %T", result[0])
	}

	// Verify it's a head tag
	if headElement.TagName != "head" {
		t.Errorf("Expected <head> tag, got <%s>", headElement.TagName)
	}

	// CRITICAL: Verify that x-data was NOT added to the head tag
	hasXData := false
	var xDataValue string
	for _, attr := range headElement.Attributes {
		if attr.Name == "x-data" {
			hasXData = true
			xDataValue = attr.Value
			break
		}
	}

	if hasXData {
		t.Errorf("FAIL: <head> tag should NOT have x-data attribute, but it has x-data=%q", xDataValue)
		t.Error("This would cause Alpine.js to parse metadata tags as reactive code, breaking the page!")
	} else {
		t.Log("SUCCESS: <head> tag correctly skipped x-data wrapping")
	}

	// Verify children were still transformed correctly
	if len(headElement.Children) < 1 {
		t.Error("Head element should have children (title, meta tags)")
	}
}

// TestBodyComponent_NoXData verifies that a body tag doesn't get x-data
func TestBodyComponent_NoXData(t *testing.T) {
	// Create a body element with some reactive content
	bodyElement := &ast.Element{
		TagName: "body",
		Children: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "container"},
				},
				Children: []ast.Node{
					&ast.ExpressionNode{Expression: "message"},
				},
			},
		},
	}

	// Data scope with reactive variables
	dataScope := map[string]any{
		"message": "Hello World",
	}

	// Wrap the body element
	result := wrapWithXData([]ast.Node{bodyElement}, dataScope)

	// Verify result
	if len(result) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(result))
	}

	resultElement, ok := result[0].(*ast.Element)
	if !ok {
		t.Fatalf("Expected element node, got %T", result[0])
	}

	// Verify it's still a body tag
	if resultElement.TagName != "body" {
		t.Errorf("Expected <body> tag, got <%s>", resultElement.TagName)
	}

	// CRITICAL: Verify that x-data was NOT added
	hasXData := false
	for _, attr := range resultElement.Attributes {
		if attr.Name == "x-data" {
			hasXData = true
			break
		}
	}

	if hasXData {
		t.Error("FAIL: <body> tag should NOT have x-data attribute")
		t.Error("The server should add x-data to <body> when needed, not the transformer")
	} else {
		t.Log("SUCCESS: <body> tag correctly skipped x-data wrapping")
	}
}

// TestHtmlComponent_NoXData verifies that an html tag doesn't get x-data
func TestHtmlComponent_NoXData(t *testing.T) {
	htmlElement := &ast.Element{
		TagName: "html",
		Attributes: []ast.Attribute{
			{Name: "lang", Value: "en"},
		},
		Children: []ast.Node{
			&ast.Element{
				TagName: "head",
				Children: []ast.Node{
					&ast.Element{
						TagName: "title",
						Children: []ast.Node{
							&ast.TextNode{Content: "My App"},
						},
					},
				},
			},
			&ast.Element{
				TagName: "body",
				Children: []ast.Node{
					&ast.TextNode{Content: "Content"},
				},
			},
		},
	}

	dataScope := map[string]any{
		"appName": "My App",
	}

	// Wrap the html element
	result := wrapWithXData([]ast.Node{htmlElement}, dataScope)

	// Verify result
	if len(result) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(result))
	}

	resultElement, ok := result[0].(*ast.Element)
	if !ok {
		t.Fatalf("Expected element node, got %T", result[0])
	}

	// Verify it's still an html tag
	if !strings.EqualFold(resultElement.TagName, "html") {
		t.Errorf("Expected <html> tag, got <%s>", resultElement.TagName)
	}

	// CRITICAL: Verify that x-data was NOT added
	hasXData := false
	for _, attr := range resultElement.Attributes {
		if attr.Name == "x-data" {
			hasXData = true
			break
		}
	}

	if hasXData {
		t.Error("FAIL: <html> tag should NOT have x-data attribute")
	} else {
		t.Log("SUCCESS: <html> tag correctly skipped x-data wrapping")
	}
}
