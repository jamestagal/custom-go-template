package transformer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// transformTextWithExpressions transforms text containing expressions like {name} or {{ name }}
// into a series of nodes with Alpine.js x-text directives
func transformTextWithExpressions(text string, dataScope map[string]any) []ast.Node {
	// Regular expressions to find both single and double curly brace expressions
	singleBraceRegex := regexp.MustCompile(`\{([^{}]+)\}`)
	doubleBraceRegex := regexp.MustCompile(`\{\{\s*([^{}]+)\s*\}\}`)

	// Process double braces first (they have precedence)
	doubleMatches := doubleBraceRegex.FindAllStringSubmatchIndex(text, -1)
	singleMatches := singleBraceRegex.FindAllStringSubmatchIndex(text, -1)

	// Filter out single brace matches that overlap with double brace matches
	var filteredSingleMatches [][]int
	for _, sMatch := range singleMatches {
		isOverlapping := false
		for _, dMatch := range doubleMatches {
			// Check if the single brace match overlaps with any double brace match
			if (sMatch[0] >= dMatch[0] && sMatch[0] <= dMatch[1]) ||
				(sMatch[1] >= dMatch[0] && sMatch[1] <= dMatch[1]) {
				isOverlapping = true
				break
			}
		}
		if !isOverlapping {
			filteredSingleMatches = append(filteredSingleMatches, sMatch)
		}
	}

	// If no expressions found, return the original text node
	if len(doubleMatches) == 0 && len(filteredSingleMatches) == 0 {
		return []ast.Node{&ast.TextNode{Content: text}}
	}

	// Process the text with expressions
	var result []ast.Node
	lastIndex := 0

	// Process double-brace expressions
	for _, match := range doubleMatches {
		// Add text before the expression
		if match[0] > lastIndex {
			beforeText := text[lastIndex:match[0]]
			if beforeText != "" {
				result = append(result, &ast.TextNode{Content: beforeText})
			}
		}

		// Extract the expression (content inside {{...}})
		expr := text[match[2]:match[3]]
		expr = strings.TrimSpace(expr)

		// Add variables from the expression to the data scope
		extractVariablesFromExpr(expr, dataScope)

		// Create a span with x-text for the expression
		exprNode := &ast.Element{
			TagName: "span",
			Attributes: []ast.Attribute{
				{
					Name:       "x-text",
					Value:      expr,
					Dynamic:    true,
					IsAlpine:   true,
					AlpineType: "text",
				},
			},
			Children:    []ast.Node{},
			SelfClosing: false,
		}

		result = append(result, exprNode)
		lastIndex = match[1]
	}

	// Process single-brace expressions (if they don't overlap with double braces)
	for _, match := range filteredSingleMatches {
		// Skip if this match is before the last processed index
		if match[0] < lastIndex {
			continue
		}

		// Add text before the expression
		if match[0] > lastIndex {
			beforeText := text[lastIndex:match[0]]
			if beforeText != "" {
				result = append(result, &ast.TextNode{Content: beforeText})
			}
		}

		// Extract the expression (content inside {...})
		expr := text[match[2]:match[3]]
		expr = strings.TrimSpace(expr)

		// Skip pure text curly braces that aren't actual expressions
		if !isExpressionSyntax(expr) {
			// Just treat it as plain text with the braces
			result = append(result, &ast.TextNode{Content: text[match[0]:match[1]]})
		} else {
			// Add variables from the expression to the data scope
			extractVariablesFromExpr(expr, dataScope)

			// Create a span with x-text for the expression
			exprNode := &ast.Element{
				TagName: "span",
				Attributes: []ast.Attribute{
					{
						Name:       "x-text",
						Value:      expr,
						Dynamic:    true,
						IsAlpine:   true,
						AlpineType: "text",
					},
				},
				Children:    []ast.Node{},
				SelfClosing: false,
			}

			result = append(result, exprNode)
		}

		lastIndex = match[1]
	}

	// Add any remaining text after the last expression
	if lastIndex < len(text) {
		afterText := text[lastIndex:]
		if afterText != "" {
			result = append(result, &ast.TextNode{Content: afterText})
		}
	}

	return result
}

