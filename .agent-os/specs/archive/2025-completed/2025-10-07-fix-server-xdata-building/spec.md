# Spec: Fix Server x-data Building Architecture

**Date**: 2025-10-07
**Status**: Draft
**Priority**: High (Bug #1 - Blocks function support)
**Complexity**: Low (Architectural fix, not feature addition)

## Problem Statement

The development server (`cmd/server/main.go`) currently bypasses the transformer and manually builds Alpine.js `x-data` attributes using regex-based extraction. This architectural violation causes:

1. **Functions extracted as truncated JSON strings** - Regex pattern `{[^}]*}` stops at the first `}`, breaking multi-line functions and nested braces
2. **Transformer bypassed** - Server doesn't use the proper rendering pipeline (parser → AST → transformer → renderer)
3. **alpineDataFormatter unused** - The transformer's function for proper x-data formatting (including method shorthand) is ignored
4. **Testing blocked** - Cannot test functions in templates (formatters, computed values, event handlers)

### Current Broken Pattern

```go
// cmd/server/main.go (current implementation - BROKEN)
functionRegex := regexp.MustCompile(`function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{[^}]*}`)
matches := functionRegex.FindAllStringSubmatch(fence.RawContent, -1)
for _, match := range matches {
    if len(match) >= 2 {
        props[match[1]] = match[0]  // Adds as STRING, not function!
    }
}

// Manual x-data building
xDataContent := "{"
for key, value := range props {
    xDataContent += fmt.Sprintf("%s: %v,", key, value)
}
xDataContent += "}"
```

**Result**: `x-data="{formatPrice: "function formatPrice(price) { retur..."}"` (truncated, quoted string)

### Expected Correct Pattern

The transformer's `alpineDataFormatter` already handles this correctly:

```go
// transformer/alpine.go (already implemented correctly)
func alpineDataFormatter(data map[string]interface{}) string {
    var parts []string
    for key, value := range data {
        switch v := value.(type) {
        case string:
            if strings.HasPrefix(v, "function ") {
                // Convert to method shorthand
                funcBody := strings.TrimPrefix(v, "function "+key)
                parts = append(parts, fmt.Sprintf("%s%s", key, funcBody))
            } else {
                parts = append(parts, fmt.Sprintf("%s: %s", key, v))
            }
        }
    }
    return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
```

**Result**: `x-data="{formatPrice(price) { return '$' + price.toFixed(2); }}"` (correct method shorthand)

## Goals

1. **Remove manual x-data building** - Delete all regex extraction and manual JSON building from server route handlers
2. **Use renderer.Render()** - Leverage the proper transformation pipeline for all template rendering
3. **Restore function support** - Add functions back to `examples/pages/comprehensive-simple.html` to verify fix
4. **Maintain performance** - Ensure no regression in render speed
5. **Preserve existing features** - All current functionality must continue working

## Non-Goals

