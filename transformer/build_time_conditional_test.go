package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestBuildTimeConditionalResolution verifies that {if published} resolves at build time
// when published: true is in the dataScope
func TestBuildTimeConditionalResolution(t *testing.T) {
	// Create a simple conditional node like {if published}
	conditionalNode := &ast.Conditional{
		IfCondition: "published",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Published content"},
		},
		ElseContent: []ast.Node{
			&ast.TextNode{Content: "Draft content"},
		},
	}

	// Create dataScope with published: true (boolean)
	dataScope := map[string]any{
		"published": true,
	}

	// Test tryResolveBuildTimeConditional directly
	resolved, branchToRender := tryResolveBuildTimeConditional("published", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected conditional to be resolved at build time, but it wasn't")
		t.Logf("dataScope: %v", dataScope)
	}

	if branchToRender != "if" {
		t.Errorf("Expected branchToRender='if', got '%s'", branchToRender)
	}
}

// TestBuildTimeConditionalWithStringTrue verifies resolution with string "true"
func TestBuildTimeConditionalWithStringTrue(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "published",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Published content"},
		},
	}

	// Create dataScope with published: "true" (string)
	dataScope := map[string]any{
		"published": "true",
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("published", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected conditional with string 'true' to be resolved at build time")
	}

	if branchToRender != "if" {
		t.Errorf("Expected branchToRender='if', got '%s'", branchToRender)
	}
}

// TestBuildTimeConditionalWithNestedContent simulates the actual news page flow
// where content comes from nested content.fields structure
func TestBuildTimeConditionalWithNestedContent(t *testing.T) {
	// Simulate the dataScope as it would be in the html.html wrapper
	// when rendering a news page
	parentDataScope := map[string]any{
		"layout": "news",
		"content": map[string]interface{}{
			"fields": map[string]interface{}{
				"published": true,
				"title":     "Test Article",
			},
		},
	}

	// Test resolveSpreadProps with "content.fields"
	spreadProps := resolveSpreadProps([]string{"content.fields"}, parentDataScope)

	// Verify published is in the spread props
	if published, ok := spreadProps["published"]; !ok {
		t.Errorf("Expected 'published' in spread props, but it wasn't found")
		t.Logf("parentDataScope: %v", parentDataScope)
		t.Logf("spreadProps: %v", spreadProps)
	} else {
		t.Logf("published value: %v (type: %T)", published, published)
		if published != true {
			t.Errorf("Expected published=true, got %v", published)
		}
	}

	// Now test the conditional resolution with the spread props as dataScope
	conditionalNode := &ast.Conditional{
		IfCondition: "published",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Published content"},
		},
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("published", conditionalNode, spreadProps)
	if !resolved {
		t.Errorf("Expected conditional to be resolved at build time with spread props")
	}
	if branchToRender != "if" {
		t.Errorf("Expected branchToRender='if', got '%s'", branchToRender)
	}
}

// TestTransformConditionalBuildTimeResolution tests the full transformConditional function
func TestTransformConditionalBuildTimeResolution(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "published",
		IfContent: []ast.Node{
			&ast.Element{
				TagName:  "div",
				Children: []ast.Node{&ast.TextNode{Content: "Published content"}},
			},
		},
		ElseContent: []ast.Node{
			&ast.Element{
				TagName:  "div",
				Children: []ast.Node{&ast.TextNode{Content: "Draft content"}},
			},
		},
	}

	dataScope := map[string]any{
		"published": true,
	}

	// Transform the conditional
	result := transformConditional(conditionalNode, dataScope)

	// Check that we got the if-branch content directly (not wrapped in <template x-if>)
	if len(result) == 0 {
		t.Errorf("Expected at least one node, got empty result")
		return
	}

	// The result should be the if-branch content (a div), NOT a template with x-if
	firstNode := result[0]
	if elem, ok := firstNode.(*ast.Element); ok {
		if elem.TagName == "template" {
			// Check if it has x-if attribute - this means build-time resolution failed
			for _, attr := range elem.Attributes {
				if attr.Name == "x-if" {
					t.Errorf("Build-time resolution failed: got <template x-if='%s'> instead of direct content", attr.Value)
					return
				}
			}
		}
		if elem.TagName != "div" {
			t.Errorf("Expected a div element, got <%s>", elem.TagName)
		}
	} else {
		t.Errorf("Expected an Element node, got %T", firstNode)
	}
}