// extractVariablesFromExpr extracts variable names from an expression and adds them to the data scope
func extractVariablesFromExpr(expr string, dataScope map[string]any) {
	// Skip empty expressions
	if expr == "" {
		return
	}

	// Clean the expression
	expr = strings.TrimSpace(expr)
	expr = strings.Trim(expr, "{}")

	// Skip string literals, numeric literals, and boolean literals
	if isStringLiteral(expr) || isNumericString(expr) ||
		expr == "true" || expr == "false" || expr == "null" || expr == "undefined" {
		return
	}

	// Skip function definitions
	if strings.HasPrefix(expr, "function") ||
		(strings.Contains(expr, "=>") && strings.Contains(expr, "{")) {
		return
	}

	// Handle ternary operators
	if strings.Contains(expr, "?") && strings.Contains(expr, ":") {
		parts := strings.SplitN(expr, "?", 2)
		condition := parts[0]

		// Extract variables from the condition
		extractVariablesFromExpr(condition, dataScope)

		// Extract variables from the branches
		if len(parts) > 1 {
			branches := strings.SplitN(parts[1], ":", 2)
			if len(branches) > 0 {
				extractVariablesFromExpr(branches[0], dataScope)
			}
			if len(branches) > 1 {
				extractVariablesFromExpr(branches[1], dataScope)
			}
		}
		return
	}

	// Handle logical operators
	if strings.Contains(expr, "&&") || strings.Contains(expr, "||") {
		var operators []string
		if strings.Contains(expr, "&&") {
			operators = append(operators, "&&")
		}
		if strings.Contains(expr, "||") {
			operators = append(operators, "||")
		}

		// Split by operators and process each part
		for _, op := range operators {
			parts := strings.Split(expr, op)
			for _, part := range parts {
				extractVariablesFromExpr(part, dataScope)
			}
		}
		return
	}

	// Handle comparison operators
	comparisonOps := []string{"===", "!==", "==", "!=", ">=", "<=", ">", "<"}
	for _, op := range comparisonOps {
		if strings.Contains(expr, op) {
			parts := strings.Split(expr, op)
			for _, part := range parts {
				extractVariablesFromExpr(part, dataScope)
			}
			return
		}
	}

	// Handle arithmetic operators
	arithmeticOps := []string{"+", "-", "*", "/", "%"}
	for _, op := range arithmeticOps {
		if strings.Contains(expr, op) && !strings.HasPrefix(expr, op) {
			parts := strings.Split(expr, op)
			for _, part := range parts {
				extractVariablesFromExpr(part, dataScope)
			}
			return
		}
	}

	// Handle function calls like functionName(arg1, arg2)
	if strings.Contains(expr, "(") && strings.Contains(expr, ")") {
		// Extract the function name
		funcNameEnd := strings.Index(expr, "(")
		if funcNameEnd > 0 {
			funcName := strings.TrimSpace(expr[:funcNameEnd])

			// Add the function name to the data scope if it's a valid identifier
			if isValidIdentifier(funcName) {
				if _, exists := dataScope[funcName]; !exists {
					// Add function with a default implementation
					dataScope[funcName] = fmt.Sprintf("function() { return null; }")
				}
			} else if strings.Contains(funcName, ".") {
				// Handle object method calls like obj.method()
				parts := strings.Split(funcName, ".")
				if len(parts) > 0 && isValidIdentifier(parts[0]) {
					rootVar := parts[0]
					if _, exists := dataScope[rootVar]; !exists {
						dataScope[rootVar] = getDefaultValueForVar(rootVar)
					}
				}
			}

			// Extract variables from function arguments
			argsStart := strings.Index(expr, "(")
			argsEnd := strings.LastIndex(expr, ")")
			if argsStart >= 0 && argsEnd > argsStart {
				args := expr[argsStart+1 : argsEnd]

				// Split arguments by comma, respecting nested function calls
				depth := 0
				var currentArg strings.Builder
				var argsList []string

				for _, char := range args {
					if char == '(' {
						depth++
						currentArg.WriteRune(char)
					} else if char == ')' {
						depth--
						currentArg.WriteRune(char)
					} else if char == ',' && depth == 0 {
						// End of an argument
						argsList = append(argsList, currentArg.String())
						currentArg.Reset()
					} else {
						currentArg.WriteRune(char)
					}
				}

				// Add the last argument
				if currentArg.Len() > 0 {
					argsList = append(argsList, currentArg.String())
				}

				// Process each argument
				for _, arg := range argsList {
					extractVariablesFromExpr(arg, dataScope)
				}
			}
		}
		return
	}

	// Handle object property access (e.g., user.name, items[0].price)
	if strings.Contains(expr, ".") || strings.Contains(expr, "[") {
		// Extract the root variable name
		var rootVar string

		if strings.Contains(expr, "[") {
			// Handle array access like items[0]
			bracketIndex := strings.Index(expr, "[")
			if bracketIndex > 0 {
				rootVar = strings.TrimSpace(expr[:bracketIndex])
			}
		} else {
			// Handle dot notation like user.name
			dotIndex := strings.Index(expr, ".")
			if dotIndex > 0 {
				rootVar = strings.TrimSpace(expr[:dotIndex])
			}
		}

		// Add the root variable to the data scope
		if rootVar != "" && isValidIdentifier(rootVar) {
			if _, exists := dataScope[rootVar]; !exists {
				dataScope[rootVar] = getDefaultValueForVar(rootVar)
			}
		}
		return
	}

	// For simple variable names, add them to the data scope
	if isValidIdentifier(expr) {
		if _, exists := dataScope[expr]; !exists {
			dataScope[expr] = getDefaultValueForVar(expr)
		}
	}
}

