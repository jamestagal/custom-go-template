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
		{
			name: "dynamic prop - simple variable lookup success",
			prop: ast.ComponentProp{
				Name:      "user",
				Value:     "{currentUser}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected: map[string]any{
				"name": "John Doe",
				"role": "admin",
				"id":   123,
			},
			description: "Should return actual value from parent scope",
		},
		{
			name: "dynamic prop - boolean variable",
			prop: ast.ComponentProp{
				Name:      "isActive",
				Value:     "{isLoggedIn}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    true,
			description: "Should preserve boolean type",
		},
		{
			name: "dynamic prop - integer variable",
			prop: ast.ComponentProp{
				Name:      "total",
				Value:     "{count}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    42,
			description: "Should preserve integer type",
		},
		{
			name: "dynamic prop - float variable",
			prop: ast.ComponentProp{
				Name:      "rate",
				Value:     "{discount}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    15.5,
			description: "Should preserve float type",
		},
		{
			name: "dynamic prop - array variable",
			prop: ast.ComponentProp{
				Name:      "items",
				Value:     "{products}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected: []any{
				map[string]any{"id": 1, "name": "Widget", "price": 29.99},
				map[string]any{"id": 2, "name": "Gadget", "price": 49.99},
			},
			description: "Should preserve array structure",
		},
		{
			name: "dynamic prop - function reference",
			prop: ast.ComponentProp{
				Name:      "formatter",
				Value:     "{formatPrice}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "function(p) { return '$' + p; }",
			description: "Should pass function as string",
		},
		{
			name: "dynamic prop - property access expression",
			prop: ast.ComponentProp{
				Name:      "userName",
				Value:     "{currentUser.name}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "currentUser.name",
			description: "Should return expression string for Alpine to evaluate",
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
			expected:    "config.debug",
			description: "Should return expression for nested property access",
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
			expected:    "count + 10",
			description: "Should return expression string for arithmetic",
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
			expected:    "nonExistentVar",
			description: "Should return expression string when variable not found",
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
			expected:    42,
			description: "Should handle whitespace and lookup variable",
		},
		{
			name: "dynamic prop - special case for validationErrors",
			prop: ast.ComponentProp{
				Name:      "errors",
				Value:     "{validationErrors}",
				IsDynamic: true,
			},
			parentScope: parentScope,
			expected:    "validationErrors",
			description: "Special case should return expression string for errors prop",
		},

		// ========== Shorthand Props ({prop}) ==========
		{
			name: "shorthand prop - variable found",
			prop: ast.ComponentProp{
				Name:        "user",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected:    "Alice",
			description: "Should lookup variable by prop name",
		},
		{
			name: "shorthand prop - boolean found",
			prop: ast.ComponentProp{
				Name:        "isLoggedIn",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected:    true,
			description: "Should preserve boolean type in shorthand",
		},
		{
			name: "shorthand prop - integer found",
			prop: ast.ComponentProp{
				Name:        "count",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected:    42,
			description: "Should preserve integer type in shorthand",
		},
		{
			name: "shorthand prop - object found",
			prop: ast.ComponentProp{
				Name:        "currentUser",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected: map[string]any{
				"name": "John Doe",
				"role": "admin",
				"id":   123,
			},
			description: "Should pass object reference in shorthand",
		},
		{
			name: "shorthand prop - array found",
			prop: ast.ComponentProp{
				Name:        "items",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected:    []any{1, 2, 3},
			description: "Should pass array reference in shorthand",
		},
		{
			name: "shorthand prop - variable not found",
			prop: ast.ComponentProp{
				Name:        "missingProp",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected:    nil,
			description: "Should return nil when variable not found",
		},
		{
			name: "shorthand prop - title found",
			prop: ast.ComponentProp{
				Name:        "title",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: parentScope,
			expected:    "Welcome",
			description: "Should lookup title by prop name",
		},

		// ========== Static Props (prop="value") ==========
		{
			name: "static prop - simple string",
			prop: ast.ComponentProp{
				Name:  "label",
				Value: `"Click Me"`,
			},
			parentScope: parentScope,
			expected:    "Click Me",
			description: "Should parse quoted string",
		},
		{
			name: "static prop - single quoted string",
			prop: ast.ComponentProp{
				Name:  "text",
				Value: `'Hello'`,
			},
			parentScope: parentScope,
			expected:    "Hello",
			description: "Should parse single quoted string",
		},
		{
			name: "static prop - boolean true",
			prop: ast.ComponentProp{
				Name:  "isActive",
				Value: "true",
			},
			parentScope: parentScope,
			expected:    true,
			description: "Should parse boolean true",
		},
		{
			name: "static prop - boolean false",
			prop: ast.ComponentProp{
				Name:  "isDisabled",
				Value: "false",
			},
			parentScope: parentScope,
			expected:    false,
			description: "Should parse boolean false",
		},
		{
			name: "static prop - integer",
			prop: ast.ComponentProp{
				Name:  "count",
				Value: "42",
			},
			parentScope: parentScope,
			expected:    42,
			description: "Should parse integer",
		},
		{
			name: "static prop - negative integer",
			prop: ast.ComponentProp{
				Name:  "offset",
				Value: "-10",
			},
			parentScope: parentScope,
			expected:    -10,
			description: "Should parse negative integer",
		},
		{
			name: "static prop - float",
			prop: ast.ComponentProp{
				Name:  "price",
				Value: "29.99",
			},
			parentScope: parentScope,
			expected:    29.99,
			description: "Should parse float",
		},
		{
			name: "static prop - negative float",
			prop: ast.ComponentProp{
				Name:  "temperature",
				Value: "-5.5",
			},
			parentScope: parentScope,
			expected:    -5.5,
			description: "Should parse negative float",
		},
		{
			name: "static prop - zero",
			prop: ast.ComponentProp{
				Name:  "initial",
				Value: "0",
			},
			parentScope: parentScope,
			expected:    0,
			description: "Should parse zero",
		},
		{
			name: "static prop - null",
			prop: ast.ComponentProp{
				Name:  "data",
				Value: "null",
			},
			parentScope: parentScope,
			expected:    nil,
			description: "Should parse null",
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
			expected:    "",
			description: "Should parse empty string",
		},
		{
			name: "static prop - string with spaces",
			prop: ast.ComponentProp{
				Name:  "message",
				Value: `"Hello World"`,
			},
			parentScope: parentScope,
			expected:    "Hello World",
			description: "Should preserve spaces in string",
		},
		{
			name: "static prop - string with escaped quotes",
			prop: ast.ComponentProp{
				Name:  "quote",
				Value: `"He said \"hello\""`,
			},
			parentScope: parentScope,
			expected:    `He said "hello"`,
			description: "Should handle escaped quotes",
		},
		{
			name: "static prop - scientific notation",
			prop: ast.ComponentProp{
				Name:  "large",
				Value: "1e5",
			},
			parentScope: parentScope,
			expected:    1e5,
			description: "Should parse scientific notation",
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
			expected:    "currentUser",
			description: "Should return expression string when parent scope is nil",
		},
		{
			name: "empty parent scope - dynamic prop",
			prop: ast.ComponentProp{
				Name:      "count",
				Value:     "{total}",
				IsDynamic: true,
			},
			parentScope: map[string]any{},
			expected:    "total",
			description: "Should return expression string when parent scope is empty",
		},
		{
			name: "nil parent scope - shorthand prop",
			prop: ast.ComponentProp{
				Name:        "user",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: nil,
			expected:    nil,
			description: "Should return nil when parent scope is nil for shorthand",
		},
		{
			name: "empty parent scope - shorthand prop",
			prop: ast.ComponentProp{
				Name:        "count",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: map[string]any{},
			expected:    nil,
			description: "Should return nil when variable not in empty scope",
		},
		{
			name: "nil parent scope - static prop",
			prop: ast.ComponentProp{
				Name:  "label",
				Value: `"Static"`,
			},
			parentScope: nil,
			expected:    "Static",
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
			expected:    10,
			description: "Should handle value without braces",
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
				Value:       "",
				IsShorthand: true,
			},
			parentScope: map[string]any{"user_name": "Bob"},
			expected:    "Bob",
			description: "Should handle prop names with underscores",
		},
		{
			name: "shorthand prop - prop name with dollar sign",
			prop: ast.ComponentProp{
				Name:        "$store",
				Value:       "",
				IsShorthand: true,
			},
			parentScope: map[string]any{"$store": map[string]any{"data": "value"}},
			expected:    map[string]any{"data": "value"},
			description: "Should handle prop names with dollar signs",
		},
		{
			name: "static prop - whitespace value",
			prop: ast.ComponentProp{
				Name:  "space",
				Value: "   ",
			},
			parentScope: map[string]any{},
			expected:    "",
			description: "Whitespace-only value should return empty string",
		},
		{
			name: "parent scope with nil values",
			prop: ast.ComponentProp{
				Name:      "data",
				Value:     "{nullValue}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"nullValue": nil},
			expected:    nil,
			description: "Should return nil when parent scope has nil value",
		},
		{
			name: "parent scope with zero values",
			prop: ast.ComponentProp{
				Name:      "count",
				Value:     "{zero}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"zero": 0},
			expected:    0,
			description: "Should return zero correctly (not treat as falsy)",
		},
		{
			name: "parent scope with false value",
			prop: ast.ComponentProp{
				Name:      "flag",
				Value:     "{disabled}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"disabled": false},
			expected:    false,
			description: "Should return false correctly (not treat as missing)",
		},
		{
			name: "parent scope with empty string",
			prop: ast.ComponentProp{
				Name:      "text",
				Value:     "{emptyStr}",
				IsDynamic: true,
			},
			parentScope: map[string]any{"emptyStr": ""},
			expected:    "",
			description: "Should return empty string from parent scope",
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
		shouldLogWarning bool
		expectedLogSubstring string
		description  string
	}{
		{
			name: "shorthand prop not found - should log warning",
			prop: ast.ComponentProp{
				Name:        "missingVar",
				Value:       "",
				IsShorthand: true,
			},
			parentScope:  map[string]any{"otherVar": "value"},
			shouldLogWarning: true,
			expectedLogSubstring: "Warning: Shorthand prop 'missingVar' not found",
			description:  "Should log warning when shorthand prop not found",
		},
		{
			name: "dynamic prop not found - should log",
			prop: ast.ComponentProp{
				Name:      "missing",
				Value:     "{nonExistent}",
				IsDynamic: true,
			},
			parentScope:  map[string]any{"other": "value"},
			shouldLogWarning: true,
			expectedLogSubstring: "dynamic expression",
			description:  "Should log info about dynamic expression",
		},
		{
			name: "shorthand prop found - no warning",
			prop: ast.ComponentProp{
				Name:        "user",
				Value:       "",
				IsShorthand: true,
			},
			parentScope:  map[string]any{"user": "Alice"},
			shouldLogWarning: false,
			expectedLogSubstring: "Warning",
			description:  "Should not log warning when prop found",
		},
		{
			name: "static prop - no warning",
			prop: ast.ComponentProp{
				Name:  "label",
				Value: `"Static"`,
			},
			parentScope:  map[string]any{},
			shouldLogWarning: false,
			expectedLogSubstring: "Warning",
			description:  "Should not log warning for static props",
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

			if tt.shouldLogWarning {
				if !strings.Contains(logOutput, tt.expectedLogSubstring) {
					t.Errorf("%s\nExpected log to contain: %q\nGot: %q",
						tt.description, tt.expectedLogSubstring, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedLogSubstring) {
					t.Errorf("%s\nExpected no warning in log\nGot: %q",
						tt.description, logOutput)
				}
			}
		})
	}
}

// TestResolvePropValueTypePreservation tests that types are correctly preserved
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
			description:  "String should remain string",
		},
		{
			name: "preserve int type",
			prop: ast.ComponentProp{
				Name:      "num",
				Value:     "{intVal}",
				IsDynamic: true,
			},
			expectedType: "int",
			description:  "Int should remain int",
		},
		{
			name: "preserve float type",
			prop: ast.ComponentProp{
				Name:      "flt",
				Value:     "{floatVal}",
				IsDynamic: true,
			},
			expectedType: "float64",
			description:  "Float should remain float64",
		},
		{
			name: "preserve bool type",
			prop: ast.ComponentProp{
				Name:      "flag",
				Value:     "{boolVal}",
				IsDynamic: true,
			},
			expectedType: "bool",
			description:  "Bool should remain bool",
		},
		{
			name: "preserve nil type",
			prop: ast.ComponentProp{
				Name:      "null",
				Value:     "{nilVal}",
				IsDynamic: true,
			},
			expectedType: "nil",
			description:  "Nil should remain nil",
		},
		{
			name: "preserve array type",
			prop: ast.ComponentProp{
				Name:      "arr",
				Value:     "{arrayVal}",
				IsDynamic: true,
			},
			expectedType: "[]interface {}",
			description:  "Array should remain array",
		},
		{
			name: "preserve object type",
			prop: ast.ComponentProp{
				Name:      "obj",
				Value:     "{objVal}",
				IsDynamic: true,
			},
			expectedType: "map[string]interface {}",
			description:  "Object should remain map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePropValue(tt.prop, parentScope)

			var actualType string
			if result == nil {
				actualType = "nil"
			} else {
				actualType = reflect.TypeOf(result).String()
			}

			if actualType != tt.expectedType {
				t.Errorf("%s\nGot type: %s, want: %s\nValue: %v",
					tt.description, actualType, tt.expectedType, result)
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
