package parser

import (
	"fmt"
	"log"
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

// ExpressionParser parses an {expression} or {{expression}} and returns an *ast.ExpressionNode
func ExpressionParser() Parser {
	return func(input string) Result {
		log.Printf("[ExpressionParser] Starting on: '%.30s...'", input)

		// Check for double braces first
		if strings.HasPrefix(input, "{{") {
			exprRes := Between(String("{{"), String("}}"), TakeUntil(String("}}")))(input)
			if exprRes.Successful {
				expressionContent := strings.TrimSpace(exprRes.Value.(string))
				log.Printf("[ExpressionParser] Parsed double-brace expression: %s", expressionContent)
				return Result{
					&ast.ExpressionNode{Expression: expressionContent},
					exprRes.Remaining,
					true,
					"",
					false,
				}
			}
		}

		// Then check for single braces
		if !strings.HasPrefix(input, "{") {
			return Result{nil, input, false, "not an expression", false}
		}

		// Check if it's a directive - must be done before attempting to parse as expression
		if isDirective(input) {
			log.Printf("[ExpressionParser] Looks like a directive, not a simple expression")
			return Result{nil, input, false, "looks like a directive, not a simple expression", false}
		}

		// Parse the expression content
		exprRes := Between(String("{"), String("}"), TakeUntil(String("}")))(input)
		if !exprRes.Successful {
			log.Printf("[ExpressionParser] Failed to parse expression: %s", exprRes.Error)
			return Result{nil, input, false, fmt.Sprintf("failed to parse expression: %s", exprRes.Error), false}
		}

		expressionContent := strings.TrimSpace(exprRes.Value.(string))
		log.Printf("[ExpressionParser] Parsed single-brace expression: %s", expressionContent)
		return Result{
			&ast.ExpressionNode{Expression: expressionContent},
			exprRes.Remaining,
			true,
			"",
			false,
		}
	}
}

// isDirective checks if an input string appears to be a directive
func isDirective(input string) bool {
	directives := []string{
		"{if", "{else if", "{else}", "{/if}",
		"{for", "{/for}", "{await", "{/await}",
		"{#if", "{#each", "{:else}", "{:else if", "{/if}", "{/#if}", "{/each}", "{/#each}",
		"{#", "{/", "{:", "{@",
	}

	for _, prefix := range directives {
		if strings.HasPrefix(input, prefix) {
			return true
		}
	}

	return false
}

// FenceParser parses the fence section (---...---) and returns an *ast.FenceSection node.
func FenceParser() Parser {
	return Map(
		Between(String("---"), String("---"), TakeUntil(String("---"))),
		func(value interface{}) (interface{}, error) {
			content := value.(string)
			log.Printf("[FenceParser] Parsed fence with %d chars", len(content))
			return &ast.FenceSection{RawContent: content}, nil
		},
	)
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
