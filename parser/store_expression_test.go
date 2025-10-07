package parser

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestStoreExpressionParser tests the parsing of store expressions
// Pattern: parseStoreExpression() [Load: 6]
// Cognitive Load: 6 (simple parsing logic with validation)
func TestStoreExpressionParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSuccess bool
		wantStore   string
		wantProp    string
		wantError   string
	}{
		{
			name:        "simple store reference",
			input:       "$auth",
			wantSuccess: true,
			wantStore:   "auth",
			wantProp:    "",
		},
		{
			name:        "store with single property",
			input:       "$auth.isLoggedIn",
			wantSuccess: true,
			wantStore:   "auth",
			wantProp:    "isLoggedIn",
		},
		{
			name:        "store with nested property",
			input:       "$auth.user.name",
			wantSuccess: true,
			wantStore:   "auth",
			wantProp:    "user.name",
		},
		{
			name:        "store with deep nested property",
			input:       "$settings.theme.colors.primary",
			wantSuccess: true,
			wantStore:   "settings",
			wantProp:    "theme.colors.primary",
		},
		{
			name:        "store with array access",
			input:       "$cart.items.length",
			wantSuccess: true,
			wantStore:   "cart",
			wantProp:    "items.length",
		},
		{
			name:        "store with underscore in name",
			input:       "$user_profile.avatar",
			wantSuccess: true,
			wantStore:   "user_profile",
			wantProp:    "avatar",
		},
		{
			name:        "store with number in name",
			input:       "$auth2.token",
			wantSuccess: true,
			wantStore:   "auth2",
			wantProp:    "token",
		},
		{
			name:        "not a store expression - missing $",
			input:       "auth.isLoggedIn",
			wantSuccess: false,
			wantError:   "not a store expression",
		},
		{
			name:        "not a store expression - $ alone",
			input:       "$",
			wantSuccess: false,
			wantError:   "invalid store name",
		},
		{
			name:        "not a store expression - $ with space",
			input:       "$ auth",
			wantSuccess: false,
			wantError:   "invalid store name",
		},
		{
			name:        "store name with invalid character",
			input:       "$auth-user.name",
			wantSuccess: true,
			wantStore:   "auth",
			wantProp:    "", // Parser stops at invalid char
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := parseStoreExpression()
			result := parser(tt.input)

			if result.Successful != tt.wantSuccess {
				t.Errorf("parseStoreExpression() success = %v, want %v", result.Successful, tt.wantSuccess)
				if !tt.wantSuccess && tt.wantError != "" && result.Error != tt.wantError {
					t.Errorf("parseStoreExpression() error = %v, want %v", result.Error, tt.wantError)
				}
				return
			}

			if !tt.wantSuccess {
				return // Expected failure, no need to check node
			}

			// Check that result contains StoreExpressionNode
			node, ok := result.Value.(*ast.StoreExpressionNode)
			if !ok {
				t.Errorf("parseStoreExpression() returned type %T, want *ast.StoreExpressionNode", result.Value)
				return
			}

			if node.StoreName != tt.wantStore {
				t.Errorf("parseStoreExpression() StoreName = %v, want %v", node.StoreName, tt.wantStore)
			}

			if node.Property != tt.wantProp {
				t.Errorf("parseStoreExpression() Property = %v, want %v", node.Property, tt.wantProp)
			}
		})
	}
}

