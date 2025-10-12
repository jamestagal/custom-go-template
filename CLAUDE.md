# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a custom Go template engine that transforms Svelte-inspired template syntax into Alpine.js-compatible HTML. The engine parses custom template syntax, transforms it through an AST, and renders reactive HTML components.

**Core Purpose**: Parse templates with syntax like `{if condition}`, `{for item in items}`, and `{variable}`, then transform them into Alpine.js directives (`x-if`, `x-for`, `x-text`, etc.).

## Architecture

The codebase follows a pipeline architecture:

```
Template Source → Parser → AST → Transformer → Rendered HTML/CSS/JS
```

### Key Packages

1. **`ast/`** - Defines the Abstract Syntax Tree node types
   - `Template` - Root node containing all parsed nodes
   - `FenceSection` - Front matter with imports, props, variables
   - `Element` - HTML elements with attributes
   - `Conditional` - If/else-if/else blocks
   - `Loop` - For loops
   - `ComponentNode` - Component references
   - `ExpressionNode` - Dynamic expressions in `{}`

2. **`parser/`** - Converts template syntax to AST
   - Uses parser combinators for composable parsing
   - **Unified Architecture (2025-10-06)**: Single parsing path using BlockConditionalParser and BlockLoopParser
   - Key files: `parser.go`, `directives.go`, `components.go`, `expressions.go`, `html.go`

   **Parser Architecture**:
   - `AnyNodeParser` - Central entry point for parsing any node type
   - `BlockConditionalParser` - Parses complete `{if}...{/if}` blocks with depth tracking
   - `BlockLoopParser` - Parses complete `{for}...{/for}` blocks with depth tracking
   - `ElementParser` → `parseChildren` → `AnyNodeParser` (consistent parsing path)
   - **DEPRECATED**: `processDirectiveNodes`, `processConditionals`, `processLoops` (see `parser/process_directives.go`)

   **Important**: The parser now uses a single, unified path. Do NOT use the old marker-based parsers or post-processing functions.

3. **`transformer/`** - Transforms AST to Alpine.js compatible nodes
   - **Critical**: This is where template syntax becomes Alpine.js directives
   - `transformer.go` - Main transformation entry point
   - `expressions.go` - Transforms `{var}` to `<span x-text="var"></span>`
   - `conditionals.go` - Transforms `{if}` to `<template x-if>`
   - `loops.go` - Transforms `{for}` to `<template x-for>`
   - `components.go` - Handles component inclusion and prop passing
   - `alpine.go` - Alpine.js-specific transformations
   - `scope.go` - Data scope management for x-data

4. **`renderer/`** - Generates final HTML/CSS/JS from transformed AST
   - `render.go` - Main rendering logic
   - `component.go` - Component rendering
   - `fence.go` - Fence section processing

5. **`scoping/`** - CSS and JS scoping utilities
   - `css.go` - Scope CSS to components
   - `js.go` - Scope JavaScript
   - `html.go` - HTML scoping utilities

6. **`loader/`** - Content JSON loading utilities
   - Loads JSON content from `content/` directory based on route paths
   - `LoadContentJSON()` - Loads and parses JSON from content files
   - `RoutePathToFilePath()` - Maps route paths to file paths (e.g., `/store-demo` → `content/pages/store-demo.json`)
   - `ExtractComponentFields()` - Extracts component data from Plenti structure (components array)
   - `IsCollectionType()` - Detects JSON format type (collection vs single)
   - Supports both Plenti collection types (with `components` array) and single types (flat JSON)
   - Used by the export let system to inject content from JSON files
   - See: `.agent-os/specs/2025-10-11-export-let-content-injection/` for full details

7. **`cmd/server/`** - Development server
   - Serves templates at http://localhost:3000
   - Registers components from `examples/components/`
   - Extracts props, variables, and functions from fence sections

## Template Syntax

### Expressions
- `{variable}` → `<span x-text="variable"></span>`
- Use single curly braces (not double like Svelte)

### Conditionals
```
{if condition}
  content
{else if otherCondition}
  content
{else}
  content
{/if}
```
Transforms to `<template x-if>`, `<template x-else-if>`, `<template x-else>`

### Loops
```
{for item in items}
  <div>{item}</div>
{/for}
```
Transforms to `<template x-for="item in items">`

