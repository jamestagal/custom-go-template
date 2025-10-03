package parser

import (
	"fmt"
	"strings"
)

// Position represents a location in the source template
type Position struct {
	Line   int
	Column int
	Offset int // Byte offset from start of input
}

// ParseError represents a detailed parsing error with context
type ParseError struct {
	Message    string
	Position   Position
	Input      string   // Full input for context extraction
	Suggestion string   // Helpful suggestion for fixing the error
	Context    string   // Snippet of code showing the problem
}

// Error implements the error interface
func (e *ParseError) Error() string {
	var sb strings.Builder

	// Main error message
	sb.WriteString(fmt.Sprintf("Parse error at line %d, column %d: %s\n",
		e.Position.Line, e.Position.Column, e.Message))

	// Show code context if available
	if e.Context != "" {
		sb.WriteString("\n")
		sb.WriteString(e.Context)
		sb.WriteString("\n")
	}

	// Add suggestion if available
	if e.Suggestion != "" {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Suggestion: %s\n", e.Suggestion))
	}

	return sb.String()
}

// GetPosition calculates line and column from input and offset
func GetPosition(input string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(input) {
		offset = len(input)
	}

	line := 1
	column := 1

	for i := 0; i < offset && i < len(input); i++ {
		if input[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return Position{
		Line:   line,
		Column: column,
		Offset: offset,
	}
}

// ExtractContext extracts a few lines of context around the error position
func ExtractContext(input string, pos Position, contextLines int) string {
	if contextLines < 1 {
		contextLines = 1
	}

	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return ""
	}

	// Adjust to 0-based indexing
	errorLine := pos.Line - 1
	if errorLine < 0 {
		errorLine = 0
	}
	if errorLine >= len(lines) {
		errorLine = len(lines) - 1
	}

	// Calculate range of lines to show
	startLine := errorLine - contextLines
	if startLine < 0 {
		startLine = 0
	}

	endLine := errorLine + contextLines + 1
	if endLine > len(lines) {
		endLine = len(lines)
	}

	var sb strings.Builder

	// Build context with line numbers
	for i := startLine; i < endLine; i++ {
		lineNum := i + 1
		prefix := "  "
		if i == errorLine {
			prefix = "→ "
		}

		sb.WriteString(fmt.Sprintf("%s%4d | %s\n", prefix, lineNum, lines[i]))

		// Add error pointer on the line after the error
		if i == errorLine {
			sb.WriteString("       | ")
			// Add spaces to align with error column
			for j := 1; j < pos.Column; j++ {
				sb.WriteString(" ")
			}
			sb.WriteString("^\n")
		}
	}

	return sb.String()
}

// NewParseError creates a detailed parse error with context
func NewParseError(message string, input string, offset int) *ParseError {
	pos := GetPosition(input, offset)
	context := ExtractContext(input, pos, 2)

	return &ParseError{
		Message:  message,
		Position: pos,
		Input:    input,
		Context:  context,
	}
}

// NewParseErrorWithSuggestion creates a parse error with a helpful suggestion
func NewParseErrorWithSuggestion(message, suggestion, input string, offset int) *ParseError {
	err := NewParseError(message, input, offset)
	err.Suggestion = suggestion
	return err
}

// Common error patterns with suggestions

// ErrUnclosedTag creates an error for unclosed HTML tags
func ErrUnclosedTag(tagName, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		fmt.Sprintf("unclosed tag <%s>", tagName),
		fmt.Sprintf("Add </%s> to close the tag", tagName),
		input,
		offset,
	)
}

// ErrUnclosedBrace creates an error for unclosed braces in expressions
func ErrUnclosedBrace(input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		"unclosed brace in expression",
		"Add } to close the expression",
		input,
		offset,
	)
}

// ErrMissingClosingDirective creates an error for missing directive closures
func ErrMissingClosingDirective(directive, input string, offset int) *ParseError {
	suggestion := ""
	switch directive {
	case "if":
		suggestion = "Add {/if} to close the conditional block"
	case "for":
		suggestion = "Add {/for} to close the loop"
	case "each":
		suggestion = "Add {/each} to close the loop"
	default:
		suggestion = fmt.Sprintf("Add {/%s} to close the block", directive)
	}

	return NewParseErrorWithSuggestion(
		fmt.Sprintf("missing closing directive for {%s}", directive),
		suggestion,
		input,
		offset,
	)
}

// ErrInvalidExpression creates an error for invalid expressions
func ErrInvalidExpression(expr, reason, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		fmt.Sprintf("invalid expression: %s", reason),
		"Check that the expression syntax is correct",
		input,
		offset,
	)
}

// ErrInvalidDirective creates an error for invalid directive syntax
func ErrInvalidDirective(directive, expected, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		fmt.Sprintf("invalid %s directive syntax", directive),
		fmt.Sprintf("Expected format: %s", expected),
		input,
		offset,
	)
}

// ErrUnexpectedToken creates an error for unexpected tokens
func ErrUnexpectedToken(found, expected, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		fmt.Sprintf("unexpected token '%s'", found),
		fmt.Sprintf("Expected %s", expected),
		input,
		offset,
	)
}

// ErrMaxDepthExceeded creates an error for exceeding nesting depth
func ErrMaxDepthExceeded(depth int, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		fmt.Sprintf("maximum nesting depth (%d) exceeded", depth),
		"Reduce nesting depth or check for infinite recursion",
		input,
		offset,
	)
}

// ErrInvalidLoopSyntax creates an error for invalid loop syntax
func ErrInvalidLoopSyntax(expr, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		"invalid loop syntax",
		"Use format: {for item in items} or {for item, index in items}",
		input,
		offset,
	)
}

// ErrInvalidAttributeSyntax creates an error for invalid attribute syntax
func ErrInvalidAttributeSyntax(attrName, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		fmt.Sprintf("invalid attribute syntax: %s", attrName),
		"Check attribute name and value format",
		input,
		offset,
	)
}

// ErrUnexpectedEOF creates an error for unexpected end of file
func ErrUnexpectedEOF(expected, input string, offset int) *ParseError {
	return NewParseErrorWithSuggestion(
		"unexpected end of file",
		fmt.Sprintf("Expected %s before end of file", expected),
		input,
		offset,
	)
}
