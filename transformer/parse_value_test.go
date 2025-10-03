package transformer

import (
	"reflect"
	"testing"
)

// TestParseValue tests the parseValue helper function that parses JavaScript literal values
// from fence section strings into appropriate Go types or Alpine.js-compatible strings.
func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		testType string // Type validation: "bool", "nil", "int", "float", "string", etc.
	}{
		// Boolean values
		{
			name:     "boolean true",
			input:    "true",
			expected: true,
			testType: "bool",
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: false,
			testType: "bool",
		},
		{
			name:     "boolean true with whitespace",
			input:    "  true  ",
			expected: true,
			testType: "bool",
		},
		{
			name:     "boolean false with whitespace",
			input:    "  false  ",
			expected: false,
			testType: "bool",
		},

		// Null values
		{
			name:     "null value",
			input:    "null",
			expected: nil,
			testType: "nil",
		},
		{
			name:     "null with whitespace",
			input:    "  null  ",
			expected: nil,
			testType: "nil",
		},

		// Integer numbers
		{
			name:     "positive integer",
			input:    "42",
			expected: 42,
			testType: "int",
		},
		{
			name:     "negative integer",
			input:    "-42",
			expected: -42,
			testType: "int",
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
			testType: "int",
		},
		{
			name:     "integer with whitespace",
			input:    "  100  ",
			expected: 100,
			testType: "int",
		},

		// Float numbers
		{
			name:     "positive float",
			input:    "3.14",
			expected: 3.14,
			testType: "float",
		},
		{
			name:     "negative float",
			input:    "-3.14",
			expected: -3.14,
			testType: "float",
		},
		{
			name:     "float with leading zero",
			input:    "0.5",
			expected: 0.5,
			testType: "float",
		},
		{
			name:     "float without leading zero",
			input:    ".5",
			expected: 0.5,
			testType: "float",
		},
		{
			name:     "float with whitespace",
			input:    "  2.5  ",
			expected: 2.5,
			testType: "float",
		},

		// Quoted strings (double quotes)
		{
			name:     "double quoted string",
			input:    `"hello"`,
			expected: "hello",
			testType: "string",
		},
		{
			name:     "double quoted empty string",
			input:    `""`,
			expected: "",
			testType: "string",
		},
		{
			name:     "double quoted string with spaces",
			input:    `"hello world"`,
			expected: "hello world",
			testType: "string",
		},
		{
			name:     "double quoted string with escaped quotes",
			input:    `"hello \"world\""`,
			expected: `hello "world"`,
			testType: "string",
		},
		{
			name:     "double quoted string with whitespace around",
			input:    `  "test"  `,
			expected: "test",
			testType: "string",
		},

		// Quoted strings (single quotes)
		{
			name:     "single quoted string",
			input:    `'hello'`,
			expected: "hello",
			testType: "string",
		},
		{
			name:     "single quoted empty string",
			input:    `''`,
			expected: "",
			testType: "string",
		},
		{
			name:     "single quoted string with spaces",
			input:    `'hello world'`,
			expected: "hello world",
			testType: "string",
		},
		{
			name:     "single quoted string with escaped quotes",
			input:    `'hello \'world\''`,
			expected: `hello 'world'`,
			testType: "string",
		},

		// Arrays (returned as strings for Alpine.js)
		{
			name:     "empty array",
			input:    "[]",
			expected: "[]",
			testType: "string",
		},
		{
			name:     "array of numbers",
			input:    "[1, 2, 3]",
			expected: "[1, 2, 3]",
			testType: "string",
		},
		{
			name:     "array of strings",
			input:    `["a", "b", "c"]`,
			expected: `["a", "b", "c"]`,
			testType: "string",
		},
		{
			name:     "nested array",
			input:    "[[1, 2], [3, 4]]",
			expected: "[[1, 2], [3, 4]]",
			testType: "string",
		},
		{
			name:     "array with mixed types",
			input:    `[1, "two", true, null]`,
			expected: `[1, "two", true, null]`,
			testType: "string",
		},
		{
			name:     "array with whitespace",
			input:    "  [1, 2, 3]  ",
			expected: "[1, 2, 3]",
			testType: "string",
		},

		// Objects (returned as strings for Alpine.js)
		{
			name:     "empty object",
			input:    "{}",
			expected: "{}",
			testType: "string",
		},
		{
			name:     "simple object",
			input:    `{ key: "value" }`,
			expected: `{ key: "value" }`,
			testType: "string",
		},
		{
			name:     "object with multiple properties",
			input:    `{ name: "John", age: 30 }`,
			expected: `{ name: "John", age: 30 }`,
			testType: "string",
		},
		{
			name:     "nested object",
			input:    `{ user: { name: "John" } }`,
			expected: `{ user: { name: "John" } }`,
			testType: "string",
		},
		{
			name:     "object with array property",
			input:    `{ items: [1, 2, 3] }`,
			expected: `{ items: [1, 2, 3] }`,
			testType: "string",
		},
		{
			name:     "object with whitespace",
			input:    `  { key: "value" }  `,
			expected: `{ key: "value" }`,
			testType: "string",
		},

		// Expressions (variable references, function calls - returned as strings)
		{
			name:     "variable reference",
			input:    "myVariable",
			expected: "myVariable",
			testType: "string",
		},
		{
			name:     "function call",
			input:    "formatPrice(item.price)",
			expected: "formatPrice(item.price)",
			testType: "string",
		},
		{
			name:     "property access",
			input:    "user.name",
			expected: "user.name",
			testType: "string",
		},
		{
			name:     "array index access",
			input:    "items[0]",
			expected: "items[0]",
			testType: "string",
		},
		{
			name:     "chained method calls",
			input:    "user.getName().toUpperCase()",
			expected: "user.getName().toUpperCase()",
			testType: "string",
		},
		{
			name:     "ternary expression",
			input:    "isActive ? 'yes' : 'no'",
			expected: "isActive ? 'yes' : 'no'",
			testType: "string",
		},
		{
			name:     "binary expression",
			input:    "count + 1",
			expected: "count + 1",
			testType: "string",
		},
		{
			name:     "template literal",
			input:    "`Hello ${name}`",
			expected: "`Hello ${name}`",
			testType: "string",
		},

		// Edge cases
		{
			name:     "empty string input",
			input:    "",
			expected: "",
			testType: "string",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
			testType: "string",
		},
		{
			name:     "string that looks like boolean",
			input:    `"true"`,
			expected: "true",
			testType: "string",
		},
		{
			name:     "string that looks like number",
			input:    `"42"`,
			expected: "42",
			testType: "string",
		},
		{
			name:     "string that looks like null",
			input:    `"null"`,
			expected: "null",
			testType: "string",
		},
		{
			name:     "malformed number",
			input:    "12.34.56",
			expected: "12.34.56", // Should return as string if can't parse
			testType: "string",
		},
		{
			name:     "unclosed string double quote",
			input:    `"hello`,
			expected: `"hello`, // Return as-is if malformed
			testType: "string",
		},
		{
			name:     "unclosed string single quote",
			input:    `'hello`,
			expected: `'hello`, // Return as-is if malformed
			testType: "string",
		},
		{
			name:     "unmatched brackets in array",
			input:    "[1, 2, 3",
			expected: "[1, 2, 3", // Return as string for Alpine.js to handle
			testType: "string",
		},
		{
			name:     "unmatched braces in object",
			input:    "{ key: 'value'",
			expected: "{ key: 'value'", // Return as string for Alpine.js to handle
			testType: "string",
		},
		{
			name:     "undefined keyword",
			input:    "undefined",
			expected: "undefined", // Return as string (variable reference)
			testType: "string",
		},
		{
			name:     "NaN keyword",
			input:    "NaN",
			expected: "NaN", // Return as string (variable reference)
			testType: "string",
		},
		{
			name:     "Infinity keyword",
			input:    "Infinity",
			expected: "Infinity", // Return as string (variable reference)
			testType: "string",
		},

		// Special string cases
		{
			name:     "string with newlines",
			input:    `"hello\nworld"`,
			expected: "hello\nworld",
			testType: "string",
		},
		{
			name:     "string with tabs",
			input:    `"hello\tworld"`,
			expected: "hello\tworld",
			testType: "string",
		},
		{
			name:     "string with unicode",
			input:    `"hello 世界"`,
			expected: "hello 世界",
			testType: "string",
		},

		// Numeric edge cases
		{
			name:     "scientific notation",
			input:    "1e5",
			expected: 1e5,
			testType: "float",
		},
		{
			name:     "negative scientific notation",
			input:    "-1.5e-3",
			expected: -1.5e-3,
			testType: "float",
		},
		{
			name:     "very large integer",
			input:    "9007199254740991", // Number.MAX_SAFE_INTEGER
			expected: 9007199254740991,
			testType: "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)

			// Check if result matches expected value
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseValue(%q) = %v (type: %T), want %v (type: %T)",
					tt.input, result, result, tt.expected, tt.expected)
			}

			// Additional type validation
			switch tt.testType {
			case "bool":
				if _, ok := result.(bool); !ok {
					t.Errorf("parseValue(%q) should return bool, got %T", tt.input, result)
				}
			case "nil":
				if result != nil {
					t.Errorf("parseValue(%q) should return nil, got %v", tt.input, result)
				}
			case "int":
				if _, ok := result.(int); !ok {
					if _, ok := result.(int64); !ok {
						t.Errorf("parseValue(%q) should return int or int64, got %T", tt.input, result)
					}
				}
			case "float":
				if _, ok := result.(float64); !ok {
					t.Errorf("parseValue(%q) should return float64, got %T", tt.input, result)
				}
			case "string":
				if _, ok := result.(string); !ok {
					t.Errorf("parseValue(%q) should return string, got %T", tt.input, result)
				}
			}
		})
	}
}

