# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-02-loop-rendering-integration/spec.md

> Created: 2025-10-02
> Status: Ready for Implementation

## Tasks

### Phase 1: Investigation and Diagnosis

#### Task 1.1: Review Current Loop Implementation
**File**: `transformer/loops.go`

- [ ] Read the complete `transformLoop()` function implementation
- [ ] Document how `node.Iterator` and `node.Collection` are currently handled
- [ ] Verify how the x-for expression is formatted
- [ ] Check how loop body content is transformed
- [ ] Identify if `transformNodes()` is called recursively on loop body
- [ ] Document any obvious issues or anti-patterns

**Output**: Written summary of current implementation in investigation notes

#### Task 1.2: Analyze Failing Tests
**Files**: `tests/alpine/alpine_integration_test.go`, `tests/alpine/loops_test.go`

- [ ] Run `go test ./tests/alpine -v -run TestAlpineIntegration/loop_rendering 2>&1 | tee loop_test_output.txt`
- [ ] Run `go test ./tests/alpine -v -run TestAlpineIntegration/nested_conditionals_and_loops 2>&1 | tee nested_test_output.txt`
- [ ] Document what the tests expect (expected output)
- [ ] Document what is actually being generated (actual output)
- [ ] Create a diff showing the specific discrepancies
- [ ] Identify root cause categories (e.g., scope issue, format issue, transformation order)

**Output**: Test failure analysis document with expected vs actual comparison

#### Task 1.3: Trace Scope Handling
**File**: `transformer/loops.go`

- [ ] Check if `node.Collection` is being added to `dataScope`
- [ ] Verify if `extractVariablesFromExpr()` is called on the collection
- [ ] Check if iterator variable is being added to parent scope (should NOT be)
- [ ] Identify if child scope is created for loop body
- [ ] Verify that loop body transformations use the child scope
- [ ] Check for scope leakage issues

**Output**: Scope flow diagram showing variable visibility at each transformation level

#### Task 1.4: Review Integration Points
**Files**: `transformer/conditionals.go`, `transformer/components.go`

- [ ] Verify that conditionals call `transformNodes()` on their content
- [ ] Check if components inside loops would have access to iterator variables
- [ ] Identify any special handling needed for loops in conditionals
- [ ] Check for any hardcoded assumptions that might break nested structures

**Output**: Integration checklist identifying potential conflict points

---

### Phase 2: Scope Handling Fixes

#### Task 2.1: Fix Collection Variable Scope
**File**: `transformer/loops.go`

- [ ] Ensure `extractVariablesFromExpr(node.Collection, dataScope)` is called
- [ ] Handle nested property accesses (e.g., `user.posts`, `category.items`)
- [ ] Add logging to verify collection is added to parent scope
- [ ] Test that collection variable is available in x-data

**Test**: Collection variable appears in final x-data scope

#### Task 2.2: Implement Loop Body Child Scope
**File**: `transformer/loops.go`, potentially `transformer/scope.go`

- [ ] Create `CreateChildScope(parentScope)` function if it doesn't exist
- [ ] Modify `transformLoop()` to create child scope for loop body
- [ ] Add iterator variable to child scope (not parent scope)
- [ ] Pass child scope to `transformNodes()` when transforming loop body
- [ ] Verify iterator is available inside loop but not outside

**Test**: Iterator variable works inside loop but doesn't leak to parent

#### Task 2.3: Fix Iterator Scope Leakage
**File**: `transformer/loops.go`

- [ ] Remove any code that adds iterator to parent scope
- [ ] Verify iterator is only in loop body scope
- [ ] Add validation that iterator doesn't appear in parent x-data
- [ ] Test nested loops to ensure iterators don't conflict

**Test**: `go test ./tests/alpine -v -run Loop` - no iterator leakage errors

#### Task 2.4: Handle Loop Index Variable (if supported)
**Files**: `ast/ast.go`, `transformer/loops.go`

- [ ] Check if AST Loop struct has an `Index` field
- [ ] If yes, add index variable to loop body scope
- [ ] Format x-for expression as `(item, index) in items`
- [ ] If no, document as future enhancement

**Test**: If supported, index variable works correctly in loop body

---

### Phase 3: Loop Transformation Fixes

#### Task 3.1: Fix x-for Expression Format
**File**: `transformer/loops.go`

- [ ] Implement: `xForExpr := fmt.Sprintf("%s in %s", node.Iterator, node.Collection)`
- [ ] Remove any extra spaces or quotes that don't match Alpine.js syntax
- [ ] Verify iterator name is preserved exactly as written
- [ ] Test with various collection expressions (simple, nested, array access)

**Test**: Generated x-for expressions match Alpine.js expectations exactly

#### Task 3.2: Fix Loop Body Transformation
**File**: `transformer/loops.go`

