package parser

import (
	"log"
	"regexp"
	"strings"
	"unicode"

	"github.com/jimafisk/custom_go_template/ast"
)

// TextParser parses text content up to any of the given delimiters
func TextParser(delimiters ...Parser) Parser {
	return func(input string) Result {
		log.Printf("[TextParser] Starting on: '%.30s...'", input)

		// If no delimiters provided, use a default set
		if len(delimiters) == 0 {
			delimiters = []Parser{String("<"), String("{")}
		}

		delimiterChoice := Choice(delimiters...)
		var consumed strings.Builder
		current := input

		for len(current) > 0 {
			// Check if current position matches any delimiter
			delimiterResult := delimiterChoice(current)
			if delimiterResult.Successful {
				log.Printf("[TextParser] Found delimiter at position %d", len(input)-len(current))
				break
			}

			// Consume one character
			consumed.WriteByte(current[0])
			current = current[1:]
		}

		// Return only if we consumed something
		content := consumed.String()
		if len(content) > 0 {
			log.Printf("[TextParser] Parsed text node with %d chars: %.30s...", len(content), content)
			return Result{&ast.TextNode{Content: content}, current, true, "", false}
		}

		// Nothing consumed - explicitly fail rather than returning an empty success
		log.Printf("[TextParser] No text before delimiters, failing")
		return Result{nil, input, false, "no text before delimiters", false}
	}
}

// ExpressionParser parses expressions in curly braces: {variable} or {$store.property}
// Pattern: Expression Router [Load: 8]
// Cognitive Load: 8 (routing logic with $ detection)
//
// Routes expression parsing based on content:
//   - {$storeName.prop} → StoreExpressionNode (via parseStoreExpression)
//   - {variable}        → ExpressionNode
//   - {obj.prop}        → ExpressionNode
//   - {expr + 1}        → ExpressionNode
func ExpressionParser() Parser {
	return func(input string) Result {
		// Use the lex-based expression parser to extract content within braces
		exprRes := LexExpressionParser()(input)
		if !exprRes.Successful {
			return exprRes
		}

		expr, ok := exprRes.Value.(string)
		if !ok {
			return Result{nil, input, false, "expression parser did not return string", false}
		}

		log.Printf("[ExpressionParser] Found expression: '%s'", expr)

		// COGNITIVE LOAD RULE: Check if expression starts with $ (store reference)
		// Route to parseStoreExpression if it's a store expression
		if len(expr) > 0 && expr[0] == '$' {
			// This is a store expression - parse it with parseStoreExpression
			storeResult := parseStoreExpression()(expr)
			if storeResult.Successful {
				log.Printf("[ExpressionParser] Routed to store parser: %s", expr)
				// Return store expression node with correct remaining
				return Result{
					Value:      storeResult.Value,
					Remaining:  exprRes.Remaining,
					Successful: true,
					Error:      "",
					Dynamic:    true,
				}
			}
			// If store parsing failed, fall through to regular expression
			log.Printf("[ExpressionParser] Store parsing failed, treating as regular expression: %s", expr)
		}

		// Regular expression (not a store)
		return Result{
			Value:      &ast.ExpressionNode{Expression: expr},
			Remaining:  exprRes.Remaining,
			Successful: true,
			Error:      "",
			Dynamic:    true,
		}
	}
}

// LexExpressionParser uses a lexing-based approach to parse expressions.
// It recognizes single curly-brace expressions: {expr}.
func LexExpressionParser() Parser {
	return func(input string) Result {
		if !strings.HasPrefix(input, "{") {
			return Result{nil, input, false, "not an expression", false}
		}

		// If it starts with "{if", "{for", "{else", or "{/", it's a directive, not an expression
		if strings.HasPrefix(input, "{if ") || strings.HasPrefix(input, "{if}") ||
			strings.HasPrefix(input, "{for ") ||
			strings.HasPrefix(input, "{else}") ||
			strings.HasPrefix(input, "{else if ") ||
			strings.HasPrefix(input, "{/") {
			return Result{nil, input, false, "not an expression (directive)", false}
		}

		// Find the closing brace
		depth := 0
		inString := false
		escaped := false
		var stringChar rune

		for i, ch := range input {
			if escaped {
				escaped = false
				continue
			}

			if ch == '\\' {
				escaped = true
				continue
			}

			if !inString {
				if ch == '"' || ch == '\'' {
					inString = true
					stringChar = ch
					continue
				}

				if ch == '{' {
					depth++
				} else if ch == '}' {
					depth--
					if depth == 0 {
						// Found the closing brace
						expr := input[1:i] // Extract expression without braces
						remaining := input[i+1:]
						return Result{expr, remaining, true, "", false}
					}
				}
			} else {
				if ch == stringChar {
					inString = false
				}
			}
		}

		// No matching closing brace found
		return Result{nil, input, false, "unclosed expression", false}
	}
}

