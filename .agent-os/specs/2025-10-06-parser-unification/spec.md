# Spec Requirements Document

> Spec: Parser Architecture Unification
> Created: 2025-10-06
> Status: Planning

## Overview

Unify the parser architecture to eliminate the two-parsing-paths bug that causes content after `{/if}` and `{/for}` directives to be incorrectly consumed into the conditional or loop content instead of being parsed as siblings.

**Current Problem**: The codebase has two different parsing paths for directives:
1. **BlockConditionalParser / BlockLoopParser** (`parser/parser.go`) - Direct block parsing with recursive depth tracking
2. **Element Parser + processConditionals/processLoops** (`parser/html.go`) - Post-processing of directive markers within HTML elements

When these two paths interact (e.g., conditionals inside loops inside HTML elements), the parser incorrectly groups sibling content into the directive's content nodes, creating malformed AST structures.

**Solution**: Remove the directive post-processing from the Element parser and rely exclusively on BlockConditionalParser and BlockLoopParser for all directive parsing, ensuring a single consistent parsing path throughout the codebase.

## User Stories

### Story 1: Template Developers Can Use Directives Without Content Consumption Bugs
**As a** template developer
**I want** to place content after `{/if}` and `{/for}` directives inside HTML elements
**So that** the content renders as siblings rather than being incorrectly consumed into the directive's content

**Acceptance Criteria**:
- Content after `{/if}` is parsed as a sibling to the conditional, not included in IfContent or ElseContent
- Content after `{/for}` is parsed outside the loop, not as the last loop iteration's content
- The parser produces the correct number of sibling nodes in the AST
- Nested structures (conditionals in loops in elements) parse correctly

### Story 2: Basic Conditionals Render All Branches Correctly
**As a** template developer
**I want** to use `{if}...{else if}...{else}` conditionals
**So that** only the matching branch renders without `{else if}` appearing as literal text

**Current Bug** (lines 30-39 in `examples/pages/home.html`):
```html
{if name.length > 3}
  <div id="praise">{name} is a long name</div>
  {if age > 1}
    <div>Has been born</div>
  {/if}
{else if name.length == 2}
  <div id="praise">{name} is medium</div>
{else}
  <div id="praise">{name} is a short name</div>
{/if}
```

**Current Behavior**: The `{else if}` and `{else}` branches may render as literal text instead of being parsed as conditional branches.

**Expected Behavior**: Only one message shows based on name.length:
- name.length > 3: "Benjamin is a long name" + "Has been born"
- name.length == 2: "Benjamin is medium"
- Otherwise: "Benjamin is a short name"

**Acceptance Criteria**:
- Only the matching conditional branch renders
- No literal `{else if}` or `{else}` text appears in output
- Nested conditionals within branches work correctly
- All conditional branches are represented correctly in the AST

### Story 3: Animals Loop Renders Content for All Iterations
**As a** template developer
**I want** to place content after a conditional inside a loop
**So that** the content renders for every loop iteration, not just items matching the conditional

**Current Bug** (lines 61-69 in `examples/pages/home.html`):
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

**Current Behavior**:
- "Bye dog."
- "Hi cat!"
- "Benjamin likes: cats" ← Only shows for cat
- "Bye bird."

**Expected Behavior**:
- "Bye dog."
- "Benjamin likes: dogs" ← Should show for dog
- "Hi cat!"
- "Benjamin likes: cats" ← Shows for cat
- "Bye bird."
- "Benjamin likes: birds" ← Should show for bird

**Acceptance Criteria**:
- The "likes" message appears 3 times (once for each animal: dog, cat, bird)
- The `<div>` and `<br>` elements are siblings to the conditional, not trapped in IfContent
- The loop AST contains 5 nodes: Conditional, TextNode, Element (div), TextNode, Element (br)
- All three animals render their respective "likes" messages

## Spec Scope

### In Scope

1. **Remove Directive Post-Processing from Element Parser**
   - Comment out or remove `processConditionals()` in `parser/html.go`
   - Comment out or remove `processLoops()` in `parser/html.go`
   - Ensure Element parser doesn't duplicate directive parsing logic

2. **Ensure Block Parsers Handle All Directive Parsing**
   - Verify `BlockConditionalParser` is called correctly via `AnyNodeParser`
   - Verify `BlockLoopParser` is called correctly via `AnyNodeParser`
   - Confirm depth tracking works correctly in nested scenarios
   - Ensure proper content boundary detection for `{/if}` and `{/for}`

