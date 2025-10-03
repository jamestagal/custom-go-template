# Spec: Fix Nested Conditionals Parsing

## Problem Statement

The parser incorrectly handles `{else if}` and `{else}` clauses that appear after a nested `{if}` block within an outer conditional. The parser closes the outer conditional prematurely when it encounters the first `{/if}`, treating subsequent `{else if}` and `{else}` as plain text expressions.

## Current Behavior

### Input Template
```html
{if name.length > 3}
  <div id="praise">{name} is a long name</div>
  {if age > 1}
    <div>Has been born</div>
  {/if}
{else if name.length == 2}
  <div id="praise">{name} is medium</div>
{else}
  <div id="praise">{name} is a short name</div>
{/if}
```

### Actual Output (Broken)
```html
<template x-if="name.length > 3">
  <div id="praise"><span x-text="name"></span> is a long name</div>
  <template x-if="age > 1">
    <div>Has been born</div>
  </template>
</template>
<span x-text="else if name.length == 2"></span>
<div id="praise"><span x-text="name"></span> is medium</div>
<span x-text="else"></span>
<div id="praise"><span x-text="name"></span> is a short name</div>
<span x-text="/if"></span>
```

The parser:
1. Opens the outer `{if name.length > 3}` conditional
2. Parses the nested `{if age > 1}` block
3. Encounters `{/if}` and **incorrectly** closes the outer conditional (should close inner)
4. Treats `{else if name.length == 2}` as a text expression (not recognized as conditional)
5. Treats `{else}` as a text expression
6. Treats `{/if}` as a text expression

## Expected Behavior

### Expected Output (Correct)
```html
<template x-if="name.length > 3">
  <div id="praise"><span x-text="name"></span> is a long name</div>
  <template x-if="age > 1">
    <div>Has been born</div>
  </template>
</template>
<template x-else-if="name.length == 2">
  <div id="praise"><span x-text="name"></span> is medium</div>
</template>
<template x-else>
  <div id="praise"><span x-text="name"></span> is a short name</div>
</template>
```

The parser should:
1. Track conditional nesting depth
2. Match each `{/if}` with its corresponding opening `{if}`
3. Only close a conditional block when the nesting depth returns to the starting level
4. Correctly recognize `{else if}` and `{else}` as part of the outer conditional

## Root Cause Analysis

### File: `parser/directives.go`

The `BlockConditionalParser()` function uses a simple approach:
1. It looks for `{if condition}` to start a block
2. It collects content until it finds `{else if}`, `{else}`, or `{/if}`
3. **It does not track nesting depth** - the first `{/if}` encountered closes the block

### Problematic Code
```go
func BlockConditionalParser() Parser {
    return func(input string) Result {
        // ... parse {if condition} ...

        depth := 0
        for i := afterCondition; i < len(remaining); {
            // Look for {/if}
            if strings.HasPrefix(remaining[i:], "{/if}") {
                // BUG: This closes the block regardless of nesting depth
                return Result{Success: true, Value: conditional}
            }
            // ... other parsing ...
        }
    }
}
```

The parser doesn't increment `depth` when it encounters nested `{if}` blocks, so it can't tell which `{/if}` belongs to which `{if}`.

## Solution Design

### Approach 1: Track Nesting Depth (Recommended)

Modify `BlockConditionalParser()` to track the nesting depth of conditionals:

```go
func BlockConditionalParser() Parser {
    return func(input string) Result {
        // ... initial parsing ...

        depth := 1  // Start at depth 1 (we're inside the outer {if})

        for i := afterCondition; i < len(remaining); {
            // Check for nested {if}
            if strings.HasPrefix(remaining[i:], "{if ") {
                depth++  // Increase depth for nested conditional
                i++
                continue
            }

            // Check for {/if}
            if strings.HasPrefix(remaining[i:], "{/if}") {
                depth--  // Decrease depth
                if depth == 0 {
                    // We've reached the end of our block
                    return Result{Success: true, Value: conditional}
                }
                // Otherwise, this {/if} belongs to a nested conditional
                i += 5  // Skip past {/if}
                continue
            }

            // Check for {else if} and {else} - only at our depth
            if depth == 1 {
                if strings.HasPrefix(remaining[i:], "{else if ") {
                    // This is our else-if clause
                    // ... parse else-if ...
                }
                if strings.HasPrefix(remaining[i:], "{else}") {
                    // This is our else clause
                    // ... parse else ...
                }
            }

            i++
        }
    }
}
```

### Approach 2: Recursive Parsing (Alternative)

Use recursive parsing where each conditional block parser can call itself for nested conditionals:

```go
func parseConditionalBlock(input string, startDepth int) (*ast.Conditional, string, error) {
    // Parse the condition
    // ...

    // Parse content until we find our matching {/if}
    currentDepth := startDepth + 1

    for ... {
        if foundNestedIf {
            // Recursively parse the nested conditional
            nested, remaining, err := parseConditionalBlock(remaining, currentDepth)
            // Add nested to our content
        }

        if foundElseIf && currentDepth == startDepth + 1 {
            // This is our else-if
        }

        if foundEndIf {
            currentDepth--
            if currentDepth == startDepth {
                // This is our closing tag
                break
            }
        }
    }
}
```

## Implementation Plan

### Phase 1: Add Depth Tracking (Recommended First Step)

1. **File**: `parser/directives.go`
2. **Function**: `BlockConditionalParser()`
3. **Changes**:
   - Initialize `depth := 1` at the start of content parsing
   - Increment `depth++` when encountering `{if ` within the block
   - Decrement `depth--` when encountering `{/if}`
   - Only close the block when `depth == 0`
   - Only recognize `{else if}` and `{else}` when `depth == 1`

### Phase 2: Add Comprehensive Tests

Create test cases in `tests/alpine/conditionals_test.go`:

```go
func TestNestedConditionalsWithElseIf(t *testing.T) {
    template := `
{if outer}
  <div>Outer true</div>
  {if inner}
    <div>Inner true</div>
  {/if}
{else if other}
  <div>Other condition</div>
{else}
  <div>All false</div>
{/if}
`

    // Test that:
    // 1. Parser creates correct AST structure
    // 2. Transformer generates proper Alpine.js templates
    // 3. All three branches (if, else-if, else) are present
}

func TestDeeplyNestedConditionals(t *testing.T) {
    template := `
{if level1}
  {if level2}
    {if level3}
      <div>Deep</div>
    {/if}
  {/if}
{else}
  <div>Level 1 false</div>
{/if}
`

    // Test that deeply nested conditionals work correctly
}

func TestNestedConditionalsInLoops(t *testing.T) {
    template := `
{for item of items}
  {if item.active}
    {if item.priority > 5}
      <div>High priority</div>
    {/if}
  {else}
    <div>Inactive</div>
  {/if}
{/for}
`

    // Test conditionals nested within loops
}
```

### Phase 3: Handle Edge Cases

1. **Multiple nesting levels**: `{if a} {if b} {if c} ... {/if} {/if} {/if}`
2. **Conditionals in loops**: `{for x of xs} {if y} ... {/if} {/for}`
3. **Loops in conditionals**: `{if a} {for x of xs} ... {/for} {/if}`
4. **Mixed nesting**: `{if a} {for x of xs} {if y} ... {/if} {/for} {/if}`

### Phase 4: Update Documentation

Update `docs/template-syntax.md` with:
- Clear examples of nested conditionals
- Limitations (if any remain)
- Best practices for complex logic

## Success Criteria

1. ✅ Parser correctly handles `{else if}` after nested `{if}` blocks
2. ✅ All test cases pass (including new nested conditional tests)
3. ✅ The `home.html` example renders correctly in browser
4. ✅ No regression in existing conditional parsing
5. ✅ Documentation updated with nested conditional examples

## Testing Strategy

### Unit Tests
- Test depth tracking with various nesting levels
- Test `{else if}` recognition at correct depth
- Test `{else}` recognition at correct depth
- Test multiple nested blocks

### Integration Tests
- Test full template parsing with nested conditionals
- Test transformation to Alpine.js syntax
- Test rendering in browser

### Regression Tests
- Ensure existing conditional tests still pass
- Ensure simple (non-nested) conditionals work
- Ensure loops still work correctly

## Timeline Estimate

- **Phase 1** (Depth Tracking): 2-4 hours
- **Phase 2** (Tests): 2-3 hours
- **Phase 3** (Edge Cases): 2-3 hours
- **Phase 4** (Documentation): 1 hour

**Total**: ~7-11 hours

## Files to Modify

1. `parser/directives.go` - Add depth tracking to `BlockConditionalParser()`
2. `tests/alpine/conditionals_test.go` - Add nested conditional tests
3. `docs/template-syntax.md` - Document nested conditionals
4. `examples/pages/home.html` - Already has test case (will work after fix)

## Dependencies

- None - this is a self-contained parser fix

## Risks

- **Breaking existing conditionals**: Mitigation - comprehensive regression tests
- **Performance impact**: Mitigation - depth tracking is O(1) per character
- **Complex edge cases**: Mitigation - extensive test coverage

## Alternative Approaches Considered

1. **Token-based parsing**: More robust but requires larger refactor
2. **AST-first parsing**: Better long-term but out of scope for this fix
3. **Simple regex matching**: Too fragile for nested structures

## Related Issues

- This fixes the rendering issue in `examples/pages/home.html` lines 30-39
- Related to general block directive parsing in `parser/directives.go`
- May help with future loop nesting issues (if any)

## References

- Current parser: `parser/directives.go` lines 200-350
- Test file: `tests/alpine/alpine_integration_test.go`
- Documentation: `docs/template-syntax.md`