// parseStoreExpression parses store expressions: $storeName or $storeName.property
// Pattern: Store Expression Parser [Load: 6]
// Cognitive Load: 6 (simple parsing with validation)
//
// Syntax:
//   - $storeName            → StoreExpressionNode{StoreName: "storeName", Property: ""}
//   - $storeName.property   → StoreExpressionNode{StoreName: "storeName", Property: "property"}
//   - $storeName.nested.prop → StoreExpressionNode{StoreName: "storeName", Property: "nested.prop"}
//
// Store names must:
//   - Start with $ followed by letter or underscore
//   - Contain only letters, digits, underscores
//
// Property paths:
//   - Optional, separated by dots
//   - Can have multiple levels (e.g., user.profile.name)
func parseStoreExpression() Parser {
	return func(input string) Result {
		// Check if starts with $ (COGNITIVE LOAD RULE: early validation)
		if len(input) == 0 || input[0] != '$' {
			return Result{nil, input, false, "not a store expression", false}
		}

		// Need at least one character after $
		if len(input) == 1 {
			return Result{nil, input, false, "invalid store name", false}
		}

		// Parse store name after $
		// Store name must start with letter or underscore
		pos := 1
		firstChar := rune(input[pos])
		if !unicode.IsLetter(firstChar) && firstChar != '_' {
			return Result{nil, input, false, "invalid store name", false}
		}

		// Continue parsing store name (alphanumeric + underscore)
		storeName := strings.Builder{}
		for pos < len(input) {
			ch := rune(input[pos])
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
				storeName.WriteRune(ch)
				pos++
			} else {
				break
			}
		}

		// Check if we have a store name (COGNITIVE LOAD RULE: validate before proceeding)
		if storeName.Len() == 0 {
			return Result{nil, input, false, "invalid store name", false}
		}

		storeNameStr := storeName.String()
		property := ""

		// Check for property access (dot notation)
		if pos < len(input) && input[pos] == '.' {
			pos++ // Skip the dot

			// Parse property path (can have multiple dots)
			propertyBuilder := strings.Builder{}
			for pos < len(input) {
				ch := rune(input[pos])
				// Property can contain letters, digits, underscores, and dots
				if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' {
					propertyBuilder.WriteRune(ch)
					pos++
				} else {
					break
				}
			}

			property = propertyBuilder.String()
			// Trim trailing dots if any
			property = strings.TrimSuffix(property, ".")
		}

		// Create StoreExpressionNode (COGNITIVE LOAD RULE: wrapped success)
		node := &ast.StoreExpressionNode{
			StoreName: storeNameStr,
			Property:  property,
		}

		remaining := input[pos:]
		log.Printf("[parseStoreExpression] Parsed store expression: %s", node.String())

		return Result{
			Value:      node,
			Remaining:  remaining,
			Successful: true,
			Error:      "",
			Dynamic:    true,
		}
	}
}

// FenceParser parses the fence section (delimited by ---).
func FenceParser() Parser {
	return Map(
		Between(String("---"), String("---"), TakeUntil(String("---"))),
		func(value interface{}) (interface{}, error) {
			content := value.(string)
			log.Printf("[FenceParser] Parsed fence with %d chars", len(content))

			// Parse the fence content to extract props and variables
			fence := parseFenceContent(content)
			return fence, nil
		},
	)
}