// TestParseValueEdgeCases tests additional edge cases and error scenarios
func TestParseValueEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldPanic bool
		description string
	}{
		{
			name:        "nil handling - should not panic",
			input:       "",
			shouldPanic: false,
			description: "Empty string should return empty string, not panic",
		},
		{
			name:        "extremely long string",
			input:       `"` + string(make([]byte, 10000)) + `"`,
			shouldPanic: false,
			description: "Very long strings should be handled gracefully",
		},
		{
			name:        "deeply nested structure",
			input:       "[[[[[[1]]]]]]",
			shouldPanic: false,
			description: "Deeply nested arrays should return as string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("parseValue(%q) panicked unexpectedly: %v", tt.input, r)
					}
				} else if tt.shouldPanic {
					t.Errorf("parseValue(%q) should have panicked but didn't", tt.input)
				}
			}()

			result := parseValue(tt.input)
			// If we get here without panic, the function handled the input
			t.Logf("parseValue(%q) = %v (type: %T) - %s", tt.input, result, result, tt.description)
		})
	}
}

// TestParseValueQuoteStripping validates that quotes are properly removed from strings
func TestParseValueQuoteStripping(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue string
		description   string
	}{
		{
			name:          "double quotes removed",
			input:         `"test"`,
			expectedValue: "test",
			description:   "Double quotes should be stripped",
		},
		{
			name:          "single quotes removed",
			input:         `'test'`,
			expectedValue: "test",
			description:   "Single quotes should be stripped",
		},
		{
			name:          "quotes inside string preserved",
			input:         `"he said 'hello'"`,
			expectedValue: "he said 'hello'",
			description:   "Quotes inside string should be preserved",
		},
		{
			name:          "mixed quotes",
			input:         `'she said "hi"'`,
			expectedValue: `she said "hi"`,
			description:   "Mixed quotes should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)
			if result != tt.expectedValue {
				t.Errorf("%s: got %q, want %q", tt.description, result, tt.expectedValue)
			}
		})
	}
}

