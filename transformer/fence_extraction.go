package transformer

import (
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// collectComponentFenceData extracts variables, prop defaults, and functions from a
// component's fence section and adds them to the provided scope map.
//
// It processes:
//   - fence.Variables: Parsed using parseValue() and added to scope
//   - fence.Props: Default values parsed using parseValue() and added to scope
//   - fence.RawContent: Function definitions extracted via regex and stored as strings
//
// Supported function patterns:
//   - function name() {} - Function declarations
//   - const/let/var name = function() {} - Function expressions
//   - const/let/var name = () => {} - Arrow functions
//   - name() {} - Method shorthand
//   - async variants of all above
//
// The function modifies the provided scope map in-place.
//
// Cognitive Load: 12 (moderate complexity with regex and multiple data sources)
func collectComponentFenceData(fence *ast.FenceSection, scope map[string]any) {
	// Handle nil fence - return early without panic (COGNITIVE LOAD RULE)
	if fence == nil {
		return
	}

	// Process variables - parse and add to scope
	for _, variable := range fence.Variables {
		parsedValue := parseValue(variable.Value)
		scope[variable.Name] = parsedValue
	}

	// Process props - parse default values and add to scope
	for _, prop := range fence.Props {
		parsedValue := parseValue(prop.DefaultValue)
		scope[prop.Name] = parsedValue
	}

	// Extract functions from RawContent if present
	if fence.RawContent != "" {
		extractFunctions(fence.RawContent, scope)
	}
}

// extractFunctions uses regex patterns to find and extract JavaScript function definitions
// from the raw fence content. Each function is stored as a string in the scope.
//
// Cognitive Load: 18 (complex regex patterns and brace matching)
func extractFunctions(rawContent string, scope map[string]any) {
	// Define regex patterns for different function syntaxes
	// Note: These patterns handle nested braces using a custom matcher function

	// JavaScript keywords that should never be considered function names
	jsKeywords := map[string]bool{
		"if": true, "else": true, "while": true, "for": true, "do": true,
		"switch": true, "case": true, "default": true, "try": true, "catch": true,
		"finally": true, "throw": true, "return": true, "break": true, "continue": true,
		"import": true, "export": true, "class": true, "extends": true, "super": true,
		"new": true, "this": true, "typeof": true, "instanceof": true, "void": true,
		"delete": true, "in": true, "with": true,
	}

	patterns := []struct {
		pattern *regexp.Regexp
		nameIdx int // Index of the capture group containing the function name
	}{
		// Pattern 1: async function name() {}
		{regexp.MustCompile(`(?m)^async function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`), 1},

		// Pattern 2: function name() {} and function* name() {} (generator)
		{regexp.MustCompile(`(?m)^function\s*\*?\s*([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`), 1},

		// Pattern 3: const/let/var name = async function() {}
		{regexp.MustCompile(`(?m)^(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*async\s+function\s*\(`), 1},

		// Pattern 4: const/let/var name = function() {}
		{regexp.MustCompile(`(?m)^(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*function\s*\(`), 1},

		// Pattern 5: const/let/var name = async (...) => {}
		{regexp.MustCompile(`(?m)^(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*async\s+`), 1},

		// Pattern 6: const/let/var name = () => {} or const/let/var name = x => {}
		{regexp.MustCompile(`(?m)^(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*`), 1},

		// Pattern 7: async name() {} (method shorthand)
		{regexp.MustCompile(`(?m)^async ([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`), 1},

		// Pattern 8: name() {} (method shorthand) - will be filtered by keyword check
		{regexp.MustCompile(`(?m)^([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`), 1},
	}

	// Track which functions we've already extracted to avoid duplicates
	extracted := make(map[int]bool)

	for _, patternInfo := range patterns {
		matches := patternInfo.pattern.FindAllStringSubmatchIndex(rawContent, -1)

		for _, match := range matches {
			// Skip if we've already extracted a function starting at this position
			if extracted[match[0]] {
				continue
			}

			// Extract function name
			nameStart := match[2*patternInfo.nameIdx]
			nameEnd := match[2*patternInfo.nameIdx+1]
			if nameStart == -1 || nameEnd == -1 {
				continue
			}
			funcName := rawContent[nameStart:nameEnd]

			// Skip JavaScript keywords (if, while, for, etc.)
			if jsKeywords[funcName] {
				continue
			}

			// Skip if this name is actually a variable assignment that's not a function
			// (e.g., "let count = 0" should not be extracted as a function)
			if !isFunctionAssignment(rawContent, match[0]) {
				continue
			}

			// Find the complete function definition
			funcStart := match[0]
			funcDef := extractCompleteFunction(rawContent, funcStart)

			if funcDef != "" {
				scope[funcName] = funcDef
				extracted[funcStart] = true
			}
		}
	}
}

// isFunctionAssignment checks if a line starting at the given position is a function assignment
// Returns false for non-function assignments like "const count = 0"
func isFunctionAssignment(content string, start int) bool {
	remaining := content[start:]

	// Get just the first line
	firstLine := remaining
	if nlIdx := strings.Index(remaining, "\n"); nlIdx != -1 {
		firstLine = remaining[:nlIdx]
	}

	// Check for arrow function on the first line
	if strings.Contains(firstLine, "=>") {
		return true
	}

	// Check for function keyword on the first line (or within 200 chars for multi-line params)
	checkRange := remaining
	if len(remaining) > 200 {
		checkRange = remaining[:200] // Look ahead a bit for function keyword
	}
	
	// Check for "function" keyword appearing soon (not way later in a different function)
	if strings.Contains(checkRange, "function") {
		// Make sure it's not finding a function that appears later
		funcIdx := strings.Index(checkRange, "function")
		beforeFunc := checkRange[:funcIdx]
		
		// If there's a blank line (double newline) before "function", it's a different statement
		if strings.Contains(beforeFunc, "\n\n") || strings.Contains(beforeFunc, "\n\r\n") {
			return false
		}
		
		// If "function" appears after a newline that's not indented (start of new statement), skip it
		lines := strings.Split(beforeFunc, "\n")
		if len(lines) > 1 {
			// There's a newline before "function"
			lastLine := lines[len(lines)-1]
			// If the last line before "function" is empty or not a continuation, it's a different statement
			if strings.TrimSpace(lastLine) == "" {
				return false
			}
			// If the last line doesn't end with "=", it's not a continuation
			if !strings.HasSuffix(strings.TrimRight(lastLine, " \t"), "=") {
				return false
			}
		}
		return true
	}

	// Check for method shorthand or function declaration
	// Look for pattern: name(params) {
	if matched, _ := regexp.MatchString(`^[ \t]*(?:async\s+)?(?:function\s*\*?\s*)?[a-zA-Z_$][a-zA-Z0-9_$]*\s*\([^)]*\)\s*\{`, remaining); matched {
		return true
	}

	// If we're here, it's likely not a function
	return false
}

// extractCompleteFunction extracts a complete function definition starting from the given
// position. It handles nested braces and various function syntaxes.
//
// Cognitive Load: 15 (complex state machine for brace matching)
func extractCompleteFunction(content string, start int) string {
	// Find where the function body or expression starts
	remaining := content[start:]

	// Check if this is an arrow function assignment (const/let/var name = ...)
	// Only check for arrow in the first line to avoid false positives from nested arrows
	isArrowAssignment := false
	if strings.HasPrefix(strings.TrimSpace(remaining), "const ") ||
		strings.HasPrefix(strings.TrimSpace(remaining), "let ") ||
		strings.HasPrefix(strings.TrimSpace(remaining), "var ") {
		// This is a variable assignment, check if it's an arrow function
		firstLine := remaining
		if nlIdx := strings.Index(remaining, "\n"); nlIdx != -1 {
			firstLine = remaining[:nlIdx]
		}
		if strings.Contains(firstLine, "=>") {
			isArrowAssignment = true
		}
	}

	// Handle arrow function with implicit return (no braces)
	// e.g., const double = x => x * 2
	if isArrowAssignment {
		arrowIdx := strings.Index(remaining, "=>")
		if arrowIdx != -1 {
			// Check if there's a brace after the arrow
			afterArrow := remaining[arrowIdx+2:]
			afterArrow = strings.TrimLeft(afterArrow, " \t")

			if !strings.HasPrefix(afterArrow, "{") {
				// Implicit return - find the end (newline or semicolon)
				endIdx := findImplicitReturnEnd(remaining)
				if endIdx != -1 {
					result := strings.TrimRight(remaining[:endIdx], " \t\n\r;")
					return result
				}
			}
		}
	}

	// Find the opening brace (or opening paren for parameter list)
	// We need to handle destructured parameters like: function displayUser({ name, email, role = 'user' }) {

	// First, find opening parenthesis for parameters
	parenIdx := strings.Index(remaining, "(")
	if parenIdx == -1 {
		return "" // No parameters found
	}

	// Match parentheses to find end of parameters
	parenCount := 0
	inString := false
	var stringChar byte
	escaped := false
	paramEnd := -1

	for i := parenIdx; i < len(remaining); i++ {
		ch := remaining[i]

		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}

		// Handle strings
		if (ch == '"' || ch == '\'') {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
			continue
		}
		if inString {
			continue
		}

		if ch == '(' {
			parenCount++
		} else if ch == ')' {
			parenCount--
			if parenCount == 0 {
				paramEnd = i
				break
			}
		}
	}

	if paramEnd == -1 {
		return "" // Unmatched parentheses
	}

	// Now find the opening brace after the parameters
	afterParams := remaining[paramEnd+1:]
	braceIdx := strings.Index(afterParams, "{")
	if braceIdx == -1 {
		// No braces found - might be arrow function with implicit return
		// Already handled above
		return ""
	}

	// Adjust braceIdx to be relative to remaining
	braceIdx = paramEnd + 1 + braceIdx

	// Match braces to find the closing brace
	braceCount := 0
	inString = false
	inRegex := false
	inTemplate := false
	escaped = false

	for i := braceIdx; i < len(remaining); i++ {
		ch := remaining[i]

		// Handle escape sequences
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}

		// Handle template literals
		if ch == '`' && !inString && !inRegex {
			inTemplate = !inTemplate
			continue
		}
		if inTemplate {
			continue
		}

		// Handle strings
		if (ch == '"' || ch == '\'') && !inRegex {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
			continue
		}
		if inString {
			continue
		}

		// Handle regex literals (simplified - may not catch all cases)
		if ch == '/' && i > 0 {
			// Check if this might be a regex
			prevNonSpace := i - 1
			for prevNonSpace >= 0 && (remaining[prevNonSpace] == ' ' || remaining[prevNonSpace] == '\t') {
				prevNonSpace--
			}
			if prevNonSpace >= 0 {
				prev := remaining[prevNonSpace]
				// Regex can follow: =, (, [, {, :, ;, !, &, |, ?, +, -, *, /, %, <, >, ^, ~, ,, return
				if prev == '=' || prev == '(' || prev == '[' || prev == '{' ||
				   prev == ':' || prev == ';' || prev == '!' || prev == '&' ||
				   prev == '|' || prev == '?' || prev == '+' || prev == '-' ||
				   prev == '*' || prev == '/' || prev == '%' || prev == '<' ||
				   prev == '>' || prev == '^' || prev == '~' || prev == ',' {
					inRegex = true
					continue
				}
			}
		}
		if inRegex && ch == '/' {
			inRegex = false
			continue
		}
		if inRegex {
			continue
		}

		// Count braces
		if ch == '{' {
			braceCount++
		} else if ch == '}' {
			braceCount--
			if braceCount == 0 {
				// Found the matching closing brace
				result := remaining[:i+1]
				// Trim trailing whitespace and semicolons
				result = strings.TrimRight(result, " \t\n\r;")
				return result
			}
		}
	}

	// If we get here, braces didn't match - return what we have
	return ""
}

// findImplicitReturnEnd finds the end of an arrow function with implicit return
// (i.e., without braces). It looks for a newline or semicolon.
//
// Cognitive Load: 8 (simple line scanning)
func findImplicitReturnEnd(content string) int {
	// Handle multiline method chains like:
	// const transform = data => data
	//   .filter(x => x.active)
	//   .map(x => x.value)

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return -1
	}

	// Start with the first line
	result := lines[0]
	length := len(result)

	// Check subsequent lines for method chaining (starts with .)
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimLeft(lines[i], " \t")
		if strings.HasPrefix(trimmed, ".") {
			// This line is part of the chain
			length += 1 + len(lines[i]) // +1 for newline
			result += "\n" + lines[i]
		} else {
			// End of the expression
			break
		}
	}

	return length
}
