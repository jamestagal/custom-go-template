# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-06-parser-unification/spec.md

> Created: 2025-10-06
> Version: 1.0.0

## Technical Requirements

### Functional Requirements

1. **Single Parsing Path for Directives**
   - All `{if}...{/if}` parsing must go through `BlockConditionalParser`
   - All `{for}...{/for}` parsing must go through `BlockLoopParser`
   - No duplicate or alternative parsing logic in Element parser

2. **Correct AST Structure**
   - Siblings after `{/if}` must be separate nodes, not in IfContent/ElseContent
   - Siblings after `{/for}` must be outside loop content
   - Loop with conditional + siblings should produce: `[Conditional, TextNode, Element, TextNode, Element]`

3. **Backward Compatibility**
   - All existing templates must continue to work
   - No breaking changes to AST node types
   - No changes to transformer or renderer required

4. **Test Coverage**
   - All existing parser tests must pass
   - Regression tests must pass after fix
   - Integration with development server must work

### Non-Functional Requirements

1. **Maintainability**
   - Clear code comments explaining architectural decision
   - No dead code (remove or clearly mark disabled code)
   - Consistent coding style

2. **Performance**
   - No significant performance degradation
   - Parsing speed should remain comparable

3. **Documentation**
   - Architecture documented in code comments
   - `KNOWN_ISSUES.md` updated
   - `CLAUDE.md` updated with parsing architecture

## Current Architecture

### Two Parsing Paths (The Problem)

**Path 1: Block Parsers** (`parser/parser.go`)
```go
// Lines 173-290: BlockConditionalParser
func (p *Parser) BlockConditionalParser(...) (*ast.Conditional, string, error)

// Lines 292-380: BlockLoopParser
func (p *Parser) BlockLoopParser(...) (*ast.Loop, string, error)
```

- Used for top-level directive parsing
- Uses recursive depth tracking via `AnyNodeParser`
- Returns directive node + remaining content
- **Works correctly in isolation**

**Path 2: Element Parser + Post-Processing** (`parser/html.go`)
```go
// Line ~200+: Element parser
func (p *Parser) Element(content string) (*ast.Element, string, error)

// Line ~400+: processConditionals
func processConditionals(nodes []ast.Node) []ast.Node

// Line ~500+: processLoops
func processLoops(nodes []ast.Node) []ast.Node
```

- Parses HTML elements and children
- Post-processes directive marker nodes
- Different code path than Block parsers
- **Interaction with Block parsers causes bugs**

### The Bug Mechanism

When parsing: `<element> → {for} → {if} → content after {/if}`

1. Element parser parses `<div class="animals">`
2. Element parser calls `AnyNodeParser` to parse children
3. `AnyNodeParser` calls `BlockLoopParser` for `{for}`
4. `BlockLoopParser` calls `AnyNodeParser` for loop content
5. `AnyNodeParser` calls `BlockConditionalParser` for `{if}`
6. **BUG**: `BlockConditionalParser` continues parsing beyond `{/if}` and includes siblings in IfContent
7. Loop receives only `Conditional` instead of `[Conditional, div, br]`

**Evidence**:
```
createLoopTemplate: received 1 content nodes: [*ast.Conditional]
transformConditional: wrapping 38 nodes in container div for if branch
```

The loop should receive 5 nodes, not 1. The Conditional should have ~5 nodes in IfContent, not 38.

## Target Architecture

### Single Parsing Path (The Solution)

**All directive parsing goes through Block parsers:**

```
Template → Element Parser → AnyNodeParser → {
  BlockConditionalParser for {if}
  BlockLoopParser for {for}
  ExpressionParser for {var}
  ...
}
```

**Element parser responsibilities:**
- Parse HTML opening/closing tags
- Parse attributes
- Delegate child parsing to `AnyNodeParser`
- **DO NOT** post-process directives

**Block parser responsibilities:**
- Parse directive syntax (`{if}`, `{for}`, etc.)
- Track depth for nesting
- Determine content boundaries
- Return directive node + remaining content

## Implementation Plan

### Step 1: Identify Directive Post-Processing Code

**File**: `parser/html.go`

