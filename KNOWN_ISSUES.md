# Known Issues

This document tracks known bugs and limitations in the custom Go template engine.

## Animals Loop Bug: Content After {/if} Incorrectly Included in Conditional

**Status**: Unresolved - Root cause identified, requires architectural fix

**Severity**: High - Affects rendering of loops with conditionals followed by additional content

**Date Identified**: October 5, 2025

**Investigation Branch**: `parser-depth-tracking-fix`

### Description

When a loop contains a conditional (`{if}...{/if}`) followed by additional content, the content that appears after `{/if}` is incorrectly included inside the conditional's IfContent or ElseContent instead of being a separate sibling in the loop.

### Template Example

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

**Location**: `examples/pages/home.html` lines 61-69

### Expected Behavior

The loop should parse content into 5 nodes:
1. Conditional (with IfContent and ElseContent)
2. TextNode (whitespace)
3. Element (`<div class="type-{animal}">`)
4. TextNode (whitespace)
5. Element (`<br>`)

The div and br should render for ALL animals (dog, cat, bird).

### Actual Behavior

The parser produces a loop with only 1 node (the Conditional), and the div and br are incorrectly included in the Conditional's IfContent (38 total nodes including whitespace).

**Rendered HTML**:
```html
<template x-for="animal in animals"><div>
  <template x-if="animal == 'cat'"><div>
    <div>Hi cat!</div>
    <div :class="animal">Benjamin likes: cats</div>  <!-- BUG: Should be outside -->
    <br></br>                                          <!-- BUG: Should be outside -->
  </div></template>
  <template x-if="!(animal == 'cat')"><div>
    <div>Bye cat.</div>
  </div></template>
</div></template>
```

### Visual Impact

The "Benjamin likes: cats" message and br only appear for "cat", not for all animals:

**Current**:
- "Bye dog."
- "Hi cat!"
- "Benjamin likes: cats" ← Only shows for cat
- "Bye bird."

**Should be**:
- "Bye dog."
- "Benjamin likes: cats" ← Should show for dog
- "Hi cat!"
- "Benjamin likes: cats" ← Shows for cat
- "Bye bird."
- "Benjamin likes: cats" ← Should show for bird

### Root Cause Analysis

#### Two Parsing Paths

The codebase has **two different parsing paths** for conditionals:

1. **BlockConditionalParser** (`parser/parser.go` line 173+)
   - Used for top-level conditional parsing
   - Uses depth tracking with recursion
   - Returns Conditional node with Remaining content

2. **Element Parser + processConditionals** (`parser/html.go`)
   - Used for parsing conditionals inside HTML elements
   - Post-processes directive marker nodes
   - Different code path than BlockConditionalParser

#### The Bug

When conditionals are **nested inside loops that are inside HTML elements**, the interaction between these two parsing paths causes the parser to continue consuming content beyond `{/if}`.

**Evidence from logs**:
```
createLoopTemplate: received 1 content nodes: [*ast.Conditional]
transformConditional: wrapping 38 nodes in container div for if branch
```

The loop only receives the Conditional (instead of 5 nodes), and the Conditional has 38 nodes in its IfContent (including the div and br that should be siblings).

#### Why This Happens

The parser correctly handles:
- ✅ Simple conditionals (no nesting)
- ✅ Nested conditionals at top level
- ❌ Conditionals inside loops inside elements (complex nesting)

The depth tracking in `BlockConditionalParser` works when called directly, but when parsing happens through the Element parser path, the depth tracking doesn't correctly handle the boundary between the conditional and subsequent siblings.

### Investigation Files

**Test Files Created**:
- `parser/conditional_bug_test.go` - Demonstrates simple conditional parsing works
- `parser/nested_conditional_loop_test.go` - Reproduces the bug with nested structure

**Test Results**:
- Simple template: ✅ Parser produces 3 nodes (Conditional + SIBLING div + br)
- Nested template: ❌ Parser produces 1 node (Conditional with SIBLING content trapped inside)