### Components
```
<ComponentName prop1="value" prop2={dynamicValue} />
```
Components are imported from `examples/components/` and registered automatically.

### Fence Section
Front matter between `---` markers containing:
- `import ComponentName from './components/ComponentName.html'`
- `import store from './stores/storeName.js'` - Import global stores
- `store storeName = { ... }` - Inline store definitions
- `prop propName = defaultValue` - Component props with default values
- `export let propName, propName2` - Props from JSON content (Svelte-compatible)
- `let/const/var variableName = value`
- Functions for Alpine.js data

### Export Let Content Injection

Components can use `export let` to declare props that come from JSON content files, following Svelte patterns for Plenti compatibility.

**Syntax:**
```html
---
export let title, description, link, image
---

<div class="card">
  <h2>{title}</h2>
  <p>{description}</p>
  <a href="{link}">Learn More</a>
</div>
```

**JSON Structure (Plenti format):**
```json
{
  "components": [
    {
      "name": "card",
      "fields": {
        "title": "Welcome",
        "description": "Content from JSON",
        "link": "/about",
        "image": "hero.jpg"
      }
    }
  ]
}
```

**How It Works:**
1. Route handler loads JSON from `content/pages/` based on URL path
2. System extracts component fields matching component name
3. Content is injected into exported props before rendering
4. Missing props with defaults use the default value (with warning)
5. Missing props without defaults cause an error

**Key Files:**
- `loader/loader.go` - Loads JSON content from files
- `renderer/content_injection.go` - Injects content into exported props
- `content/pages/*.json` - Content files with Plenti structure

**See:** `.agent-os/specs/2025-10-11-export-let-content-injection/` for full implementation details

### Global Store System

The template engine supports Alpine.js global stores for shared reactive state across components.

**Store Syntax**: `{$storeName.property}` transforms to `$store.storeName.property`

**Store Definitions**:
1. **Inline stores**: Define in fence section with `store name = { ... }`
2. **External files**: Import from `stores/` directory with `import store from './stores/name.js'`
3. **Component imports**: Components can import stores they depend on

**Priority order**: Inline > Imported > External (when conflicts exist)

**Key Files**:
- `ast/store.go` - StoreDefinitionNode AST node
- `parser/store_expression.go` - Parse `{$store.prop}` syntax
- `transformer/stores.go` - Transform store expressions to Alpine.js
- `transformer/store_tracking.go` - Track referenced stores
- `renderer/stores.go` - Render Alpine.store() initializations

**Store Transformation Examples**:
```
{$auth.isLoggedIn}           → <span x-text="$store.auth.isLoggedIn"></span>
{if $auth.isLoggedIn}        → <template x-if="$store.auth.isLoggedIn">
{for item in $cart.items}    → <template x-for="item in $store.cart.items">
@click="$store.auth.login()" → @click="$store.auth.login()" (preserved)
```

**Important Implementation Details**:
- Store expressions are tracked during transformation for proper initialization
- Literal content elements (`<pre>`, `<code>`, `<textarea>`) skip store transformation
- Already-transformed `$store.*` expressions are NOT re-transformed (prevents double prefix bug)
- Store methods and computed properties (getters) are fully supported
- Stores are initialized via `Alpine.store()` in rendered JavaScript

**Store System Integration** (see `.agent-os/specs/2025-10-07-global-store-system/`):
- Phase 1: AST & Parser - Store definition nodes and `{$store.prop}` parsing
- Phase 2: Transformation - Transform store expressions to Alpine.js syntax
- Phase 3: Rendering - Generate Alpine.store() initialization code
- Phase 4: Integration - Wire stores into component rendering pipeline
- Phase 5: Documentation - Developer guides and examples

## Common Commands

### Build
```bash
go build ./...
```

### Run Tests
```bash
# All tests
go test ./... -v

# Specific package
go test ./transformer -v
go test ./tests/alpine -v

# Specific test
go test ./transformer -run TestConditionals -v
```

### Run Development Server
```bash
go run cmd/server/main.go
# Visit http://localhost:3000
```

### Test Single Component
```bash
go run cmd/test_component/main.go
```

## Critical Implementation Rules

### Parser Architecture (Updated 2025-10-06)

**IMPORTANT**: The parser now uses a unified, single-path architecture:

1. **Always use BlockConditionalParser and BlockLoopParser** via `AnyNodeParser`
2. **Never use** the old `processDirectiveNodes`, `processConditionals`, or `processLoops` functions
3. **Element children parsing**: `parseChildren` calls `AnyNodeParser` directly, not `parseChildNode`

**Why this matters**:
- The old dual-path approach caused content after `{/if}` and `{/for}` to be incorrectly consumed into the directive
- BlockConditionalParser and BlockLoopParser use depth tracking to correctly identify directive boundaries
- Post-processing tried to re-organize already-parsed nodes, causing bugs

**See**: `.agent-os/specs/2025-10-06-parser-unification/` for full details

### Data Scope Management
- All variables referenced in templates MUST be in Alpine.js x-data scope
- Track variables across fence sections, expressions, conditionals, loops
- The transformer builds a data scope map that becomes the x-data object
- Component props must be extracted and passed correctly

**Props vs Stores - When to Use Which**:

Use **Props** when:
- Data is component-specific and passed from parent
- Each component instance needs different values
- Data flows one direction (parent → child)
- Example: `<UserCard name="John" age={30} />`

Use **Stores** when:
- State is shared across multiple components
- Multiple components read AND write the same data
- State persists across component instances
- Need global application state (auth, theme, cart)
- Example: `{if $auth.isLoggedIn}` used by LoginButton, UserMenu, ProtectedRoute

### Transformation Order
1. Parse template to AST (using unified parser path)
2. Extract fence data (props, variables, functions)
3. Transform AST nodes to Alpine.js equivalents
4. Build x-data scope from all discovered variables
5. Render final HTML with Alpine.js directives

### Testing Requirements
- Write tests BEFORE implementing features
- Test edge cases: nested conditionals, nested loops, components in loops
- Use the `tests/alpine/` and `tests/components/` directories
- Key test files show patterns: `alpine_integration_test.go`, `components_test.go`
- Regression tests: `parser/conditional_bug_test.go`, `parser/nested_conditional_loop_test.go`

### Error Handling
- Provide detailed, actionable error messages
- Include line/column info where possible
- Fail gracefully without cascading errors

### Whitespace Handling
- Preserve meaningful whitespace in text nodes
- The transformer has specific whitespace handling in `text.go`
- Tests validate whitespace preservation

## Component System

Components are defined in `examples/components/` as `.html` files. Each component:
- Can have a fence section with props
- Is registered on server startup
- Props are extracted via `extractComponentProps()` in `cmd/server/main.go`
- Component AST is stored and reused when component is referenced

## Alpine.js Integration

The engine targets Alpine.js 3.x. Key integration points:
- x-data directive wraps reactive state
- x-text for text expressions
- x-if/x-else for conditionals
- x-for for loops
- x-bind: for attribute binding
- @click and other event handlers

## Known Issues & Patterns

### Parser Unification (Fixed 2025-10-06)
The "two-parsing-paths" bug has been FIXED. See `.agent-os/specs/2025-10-06-parser-unification/COMPLETION_SUMMARY.md` for details.

**Before**: Dual parsing paths caused content after `{/if}` and `{/for}` to be trapped incorrectly
**After**: Single unified path using BlockConditionalParser and BlockLoopParser

### Recursive Component Transformation
See `docs/RecursiveComponentTranformationChecklist.md` for handling deeply nested components.

### Root Cause Analysis
See `docs/RootCauseAnalysis.md` for debugging transformation issues.

### Alpine Data Wrapper
Components need proper x-data initialization. The transformer determines when to wrap content with x-data directive.

## File Organization

- `examples/` - Template files for testing
  - `pages/` - Page templates
  - `components/` - Reusable components
- `tests/` - Test files organized by feature
  - `alpine/` - Alpine.js integration tests
  - `components/` - Component tests
- `docs/` - Design docs and specifications
- `.agent-os/specs/` - Formal specifications for major changes
- `cmd/` - Executables
  - `server/` - Development server
  - `test_*/` - Testing utilities

## Important Notes

- Always prefer editing existing files over creating new ones
- Variable syntax uses single curly braces `{var}`, not double `{{var}}`
- The fence section uses JavaScript syntax for variables and functions
- Components must be registered before use
- Alpine.js directives are generated automatically; don't write them manually in templates
- **Parser Rule**: Always use `AnyNodeParser` for parsing child nodes to ensure unified parsing path
