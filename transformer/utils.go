package transformer

import (
	"strings"
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