// TestStoreExpressionInTemplateExpression tests store expressions within {braces}
// Pattern: Integration test [Load: 8]
func TestStoreExpressionInTemplateExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSuccess bool
		wantStore   string
		wantProp    string
	}{
		{
			name:        "store in text expression",
			input:       "{$auth.isLoggedIn}",
			wantSuccess: true,
			wantStore:   "auth",
			wantProp:    "isLoggedIn",
		},
		{
			name:        "store with nested property in braces",
			input:       "{$user.profile.settings.theme}",
			wantSuccess: true,
			wantStore:   "user",
			wantProp:    "profile.settings.theme",
		},
		{
			name:        "store reference without property",
			input:       "{$cart}",
			wantSuccess: true,
			wantStore:   "cart",
			wantProp:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the LexExpressionParser which should extract content within braces
			result := LexExpressionParser()(tt.input)

			if !result.Successful {
				t.Errorf("LexExpressionParser() failed: %v", result.Error)
				return
			}

			expr, ok := result.Value.(string)
			if !ok {
				t.Errorf("LexExpressionParser() returned type %T, want string", result.Value)
				return
			}

			// Now parse the expression content as a store expression
			storeResult := parseStoreExpression()(expr)

			if storeResult.Successful != tt.wantSuccess {
				t.Errorf("parseStoreExpression() success = %v, want %v", storeResult.Successful, tt.wantSuccess)
				return
			}

			if !tt.wantSuccess {
				return
			}

			node, ok := storeResult.Value.(*ast.StoreExpressionNode)
			if !ok {
				t.Errorf("parseStoreExpression() returned type %T, want *ast.StoreExpressionNode", storeResult.Value)
				return
			}

			if node.StoreName != tt.wantStore {
				t.Errorf("StoreName = %v, want %v", node.StoreName, tt.wantStore)
			}

			if node.Property != tt.wantProp {
				t.Errorf("Property = %v, want %v", node.Property, tt.wantProp)
			}
		})
	}
}

// TestStoreExpressionParserEdgeCases tests edge cases and boundary conditions
// Pattern: Edge case testing [Load: 5]
func TestStoreExpressionParserEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSuccess bool
		description string
	}{
		{
			name:        "empty string",
			input:       "",
			wantSuccess: false,
			description: "Empty input should fail",
		},
		{
			name:        "dollar sign at end of string",
			input:       "text$",
			wantSuccess: false,
			description: "Dollar not at start should fail",
		},
		{
			name:        "store name starting with number",
			input:       "$123store",
			wantSuccess: false,
			description: "Store name cannot start with number",
		},
		{
			name:        "property ending with dot",
			input:       "$auth.user.",
			wantSuccess: true,
			description: "Trailing dot should be ignored",
		},
		{
			name:        "multiple consecutive dots",
			input:       "$auth..user",
			wantSuccess: true,
			description: "Multiple dots should parse storeName correctly",
		},
		{
			name:        "very long property chain",
			input:       "$app.state.user.profile.settings.theme.colors.primary.hex",
			wantSuccess: true,
			description: "Long property chains should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := parseStoreExpression()
			result := parser(tt.input)

			if result.Successful != tt.wantSuccess {
				t.Errorf("%s: success = %v, want %v", tt.description, result.Successful, tt.wantSuccess)
			}
		})
	}
}

// TestExpressionParserRouting tests that ExpressionParser correctly routes to store parser
// Pattern: Integration test [Load: 7]
// Task 1.4: Main integration tests for store expression routing
func TestExpressionParserRouting(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantNodeType string
		wantStore    string
		wantProp     string
		wantExpr     string
		description  string
	}{
		{
			name:         "regular variable expression",
			input:        "{userName}",
			wantNodeType: "Expression",
			wantExpr:     "userName",
			description:  "Regular variables should use ExpressionNode",
		},
		{
			name:         "store expression",
			input:        "{$auth.isLoggedIn}",
			wantNodeType: "StoreExpression",
			wantStore:    "auth",
			wantProp:     "isLoggedIn",
			description:  "Store expressions should use StoreExpressionNode",
		},
		{
			name:         "complex expression",
			input:        "{count + 1}",
			wantNodeType: "Expression",
			wantExpr:     "count + 1",
			description:  "Complex expressions should use ExpressionNode",
		},
		{
			name:         "store without property",
			input:        "{$cart}",
			wantNodeType: "StoreExpression",
			wantStore:    "cart",
			wantProp:     "",
			description:  "Store reference alone should work",
		},
		{
			name:         "store with nested property",
			input:        "{$user.profile.name}",
			wantNodeType: "StoreExpression",
			wantStore:    "user",
			wantProp:     "profile.name",
			description:  "Nested store property access should work",
		},
		{
			name:         "object property access (not store)",
			input:        "{user.name}",
			wantNodeType: "Expression",
			wantExpr:     "user.name",
			description:  "Regular object property access without $ should be ExpressionNode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the main ExpressionParser which should route correctly
			result := ExpressionParser()(tt.input)

			if !result.Successful {
				t.Errorf("%s: ExpressionParser() failed: %v", tt.description, result.Error)
				return
			}

			// Check node type
			switch tt.wantNodeType {
			case "Expression":
				node, ok := result.Value.(*ast.ExpressionNode)
				if !ok {
					t.Errorf("%s: got node type %T, want *ast.ExpressionNode", tt.description, result.Value)
					return
				}
				if node.Expression != tt.wantExpr {
					t.Errorf("%s: Expression = %v, want %v", tt.description, node.Expression, tt.wantExpr)
				}

			case "StoreExpression":
				node, ok := result.Value.(*ast.StoreExpressionNode)
				if !ok {
					t.Errorf("%s: got node type %T, want *ast.StoreExpressionNode", tt.description, result.Value)
					return
				}
				if node.StoreName != tt.wantStore {
					t.Errorf("%s: StoreName = %v, want %v", tt.description, node.StoreName, tt.wantStore)
				}
				if node.Property != tt.wantProp {
					t.Errorf("%s: Property = %v, want %v", tt.description, node.Property, tt.wantProp)
				}

			default:
				t.Errorf("Unknown node type: %s", tt.wantNodeType)
			}
		})
	}
}

