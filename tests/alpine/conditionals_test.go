package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/tests/testutils"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestConditionalTransformation(t *testing.T) {
	tests := []struct {
		name        string
		template    *ast.Template
		props       map[string]any
		contains    []string
		notContains []string
	}{
		{
			name: "simple if condition",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "container",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "isActive",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "Active state"},
								},
							},
						},
					},
				},
			},
			// NOTE: Don't provide isActive in props to test RUNTIME x-if generation
			// When condition values are in props, build-time conditional elimination kicks in
			props: map[string]any{},
			contains: []string{
				`<div class="container">`,
				`<template x-if="isActive">`,
				`Active state`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
		},
		{
			name: "if-else condition",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "status-indicator",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "status === 'active'",
								IfContent: []ast.Node{
									&ast.Element{
										TagName: "span",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "active",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Active"},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.Element{
										TagName: "span",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "inactive",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Inactive"},
										},
									},
								},
							},
						},
					},
				},
			},
			// NOTE: Don't provide status in props to test RUNTIME x-if generation
			// When condition values are in props, build-time conditional elimination kicks in
			props: map[string]any{},
			contains: []string{
				`<div class="status-indicator">`,
				`<template x-if="status === 'active'">`,
				`<span class="active">Active</span>`,
				`</template>`,
				// Alpine.js 3.x: else uses negated x-if instead of x-else
				`<template x-if="!(status === 'active')">`,
				`<span class="inactive">Inactive</span>`,
				`</template>`,
				`</div>`,
			},
			notContains: []string{
				// Alpine.js 3.x doesn't support x-else directive
				`<template x-else>`,
			},
		},
		{
			name: "if-else-if-else condition",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "status-display",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "status === 'active'",
								IfContent: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "active-status",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Status: Active"},
										},
									},
								},
								ElseIfConditions: []string{"status === 'pending'"},
								ElseIfContent: [][]ast.Node{
									{
										&ast.Element{
											TagName: "div",
											Attributes: []ast.Attribute{
												{
													Name:  "class",
													Value: "pending-status",
												},
											},
											Children: []ast.Node{
												&ast.TextNode{Content: "Status: Pending"},
											},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "inactive-status",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Status: Inactive"},
										},
									},
								},
							},
						},
					},
				},
			},
			// NOTE: Don't provide status in props to test RUNTIME x-if generation
			// When condition values are in props, build-time conditional elimination kicks in
			props: map[string]any{},
			contains: []string{
				`<div class="status-display">`,
				`<template x-if="status === 'active'">`,
				`<div class="active-status">Status: Active</div>`,
				`</template>`,
				// Actual format from transformer - else-if with previous condition negated
				`<template x-if="(!(status === 'active')) && (status === 'pending')">`,
				`<div class="pending-status">Status: Pending</div>`,
				`</template>`,
				// Alpine.js 3.x: else uses negated x-if with all previous conditions
				`<template x-if="!(status === 'active') && !(status === 'pending')">`,
				`<div class="inactive-status">Status: Inactive</div>`,
				`</template>`,
				`</div>`,
			},
			notContains: []string{
				`<template x-else-if="status === 'pending'">`,
				// Alpine.js 3.x doesn't support x-else directive
				`<template x-else>`,
			},
		},
		{
			name: "nested conditionals",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "user-profile",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "isLoggedIn",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "Welcome, {username}!"},
									&ast.Conditional{
										IfCondition: "isAdmin",
										IfContent: []ast.Node{
											&ast.TextNode{Content: "You have admin privileges."},
										},
										ElseContent: []ast.Node{
											&ast.TextNode{Content: "You have user privileges."},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.TextNode{Content: "Please log in."},
								},
							},
						},
					},
				},
			},
			// NOTE: Don't provide isLoggedIn/isAdmin in props to test RUNTIME x-if generation
			// When boolean condition values are in props, build-time conditional elimination kicks in
			// Keep username for expression interpolation testing
			props: map[string]any{
				"username": "JohnDoe",
			},
			contains: []string{
				`<div class="user-profile">`,
				`<template x-if="isLoggedIn">`,
				`Welcome,`,
				`<template x-if="isAdmin">`,
				`You have admin privileges.`,
				`</template>`,
				// Alpine.js 3.x: else uses negated x-if
				`<template x-if="!(isAdmin)">`,
				`You have user privileges.`,
				`</template>`,
				`</template>`,
				`<template x-if="!(isLoggedIn)">`,
				`Please log in.`,
				`</template>`,
				`</div>`,
			},
			notContains: []string{
				// Alpine.js 3.x doesn't support x-else directive
				`<template x-else>`,
			},
		},
		{
			name: "conditionals with expressions",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "cart",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "items.length > 0",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "You have {items.length} items in your cart."},
									&ast.TextNode{Content: "Total: ${total}"},
								},
								ElseContent: []ast.Node{
									&ast.TextNode{Content: "Your cart is empty."},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"items": []string{"item1", "item2", "item3"},
				"total": 42.99,
			},
			contains: []string{
				`<div class="cart">`,
				`<template x-if="items.length > 0">`,
				`You have <span x-text="items.length"></span> items in your cart.`,
				`</template>`,
				// Alpine.js 3.x: else uses negated x-if
				`<template x-if="!(items.length > 0)">`,
				`Your cart is empty.`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
			notContains: []string{
				// Alpine.js 3.x doesn't support x-else directive
				`<template x-else>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the template
			transformed := transformer.TransformAST(tt.template, tt.props)

			// Render the transformed template
			var result string
			for _, node := range transformed.RootNodes {
				result += testutils.RenderNode(node)
			}

			// Normalize and compare
			resultNormalized := testutils.NormalizeWhitespace(result)

			// Check that output contains expected strings
			for _, s := range tt.contains {
				normalizedSubstr := testutils.NormalizeWhitespace(s)
				if !strings.Contains(resultNormalized, normalizedSubstr) {
					t.Errorf("Expected output to contain (normalized):\n%s\n\nBut it doesn't. Full output (normalized):\n%s", normalizedSubstr, resultNormalized)
				}
			}

			// Check that output doesn't contain unwanted strings
			for _, s := range tt.notContains {
				normalizedSubstr := testutils.NormalizeWhitespace(s)
				if strings.Contains(resultNormalized, normalizedSubstr) {
					t.Errorf("Expected output NOT to contain (normalized):\n%s\n\nBut it does. Full output (normalized):\n%s", normalizedSubstr, resultNormalized)
				}
			}
		})
	}
}

// TestNestedConditionalsWithElseIf tests that {else if} after nested {if} blocks
// are correctly parsed as conditional clauses, not treated as text
func TestNestedConditionalsWithElseIf(t *testing.T) {
	// THIS IS THE BUG: When parsing this template string, the parser should create
	// a conditional with all three branches (if/else-if/else). However, due to the
	// nested {if} block, the {else if} and {else} are incorrectly treated as text.
	templateStr := `<div class="nested-test">
{if outer}
  Outer true
  {if inner}
    Inner true
  {/if}
{else if other}
  Other branch
{else}
  Default branch
{/if}
</div>`

	// Parse the template string
	template, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// NOTE: Don't provide boolean values in props to test RUNTIME x-if generation
	// When condition values are in props, build-time conditional elimination kicks in
	props := map[string]any{}

	// Transform the template
	transformed := transformer.TransformAST(template, props)

	// Render the transformed template
	var result string
	for _, node := range transformed.RootNodes {
		result += testutils.RenderNode(node)
	}

	// Normalize output
	resultNormalized := testutils.NormalizeWhitespace(result)

	// Expected: All three branches should be present in the output
	// If the bug exists, {else if} and {else} will be rendered as text
	// Alpine.js 3.x: else uses negated x-if instead of x-else
	expectedStrings := []string{
		`<template x-if="outer">`,
		`Outer true`,
		`<template x-if="inner">`,
		`Inner true`,
		`</template>`,
		`<template x-if="(!(outer)) && (other)">`,
		`Other branch`,
		// Actual format from transformer (simpler parentheses)
		`<template x-if="!(outer) && !(other)">`,
		`Default branch`,
	}

	for _, expected := range expectedStrings {
		normalized := testutils.NormalizeWhitespace(expected)
		if !strings.Contains(resultNormalized, normalized) {
			t.Errorf("Expected output to contain:\n%s\n\nBut it doesn't. Full output:\n%s", normalized, resultNormalized)
		}
	}

	// Verify the AST structure has all three branches
	// If the bug exists, ElseIfConditions and ElseContent will be empty or missing
	if len(template.RootNodes) == 0 {
		t.Fatal("Template has no root nodes")
	}

	elem, ok := template.RootNodes[0].(*ast.Element)
	if !ok {
		t.Fatal("First root node is not an element")
	}

	if len(elem.Children) == 0 {
		t.Fatal("Element has no children")
	}

	cond, ok := elem.Children[0].(*ast.Conditional)
	if !ok {
		t.Fatal("First child is not a conditional")
	}

	// Verify all three branches exist in AST
	if cond.IfCondition == "" {
		t.Error("Missing if condition")
	}
	if len(cond.ElseIfConditions) != 1 {
		t.Errorf("Expected 1 else-if condition, got %d. This indicates the parser bug!", len(cond.ElseIfConditions))
	}
	if len(cond.ElseContent) == 0 {
		t.Error("Missing else content. This indicates the parser bug!")
	}
}

// TestDeeplyNestedConditionals tests that the parser correctly handles
// 3+ levels of conditional nesting
func TestDeeplyNestedConditionals(t *testing.T) {
	templateStr := `<div class="deep-nest">
{if level1}
  Level 1
  {if level2}
    Level 2
    {if level3}
      Level 3
    {else}
      Level 3 else
    {/if}
  {else}
    Level 2 else
  {/if}
{else}
  Level 1 else
{/if}
</div>`

	// Parse the template string
	template, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// NOTE: Don't provide boolean values in props to test RUNTIME x-if generation
	// When condition values are in props, build-time conditional elimination kicks in
	props := map[string]any{}

	// Transform the template
	transformed := transformer.TransformAST(template, props)

	// Render the transformed template
	var result string
	for _, node := range transformed.RootNodes {
		result += testutils.RenderNode(node)
	}

	// Normalize output
	resultNormalized := testutils.NormalizeWhitespace(result)

	// Expected: All nested levels should be present with proper structure
	// Alpine.js 3.x: else uses negated x-if instead of x-else
	expectedStrings := []string{
		`<template x-if="level1">`,
		`Level 1`,
		`<template x-if="level2">`,
		`Level 2`,
		`<template x-if="level3">`,
		`Level 3`,
		`<template x-if="!(level3)">`,
		`Level 3 else`,
		`<template x-if="!(level2)">`,
		`Level 2 else`,
		`<template x-if="!(level1)">`,
		`Level 1 else`,
	}

	for _, expected := range expectedStrings {
		normalized := testutils.NormalizeWhitespace(expected)
		if !strings.Contains(resultNormalized, normalized) {
			t.Errorf("Expected output to contain:\n%s\n\nBut it doesn't. Full output:\n%s", normalized, resultNormalized)
		}
	}

	// Verify AST structure maintains nesting
	if len(template.RootNodes) == 0 {
		t.Fatal("Template has no root nodes")
	}

	elem, ok := template.RootNodes[0].(*ast.Element)
	if !ok {
		t.Fatal("First root node is not an element")
	}

	// Verify Level 1 conditional
	level1Cond, ok := elem.Children[0].(*ast.Conditional)
	if !ok {
		t.Fatal("Level 1: Expected conditional node")
	}

	if level1Cond.IfCondition != "level1" {
		t.Errorf("Level 1: Expected condition 'level1', got '%s'", level1Cond.IfCondition)
	}

	// Verify Level 2 conditional is nested inside Level 1
	if len(level1Cond.IfContent) < 2 {
		t.Fatal("Level 1: Expected at least 2 children (text + conditional)")
	}

	level2Cond, ok := level1Cond.IfContent[1].(*ast.Conditional)
	if !ok {
		t.Fatal("Level 2: Expected conditional node nested in Level 1")
	}

	if level2Cond.IfCondition != "level2" {
		t.Errorf("Level 2: Expected condition 'level2', got '%s'", level2Cond.IfCondition)
	}

	// Verify Level 3 conditional is nested inside Level 2
	if len(level2Cond.IfContent) < 2 {
		t.Fatal("Level 2: Expected at least 2 children (text + conditional)")
	}

	level3Cond, ok := level2Cond.IfContent[1].(*ast.Conditional)
	if !ok {
		t.Fatal("Level 3: Expected conditional node nested in Level 2")
	}

	if level3Cond.IfCondition != "level3" {
		t.Errorf("Level 3: Expected condition 'level3', got '%s'", level3Cond.IfCondition)
	}
}

// TestNestedConditionalsInLoops tests that {if} blocks inside {for} loops
// are correctly parsed and don't interfere with loop iteration
func TestNestedConditionalsInLoops(t *testing.T) {
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:  "class",
						Value: "loop-conditional",
					},
				},
				Children: []ast.Node{
					&ast.Loop{
						Iterator:   "item",
						Collection: "items",
						Content: []ast.Node{
							&ast.Element{
								TagName: "div",
								Attributes: []ast.Attribute{
									{
										Name:  "class",
										Value: "item",
									},
								},
								Children: []ast.Node{
									&ast.TextNode{Content: "{item.name}"},
									&ast.Conditional{
										IfCondition: "item.active",
										IfContent: []ast.Node{
											&ast.TextNode{Content: "Active"},
										},
										ElseContent: []ast.Node{
											&ast.TextNode{Content: "Inactive"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	props := map[string]any{
		"items": []map[string]any{
			{"name": "Item 1", "active": true},
			{"name": "Item 2", "active": false},
		},
	}

	// Transform the template
	transformed := transformer.TransformAST(template, props)

	// Render the transformed template
	var result string
	for _, node := range transformed.RootNodes {
		result += testutils.RenderNode(node)
	}

	// Normalize output
	resultNormalized := testutils.NormalizeWhitespace(result)

	// Expected: Loop should contain conditionals correctly
	// Alpine.js 3.x: else uses negated x-if instead of x-else
	expectedStrings := []string{
		`<div class="item">`,
		`<span x-text="item.name"></span>`,
		`<template x-if="item.active">`,
		`Active`,
		`<template x-if="!(item.active)">`,
		`Inactive`,
	}

	for _, expected := range expectedStrings {
		normalized := testutils.NormalizeWhitespace(expected)
		if !strings.Contains(resultNormalized, normalized) {
			t.Errorf("Expected output to contain:\n%s\n\nBut it doesn't. Full output:\n%s", normalized, resultNormalized)
		}
	}

	// Verify AST structure
	elem, ok := template.RootNodes[0].(*ast.Element)
	if !ok {
		t.Fatal("First root node is not an element")
	}

	loop, ok := elem.Children[0].(*ast.Loop)
	if !ok {
		t.Fatal("Expected loop node")
	}

	if loop.Iterator != "item" {
		t.Errorf("Expected iterator 'item', got '%s'", loop.Iterator)
	}

	if loop.Collection != "items" {
		t.Errorf("Expected collection 'items', got '%s'", loop.Collection)
	}

	// Verify conditional is inside loop content
	loopElem, ok := loop.Content[0].(*ast.Element)
	if !ok {
		t.Fatal("Expected element inside loop")
	}

	found := false
	for _, child := range loopElem.Children {
		if _, ok := child.(*ast.Conditional); ok {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected conditional node inside loop content")
	}
}

// TestLoopsInConditionals tests that {for} loops inside {if} blocks
// are correctly parsed and conditional depth doesn't affect loop parsing
func TestLoopsInConditionals(t *testing.T) {
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:  "class",
						Value: "conditional-loop",
					},
				},
				Children: []ast.Node{
					&ast.Conditional{
						IfCondition: "hasItems",
						IfContent: []ast.Node{
							&ast.TextNode{Content: "Items:"},
							&ast.Loop{
								Iterator:   "item",
								Collection: "items",
								Content: []ast.Node{
									&ast.Element{
										TagName:    "span",
										Attributes: []ast.Attribute{},
										Children: []ast.Node{
											&ast.TextNode{Content: "{item}"},
										},
									},
								},
							},
						},
						ElseContent: []ast.Node{
							&ast.TextNode{Content: "No items"},
						},
					},
				},
			},
		},
	}

	// NOTE: Don't provide hasItems/items in props to test RUNTIME x-if/x-for generation
	// When condition values are in props, build-time conditional elimination kicks in
	props := map[string]any{}

	// Transform the template
	transformed := transformer.TransformAST(template, props)

	// Render the transformed template
	var result string
	for _, node := range transformed.RootNodes {
		result += testutils.RenderNode(node)
	}

	// Normalize output
	resultNormalized := testutils.NormalizeWhitespace(result)

	// Expected: Conditional should contain loop correctly
	// Alpine.js 3.x: else uses negated x-if instead of x-else
	// x-for uses "(, item) in items" format for index support
	// {item} in TextNode is parsed and converted to <span x-text="item">
	expectedStrings := []string{
		`<template x-if="hasItems">`,
		`Items:`,
		`<template x-for="(, item) in items">`,
		`<span x-text="item">`,
		`<template x-if="!(hasItems)">`,
		`No items`,
	}

	for _, expected := range expectedStrings {
		normalized := testutils.NormalizeWhitespace(expected)
		if !strings.Contains(resultNormalized, normalized) {
			t.Errorf("Expected output to contain:\n%s\n\nBut it doesn't. Full output:\n%s", normalized, resultNormalized)
		}
	}

	// Verify AST structure
	elem, ok := template.RootNodes[0].(*ast.Element)
	if !ok {
		t.Fatal("First root node is not an element")
	}

	cond, ok := elem.Children[0].(*ast.Conditional)
	if !ok {
		t.Fatal("Expected conditional node")
	}

	if cond.IfCondition != "hasItems" {
		t.Errorf("Expected condition 'hasItems', got '%s'", cond.IfCondition)
	}

	// Verify loop is inside conditional content
	found := false
	for _, child := range cond.IfContent {
		if loop, ok := child.(*ast.Loop); ok {
			found = true
			if loop.Iterator != "item" {
				t.Errorf("Expected iterator 'item', got '%s'", loop.Iterator)
			}
			if loop.Collection != "items" {
				t.Errorf("Expected collection 'items', got '%s'", loop.Collection)
			}
			break
		}
	}

	if !found {
		t.Error("Expected loop node inside conditional content")
	}
}

// TestMixedNesting tests complex combinations of loops and conditionals
// to ensure they work together correctly
func TestMixedNesting(t *testing.T) {
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:  "class",
						Value: "mixed-nest",
					},
				},
				Children: []ast.Node{
					&ast.Conditional{
						IfCondition: "showUsers",
						IfContent: []ast.Node{
							&ast.Loop{
								Iterator:   "user",
								Collection: "users",
								Content: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "user",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "{user.name}"},
											&ast.Conditional{
												IfCondition: "user.isAdmin",
												IfContent: []ast.Node{
													&ast.TextNode{Content: "Admin"},
												},
												ElseIfConditions: []string{"user.isModerator"},
												ElseIfContent: [][]ast.Node{
													{
														&ast.TextNode{Content: "Moderator"},
													},
												},
												ElseContent: []ast.Node{
													&ast.TextNode{Content: "User"},
												},
											},
										},
									},
								},
							},
						},
						ElseContent: []ast.Node{
							&ast.TextNode{Content: "Users hidden"},
						},
					},
				},
			},
		},
	}

	// NOTE: Don't provide showUsers in props to test RUNTIME x-if generation
	// When condition values are in props, build-time conditional elimination kicks in
	// users array causes build-time loop expansion, inner conditionals remain runtime
	props := map[string]any{
		"users": []map[string]any{
			{"name": "Alice", "isAdmin": true, "isModerator": false},
			{"name": "Bob", "isAdmin": false, "isModerator": true},
			{"name": "Charlie", "isAdmin": false, "isModerator": false},
		},
	}

	// Transform the template
	transformed := transformer.TransformAST(template, props)

	// Render the transformed template
	var result string
	for _, node := range transformed.RootNodes {
		result += testutils.RenderNode(node)
	}

	// Normalize output
	resultNormalized := testutils.NormalizeWhitespace(result)

	// Expected: Complex nesting should work correctly
	// Alpine.js 3.x: else uses negated x-if instead of x-else
	// Loop is expanded at build-time with users array
	// Inner conditionals remain runtime (checking user properties)
	expectedStrings := []string{
		`<template x-if="showUsers">`,
		`<div class="user">`,
		`<span x-text="user.name"></span>`,
		`<template x-if="user.isAdmin">`,
		`Admin`,
		`<template x-if="(!(user.isAdmin)) && (user.isModerator)">`,
		`Moderator`,
		// Actual format: simpler parentheses for else clause
		`<template x-if="!(user.isAdmin) && !(user.isModerator)">`,
		`User`,
		`</template>`,
		`<template x-if="!(showUsers)">`,
		`Users hidden`,
	}

	for _, expected := range expectedStrings {
		normalized := testutils.NormalizeWhitespace(expected)
		if !strings.Contains(resultNormalized, normalized) {
			t.Errorf("Expected output to contain:\n%s\n\nBut it doesn't. Full output:\n%s", normalized, resultNormalized)
		}
	}

	// Verify AST structure - outer conditional
	elem, ok := template.RootNodes[0].(*ast.Element)
	if !ok {
		t.Fatal("First root node is not an element")
	}

	outerCond, ok := elem.Children[0].(*ast.Conditional)
	if !ok {
		t.Fatal("Expected outer conditional node")
	}

	// Verify loop inside conditional
	var loop *ast.Loop
	for _, child := range outerCond.IfContent {
		if l, ok := child.(*ast.Loop); ok {
			loop = l
			break
		}
	}

	if loop == nil {
		t.Fatal("Expected loop inside outer conditional")
	}

	// Verify conditional inside loop
	loopElem, ok := loop.Content[0].(*ast.Element)
	if !ok {
		t.Fatal("Expected element inside loop")
	}

	var innerCond *ast.Conditional
	for _, child := range loopElem.Children {
		if c, ok := child.(*ast.Conditional); ok {
			innerCond = c
			break
		}
	}

	if innerCond == nil {
		t.Fatal("Expected conditional inside loop")
	}

	// Verify inner conditional has if/else-if/else structure
	if innerCond.IfCondition == "" {
		t.Error("Missing if condition in inner conditional")
	}
	if len(innerCond.ElseIfConditions) != 1 {
		t.Errorf("Expected 1 else-if in inner conditional, got %d", len(innerCond.ElseIfConditions))
	}
	if len(innerCond.ElseContent) == 0 {
		t.Error("Missing else content in inner conditional")
	}
}
