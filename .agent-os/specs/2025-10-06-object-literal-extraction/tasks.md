# Implementation Tasks: Object Literal Extraction

**Spec**: Object Literal Extraction for Fence Parser
**Estimated Total**: 6-7 hours
**Agent**: go-backend

## Task 1: Create JavaScript Formatting Functions

**File**: Create `renderer/js_format.go`

**Cognitive Load**: 5/30
**Estimated Time**: 1.5 hours

### Implementation

Create pure formatting functions for JavaScript literals:

```go
package renderer

import (
	"bytes"
	"fmt"
	"strings"
)

// FormatValueForXData formats a Go value as a JavaScript literal
// suitable for use in Alpine.js x-data attributes
func FormatValueForXData(value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		return formatObjectLiteral(v)
	case []interface{}:
		return formatArrayLiteral(v)
	case string:
		return fmt.Sprintf("'%s'", escapeJSString(v))
	case bool:
		return fmt.Sprintf("%t", v)
	case float64:
		return fmt.Sprintf("%v", v)
	case int:
		return fmt.Sprintf("%d", v)
	case nil:
		return "null"
	default:
		// Fallback: convert to string
		return fmt.Sprintf("'%v'", v)
	}
}

// formatObjectLiteral formats a map as a JavaScript object literal
func formatObjectLiteral(obj map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("{ ")

	first := true
	for key, value := range obj {
		if !first {
			buf.WriteString(", ")
		}
		first = false

		// Write key
		buf.WriteString(key)
		buf.WriteString(": ")

		// Write value (recursive)
		buf.WriteString(FormatValueForXData(value))
	}

	buf.WriteString(" }")
	return buf.String()
}

// formatArrayLiteral formats a slice as a JavaScript array literal
func formatArrayLiteral(arr []interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("[")

	for i, value := range arr {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(FormatValueForXData(value))
	}

	buf.WriteString("]")
	return buf.String()
}

// escapeJSString escapes special characters for JavaScript string literals
// Uses single quotes for strings to avoid conflicts with HTML attributes
func escapeJSString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)   // Backslash first
	s = strings.ReplaceAll(s, `'`, `\'`)   // Single quote
	s = strings.ReplaceAll(s, "\n", `\n`)  // Newline
	s = strings.ReplaceAll(s, "\r", `\r`)  // Carriage return
	s = strings.ReplaceAll(s, "\t", `\t`)  // Tab
	return s
}
```

### Tests

Create `renderer/js_format_test.go`:

```go
package renderer

import "testing"