- [ ] Ensure `transformNodes(node.Content, loopBodyScope, false)` is called
- [ ] Verify expressions inside loop body are transformed to x-text spans
- [ ] Check that child elements are transformed recursively
- [ ] Ensure loop body is placed as children of template element

**Test**: Loop body content is fully transformed with proper x-text wrapping

#### Task 3.3: Create Template Element Correctly
**File**: `transformer/loops.go`

- [ ] Create `&ast.Element{TagName: "template"}` for loop wrapper
- [ ] Add x-for attribute with correct properties:
  - `Name: "x-for"`
  - `Value: xForExpr`
  - `Dynamic: true`
  - `IsAlpine: true`
  - `AlpineType: "for"`
- [ ] Set `Children: transformedContent`
- [ ] Return as single-element slice

**Test**: Template element structure matches expected AST output

#### Task 3.4: Add Error Handling and Validation
**File**: `transformer/loops.go`

- [ ] Add check for empty `node.Iterator` with warning and fallback
- [ ] Add check for empty `node.Collection` with warning and skip
- [ ] Handle empty `node.Content` gracefully
- [ ] Add error logging for malformed loop nodes

**Test**: Malformed loops produce helpful errors without crashing

---

### Phase 4: Nested Structure Support

#### Task 4.1: Test Nested Loops
**File**: Create/update `tests/alpine/nested_loops_test.go`

- [ ] Create test for loop within loop (2 levels)
- [ ] Create test for loop within loop within loop (3 levels)
- [ ] Test different iterator names (item, category, product, etc.)
- [ ] Test accessing outer iterator from inner loop
- [ ] Verify each loop has correct scope isolation

**Test**: `go test ./tests/alpine -v -run NestedLoops` passes

#### Task 4.2: Test Loops in Conditionals
**File**: Update `tests/alpine/alpine_integration_test.go`

- [ ] Test loop inside if block
- [ ] Test loop inside else block
- [ ] Test conditional inside loop
- [ ] Test mixed nesting (loop in if in loop)
- [ ] Verify both x-if and x-for render correctly

**Test**: `go test ./tests/alpine -v -run TestAlpineIntegration/nested_conditionals_and_loops` passes

#### Task 4.3: Test Conditionals in Loops
**File**: Create/update test file

- [ ] Create test for {if} inside {for}
- [ ] Test {else if} and {else} inside loops
- [ ] Verify conditional can access iterator variable
- [ ] Test nested conditionals inside loops

**Test**: All conditional + loop combinations work correctly

#### Task 4.4: Verify Scope Isolation in Nested Structures
**Files**: `transformer/loops.go`, test files

- [ ] Add debug logging to track scope creation
- [ ] Run nested tests with logging enabled
- [ ] Verify each nested level has correct scope chain
- [ ] Check that variables resolve to correct scope level
- [ ] Ensure no scope pollution between sibling structures

**Test**: Scope chain is correct for all nesting levels

---

### Phase 5: Components in Loops

#### Task 5.1: Review Component Transformation in Loop Context
**File**: `transformer/components.go`

**Note**: This task may be blocked by Spec 1 (Recursive Component Transformation)

- [ ] Check if component transformation receives loop body scope
- [ ] Verify `resolvePropValue()` can look up iterator variables
- [ ] Test simple component in loop (no nested components)
- [ ] Identify any issues specific to loop context

**Test**: Simple component in loop renders correctly

#### Task 5.2: Test Component Props Bound to Iterator
**File**: Create test in `tests/alpine/components_in_loops_test.go`

- [ ] Create test: `<ComponentName prop={item}>`
- [ ] Create test: `<ComponentName prop={item.field}>`
- [ ] Create test: component with multiple props from iterator
- [ ] Verify props resolve correctly in component scope

**Test**: Component props correctly resolve iterator variables

#### Task 5.3: Test Nested Components in Loops
**File**: Same as 5.2

**Dependency**: Requires Spec 1 to be completed

- [ ] Create test: component with child component in loop
- [ ] Verify both parent and child components render
- [ ] Check that iterator variable is available at all component levels
- [ ] Test prop passing through component hierarchy in loop

**Test**: Nested components in loops work correctly

#### Task 5.4: Integration Testing with Components
**File**: Update integration tests

- [ ] Test real-world scenario: product list with ProductCard components
- [ ] Test: list of user comments with nested reply components
- [ ] Verify x-data scoping doesn't conflict between loop and component
- [ ] Check that Alpine.js reactivity works correctly

**Test**: Full integration scenarios pass

---

### Phase 6: Testing and Validation

#### Task 6.1: Fix Target Test Cases
**Files**: `tests/alpine/alpine_integration_test.go`

