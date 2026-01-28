# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-06-object-literal-extraction/spec.md

> Created: 2025-10-06
> Version: 1.0.0

## Technical Requirements

### 1. JavaScript Literal Formatting Functions

Create new file `renderer/js_literals.go` with the following functions:

#### formatValueForXData(value interface{}) string
**Purpose**: Top-level dispatcher that routes any value type to the appropriate formatting function.

**Logic**:
- Check value type using type assertion
- Route to specialized formatter:
  - `nil` → `"null"`
  - `bool` → `"true"` or `"false"`
  - `int`, `int64`, `float64`, etc. → `fmt.Sprintf("%v", value)`
  - `string` → `escapeString(value)` wrapped in quotes
  - `[]interface{}` → `formatArrayLiteral(value)`
  - `map[string]interface{}` → `formatObjectLiteral(value)`
  - Unknown types → JSON marshal as fallback (defensive)

#### formatObjectLiteral(obj map[string]interface{}) string
**Purpose**: Format a map as a JavaScript object literal.

**Logic**:
```go
// Pseudocode
result := "{"
keys := sortedKeys(obj)  // Sort for deterministic output
for each key in keys:
    formattedValue := formatValueForXData(obj[key])
    result += key + ": " + formattedValue + ", "
result = trimTrailingComma(result)
result += "}"
return result
```

**Edge Cases**:
- Empty objects → `"{}"`
- Nested objects → Recursive call to formatValueForXData
- Keys with special characters → Quote key if needed (future enhancement)

#### formatArrayLiteral(arr []interface{}) string
**Purpose**: Format a slice as a JavaScript array literal.

**Logic**:
```go
// Pseudocode
result := "["
for each element in arr:
    formattedValue := formatValueForXData(element)
    result += formattedValue + ", "
result = trimTrailingComma(result)
result += "]"
return result
```

**Edge Cases**:
- Empty arrays → `"[]"`
- Nested arrays → Recursive call to formatValueForXData
- Mixed type arrays → Each element formatted independently

#### escapeString(str string) string
**Purpose**: Escape special characters in string values for JavaScript literals.

**Logic**:
- Replace `\` with `\\`
- Replace `"` with `\"`
- Replace newlines with `\n`
- Replace tabs with `\t`
- Replace carriage returns with `\r`
- Preserve Unicode characters (UTF-8 safe)

**Return**: String WITHOUT surrounding quotes (caller adds quotes)

### 2. Integration Points

#### A. Page-Level x-data Generation (renderer/render.go)

**Current Code**:
```go
// In buildXDataFromScope()
jsonData, err := json.Marshal(value)
if err != nil {
    return "", fmt.Errorf("failed to marshal data: %w", err)
}
dataMap[key] = string(jsonData)
```

**Updated Code**:
```go
// In buildXDataFromScope()
formattedValue := formatValueForXData(value)
dataMap[key] = formattedValue
```

**Impact**:
- Page-level variables become JavaScript literals
- No JSON escaping artifacts
- Nested objects in fence sections work correctly

#### B. Component x-data Generation (renderer/component.go)

**Current Code** (in `renderComponent` function):
```go
// When building component's x-data
jsonValue, err := json.Marshal(propValue)
if err != nil {
    return "", fmt.Errorf("failed to marshal prop: %w", err)
}
componentData[propName] = string(jsonValue)
```

**Updated Code**:
```go
// When building component's x-data
formattedValue := formatValueForXData(propValue)
componentData[propName] = formattedValue
```

**Impact**:
- Component props receive properly formatted objects
- Nested component data works correctly
- Arrays in component props maintain structure

### 3. Type Preservation Strategy

| Go Type | JavaScript Output | Example |
|---------|------------------|---------|
| `nil` | `null` | `null` |
| `bool` (true) | `true` | `true` |
| `bool` (false) | `false` | `false` |
| `int`, `int64` | Number | `42` |
| `float64` | Number | `3.14` |
| `string` | Quoted string | `"Hello"` |
| `[]interface{}` | Array | `[1, 2, 3]` |
| `map[string]interface{}` | Object | `{name: "Alice", age: 30}` |

### 4. Escaping Strategy

**String Escaping Requirements**:
- Backslash: `\` → `\\`
- Double quote: `"` → `\"`
- Newline: `\n` → `\\n`
- Tab: `\t` → `\\t`
- Carriage return: `\r` → `\\r`

**Why**: JavaScript literals in x-data attributes are eventually parsed by Alpine.js. Improper escaping causes syntax errors or data corruption.

**Example**:
```
Input: She said "hello"
Output: "She said \"hello\""
```

### 5. Backward Compatibility

