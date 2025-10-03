# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-03-nested-conditionals-fix/spec.md

> Created: 2025-10-03
> Status: Ready for Implementation

## Tasks

### Phase 1: Test Creation (TDD Approach)

- [ ] Create `TestNestedConditionalsWithElseIf` in `tests/alpine/conditionals_test.go`
  - Test `{else if}` after nested `{if}` block
  - Verify AST contains all three branches (if/else-if/else)
  - Verify Alpine.js output has proper `<template x-if>`, `<template x-else-if>`, `<template x-else>`

- [ ] Create `TestDeeplyNestedConditionals` in `tests/alpine/conditionals_test.go`
  - Test 3+ levels of nesting
  - Verify each level creates proper AST nodes
  - Verify Alpine.js output maintains nesting structure

- [ ] Create `TestNestedConditionalsInLoops` in `tests/alpine/conditionals_test.go`
  - Test `{if}` blocks inside `{for}` loops
  - Verify loop iteration doesn't interfere with conditional depth tracking

- [ ] Create `TestLoopsInConditionals` in `tests/alpine/conditionals_test.go`
  - Test `{for}` loops inside `{if}` blocks
  - Verify conditional depth doesn't affect loop parsing

- [ ] Create `TestMixedNesting` in `tests/alpine/conditionals_test.go`
  - Test combinations of loops and conditionals
  - Verify complex nesting scenarios work correctly

- [ ] Run tests and verify they fail (confirming the bug exists)
  - `go test ./tests/alpine -run TestNested -v`
  - Document failure output

### Phase 2: Parser Implementation

- [ ] Add depth counter to `BlockConditionalParser()` in `parser/directives.go`
  - Initialize `depth := 1` at function start
  - Add comment explaining depth tracking purpose

- [ ] Implement depth increment logic
  - Detect `{if ` prefix
  - Increment `depth++`
  - Add continue to skip further processing

- [ ] Implement depth decrement logic
  - Detect `{/if}` prefix
  - Decrement `depth--`
  - Only return if `depth == 0`
  - Otherwise continue parsing

- [ ] Update `{else if}` recognition
  - Only parse when `depth == 1`
  - Ignore `{else if}` at other depths

- [ ] Update `{else}` recognition
  - Only parse when `depth == 1`
  - Ignore `{else}` at other depths

### Phase 3: Validation

- [ ] Run new test cases
  - `go test ./tests/alpine -run TestNested -v`
  - Verify all new tests pass

- [ ] Run full test suite
  - `go test ./... -v`
  - Verify no regression (all 294+ tests pass)

- [ ] Test `home.html` manually
  - `go run cmd/server/main.go`
  - Visit `http://localhost:3000/home.html`
  - Verify lines 30-39 render correctly with all three conditional branches

- [ ] Test edge cases
  - Empty nested conditionals
  - Adjacent nested conditionals
  - Malformed templates (missing `{/if}`)

### Phase 4: Documentation

- [ ] Create or update `docs/template-syntax.md`
  - Add "Nested Conditionals" section
  - Include examples of nested conditionals
  - Explain depth tracking behavior

- [ ] Update inline code comments
  - Add comments explaining depth tracking in `parser/directives.go`
  - Document any edge cases handled

- [ ] Update CHANGELOG (if exists)
  - Document bug fix
  - Note breaking changes (if any)

### Phase 5: Cleanup

- [ ] Review code for clarity
  - Ensure variable names are descriptive
  - Remove any debug logging

- [ ] Format code
  - `go fmt ./...`

- [ ] Run linter (if configured)
  - `golangci-lint run` or equivalent

- [ ] Final test run
  - `go test ./... -v`
  - Confirm 100% pass rate
