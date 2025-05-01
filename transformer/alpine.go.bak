package transformer

import (
	"encoding/json"
	"log"
	"sort"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// alpineDataFormatter formats the data scope as JSON for Alpine.js x-data attribute
func alpineDataFormatter(dataScope map[string]any) string {
	// Filter out null values
	filteredData := make(map[string]any)
	for k, v := range dataScope {
		if v != nil {
			filteredData[k] = v
		}
	}
	
	// Sort keys to ensure consistent order
	keys := make([]string, 0, len(filteredData))
	for k := range filteredData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Build ordered map
	orderedData := make(map[string]any)
	for _, k := range keys {
		orderedData[k] = filteredData[k]
	}
	
	// Use specific JSON formatting for Alpine compatibility
	jsonBytes, err := json.Marshal(orderedData)
	if err != nil {
		log.Printf("Error marshaling alpine data: %v", err)
		return "{}"
	}
	
	// Use single quotes for Alpine data to avoid HTML attribute escaping issues
	return "'" + string(jsonBytes) + "'"
}

// isFunctionExpression checks if a string appears to be a JavaScript function
func isFunctionExpression(str string) bool {
	str = strings.TrimSpace(str)

	// Check for arrow functions
	if strings.Contains(str, "=>") {
		return true
	}

	// Check for function expressions
	if strings.HasPrefix(str, "function") {
		return true
	}

	// Check for method shorthand
	if strings.Contains(str, "(") && strings.Contains(str, ")") && strings.Contains(str, "{") {
		return true
	}

	return false
}

// escapeJSONString escapes a string for JSON while preserving certain characters
func escapeJSONString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// wrapWithAlpineData wraps the given nodes with an Alpine.js x-data wrapper
// This is a more robust implementation than the previous createAlpineWrapper
func wrapWithAlpineData(nodes []ast.Node, dataScope map[string]any) *ast.Element {
	// First, extract any fence sections from the nodes
	var processedNodes []ast.Node

	for _, node := range nodes {
		if textNode, isText := node.(*ast.TextNode); isText {
			content := textNode.Content

			// Check if this text node contains a fence section
			if strings.HasPrefix(strings.TrimSpace(content), "---") && strings.Contains(content, "---\n") {
				// Extract the fence content between the triple dashes
				parts := strings.Split(content, "---")
				if len(parts) >= 3 {
					fenceContent := strings.TrimSpace(parts[1])

					// Parse variable declarations from the fence content
					lines := strings.Split(fenceContent, "\n")
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
								if isFunctionExpression(varValue) {
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

					// Skip adding this fence section node to the processed nodes
					continue
				}
			}
		}

		// Add the node to the processed list
		processedNodes = append(processedNodes, node)
	}

	// Format the data scope as JSON
	dataJSON := alpineDataFormatter(dataScope)

	// Create the wrapper element
	wrapper := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:     "x-data",
				Value:    dataJSON,
				IsAlpine: true,
				AlpineType: "data",
			},
		},
		Children:    processedNodes,
		SelfClosing: false,
	}

	return wrapper
}

