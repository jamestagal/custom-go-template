# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-03-nested-conditionals-fix/spec.md

> Created: 2025-10-03
> Version: 1.0.0

## Technical Requirements

### 1. Parser Modification (parser/directives.go)

**File**: `parser/directives.go`
**Function**: `BlockConditionalParser()`

**Current Problem**:
```go
// Current code (simplified)
for i := afterCondition; i < len(remaining); {
    if strings.HasPrefix(remaining[i:], "{/if}") {
        // BUG: Closes block immediately without checking depth
        return Result{Success: true, Value: conditional}
    }
}
```

**Required Changes**:
```go
// Add depth tracking
depth := 1  // Start at depth 1 (we're inside the opening {if})

for i := afterCondition; i < len(remaining); {
    // Increment depth for nested {if}
    if strings.HasPrefix(remaining[i:], "{if ") {
        depth++
        i++
        continue
    }

    // Decrement depth for {/if}
    if strings.HasPrefix(remaining[i:], "{/if}") {
        depth--
        if depth == 0 {
            // This is our closing tag
            return Result{Success: true, Value: conditional}
        }
        // This {/if} belongs to a nested conditional
        i += 5
        continue
    }

    // Only recognize {else if} and {else} at our depth (depth == 1)
    if depth == 1 {
        if strings.HasPrefix(remaining[i:], "{else if ") {
            // Parse else-if clause
        }
        if strings.HasPrefix(remaining[i:], "{else}") {
            // Parse else clause
        }
    }

    i++
}
```

### 2. Test Coverage (tests/alpine/conditionals_test.go)

**New Test Cases Required**:

1. **TestNestedConditionalsWithElseIf**: Test `{else if}` after nested `{if}` block
2. **TestDeeplyNestedConditionals**: Test 3+ levels of nesting
3. **TestNestedConditionalsInLoops**: Test `{if}` blocks inside `{for}` loops
4. **TestLoopsInConditionals**: Test `{for}` loops inside `{if}` blocks
5. **TestMixedNesting**: Test combinations of loops and conditionals

**Example Test**:
```go
func TestNestedConditionalsWithElseIf(t *testing.T) {
    template := `
{if outer}
  <div>Outer</div>
  {if inner}
    <div>Inner</div>
  {/if}
{else if other}
  <div>Other</div>
{else}
  <div>Default</div>
{/if}`

    // Parse and verify AST structure
    // Transform and verify Alpine.js output
    // Ensure all three branches present
}
```

### 3. Validation Requirements

- **No Regression**: All 294+ existing tests must continue to pass
- **AST Structure**: Nested conditionals must create proper nested AST nodes
- **Alpine Output**: Generated HTML must have correct `<template x-if>`, `<template x-else-if>`, `<template x-else>` nesting
- **Edge Cases**: Handle up to 5 levels of nesting without performance degradation

### 4. Documentation Update

**File**: `docs/template-syntax.md`

Add section:
```markdown
### Nested Conditionals

You can nest conditionals arbitrarily deep:

\`\`\`
{if condition1}
  <div>Level 1</div>
  {if condition2}
    <div>Level 2</div>
  {/if}
{else if condition3}
  <div>Alternative</div>
{else}
  <div>Default</div>
{/if}
\`\`\`

The parser tracks nesting depth to ensure each `{/if}` matches its corresponding `{if}`.
```

## Approach

### Implementation Steps

1. **Add Depth Counter** - Initialize `depth := 1` at start of `BlockConditionalParser()`
2. **Track Opening Tags** - Increment depth when encountering `{if ` prefix
3. **Track Closing Tags** - Decrement depth when encountering `{/if}` prefix
4. **Conditional Closure** - Only return from function when `depth == 0`
5. **Clause Recognition** - Only parse `{else if}` and `{else}` when `depth == 1`
6. **Test First** - Write failing tests before implementing fix
7. **Validate** - Run full test suite to ensure no regression

### Edge Cases to Handle

- **Empty Nested Conditionals**: `{if outer}{if inner}{/if}{/if}`
- **Adjacent Nested Conditionals**: Multiple nested conditionals in sequence
- **Malformed Templates**: Missing `{/if}` tags (should error gracefully)
- **Whitespace Variations**: `{if}`, `{ if }`, tabs vs spaces

## Performance Considerations

- **Time Complexity**: O(n) where n is template length (same as before)
- **Space Complexity**: O(1) for depth counter (minimal overhead)
- **No Impact**: Depth tracking adds negligible overhead (~1 integer increment/decrement per conditional tag)

## External Dependencies

None - this is a pure parser logic fix using existing Go standard library.