// unwrapQuotedLiteral removes outer quotes from JavaScript literals if present
// Pattern: Value Sanitization [Load: 6]
// Cognitive Load: 6 (string manipulation, conditional unwrapping)
//
// Handles cases where multi-line parsing wraps JavaScript literals in quotes:
//   - '[...]' → [...]  (array literal)
//   - '{...}' → {...}  (object literal)
//   - '"text"' → "text" (already quoted string - keep as-is)
//   - 'text' → text (unquoted string)
//
// This fixes the bug where notifications appears as 'notifications:'[...]” with outer quotes
func unwrapQuotedLiteral(value string) string {
	trimmed := strings.TrimSpace(value)

	// Need at least 2 characters for quotes
	if len(trimmed) < 2 {
		return value
	}

	// Check if outer characters are quotes
	firstChar := trimmed[0]
	lastChar := trimmed[len(trimmed)-1]

	// Single-quoted or double-quoted
	if (firstChar == '\'' && lastChar == '\'') || (firstChar == '"' && lastChar == '"') {
		inner := trimmed[1 : len(trimmed)-1]
		innerTrimmed := strings.TrimSpace(inner)

		// Check if inner content is a JavaScript literal (array or object)
		if len(innerTrimmed) > 0 {
			innerFirst := innerTrimmed[0]
			innerLast := innerTrimmed[len(innerTrimmed)-1]

			// Array literal: '[...]' → [...]
			if innerFirst == '[' && innerLast == ']' {
				log.Printf("[unwrapQuotedLiteral] Unwrapped array literal: %q → %q", trimmed[:min(50, len(trimmed))], innerTrimmed[:min(50, len(innerTrimmed))])
				return innerTrimmed
			}

			// Object literal: '{...}' → {...}
			if innerFirst == '{' && innerLast == '}' {
				log.Printf("[unwrapQuotedLiteral] Unwrapped object literal: %q → %q", trimmed[:min(50, len(trimmed))], innerTrimmed[:min(50, len(innerTrimmed))])
				return innerTrimmed
			}
		}
	}

	// Not a quoted literal, return as-is
	return value
}

