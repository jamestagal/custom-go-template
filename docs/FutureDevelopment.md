# Future Development Tasks

This document tracks enhancement tasks for future development of the Custom Go Template engine.

## Component Prop Scoping

### Current Limitations

The current implementation uses a global x-data scope workaround which has several limitations:

1. **No component isolation**: All props must exist in the global scope
2. **Naming conflicts**: Component prop names can clash with page variables
3. **No prop validation**: Can't enforce required props or default values properly
4. **Performance**: Can't optimize component re-rendering based on prop changes

### Proposed Enhancement

A proper component prop scoping solution would involve:

- Each component instance gets wrapped with its own `<div x-data="{ ...props }">`
- Prop expressions are evaluated in parent scope and serialized to component scope
- Component default values merge with passed props
- Scoped styles and scripts per component instance

### Tasks

1. **Create spec document for proper component prop scoping with isolation**
   - Document the architectural changes needed
   - Define the transformation pipeline modifications
   - Specify the scope inheritance model
   - Define prop serialization strategy

2. **Implement component prop scoping: Each component instance gets wrapped with own x-data scope**
   - Modify transformer to wrap component instances with isolated x-data
   - Ensure proper Alpine.js initialization order
   - Handle nested component scoping

3. **Implement prop expression evaluation in parent scope with serialization to component scope**
   - Evaluate prop expressions in parent context
   - Serialize evaluated values to component's x-data
   - Handle dynamic prop updates
   - Support both literal values and expressions

4. **Implement component default value merging with passed props**
   - Extract default prop values from component fence section
   - Merge passed props with defaults
   - Handle undefined/null prop values
   - Preserve type information during merge

5. **Add prop validation: enforce required props and default values**
   - Define prop validation syntax in fence section
   - Implement required prop checking at transformation time
   - Validate prop types when specified
   - Provide helpful error messages for missing/invalid props

6. **Optimize component re-rendering based on prop changes**
   - Track which props are used in component template
   - Implement change detection for props
   - Optimize Alpine.js reactivity for component updates
   - Consider memoization for expensive prop transformations

## Component Style Extraction

### Current Limitation

Component scoped styles defined in `<style>` tags within component files are not automatically extracted and included in the parent page output. Currently, styles must be manually copied to the parent page's style section.

### How Svelte/Plenti Handles Component Styles

In Svelte (and by extension, Plenti which uses Svelte), component styles work as follows:

1. **Scoped CSS**: Each component's `<style>` block is automatically scoped to that component
2. **Hash-based Class Names**: Svelte adds unique hash-based class names (e.g., `svelte-1m9hqrl`) to elements
3. **CSS Extraction**: All component styles are extracted and bundled into a single CSS file
4. **Automatic Inclusion**: The bundled CSS is automatically included in the page output
5. **Cascade Behavior**: Component styles from child components are accessible to parent components

**Example from Plenti:**
- Component `progress_circle.svelte` has `<style>` with `#progress { ... }`
- Svelte compiles this to `.svelte-1m9hqrl` classes
- All styles from `progress_circle`, `WebLensPopup`, and `footer` components are bundled
- The footer component (parent) can use styles from its child components without manual importing

### Proposed Enhancement

Implement automatic extraction and inclusion of component scoped styles into parent page output, similar to how Svelte handles it.

### Task

7. **Implement automatic extraction and inclusion of component scoped styles into parent page output**
   - Extract `<style>` blocks from component templates during parsing
   - Generate unique scope identifiers for each component (e.g., hash-based class names)
   - Add scope identifiers to component elements during transformation
   - Transform CSS selectors to include scope identifiers
   - Track which components are used on a page
   - Aggregate all component styles (component + nested component styles)
   - Deduplicate identical style blocks
   - Inject aggregated scoped styles into page `<style>` section or separate CSS file
   - Ensure child component styles are accessible to parent components
   - Handle style precedence and cascade order
   - Consider performance: Cache transformed styles per component

## Alpine.js x-data Formatting

### Current Implementation

As of 2025-10-07, the system has **two separate x-data formatters** serving different architectural levels:

#### 1. Transformer Level: `alpineDataFormatter()`
**Location**: `transformer/alpine.go:694-760`

**Purpose**: Formats x-data for AST-level component wrapping during transformation

**Features**:
- Function detection via `isFunctionExpression()` (supports 6+ patterns)
- Value formatting via `FormatGoValueToJS()` (preserves function syntax)
- Topological sorting for dependency ordering
- Self-reference detection with function wrapper syntax
- Iterator cleanup (removes leaked loop variables)

