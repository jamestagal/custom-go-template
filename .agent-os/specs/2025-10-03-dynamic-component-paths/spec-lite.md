# Spec 4: Dynamic Component Paths (Lite)

**Date**: 2025-10-03
**Status**: Draft
**Effort**: Medium (4-6 hours)

## What

Add support for dynamic component paths using `<=` syntax from the original project:

```html
<='./views/{comp}.html' age={age + 1} />
<='{path}' />
```

## Why

- Enables runtime component selection based on variables
- Achieves 100% feature parity with original project
- More flexible component composition
- Reduces need for {if} blocks for conditional components

## How

1. Add `DynamicComponentNode` to AST
2. Create `DynamicComponentParser()` to parse `<=` syntax
3. Add `transformDynamicComponent()` to resolve paths and transform
4. Write comprehensive tests (parser, transformer, integration)
5. Update documentation

## Tasks

1. Add DynamicComponentNode to AST (Load: 5)
2. Implement DynamicComponentParser (Load: 15)
3. Implement transformDynamicComponent (Load: 20)
4. Integration and Testing (Load: 12)
5. Documentation and Cleanup (Load: 8)

**Total Load**: 60

## Success

- `<=` syntax parses correctly
- Dynamic paths with variables work
- Props pass to dynamic components
- Clear error messages
- All tests pass
- No regressions
