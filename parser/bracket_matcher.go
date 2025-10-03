package parser

// BracketMatcher tracks bracket/brace/parenthesis matching for multi-line parsing.
// It uses a stack-based approach to ensure all opening brackets have matching closing brackets,
// and properly handles string literals to avoid counting brackets inside strings.
type BracketMatcher struct {
	stack      []rune // Stack of opening brackets: '[', '{', '('
	inString   bool   // True when inside a string literal
	stringChar rune   // The quote character that started the current string (' or ")
	escaped    bool   // True when the previous character was a backslash
	hasError   bool   // True if we encountered mismatched or extra closing brackets
}

// newBracketMatcher creates a new bracket matcher instance
func newBracketMatcher() *BracketMatcher {
	return &BracketMatcher{
		stack:      []rune{},
		inString:   false,
		stringChar: 0,
		escaped:    false,
		hasError:   false,
	}
}

// processChar processes a single character, updating the internal state
func (bm *BracketMatcher) processChar(char rune) {
	// Handle escape sequences
	if bm.escaped {
		bm.escaped = false
		return
	}

	if char == '\\' {
		bm.escaped = true
		return
	}

	// Handle string literals
	if bm.inString {
		// Check if this is the closing quote (same as opening quote)
		if char == bm.stringChar {
			bm.inString = false
			bm.stringChar = 0
		}
		// Inside string, ignore all other characters (including brackets)
		return
	}

	// Check if entering a string
	if char == '"' || char == '\'' {
		bm.inString = true
		bm.stringChar = char
		return
	}

	// Handle opening brackets/braces/parens
	if char == '[' || char == '{' || char == '(' {
		bm.stack = append(bm.stack, char)
		return
	}

	// Handle closing brackets/braces/parens
	if char == ']' || char == '}' || char == ')' {
		// Check if we have a matching opening bracket
		if len(bm.stack) > 0 {
			opening := bm.stack[len(bm.stack)-1]
			if bm.isMatchingPair(opening, char) {
				// Pop the matching opening bracket
				bm.stack = bm.stack[:len(bm.stack)-1]
			} else {
				// Mismatched bracket pair (e.g., '[' closed with '}')
				bm.hasError = true
			}
		} else {
			// Closing bracket with no matching opening bracket
			bm.hasError = true
		}
	}
}

// isMatchingPair checks if an opening and closing bracket match
func (bm *BracketMatcher) isMatchingPair(opening, closing rune) bool {
	pairs := map[rune]rune{
		'[': ']',
		'{': '}',
		'(': ')',
	}
	return pairs[opening] == closing
}

// isComplete returns true if all brackets are matched (stack is empty) and no errors
func (bm *BracketMatcher) isComplete() bool {
	return len(bm.stack) == 0 && !bm.inString && !bm.hasError
}

// hasOpenBrackets returns true if there are unclosed brackets
func (bm *BracketMatcher) hasOpenBrackets() bool {
	return len(bm.stack) > 0
}
