# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-06-parser-unification/spec.md

> Created: 2025-10-06
> Status: Ready for Implementation

## Tasks

### Phase 1: Investigation & Preparation

- [ ] **Task 1.1**: Review Investigation Documents
  - Read KNOWN_ISSUES.md completely
  - Read INVESTIGATION_SUMMARY.md completely
  - Understand the two parsing paths bug
  - Review test files: parser/conditional_bug_test.go and parser/nested_conditional_loop_test.go
  - **Estimated Time**: 30 minutes

- [ ] **Task 1.2**: Reproduce Bugs
  - Run development server: `go run cmd/server/main.go`
  - Visit http://localhost:3000
  - Observe Basic Conditionals bug (lines 30-39)
  - Observe Animals Loop bug (lines 61-69)
  - Count "likes" messages (should be 3, currently 1)
  - Take screenshots of current buggy behavior
  - **Estimated Time**: 15 minutes

- [ ] **Task 1.3**: Run Baseline Tests
  - Run: `go test ./parser -v`
  - Run: `go test ./parser -run TestConditionalBug -v` (should pass)
  - Run: `go test ./parser -run TestNestedConditionalLoop -v` (should fail)
  - Run: `go test ./... -v` (note any existing failures)
  - Record baseline test results
  - **Estimated Time**: 10 minutes

### Phase 2: Code Analysis

- [ ] **Task 2.1**: Locate processConditionals Function
  - Open parser/html.go
  - Find processConditionals function (approximately line 400+)
  - Read and understand what it does
  - Find all calls to processConditionals
  - Note line numbers for all locations
  - **Estimated Time**: 15 minutes

- [ ] **Task 2.2**: Locate processLoops Function
  - In parser/html.go
  - Find processLoops function (approximately line 500+)
  - Read and understand what it does
  - Find all calls to processLoops
  - Note line numbers for all locations
  - **Estimated Time**: 15 minutes

- [ ] **Task 2.3**: Verify Block Parser Integration
  - Open parser/parser.go
  - Review AnyNodeParser function
  - Verify it calls BlockConditionalParser for {if}
  - Verify it calls BlockLoopParser for {for}
  - Confirm Element parser calls AnyNodeParser for children
  - **Estimated Time**: 20 minutes

### Phase 3: Implementation

- [ ] **Task 3.1**: Comment Out processConditionals
  - Add explanatory comment block above function
  - Include reference to this spec
  - Comment out entire processConditionals function
  - **Estimated Time**: 5 minutes

- [ ] **Task 3.2**: Comment Out processConditionals Calls
  - Find all calls to processConditionals in parser/html.go
  - Comment out each call
  - Add inline comment referencing spec
  - **Estimated Time**: 5 minutes

- [ ] **Task 3.3**: Comment Out processLoops
  - Add explanatory comment block above function
  - Include reference to this spec
  - Comment out entire processLoops function
  - **Estimated Time**: 5 minutes

- [ ] **Task 3.4**: Comment Out processLoops Calls
  - Find all calls to processLoops in parser/html.go
  - Comment out each call
  - Add inline comment referencing spec
  - **Estimated Time**: 5 minutes

### Phase 4: Testing

- [ ] **Task 4.1**: Test Parser Unit Tests
  - Run: `go test ./parser -v`
  - Verify all tests pass
  - Run: `go test ./parser -run TestConditionalBug -v`
  - Run: `go test ./parser -run TestNestedConditionalLoop -v`
  - **Expected**: TestNestedConditionalLoop should now PASS
  - **Estimated Time**: 10 minutes

- [ ] **Task 4.2**: Test Basic Conditionals Fix
  - Start development server: `go run cmd/server/main.go`
  - Visit http://localhost:3000
  - Check Basic Conditionals section (lines 30-39)
  - Verify only ONE message shows
  - Verify no literal {else if} or {else} text
  - Verify nested "Has been born" conditional works
  - **Estimated Time**: 10 minutes

- [ ] **Task 4.3**: Test Animals Loop Fix
  - In development server at http://localhost:3000
  - Check Animals Loop section (lines 61-69)
  - Count "likes" messages - should be 3 (dog, cat, bird)
  - Verify "Benjamin likes: dogs" appears
  - Verify "Benjamin likes: cats" appears
  - Verify "Benjamin likes: birds" appears
  - Verify correct greeting for each animal
  - **Estimated Time**: 10 minutes

- [ ] **Task 4.4**: Run Full Test Suite
  - Run: `go test ./... -v`
  - Verify no new test failures
  - Check tests in: parser, transformer, renderer, tests/alpine, tests/components
  - Compare to baseline test results from Task 1.3
  - **Estimated Time**: 15 minutes

- [ ] **Task 4.5**: Manual Edge Case Testing
  - Test simple conditional (no nesting)
  - Test conditional with siblings
  - Test loop with conditional
  - Test nested conditionals
  - Test nested loops
  - Test components in loops
  - Test all examples in examples/pages/ directory
  - **Estimated Time**: 20 minutes

### Phase 5: Documentation