- Fixing other bugs (Bug #2: Component props in loops, Bug #3: Attribute expressions, Bug #4: Multi-line variables)
- Adding new template features
- Changing the transformer's alpineDataFormatter implementation (already correct)
- Modifying the parser or AST structure

## Success Criteria

### Functional Requirements

1. ✅ **Functions render correctly** - `formatPrice(price)` in fence section becomes proper Alpine.js method
2. ✅ **Multi-line functions supported** - Functions with nested braces, template literals work
3. ✅ **Zero console errors** - No JavaScript errors when using functions
4. ✅ **All existing features work** - Props, variables, conditionals, loops, components unchanged

### Technical Requirements

1. ✅ **Server uses renderer.Render()** - No manual x-data building in route handlers
2. ✅ **alpineDataFormatter invoked** - Transformer's formatter is used for all x-data generation
3. ✅ **Proper scope tracking** - All variables/functions tracked in transformer's scope map
4. ✅ **Type safety maintained** - No interface{} type assertions removed

### Testing Requirements

1. ✅ **Add functions to comprehensive-simple.html** - Restore `getGreeting()` and `formatPrice()` functions
2. ✅ **Integration test passes** - Server renders functions correctly at http://localhost:3333/comprehensive-simple
3. ✅ **Unit tests pass** - All existing transformer and renderer tests continue passing
4. ✅ **Manual verification** - Browser console shows zero errors, functions execute correctly

## Architecture

### Current Architecture (Broken)

```
Template File
    ↓
Server reads file
    ↓
Parser extracts fence section (✓ correct)
    ↓
❌ Server uses REGEX to extract functions (BROKEN)
    ↓
❌ Server manually builds x-data as JSON string (BROKEN)
    ↓
Server injects x-data into HTML (bypassing transformer)
    ↓
Browser receives invalid x-data
```

### Target Architecture (Correct)

```
Template File
    ↓
renderer.Render(template, componentRegistry)
    ↓
Parser → AST (fence section + body)
    ↓
Transformer:
  - Extract fence data (props, variables, functions)
  - Transform AST nodes to Alpine.js directives
  - Build scope map from all discovered variables
  - ✓ Use alpineDataFormatter for x-data generation
    ↓
Renderer generates final HTML with correct x-data
    ↓
Browser receives valid x-data with method shorthand
```

## Implementation Plan

### Phase 1: Update Server Route Handlers

**File**: `cmd/server/main.go`

**Changes**:
1. Remove manual fence data extraction (regex patterns for functions, variables)
2. Remove manual x-data building logic
3. Call `renderer.Render()` for all template rendering
4. Pass component registry to renderer
5. Ensure proper error handling

**Estimated LOC**: ~100 lines removed, ~20 lines added

### Phase 2: Verify Transformer Integration

**File**: `transformer/alpine.go`

**Verification** (no changes needed):
1. Confirm `alpineDataFormatter` is called during transformation
2. Verify method shorthand conversion works for functions
3. Ensure all fence data types (props, vars, functions) are handled

**Estimated LOC**: 0 (verification only)

### Phase 3: Restore Functions to Test File

**File**: `examples/pages/comprehensive-simple.html`

**Changes**:
1. Add `getGreeting()` function to fence section
2. Add `formatPrice()` function to fence section
3. Use functions in template body (test expressions)

**Estimated LOC**: ~30 lines added

### Phase 4: Integration Testing

**Actions**:
1. Run server: `go run cmd/server/main.go`
2. Visit http://localhost:3333/comprehensive-simple
3. Check browser console (expect zero errors)
4. Verify functions execute correctly in template
5. Run existing test suite: `go test ./... -v`

**Expected Result**: All tests pass, zero console errors, functions work

## Technical Specifications

See `sub-specs/technical-spec.md` for detailed implementation guide.

## Risks and Mitigations

### Risk 1: Breaking Existing Templates
**Likelihood**: Low
**Impact**: High
**Mitigation**: Run full test suite after changes, manually test all example pages

### Risk 2: Performance Regression
**Likelihood**: Very Low
**Impact**: Medium
**Mitigation**: The renderer already exists and is optimized; removing manual code should improve performance

### Risk 3: Missing Edge Cases
**Likelihood**: Low
**Impact**: Medium
**Mitigation**: Start with simple functions (getGreeting, formatPrice), then test complex cases (arrow functions, async, nested)

## Dependencies

### Internal Dependencies
- `renderer` package - Must be properly integrated into server
- `transformer` package - Already implements alpineDataFormatter correctly
- `parser` package - Already extracts fence sections correctly

### External Dependencies
- None (Alpine.js already integrated)

## Timeline

**Estimated Duration**: 2-4 hours

- **Phase 1** (Server updates): 1-2 hours
- **Phase 2** (Verification): 30 minutes
- **Phase 3** (Test file updates): 30 minutes
- **Phase 4** (Integration testing): 30-60 minutes

## Open Questions

None - Solution is clear and well-defined.

## Related Documentation

- `.agent-os/bugs/attribute-expression-transform-bug.md` - Bug #3 (different issue)
- `docs/COMPREHENSIVE_TEST_CHECKLIST.md` - Documents Bug #1 impact on testing
- `CLAUDE.md` - Project architecture overview
- `transformer/alpine.go` - alpineDataFormatter implementation
- `renderer/render.go` - Proper rendering pipeline

## Approval Checklist

- [ ] Requirements reviewed and approved
- [ ] Technical approach validated
- [ ] Success criteria clear and measurable
- [ ] Risks identified and mitigated
- [ ] Timeline reasonable
- [ ] Ready for task breakdown (`/create-tasks`)
