# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-03-fence-multiline-props-fix/spec.md

> Created: 2025-10-03
> Status: Ready for Implementation

## Tasks

### Phase 1: Implement Bracket Matcher

- [ ] **Task 1.1:** Create `parser/bracket_matcher.go`
  - Implement `BracketMatcher` struct with stack
  - Add `processChar()` method for character-by-character processing
  - Implement string literal detection (single and double quotes)
  - Add escape character handling for quotes
  - Create `isComplete()` method to check if all brackets are matched
  - **Estimated time:** 1 hour
  - **Files:** `parser/bracket_matcher.go` (new)

- [ ] **Task 1.2:** Write bracket matcher unit tests
  - Test simple arrays: `[1, 2, 3]`
  - Test simple objects: `{a: 1, b: 2}`
  - Test nested structures: `{a: [1, 2], b: {c: 3}}`
  - Test strings containing brackets: `{msg: "test [array]"}`
  - Test escaped quotes: `{msg: "it's \"quoted\""}`
  - Test edge cases: empty arrays `[]`, empty objects `{}`
  - Test error cases: mismatched brackets `[}`, unclosed brackets `[1, 2`
  - **Estimated time:** 1 hour
  - **Files:** `parser/bracket_matcher_test.go` (new)

### Phase 2: Update Fence Parser

- [ ] **Task 2.1:** Add multi-line prop value parser
  - Implement `parseMultiLinePropValue()` in `cmd/server/main.go`
  - Use `BracketMatcher` to accumulate lines until complete
  - Handle line accumulation with proper newline insertion
  - Return full value and ending line index
  - **Estimated time:** 1.5 hours
  - **Files:** `cmd/server/main.go`

- [ ] **Task 2.2:** Update `extractComponentProps()` function
  - Detect when prop value starts with `[` or `{`
  - Call `parseMultiLinePropValue()` for multi-line values
  - Advance line index to skip already-processed lines
  - Maintain fast path for single-line props
  - **Estimated time:** 1 hour
  - **Files:** `cmd/server/main.go`

- [ ] **Task 2.3:** Add error handling and validation
  - Detect unclosed brackets at end of fence section
  - Provide clear error message with line number
  - Detect mismatched brackets
  - Handle edge case: fence section ends mid-prop
  - **Estimated time:** 0.5 hours
  - **Files:** `cmd/server/main.go`

### Phase 3: Testing

- [ ] **Task 3.1:** Create fence multi-line parsing tests
  - Test multi-line array parsing
  - Test multi-line object parsing
  - Test nested structures (arrays in objects, objects in arrays)
  - Test function expressions: `new Date().getFullYear()`
  - Test mixed single-line and multi-line props in same fence
  - **Estimated time:** 1.5 hours
  - **Files:** `parser/fence_multiline_test.go` (new)

- [ ] **Task 3.2:** Integration testing with real components
  - Update Footer.html to use multi-line format for `links` prop
  - Update Footer.html to use multi-line format for `socialLinks` prop
  - Test Footer component rendering via dev server
  - Verify Alpine.js receives correct data in x-data
  - Test in browser: check console for x-data object
  - **Estimated time:** 1 hour
  - **Files:** `examples/components/Footer.html`

- [ ] **Task 3.3:** Regression testing
  - Run full test suite: `go test ./... -v`
  - Verify existing components still work (Header, UserProfile)
  - Test simple.html and comprehensive.html pages
  - Verify no performance degradation
  - **Estimated time:** 0.5 hours
  - **Files:** All existing test files

### Phase 4: Documentation

- [ ] **Task 4.1:** Update CLAUDE.md
  - Document multi-line prop syntax
  - Add examples of valid multi-line arrays and objects
  - Document bracket matching behavior
  - Add troubleshooting for common errors
  - **Estimated time:** 0.5 hours
  - **Files:** `CLAUDE.md`

- [ ] **Task 4.2:** Add inline code comments
  - Comment bracket matcher algorithm
  - Comment multi-line parsing logic
  - Document why function expressions are kept as strings
  - **Estimated time:** 0.5 hours
  - **Files:** `parser/bracket_matcher.go`, `cmd/server/main.go`

## Total Estimated Time: 9 hours

## Success Criteria

- [ ] Footer.html renders correctly with multi-line `links` and `socialLinks` arrays
- [ ] All bracket matcher tests pass
- [ ] All fence parsing tests pass
- [ ] No regression in existing components
- [ ] Function expressions preserved as strings for Alpine.js runtime evaluation
- [ ] Clear error messages for malformed multi-line props
- [ ] Documentation updated with examples and usage guidelines

## Dependencies

- No external dependencies required
- All implementation using Go standard library

## Rollback Plan

If issues arise:
1. Revert changes to `cmd/server/main.go` (keep original `extractComponentProps()`)
2. Remove new files: `parser/bracket_matcher.go`, `parser/bracket_matcher_test.go`
3. Components with multi-line props will need to be reformatted to single-line
4. No data loss or breaking changes to existing functionality