Look for these functions:
- `processConditionals(nodes []ast.Node) []ast.Node`
- `processLoops(nodes []ast.Node) []ast.Node`
- Any calls to these functions in Element parser

### Step 2: Comment Out Post-Processing

**Approach**: Comment out rather than delete to preserve history

```go
// DISABLED: Using BlockConditionalParser exclusively for all conditional parsing
// This post-processing created interaction bugs with nested structures
// See: .agent-os/specs/2025-10-06-parser-unification/spec.md
//
// func processConditionals(nodes []ast.Node) []ast.Node {
//   ... existing code ...
// }

// DISABLED: Using BlockLoopParser exclusively for all loop parsing
// This post-processing created interaction bugs with nested structures
// See: .agent-os/specs/2025-10-06-parser-unification/spec.md
//
// func processLoops(nodes []ast.Node) []ast.Node {
//   ... existing code ...
// }
```

**Also comment out calls to these functions:**
```go
// children = processConditionals(children)
// children = processLoops(children)
```

### Step 3: Verify AnyNodeParser Integration

**File**: `parser/parser.go`

Ensure `AnyNodeParser` is being called correctly:

```go
func (p *Parser) AnyNodeParser(content string, depth int) (ast.Node, string, error) {
  // Check for conditionals
  if strings.HasPrefix(content, "{if ") {
    return p.BlockConditionalParser(content, depth)
  }

  // Check for loops
  if strings.HasPrefix(content, "{for ") {
    return p.BlockLoopParser(content, depth)
  }

  // ... other node types ...
}
```

**Verify this is called from**:
- Element parser when parsing children
- Loop parser when parsing loop content
- Conditional parser when parsing branch content

### Step 4: Test Against Basic Conditionals Bug

**File**: `examples/pages/home.html` lines 30-39

Run development server:
```bash
go run cmd/server/main.go
# Visit http://localhost:3000
```

**Expected Behavior**:
- Only one message shows based on name.length
- No literal `{else if}` or `{else}` text

**Current Bug**: Multiple branches may render or show literal directive text

**Verification**:
```bash
# Look for this in rendered HTML:
# Should see only ONE of:
# - "Benjamin is a long name" + "Has been born"
# - "Benjamin is medium"
# - "Benjamin is a short name"
```

### Step 5: Test Against Animals Loop Bug

**File**: `examples/pages/home.html` lines 61-69

**Expected Behavior**:
```html
Bye dog.
Benjamin likes: dogs
<br>
Hi cat!
Benjamin likes: cats
<br>
Bye bird.
Benjamin likes: birds
<br>
```

**Current Bug**:
```html
Bye dog.
Hi cat!
Benjamin likes: cats  ← Only shows for cat
<br>
Bye bird.
```

**Verification**:
```bash
# Count occurrences of "likes" in rendered output
# Should be 3 (one for each animal)
# Currently only 1 (for cat)
```

### Step 6: Run Parser Tests

**Test Files**:
```bash
# All parser tests
go test ./parser -v

# Specific regression tests
go test ./parser -run TestConditionalBug -v
go test ./parser -run TestNestedConditionalLoop -v
```

**Expected Results**:
- `TestConditionalBug` - Should pass (already passes)
- `TestNestedConditionalLoop` - Should pass (currently fails, will pass after fix)

**Key assertion in TestNestedConditionalLoop**:
```go
// Loop should have 5 nodes: Conditional, TextNode, Element, TextNode, Element
// Currently has 1 node: Conditional (with siblings trapped inside)
if len(loop.Content) != 5 {
  t.Errorf("Expected 5 nodes in loop content, got %d", len(loop.Content))
}
```

### Step 7: Run Full Test Suite

```bash
# All tests
go test ./... -v

# Key test directories
go test ./tests/alpine -v
go test ./tests/components -v
go test ./transformer -v
go test ./renderer -v
```

**Watch for**:
- Any new test failures
- Any changes in test output
- Any regressions in existing functionality

### Step 8: Update Documentation

**File**: `KNOWN_ISSUES.md`

