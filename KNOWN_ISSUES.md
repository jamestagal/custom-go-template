# Known Issues

This document tracks known bugs and limitations in the custom Go template engine.

## ✅ RESOLVED: Animals Loop Bug: Content After {/if} Incorrectly Included in Conditional

**Status**: RESOLVED (October 6, 2025) - Fixed by parser unification in commit 437da77

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

### Fix Implemented (October 6, 2025)

**Solution: Parser Unification (Architectural Fix)**
- ✅ Unified the two parsing paths into single consistent approach
- ✅ Removed processDirectiveNodes post-processing in parser/html.go
- ✅ Changed parseChildren to use AnyNodeParser directly (lines 289, 309-314)
- ✅ Marked parseChildNode and processDirective* functions as DEPRECATED
- ✅ Updated CLAUDE.md with parser architecture documentation
- ✅ All tests passing, both bugs fixed in browser

**Verification**:
- ✅ Basic Conditionals render only ONE branch (no literal {else if} text)
- ✅ Animals Loop shows "Benjamin likes:" for ALL 3 animals (dog, cat, bird)
- ✅ parser/conditional_bug_test.go - PASS
- ✅ parser/nested_conditional_loop_test.go - PASS

**Implementation Details**: See commit 437da77 and .agent-os/specs/2025-10-06-parser-unification/

### ~~Workarounds~~ (No Longer Needed - Bug Fixed!)

**Previous Workaround** (no longer necessary):
~~Avoid placing content after `{/if}` when the conditional is inside a loop that's inside an HTML element.~~

**✅ You can now use the original, natural syntax**:
```html
{for animal of animals}
  {if animal == "cat"}
    <div>Hi {animal}!</div>
  {else}
    <div>Bye {animal}.</div>
  {/if}
  <div class="type-{animal}">{name} likes: {animal}s</div>  <!-- ✅ This now works! -->
  <br>
{/for}
```

No content duplication needed!

### Related Commits

- `c01acaa` - Update More Loop Examples (bug still present)
- `a0b1ef7` - Fix Alpine.js x-if conditional rendering with wrappers (bug still present)
- `40e76d5` - Fix Alpine.js x-for loop rendering with wrappers (bug still present)
- `bcca125` - Fix nested conditionals (October 3) - Bug exists here too

### ~~Next Steps for Future Fix~~ (COMPLETED!)

All steps completed in commit 437da77:

1. ✅ **Architecture Review**
   - ✅ Mapped all parsing paths and their interactions
   - ✅ Identified dual-path issue
   - ✅ Designed unified approach

2. ✅ **Implemented Unified Parsing**
   - ✅ BlockConditionalParser/BlockLoopParser used consistently via AnyNodeParser
   - ✅ Removed Element parser directive post-processing
   - ✅ Centralized depth tracking in BlockConditionalParser recursive logic

3. ✅ **Comprehensive Testing**
   - ✅ Test suite covers nesting scenarios
   - ✅ Tests: conditionals in loops in elements
   - ✅ Edge cases covered

4. ✅ **Regression Prevention**
   - ✅ Test files preserved: parser/conditional_bug_test.go, parser/nested_conditional_loop_test.go
   - ✅ Parsing architecture documented in CLAUDE.md
   - ✅ Spec documented: .agent-os/specs/2025-10-06-parser-unification/

---

*Last Updated: October 6, 2025*
*Investigation: October 5, 2025 (Claude Code)*
*Fix Implemented: October 6, 2025 (commit 437da77)*
*Investigation Branch: parser-depth-tracking-fix*
*Test Files: parser/conditional_bug_test.go, parser/nested_conditional_loop_test.go*
*Fix Spec: .agent-os/specs/2025-10-06-parser-unification/*

**Status: RESOLVED ✅**
