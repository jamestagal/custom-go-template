package parser

import (
	"testing"
)

func TestBracketMatcher_SimpleArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "simple array",
			input:    "[1, 2, 3]",
			expected: true,
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: true,
		},
		{
			name:     "array with strings",
			input:    `["a", "b", "c"]`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newBracketMatcher()
			for _, char := range tt.input {
				matcher.processChar(char)
			}

			if got := matcher.isComplete(); got != tt.expected {
				t.Errorf("isComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBracketMatcher_SimpleObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "simple object",
			input:    `{a: 1, b: 2}`,
			expected: true,
		},
		{
			name:     "empty object",
			input:    `{}`,
			expected: true,
		},
		{
			name:     "object with string values",
			input:    `{name: "John", age: "30"}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newBracketMatcher()
			for _, char := range tt.input {
				matcher.processChar(char)
			}

			if got := matcher.isComplete(); got != tt.expected {
				t.Errorf("isComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBracketMatcher_NestedStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "array in object",
			input:    `{a: [1, 2], b: {c: 3}}`,
			expected: true,
		},
		{
			name:     "object in array",
			input:    `[{a: 1}, {b: 2}]`,
			expected: true,
		},
		{
			name:     "deeply nested",
			input:    `{a: {b: {c: [1, 2, 3]}}}`,
			expected: true,
		},
		{
			name:     "mixed nesting",
			input:    `[{a: [1, 2]}, {b: {c: 3}}]`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newBracketMatcher()
			for _, char := range tt.input {
				matcher.processChar(char)
			}

			if got := matcher.isComplete(); got != tt.expected {
				t.Errorf("isComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBracketMatcher_StringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "string with brackets",
			input:    `{msg: "test [array]"}`,
			expected: true,
		},
		{
			name:     "string with braces",
			input:    `{msg: "test {object}"}`,
			expected: true,
		},
		{
			name:     "escaped quotes with double",
			input:    `{msg: "it's \"quoted\""}`,
			expected: true,
		},
		{
			name:     "escaped quotes with single",
			input:    `{msg: 'it\'s "quoted"'}`,
			expected: true,
		},
		{
			name:     "mixed quotes",
			input:    `{a: "test's", b: 'he said "hello"'}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newBracketMatcher()
			for _, char := range tt.input {
				matcher.processChar(char)
			}

			if got := matcher.isComplete(); got != tt.expected {
				t.Errorf("isComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBracketMatcher_Incomplete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "unclosed array",
			input:    "[1, 2, 3",
			expected: false,
		},
		{
			name:     "unclosed object",
			input:    "{a: 1, b: 2",
			expected: false,
		},
		{
			name:     "nested unclosed",
			input:    "{a: [1, 2}",
			expected: false,
		},
		{
			name:     "extra closing bracket",
			input:    "[1, 2]]",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newBracketMatcher()
			for _, char := range tt.input {
				matcher.processChar(char)
			}

			if got := matcher.isComplete(); got != tt.expected {
				t.Errorf("isComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBracketMatcher_Parentheses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "function call",
			input:    "new Date().getFullYear()",
			expected: true,
		},
		{
			name:     "function in array",
			input:    "[func(a, b), func2()]",
			expected: true,
		},
		{
			name:     "unclosed parenthesis",
			input:    "func(a, b",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newBracketMatcher()
			for _, char := range tt.input {
				matcher.processChar(char)
			}

			if got := matcher.isComplete(); got != tt.expected {
				t.Errorf("isComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBracketMatcher_RealWorldFooterLinks(t *testing.T) {
	// This is the actual multi-line array from Footer.html
	input := `[
  { label: "Home", url: "/" },
  { label: "About", url: "/about" },
  { label: "Products", url: "/products" },
  { label: "Contact", url: "/contact" },
  { label: "Terms of Service", url: "/terms" },
  { label: "Privacy Policy", url: "/privacy" }
]`

	matcher := newBracketMatcher()
	for _, char := range input {
		matcher.processChar(char)
	}

	if !matcher.isComplete() {
		t.Errorf("Footer links array should be complete, but isComplete() = false")
	}
}

func TestBracketMatcher_RealWorldSocialLinks(t *testing.T) {
	// This is the actual multi-line array from Footer.html
	input := `[
  { label: "Twitter", url: "#", icon: "🐦" },
  { label: "Facebook", url: "#", icon: "📘" },
  { label: "Instagram", url: "#", icon: "📷" },
  { label: "GitHub", url: "#", icon: "🐙" }
]`

	matcher := newBracketMatcher()
	for _, char := range input {
		matcher.processChar(char)
	}

	if !matcher.isComplete() {
		t.Errorf("Social links array should be complete, but isComplete() = false")
	}
}
