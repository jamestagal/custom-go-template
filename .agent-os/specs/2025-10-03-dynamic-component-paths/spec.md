# Spec 4: Dynamic Component Paths

**Status**: Draft
**Created**: 2025-10-03
**Priority**: High
**Estimated Effort**: Medium (4-6 hours)

## Problem Statement

The custom Go template engine currently only supports static component paths where component names are hardcoded at parse time:

```html
<ComponentName prop={value} />
```

The original developer implemented an innovative `<=` syntax for dynamic component paths, allowing runtime component selection based on variables:

```html
<='./views/{comp}.html' age={age + 1} />
<='{path}' />
```

This feature enables:
- Dynamic component loading based on user data
- Conditional component rendering without {if} blocks
- More flexible component composition patterns

Currently missing from our implementation, this would bring us to 100% feature parity with the original project.

## Requirements

### Functional Requirements

1. **Parse Dynamic Component Syntax**
   - Support `<='path' />` syntax
   - Extract dynamic path expression within quotes
   - Support both single and double quotes
   - Handle props after path expression

2. **Variable Interpolation in Paths**
   - Support `{variable}` within path strings
   - Example: `<='./views/{comp}.html' />`
   - Extract variables from path expression

3. **Runtime Component Resolution**
   - Resolve component path at transformation time
   - Look up component template from registry
   - Fall back gracefully if component not found

4. **Prop Passing**
   - Support all prop types (dynamic, shorthand, static)
   - Example: `<='path' prop={value} {shorthand} static="value" />`

### Non-Functional Requirements

1. **Performance**: No significant overhead vs static components
2. **Error Handling**: Clear messages when path resolution fails
3. **Testing**: Comprehensive test coverage for all path patterns
4. **Cognitive Load**: Keep all functions < 30 complexity

## Success Criteria

- [ ] Parser recognizes `<=` syntax
- [ ] Dynamic paths with variables parse correctly
- [ ] Component resolution works at transformation time
- [ ] Props pass correctly to dynamic components
- [ ] Error messages are clear and actionable
- [ ] All tests pass with > 90% coverage
- [ ] No regressions in existing functionality

## Technical Approach

### AST Changes

Add new node type `ast.DynamicComponentNode`:
```go
type DynamicComponentNode struct {
    PathExpression string              // The dynamic path with variables
    Props          []Attribute          // Props to pass
    SelfClosing    bool
}
```

### Parser Changes

Add `DynamicComponentParser()` in `parser/components.go`:
```go
func DynamicComponentParser() Parser {
    return func(input string) Result {
        // 1. Match <= prefix
        // 2. Extract quoted path expression
        // 3. Parse props
        // 4. Return DynamicComponentNode
    }
}
```

### Transformer Changes

Add `transformDynamicComponent()` in `transformer/components.go`:
```go
func transformDynamicComponent(node *ast.DynamicComponentNode, dataScope map[string]any) []ast.Node {
    // 1. Extract variables from path expression
    // 2. Resolve path (if possible at compile time)
    // 3. Look up component template
    // 4. Transform like regular component
}
```

### Renderer Changes

Minimal changes - dynamic components render like regular components after transformation.

## Testing Strategy

### Unit Tests

1. **Parser Tests** (`parser/components_test.go`)
   - Parse `<='./path.html' />`
   - Parse `<='./views/{comp}.html' />`
   - Parse with props: `<='path' prop={value} />`
   - Parse both quote styles

2. **Transformer Tests** (`transformer/components_test.go`)
   - Transform with static path
   - Transform with dynamic variables
   - Handle missing component gracefully
   - Prop passing to dynamic components

3. **Integration Tests** (`tests/alpine/dynamic_components_test.go`)
   - Full template with dynamic components
   - Nested dynamic components
   - Dynamic components in loops
   - Error cases

### Edge Cases

- Missing component file
- Invalid path expression
- Circular component references
- Dynamic components with recursive content

## Implementation Plan

### Task 1: Add DynamicComponentNode to AST ✅
- Create new node type in `ast/ast.go`
- Add to type switch cases in transformer
- Update visitor patterns if needed

**Cognitive Load**: 5

### Task 2: Implement DynamicComponentParser ✅
- Add parser in `parser/components.go`
- Handle `<=` prefix
- Extract quoted path
- Parse props
- Write parser tests

**Cognitive Load**: 15

### Task 3: Implement transformDynamicComponent ✅
- Add transformation logic in `transformer/components.go`
- Extract path variables
- Resolve component template
- Pass props correctly
- Write transformation tests

**Cognitive Load**: 20

### Task 4: Integration and Testing ✅
- Add integration tests
- Test with Alpine.js output
- Verify prop passing
- Test error cases

**Cognitive Load**: 12

### Task 5: Documentation and Cleanup ✅
- Update CLAUDE.md with new syntax
- Add examples to examples/
- Document dynamic component patterns
- Final test run

**Cognitive Load**: 8

**Total Cognitive Load**: 60 (distributed across 5 tasks)

## Dependencies

- Spec 1 (Recursive Component Transformation): ✅ Complete
- Spec 2 (Function Expression Handling): ✅ Complete
- Spec 3 (Loop Rendering): ✅ Complete
- Component registry system: ✅ Exists in cmd/server/main.go

## Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Path resolution complexity | High | Medium | Start with static paths, add variables incrementally |
| Performance overhead | Medium | Low | Cache resolved components |
| Circular references | High | Low | Add reference tracking, fail gracefully |
| Test coverage gaps | Medium | Medium | TDD approach, write tests first |

## Comparison with Original Implementation

**Original Approach** (from main.go lines 708-733):
```go
if strings.HasPrefix(markup[i:], "<=") {
    dynamicCompPath := markup[startDynamicCompPathIndex:endDynamicCompPathIndex]
    // Evaluate path from props at runtime
}
```

**Our Approach**:
- Parse to AST (more structured)
- Transform through pipeline (more testable)
- Use parser combinators (more composable)
- Type-safe node handling (fewer bugs)

## Future Enhancements

- Optional Goja integration for computed paths
- Path aliases/shortcuts
- Component preloading hints
- Development-time path validation

## References

- Original implementation: `/Users/benjaminwaller/Projects/Jim Fisk/Jim go template/jim_custom_go_template/main.go` (lines 708-733)
- Comparison document: `docs/OriginalVsCurrentComparison.md`
- Parser patterns: `parser/components.go`
- Component transformation: `transformer/components.go`