- [ ] Run `go test ./tests/alpine -v -run TestAlpineIntegration/loop_rendering`
- [ ] Verify test passes completely
- [ ] Run `go test ./tests/alpine -v -run TestAlpineIntegration/nested_conditionals_and_loops`
- [ ] Verify test passes completely
- [ ] Document any remaining issues

**Success Criteria**: Both target tests pass without errors

#### Task 6.2: Run Full Loop Test Suite
**Command**: `go test ./tests/alpine -v -run Loop`

- [ ] Fix any failing loop-specific tests
- [ ] Add any missing test coverage identified during implementation
- [ ] Verify all loop variations work (simple, nested, with conditionals)

**Success Criteria**: All loop tests pass

#### Task 6.3: Run Full Alpine Integration Suite
**Command**: `go test ./tests/alpine -v`

- [ ] Run complete test suite
- [ ] Identify any regressions in non-loop tests
- [ ] Fix any regressions caused by loop changes
- [ ] Verify no scope or transformation issues in other features

**Success Criteria**: All Alpine.js integration tests pass

#### Task 6.4: Manual Browser Testing
**Command**: `go run cmd/server/main.go`

- [ ] Create example page with simple loop
- [ ] Create example with nested loops
- [ ] Create example with loop in conditional
- [ ] Open http://localhost:3000 and verify rendering
- [ ] Check browser console for Alpine.js errors
- [ ] Test reactivity (data changes trigger re-renders)

**Success Criteria**: All examples render correctly with no console errors

#### Task 6.5: Performance and Edge Case Testing

- [ ] Test loop with empty collection
- [ ] Test loop with single item
- [ ] Test loop with large collection (100+ items)
- [ ] Test loop with complex nested structures
- [ ] Verify performance is acceptable
- [ ] Check memory usage doesn't leak

**Success Criteria**: All edge cases handled gracefully

---

### Phase 7: Documentation and Cleanup

#### Task 7.1: Add Code Comments
**File**: `transformer/loops.go`

- [ ] Document `transformLoop()` function purpose and parameters
- [ ] Add comments explaining scope handling strategy
- [ ] Document x-for expression format requirements
- [ ] Add examples in comments showing input/output

**Output**: Well-commented loop transformation code

#### Task 7.2: Update Debug Logging
**File**: `transformer/loops.go`

- [ ] Add structured logging for loop transformation start/end
- [ ] Log iterator and collection names
- [ ] Log scope state before/after loop transformation
- [ ] Ensure logs are helpful for debugging but not overly verbose

**Output**: Useful debug logs for troubleshooting

#### Task 7.3: Remove Temporary Debug Code

- [ ] Remove any temporary print statements added during debugging
- [ ] Remove commented-out experimental code
- [ ] Clean up any TODO comments that were resolved
- [ ] Ensure code follows project style guidelines

**Output**: Clean, production-ready code

#### Task 7.4: Verify Example Files
**Directory**: `examples/`

- [ ] Check if example files use loop syntax
- [ ] Update examples if loop syntax has changed
- [ ] Add new example showing loops if none exists
- [ ] Verify all examples work with new loop implementation

**Output**: Working example files demonstrating loop features

---

## Implementation Notes

### Recommended Implementation Order

1. **Phase 1** (Investigation) - MUST be done first to understand current state
2. **Phase 2** (Scope Handling) - Foundation for correct loop transformation
3. **Phase 3** (Loop Transformation) - Core functionality fixes
4. **Phase 4** (Nested Structures) - Build on working basic loops
5. **Phase 5** (Components in Loops) - May need to wait for Spec 1 completion
6. **Phase 6** (Testing) - Continuous throughout implementation
7. **Phase 7** (Documentation) - Final cleanup

### Dependencies

- **Spec 1 (Recursive Component Transformation)**: Required for Task 5.3 and 5.4
- **Spec 2 (Function Expression Handling)**: Independent, can be done in parallel

### Key Files Modified

- `transformer/loops.go` - Primary implementation file
- `transformer/scope.go` - May need scope utility functions
- `tests/alpine/alpine_integration_test.go` - Validation
- `tests/alpine/loops_test.go` - Loop-specific tests
- `tests/alpine/nested_loops_test.go` - New test file for nested scenarios

### Testing Strategy

- Test-first approach: Review failing tests before implementing fixes
- Unit tests for individual loop transformation
- Integration tests for nested structures
- Browser tests for Alpine.js reactivity
- Regression tests to ensure no breaking changes

### Success Metrics

- [ ] `TestAlpineIntegration/loop_rendering` passes
- [ ] `TestAlpineIntegration/nested_conditionals_and_loops` passes
- [ ] All loop-related tests pass
- [ ] No Alpine.js console errors in browser
- [ ] Iterator variables properly scoped
- [ ] Collection variables added to x-data
- [ ] Nested loops work correctly
- [ ] Loops in conditionals work correctly
- [ ] Components in loops work (after Spec 1)
