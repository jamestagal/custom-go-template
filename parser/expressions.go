package parser

import (
	"log"
	"regexp"
	"strings"

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

// ExpressionParser parses expressions in curly braces: {variable}
func ExpressionParser() Parser {
	return func(input string) Result {
		// Use the lex-based expression parser
		exprRes := LexExpressionParser()(input)
		if !exprRes.Successful {
			return exprRes
		}

		expr, ok := exprRes.Value.(string)
		if !ok {
			return Result{nil, input, false, "expression parser did not return string", false}
		}

		log.Printf("[ExpressionParser] Found expression: '%s'", expr)
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

// parseFenceContent extracts props and variables from fence section content
// Now handles multi-line values for arrays and objects
func parseFenceContent(content string) *ast.FenceSection {
	fence := &ast.FenceSection{
		RawContent: content,
		Props:      []ast.PropNode{},
		Variables:  []ast.VariableNode{},
		Imports:    []ast.ImportNode{},
	}

	lines := strings.Split(content, "\n")

	// Regex patterns for parsing
	propRegex := regexp.MustCompile(`^\s*prop\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+)$`)
	varRegex := regexp.MustCompile(`^\s*(let|const|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+)$`)
	importRegex := regexp.MustCompile(`^\s*import\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s+from\s+['"](.+?)['"](?:;)?$`)

	// Process lines with multi-line support
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			i++
			continue
		}

		// Check for prop declaration
		if matches := propRegex.FindStringSubmatch(trimmedLine); matches != nil {
			propName := matches[1]
			firstLineValue := matches[2]

			// Check if this is a multi-line value (starts with [ or {)
			firstChar := strings.TrimSpace(firstLineValue)
			if len(firstChar) > 0 && (firstChar[0] == '[' || firstChar[0] == '{' || firstChar[0] == '(') {
				// Multi-line value - accumulate lines until brackets are matched
				fullValue, endIndex := parseMultiLineValue(lines, i, firstLineValue)
				if fullValue != "" {
					preview := fullValue
					if len(fullValue) > 50 {
						preview = fullValue[:50] + "..."
					}
					log.Printf("[parseFenceContent] Found multi-line prop: %s = %s (total %d chars)", propName, preview, len(fullValue))
					fence.Props = append(fence.Props, ast.PropNode{
						Name:         propName,
						DefaultValue: fullValue,
					})
					i = endIndex + 1
					continue
				}
			}

			// Single-line value
			// Remove trailing semicolon if present
			value := strings.TrimSpace(firstLineValue)
			value = strings.TrimSuffix(value, ";")

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

			// Check if this is a multi-line value
			firstChar := strings.TrimSpace(firstLineValue)
			if len(firstChar) > 0 && (firstChar[0] == '[' || firstChar[0] == '{' || firstChar[0] == '(') {
				// Multi-line value
				fullValue, endIndex := parseMultiLineValue(lines, i, firstLineValue)
				if fullValue != "" {
					preview := fullValue
					if len(fullValue) > 50 {
						preview = fullValue[:50] + "..."
					}
					log.Printf("[parseFenceContent] Found multi-line variable: %s %s = %s (total %d chars)", keyword, varName, preview, len(fullValue))
					fence.Variables = append(fence.Variables, ast.VariableNode{
						Keyword: keyword,
						Name:    varName,
						Value:   fullValue,
					})
					i = endIndex + 1
					continue
				}
			}

			// Single-line value
			value := strings.TrimSpace(firstLineValue)
			value = strings.TrimSuffix(value, ";")

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

	log.Printf("[parseFenceContent] Extracted %d props, %d variables, %d imports",
		len(fence.Props), len(fence.Variables), len(fence.Imports))

	return fence
}

// parseMultiLineValue accumulates lines until all brackets/braces are matched
// Returns the full value string and the index of the last line consumed
func parseMultiLineValue(lines []string, startIndex int, firstLineValue string) (string, int) {
	matcher := newBracketMatcher()
	var accumulator strings.Builder

	log.Printf("[parseMultiLineValue] Starting with firstLineValue: %q, startIndex=%d, total lines=%d", firstLineValue, startIndex, len(lines))

	// Start with the first line value
	accumulator.WriteString(firstLineValue)

	// Process characters from first line
	for _, char := range firstLineValue {
		matcher.processChar(char)
	}

	log.Printf("[parseMultiLineValue] After first line: isComplete=%v, stack depth=%d", matcher.isComplete(), len(matcher.stack))

	// Check if already complete (single-line case)
	if matcher.isComplete() {
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
		}

		log.Printf("[parseMultiLineValue] After line %d: isComplete=%v, stack depth=%d", i, matcher.isComplete(), len(matcher.stack))

		// Check if complete
		if matcher.isComplete() {
			log.Printf("[parseMultiLineValue] Multi-line complete at line %d: %q", i, accumulator.String())
			return accumulator.String(), i
		}
	}

	// Unclosed brackets - return empty string to indicate error
	// The caller will fall back to single-line parsing
	log.Printf("[parseMultiLineValue] Warning: unclosed brackets in multi-line value starting at line %d, final stack depth=%d", startIndex, len(matcher.stack))
	return "", -1
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
func StyleParser() Parser {
	return Map(
		Between(String("<style>"), String("</style>"), TakeUntil(String("</style>"))),
		func(value interface{}) (interface{}, error) {
			content := value.(string)
			log.Printf("[StyleParser] Parsed style with %d chars", len(content))
			return &ast.StyleSection{Content: content}, nil
		},
	)
}