- [ ] **Task 5.1**: Update KNOWN_ISSUES.md
  - Mark Animals Loop bug as RESOLVED
  - Add resolution date: 2025-10-06
  - Add reference to this spec
  - Note the fix approach
  - **Estimated Time**: 10 minutes

- [ ] **Task 5.2**: Update CLAUDE.md
  - Add Parser Architecture section (if not exists)
  - Document unified parsing approach
  - Note removal of post-processing
  - Explain why change was made
  - Link to this spec
  - **Estimated Time**: 15 minutes

- [ ] **Task 5.3**: Add Code Comments
  - Review all commented out code
  - Ensure all have clear explanations
  - Ensure all reference this spec
  - Add comments to AnyNodeParser noting it's the single entry point
  - **Estimated Time**: 10 minutes

### Phase 6: Validation & Completion

- [ ] **Task 6.1**: Final Visual Verification
  - Run development server one more time
  - Load http://localhost:3000
  - Take screenshots of fixed behavior
  - Compare to screenshots from Task 1.2
  - Document the visual improvement
  - **Estimated Time**: 10 minutes

- [ ] **Task 6.2**: Code Review Checklist
  - All processConditionals code commented out
  - All processLoops code commented out
  - All comments include spec reference
  - No dead code paths remaining
  - Consistent code style
  - No debug logging left in
  - **Estimated Time**: 10 minutes

- [ ] **Task 6.3**: Test Results Summary
  - Verify TestConditionalBug passes
  - Verify TestNestedConditionalLoop passes
  - Verify all parser tests pass
  - Verify full test suite passes
  - Verify Basic Conditionals fixed
  - Verify Animals Loop fixed
  - No regressions found
  - **Estimated Time**: 5 minutes

- [ ] **Task 6.4**: Commit Changes
  - Stage modified files (parser/html.go, KNOWN_ISSUES.md, CLAUDE.md)
  - Write clear commit message
  - Reference this spec in commit message
  - Commit changes
  - **Estimated Time**: 5 minutes

### Phase 7: Post-Implementation

- [ ] **Task 7.1**: Update Spec Status
  - Update this file with completion date
  - Mark spec as Completed in spec.md
  - Add any lessons learned
  - **Estimated Time**: 5 minutes

- [ ] **Task 7.2**: Consider Future Improvements
  - Should commented code be deleted in future release?
  - Are there other parsing paths to unify?
  - Should parser architecture be documented further?
  - Create follow-up specs if needed
  - **Estimated Time**: 15 minutes

## Task Summary

**Total Tasks**: 27
**Estimated Total Time**: 4.5 hours

**By Phase**:
- Phase 1 (Investigation): 55 minutes
- Phase 2 (Code Analysis): 50 minutes
- Phase 3 (Implementation): 20 minutes
- Phase 4 (Testing): 65 minutes
- Phase 5 (Documentation): 35 minutes
- Phase 6 (Validation): 30 minutes
- Phase 7 (Post-Implementation): 20 minutes

## Critical Path

The following tasks MUST be completed in order:

1. Task 1.1 → 1.2 → 1.3 (Understand problem before fixing)
2. Task 2.1 → 2.2 → 2.3 (Analyze before implementing)
3. Task 3.1 → 3.2 → 3.3 → 3.4 (Implementation sequence)
4. Task 4.1 → 4.2 → 4.3 → 4.4 (Test incrementally)
5. Task 6.3 → 6.4 (Verify before committing)

## Success Metrics

- [ ] Both bugs fixed (Basic Conditionals + Animals Loop)
- [ ] TestNestedConditionalLoop passes (currently fails)
- [ ] All existing tests continue to pass
- [ ] No new test failures introduced
- [ ] Development server renders correctly
- [ ] Documentation updated
- [ ] Code changes committed

## Notes

### Implementation Strategy

**Conservative Approach**: Comment out code rather than delete to preserve history and allow easy rollback if needed.

**Testing First**: Verify bugs exist before fixing, verify fixes work after implementation.

**Incremental Changes**: Make one change at a time, test after each change.

### Potential Issues

1. **Unknown Dependencies**: processConditionals/processLoops might be called from unexpected places
   - Mitigation: Search entire codebase for function calls before disabling

2. **Edge Cases**: Some template patterns might depend on post-processing behavior
   - Mitigation: Comprehensive manual testing of various template patterns

3. **Test Failures**: Existing tests might expect post-processing behavior
   - Mitigation: Review test expectations and update if they're testing wrong behavior

### Rollback Plan

If issues arise:
1. Uncommit changes: `git reset --soft HEAD~1`
2. Uncomment code in parser/html.go
3. Run tests to verify restoration
4. Document what went wrong
5. Create new spec addressing the issues found

### Definition of Done

This spec is complete when:
- [ ] All tasks marked complete
- [ ] Both bugs visually confirmed as fixed in dev server
- [ ] TestNestedConditionalLoop passes
- [ ] All existing tests pass
- [ ] Documentation updated
- [ ] Changes committed with clear message
- [ ] No regressions in any functionality