3. **Fix Basic Conditionals Bug**
   - Diagnose why `{else if}` and `{else}` branches render as literal text
   - Ensure all conditional branches are parsed correctly
   - Verify nested conditionals within conditional branches work

4. **Fix Animals Loop Bug**
   - Ensure content after `{/if}` inside `{for}` is parsed as sibling
   - Verify loop AST contains correct number of content nodes
   - Confirm transformer receives correct AST structure

5. **Testing & Validation**
   - Run existing test suite to ensure no regressions
   - Verify `parser/conditional_bug_test.go` passes
   - Verify `parser/nested_conditional_loop_test.go` passes (currently fails)
   - Test against `examples/pages/home.html` for visual confirmation
   - Confirm both bugs are fixed in development server

6. **Documentation**
   - Update `KNOWN_ISSUES.md` to mark bugs as resolved
   - Document the unified parsing architecture in `CLAUDE.md`
   - Add comments explaining why directive post-processing was removed

## Out of Scope

### Not Included in This Spec

1. **Performance Optimization**
   - This spec focuses on correctness, not speed
   - Performance improvements can be addressed in future specs

2. **New Parser Features**
   - No new directives or syntax
   - No new parsing capabilities beyond fixing existing bugs

3. **AST Node Type Changes**
   - No modifications to AST node structures
   - No new node types

4. **Transformer or Renderer Changes**
   - Changes only to parser architecture
   - Transformer and renderer work correctly with proper AST

5. **Alternative Directive Syntax**
   - Keep existing `{if}`, `{for}`, `{/if}`, `{/for}` syntax
   - No syntax changes

6. **Error Message Improvements**
   - Focus on fixing bugs, not enhancing error reporting
   - Error message improvements can be future work

7. **Additional Test Coverage**
   - Write tests only for the two specific bugs
   - Comprehensive test expansion is out of scope

## Expected Deliverable

### Success Criteria

1. **Both Bugs Fixed**
   - Basic Conditionals: Only one message shows (no literal `{else if}` text)
   - Animals Loop: "likes" message appears 3 times (for dog, cat, bird)

2. **All Existing Tests Pass**
   - No regressions in parser tests
   - No regressions in transformer tests
   - No regressions in integration tests

3. **New Regression Tests Pass**
   - `parser/conditional_bug_test.go` passes
   - `parser/nested_conditional_loop_test.go` passes (currently fails)

4. **Clean Architecture**
   - Single parsing path for all directives
   - No duplicate parsing logic
   - Clear separation of concerns

5. **Working Development Server**
   - `examples/pages/home.html` renders correctly
   - Visual inspection confirms both bugs fixed
   - No console errors or warnings

### Deliverables

1. **Modified Files**
   - `parser/html.go` - Remove directive post-processing
   - Any other parser files requiring updates

2. **Test Results**
   - All parser tests passing
   - Regression tests passing
   - Development server visual confirmation

3. **Documentation**
   - Updated `KNOWN_ISSUES.md` marking bugs as resolved
   - Updated `CLAUDE.md` with architecture notes
   - Code comments explaining changes

### Definition of Done

- [ ] `processConditionals()` and `processLoops()` removed or disabled in Element parser
- [ ] BlockConditionalParser and BlockLoopParser handle all directive parsing
- [ ] Basic Conditionals bug fixed (verified in dev server)
- [ ] Animals Loop bug fixed (verified in dev server)
- [ ] All existing tests pass
- [ ] `parser/conditional_bug_test.go` passes
- [ ] `parser/nested_conditional_loop_test.go` passes
- [ ] No new test failures introduced
- [ ] Code changes documented with comments
- [ ] `KNOWN_ISSUES.md` updated to reflect resolution
- [ ] Changes committed with clear commit message

## Spec Documentation

- Tasks: @.agent-os/specs/2025-10-06-parser-unification/tasks.md
- Technical Specification: @.agent-os/specs/2025-10-06-parser-unification/sub-specs/technical-spec.md
- Investigation Background: @KNOWN_ISSUES.md
- Investigation Summary: @INVESTIGATION_SUMMARY.md
- Test Files: @parser/conditional_bug_test.go, @parser/nested_conditional_loop_test.go