Mark bugs as resolved:
```markdown
## Animals Loop Bug: Content After {/if} Incorrectly Included in Conditional

**Status**: ✅ RESOLVED - Fixed in 2025-10-06
**Resolution**: Unified parser architecture (removed Element parser directive post-processing)
**Spec**: @.agent-os/specs/2025-10-06-parser-unification/spec.md
```

**File**: `CLAUDE.md`

Add architecture note in "Architecture" section:
```markdown
### Parser Architecture (Updated 2025-10-06)

All directive parsing (`{if}`, `{for}`, etc.) goes through dedicated Block parsers:
- `BlockConditionalParser` - Handles all `{if}...{/if}` parsing
- `BlockLoopParser` - Handles all `{for}...{/for}` parsing

The Element parser delegates to these via `AnyNodeParser` but does NOT post-process
directive markers. This unified approach prevents interaction bugs in nested structures.

**Previous Architecture** (before 2025-10-06):
The Element parser had post-processing functions (`processConditionals`, `processLoops`)
that created a second parsing path. This caused content after `{/if}` to be incorrectly
consumed into conditional content when nested inside loops inside elements.

**See**: @.agent-os/specs/2025-10-06-parser-unification/spec.md
```

## Files to Modify

### Primary Changes

1. **`parser/html.go`**
   - Comment out `processConditionals()` function
   - Comment out `processLoops()` function
   - Comment out calls to these functions
   - Add comments explaining the change

### Documentation Updates

2. **`KNOWN_ISSUES.md`**
   - Mark Animals Loop bug as resolved
   - Mark Basic Conditionals bug as resolved
   - Link to this spec

3. **`CLAUDE.md`**
   - Add parser architecture explanation
   - Note the unified parsing approach
   - Link to this spec

## Files to Test

### Regression Tests

1. **`parser/conditional_bug_test.go`**
   - Tests simple conditional parsing
   - Should pass before and after

2. **`parser/nested_conditional_loop_test.go`**
   - Tests conditional inside loop inside element
   - Currently fails, should pass after fix

### Integration Tests

3. **`examples/pages/home.html`**
   - Visual test via development server
   - Lines 30-39: Basic Conditionals
   - Lines 61-69: Animals Loop
   - Both should render correctly

### Existing Test Suites

4. **`parser/*_test.go`**
   - All parser unit tests
   - Should continue to pass

5. **`tests/alpine/*_test.go`**
   - Alpine.js integration tests
   - Should continue to pass

6. **`tests/components/*_test.go`**
   - Component tests
   - Should continue to pass

7. **`transformer/*_test.go`**
   - Transformer tests
   - Should continue to pass

## Success Criteria

### Bug Fixes Verified

1. **Basic Conditionals** (`home.html` lines 30-39)
   - ✅ Only one message shows based on name.length
   - ✅ No literal `{else if}` or `{else}` text in output
   - ✅ Nested `{if age > 1}` inside first branch works

2. **Animals Loop** (`home.html` lines 61-69)
   - ✅ "likes" message appears 3 times (for dog, cat, bird)
   - ✅ Loop AST has 5 content nodes (not 1)
   - ✅ Conditional IfContent has ~5 nodes (not 38)

### Tests Passing

3. **Parser Tests**
   - ✅ `parser/conditional_bug_test.go` passes
   - ✅ `parser/nested_conditional_loop_test.go` passes
   - ✅ All other parser tests pass

4. **Integration Tests**
   - ✅ `tests/alpine/*` tests pass
   - ✅ `tests/components/*` tests pass
   - ✅ `transformer/*` tests pass

### Code Quality

5. **Clean Code**
   - ✅ Commented out code has clear explanation
   - ✅ Spec linked in comments
   - ✅ No dead code paths
   - ✅ Consistent style

6. **Documentation**
   - ✅ `KNOWN_ISSUES.md` updated
   - ✅ `CLAUDE.md` updated
   - ✅ Code comments added

### No Regressions

7. **Existing Functionality**
   - ✅ No new test failures
   - ✅ No rendering regressions
   - ✅ Development server works correctly

## External Dependencies

**None** - This is an internal refactoring that doesn't add new dependencies.

**Existing Dependencies** (unchanged):
- Go standard library
- Alpine.js (for rendering, not parsing)
- Development server dependencies

## Testing Strategy

