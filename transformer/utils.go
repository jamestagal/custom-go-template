package transformer

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// isJSObjectLiteral checks if a string appears to be a JavaScript object literal
func isJSObjectLiteral(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}

// isJSArrayLiteral checks if a string appears to be a JavaScript array literal
func isJSArrayLiteral(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")
}

// isJSFunctionLiteral checks if a string appears to be a JavaScript function literal
func isJSFunctionLiteral(s string) bool {
	s = strings.TrimSpace(s)
	return strings.Contains(s, "=>") ||
	       strings.HasPrefix(s, "function") ||
	       (strings.Contains(s, "(") && strings.Contains(s, ")") && strings.Contains(s, "{") && strings.Contains(s, "}"))
}

// parseValue parses JavaScript literal values from fence section strings into
// appropriate Go types or Alpine.js-compatible strings.
//
// Returns:
//   - bool for "true" and "false"
//   - nil for "null"
//   - int for integers (e.g., "42", "-42")
//   - float64 for floats (e.g., "3.14", "1e5")
//   - string for quoted strings (quotes stripped, escape sequences handled)
//   - string for arrays (e.g., "[1,2,3]" returned as-is for Alpine.js)
//   - string for objects (e.g., "{key: 'value'}" returned as-is for Alpine.js)
//   - string for expressions/variables (e.g., "userName", "getTotal()")
func parseValue(value string) interface{} {
	// CRITICAL DEBUG: Log the raw input value
	log.Printf("[parseValue] ENTRY: input value type=%T, len=%d, raw=%q", value, len(value), value)

	// Trim whitespace from input
	value = strings.TrimSpace(value)
	log.Printf("[parseValue] After TrimSpace: len=%d, value=%q", len(value), value)

	// Handle empty string
	if value == "" {
		log.Printf("[parseValue] Empty string, returning empty")
		return ""
	}

	// Handle booleans
	if value == "true" {
		log.Printf("[parseValue] Returning bool true")
		return true
	}
	if value == "false" {
		log.Printf("[parseValue] Returning bool false")
		return false
	}

	// Handle null
	if value == "null" {
		log.Printf("[parseValue] Returning nil")
		return nil
	}

	// Handle quoted strings (both double and single quotes)
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			// Check if quotes are properly closed (not unclosed strings)
			quote := value[0]
			if value[len(value)-1] == quote {
				// CRITICAL: Check if the content inside quotes is an array/object literal
				unquoted := value[1 : len(value)-1]
				unquotedTrimmed := strings.TrimSpace(unquoted)

				log.Printf("[parseValue] Found quoted string, quote=%c, unquoted=%q", quote, unquoted)

				// Check if this is a quoted JavaScript literal (array or object)
				if (strings.HasPrefix(unquotedTrimmed, "[") && strings.HasSuffix(unquotedTrimmed, "]")) ||
				   (strings.HasPrefix(unquotedTrimmed, "{") && strings.HasSuffix(unquotedTrimmed, "}")) {
					// This is a quoted JavaScript literal - unwrap it!
					log.Printf("[parseValue] UNWRAPPING quoted JS literal: %q", unquotedTrimmed)
					return unquotedTrimmed
				}

				// Regular string value - handle escape sequences
				result := unescapeString(unquoted)
				log.Printf("[parseValue] Unquoted string result: %q", result)
				return result
			}
		}
	}

	// Handle arrays - return as string for Alpine.js
	if strings.HasPrefix(value, "[") {
		log.Printf("[parseValue] Detected array literal, returning as-is: %q", value)
		// Even if brackets don't match, return as string for Alpine.js to handle
		return value
	}

	// Handle objects - return as string for Alpine.js
	if strings.HasPrefix(value, "{") {
		log.Printf("[parseValue] Detected object literal, returning as-is: %q", value)
		// Even if braces don't match, return as string for Alpine.js to handle
		return value
	}

	// Try to parse as number
	if num, isNum := tryParseNumber(value); isNum {
		log.Printf("[parseValue] Parsed as number: %v", num)
		return num
	}

	// Everything else is an expression or variable reference - return as string
	log.Printf("[parseValue] Returning as expression/variable: %q", value)
	return value
}

// tryParseNumber attempts to parse a string as a number (int or float64).
// Returns (value, true) if successfully parsed, (nil, false) otherwise.
func tryParseNumber(value string) (interface{}, bool) {
	// Check if it looks like a number at all
	if value == "" {
		return nil, false
	}

	// Check for scientific notation (contains 'e' or 'E')
	if strings.ContainsAny(value, "eE") {
		// Parse as float (scientific notation)
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f, true
		}
		return nil, false
	}

	// Check if it contains a decimal point
	if strings.Contains(value, ".") {
		// Parse as float
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f, true
		}
		return nil, false
	}

	// Try to parse as integer
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		// Return as int (which is int64 on 64-bit systems, int32 on 32-bit systems)
		// This ensures the type matches Go's natural int literal type
		return int(i), true
	}

	return nil, false
}

// unescapeString handles escape sequences in strings.
// Supports: \", \', \\, \n, \t, and standard Go escape sequences.
func unescapeString(s string) string {
	// Use strings.Builder for efficient string building
	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			// Handle escape sequence
			next := s[i+1]
			switch next {
			case 'n':
				result.WriteByte('\n')
				i++ // Skip next character
			case 't':
				result.WriteByte('\t')
				i++ // Skip next character
			case 'r':
				result.WriteByte('\r')
				i++ // Skip next character
			case '\\':
				result.WriteByte('\\')
				i++ // Skip next character
			case '"':
				result.WriteByte('"')
				i++ // Skip next character
			case '\'':
				result.WriteByte('\'')
				i++ // Skip next character
			default:
				// For unknown escape sequences, keep the backslash and character
				result.WriteByte('\\')
			}
		} else {
			result.WriteByte(s[i])
		}
	}

	return result.String()
}

// filterOutFence removes all FenceSection nodes from a slice of AST nodes.
// It preserves all other node types in their original order without modifying them.
// Returns a new slice containing only non-FenceSection nodes.
//
// Cognitive Load: 4 (simple filter operation)
func filterOutFence(nodes []ast.Node) []ast.Node {
	// Handle nil input - return empty slice (not nil)
	if nodes == nil {
		return []ast.Node{}
	}

	// Preallocate result slice (COGNITIVE LOAD RULE: preallocate when size known)
	// We don't know exact size, but at most it's len(nodes)
	result := make([]ast.Node, 0, len(nodes))

	// Filter out FenceSection nodes
	for _, node := range nodes {
		if _, isFence := node.(*ast.FenceSection); !isFence {
			result = append(result, node)
		}
	}

	return result
}

// getMapKeys returns the keys of a map as a sorted slice for debugging
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
