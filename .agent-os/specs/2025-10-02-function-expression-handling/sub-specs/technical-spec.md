# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-02-function-expression-handling/spec.md

## Technical Requirements

### 1. Enhanced isFunctionExpression()

**Location**: `transformer/alpine.go`

**Current Signature**: `func isFunctionExpression(expr string) bool`

**Purpose**: Detect if a string contains a JavaScript function definition

**Improved Implementation**:

```go
func isFunctionExpression(expr string) bool {
    expr = strings.TrimSpace(expr)

    // Pattern 1: function declarations - function name() {}
    if strings.HasPrefix(expr, "function") {
        return true
    }

    // Pattern 2: arrow functions - () => {} or (x) => {} or x => {}
    if strings.Contains(expr, "=>") {
        return true
    }

    // Pattern 3: async functions - async function name() {}
    if strings.HasPrefix(expr, "async") {
        return true
    }

    // Pattern 4: getters/setters - get name() {} or set name(v) {}
    if strings.HasPrefix(expr, "get ") || strings.HasPrefix(expr, "set ") {
        return true
    }

    // Pattern 5: method shorthand - name() { ... }
    // Look for pattern: identifier followed by ( with balanced braces
    methodShorthandRegex := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*\s*\([^)]*\)\s*\{`)
    if methodShorthandRegex.MatchString(expr) {
        return true
    }

    return false
}
```

**Test Cases**:
- `function greet() { return 'hello'; }` → true
- `function() { return 42; }` → true
- `() => { return 42; }` → true
- `(x) => x * 2` → true
- `x => x * 2` → true
- `async function fetch() {}` → true
- `get value() { return this._value; }` → true
- `set value(v) { this._value = v; }` → true
- `greet() { return 'hello'; }` → true
- `const greet = function() {}` → false (handled differently)
- `"hello"` → false
- `42` → false
- `true` → false

### 2. Refactored formatGoValueToJS()

**Location**: `transformer/alpine.go`

**Current Signature**: `func formatGoValueToJS(value any) string`

**Purpose**: Convert Go values to JavaScript literal syntax

**Implementation Requirements**:

```go
func formatGoValueToJS(value any) string {
    if value == nil {
        return "null"
    }

    switch v := value.(type) {
    case string:
        // Check if this string is a function definition
        if isFunctionExpression(v) {
            // Return function without quotes
            return v
        }

        // Check if it looks like a complex JS object/array literal
        trimmed := strings.TrimSpace(v)
        if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
           (strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
            // Could be an object or array literal - preserve as-is
            return v
        }

        // Check if it's a variable reference (simple identifier)
        if isValidIdentifier(v) {
            // Variable reference - no quotes
            return v
        }

        // Regular string - add quotes and escape
        escaped := strings.ReplaceAll(v, `\`, `\\`)
        escaped = strings.ReplaceAll(escaped, `'`, `\'`)
        return fmt.Sprintf("'%s'", escaped)

    case bool:
        if v {
            return "true"
        }
        return "false"

    case int, int8, int16, int32, int64:
        return fmt.Sprintf("%d", v)

    case uint, uint8, uint16, uint32, uint64:
        return fmt.Sprintf("%d", v)

    case float32, float64:
        return fmt.Sprintf("%v", v)

    case []interface{}:
        // Array - format elements
        var elements []string
        for _, item := range v {
            elements = append(elements, formatGoValueToJS(item))
        }
        return "[" + strings.Join(elements, ", ") + "]"

    case map[string]interface{}:
        // Object - format key-value pairs
        var pairs []string
        for key, val := range v {
            formattedValue := formatGoValueToJS(val)
            pairs = append(pairs, fmt.Sprintf("'%s': %s", key, formattedValue))
        }
        return "{" + strings.Join(pairs, ", ") + "}"

    default:
        // Fallback - convert to string and quote
        str := fmt.Sprintf("%v", v)
        escaped := strings.ReplaceAll(str, `'`, `\'`)
        return fmt.Sprintf("'%s'", escaped)
    }
}
```

### 3. Helper: isValidIdentifier()

**Location**: `transformer/alpine.go`

**Purpose**: Check if a string is a valid JavaScript identifier (variable name)

**Implementation**:

```go
func isValidIdentifier(s string) bool {
    if len(s) == 0 {
        return false
    }

    // Check for JavaScript keywords
    keywords := map[string]bool{
        "break": true, "case": true, "catch": true, "class": true, "const": true,
        "continue": true, "debugger": true, "default": true, "delete": true,
        "do": true, "else": true, "export": true, "extends": true, "finally": true,
        "for": true, "function": true, "if": true, "import": true, "in": true,
        "instanceof": true, "let": true, "new": true, "return": true, "super": true,
        "switch": true, "this": true, "throw": true, "try": true, "typeof": true,
        "var": true, "void": true, "while": true, "with": true, "yield": true,
        "true": true, "false": true, "null": true, "undefined": true,
    }

    if keywords[s] {
        return false
    }

    // First character must be letter, underscore, or dollar sign
    firstChar := s[0]
    if !((firstChar >= 'a' && firstChar <= 'z') ||
         (firstChar >= 'A' && firstChar <= 'Z') ||
         firstChar == '_' || firstChar == '$') {
        return false
    }

    // Subsequent characters can also be digits
    for i := 1; i < len(s); i++ {
        c := s[i]
        if !((c >= 'a' && c <= 'z') ||
             (c >= 'A' && c <= 'Z') ||
             (c >= '0' && c <= '9') ||
             c == '_' || c == '$') {
            return false
        }
    }

    return true
}
```

### 4. Updated alpineDataFormatter()

**Location**: `transformer/alpine.go`

**Current Issues**:
- Uses `json.Marshal()` which quotes everything
- Test-specific hardcoded values

**Required Changes**:

```go
func alpineDataFormatter(dataScope map[string]any) string {
    // Remove test-specific code (lines 24-44)

    // Ensure critical variables exist (can keep this logic)
    ensureCriticalVariables(dataScope)

    // Sort keys for consistent output
    keys := make([]string, 0, len(dataScope))
    for key := range dataScope {
        keys = append(keys, key)
    }
    sort.Strings(keys)

    // Build object literal using formatGoValueToJS
    var parts []string
    for _, key := range keys {
        // Skip internal Alpine.js variables
        if strings.HasPrefix(key, "$") {
            continue
        }

        value := dataScope[key]

        // Skip nil values (optional - or format as undefined)
        if value == nil {
            parts = append(parts, fmt.Sprintf(`"%s": undefined`, key))
            continue
        }

        // Use the fixed helper to format the value
        formattedValue := formatGoValueToJS(value)

        // Build key-value pair
        parts = append(parts, fmt.Sprintf(`"%s": %s`, key, formattedValue))
    }

    result := "{" + strings.Join(parts, ", ") + "}"
    log.Printf("Generated x-data object literal: %s", result)
    return result
}
```

### 5. Test Case Validation

**Test File**: `tests/alpine/alpine_data_wrapper_test.go`

**Failing Test**: `TestAlpineDataWrapper/Function_Expressions`

**Expected Output**:
```html
<div x-data="{&quot;count&quot;:0,&quot;increment&quot;:function() { return count++ }}">
```

**Current Wrong Output**:
```html
<div x-data="{&quot;count&quot;:0,&quot;increment&quot;:&quot;function() { return count++ }&quot;}">
```

**Key Difference**: The function should NOT be quoted

**Verification After Fix**:
```bash
go test ./tests/alpine -v -run TestAlpineDataWrapper/Function_Expressions
```

### 6. Edge Cases to Handle

**Complex Objects**:
- Objects with methods: `{ count: 0, increment() { this.count++ } }`
- Nested objects: `{ user: { name: 'John', greet() { return 'Hi'; } } }`
- Arrays with functions: `[() => {}, () => {}]`

**Alpine.js Magic Properties**:
- `$refs`, `$el`, `$dispatch`, etc. - These can be in the data scope
- Format as regular properties (they're just references)

**Method Shorthand Detection**:
```javascript
{
  count: 0,
  increment() {  // This is method shorthand
    return count++
  }
}
```

The entire value string might be: `"function() { return count++ }"` or `"increment() { return count++ }"`

### 7. Preserving Complex JS Objects

**Existing Function**: `isComplexJSObject()`

**Purpose**: Detect objects that should be preserved as-is

**Integration**: If a string value passes `isComplexJSObject()`, preserve it without modification in `formatGoValueToJS()`

### 8. HTML Entity Escaping

**Note**: The `&quot;` in test expectations is HTML entity escaping by the renderer

**Transformer Output**: `{"count":0,"increment":function() { return count++ }}`

**After Rendering to HTML Attribute**: `{&quot;count&quot;:0,&quot;increment&quot;:function() { return count++ }}`

**Important**: The transformer should output raw quotes. The renderer will handle HTML entity escaping.

### 9. Logging and Debugging

**Add Debug Logs**:
```go
log.Printf("Formatting value '%v' (type: %T)", value, value)
log.Printf("  Is function expression: %v", isFunctionExpression(stringValue))
log.Printf("  Formatted as: %s", formattedValue)
```

### 10. Backward Compatibility

**Ensure**:
- All existing `TestAlpineDataWrapper` subtests still pass
- `TestIsComplexJSObject` tests still pass
- No regression in other Alpine.js integration tests

## External Dependencies

No new external dependencies required. Uses only Go standard library (`strings`, `fmt`, `regexp`, `sort`).
