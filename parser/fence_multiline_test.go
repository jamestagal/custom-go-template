package parser

import (
	"testing"
)

func TestParseFenceContent_MultiLineArray(t *testing.T) {
	input := `prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "links" {
		t.Errorf("Expected prop name 'links', got '%s'", fence.Props[0].Name)
	}

	expectedValue := `[
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]`

	if fence.Props[0].DefaultValue != expectedValue {
		t.Errorf("Expected full array value, got:\n%s\n\nExpected:\n%s",
			fence.Props[0].DefaultValue, expectedValue)
	}
}

func TestParseFenceContent_MultiLineObject(t *testing.T) {
	input := `prop config = {
  theme: "dark",
  options: {
    nested: true
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "config" {
		t.Errorf("Expected prop name 'config', got '%s'", fence.Props[0].Name)
	}

	expectedValue := `{
  theme: "dark",
  options: {
    nested: true
  }
}`

	if fence.Props[0].DefaultValue != expectedValue {
		t.Errorf("Expected full object value, got:\n%s\n\nExpected:\n%s",
			fence.Props[0].DefaultValue, expectedValue)
	}
}

func TestParseFenceContent_FooterLinks(t *testing.T) {
	// Real-world test case from Footer.html
	input := `prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" },
  { label: "Products", url: "/products" },
  { label: "Contact", url: "/contact" },
  { label: "Terms of Service", url: "/terms" },
  { label: "Privacy Policy", url: "/privacy" }
]`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "links" {
		t.Errorf("Expected prop name 'links', got '%s'", fence.Props[0].Name)
	}

	// Check that the value is the complete array, not just "["
	if fence.Props[0].DefaultValue == "[" {
		t.Error("Bug detected: prop value was truncated to first line only")
	}

	// The value should contain all the links
	value := fence.Props[0].DefaultValue
	if !contains(value, "Home") || !contains(value, "Privacy Policy") {
		t.Errorf("Multi-line array incomplete: %s", value)
	}
}

func TestParseFenceContent_MixedProps(t *testing.T) {
	input := `prop companyName = "Custom Template Co."
prop year = new Date().getFullYear()
prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]
prop count = 42`

	fence := parseFenceContent(input)

	if len(fence.Props) != 4 {
		t.Errorf("Expected 4 props, got %d", len(fence.Props))
		return
	}

	// Check each prop
	tests := []struct {
		name  string
		value string
	}{
		{"companyName", `"Custom Template Co."`},
		{"year", "new Date().getFullYear()"},
		{"links", "["},  // Should start with [
		{"count", "42"},
	}

	for i, tt := range tests {
		if i >= len(fence.Props) {
			t.Errorf("Missing prop at index %d", i)
			continue
		}

		if fence.Props[i].Name != tt.name {
			t.Errorf("Prop %d: expected name '%s', got '%s'", i, tt.name, fence.Props[i].Name)
		}

		// For the links prop, verify it's multi-line
		if tt.name == "links" {
			if fence.Props[i].DefaultValue == "[" {
				t.Errorf("Bug: links prop was truncated to single line")
			}
			if !contains(fence.Props[i].DefaultValue, "Home") {
				t.Errorf("links prop incomplete: %s", fence.Props[i].DefaultValue)
			}
		} else {
			// For single-line props, check exact value
			if fence.Props[i].DefaultValue != tt.value {
				t.Errorf("Prop %s: expected value '%s', got '%s'",
					tt.name, tt.value, fence.Props[i].DefaultValue)
			}
		}
	}
}

func TestParseFenceContent_FunctionExpression(t *testing.T) {
	input := `prop year = new Date().getFullYear()`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	expected := "new Date().getFullYear()"
	if fence.Props[0].DefaultValue != expected {
		t.Errorf("Expected function expression '%s', got '%s'",
			expected, fence.Props[0].DefaultValue)
	}
}

func TestParseFenceContent_NestedArraysAndObjects(t *testing.T) {
	input := `prop data = {
  items: [
    { id: 1, tags: ["a", "b"] },
    { id: 2, tags: ["c", "d"] }
  ],
  meta: { count: 2 }
}`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	value := fence.Props[0].DefaultValue

	// Verify it contains all the nested content
	requiredStrings := []string{"items", "tags", "meta", "count"}
	for _, str := range requiredStrings {
		if !contains(value, str) {
			t.Errorf("Multi-line prop missing '%s': %s", str, value)
		}
	}
}

func TestParseFenceContent_Variables(t *testing.T) {
	input := `let items = [
  { id: 1 },
  { id: 2 }
]
const config = {
  enabled: true
}`

	fence := parseFenceContent(input)

	if len(fence.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(fence.Variables))
		return
	}

	// Check items variable
	if fence.Variables[0].Name != "items" {
		t.Errorf("Expected variable name 'items', got '%s'", fence.Variables[0].Name)
	}
	if fence.Variables[0].Keyword != "let" {
		t.Errorf("Expected keyword 'let', got '%s'", fence.Variables[0].Keyword)
	}
	if !contains(fence.Variables[0].Value, "id") {
		t.Errorf("Variable value incomplete: %s", fence.Variables[0].Value)
	}

	// Check config variable
	if fence.Variables[1].Name != "config" {
		t.Errorf("Expected variable name 'config', got '%s'", fence.Variables[1].Name)
	}
	if fence.Variables[1].Keyword != "const" {
		t.Errorf("Expected keyword 'const', got '%s'", fence.Variables[1].Keyword)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
		(s[0:len(substr)] == substr ||
		 s[len(s)-len(substr):] == substr ||
		 containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