// parseSimpleObject parses a simple JavaScript object into a map
func parseSimpleObject(objStr string) map[string]any {
	result := make(map[string]any)

	// Remove the outer braces
	content := strings.TrimSpace(objStr[1 : len(objStr)-1])

	// Split by commas, but respect nested objects
	pairs := splitRespectingBraces(content, ',')

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
			} else if isNumericValue(value) {
				// Try to parse as number
				if strings.Contains(value, ".") {
					// Float
					result[key] = parseFloat(value)
				} else {
					// Integer
					result[key] = parseInt(value)
				}
			} else if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				// String
				result[key] = strings.Trim(value, "\"")
			} else if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				// String with single quotes
				result[key] = strings.Trim(value, "'")
			} else if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
				// Nested object
				result[key] = parseSimpleObject(value)
			} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				// Array
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

	// Split by commas, but respect nested structures
	elements := splitRespectingBraces(content, ',')

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
		} else if isNumericValue(elem) {
			// Try to parse as number
			if strings.Contains(elem, ".") {
				// Float
				result = append(result, parseFloat(elem))
			} else {
				// Integer
				result = append(result, parseInt(elem))
			}
		} else if strings.HasPrefix(elem, "\"") && strings.HasSuffix(elem, "\"") {
			// String
			result = append(result, strings.Trim(elem, "\""))
		} else if strings.HasPrefix(elem, "'") && strings.HasSuffix(elem, "'") {
			// String with single quotes
			result = append(result, strings.Trim(elem, "'"))
		} else if strings.HasPrefix(elem, "{") && strings.HasSuffix(elem, "}") {
			// Object
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

// splitRespectingBraces splits a string by a delimiter while respecting braces and brackets
func splitRespectingBraces(s string, delimiter rune) []string {
	var result []string
	var current strings.Builder

	braceLevel := 0
	bracketLevel := 0
	inQuotes := false
	quoteChar := rune(0)

	for _, ch := range s {
		if !inQuotes {
			if ch == '{' {
				braceLevel++
			} else if ch == '}' {
				braceLevel--
			} else if ch == '[' {
				bracketLevel++
			} else if ch == ']' {
				bracketLevel--
			} else if ch == '"' || ch == '\'' {
				inQuotes = true
				quoteChar = ch
			} else if ch == delimiter && braceLevel == 0 && bracketLevel == 0 {
				result = append(result, current.String())
				current.Reset()
				continue
			}
		} else if ch == quoteChar && inQuotes {
			inQuotes = false
		}

		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// isNumericValue checks if a string is a valid number
func isNumericValue(s string) bool {
	s = strings.TrimSpace(s)

	// Check for numeric format
	hasDigit := false
	hasDot := false

	for i, ch := range s {
		if ch == '-' && i == 0 {
			// Negative sign at beginning is ok
			continue
		}
		if ch == '.' && !hasDot {
			hasDot = true
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
		hasDigit = true
	}

	return hasDigit
}

// parseInt parses a string to an integer
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	var result int
	var negative bool

	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		result = result*10 + int(ch-'0')
	}

	if negative {
		result = -result
	}

	return result
}

// parseFloat parses a string to a float
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	var result float64
	var negative bool

	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	parts := strings.Split(s, ".")

	// Integer part
	for _, ch := range parts[0] {
		if ch < '0' || ch > '9' {
			break
		}
		result = result*10 + float64(ch-'0')
	}

	// Decimal part
	if len(parts) > 1 {
		decimal := 0.0
		multiplier := 0.1

		for _, ch := range parts[1] {
			if ch < '0' || ch > '9' {
				break
			}
			decimal += float64(ch-'0') * multiplier
			multiplier *= 0.1
		}

		result += decimal
	}

	if negative {
		result = -result
	}

	return result
}

// ensureVariablesInScope ensures all referenced variables exist in the data scope
// This is critical for Alpine.js to work correctly with expressions
func ensureVariablesInScope(nodes []ast.Node, dataScope map[string]any) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.ExpressionNode:
			// Add variables from expressions
			AddExprVarsToScope(n.Expression, dataScope)

		case *ast.Element:
			// Check attributes for expressions
			for _, attr := range n.Attributes {
				if attr.Dynamic || attr.IsAlpine {
					AddExprVarsToScope(attr.Value, dataScope)
				}
			}

			// Recursively process children
			ensureVariablesInScope(n.Children, dataScope)

		case *ast.Conditional:
			// Add variables from condition
			AddExprVarsToScope(n.IfCondition, dataScope)

			// Process branches
			ensureVariablesInScope(n.IfContent, dataScope)
			for i, condition := range n.ElseIfConditions {
				AddExprVarsToScope(condition, dataScope)
				if i < len(n.ElseIfContent) {
					ensureVariablesInScope(n.ElseIfContent[i], dataScope)
				}
			}
			ensureVariablesInScope(n.ElseContent, dataScope)

		case *ast.Loop:
			// Add loop variables
			dataScope[n.Iterator] = nil
			if n.Value != "" {
				dataScope[n.Value] = nil
			}

			// Add array variable
			AddExprVarsToScope(n.Collection, dataScope)

			// Process loop body
			ensureVariablesInScope(n.Content, dataScope)
		}
	}
}

// TransformWithAlpineData transforms the given nodes with an Alpine.js data wrapper
// This is the main entry point for applying Alpine.js data binding to templates
func TransformWithAlpineData(nodes []ast.Node, dataScope map[string]any) ast.Node {
	// Ensure all variables referenced in the nodes exist in the data scope
	ensureVariablesInScope(nodes, dataScope)

	// Wrap the nodes with Alpine.js data wrapper
	return wrapWithAlpineData(nodes, dataScope)
}
