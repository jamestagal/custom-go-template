package parser

import (
	"strings"
)

// String creates a parser that matches a specific string
func String(match string) Parser {
	return func(input string) Result {
		if strings.HasPrefix(input, match) {
			return Result{match, input[len(match):], true, "", false} // Added Dynamic
		}
		return Result{nil, input, false, "string not matched: " + match, false} // Added Dynamic
	}
}

// AnyChar parses a single character
func AnyChar() Parser {
	return func(input string) Result {
		if len(input) == 0 {
			return Result{nil, input, false, "end of input", false} // Added Dynamic
		}
		return Result{string(input[0]), input[1:], true, "", false} // Added Dynamic
	}
}

// TakeUntil consumes input until the delimiter parser succeeds
func TakeUntil(delimiter Parser) Parser {
	return func(input string) Result {
		var consumed strings.Builder
		current := input
		for len(current) > 0 {
			res := delimiter(current)
			if res.Successful {
				return Result{consumed.String(), current, true, "", false} // Added Dynamic
			}
			charRes := AnyChar()(current)
			if !charRes.Successful {
				return Result{nil, input, false, "failed to consume character", false} // Added Dynamic
			}
			consumed.WriteString(charRes.Value.(string))
			current = charRes.Remaining
		}
		return Result{nil, input, false, "delimiter not found", false} // Added Dynamic
	}
}

// TakeUntilAny consumes input until any of the delimiter parsers succeed
func TakeUntilAny(delimiters ...Parser) Parser {
	delimiterChoice := Choice(delimiters...)
	return func(input string) Result {
		var consumed strings.Builder
		current := input
		for len(current) > 0 {
			res := delimiterChoice(current)
			if res.Successful {
				return Result{consumed.String(), current, true, "", false} // Added Dynamic
			}
			charRes := AnyChar()(current)
			if !charRes.Successful {
				return Result{nil, input, false, "failed to consume character", false} // Added Dynamic
			}
			consumed.WriteString(charRes.Value.(string))
			current = charRes.Remaining
		}
		return Result{consumed.String(), current, true, "", false} // Added Dynamic
	}
}

// Whitespace consumes whitespace characters
func Whitespace() Parser {
	return func(input string) Result {
		i := 0
		for i < len(input) && (input[i] == ' ' || input[i] == '\t' || input[i] == '\n' || input[i] == '\r') {
			i++
		}
		return Result{input[:i], input[i:], true, "", false} // Added Dynamic
	}
}

// Identifier parses HTML tag names and attribute names
func Identifier() Parser {
	return func(input string) Result {
		if len(input) == 0 {
			return Result{nil, input, false, "empty input", false} // Added Dynamic
		}

		// Check first character - allow @, : for Alpine directives
		firstChar := input[0]
		if !((firstChar >= 'a' && firstChar <= 'z') ||
			(firstChar >= 'A' && firstChar <= 'Z') ||
			firstChar == '@' || firstChar == ':' ||
			firstChar == '_') {
			return Result{nil, input, false, "not an identifier", false} // Added Dynamic
		}

		// Continue with remaining characters
		i := 1
		for i < len(input) && ((input[i] >= 'a' && input[i] <= 'z') ||
			(input[i] >= 'A' && input[i] <= 'Z') ||
			(input[i] >= '0' && input[i] <= '9') ||
			input[i] == '-' || input[i] == '_' ||
			input[i] == ':') { // Allow : in the middle of an identifier too
			i++
		}

		return Result{input[:i], input[i:], true, "", false} // Added Dynamic
	}
}

// DoctypeParser parses HTML DOCTYPE declarations
func DoctypeParser() Parser {
	return func(input string) Result {
		start := "<!DOCTYPE"
		// Case-insensitive check for <!DOCTYPE
		if len(input) < len(start) || !strings.EqualFold(input[:len(start)], start) {
			return Result{nil, input, false, "not a doctype", false} // Added Dynamic
		}
		endPos := strings.Index(input, ">")
		if endPos == -1 {
			return Result{nil, input, false, "doctype not closed", false} // Added Dynamic
		}
		// Consume the doctype but return nil value so Many ignores it
		return Result{nil, input[endPos+1:], true, "", false} // Added Dynamic
	}
}
