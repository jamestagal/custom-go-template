# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-03-component-props-rendering-fix/spec.md

> Created: 2025-10-03
> Status: Ready for Implementation

## Tasks

### Investigation Phase

- [ ] Add debug logging to trace Footer component `links` prop through pipeline
- [ ] Trace value from fence parser → AST → transformer → renderer → final HTML
- [ ] Identify exact function/line where truncation occurs
- [ ] Document current behavior with code references and examples
- [ ] Verify fence parser is correctly extracting multi-line values (run existing tests)

### Fix parseValue() Function

- [ ] Write unit tests for `parseValue()` with multi-line JavaScript arrays
- [ ] Write unit tests for `parseValue()` with multi-line JavaScript objects
- [ ] Write unit tests for `parseValue()` with nested structures
- [ ] Update `parseValue()` in `cmd/server/main.go` to detect JavaScript syntax
- [ ] Return raw JavaScript string instead of attempting JSON unmarshaling
- [ ] Verify all new unit tests pass

### Fix Component Transformation

- [ ] Write tests in `transformer/components_test.go` for multi-line prop values
- [ ] Update `TransformComponent()` to preserve JavaScript prop values
- [ ] Ensure props are not marshaled to JSON during transformation
- [ ] Update data scope building to handle JavaScript expressions
- [ ] Add comments explaining why JavaScript syntax is preserved
- [ ] Verify all transformer tests pass

### Fix Component Rendering

- [ ] Write integration test for Footer component with full links array
- [ ] Update `renderer/component.go` if needed to output JavaScript expressions
- [ ] Ensure proper HTML attribute escaping for x-data
- [ ] Handle edge cases: quotes, newlines, special characters in prop values
- [ ] Verify rendered HTML has complete prop values (not truncated)

### Browser Testing

- [ ] Start development server: `go run cmd/server/main.go`
- [ ] Navigate to page with Footer component
- [ ] Inspect HTML source - verify x-data has complete links array
- [ ] Open browser console - verify Alpine.js receives correct data
- [ ] Test component functionality - verify links render correctly
- [ ] Test with multiple components with complex props

### Regression Testing

- [ ] Run full test suite: `go test ./... -v`
- [ ] Verify all existing tests pass
- [ ] Check other components still render correctly
- [ ] Test nested components with props
- [ ] Test components in loops with props

### Documentation

- [ ] Add code comments to parseValue() explaining the fix
- [ ] Add code comments to component transformation explaining JavaScript preservation
- [ ] Update any relevant docs in `docs/` if patterns changed
- [ ] Document edge cases and limitations in code comments

### Final Validation

- [ ] Footer component renders with full links array in browser
- [ ] All navigation links are clickable and functional
- [ ] Inspect element shows complete x-data attribute value
- [ ] No console errors from Alpine.js
- [ ] All tests passing
- [ ] Code review ready