// TestFullComponentTransformationWithPublished tests the complete flow:
// Register component -> Transform with resolved props -> Check conditional resolution
func TestFullComponentTransformationWithPublished(t *testing.T) {
	// Create a component AST that looks like news.html:
	// ---
	// export let published, title
	// ---
	// {if published}
	// <div class="content">Published!</div>
	// {/if}
	componentAST := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				ExportedProps: []string{"published", "title"},
			},
			&ast.Conditional{
				IfCondition: "published",
				IfContent: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{Name: "class", Value: "content"},
						},
						Children: []ast.Node{
							&ast.TextNode{Content: "Published!"},
						},
					},
				},
				ElseContent: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.TextNode{Content: "Draft"},
						},
					},
				},
			},
		},
	}

	// Register the component
	RegisterComponent("test-news", componentAST, []string{"published", "title"})

	// Resolved props as they would come from content.fields spread
	resolvedProps := map[string]any{
		"published": true,
		"title":     "Test Article",
	}

	// Transform the component with resolved props
	parentDataScope := map[string]any{}
	result := TransformComponentWithResolvedProps("test-news", resolvedProps, parentDataScope)

	t.Logf("Result has %d nodes", len(result))

	// The result should be wrapped in an x-data div
	if len(result) == 0 {
		t.Fatalf("Expected at least one node, got empty result")
	}

	// Find the content - it should be a div with class="content", NOT a <template x-if>
	found := false
	foundTemplateXIf := false

	var checkNodes func(nodes []ast.Node)
	checkNodes = func(nodes []ast.Node) {
		for _, node := range nodes {
			if elem, ok := node.(*ast.Element); ok {
				if elem.TagName == "template" {
					for _, attr := range elem.Attributes {
						if attr.Name == "x-if" && attr.Value == "published" {
							foundTemplateXIf = true
							t.Logf("Found <template x-if='published'> - build-time resolution FAILED")
						}
					}
				}
				if elem.TagName == "div" {
					for _, attr := range elem.Attributes {
						if attr.Name == "class" && attr.Value == "content" {
							found = true
							t.Logf("Found <div class='content'> - build-time resolution SUCCEEDED")
						}
					}
				}
				// Check children
				checkNodes(elem.Children)
			}
		}
	}

	checkNodes(result)

	if foundTemplateXIf && !found {
		t.Errorf("Build-time resolution failed: got <template x-if='published'> instead of direct content")
	}
	if !found && !foundTemplateXIf {
		t.Errorf("Could not find expected content in result")
	}
}

// =============================================================================
// Tests for Negation Pattern Build-Time Resolution
// =============================================================================

// TestBuildTimeConditionalWithNegation verifies {if !compact} with compact=false resolves to if-branch
func TestBuildTimeConditionalWithNegation(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "!compact",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Full view"},
		},
		ElseContent: []ast.Node{
			&ast.TextNode{Content: "Compact view"},
		},
	}

	// compact: false means !compact is true → render if-branch
	dataScope := map[string]any{
		"compact": false,
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("!compact", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected negation '!compact' with compact=false to be resolved at build time")
	}

	if branchToRender != "if" {
		t.Errorf("Expected branchToRender='if' (since !false = true), got '%s'", branchToRender)
	}
}

// TestBuildTimeConditionalWithNegationElseBranch verifies {if !compact} with compact=true resolves to else-branch
func TestBuildTimeConditionalWithNegationElseBranch(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "!compact",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Full view"},
		},
		ElseContent: []ast.Node{
			&ast.TextNode{Content: "Compact view"},
		},
	}

	// compact: true means !compact is false → render else-branch
	dataScope := map[string]any{
		"compact": true,
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("!compact", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected negation '!compact' with compact=true to be resolved at build time")
	}

	if branchToRender != "else" {
		t.Errorf("Expected branchToRender='else' (since !true = false), got '%s'", branchToRender)
	}
}

// TestBuildTimeConditionalWithNegationNoElse verifies {if !compact} with compact=true and no else-branch resolves to "none"
func TestBuildTimeConditionalWithNegationNoElse(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "!compact",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Full view"},
		},
		// No ElseContent
	}

	// compact: true means !compact is false → nothing to render
	dataScope := map[string]any{
		"compact": true,
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("!compact", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected negation '!compact' with compact=true to be resolved at build time")
	}

	if branchToRender != "none" {
		t.Errorf("Expected branchToRender='none' (no else branch), got '%s'", branchToRender)
	}
}