**Status**: ⚠️ **UNEXPORTED** - Cannot be called from outside transformer package

#### 2. Server Level: `buildXDataFromProps()`
**Location**: `cmd/server/main.go:187-244`

**Purpose**: Formats x-data for HTML-level injection into `<body>`/`<html>` tags

**Features**:
- Generates JavaScript object literal (NOT JSON)
- Function detection via prefix/arrow function checks
- Function minification for HTML attributes
- Proper JavaScript escaping (single quotes)
- HTML entity escaping for attributes
- Sorted keys for consistent output

**Why Both Exist**:
- Transformer works on AST nodes (component wrapping)
- Server works on final HTML strings (tag injection)
- Transformer can't modify `<body>` tag (not in AST)
- Server can't use unexported `alpineDataFormatter`

### Known Limitations

1. **Code Duplication**: Two formatters implement similar logic
2. **Parser Gap**: Functions not extracted into `fence.Variables` (requires `extractFunctionsFromFence()` workaround)
3. **Arrow Functions**: Complex arrow functions may not be captured by regex
4. **Function Minification**: Basic whitespace removal, not full JS minification

### Proposed Enhancement

**Option 1: Export alpineDataFormatter** (RECOMMENDED for v1.0+)

Export the transformer's function for reuse by server:

```go
// transformer/alpine.go
func AlpineDataFormatter(dataScope map[string]any) string {
    return alpineDataFormatter(dataScope)
}
```

Then use in server:
```go
// cmd/server/main.go
xDataValue := transformer.AlpineDataFormatter(props)
```

**Benefits**:
- ✅ Single source of truth for x-data formatting
- ✅ Reuses transformer's topological sorting
- ✅ Reuses self-reference detection
- ✅ Removes ~100 lines of duplicate code

**Risks**:
- ⚠️ Low risk: Server currently working, refactor could introduce bugs
- ⚠️ May need to handle server-specific formatting needs

**Option 2: Unify x-data Injection at Transformer Level** (FUTURE)

Move ALL x-data generation to transformer:

```
Current:
  renderer.Render() → markup (no x-data on <body>)
  server manually injects x-data into <body>

Future:
  renderer.Render() → markup (x-data already on <body>)
  server just sends response
```

**Benefits**:
- ✅ True separation of concerns
- ✅ Server becomes thin wrapper
- ✅ All x-data logic in one place

**Risks**:
- ⚠️ High complexity: Requires AST changes to include `<body>` tag
- ⚠️ May break existing architecture

**Option 3: Enhanced Parser for Function Extraction** (LONG-TERM)

Modify parser to extract function declarations into `fence.Variables`:

```go
// parser/fence.go
type Variable struct {
    Name         string
    Value        string
    IsFunction   bool  // NEW: Flag for function declarations
    FunctionBody string // NEW: Full function body
}
```

**Benefits**:
- ✅ No `extractFunctionsFromFence()` workaround needed
- ✅ Functions treated as first-class fence declarations
- ✅ Better error reporting for invalid functions

**Risks**:
- ⚠️ Medium complexity: Parser changes
- ⚠️ Need to handle function vs variable distinction

### Tasks

8. **Export alpineDataFormatter from transformer package**
   - Add public `AlpineDataFormatter()` wrapper function
   - Update server to use exported function
   - Remove `buildXDataFromProps()` from server
   - Verify all tests still pass
   - Update documentation

9. **Enhanced parser: Extract function declarations into fence.Variables**
   - Add `IsFunction` flag to `Variable` struct
   - Implement function declaration parsing in fence parser
   - Store full function body in AST
   - Remove `extractFunctionsFromFence()` workaround from server
   - Add tests for function parsing edge cases

10. **Unify x-data injection at transformer level**
    - Modify transformer to inject x-data into `<body>`/`<html>` tags
    - Update renderer to return complete HTML with x-data
    - Remove server-level x-data injection
    - Update `needsAlpineWrapper()` logic to handle page-level tags
    - Verify no regressions in component rendering

### Implementation Notes

**When to Implement**:
- ⏰ **Task 8**: After v1.0 release and all spec tasks complete
- ⏰ **Task 9**: When adding more complex function support (async, generators)
- ⏰ **Task 10**: Long-term architectural refactor (v2.0+)

**Prerequisites**:
- ✅ Tasks 1-4 of `.agent-os/specs/2025-10-07-fix-server-xdata-building/` complete
- ✅ Functions tested and working in browser
- ✅ No regressions discovered
- ✅ All existing tests passing

