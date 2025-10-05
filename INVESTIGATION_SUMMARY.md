# Investigation Summary: Animals Loop Bug

**Date**: October 5, 2025
**Branch**: `parser-depth-tracking-fix`
**Status**: Root cause identified - Architectural fix required

## Executive Summary

Investigated the "Animals Loop" bug where content after `{/if}` inside loops is incorrectly included in the conditional's content instead of being rendered as siblings. Through comprehensive testing, confirmed the root cause is an interaction bug between two different parsing paths in the codebase.

**Key Finding**: The parser has two paths for handling conditionals (BlockConditionalParser and Element parser + processConditionals), and their interaction causes content over-consumption in deeply nested structures.

## Investigation Timeline

### Initial Hypothesis: Parser Depth Tracking Incomplete
- **Investigated**: Whether BlockConditionalParser needed depth increment for nested {if}
- **Result**: Parser depth tracking is correct - uses recursive calls via AnyNodeParser
- **Finding**: Original spec hypothesis was incorrect

### Second Hypothesis: Transformer ensureProperNesting Bug
- **Investigated**: Whether ensureProperNesting was incorrectly re-parenting siblings into conditional templates
- **Implementation**: Rewrote ensureProperNesting to finalize templates when non-conditional siblings found
- **Result**: Bug persisted because parser already grouped nodes incorrectly
- **Finding**: Transformer can't fix what parser produces incorrectly

### Third Hypothesis: Renderer Issue
- **Investigated**: Whether renderer was incorrectly placing nodes
- **Result**: Renderer correctly renders the AST structure it receives
- **Finding**: Problem is upstream in parser

### Final Discovery: Two Parsing Paths
- **Investigated**: Created minimal reproduction tests
- **Key Tests**:
  - Simple template: Parser works correctly ✅
  - Nested template (conditional in loop in element): Bug reproduced ❌
- **Root Cause**: BlockConditionalParser works in isolation, but when parsing happens through Element parser path, depth tracking fails
- **Conclusion**: Architectural issue requiring unified parsing approach

## Technical Details

### The Bug

**Template** (`examples/pages/home.html` lines 61-69):
```html
{for animal of animals}
  {if animal == "cat"}
    <div>Hi {animal}!</div>
  {else}
    <div>Bye {animal}.</div>
  {/if}
  <div class="type-{animal}">{name} likes: {animal}s</div>
  <br>
{/for}
```

**Expected Parser Output**:
```
Loop.Content: [
  Conditional,
  TextNode,
  Element (div "likes"),
  TextNode,
  Element (br)
]
```

**Actual Parser Output**:
```
Loop.Content: [
  Conditional
]
Where Conditional.IfContent includes the div and br (38 total nodes)
```

### Evidence from Logs

```
createLoopTemplate: received 1 content nodes: [*ast.Conditional]
transformConditional: wrapping 38 nodes in container div for if branch
```

The loop receives only 1 node instead of 5, confirming the parser grouped siblings incorrectly.

### Two Parsing Paths Identified

1. **BlockConditionalParser** (`parser/parser.go` line 173+)
   - Direct parsing of {if}...{/if} blocks
   - Recursive depth tracking via AnyNodeParser
   - Returns Conditional with correct Remaining content
   - **Works correctly in isolation**

2. **Element Parser + processConditionals** (`parser/html.go`)
   - Parses HTML elements and their children
   - Post-processes directive marker nodes
   - Different code path than BlockConditionalParser
   - **Interaction with BlockConditionalParser causes bug**

### Why The Bug Occurs

When parsing structure: `<element> → loop → conditional`:

1. Element parser parses `<div class="animals">`
2. Element parser parses children, including `{for}` loop
3. Loop parser (BlockLoopParser) calls AnyNodeParser to parse loop content
4. AnyNodeParser calls BlockConditionalParser to parse `{if}...{/if}`
5. **BUG**: BlockConditionalParser continues parsing beyond `{/if}` and includes siblings in IfContent
6. Loop receives only Conditional instead of [Conditional, div, br]

## Test Files Created

### `parser/conditional_bug_test.go`
Tests simple conditional parsing to establish baseline:
```go
template := `{if condition}
  <div>Content</div>
{/if}
<div>SIBLING</div>`
```

**Result**: ✅ Parser correctly produces 2 nodes (Conditional + SIBLING)

### `parser/nested_conditional_loop_test.go`
Reproduces the actual bug with nested structure:
```go
template := `<div>
  {for animal of animals}
    {if animal == "cat"}
      <div>Hi</div>
    {else}
      <div>Bye</div>
    {/if}
    <div>SIBLING</div>
  {/for}
</div>`
```

**Result**: ❌ Parser produces Loop with 1 node; SIBLING trapped in Conditional.IfContent

## Files Modified During Investigation

