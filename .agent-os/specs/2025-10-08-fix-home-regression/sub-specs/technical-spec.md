# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-08-fix-home-regression/spec.md

## Technical Requirements

### Investigation Phase

**Use go-backend agent exclusively for all investigation and implementation work.**

1. **Identify Root Cause**
   - Analyze `parser/fence.go` - `ParseFenceContentWithStores()` function
   - Compare with `ParseFenceContent()` to identify differences in function preservation
   - Trace how fence functions are extracted and stored during component registration
   - Identify where functions are lost in the parsing/rendering pipeline

2. **Analyze Component Flow**
   - Review `cmd/server/main.go` - component registration at lines 111-127 (conditional fence re-parsing)
   - Review `renderer/component.go` - how fence functions are rendered in x-data
   - Review `renderer/fence.go` - fence section rendering logic
   - Identify all code paths that handle fence functions

### Implementation Requirements

1. **Preserve Functions in ParseFenceContentWithStores**
   - Modify `parser/fence.go` to preserve function definitions when parsing with store registry
   - Ensure function AST nodes are maintained in FenceSection.Functions
   - Maintain compatibility with store import resolution

2. **Component Registration Fix**
   - Review conditional fence re-parsing logic in `cmd/server/main.go` (lines 111-127)
   - Ensure functions are preserved when re-parsing fence sections with store imports
   - Consider alternative approaches: merge fence data instead of replace, or skip re-parsing if no stores

3. **Rendering Pipeline Fix**
   - Ensure `renderer/fence.go` includes all functions in x-data output
   - Verify function rendering works with both store and non-store components
   - Maintain proper JavaScript syntax for function definitions

### Error Handling

1. **Graceful Degradation**
   - If function parsing fails, log clear error message
   - Don't silently strip functions - fail loudly with actionable error
   - Preserve original fence content if parsing fails

2. **Validation**
   - Add validation that fence functions list is preserved after parsing
   - Add warning if function count changes during parsing
   - Log which functions are found vs rendered

### Testing Requirements

1. **Unit Tests**
   - Test `ParseFenceContentWithStores()` preserves functions
   - Test fence parsing with stores AND functions
   - Test component registration preserves all fence data

2. **Integration Tests**
   - Test UserProfile component renders with functions
   - Test components with both store imports AND functions
   - Test components with only stores, only functions, and both

3. **Regression Tests**
   - Test home.html renders without errors
   - Test all UserProfile instances display correctly
   - Test store-components-demo continues working

### Files to Modify

**Parser**:
- `parser/fence.go` - ParseFenceContentWithStores() function preservation
- `parser/expressions.go` - If function expression parsing affected

**Server**:
- `cmd/server/main.go` - Component registration fence re-parsing logic (lines 111-127)

**Renderer**:
- `renderer/fence.go` - Ensure functions included in x-data
- `renderer/component.go` - Component rendering with functions

**Tests**:
- `parser/fence_test.go` - Add function preservation tests
- `tests/integration/component_functions_test.go` - New integration tests
- `tests/integration/home_page_test.go` - New regression test

### Performance Considerations

- Fence parsing should not be significantly slower
- Avoid re-parsing fence sections multiple times
- Consider caching parsed fence data with functions

### Success Criteria

1. ✅ Home page renders without console errors
2. ✅ All UserProfile components display complete data
3. ✅ Store demo page continues working
4. ✅ Tests pass: `go test ./parser -v`, `go test ./tests/integration -v`
5. ✅ No new console errors on any existing pages