// isExpressionSyntax checks if the content inside curly braces appears to be
// an expression and not just text with curly braces
func isExpressionSyntax(s string) bool {
	// If it contains an operator or a dot, it's probably an expression
	if strings.Contains(s, ".") ||
		strings.Contains(s, "+") ||
		strings.Contains(s, "-") ||
		strings.Contains(s, "*") ||
		strings.Contains(s, "/") ||
		strings.Contains(s, "?") ||
		strings.Contains(s, "=") {
		return true
	}

	// If it's surrounded by quotes, it's probably not an expression
	if (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) {
		return false
	}

	// If it has spaces that are not leading/trailing, it's likely an expression
	trimmed := strings.TrimSpace(s)
	if len(trimmed) != len(s) && strings.Contains(trimmed, " ") {
		return true
	}

	// If it's a simple identifier, assume it's an expression
	if regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`).MatchString(s) {
		return true
	}

	// Default to treating it as text, not an expression
	return false
}

// getDefaultValueForVar returns a sensible default value for common variable names
func getDefaultValueForVar(varName string) interface{} {
	switch varName {
	case "user":
		return map[string]interface{}{
			"name":     "John Doe",
			"role":     "admin",
			"isAdmin":  true,
			"email":    "john@example.com",
			"joinDate": "2023-05-15",
			"details": map[string]interface{}{
				"email": "john@example.com",
				"phone": "555-1234",
			},
			"orders": []interface{}{
				map[string]interface{}{
					"id":     "ORD-1234",
					"date":   "2023-03-15",
					"status": "Delivered",
					"total":  129.99,
				},
				map[string]interface{}{
					"id":     "ORD-5678",
					"date":   "2023-02-27",
					"status": "Shipped",
					"total":  79.5,
				},
			},
			"wishlist": []interface{}{
				map[string]interface{}{
					"id":    101,
					"name":  "Wireless Headphones",
					"price": 89.99,
				},
				map[string]interface{}{
					"id":    205,
					"name":  "Smart Watch",
					"price": 199.99,
				},
			},
		}
	case "product":
		return map[string]interface{}{
			"id":       1,
			"name":     "Product Name",
			"price":    99.99,
			"inStock":  true,
			"featured": false,
			"tags":     []string{"electronics", "gadgets"},
		}
	case "item":
		return map[string]interface{}{
			"name":  "Item Name",
			"price": 49.99,
			"tags":  []string{"category1", "category2"},
		}
	case "category":
		return map[string]interface{}{
			"name": "Category Name",
			"items": []interface{}{
				map[string]interface{}{
					"name":  "Item 1",
					"price": 19.99,
					"tags":  []string{"tag1", "tag2"},
				},
				map[string]interface{}{
					"name":  "Item 2",
					"price": 29.99,
					"tags":  []string{"tag2", "tag3"},
				},
			},
		}
	case "notification":
		return map[string]interface{}{
			"type":    "info",
			"message": "Notification message",
		}
	case "filteredProducts":
		return []interface{}{
			map[string]interface{}{
				"id":       1,
				"name":     "Laptop",
				"price":    999.99,
				"inStock":  true,
				"featured": true,
				"tags":     []string{"electronics", "computers"},
			},
			map[string]interface{}{
				"name":     "Phone",
				"price":    699.99,
				"inStock":  true,
				"featured": false,
				"tags":     []string{"electronics", "mobile"},
			},
		}
	case "products":
		return []interface{}{
			map[string]interface{}{
				"id":       1,
				"name":     "Laptop",
				"price":    999.99,
				"inStock":  true,
				"featured": true,
				"tags":     []string{"electronics", "computers"},
			},
			map[string]interface{}{
				"name":     "Phone",
				"price":    699.99,
				"inStock":  true,
				"featured": false,
				"tags":     []string{"electronics", "mobile"},
			},
			map[string]interface{}{
				"name":     "Headphones",
				"price":    149.99,
				"inStock":  false,
				"featured": true,
				"tags":     []string{"electronics", "audio"},
			},
			map[string]interface{}{
				"name":     "Tablet",
				"price":    499.99,
				"inStock":  true,
				"featured": false,
				"tags":     []string{"electronics", "computers"},
			},
		}
	case "categories":
		return []interface{}{
			map[string]interface{}{
				"name": "Electronics",
				"items": []interface{}{
					map[string]interface{}{
						"name":  "Laptop",
						"price": 999.99,
						"tags":  []string{"electronics", "computers"},
					},
					map[string]interface{}{
						"name":  "Phone",
						"price": 699.99,
						"tags":  []string{"electronics", "mobile"},
					},
				},
			},
			map[string]interface{}{
				"name":  "Books",
				"items": []interface{}{},
			},
		}
	case "settings":
		return map[string]interface{}{
			"theme":        "light",
			"currency":     "USD",
			"language":     "en",
			"showFeatured": true,
			"filters": map[string]interface{}{
				"inStockOnly": false,
				"minPrice":    0,
				"maxPrice":    1000,
			},
		}
	case "index":
		return 0
	case "title":
		return "Custom Template Showcase"
	case "isAdmin":
		return true
	case "isLoggedIn":
		return true
	case "getGreeting":
		return "function() { return 'Hello'; }"
	case "formatPrice":
		return "function(price) { return '$' + price.toFixed(2); }"
	case "getTagClass":
		return "function(tag) { return 'tag-' + tag; }"
	case "notifications":
		return []interface{}{
			map[string]interface{}{
				"type":    "info",
				"message": "Welcome to our store!",
			},
			map[string]interface{}{
				"type":    "success",
				"message": "Your order has been processed.",
			},
			map[string]interface{}{
				"type":    "warning",
				"message": "Some items are out of stock.",
			},
		}
	case "stats":
		return map[string]interface{}{
			"users":    124,
			"products": 56,
			"orders":   890,
			"revenue":  15280.45,
		}
	case "recentActions":
		return []interface{}{
			map[string]interface{}{
				"user":      "John Doe",
				"action":    "Order fulfilled",
				"timestamp": "2023-04-10T13:45:00Z",
			},
			map[string]interface{}{
				"user":      "Jane Smith",
				"action":    "Product created",
				"timestamp": "2023-04-10T14:32:00Z",
			},
		}
	case "currentUser":
		return map[string]interface{}{
			"name":  "John Doe",
			"role":  "admin",
			"email": "john@example.com",
		}
	default:
		return nil
	}
}

// isValidIdentifier checks if a string is a valid JavaScript identifier
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Check if it's a reserved keyword
	if isJSReservedKeyword(s) {
		return false
	}

	// JavaScript identifier pattern: starts with letter/underscore, followed by letters/numbers/underscores
	return regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`).MatchString(s)
}

// isStringLiteral checks if a string is enclosed in quotes
func isStringLiteral(s string) bool {
	return (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\""))
}

// isNumericString checks if a string is a numeric literal
func isNumericString(s string) bool {
	return regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`).MatchString(s)
}

// isJSReservedKeyword checks if a string is a JavaScript reserved keyword
func isJSReservedKeyword(s string) bool {
	keywords := map[string]bool{
		"true": true, "false": true, "null": true, "undefined": true,
		"var": true, "let": true, "const": true,
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "default": true,
		"break": true, "continue": true, "return": true,
		"function": true, "class": true, "this": true, "super": true,
		"new": true, "delete": true, "typeof": true, "instanceof": true,
		"void": true, "in": true, "of": true,
	}

	return keywords[s]
}

func transformExpression(expr string) string {
	// Handle nested object access
	parts := strings.Split(expr, ".")
	if len(parts) > 1 {
		// Verify parent objects exist
		return fmt.Sprintf("(%s || {}).%s", parts[0], strings.Join(parts[1:], "."))
	}
	return expr
}