**Requirement**: Existing simple props must continue working.

**Test Cases**:
- String prop: `name="Alice"` → x-data contains `name: "Alice"`
- Number prop: `count={42}` → x-data contains `count: 42`
- Boolean prop: `active={true}` → x-data contains `active: true`

**Verification**: All existing component tests must pass without modification.

## Approach

### Implementation Steps

1. **Create `renderer/js_literals.go`**
   - Implement `formatValueForXData` with type routing
   - Implement `formatObjectLiteral` with sorted keys
   - Implement `formatArrayLiteral` with recursive formatting
   - Implement `escapeString` with all special characters

2. **Create `renderer/js_literals_test.go`**
   - Unit tests for each formatting function
   - Test cases: primitives, nested objects, arrays, mixed types
   - Edge cases: empty structures, special characters, nil values

3. **Update `renderer/render.go`**
   - Replace JSON marshal in `buildXDataFromScope()`
   - Add import for js_literals functions

4. **Update `renderer/component.go`**
   - Replace JSON marshal in component prop handling
   - Add import for js_literals functions

5. **Create Integration Test**
   - New file: `tests/components/object_props_test.go`
   - Test nested objects passing to components
   - Test arrays passing to components
   - Test multiple component instances with different data

6. **Update Examples**
   - Modify `examples/pages/home.html` with 3 UserProfile instances
   - Ensure UserProfile component can display nested data

7. **Run Full Test Suite**
   - Verify all existing tests pass
   - Verify new tests pass
   - Manual verification in dev server

### Key Design Decisions

**Decision 1**: Use sorted keys for object formatting
**Rationale**: Deterministic output makes testing easier and diffs cleaner

**Decision 2**: Recursive value formatting
**Rationale**: Handles arbitrary nesting depth without special cases

**Decision 3**: Fallback to JSON marshal for unknown types
**Rationale**: Defensive programming; graceful degradation for edge cases

**Decision 4**: No quotes around object keys (unless needed)
**Rationale**: JavaScript allows unquoted keys for valid identifiers; cleaner output

**Decision 5**: Create separate file for literal formatting
**Rationale**: Single responsibility; easier to test and maintain

## External Dependencies

### Current Dependencies (No Changes)
- **Goja**: JavaScript engine for executing fence section code
  - Used to evaluate JavaScript and extract values
  - No modifications needed

- **Alpine.js**: Frontend reactive framework
  - Consumes x-data attributes
  - Expects JavaScript object literals
  - No modifications needed

### Go Standard Library
- `fmt` - String formatting
- `sort` - Sorting object keys
- `strings` - String manipulation
- `reflect` - Type introspection (if needed for type assertions)

### Testing Dependencies
- `testing` - Go testing framework
- Existing test helpers in `tests/` directory

## Testing Strategy

### Unit Tests (renderer/js_literals_test.go)

**Test Coverage**:
- `TestFormatValueForXData_Primitives`: nil, bool, int, float, string
- `TestFormatValueForXData_Strings`: special characters, quotes, newlines
- `TestFormatObjectLiteral`: simple objects, nested objects, empty objects
- `TestFormatArrayLiteral`: simple arrays, nested arrays, empty arrays, mixed types
- `TestEscapeString`: all special characters, unicode, empty string

### Integration Tests (tests/components/object_props_test.go)

**Test Coverage**:
- Component receives object with nested properties
- Component receives array prop
- Multiple components with different object props
- Object with mixed types (strings, numbers, booleans, arrays)

### Regression Tests

**Requirement**: All existing tests must pass
- `tests/alpine/alpine_integration_test.go`
- `tests/components/components_test.go`
- `transformer/*_test.go`
- `parser/*_test.go`

### Manual Verification

**Steps**:
1. Start dev server: `go run cmd/server/main.go`
2. Visit http://localhost:3000
3. Verify 3 UserProfile components display different data
4. Inspect rendered HTML for proper x-data syntax
5. Open browser console; verify no Alpine.js errors

## Performance Considerations

**Expected Impact**: Minimal
- Literal formatting is simpler than JSON marshaling
- No reflection needed for most types
- String building is efficient for small-to-medium objects

**Mitigation**:
- Use `strings.Builder` for large objects (if needed)
- Cache formatted values (future optimization if needed)

## Error Handling

**Strategy**: Graceful degradation

**Error Cases**:
1. Unknown type → Fallback to JSON marshal, log warning
2. Circular references → Not possible with current Goja extraction (returns primitives/maps/slices)
3. Invalid string characters → Escape aggressively

**No Breaking Changes**: If formatting fails, system should still produce valid (though possibly suboptimal) JavaScript.