// parseFenceContent extracts props, variables, functions, stores, and imports from fence section content
// Now handles multi-line values for arrays, objects, and ternary expressions
// Pattern: Fence Content Parser [Load: 12]
// Cognitive Load: 12 (multiple regex patterns, line-by-line parsing, multi-line support)
func parseFenceContent(content string) *ast.FenceSection {
	fence := &ast.FenceSection{
		RawContent:    content,
		Props:         []ast.PropNode{},
		ExportedProps: []string{},
		Variables:     []ast.VariableNode{},
		Functions:     []ast.FunctionNode{},
		Imports:       []ast.ImportNode{},
		Stores:        make(map[string]string),
	}

	lines := strings.Split(content, "\n")

	// Regex patterns for parsing (COGNITIVE LOAD RULE: precompile patterns)
	propRegex := regexp.MustCompile(`^\s*prop\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+)$`)
	varRegex := regexp.MustCompile(`^\s*(let|const|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+)$`)
	storeRegex := regexp.MustCompile(`^\s*store\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+)$`)
	importRegex := regexp.MustCompile(`^\s*import\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s+from\s+['"](.+?)['"](?:;)?$`)

	exportLetRegex := regexp.MustCompile(`^\s*export\s+let\s+(.*)$`)
	// Function patterns: both regular functions and getters
	functionRegex := regexp.MustCompile(`^\s*function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(([^)]*)\)\s*\{`)
	getterRegex := regexp.MustCompile(`^\s*get\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(\s*\)\s*\{`)

	// Process lines with multi-line support
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			i++
			continue
		}
		// Check for export let declaration (COGNITIVE LOAD RULE: parse before variables)
		if matches := exportLetRegex.FindStringSubmatch(trimmedLine); matches != nil {
			propList := matches[1]

			// Remove optional semicolon
			propList = strings.TrimSuffix(strings.TrimSpace(propList), ";")

			// Split by comma and extract prop names
			if propList != "" {
				propNames := strings.Split(propList, ",")
				for _, propName := range propNames {
					propName = strings.TrimSpace(propName)
					if propName != "" {
						log.Printf("[parseFenceContent] Found exported prop: %s", propName)
						fence.ExportedProps = append(fence.ExportedProps, propName)
					}
				}
			}

			i++
			continue
		}

		// Check for function declaration (COGNITIVE LOAD RULE: parse before variables to avoid conflicts)
		if matches := functionRegex.FindStringSubmatch(trimmedLine); matches != nil {
			funcName := matches[1]
			funcParams := matches[2]

			// Find the matching closing brace for this function
			funcBody, endIndex := parseFunctionBody(lines, i)
			if funcBody != "" {
				log.Printf("[parseFenceContent] Found function: %s(%s)", funcName, funcParams)
				fence.Functions = append(fence.Functions, ast.FunctionNode{
					Name:     funcName,
					Params:   funcParams,
					Body:     funcBody,
					IsGetter: false,
				})
				i = endIndex + 1
				continue
			}
		}

		// Check for getter declaration
		if matches := getterRegex.FindStringSubmatch(trimmedLine); matches != nil {
			getterName := matches[1]

			// Find the matching closing brace for this getter
			getterBody, endIndex := parseFunctionBody(lines, i)
			if getterBody != "" {
				log.Printf("[parseFenceContent] Found getter: get %s()", getterName)
				fence.Functions = append(fence.Functions, ast.FunctionNode{
					Name:     getterName,
					Params:   "", // Getters have no params
					Body:     getterBody,
					IsGetter: true,
				})
				i = endIndex + 1
				continue
			}
		}

		// Check for store declaration
		if matches := storeRegex.FindStringSubmatch(trimmedLine); matches != nil {
			storeName := matches[1]
			firstLineValue := matches[2]

			// Check if this is a potentially multi-line value
			if isMultiLineValue(firstLineValue) {
				// Multi-line value - accumulate lines until complete
				fullValue, endIndex := parseMultiLineValue(lines, i, firstLineValue)
				if fullValue != "" {
					// Unwrap quoted literals (COGNITIVE LOAD RULE: sanitize before storing)
					unwrapped := unwrapQuotedLiteral(fullValue)

					preview := unwrapped
					if len(unwrapped) > 50 {
						preview = unwrapped[:50] + "..."
					}
					log.Printf("[parseFenceContent] Found multi-line store: %s = %s (total %d chars)", storeName, preview, len(unwrapped))
					fence.Stores[storeName] = unwrapped
					i = endIndex + 1
					continue
				}
			}

			// Single-line value
			value := strings.TrimSpace(firstLineValue)
			value = strings.TrimSuffix(value, ";")
			value = unwrapQuotedLiteral(value) // Also unwrap single-line values

			log.Printf("[parseFenceContent] Found store: %s = %s", storeName, value)
			fence.Stores[storeName] = value
			i++
			continue
		}

		// Check for prop declaration
		if matches := propRegex.FindStringSubmatch(trimmedLine); matches != nil {
			propName := matches[1]
			firstLineValue := matches[2]

			// Check if this is a potentially multi-line value
			if isMultiLineValue(firstLineValue) {
				// Multi-line value - accumulate lines until complete
				fullValue, endIndex := parseMultiLineValue(lines, i, firstLineValue)
				if fullValue != "" {
					// Unwrap quoted literals (COGNITIVE LOAD RULE: sanitize before storing)
					unwrapped := unwrapQuotedLiteral(fullValue)

					preview := unwrapped
					if len(unwrapped) > 50 {
						preview = unwrapped[:50] + "..."
					}
					log.Printf("[parseFenceContent] Found multi-line prop: %s = %s (total %d chars)", propName, preview, len(unwrapped))
					fence.Props = append(fence.Props, ast.PropNode{
						Name:         propName,
						DefaultValue: unwrapped,
					})
					i = endIndex + 1
					continue
				}
			}

			// Single-line value
			value := strings.TrimSpace(firstLineValue)
			value = strings.TrimSuffix(value, ";")
			value = unwrapQuotedLiteral(value) // Also unwrap single-line values

			log.Printf("[parseFenceContent] Found prop: %s = %s", propName, value)
			fence.Props = append(fence.Props, ast.PropNode{
				Name:         propName,
				DefaultValue: value,
			})
			i++
			continue
		}

		// Check for variable declaration
		if matches := varRegex.FindStringSubmatch(trimmedLine); matches != nil {
			keyword := matches[1]
			varName := matches[2]
			firstLineValue := matches[3]

			// Check if this is a potentially multi-line value
			if isMultiLineValue(firstLineValue) {
				// Multi-line value
				fullValue, endIndex := parseMultiLineValue(lines, i, firstLineValue)
				if fullValue != "" {
					// Unwrap quoted literals (COGNITIVE LOAD RULE: sanitize before storing)
					unwrapped := unwrapQuotedLiteral(fullValue)

					preview := unwrapped
					if len(unwrapped) > 50 {
						preview = unwrapped[:50] + "..."
					}
					log.Printf("[parseFenceContent] Found multi-line variable: %s %s = %s (total %d chars)", keyword, varName, preview, len(unwrapped))
					fence.Variables = append(fence.Variables, ast.VariableNode{
						Keyword: keyword,
						Name:    varName,
						Value:   unwrapped,
					})
					i = endIndex + 1
					continue
				}
			}

			// Single-line value
			value := strings.TrimSpace(firstLineValue)
			value = strings.TrimSuffix(value, ";")
			value = unwrapQuotedLiteral(value) // Also unwrap single-line values

			log.Printf("[parseFenceContent] Found variable: %s %s = %s", keyword, varName, value)
			fence.Variables = append(fence.Variables, ast.VariableNode{
				Keyword: keyword,
				Name:    varName,
				Value:   value,
			})
			i++
			continue
		}

		// Check for import statement
		if matches := importRegex.FindStringSubmatch(trimmedLine); matches != nil {
			importName := matches[1]
			importPath := matches[2]

			log.Printf("[parseFenceContent] Found import: %s from %s", importName, importPath)
			fence.Imports = append(fence.Imports, ast.ImportNode{
				Name: importName,
				Path: importPath,
			})
			i++
			continue
		}

		// Unknown line, skip it
		i++
	}

	log.Printf("[parseFenceContent] Extracted %d props, %d exported props, %d variables, %d functions, %d imports, %d stores",
		len(fence.Props), len(fence.ExportedProps), len(fence.Variables), len(fence.Functions), len(fence.Imports), len(fence.Stores))

	return fence
}