// TestParseValueNumberParsing validates correct number type detection
func TestParseValueNumberParsing(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldBeInt bool
		description string
	}{
		{
			name:        "integer without decimal",
			input:       "42",
			shouldBeInt: true,
			description: "Whole numbers should be parsed as int",
		},
		{
			name:        "number with decimal point",
			input:       "42.0",
			shouldBeInt: false,
			description: "Numbers with decimal point should be float even if whole",
		},
		{
			name:        "fractional number",
			input:       "3.14",
			shouldBeInt: false,
			description: "Fractional numbers should be float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)

			if tt.shouldBeInt {
				switch result.(type) {
				case int, int64:
					// Correct
				default:
					t.Errorf("%s: got type %T, want int/int64", tt.description, result)
				}
			} else {
				if _, ok := result.(float64); !ok {
					t.Errorf("%s: got type %T, want float64", tt.description, result)
				}
			}
		})
	}
}

// TestParseValueComplexExpressions validates that complex JavaScript expressions are returned as strings
func TestParseValueComplexExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "arrow function",
			input: "(x) => x * 2",
		},
		{
			name:  "object destructuring",
			input: "{ name, age }",
		},
		{
			name:  "spread operator",
			input: "...items",
		},
		{
			name:  "logical operators",
			input: "isActive && isVisible || isDefault",
		},
		{
			name:  "comparison operators",
			input: "count >= 10",
		},
		{
			name:  "typeof expression",
			input: "typeof value",
		},
		{
			name:  "new expression",
			input: "new Date()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)

			// Complex expressions should be returned as strings
			if str, ok := result.(string); ok {
				if str != tt.input && str != trimSpaces(tt.input) {
					t.Errorf("parseValue(%q) = %q, expected to preserve input as string", tt.input, str)
				}
			} else {
				t.Errorf("parseValue(%q) should return string for complex expression, got %T", tt.input, result)
			}
		})
	}
}

// Helper function for tests
func trimSpaces(s string) string {
	// Simple space trimming for comparison
	return s // Can be enhanced if needed
}
