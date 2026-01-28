# Nested Conditionals Parsing Fix - Lite Summary

Fix parser bug where `{else if}` and `{else}` after nested `{if}` blocks are treated as text instead of conditional clauses. Add depth tracking to `BlockConditionalParser()` so each `{if}` increments depth, each `{/if}` decrements it, and blocks only close at depth 0. This enables proper rendering of arbitrarily nested conditional logic in templates.

## Key Points
- Add depth counter to `BlockConditionalParser()` in `parser/directives.go`
- Track nesting: increment on `{if}`, decrement on `{/if}`, close only at depth 0
- Recognize `{else if}`/`{else}` only at current block depth (depth == 1)
- Add comprehensive test coverage for nested conditionals
- Fix `home.html` lines 30-39 to render all three conditional branches correctly