// parseFunctionBody extracts the complete function body from lines starting at startIndex
// Returns the full function body (including function declaration) and the ending line index
// Pattern: Brace Matching Parser [Load: 8]
// Cognitive Load: 8 (depth tracking, string handling, multi-line accumulation)
func parseFunctionBody(lines []string, startIndex int) (string, int) {
	var accumulator strings.Builder
	braceDepth := 0
	inString := false
	stringChar := rune(0)
	escaped := false

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]

		// Append line to accumulator
		if i > startIndex {
			accumulator.WriteString("\n")
		}
		accumulator.WriteString(line)

		// Process each character to track braces
		for _, char := range line {
			// Handle escape sequences
			if escaped {
				escaped = false
				continue
			}

			if char == '\\' {
				escaped = true
				continue
			}

			// Handle string literals
			if inString {
				if char == stringChar {
					inString = false
					stringChar = 0
				}
				continue
			}

			// Check if entering a string
			if char == '"' || char == '\'' || char == '`' {
				inString = true
				stringChar = char
				continue
			}

			// Track braces outside of strings
			if char == '{' {
				braceDepth++
			} else if char == '}' {
				braceDepth--
				if braceDepth == 0 {
					// Function complete
					return accumulator.String(), i
				}
			}
		}
	}

	// Unclosed function - return empty to indicate error
	log.Printf("[parseFunctionBody] WARNING: Unclosed function starting at line %d", startIndex)
	return "", -1
}

// isMultiLineValue checks if a value is likely to span multiple lines
// This includes:
// - Values starting with [ or { (arrays/objects)
// - Values containing ternary operators (? :)
// - Values ending with an opening bracket/brace (incomplete expression)
func isMultiLineValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 0 {
		return false
	}

	// Check if starts with array/object/paren
	firstChar := trimmed[0]
	if firstChar == '[' || firstChar == '{' || firstChar == '(' {
		return true
	}

	// Check if contains a ternary operator followed by [ or {
	// This handles cases like: isLoggedIn ? [...]
	if strings.Contains(trimmed, "?") {
		// Look for ? followed by whitespace and then [ or {
		ternaryPattern := regexp.MustCompile(`\?\s*[\[{]`)
		if ternaryPattern.MatchString(trimmed) {
			return true
		}
	}

	// Check if line ends with opening bracket/brace (incomplete)
	lastChar := trimmed[len(trimmed)-1]
	if lastChar == '[' || lastChar == '{' || lastChar == '(' {
		return true
	}

	return false
}

