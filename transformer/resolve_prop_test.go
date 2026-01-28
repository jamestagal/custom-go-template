package transformer

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestResolvePropValue tests the resolvePropValue helper function that resolves prop values
// by evaluating expressions against parent scope. This tests all three prop types:
// 1. Dynamic props (prop={expression})
// 2. Shorthand props ({prop})
// 3. Static props (prop="value")
//
// NOTE: After x-data optimization work, simple variable references now return __VAR_REF__ markers
// to preserve reactivity. The formatComponentData function strips these markers when generating x-data.
//
// IMPORTANT: The parser creates shorthand props with BOTH IsShorthand=true AND IsDynamic=true,
// and sets Value=propName. So they're handled as dynamic props that return __VAR_REF__ markers.
func TestResolvePropValue(t *testing.T) {
	// Create a realistic parent scope with various data types
	parentScope := map[string]any{
		"currentUser": map[string]any{
			"name": "John Doe",
			"role": "admin",
			"id":   123,
		},
		"products": []any{
			map[string]any{"id": 1, "name": "Widget", "price": 29.99},
			map[string]any{"id": 2, "name": "Gadget", "price": 49.99},
		},
		"formatPrice":     "function(p) { return '$' + p; }",
		"isLoggedIn":      true,
		"count":           42,
		"discount":        15.5,
		"title":           "Welcome",
		"user":            "Alice",
		"validationErrors": []any{"error1", "error2"},
		"items":           []any{1, 2, 3},
		"config": map[string]any{
			"debug": true,
			"port":  8080,
		},
	}

	tests := []struct {
		name        string
		prop        ast.ComponentProp
		parentScope map[string]any
		expected    any
		description string
	}{
		// ========== Dynamic Props (prop={expression}) ==========
		// UPDATED 2025-01-25: Simple variable references now return ACTUAL values for build-time expansion
		{
			name: "dynamic prop - simple variable lookup success",
			prop: ast.ComponentProp{
				Name:      "user",
				Value:     "{currentUser}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    parentScope["currentUser"], // Actual map value
			description: "Should return actual value for build-time expansion",
		},
		{
			name: "dynamic prop - boolean variable",
			prop: ast.ComponentProp{
				Name:      "isActive",
				Value:     "{isLoggedIn}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    true, // Actual bool value
			description: "Should return actual bool value for build-time expansion",
		},
		{
			name: "dynamic prop - integer variable",
			prop: ast.ComponentProp{
				Name:      "total",
				Value:     "{count}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    42, // Actual int value
			description: "Should return actual int value for build-time expansion",
		},
		{
			name: "dynamic prop - float variable",
			prop: ast.ComponentProp{
				Name:      "rate",
				Value:     "{discount}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    15.5, // Actual float value
			description: "Should return actual float value for build-time expansion",
		},
		{
			name: "dynamic prop - array variable",
			prop: ast.ComponentProp{
				Name:      "items",
				Value:     "{products}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    parentScope["products"], // Actual array value
			description: "Should return actual array value for build-time expansion",
		},
		{
			name: "dynamic prop - function reference",
			prop: ast.ComponentProp{
				Name:      "formatter",
				Value:     "{formatPrice}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "function(p) { return '$' + p; }", // Actual string value
			description: "Should return actual function string for build-time expansion",
		},
		{
			name: "dynamic prop - property access expression",
			prop: ast.ComponentProp{
				Name:      "userName",
				Value:     "{currentUser.name}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "John Doe", // Resolved nested property value
			description: "Should resolve nested property and return actual value",
		},
		{
			name: "dynamic prop - array index access",
			prop: ast.ComponentProp{
				Name:      "firstProduct",
				Value:     "{products[0]}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "products[0]",
			description: "Should return expression string for complex access",
		},
		{
			name: "dynamic prop - nested property access",
			prop: ast.ComponentProp{
				Name:      "debug",
				Value:     "{config.debug}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    true, // Resolved nested property value
			description: "Should resolve nested property and return actual value",
		},
		{
			name: "dynamic prop - method call expression",
			prop: ast.ComponentProp{
				Name:      "price",
				Value:     "{formatPrice(item.price)}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "formatPrice(item.price)",
			description: "Should return expression string for method calls",
		},
		{
			name: "dynamic prop - arithmetic expression",
			prop: ast.ComponentProp{
				Name:      "total",
				Value:     "{count + 10}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    float64(52), // count=42 + 10 = 52 (evaluated at build-time)
			description: "Should evaluate arithmetic to actual value when variables resolvable",
		},
		{
			name: "dynamic prop - comparison expression",
			prop: ast.ComponentProp{
				Name:      "isValid",
				Value:     "{count > 0}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "count > 0",
			description: "Should return expression string for comparisons",
		},
		{
			name: "dynamic prop - logical expression",
			prop: ast.ComponentProp{
				Name:      "show",
				Value:     "{isLoggedIn && count > 0}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "isLoggedIn && count > 0",
			description: "Should return expression string for logical operations",
		},
		{
			name: "dynamic prop - ternary expression",
			prop: ast.ComponentProp{
				Name:      "label",
				Value:     "{isLoggedIn ? 'Logout' : 'Login'}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "isLoggedIn ? 'Logout' : 'Login'",
			description: "Should return expression string for ternary",
		},
		{
			name: "dynamic prop - variable not found in parent scope",
			prop: ast.ComponentProp{
				Name:      "missing",
				Value:     "{nonExistentVar}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "__VAR_REF__nonExistentVar",
			description: "Should return __VAR_REF__ marker even when variable not found",
		},
		{
			name: "dynamic prop - template literal expression",
			prop: ast.ComponentProp{
				Name:      "message",
				Value:     "{`Hello ${user.name}`}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "`Hello ${user.name}`",
			description: "Should return template literal as expression string",
		},
		{
			name: "dynamic prop - with extra whitespace",
			prop: ast.ComponentProp{
				Name:      "data",
				Value:     "{  count  }",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    42, // Actual int value (whitespace is trimmed)
			description: "Should handle whitespace and return actual value",
		},
		{
			name: "dynamic prop - special case for validationErrors",
			prop: ast.ComponentProp{
				Name:      "errors",
				Value:     "{validationErrors}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    parentScope["validationErrors"], // Actual array value
			description: "Should return actual array value for validationErrors",
		},

		// ========== Shorthand Props ({prop}) ==========
		// UPDATED 2025-01-25: Shorthand props now return actual values for build-time expansion
		{
			name: "shorthand prop - variable found",
			prop: ast.ComponentProp{
				Name:        "user",
				Value:       "user", // Parser sets value to prop name
				IsShorthand: true,
				IsDynamic:   true, // Parser ALSO sets IsDynamic
			},
			parentScope: parentScope,
			expected:    "Alice", // Actual string value
			description: "Should return actual value for build-time expansion",
		},
		{
			name: "shorthand prop - boolean found",
			prop: ast.ComponentProp{
				Name:        "isLoggedIn",
				Value:       "isLoggedIn",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: parentScope,
			expected:    true, // Actual bool value
			description: "Should return actual bool value for build-time expansion",
		},
		{
			name: "shorthand prop - integer found",
			prop: ast.ComponentProp{
				Name:        "count",
				Value:       "count",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: parentScope,
			expected:    42, // Actual int value
			description: "Should return actual int value for build-time expansion",
		},
		{
			name: "shorthand prop - object found",
			prop: ast.ComponentProp{
				Name:        "currentUser",
				Value:       "currentUser",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: parentScope,
			expected:    parentScope["currentUser"], // Actual map value
			description: "Should return actual map value for build-time expansion",
		},
		{
			name: "shorthand prop - array found",
			prop: ast.ComponentProp{
				Name:        "items",
				Value:       "items",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: parentScope,
			expected:    parentScope["items"], // Actual array value
			description: "Should return actual array value for build-time expansion",
		},
		{
			name: "shorthand prop - variable not found",
			prop: ast.ComponentProp{
				Name:        "missingProp",
				Value:       "missingProp",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: parentScope,
			expected:    "__VAR_REF__missingProp", // Still returns marker when not found
			description: "Should return __VAR_REF__ marker when variable not in scope",
		},
		{
			name: "shorthand prop - title found",
			prop: ast.ComponentProp{
				Name:        "title",
				Value:       "title",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: parentScope,
			expected:    "Welcome", // Actual string value
			description: "Should return actual string value for build-time expansion",
		},

		// ========== Static Props (prop="value") ==========
		// Static props return parsed values (unchanged behavior)
		{
			name: "static prop - simple string",
			prop: ast.ComponentProp{
				Name:  "label",
				Value: `"Click Me"`,
			},
			parentScope: parentScope,
			expected:    `"Click Me"`,
			description: "Should return string with quotes (static props don't parse)",
		},
		{
			name: "static prop - single quoted string",
			prop: ast.ComponentProp{
				Name:  "text",
				Value: `'Hello'`,
			},
			parentScope: parentScope,
			expected:    `'Hello'`,
			description: "Should return string with quotes (static props don't parse)",
		},
		{
			name: "static prop - boolean true",
			prop: ast.ComponentProp{
				Name:  "isActive",
				Value: "true",
			},
			parentScope: parentScope,
			expected:    "true",
			description: "Should return boolean as string (static props)",
		},
		{
			name: "static prop - boolean false",
			prop: ast.ComponentProp{
				Name:  "isDisabled",
				Value: "false",
			},
			parentScope: parentScope,
			expected:    "false",
			description: "Should return boolean as string (static props)",
		},
		{
			name: "static prop - integer",
			prop: ast.ComponentProp{
				Name:  "count",
				Value: "42",
			},
			parentScope: parentScope,
			expected:    "42",
			description: "Should return integer as string (static props)",
		},
		{
			name: "static prop - negative integer",
			prop: ast.ComponentProp{
				Name:  "offset",
				Value: "-10",
			},
			parentScope: parentScope,
			expected:    "-10",
			description: "Should return negative integer as string (static props)",
		},
		{
			name: "static prop - float",
			prop: ast.ComponentProp{
				Name:  "price",
				Value: "29.99",
			},
			parentScope: parentScope,
			expected:    "29.99",
			description: "Should return float as string (static props)",
		},
		{
			name: "static prop - negative float",
			prop: ast.ComponentProp{
				Name:  "temperature",
				Value: "-5.5",
			},
			parentScope: parentScope,
			expected:    "-5.5",
			description: "Should return negative float as string (static props)",
		},
		{
			name: "static prop - zero",
			prop: ast.ComponentProp{
				Name:  "initial",
				Value: "0",
			},
			parentScope: parentScope,
			expected:    "0",
			description: "Should return zero as string (static props)",
		},
		{
			name: "static prop - null",
			prop: ast.ComponentProp{
				Name:  "data",
				Value: "null",
			},
			parentScope: parentScope,
			expected:    "null",
			description: "Should return null as string (static props)",
		},
		{
			name: "static prop - array",
			prop: ast.ComponentProp{
				Name:  "items",
				Value: "[1, 2, 3]",
			},
			parentScope: parentScope,
			expected:    "[1, 2, 3]",
			description: "Should return array as string for Alpine",
		},
		{
			name: "static prop - array with strings",
			prop: ast.ComponentProp{
				Name:  "colors",
				Value: `["red", "green", "blue"]`,
			},
			parentScope: parentScope,
			expected:    `["red", "green", "blue"]`,
			description: "Should return array as string for Alpine",
		},
		{
			name: "static prop - object",
			prop: ast.ComponentProp{
				Name:  "config",
				Value: `{ key: "value" }`,
			},
			parentScope: parentScope,
			expected:    `{ key: "value" }`,
			description: "Should return object as string for Alpine",
		},
		{
			name: "static prop - empty string",
			prop: ast.ComponentProp{
				Name:  "placeholder",
				Value: `""`,
			},
			parentScope: parentScope,
			expected:    `""`,
			description: "Should return empty string with quotes (static props)",
		},
		{
			name: "static prop - string with spaces",
			prop: ast.ComponentProp{
				Name:  "message",
				Value: `"Hello World"`,
			},
			parentScope: parentScope,
			expected:    `"Hello World"`,
			description: "Should preserve spaces in string (static props)",
		},
		{
			name: "static prop - string with escaped quotes",
			prop: ast.ComponentProp{
				Name:  "quote",
				Value: `"He said \"hello\""`,
			},
			parentScope: parentScope,
			expected:    `"He said \"hello\""`,
			description: "Should preserve escaped quotes (static props)",
		},
		{
			name: "static prop - scientific notation",
			prop: ast.ComponentProp{
				Name:  "large",
				Value: "1e5",
			},
			parentScope: parentScope,
			expected:    "1e5",
			description: "Should return scientific notation as string (static props)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePropValue(tt.prop, tt.parentScope)

			// Use DeepEqual to compare complex types
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("%s\nresolvePropValue() = %v (type: %T)\nwant %v (type: %T)",
					tt.description, result, result, tt.expected, tt.expected)
			}
		})
	}
}

// TestResolvePropValueEdgeCases tests edge cases and error scenarios
func TestResolvePropValueEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		prop        ast.ComponentProp
		parentScope map[string]any
		expected    any
		description string
	}{
		{
			name: "nil parent scope - dynamic prop",
			prop: ast.ComponentProp{
				Name:      "user",
				Value:     "{currentUser}",
				IsDynamic: true,
			},
			parentScope: nil,
			expected:    "__VAR_REF__currentUser",
			description: "Should return __VAR_REF__ marker when parent scope is nil",
		},
		{
			name: "empty parent scope - dynamic prop",
			prop: ast.ComponentProp{
				Name:      "count",
				Value:     "{total}",
				IsDynamic: true,
			},
			parentScope: map[string]any{},
			expected:    "__VAR_REF__total",
			description: "Should return __VAR_REF__ marker when parent scope is empty",
		},
		{
			name: "nil parent scope - shorthand prop",
			prop: ast.ComponentProp{
				Name:        "user",
				Value:       "user",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: nil,
			expected:    "__VAR_REF__user",
			description: "Should return __VAR_REF__ marker when parent scope is nil for shorthand",
		},
		{
			name: "empty parent scope - shorthand prop",
			prop: ast.ComponentProp{
				Name:        "count",
				Value:       "count",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: map[string]any{},
			expected:    "__VAR_REF__count",
			description: "Should return __VAR_REF__ marker when parent scope is empty for shorthand",
		},
		{
			name: "nil parent scope - static prop",
			prop: ast.ComponentProp{
				Name:  "label",
				Value: `"Static"`,
			},
			parentScope: nil,
			expected:    `"Static"`,
			description: "Static props should work without parent scope",
		},
		{
			name: "dynamic prop with braces already trimmed",
			prop: ast.ComponentProp{
				Name:      "value",
				Value:     "count",
				IsDynamic: true,
			},
			parentScope: map[string]any{"count": 10},
			expected:    10, // Actual value for build-time expansion
			description: "Should return actual value for build-time expansion",
		},
		{
			name: "dynamic prop - empty braces",
			prop: ast.ComponentProp{
				Name:      "value",
				Value:     "{}",
				IsDynamic: true,
			},
			parentScope: map[string]any{},
			expected:    "",
			description: "Should return empty string for empty braces",
		},
		{
			name: "static prop - unquoted expression",
			prop: ast.ComponentProp{
				Name:  "expr",
				Value: "someVariable",
			},
			parentScope: map[string]any{},
			expected:    "someVariable",
			description: "Unquoted value should be returned as expression string",
		},
		{
			name: "dynamic prop - complex chained expression",
			prop: ast.ComponentProp{
				Name:      "result",
				Value:     "{user.address.city.toUpperCase()}",
				IsDynamic: true,
			},
			parentScope: map[string]any{},
			expected:    "user.address.city.toUpperCase()",
			description: "Should return complex chained expression as string",
		},
		{
			name: "shorthand prop - prop name with underscore",
			prop: ast.ComponentProp{
				Name:        "user_name",
				Value:       "user_name",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: map[string]any{"user_name": "Bob"},
			expected:    "Bob", // Actual value for build-time expansion
			description: "Should return actual value for build-time expansion",
		},
		{
			name: "shorthand prop - prop name with dollar sign",
			prop: ast.ComponentProp{
				Name:        "$store",
				Value:       "$store",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope: map[string]any{"$store": map[string]any{"data": "value"}},
			expected:    map[string]any{"data": "value"}, // Actual value for build-time expansion
			description: "Should return actual value for build-time expansion",
		},
		{
			name: "static prop - whitespace value",
			prop: ast.ComponentProp{
				Name:  "space",
				Value: "   ",
			},
			parentScope: map[string]any{},
			expected:    "   ",
			description: "Whitespace value should be returned as-is (static props)",
		},
		{
			name: "parent scope with nil values",
			prop: ast.ComponentProp{
				Name:      "data",
				Value:     "{nullValue}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"nullValue": nil},
			expected:    "__VAR_REF__nullValue",
			description: "Should return __VAR_REF__ marker even when value is nil",
		},
		{
			name: "parent scope with zero values",
			prop: ast.ComponentProp{
				Name:      "count",
				Value:     "{zero}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"zero": 0},
			expected:    0, // Actual value for build-time expansion
			description: "Should return actual zero value for build-time expansion",
		},
		{
			name: "parent scope with false value",
			prop: ast.ComponentProp{
				Name:      "flag",
				Value:     "{disabled}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"disabled": false},
			expected:    false, // Actual value for build-time expansion
			description: "Should return actual false value for build-time expansion",
		},
		{
			name: "parent scope with empty string",
			prop: ast.ComponentProp{
				Name:      "text",
				Value:     "{emptyStr}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"emptyStr": ""},
			expected:    "", // Actual value for build-time expansion
			description: "Should return actual empty string for build-time expansion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePropValue(tt.prop, tt.parentScope)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("%s\nresolvePropValue() = %v (type: %T)\nwant %v (type: %T)",
					tt.description, result, result, tt.expected, tt.expected)
			}
		})
	}
}

// TestResolvePropValueLogging tests that appropriate warnings are logged
func TestResolvePropValueLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		// Restore default log output to os.Stderr (not nil!)
		log.SetOutput(os.Stderr)
	}()

	tests := []struct {
		name         string
		prop         ast.ComponentProp
		parentScope  map[string]any
		shouldLog    bool
		expectedLogSubstring string
		description  string
	}{
		{
			name: "shorthand prop not found - should log",
			prop: ast.ComponentProp{
				Name:        "missingVar",
				Value:       "missingVar",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope:  map[string]any{"otherVar": "value"},
			shouldLog:    true,
			expectedLogSubstring: "not in parent scope",
			description:  "Should log when shorthand prop variable not found",
		},
		{
			name: "dynamic prop not found - should log",
			prop: ast.ComponentProp{
				Name:      "missing",
				Value:     "{nonExistent}",
				IsDynamic: true,
			},
			parentScope:  map[string]any{"other": "value"},
			shouldLog:    true,
			expectedLogSubstring: "not in parent scope",
			description:  "Should log about variable not found",
		},
		{
			name: "shorthand prop found - should log resolved",
			prop: ast.ComponentProp{
				Name:        "user",
				Value:       "user",
				IsShorthand: true,
				IsDynamic:   true,
			},
			parentScope:  map[string]any{"user": "Alice"},
			shouldLog:    true,
			expectedLogSubstring: "Resolved variable",
			description:  "Should log that variable was resolved to actual value",
		},
		{
			name: "static prop - no logging",
			prop: ast.ComponentProp{
				Name:  "label",
				Value: `"Static"`,
			},
			parentScope:  map[string]any{},
			shouldLog:    false,
			expectedLogSubstring: "extractPropValue",
			description:  "Should not log for static props",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log buffer
			logBuffer.Reset()

			// Call function
			_ = resolvePropValue(tt.prop, tt.parentScope)

			// Check log output
			logOutput := logBuffer.String()

			if tt.shouldLog {
				if !strings.Contains(logOutput, tt.expectedLogSubstring) {
					t.Errorf("%s\nExpected log to contain: %q\nGot: %q",
						tt.description, tt.expectedLogSubstring, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedLogSubstring) {
					t.Errorf("%s\nExpected no log containing %q\nGot: %q",
						tt.description, tt.expectedLogSubstring, logOutput)
				}
			}
		})
	}
}

// TestResolvePropValueTypePreservation tests that actual values are returned for build-time expansion
// (Updated 2025-01-25: Changed from __VAR_REF__ markers to actual values for build-time loop expansion)
func TestResolvePropValueTypePreservation(t *testing.T) {
	parentScope := map[string]any{
		"strVal":   "text",
		"intVal":   42,
		"floatVal": 3.14,
		"boolVal":  true,
		"nilVal":   nil,
		"arrayVal": []any{1, 2, 3},
		"objVal":   map[string]any{"key": "value"},
	}

	tests := []struct {
		name         string
		prop         ast.ComponentProp
		expectedType string
		expectedValue any
		description  string
	}{
		{
			name: "preserve string type",
			prop: ast.ComponentProp{
				Name:      "str",
				Value:     "{strVal}",
				IsDynamic: true,
			},
			expectedType: "string",
			expectedValue: "text",
			description:  "Simple variable should return actual string value",
		},
		{
			name: "preserve int type",
			prop: ast.ComponentProp{
				Name:      "num",
				Value:     "{intVal}",
				IsDynamic: true,
			},
			expectedType: "int",
			expectedValue: 42,
			description:  "Simple variable should return actual int value",
		},
		{
			name: "preserve float type",
			prop: ast.ComponentProp{
				Name:      "flt",
				Value:     "{floatVal}",
				IsDynamic: true,
			},
			expectedType: "float64",
			expectedValue: 3.14,
			description:  "Simple variable should return actual float value",
		},
		{
			name: "preserve bool type",
			prop: ast.ComponentProp{
				Name:      "flag",
				Value:     "{boolVal}",
				IsDynamic: true,
			},
			expectedType: "bool",
			expectedValue: true,
			description:  "Simple variable should return actual bool value",
		},
		{
			name: "preserve nil type",
			prop: ast.ComponentProp{
				Name:      "null",
				Value:     "{nilVal}",
				IsDynamic: true,
			},
			expectedType: "string",
			expectedValue: "__VAR_REF__nilVal", // nil values return reference marker
			description:  "Nil variable should return __VAR_REF__ marker (not in scope effectively)",
		},
		{
			name: "preserve array type",
			prop: ast.ComponentProp{
				Name:      "arr",
				Value:     "{arrayVal}",
				IsDynamic: true,
			},
			expectedType: "[]interface {}",
			expectedValue: []any{1, 2, 3},
			description:  "Simple variable should return actual array value",
		},
		{
			name: "preserve object type",
			prop: ast.ComponentProp{
				Name:      "obj",
				Value:     "{objVal}",
				IsDynamic: true,
			},
			expectedType: "map[string]interface {}",
			expectedValue: map[string]any{"key": "value"},
			description:  "Simple variable should return actual object value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePropValue(tt.prop, parentScope)

			actualType := reflect.TypeOf(result).String()

			if actualType != tt.expectedType {
				t.Errorf("%s\nGot type: %s, want: %s\nValue: %v",
					tt.description, actualType, tt.expectedType, result)
			}

			// For string results, check exact match or prefix for __VAR_REF__
			if strResult, ok := result.(string); ok {
				if strExpected, ok := tt.expectedValue.(string); ok {
					if strResult != strExpected {
						t.Errorf("%s\nGot: %s, want: %s",
							tt.description, strResult, strExpected)
					}
				}
			}
		})
	}
}

// TestResolvePropValueComplexExpressions tests complex JavaScript expressions
func TestResolvePropValueComplexExpressions(t *testing.T) {
	parentScope := map[string]any{
		"items": []any{1, 2, 3},
		"user":  map[string]any{"name": "John"},
	}

	tests := []struct {
		name        string
		prop        ast.ComponentProp
		expected    string
		description string
	}{
		{
			name: "array method call",
			prop: ast.ComponentProp{
				Name:      "filtered",
				Value:     "{items.filter(x => x > 1)}",
				IsDynamic: true,
			},
			expected:    "items.filter(x => x > 1)",
			description: "Should return array method expression",
		},
		{
			name: "chained property access",
			prop: ast.ComponentProp{
				Name:      "userName",
				Value:     "{user.profile.name}",
				IsDynamic: true,
			},
			expected:    "user.profile.name",
			description: "Should return chained property expression",
		},
		{
			name: "destructuring syntax",
			prop: ast.ComponentProp{
				Name:      "spread",
				Value:     "{...items}",
				IsDynamic: true,
			},
			expected:    "...items",
			description: "Should return spread operator expression",
		},
		{
			name: "optional chaining",
			prop: ast.ComponentProp{
				Name:      "optional",
				Value:     "{user?.address?.city}",
				IsDynamic: true,
			},
			expected:    "user?.address?.city",
			description: "Should return optional chaining expression",
		},
		{
			name: "nullish coalescing",
			prop: ast.ComponentProp{
				Name:      "default",
				Value:     "{value ?? 'default'}",
				IsDynamic: true,
			},
			expected:    "value ?? 'default'",
			description: "Should return nullish coalescing expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePropValue(tt.prop, parentScope)

			if result != tt.expected {
				t.Errorf("%s\nresolvePropValue() = %q\nwant %q",
					tt.description, result, tt.expected)
			}
		})
	}
}