### Unit Testing

**Parser Tests**:
```bash
go test ./parser -v -run TestConditionalBug
go test ./parser -v -run TestNestedConditionalLoop
go test ./parser -v
```

**Expected Output**:
```
TestConditionalBug: PASS
TestNestedConditionalLoop: PASS (currently fails)
```

### Integration Testing

**Development Server**:
```bash
go run cmd/server/main.go
# Visit http://localhost:3000
# Manually verify:
# 1. Basic Conditionals section
# 2. Animals Loop section
```

**Visual Checklist**:
- [ ] Only one conditional message shows
- [ ] "Benjamin likes: dogs" appears
- [ ] "Benjamin likes: cats" appears
- [ ] "Benjamin likes: birds" appears
- [ ] Total of 3 "likes" messages (not 1)

### Regression Testing

**Full Test Suite**:
```bash
go test ./... -v
```

**Watch for**:
- New failures in any package
- Changes to test output
- Unexpected behavior

### Manual Testing

**Test Cases**:

1. **Simple Conditional**
   ```html
   {if condition}
     <div>Content</div>
   {/if}
   ```
   - Should work as before

2. **Conditional with Siblings**
   ```html
   {if condition}
     <div>Content</div>
   {/if}
   <div>SIBLING</div>
   ```
   - SIBLING should be separate node

3. **Loop with Conditional**
   ```html
   {for item of items}
     {if item == "special"}
       <div>Special</div>
     {/if}
     <div>Always shows</div>
   {/for}
   ```
   - "Always shows" should appear for all items

4. **Nested Conditionals**
   ```html
   {if outer}
     <div>Outer content</div>
     {if inner}
       <div>Inner content</div>
     {/if}
   {/if}
   ```
   - Should work as before

## Risk Assessment

### Low Risk

- Commenting out code (not deleting) allows easy rollback
- Block parsers already work correctly in isolation
- Comprehensive test suite catches regressions

### Medium Risk

- Changes to core parsing logic
- Potential for unexpected edge cases
- Requires thorough testing

### Mitigation Strategies

1. **Incremental Testing**
   - Test after commenting out processConditionals
   - Test after commenting out processLoops
   - Test full suite at each step

2. **Git Safety**
   - Make changes in feature branch
   - Commit after each successful test pass
   - Easy rollback if issues found

3. **Comprehensive Validation**
   - Run full test suite
   - Manual testing in dev server
   - Check multiple template examples

4. **Documentation**
   - Clear comments explain changes
   - Spec documents rationale
   - Easy for future developers to understand

## Implementation Checklist

- [ ] Read investigation documents (KNOWN_ISSUES.md, INVESTIGATION_SUMMARY.md)
- [ ] Locate `processConditionals` in parser/html.go
- [ ] Locate `processLoops` in parser/html.go
- [ ] Comment out `processConditionals` function with explanation
- [ ] Comment out `processLoops` function with explanation
- [ ] Comment out calls to these functions
- [ ] Run `parser/conditional_bug_test.go`
- [ ] Run `parser/nested_conditional_loop_test.go`
- [ ] Run full parser test suite
- [ ] Start development server
- [ ] Check Basic Conditionals (lines 30-39)
- [ ] Check Animals Loop (lines 61-69)
- [ ] Count "likes" messages (should be 3)
- [ ] Run full test suite `go test ./... -v`
- [ ] Update KNOWN_ISSUES.md
- [ ] Update CLAUDE.md
- [ ] Commit changes with clear message

## Notes

### Why Comment Out Instead of Delete?

Preserving the original code (commented out) provides:
- History of what was tried
- Reference for understanding the bug
- Easy rollback if needed
- Documentation of architectural evolution

### Why This Fix Works

The Block parsers (`BlockConditionalParser`, `BlockLoopParser`) already implement correct depth tracking and content boundary detection. The bug only occurs when their output is post-processed by Element parser functions.

By removing the post-processing, we allow the Block parsers to work as designed, eliminating the interaction bug.

### Future Considerations

If this fix works, consider:
- Deleting commented code after a few releases
- Adding more parser architecture documentation
- Expanding test coverage for nested structures
- Performance profiling to ensure no degradation
