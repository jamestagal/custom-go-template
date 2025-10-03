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

// ExpressionParser parses an {expression} or {expression} and returns an *ast.ExpressionNode
// This version is more flexible with whitespace inside the braces
func ExpressionParser() Parser {
	return func(input string) Result {
		log.Printf("[ExpressionParser] Starting on: '%.30s...'", input)

		// Check if it starts with a brace
		if !strings.HasPrefix(input, "{") {
			return Result{nil, input, false, "not an expression", false}
		}

		// Check if it's a directive - must be done before attempting to parse as expression
		if isDirective(input) {
			log.Printf("[ExpressionParser] Looks like a directive, not a simple expression")
			return Result{nil, input, false, "looks like a directive, not a simple expression", false}
		}

		// Manual parsing with whitespace handling
		i := 1 // Skip the opening brace

		// Track the expression content
		start := i

		// Find the closing brace, handling nested braces
		braceDepth := 1
		for i < len(input) && braceDepth > 0 {
			if input[i] == '{' {
				braceDepth++
			} else if input[i] == '}' {
				braceDepth--
			}

			if braceDepth > 0 {
				i++
			}
		}

		// If we found a closing brace
		if i < len(input) && input[i] == '}' {
			expressionContent := strings.TrimSpace(input[start:i])
			log.Printf("[ExpressionParser] Parsed expression with whitespace handling: %s", expressionContent)

			return Result{
				&ast.ExpressionNode{Expression: expressionContent},
				input[i+1:],
				true,
				"",
				false,
			}
		}

		log.Printf("[ExpressionParser] Failed to find closing brace for expression")
		return Result{nil, input, false, "unclosed expression", false}
	}
}

// isDirective checks if an input string appears to be a directive, handling whitespace
func isDirective(input string) bool {
	// Trim whitespace at the start for consistent checking
	trimmed := strings.TrimLeft(input, " \t\n\r")

	// First character must be {
	if !strings.HasPrefix(trimmed, "{") {
		return false
	}

	// Check for directive prefixes after the opening brace and potential whitespace
	i := 1
	// Skip whitespace after the opening brace
	for i < len(trimmed) && (trimmed[i] == ' ' || trimmed[i] == '\t') {
		i++
	}

	// Now check for directive keywords
	if i < len(trimmed) {
		prefixes := []string{
			"if ", "#if ", "else", "/if", "end",
			"for ", "#each ", "/for", "#for", "/each", "/#each",
			"await ", "/await", "#await", "/await",
		}

		for _, prefix := range prefixes {
			if strings.HasPrefix(trimmed[i:], prefix) {
				return true
			}
		}
	}

	return false
}

// FenceParser parses the fence section (---...---) and returns an *ast.FenceSection node.
// This now properly extracts props and variables from the fence content.
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
func parseFenceContent(content string) *ast.FenceSection {
	fence := &ast.FenceSection{
		RawContent: content,
		Props:      []ast.PropNode{},
		Variables:  []ast.VariableNode{},
		Imports:    []ast.ImportNode{},
	}

	lines := strings.Split(content, "\n")

	// Regex patterns for parsing
	// Note: Use (?:;)? to optionally match trailing semicolon, then trim it
	propRegex := regexp.MustCompile(`^\s*prop\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+?)(?:;)?$`)
	varRegex := regexp.MustCompile(`^\s*(let|const|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(.+?)(?:;)?$`)
	importRegex := regexp.MustCompile(`^\s*import\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s+from\s+['"](.+?)['"](?:;)?$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for prop declaration
		if matches := propRegex.FindStringSubmatch(line); matches != nil {
			propName := matches[1]
			propValue := strings.TrimSpace(matches[2])

			log.Printf("[parseFenceContent] Found prop: %s = %s", propName, propValue)

			fence.Props = append(fence.Props, ast.PropNode{
				Name:         propName,
				DefaultValue: propValue,
			})
			continue
		}

		// Check for variable declaration
		if matches := varRegex.FindStringSubmatch(line); matches != nil {
			keyword := matches[1]
			varName := matches[2]
			varValue := strings.TrimSpace(matches[3])

			log.Printf("[parseFenceContent] Found variable: %s %s = %s", keyword, varName, varValue)

			fence.Variables = append(fence.Variables, ast.VariableNode{
				Keyword: keyword,
				Name:    varName,
				Value:   varValue,
			})
			continue
		}

		// Check for import statement
		if matches := importRegex.FindStringSubmatch(line); matches != nil {
			importName := matches[1]
			importPath := matches[2]

			log.Printf("[parseFenceContent] Found import: %s from %s", importName, importPath)

			fence.Imports = append(fence.Imports, ast.ImportNode{
				Name: importName,
				Path: importPath,
			})
			continue
		}
	}

	log.Printf("[parseFenceContent] Extracted %d props, %d variables, %d imports",
		len(fence.Props), len(fence.Variables), len(fence.Imports))

	return fence
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