// TestStoreExpressionInTextContent tests store expressions in text content
// Pattern: Integration test [Load: 8]
// Task 1.4: Test store expressions work in text content
func TestStoreExpressionInTextContent(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		wantStoreRef bool
		wantStore    string
		wantProp     string
		description  string
	}{
		{
			name:         "store in paragraph text",
			template:     "<p>{$auth.user.name}</p>",
			wantStoreRef: true,
			wantStore:    "auth",
			wantProp:     "user.name",
			description:  "Store expression in paragraph should parse correctly",
		},
		{
			name:         "mixed regular and store expressions",
			template:     "<p>{title}: {$auth.user.name}</p>",
			wantStoreRef: true,
			wantStore:    "auth",
			wantProp:     "user.name",
			description:  "Mixed expressions should both parse correctly",
		},
		{
			name:         "store in heading",
			template:     "<h1>Welcome {$auth.user.name}</h1>",
			wantStoreRef: true,
			wantStore:    "auth",
			wantProp:     "user.name",
			description:  "Store expression in heading should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the full template
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Errorf("%s: ParseTemplate() error = %v", tt.description, err)
				return
			}

			// Find StoreExpressionNode in the AST
			foundStore := false
			var storeNode *ast.StoreExpressionNode

			var findStore func(ast.Node)
			findStore = func(node ast.Node) {
				if node == nil {
					return
				}

				switch n := node.(type) {
				case *ast.StoreExpressionNode:
					foundStore = true
					storeNode = n
				case *ast.Element:
					for _, child := range n.Children {
						findStore(child)
					}
				case *ast.Template:
					for _, root := range n.RootNodes {
						findStore(root)
					}
				}
			}

			findStore(tmpl)

			if tt.wantStoreRef && !foundStore {
				t.Errorf("%s: expected store reference but none found", tt.description)
				return
			}

			if tt.wantStoreRef && foundStore {
				if storeNode.StoreName != tt.wantStore {
					t.Errorf("%s: StoreName = %v, want %v", tt.description, storeNode.StoreName, tt.wantStore)
				}
				if storeNode.Property != tt.wantProp {
					t.Errorf("%s: Property = %v, want %v", tt.description, storeNode.Property, tt.wantProp)
				}
			}
		})
	}
}

