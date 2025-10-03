# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-03-component-props-rendering-fix/spec.md

> Created: 2025-10-03
> Version: 1.0.0

## Technical Requirements

### 1. Data Flow Investigation

**Trace prop values through the pipeline:**
- Component AST creation (fence parser) → Already working correctly
- Component transformation (`transformer/components.go`) → Investigate how props are extracted and transformed
- Component rendering (`renderer/component.go`) → Investigate how props are passed to Alpine.js
- Data scope building (`cmd/server/main.go` parseValue()) → **Primary issue location**

**Key files to analyze:**
- `transformer/components.go` - `TransformComponent()` function
- `renderer/component.go` - Component rendering logic
- `cmd/server/main.go` - `parseValue()` function that processes prop values
- `transformer/scope.go` - Data scope management

### 2. Fix parseValue() Function

**Current behavior:**
```go
// In cmd/server/main.go
func parseValue(value string) interface{} {
    // Attempts json.Unmarshal on value
    // Fails for JavaScript syntax (unquoted keys)
    // Returns truncated string
}
```

**Required changes:**
- Handle multi-line JavaScript array/object syntax
- Preserve JavaScript syntax as-is (don't attempt to parse as JSON)
- Return complete string value for Alpine.js to interpret
- Consider: Should we convert JavaScript syntax to valid JSON, or pass raw JavaScript?

**Decision needed:** Alpine.js x-data accepts JavaScript expressions, not JSON. We should preserve the JavaScript syntax as-is.

### 3. Component Data Scope Building

**Current issue:**
When building the x-data object for a component with props, the prop values are being marshaled to JSON. This fails for JavaScript syntax and truncates the value.

**Required fix:**
- In `transformer/components.go`, when building component data scope
- Preserve prop values as JavaScript expressions
- Ensure the renderer outputs these as-is in the x-data attribute
- Handle string escaping properly for attribute context

### 4. Testing Requirements

**Test cases to add:**
1. Component with array prop containing multiple objects
2. Component with nested object prop
3. Component with multi-line array spanning 10+ lines
4. Component with prop containing string values with special characters
5. Integration test: Full page render with Footer component

**Test files:**
- `tests/alpine/component_props_test.go` - Add multi-line prop tests
- `transformer/components_test.go` - Add transformation tests

### 5. String Escaping

**Consideration:**
When outputting JavaScript values in HTML attributes, proper escaping is critical:
- Double quotes in prop values must be escaped
- Newlines must be preserved or converted to spaces
- Single quotes vs double quotes in attribute context

## Approach

### Phase 1: Investigation (30 min)
1. Add debug logging to trace a sample prop value through the pipeline
2. Identify exact location where truncation occurs
3. Document current data flow with specific function calls

### Phase 2: Fix parseValue() (1 hour)
1. Update `parseValue()` in `cmd/server/main.go` to detect JavaScript array/object syntax
2. Return raw JavaScript string instead of attempting JSON parse
3. Add unit tests for parseValue() with multi-line inputs

### Phase 3: Fix Component Transformation (1-2 hours)
1. Update `transformer/components.go` to preserve JavaScript prop values
2. Ensure prop values are not marshaled to JSON
3. Update data scope building to handle JavaScript expressions
4. Add tests for component transformation with complex props

### Phase 4: Fix Rendering (1 hour)
1. Update `renderer/component.go` if needed to output JavaScript expressions correctly
2. Ensure proper HTML attribute escaping
3. Test rendered HTML in browser with Alpine.js

### Phase 5: Integration Testing (1 hour)
1. Test Footer component with full links array
2. Verify in browser that Alpine.js receives and processes the data correctly
3. Add integration tests to prevent regression

## External Dependencies

None. This is an internal fix to the transformation and rendering pipeline.

All required code changes are within existing packages:
- `cmd/server/main.go`
- `transformer/components.go`
- `renderer/component.go`
- Test files in `tests/alpine/`
