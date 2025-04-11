package alpine

import (
	"testing"
	
	"github.com/jimafisk/custom_go_template/renderer"
)

func TestIsComplexJSObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "simple object",
			input: "{ a: 1, b: 2 }",
			want:  true, // Should be true because it has key-value pairs
		},
		{
			name:  "object with method",
			input: "{ method() { return 42 } }",
			want:  true,
		},
		{
			name:  "object with getter",
			input: "{ get prop() { return value } }",
			want:  true,
		},
		{
			name:  "object with setter",
			input: "{ set prop(value) { this.value = value } }",
			want:  true,
		},
		{
			name:  "object with arrow function",
			input: "{ method: () => { return 42 } }",
			want:  true,
		},
		{
			name:  "object with function property",
			input: "{ method: function() { return 42 } }",
			want:  true,
		},
		{
			name:  "object with alpine magic property",
			input: "{ $refs: {}, $el: null }",
			want:  true,
		},
		{
			name:  "object with init method",
			input: "{ init() { console.log('initialized') } }",
			want:  true,
		},
		{
			name:  "nested object with methods",
			input: "{ user: { profile: { update() { return true } } } }",
			want:  true,
		},
		{
			name:  "object with this reference",
			input: "{ count: 0, increment() { this.count++ } }",
			want:  true,
		},
		{
			name:  "not an object",
			input: "42",
			want:  false,
		},
		{
			name:  "empty object",
			input: "{}",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderer.IsComplexJSObject(tt.input)
			if got != tt.want {
				t.Errorf("IsComplexJSObject() = %v, want %v for input: %s", got, tt.want, tt.input)
			}
		})
	}
}

func TestCleanupObjectLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "add missing comma",
			input: "{ a: 1 b: 2 }",
			want:  "{ a: 1, b: 2 }",
		},
		{
			name:  "remove trailing comma",
			input: "{ a: 1, b: 2, }",
			want:  "{ a: 1, b: 2 }",
		},
		{
			name:  "fix method definition",
			input: "{ method() { return 42 } }",
			want:  "{ method() { return 42 } }",
		},
		{
			name:  "fix nested object",
			input: "{ user: { name: 'John' age: 30 } }",
			want:  "{ user: { name: 'John', age: 30 } }",
		},
		{
			name:  "preserve complex method",
			input: "{ toggle() { this.open = !this.open } }",
			want:  "{ toggle() { this.open = !this.open } }",
		},
		{
			name:  "fix multiple issues",
			input: "{ a: 1 b: 2, c: 3, }",
			want:  "{ a: 1, b: 2, c: 3 }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderer.CleanupObjectLiteral(tt.input)
			if got != tt.want {
				t.Errorf("CleanupObjectLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanupMethodDefinition(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "regular method",
			input: "method() { return 42 }",
			want:  "method() { return 42 }",
		},
		{
			name:  "async method",
			input: "async fetch() { return await api.get() }",
			want:  "async fetch() { return await api.get() }",
		},
		{
			name:  "getter",
			input: "get value() { return this._value }",
			want:  "get value() { return this._value }",
		},
		{
			name:  "setter",
			input: "set value(v) { this._value = v }",
			want:  "set value(v) { this._value = v }",
		},
		{
			name:  "function expression",
			input: "function() { return 42 }",
			want:  "function() { return 42 }",
		},
		{
			name:  "arrow function",
			input: "() => { return 42 }",
			want:  "() => { return 42 }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderer.CleanupMethodDefinition(tt.input)
			if got != tt.want {
				t.Errorf("CleanupMethodDefinition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatJSValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "string value",
			input: "hello",
			want:  "'hello'",
		},
		{
			name:  "integer value",
			input: 42,
			want:  "42",
		},
		{
			name:  "boolean value",
			input: true,
			want:  "true",
		},
		{
			name:  "nil value",
			input: nil,
			want:  "null",
		},
		{
			name:  "float value",
			input: 3.14,
			want:  "3.14",
		},
		{
			name:  "slice value",
			input: []any{1, "two", true},
			want:  "[1, 'two', true]",
		},
		{
			name:  "map value",
			input: map[string]any{"name": "John", "age": 30},
			want:  "{name: 'John', age: 30}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderer.FormatJSValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatJSValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
