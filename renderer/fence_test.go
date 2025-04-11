package renderer

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvalJS_SimpleExpressions(t *testing.T) {
	tests := []struct {
		name      string
		jsCode    string
		propsDecl string
		want      any
	}{
		{
			name:      "simple string",
			jsCode:    "'hello world'",
			propsDecl: "",
			want:      "hello world",
		},
		{
			name:      "simple number",
			jsCode:    "42",
			propsDecl: "",
			want:      int64(42),
		},
		{
			name:      "simple boolean",
			jsCode:    "true",
			propsDecl: "",
			want:      true,
		},
		{
			name:      "simple array",
			jsCode:    "[1, 2, 3]",
			propsDecl: "",
			want:      []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			name:      "simple object",
			jsCode:    "({a: 1, b: 2})",
			propsDecl: "",
			want:      map[string]interface{}{"a": int64(1), "b": int64(2)},
		},
		{
			name:      "variable from props",
			jsCode:    "x + y",
			propsDecl: "let x = 10; let y = 20;",
			want:      int64(30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvalJS(tt.jsCode, tt.propsDecl)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvalJS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalJS_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name      string
		jsCode    string
		propsDecl string
		wantType  string // Instead of exact value, we check the type for complex objects
	}{
		{
			name:      "object with method",
			jsCode:    "({ method() { return 42; } })",
			propsDecl: "",
			wantType:  "string", // Should return the object literal as a string for Alpine.js
		},
		{
			name:      "alpine magic properties",
			jsCode:    "{ $refs: {}, $el: null }",
			propsDecl: "",
			wantType:  "string", // Should return the object literal as a string
		},
		{
			name:      "arrow function",
			jsCode:    "() => { return 42; }",
			propsDecl: "",
			wantType:  "string", // Should return the function as a string
		},
		{
			name:      "getter method",
			jsCode:    "{ get prop() { return 42; } }",
			propsDecl: "",
			wantType:  "string", // Should return the object as a string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvalJS(tt.jsCode, tt.propsDecl)
			gotType := reflect.TypeOf(got).String()
			if gotType != tt.wantType {
				t.Errorf("EvalJS() type = %v, want type %v", gotType, tt.wantType)
			}
		})
	}
}

func TestEvalJS_AlpineJSPatterns(t *testing.T) {
	tests := []struct {
		name      string
		jsCode    string
		propsDecl string
		shouldContain string // Check if result contains expected substring
	}{
		{
			name:         "x-data object",
			jsCode:       "{ count: 0, increment() { this.count++ } }",
			propsDecl:    "",
			shouldContain: "increment",
		},
		{
			name:         "complex nested object",
			jsCode:       "{ user: { name: 'John', profile: { age: 30 } } }",
			propsDecl:    "",
			shouldContain: "profile",
		},
		{
			name:         "alpine refs",
			jsCode:       "{ init() { this.$refs.button.focus() } }",
			propsDecl:    "",
			shouldContain: "$refs",
		},
		{
			name:         "alpine events",
			jsCode:       "{ handleClick(event) { console.log(event) } }",
			propsDecl:    "",
			shouldContain: "handleClick",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvalJS(tt.jsCode, tt.propsDecl)
			gotStr, ok := got.(string)
			if !ok {
				t.Errorf("EvalJS() returned type %T, want string", got)
				return
			}
			
			if !contains(gotStr, tt.shouldContain) {
				t.Errorf("EvalJS() = %v, should contain %v", gotStr, tt.shouldContain)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestEvalJS_EvaluationStrategies(t *testing.T) {
	// Test each evaluation strategy separately
	t.Run("DirectEvaluation", func(t *testing.T) {
		result := EvalJS("1 + 2", "")
		if result != int64(3) {
			t.Errorf("Direct evaluation failed, got %v, want 3", result)
		}
	})

	t.Run("ExpressionWrapping", func(t *testing.T) {
		result := EvalJS("{a: 1, b: 2}", "")
		if _, ok := result.(string); !ok {
			t.Errorf("Expression wrapping failed, got type %T, want string", result)
		}
	})

	t.Run("FunctionWrapping", func(t *testing.T) {
		result := EvalJS("function() { return 42; }", "")
		if _, ok := result.(string); !ok {
			t.Errorf("Function wrapping failed, got type %T, want string", result)
		}
	})

	t.Run("ObjectAssignment", func(t *testing.T) {
		result := EvalJS("{ method() { return 42; } }", "")
		if _, ok := result.(string); !ok {
			t.Errorf("Object assignment failed, got type %T, want string", result)
		}
	})
}