// TestBuildTimeConditionalWithNegationParentheses verifies !(compact) pattern
func TestBuildTimeConditionalWithNegationParentheses(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "!(compact)",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Full view"},
		},
	}

	dataScope := map[string]any{
		"compact": false,
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("!(compact)", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected negation '!(compact)' with compact=false to be resolved at build time")
	}

	if branchToRender != "if" {
		t.Errorf("Expected branchToRender='if', got '%s'", branchToRender)
	}
}

// TestBuildTimeConditionalWithNegationPropertyPath verifies {if !user.avatar} pattern
func TestBuildTimeConditionalWithNegationPropertyPath(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "!user.avatar",
		IfContent: []ast.Node{
			&ast.TextNode{Content: "Show default avatar"},
		},
		ElseContent: []ast.Node{
			&ast.TextNode{Content: "Show user avatar"},
		},
	}

	// user.avatar is nil/empty → !user.avatar is true
	dataScope := map[string]any{
		"user": map[string]interface{}{
			"name":   "John",
			"avatar": nil, // No avatar
		},
	}

	resolved, branchToRender := tryResolveBuildTimeConditional("!user.avatar", conditionalNode, dataScope)

	if !resolved {
		t.Errorf("Expected negation '!user.avatar' with avatar=nil to be resolved at build time")
	}

	if branchToRender != "if" {
		t.Errorf("Expected branchToRender='if' (since !nil = true), got '%s'", branchToRender)
	}
}

// TestTransformConditionalNegationBuildTime tests full transform with negation
func TestTransformConditionalNegationBuildTime(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "!showRole",
		IfContent: []ast.Node{
			&ast.Element{
				TagName:  "div",
				Children: []ast.Node{&ast.TextNode{Content: "Role hidden"}},
			},
		},
		ElseContent: []ast.Node{
			&ast.Element{
				TagName: "span",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "role"},
				},
				Children: []ast.Node{&ast.TextNode{Content: "Admin"}},
			},
		},
	}

	// Initialize runtime tracker to prevent nil pointer
	runtimeTracker = NewRuntimeVarTracker()

	// showRole: false means !showRole is true → render if-branch (Role hidden)
	dataScope := map[string]any{
		"showRole": false,
	}

	result := transformConditional(conditionalNode, dataScope)

	if len(result) == 0 {
		t.Fatalf("Expected at least one node, got empty result")
	}

	// Should get direct content (div with "Role hidden"), NOT a <template x-if>
	firstNode := result[0]
	elem, ok := firstNode.(*ast.Element)
	if !ok {
		t.Fatalf("Expected Element, got %T", firstNode)
	}

	if elem.TagName == "template" {
		t.Errorf("Build-time resolution failed: got <template> instead of direct content")
		return
	}

	if elem.TagName != "div" {
		t.Errorf("Expected <div>, got <%s>", elem.TagName)
	}
}

// TestRuntimeVarNotTrackedForEliminatedBranch verifies that variables in eliminated branches are NOT tracked
func TestRuntimeVarNotTrackedForEliminatedBranch(t *testing.T) {
	conditionalNode := &ast.Conditional{
		IfCondition: "showRole",
		IfContent: []ast.Node{
			// This branch has {roleBadge.bg} which references roleBadge
			&ast.ExpressionNode{Expression: "roleBadge.bg"},
		},
		ElseContent: []ast.Node{
			&ast.TextNode{Content: "No role"},
		},
	}

	// Initialize fresh tracker
	runtimeTracker = NewRuntimeVarTracker()

	// showRole: false → the if-branch with roleBadge is eliminated
	dataScope := map[string]any{
		"showRole":  false,
		"roleBadge": "some-getter-that-should-not-be-tracked",
	}

	transformConditional(conditionalNode, dataScope)

	// roleBadge should NOT be tracked since its branch was eliminated
	if runtimeTracker.IsTracked("roleBadge") {
		t.Errorf("Variable 'roleBadge' should NOT be tracked when its branch is eliminated at build-time")
	}
}