// parseMultiLineValue accumulates lines until all brackets/braces are matched
// Returns the full value string and the index of the last line consumed
// Now handles ternary operators properly
func parseMultiLineValue(lines []string, startIndex int, firstLineValue string) (string, int) {
	matcher := newBracketMatcher()
	ternaryMatcher := newTernaryMatcher()
	var accumulator strings.Builder

	log.Printf("[parseMultiLineValue] Starting with firstLineValue: %q, startIndex=%d, total lines=%d", firstLineValue, startIndex, len(lines))

	// Start with the first line value
	accumulator.WriteString(firstLineValue)

	// Process characters from first line
	for _, char := range firstLineValue {
		matcher.processChar(char)
		ternaryMatcher.processChar(char)
	}

	log.Printf("[parseMultiLineValue] After first line: bracketsComplete=%v, ternaryComplete=%v, stack depth=%d, ternary depth=%d",
		matcher.isComplete(), ternaryMatcher.isComplete(), len(matcher.stack), ternaryMatcher.depth)

	// Check if already complete (single-line case)
	// Must have both brackets matched AND ternary complete (or no ternary)
	if matcher.isComplete() && ternaryMatcher.isComplete() {
		log.Printf("[parseMultiLineValue] Single-line complete: %q", accumulator.String())
		return accumulator.String(), startIndex
	}

	// Continue with subsequent lines
	for i := startIndex + 1; i < len(lines); i++ {
		line := lines[i]

		log.Printf("[parseMultiLineValue] Processing line %d: %q", i, line)

		// Add newline before appending next line
		accumulator.WriteString("\n")
		accumulator.WriteString(line)

		// Process each character in this line
		for _, char := range line {
			matcher.processChar(char)
			ternaryMatcher.processChar(char)
		}

		log.Printf("[parseMultiLineValue] After line %d: bracketsComplete=%v, ternaryComplete=%v, stack depth=%d, ternary depth=%d",
			i, matcher.isComplete(), ternaryMatcher.isComplete(), len(matcher.stack), ternaryMatcher.depth)

		// Check if complete - both brackets and ternary must be complete
		if matcher.isComplete() && ternaryMatcher.isComplete() {
			log.Printf("[parseMultiLineValue] Multi-line complete at line %d: %q", i, accumulator.String())
			return accumulator.String(), i
		}
	}

	// Unclosed brackets or incomplete ternary - return empty string to indicate error
	// The caller will fall back to single-line parsing
	log.Printf("[parseMultiLineValue] Warning: incomplete expression starting at line %d, bracket depth=%d, ternary depth=%d",
		startIndex, len(matcher.stack), ternaryMatcher.depth)
	return "", -1
}

// ternaryMatcher tracks ternary operator (? :) nesting
type ternaryMatcher struct {
	depth      int // Number of ? without matching :
	inString   bool
	stringChar rune
	escaped    bool
}

func newTernaryMatcher() *ternaryMatcher {
	return &ternaryMatcher{
		depth:      0,
		inString:   false,
		stringChar: 0,
		escaped:    false,
	}
}

func (tm *ternaryMatcher) processChar(char rune) {
	// Handle escape sequences
	if tm.escaped {
		tm.escaped = false
		return
	}

	if char == '\\' {
		tm.escaped = true
		return
	}

	// Handle string literals
	if tm.inString {
		if char == tm.stringChar {
			tm.inString = false
			tm.stringChar = 0
		}
		return
	}

	// Check if entering a string
	if char == '"' || char == '\'' {
		tm.inString = true
		tm.stringChar = char
		return
	}

	// Track ternary operators outside of strings
	if char == '?' {
		tm.depth++
	} else if char == ':' {
		if tm.depth > 0 {
			tm.depth--
		}
	}
}