**Trigger Conditions for Task 8**:
- Need to add complex features to `buildXDataFromProps()` (e.g., topological sorting)
- Discovering bugs in server's function detection
- Community requests for consistent x-data formatting

### Reference Documentation

**Completion Reports**:
- `.agent-os/specs/2025-10-07-fix-server-xdata-building/TASK1_COMPLETION_REPORT.md` - Server refactor details
- `.agent-os/specs/2025-10-07-fix-server-xdata-building/TASK2_COMPLETION_REPORT.md` - Transformer integration verification

**Key Files**:
- `transformer/alpine.go:694` - `alpineDataFormatter()` implementation
- `transformer/alpine.go:54` - `isFunctionExpression()` detection
- `transformer/alpine.go:294` - `FormatGoValueToJS()` formatting
- `cmd/server/main.go:187` - `buildXDataFromProps()` server implementation
- `cmd/server/main.go:185` - `extractFunctionsFromFence()` workaround

## Priority and Dependencies

### High Priority
- Task 7: Component style extraction (immediate pain point)
- Task 1: Create spec document (foundation for other tasks)

### Medium Priority
- Task 2: Component prop scoping implementation
- Task 4: Default value merging
- Task 8: Export alpineDataFormatter (code deduplication, low risk)

### Low Priority (depends on Tasks 1-4)
- Task 3: Prop expression evaluation
- Task 5: Prop validation
- Task 6: Re-rendering optimization
- Task 9: Enhanced parser for functions (medium risk)

### Very Low Priority (architectural refactors)
- Task 10: Unify x-data injection (high complexity, v2.0+)
- Task 11: Fix loop index transformation (cosmetic issue only)

## Loop Index Transformation

### Current Issue

When generating Alpine.js `x-for` directives, the transformer creates unnamed index variables, resulting in cosmetic HTML issues:

```html
<!-- Current output -->
<template x-for="(product, ) in products">
  <!-- Notice the unnamed index: (product, ) -->
</template>

<!-- Expected output -->
<template x-for="(product, index) in products">
  <!-- Named index: (product, index) -->
</template>
```

**Impact**:
- ⚠️ Cosmetic only - functionality not affected
- Functions and rendering work correctly
- May cause confusion when reading generated HTML
- Alpine.js handles it gracefully (ignores unnamed parameters)

### Root Cause

The transformer's loop transformation (`transformer/loops.go`) generates x-for attributes but doesn't always provide a name for the index variable when one is used in the source template.

**Example from source template**:
```html
{for product in products}
  <!-- No index used -->
{/for}

{for product, index in products}
  <!-- Index explicitly used -->
{/for}
```

Both cases currently may generate `(product, )` instead of properly handling the index.

### Observed In

- **Spec**: `.agent-os/specs/2025-10-07-fix-server-xdata-building/`
- **Page**: `examples/pages/comprehensive-simple.html`
- **Sections**: Loop rendering (Section 3, 5, 6)
- **Date Discovered**: 2025-10-07

### Proposed Enhancement

Improve the loop transformer to:
1. Detect when index variable is NOT used in template
2. Generate x-for without index parameter: `x-for="product in products"`
3. When index IS used, generate with named index: `x-for="(product, index) in products"`

### Tasks

11. **Fix loop index transformation to generate named index variables**
    - Analyze `transformer/loops.go` for x-for generation logic
    - Detect if index variable is referenced in loop body
    - If index not used: Generate `x-for="item in items"` (no index)
    - If index used: Generate `x-for="(item, index) in items"` (named index)
    - Add tests for both cases
    - Verify no regressions in existing loop tests

### Implementation Notes

**Priority**: Low (cosmetic issue only)

**When to Implement**:
- ⏰ When adding more sophisticated loop features
- ⏰ During general transformer refactoring
- ⏰ If community reports confusion about generated HTML

**Prerequisites**:
- ✅ Fix-server-xdata-building spec complete
- ✅ All loop tests passing
- ✅ Understanding of Alpine.js x-for syntax requirements

**Related Files**:
- `transformer/loops.go` - Loop transformation logic
- `tests/alpine/loop_test.go` - Loop test cases
- `parser/loop.go` - Loop parsing logic

## Notes

These enhancements would significantly improve:
- Developer experience (less manual workarounds)
- Component reusability and isolation
- Performance and optimization opportunities
- Type safety and error prevention
- Code maintainability (reduce duplication)
- Architectural consistency

The current implementation is functional and production-ready. These are optimization opportunities, not critical bugs.