**Log Evidence**:
```
Simple test:
  Loop Content: 3 nodes [Conditional, Element (SIBLING), Element (br)]

Nested test:
  Loop Content: 1 node [Conditional]
  Conditional IfContent: 26 nodes (includes SIBLING div and br)
```

### Files Involved

**Parser**:
- `parser/parser.go` - BlockConditionalParser (line 173-290)
- `parser/html.go` - Element parser with directive processing

**Transformer**:
- `transformer/conditionals.go` - transformConditional creates wrapper divs
- `transformer/loops.go` - createLoopTemplate receives incorrect node count
- `transformer/template_nesting.go` - ensureProperNesting (improved but can't fix parser issue)

**Examples**:
- `examples/pages/home.html` - Animals Loop section (lines 59-70)

### Previous Investigation Attempts

1. **Initial hypothesis**: Parser depth tracking incomplete ❌
   - Parser depth tracking is correct for top-level parsing
   - Issue is in interaction between two parsing paths

2. **Second hypothesis**: Transformer ensureProperNesting bug ❌
   - Improved ensureProperNesting to not buffer siblings
   - Bug persists because parser already grouped nodes incorrectly

3. **Third hypothesis**: Renderer issue ❌
   - Renderer correctly renders the AST it receives
   - Problem is upstream in parser

4. **Final finding**: Two parsing paths with interaction bug ✅
   - BlockConditionalParser works correctly in isolation
   - Element parser + processConditionals creates different behavior
   - Interaction causes over-consumption of content

### Recommended Fix

**Option 1: Quick Fix (Band-aid)**
- Add depth increment when BlockConditionalParser encounters nested conditionals
- May not fully address the interaction bug
- Risk of breaking other cases

**Option 2: Proper Fix (Architectural)**
- Unify the two parsing paths into a single consistent approach
- Ensure all conditionals use the same depth tracking logic
- Separate responsibilities clearly between parser phases
- This is the recommended approach but requires significant refactoring

**Option 3: Workaround**
- Document the limitation
- Advise users to structure templates differently
- Not ideal but avoids risky changes

### Workarounds

**For Users**:
Avoid placing content after `{/if}` when the conditional is inside a loop that's inside an HTML element.

**Restructure like this**:
```html
{for animal of animals}
  <div>
    {if animal == "cat"}
      <div>Hi {animal}!</div>
      <div class="type-{animal}">{name} likes: {animal}s</div>
      <br>
    {else}
      <div>Bye {animal}.</div>
      <div class="type-{animal}">{name} likes: {animal}s</div>
      <br>
    {/if}
  </div>
{/for}
```

**Trade-off**: Content duplication in each branch.

### Related Commits

- `c01acaa` - Update More Loop Examples (bug still present)
- `a0b1ef7` - Fix Alpine.js x-if conditional rendering with wrappers (bug still present)
- `40e76d5` - Fix Alpine.js x-for loop rendering with wrappers (bug still present)
- `bcca125` - Fix nested conditionals (October 3) - Bug exists here too

### Next Steps for Future Fix

1. **Architecture Review**
   - Map out all parsing paths and their interactions
   - Identify why two paths exist and if both are needed
   - Design unified approach

2. **Implement Unified Parsing**
   - Ensure BlockConditionalParser is used consistently
   - Remove or refactor Element parser directive processing
   - Centralize depth tracking logic

3. **Comprehensive Testing**
   - Expand test suite to cover all nesting scenarios
   - Test: conditionals in loops in elements in conditionals
   - Ensure edge cases are covered

4. **Regression Prevention**
   - Keep test files created during investigation
   - Add CI checks for this specific pattern
   - Document parsing architecture

---

*Last Updated: October 5, 2025*
*Investigation by: Claude Code*
*Investigation Branch: parser-depth-tracking-fix*
*Test Files: parser/conditional_bug_test.go, parser/nested_conditional_loop_test.go*
