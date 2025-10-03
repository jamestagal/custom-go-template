# Spec Requirements Document

> Spec: Nested Conditionals Parsing Fix
> Created: 2025-10-03
> Status: Planning

## Overview

Fix the parser bug where `{else if}` and `{else}` clauses appearing after nested `{if}` blocks are incorrectly treated as text expressions instead of conditional clauses, enabling proper rendering of arbitrarily nested conditional logic in templates.

## User Stories

### Parser Developer

As a parser developer, I want the conditional parser to correctly track nesting depth, so that nested `{if}` blocks don't cause outer `{else if}` and `{else}` clauses to be misinterpreted.

**Workflow**: When parsing a template with nested conditionals, the parser should increment a depth counter for each opening `{if}`, decrement for each closing `{/if}`, and only recognize `{else if}`/`{else}` at the current block's depth level. This ensures that the first `{/if}` encountered closes the innermost conditional, not the outer one.

### Template Author

As a template author, I want to write nested conditional logic naturally, so that I can express complex UI logic without workarounds.

**Workflow**: Authors can write templates like:
```
{if outer_condition}
  <div>Outer true</div>
  {if inner_condition}
    <div>Inner true</div>
  {/if}
{else if fallback_condition}
  <div>Fallback</div>
{else}
  <div>Default</div>
{/if}
```

And have all branches render correctly without the parser treating `{else if}` as plain text.

### End User

As an end user viewing a page, I want conditional UI elements to display correctly based on data state, so that I see the appropriate content for my situation.

**Workflow**: When viewing `home.html` with nested conditionals, users should see the correct conditional branch render (e.g., "is a long name" vs "is medium" vs "is a short name") based on the `name.length` value, without seeing raw template syntax like `{else if}` displayed as text.

## Spec Scope

1. **Depth Tracking** - Add nesting depth counter to `BlockConditionalParser()` that increments on `{if}` and decrements on `{/if}`
2. **Conditional Closure Logic** - Only close a conditional block when depth returns to 0, not on the first `{/if}` encountered
3. **Clause Recognition** - Only recognize `{else if}` and `{else}` when at depth 1 (the current block's level)
4. **Comprehensive Tests** - Add test cases for nested conditionals, deeply nested conditionals, and conditionals within loops
5. **Documentation Update** - Add nested conditional examples to template syntax documentation

## Out of Scope

- Refactoring to token-based parsing (future enhancement)
- AST-first parsing architecture (out of scope for this fix)
- Loop nesting improvements (separate concern)
- Performance optimization beyond O(1) depth tracking

## Expected Deliverable

1. `examples/pages/home.html` lines 30-39 render correctly with all three conditional branches (if/else-if/else) working
2. All existing conditional tests continue to pass (no regression)
3. New test cases pass for nested conditionals, deeply nested conditionals, and mixed nesting with loops

## Spec Documentation

- Tasks: @.agent-os/specs/2025-10-03-nested-conditionals-fix/tasks.md
- Technical Specification: @.agent-os/specs/2025-10-03-nested-conditionals-fix/sub-specs/technical-spec.md