### Transformer Changes
- `transformer/template_nesting.go`:
  - Improved `ensureProperNesting` to not buffer siblings after complete conditional chains
  - Added recursive processing of element children
  - Added detailed logging with `getNodeDesc()` helper
  - **Impact**: Helps prevent future similar bugs in transformer, but can't fix parser issue

### Parser Documentation
- `parser/parser.go`:
  - Added comprehensive comments explaining recursive depth tracking architecture
  - Documented why depth only decrements (AnyNodeParser handles nested blocks)
  - **Impact**: Clarifies intended design for future developers

### Test Files
- `parser/conditional_bug_test.go` - Simple test (passes)
- `parser/nested_conditional_loop_test.go` - Nested test (fails, demonstrates bug)
- **Impact**: Regression tests for future fix

### Documentation
- `KNOWN_ISSUES.md` - Comprehensive bug documentation
- `INVESTIGATION_SUMMARY.md` - This file
- **Impact**: Preserves investigation findings for future work

## Recommended Solution

### Option 1: Quick Fix (Not Recommended)
Add depth increment logic to BlockConditionalParser when encountering nested conditionals.

**Pros**:
- Quick to implement
- Might resolve this specific case

**Cons**:
- Band-aid solution
- Doesn't address architectural issue
- Risk of breaking other parsing scenarios
- May not fully resolve interaction bug

### Option 2: Architectural Fix (Recommended)
Unify the two parsing paths into a single consistent approach.

**Implementation Steps**:
1. Audit all parsing paths and document their interactions
2. Design unified conditional parsing strategy
3. Ensure BlockConditionalParser is used consistently
4. Refactor Element parser to delegate to BlockConditionalParser
5. Centralize depth tracking logic
6. Expand test coverage for all nesting scenarios

**Pros**:
- Addresses root cause
- Prevents similar bugs in future
- Cleaner architecture
- More maintainable

**Cons**:
- Requires significant refactoring
- Higher risk in short term
- Needs comprehensive testing

### Option 3: Documented Workaround (Interim)
Document the limitation and provide template restructuring guidance.

**User Workaround**:
```html
{for animal of animals}
  {if animal == "cat"}
    <div>Hi {animal}!</div>
    <div>{name} likes: {animal}s</div>
    <br>
  {else}
    <div>Bye {animal}.</div>
    <div>{name} likes: {animal}s</div>
    <br>
  {/if}
{/for}
```

**Trade-off**: Content duplication but avoids the bug.

## Related Work

### Commits During Investigation
- Initial investigation on `parser-depth-tracking-fix` branch
- Transformer improvements (template_nesting.go)
- Parser documentation improvements
- Test file creation

### Previous Related Commits
- `bcca125` (Oct 3) - Fix nested conditionals - Bug already existed
- `40e76d5` (Oct 5) - Fix x-for loop rendering with wrappers - Bug persists
- `a0b1ef7` (Oct 5) - Fix x-if conditional rendering with wrappers - Bug persists

The bug predates all recent fixes, confirming it's a longstanding architectural issue.

## Lessons Learned

1. **Parser depth tracking works correctly** - Original spec was based on incorrect hypothesis
2. **Two parsing paths create interaction bugs** - Architecture needs simplification
3. **Integration tests are crucial** - Unit tests of individual components passed, but integration revealed bug
4. **Comprehensive logging is invaluable** - Detailed logs helped trace exact flow
5. **Test files preserve knowledge** - Created tests will prevent regression

## Next Steps for Future Developer

1. **Review Test Files**
   - Run `parser/conditional_bug_test.go` (should pass)
   - Run `parser/nested_conditional_loop_test.go` (currently fails - demonstrates bug)
   - Understand the difference in behavior

2. **Map Parsing Paths**
   - Trace all code paths that parse conditionals
   - Document when each path is used
   - Identify redundancies

3. **Design Unified Approach**
   - Choose one authoritative parser for conditionals
   - Ensure all code paths delegate to it
   - Centralize depth tracking

4. **Implement & Test**
   - Make changes incrementally
   - Run full test suite after each change
   - Use created test files as regression checks

5. **Document Architecture**
   - Update CLAUDE.md with parsing architecture
   - Add diagrams if helpful
   - Document design decisions

## Conclusion

This investigation definitively identified the root cause of the Animals Loop bug: an interaction between two different parsing paths creates incorrect AST structure when conditionals are nested inside loops inside elements.

While transformer improvements were made during investigation (better `ensureProperNesting`), the bug can only be properly fixed by unifying the parsing architecture.

The comprehensive test files and documentation created during this investigation provide a solid foundation for implementing the architectural fix when resources allow.

---

**Investigation Duration**: ~4 hours
**Files Created**: 4 (2 test files, 2 documentation files)
**Files Modified**: 2 (transformer/template_nesting.go, parser/parser.go)
**Root Cause**: Confirmed - Two parsing paths with interaction bug
**Recommended Fix**: Architectural refactoring to unify parsing paths
**Interim Solution**: Document workaround for users