func (tm *ternaryMatcher) isComplete() bool {
	return tm.depth == 0 && !tm.inString
}

// ScriptParser parses the script section and returns an *ast.ScriptSection node.
func ScriptParser() Parser {
	return Map(
		Between(String("<script>"), String("</script>"), TakeUntil(String("</script>"))),
		func(value interface{}) (interface{}, error) {
			content := value.(string)
			log.Printf("[ScriptParser] Parsed script with %d chars", len(content))
			return &ast.ScriptSection{Content: content}, nil
		},
	)
}

// StyleParser parses the style section and returns an *ast.StyleSection node.
// Handles both <style> and <style ...attributes...>
func StyleParser() Parser {
	return func(input string) Result {
		// Check if starts with <style
		if !strings.HasPrefix(input, "<style") {
			return Result{nil, input, false, "not a style tag", false}
		}

		// Find the end of the opening tag (either > or />)
		openTagEnd := strings.Index(input, ">")
		if openTagEnd == -1 {
			return Result{nil, input, false, "unclosed style opening tag", false}
		}

		// Skip past the opening tag
		contentStart := openTagEnd + 1

		// Find the closing </style> tag
		closeTagStart := strings.Index(input[contentStart:], "</style>")
		if closeTagStart == -1 {
			return Result{nil, input, false, "missing </style> closing tag", false}
		}

		// Extract the content between tags
		content := input[contentStart : contentStart+closeTagStart]

		// Calculate remaining input after </style>
		remaining := input[contentStart+closeTagStart+len("</style>"):]

		log.Printf("[StyleParser] Parsed style with %d chars (with attributes support)", len(content))

		return Result{
			Value:      &ast.StyleSection{Content: content},
			Remaining:  remaining,
			Successful: true,
			Error:      "",
		}
	}
}

// ParseFenceContentWithStores is a wrapper for parseFenceContent that supports external store imports
// Pattern: Parser with Registry Support [Load: 6]
// Cognitive Load: 6 (delegate to parseFenceContent, then process store imports)
func ParseFenceContentWithStores(content string, storeRegistry map[string]string) *ast.FenceSection {
	// First, parse normally to get inline stores AND functions
	fence := parseFenceContent(content)

	// If no registry provided, return fence as-is
	if storeRegistry == nil {
		return fence
	}

	// Now process store imports
	// Pattern: import store from './stores/name.js'
	storeImportRegex := regexp.MustCompile(`^\s*import\s+store\s+from\s+['"](.+?)['"](?:;)?$`)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// Check for store import
		if matches := storeImportRegex.FindStringSubmatch(trimmedLine); matches != nil {
			importPath := matches[1]

			// Extract store name from path
			storeName := ExtractStoreNameFromPath(importPath)

			if storeName != "" {
				// Look up store content in registry
				if storeContent, exists := storeRegistry[storeName]; exists {
					// Only add if not already defined inline (inline overrides import)
					if _, alreadyDefined := fence.Stores[storeName]; !alreadyDefined {
						fence.Stores[storeName] = storeContent
						log.Printf("[ParseFenceContentWithStores] Loaded external store: %s from %s", storeName, importPath)
					} else {
						log.Printf("[ParseFenceContentWithStores] Inline store '%s' overrides imported store", storeName)
					}
				} else {
					log.Printf("[ParseFenceContentWithStores] WARNING: Store '%s' not found in registry (from %s)", storeName, importPath)
				}
			}
		}
	}

	return fence
}

// ExtractStoreNameFromPath extracts the store name from an import path
// Supports various path formats:
// - './stores/auth.js' -> 'auth'
// - 'stores/auth.js' -> 'auth'
// - '../stores/auth.js' -> 'auth'
// - '/stores/auth.js' -> 'auth'
// Pattern: Path Extraction Utility [Load: 6]
func ExtractStoreNameFromPath(path string) string {
	// Remove quotes if present
	path = strings.Trim(path, `"'`)

	// Find the last slash
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		// No slash, entire path is filename
		return strings.TrimSuffix(path, ".js")
	}

	// Extract filename after last slash
	filename := path[lastSlash+1:]

	// Remove .js extension
	storeName := strings.TrimSuffix(filename, ".js")

	return storeName
}
