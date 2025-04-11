package transformer

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// transformTextWithExpressions transforms text containing expressions like {name} or {name }
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

	// Combine and sort all matches by their start position
	allMatches := append(doubleMatches, filteredSingleMatches...)
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i][0] < allMatches[j][0]
	})

	// If no expressions found, return the original text node
	if len(allMatches) == 0 {
		return []ast.Node{&ast.TextNode{Content: text}
	}

	// Process the text with expressions
	var result []ast.Node
	lastIndex := 0

	for _, match := range allMatches {
		// Add text before the expression
		if match[0] > lastIndex {
			beforeText := text[lastIndex:match[0]]
			if beforeText != "" {
				result = append(result, &ast.TextNode{Content: beforeText})
			}
		}

		// Extract the expression without braces
		var expr string

		// Check if this is a double brace expression
		if strings.HasPrefix(text[match[0]:match[1]], "{") {
			// Double brace expression - extract and trim more aggressively
			expr = text[match[2]:match[3]]
			expr = strings.TrimSpace(expr)
		} else {
			// Single brace expression
			expr = text[match[2]:match[3]]
			expr = strings.TrimSpace(expr)
		}

		// Add variables from the expression to the data scope
		AddExprVarsToScope(expr, dataScope)

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

	// Add any remaining text after the last expression
	if lastIndex < len(text) {
		afterText := text[lastIndex:]
		if afterText != "" {
			result = append(result, &ast.TextNode{Content: afterText})
		}
	}

	return result
}

// isNumeric checks if a string is a numeric literal
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// isJSKeyword checks if a string is a JavaScript keyword
func isJSKeyword(s string) bool {
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

// formatDataObject converts the data scope to a JavaScript object literal for Alpine.js x-data
func formatDataObject(dataScope map[string]any) string {
	// For Alpine.js, we need to format the data as a JavaScript object literal
	// rather than a JSON string, as Alpine expects the x-data attribute to contain
	// valid JavaScript object syntax

	if len(dataScope) == 0 {
		return "{}"
	}

	// For complex object structures that might contain methods, prefer a multiline format
	// which is more compatible with Alpine.js
	var builder strings.Builder
	builder.WriteString("{\n")

	// Track if we need to add a comma
	needsComma := false

	// Build a JavaScript object literal string
	for key, value := range dataScope {
		// Add comma if needed
		if needsComma {
			builder.WriteString(",\n")
		}
		needsComma = true

		// Format the key
		builder.WriteString("  ")
		// Check if the key needs quotes
		if needsQuotes(key) {
			builder.WriteString(fmt.Sprintf("'%s'", escapeJSString(key)))
		} else {
			builder.WriteString(key)
		}
		builder.WriteString(": ")

		// Format the value
		if strValue, ok := value.(string); ok {
			// Special handling for string values that might be JavaScript expressions
			if isMethodDefinition(strValue) {
				// For method definitions, don't add quotes and clean up the syntax
				cleanMethod := cleanupMethodDefinition(strValue)
				builder.WriteString(cleanMethod)
			} else if isJSExpression(strValue) {
				// For JS expressions, preserve them as-is
				cleanExpr := strings.TrimSuffix(strings.TrimSpace(strValue), ";")
				builder.WriteString(cleanExpr)
			} else {
				// Regular string value
				builder.WriteString("'")
				builder.WriteString(escapeJSString(strValue))
				builder.WriteString("'")
			}
		} else {
			// Non-string values
			formattedValue := formatValue(value)
			builder.WriteString(formattedValue)
		}
	}

	builder.WriteString("\n}")
	return builder.String()
}

// needsQuotes checks if a property key needs quotes in JavaScript
func needsQuotes(key string) bool {
	// JavaScript identifier pattern
	identifierPattern := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)
	return !identifierPattern.MatchString(key)
}

// cleanupMethodDefinition removes trailing semicolons and ensures proper formatting for object methods
func cleanupMethodDefinition(method string) string {
	// Remove trailing semicolons
	method = strings.TrimSpace(method)
	method = strings.TrimSuffix(method, ";")

	// Check for the arrow function with braces pattern
	if strings.Contains(method, "=>") && strings.Contains(method, "{") {
		// Make sure there's proper spacing around the arrow
		arrowPattern := regexp.MustCompile(`\)\s*=>\s*{`)
		if !arrowPattern.MatchString(method) {
			method = regexp.MustCompile(`\)=>\{`).ReplaceAllString(method, ") => {")
		}
	}

	return method
}

// isMethodDefinition checks if a string looks like a JavaScript method definition
func isMethodDefinition(s string) bool {
	// Look for patterns like "function(...) {...}" or "(...) => {...}" or "name(...) {...}"
	s = strings.TrimSpace(s)

	// Check for arrow function: (...) => {...}
	if strings.Contains(s, "=>") {
		return true
	}

	// Check for function keyword: function(...) {...}
	if strings.HasPrefix(s, "function") {
		return true
	}

	// Check for method shorthand: name(...) {...}
	methodPattern := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*\s*\(.*\)\s*\{`)
	if methodPattern.MatchString(s) {
		return true
	}

	// Check for getter/setter syntax: get prop() {...} or set prop(...) {...}
	getterSetterPattern := regexp.MustCompile(`^(get|set)\s+[a-zA-Z_$][a-zA-Z0-9_$]*\s*\(.*\)\s*\{`)
	if getterSetterPattern.MatchString(s) {
		return true
	}

	// Check for async methods: async function(...) or async name(...)
	if strings.HasPrefix(s, "async ") {
		return true
	}

	return false
}

// formatValue formats a value for use in a JavaScript object literal
func formatValue(value any) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		// Check if it's a method definition
		if isMethodDefinition(v) {
			// Remove any trailing semicolons for method definitions in object literals
			return cleanupMethodDefinition(v)
		}

		// Check if it's a JavaScript expression that should be preserved
		// This includes object literals, array literals, and other JS expressions
		if isJSExpression(v) {
			// Remove any trailing semicolons for JS expressions in object literals
			return strings.TrimSuffix(strings.TrimSpace(v), ";")
		}

		// Escape quotes and wrap in quotes
		escaped := escapeJSString(v)
		return "'" + escaped + "'"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// Use the default string representation for numbers
		return stringify(v)
	case []any:
		// Format arrays
		var items []string
		for _, item := range v {
			items = append(items, formatValue(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		// Format nested objects
		var pairs []string
		for k, val := range v {
			// Check if the value is a method definition
			if strVal, ok := val.(string); ok && isMethodDefinition(strVal) {
				// Remove any trailing semicolons for method definitions
				cleanMethod := cleanupMethodDefinition(strVal)
				pairs = append(pairs, k+": "+cleanMethod)
			} else {
				pairs = append(pairs, k+": "+formatValue(val))
			}
		}
		return "{ " + strings.Join(pairs, ", ") + " }"
	default:
		// Try to convert to JSON as a fallback
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			log.Printf("Error marshaling value: %v", err)
			return "null /* Error serializing value */"
		}
		// For complex objects, return the JSON string
		jsonStr := string(jsonBytes)
		// Convert double quotes to single quotes for better Alpine.js compatibility
		jsonStr = strings.ReplaceAll(jsonStr, "\"", "'")
		return jsonStr
	}
}

// isJSExpression checks if a string appears to be a JavaScript expression that should be preserved without quotes
func isJSExpression(s string) bool {
	s = strings.TrimSpace(s)

	// Check for object literals
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return true
	}

	// Check for array literals
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return true
	}

	// Check for function expressions (already covered by isMethodDefinition, but for explicitness)
	if strings.HasPrefix(s, "function(") || strings.Contains(s, "=>") {
		return true
	}

	// Check for new operator
	if strings.HasPrefix(s, "new ") {
		return true
	}

	// Check for ternary operators
	if strings.Contains(s, "?") && strings.Contains(s, ":") {
		return true
	}

	return false
}

// escapeJSString escapes a string for use in JavaScript
func escapeJSString(s string) string {
	// Replace backslashes first
	s = strings.ReplaceAll(s, "\\", "\\\\")

	// Replace quotes
	s = strings.ReplaceAll(s, "'", "\\'")

	// Replace newlines and other special characters
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")

	return s
}

// stringify converts a value to its string representation
func stringify(value any) string {
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}