// TestBackwardCompatibility tests that existing expression parsing still works
// Pattern: Regression test [Load: 6]
// Task 1.4: Ensure no regressions with existing functionality
func TestBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		description string
	}{
		{
			name:        "simple variable",
			template:    "<p>{userName}</p>",
			description: "Simple variable should still parse as ExpressionNode",
		},
		{
			name:        "object property",
			template:    "<p>{user.name}</p>",
			description: "Object property access should still work",
		},
		{
			name:        "array access",
			template:    "<p>{items.length}</p>",
			description: "Array property should still work",
		},
		{
			name:        "complex expression",
			template:    "<p>{count + 1}</p>",
			description: "Math expressions should still work",
		},
		{
			name:        "ternary expression",
			template:    `<p>{isLoggedIn ? 'Welcome' : 'Login'}</p>`,
			description: "Ternary expressions should still work",
		},
		{
			name: "if with regular variable",
			template: `{if isLoggedIn}
				<p>Welcome</p>
			{/if}`,
			description: "Conditionals with regular variables should work",
		},
		{
			name: "for with regular array",
			template: `{for item in items}
				<div>{item}</div>
			{/for}`,
			description: "Loops with regular arrays should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse template
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Errorf("%s: ParseTemplate() error = %v (expected success for backward compatibility)", tt.description, err)
				return
			}

			// Just verify it parsed successfully
			if tmpl == nil {
				t.Errorf("%s: got nil template", tt.description)
			}
		})
	}
}

// TestStoreExpressionInConditionals tests that store expressions in conditionals are stored as strings
// Pattern: Integration test [Load: 7]
// Task 1.4: Store expressions in conditionals are stored as condition strings (transformation happens in Phase 2)
func TestStoreExpressionInConditionals(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantCond    string
		description string
	}{
		{
			name: "store in if condition",
			template: `{if $auth.isLoggedIn}
				<p>Welcome!</p>
			{/if}`,
			wantCond:    "$auth.isLoggedIn",
			description: "Store in if condition should be stored as string",
		},
		{
			name: "store in else-if condition",
			template: `{if isAdmin}
				<p>Admin</p>
			{else if $auth.isLoggedIn}
				<p>User</p>
			{/if}`,
			wantCond:    "$auth.isLoggedIn",
			description: "Store in else-if should be stored as string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the full template
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Errorf("%s: ParseTemplate() error = %v", tt.description, err)
				return
			}

			// Find Conditional in the AST and check condition string
			foundCond := false

			var findCond func(ast.Node)
			findCond = func(node ast.Node) {
				if node == nil {
					return
				}

				switch n := node.(type) {
				case *ast.Conditional:
					// Check IfCondition or ElseIfConditions
					if n.IfCondition == tt.wantCond {
						foundCond = true
					}
					for _, cond := range n.ElseIfConditions {
						if cond == tt.wantCond {
							foundCond = true
						}
					}
				case *ast.Element:
					for _, child := range n.Children {
						findCond(child)
					}
				case *ast.Template:
					for _, root := range n.RootNodes {
						findCond(root)
					}
				}
			}

			findCond(tmpl)

			if !foundCond {
				t.Errorf("%s: expected condition string %q but not found", tt.description, tt.wantCond)
			}
		})
	}
}

// TestStoreExpressionInLoops tests that store expressions in loops are stored as strings
// Pattern: Integration test [Load: 7]
// Task 1.4: Store expressions in loops are stored as collection strings (transformation happens in Phase 2)
func TestStoreExpressionInLoops(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantColl    string
		description string
	}{
		{
			name: "store in for loop collection",
			template: `{for item in $cart.items}
				<div>{item.name}</div>
			{/for}`,
			wantColl:    "$cart.items",
			description: "Store in loop collection should be stored as string",
		},
		{
			name: "store in nested loop",
			template: `{for category in $store.categories}
				{for item in category.items}
					<div>{item}</div>
				{/for}
			{/for}`,
			wantColl:    "$store.categories",
			description: "Store in outer loop should be stored as string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the full template
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Errorf("%s: ParseTemplate() error = %v", tt.description, err)
				return
			}

			// Find Loop in the AST and check collection string
			foundLoop := false

			var findLoop func(ast.Node)
			findLoop = func(node ast.Node) {
				if node == nil {
					return
				}

				switch n := node.(type) {
				case *ast.Loop:
					// Check Collection string
					if n.Collection == tt.wantColl {
						foundLoop = true
					}
					// Check loop content
					for _, child := range n.Content {
						findLoop(child)
					}
				case *ast.Element:
					for _, child := range n.Children {
						findLoop(child)
					}
				case *ast.Template:
					for _, root := range n.RootNodes {
						findLoop(root)
					}
				}
			}

			findLoop(tmpl)

			if !foundLoop {
				t.Errorf("%s: expected collection string %q but not found", tt.description, tt.wantColl)
			}
		})
	}
}