func TestFormatValueForXData_String(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"it's", "'it\\'s'"},
		{"line1\nline2", "'line1\\nline2'"},
		{`path\to\file`, "'path\\\\to\\\\file'"},
	}

	for _, tt := range tests {
		result := FormatValueForXData(tt.input)
		if result != tt.expected {
			t.Errorf("FormatValueForXData(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatValueForXData_Primitives(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{nil, "null"},
	}

	for _, tt := range tests {
		result := FormatValueForXData(tt.input)
		if result != tt.expected {
			t.Errorf("FormatValueForXData(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatValueForXData_Object(t *testing.T) {
	obj := map[string]interface{}{
		"name":  "Jane",
		"age":   30,
		"active": true,
	}

	result := FormatValueForXData(obj)

	// Check it contains all fields (order may vary for maps)
	if !contains(result, "name: 'Jane'") {
		t.Errorf("Expected name field, got: %s", result)
	}
	if !contains(result, "age: 30") {
		t.Errorf("Expected age field, got: %s", result)
	}
	if !contains(result, "active: true") {
		t.Errorf("Expected active field, got: %s", result)
	}
}

func TestFormatValueForXData_NestedObject(t *testing.T) {
	obj := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Jane",
			"email": "jane@example.com",
		},
	}

	result := FormatValueForXData(obj)

	if !contains(result, "user: {") {
		t.Errorf("Expected nested object, got: %s", result)
	}
	if !contains(result, "name: 'Jane'") {
		t.Errorf("Expected nested name, got: %s", result)
	}
}

func TestFormatValueForXData_Array(t *testing.T) {
	arr := []interface{}{"one", "two", "three"}

	result := FormatValueForXData(arr)
	expected := "['one', 'two', 'three']"

	if result != expected {
		t.Errorf("FormatValueForXData(array) = %q, want %q", result, expected)
	}
}

func TestFormatValueForXData_ArrayOfObjects(t *testing.T) {
	arr := []interface{}{
		map[string]interface{}{"name": "Jane"},
		map[string]interface{}{"name": "John"},
	}

	result := FormatValueForXData(arr)

	if !contains(result, "name: 'Jane'") || !contains(result, "name: 'John'") {
		t.Errorf("Expected array of objects, got: %s", result)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
```

### Acceptance Criteria
- [ ] All formatting functions implemented
- [ ] All tests pass
- [ ] String escaping handles special characters correctly
- [ ] Nested objects format correctly
- [ ] Arrays format correctly

---

## Task 2: Update Page-Level x-data Formatting

**File**: Modify `cmd/server/main.go` (or wherever root x-data is rendered)

**Cognitive Load**: 4/30
**Estimated Time**: 1 hour

### Implementation

Find where page props are marshaled to JSON for the `<body x-data='...'>` attribute and replace with JavaScript literal formatting.

**Current pattern** (likely):
```go
propsJSON, err := json.Marshal(props)
xDataValue := string(propsJSON)
```

**New pattern**:
```go
import "path/to/renderer"

// Build x-data as JavaScript object literal
var buf bytes.Buffer
buf.WriteString("{")
first := true
for key, value := range props {
	if !first {
		buf.WriteString(", ")
	}
	first = false

	buf.WriteString(key)
	buf.WriteString(": ")
	buf.WriteString(renderer.FormatValueForXData(value))
}
buf.WriteString("}")
xDataValue := buf.String()
```

### Test

Manual verification:
1. Define object in home.html fence:
   ```javascript
   let user1 = { name: "Test", age: 25 }
   ```

2. Check rendered `<body x-data='...'>` contains:
   ```javascript
   user1: { name: 'Test', age: 25 }
   ```
   NOT:
   ```javascript
   user1: "{\n  name: \"Test\",\n  age: 25\n}"
   ```

### Acceptance Criteria
- [ ] Page-level props use JavaScript literal formatting
- [ ] Objects appear as objects, not strings
- [ ] Existing simple props still work
- [ ] Browser console shows no errors

---

## Task 3: Update Component x-data Formatting

**File**: Modify `transformer/components.go:154-206`

**Cognitive Load**: 6/30
**Estimated Time**: 1.5 hours

### Implementation

The component transformer already has custom formatting logic for functions and getters. Extend it to handle objects and arrays.

**Current code** (lines 154-206) handles:
- Functions: `function name() { }` → `name() { }`
- Getters: `get name() { }` → stays as-is
- Everything else: wrapped in quotes

**Add**:
```go
import "path/to/renderer"

// In the loop that formats component props (around line 180-205)
for propName, propValue := range componentData {
	// ... existing code ...

	// After function/getter checks, before quoting as string:

	// NEW: Check if value is an object or array
	switch v := propValue.(type) {
	case map[string]interface{}:
		// Format as JavaScript object literal
		result.WriteString(propName)
		result.WriteString(": ")
		result.WriteString(renderer.FormatValueForXData(v))
		continue
	case []interface{}:
		// Format as JavaScript array literal
		result.WriteString(propName)
		result.WriteString(": ")
		result.WriteString(renderer.FormatValueForXData(v))
		continue
	}

	// ... existing code for other types ...
}
```

### Test

Create integration test `tests/components/object_props_test.go`:

```go
package components

import (
	"testing"
	"strings"
)

func TestObjectPropPassing(t *testing.T) {
	// Test that object props pass correctly to components

	// Setup: Create test page with object variable
	// and component that uses it

	// Render and verify x-data contains object literal,
	// not string representation

	// This test will be similar to existing component tests
	// but specifically checks object formatting
}
```

### Acceptance Criteria
- [ ] Objects in component props format as JavaScript literals
- [ ] Arrays in component props format correctly
- [ ] Existing function/getter formatting still works
- [ ] Integration test passes

---

## Task 4: End-to-End Verification with UserProfile

**File**: Update `examples/pages/home.html` test

**Cognitive Load**: 3/30
**Estimated Time**: 1 hour

### Implementation

1. Restore the user object definitions in home.html fence section (already there)
2. Update component calls to pass objects:
   ```html
   <="./components/UserProfile.html" user={user1} showRole={true} />
   <='{path}' user={user2} showRole={true} />
   <="./components/{comp}.html" user={user3} showRole={true} />
   ```

3. Start server and verify in browser:
   - No JavaScript errors
   - 3 different UserProfile cards show different names
   - Role badges show different colors
   - Emails show different values

### Visual Verification

Browser should show:

**Card 1**:
- Name: Benjamin
- Email: benjamin@example.com
- Role: admin (red badge)
- Member since: 2024-10-06

**Card 2**:
- Name: Dynamic User
- Email: dynamic@example.com
- Role: user (green badge)
- Member since: 2024-05-15

**Card 3**:
- Name: Variable Path User
- Email: variable@example.com
- Role: editor (blue badge)
- Member since: 2024-03-20

### Acceptance Criteria
- [ ] All 3 cards display different data
- [ ] No console errors
- [ ] Objects are accessible in Alpine.js (can inspect in devtools)
- [ ] Screenshots captured showing success

---

## Task 5: Documentation and Cleanup

**Files**: Update documentation

**Cognitive Load**: 2/30
**Estimated Time**: 30 minutes

### Implementation

1. Update `docs/SESSION_SUMMARY_2025-10-06.md`:
   - Add section for object literal extraction implementation
   - Document the fix and its importance for Plenti

2. Update `.agent-os/specs/2025-10-06-object-literal-extraction/`:
   - Create `COMPLETION_SUMMARY.md` documenting success
   - List all files changed
   - Include before/after examples

3. Update `CLAUDE.md` if needed:
   - Add note about object prop passing support
   - Update fence section documentation

### Acceptance Criteria
- [ ] Session summary updated
- [ ] Completion summary created
- [ ] Documentation reflects new capability

---

## Success Metrics

At completion:
1. ✅ Objects extracted from fence as structured data, not strings
2. ✅ Objects pass correctly as component props
3. ✅ UserProfile example shows 3 different users with different data
4. ✅ All tests pass (unit + integration)
5. ✅ Zero JavaScript errors in browser
6. ✅ Code documented and spec marked complete

## Notes for go-backend Agent

- Use modern Go 1.21+ features
- Follow existing code patterns in renderer/fence.go
- Maintain consistency with transformer/components.go formatting style
- Add comprehensive error handling
- Include logging for debugging object formatting
- Test edge cases: null, nested, mixed types
- Verify backward compatibility with existing simple props
