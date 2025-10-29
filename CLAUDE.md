# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Session Management

**IMPORTANT - Context Usage Warning**: Always warn the user about remaining context tokens before starting a new task to ensure there is enough context left to complete the task. If context is running low (below 30-40%), ask the user to use `/compact` before proceeding with complex tasks.

**Context Guidelines**:
- ✅ **>50% remaining**: Sufficient for most tasks
- ⚠️ **30-50% remaining**: Good for small-to-medium tasks; warn before large implementations
- 🔴 **<30% remaining**: Should compact before starting new complex work

## Project Overview

This is a custom Go template engine that transforms Svelte-inspired template syntax into Alpine.js-compatible HTML. The engine parses custom template syntax, transforms it through an AST, and renders reactive HTML components.

**Core Purpose**: Parse templates with syntax like `{if condition}`, `{for item in items}`, and `{variable}`, then transform them into Alpine.js directives (`x-if`, `x-for`, `x-text`, etc.).

## Plenti Architecture Pattern

The template engine supports **two rendering modes**:

### 1. Standalone Rendering (Legacy)
- Complete HTML pages with `<!DOCTYPE>`, `<html>`, `<body>` tags
- All content hardcoded in template
- Uses `renderTemplate()` function
- Example: `/store-demo` (self-contained page)

### 2. Wrapper Rendering (Plenti Pattern)
- Content-only templates (no HTML wrapper)
- Content loaded from JSON files in `content/pages/`
- Global HTML wrapper at `layouts/global/html.html`
- Uses `renderWithWrapper()` function
- Example: `/jim-test`, `/pages`, `/` (home)

**Plenti Pattern Flow:**
```
Request → renderWithWrapper() → Load JSON → Parse Wrapper → Inject Content → Render
```

**Migration Example (jim-test):**

Before:
```html
<!DOCTYPE html>
<html>
<head>...</head>
<body>
  <h1>Hello Benjamin!</h1>
</body>
</html>
```

After:
```html
---
export let components
---

{for component in components}
  {if component.name === 'demo_header'}
    <h1>{component.fields.salutation} {component.fields.name}!</h1>
  {/if}
{/for}
```

With `content/pages/jim-test.json`:
```json
{
  "path": "/jim-test",
  "components": [
    {
      "name": "demo_header",
      "fields": {
        "salutation": "Hello",
        "name": "Benjamin"
      }
    }
  ]
}
```

**See:** `MIGRATION_GUIDE.md` for complete migration documentation.


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

7. **`analyzer/`** - Runtime vs build-time expression analysis
   - `scope.go` - ScopeAnalyzer for distinguishing runtime-only from build-resolvable expressions
   - `IsRuntimeExpression()` - Detects loop variables, Alpine stores, operators
   - Checks dataScope for nil-valued entries (loop variable markers)
   - Used by runtime component resolution system
   - See: `.agent-os/specs/2025-10-15-runtime-component-resolution/` for full details

8. **`builder/`** - Component registry generation
   - `registry_generator.go` - Converts component ASTs to JavaScript template functions
   - `GenerateComponentRegistry()` - Creates ES module with all component templates
   - Converts `{expr}` to `${props.expr}` for JavaScript template literals
   - Preserves Alpine.js directives (x-text, x-if, etc.)
   - Context tracking for literal content blocks (style/script tags)
   - Auto-generates `static/js/component-registry.js` on server startup

9. **`cmd/server/`** - Development server
   - Serves templates at http://localhost:3000
   - Registers components from `examples/components/`
   - Extracts props, variables, and functions from fence sections
   - Auto-generates component registry on startup (65 components)
   - Serves runtime JavaScript: `/js/component-registry.js`, `/js/runtime-components.js`

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

### Runtime Component Resolution

The system supports **dynamic component resolution** for components whose names are only known at runtime (e.g., in loops).

**Syntax:**
```html
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**How It Works:**

1. **Scope Analysis** (`analyzer/scope.go`):
   - Detects if component name is runtime-only (loop variable, Alpine store, operator)
   - Checks dataScope for nil-valued entries (loop variable markers)
   - Example: `component.name` has `component` marked as `nil` in dataScope → runtime

2. **Build-Time vs Runtime:**
   - **Build-Time**: String literals like `"Hero2436"` → component inlined directly
   - **Runtime**: Loop variables like `component.name` → runtime wrapper emitted

3. **Runtime Wrapper** (emitted for runtime expressions):
   ```html
   <template x-for="(component, ) in components">
     <div class="dyn-comp-runtime"
          x-data="{compName: component.name, compProps: {...}}"
          x-init="$renderDynamicComponent($el, compName, compProps)">
     </div>
   </template>
   ```

4. **Client-Side Resolution** (`static/js/runtime-components.js`):
   - Alpine.js magic: `$renderDynamicComponent(el, name, props)`
   - Loads component registry from `/js/component-registry.js`
   - Renders component template function with props
   - Re-initializes Alpine directives with `Alpine.initTree(el)`

5. **Component Registry** (`static/js/component-registry.js`):
   - Auto-generated on server startup (65 components)
   - ES module format: `export default { 'Hero2436': (props) => \`...\`, ... }`
   - Template functions convert `{expr}` to `${props.expr}`
   - Alpine directives preserved for client-side hydration

**Key Files:**
- `analyzer/scope.go` - Runtime expression detection
- `transformer/dynamic_component_by_name.go` - Routing and wrapper emission
- `builder/registry_generator.go` - Component registry generation
- `static/js/runtime-components.js` - Alpine.js magic function
- `static/js/component-registry.js` - Auto-generated component templates

**See:** `.agent-os/specs/2025-10-15-runtime-component-resolution/` for full implementation details

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

## Build-Time Loop Expansion

The template engine expands loops at build time (like Svelte) instead of generating runtime Alpine.js x-for templates. This allows loop variables to be available during transformation, enabling dynamic component name resolution.

### How It Works

**Template:**
```html
---
export let components
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Build Process:**
1. Loop transformer resolves `components` array from dataScope (from JSON)
2. For each component, creates iteration scope with actual component data
3. Transforms body nodes with iteration scope (component.name resolves!)
4. Appends transformed nodes to output
5. Result: Fully expanded HTML, no x-for templates

**Output (2 components in array):**
```html
<div class="hero" x-data='{"title":"Welcome"}'>
  <h1 x-text="title">Welcome</h1>
</div>

<div class="services" x-data='{"title":"Our Services"}'>
  <h2 x-text="title">Our Services</h2>
</div>
```

### Hybrid Approach

The system uses **build-time expansion when possible**, **runtime fallback when needed**:

**Build-Time Expansion** (when collection resolvable):
- Regular arrays in dataScope: `items`, `components`, `users`
- Collections from JSON content files
- Produces fully expanded HTML (no x-for)

**Runtime Fallback** (when collection not resolvable):
- Store collections: `$store.cart.items`
- Complex expressions: `Array(count)`, `filteredItems`
- Generates Alpine x-for template for runtime evaluation

### Benefits

1. **Component Name Resolution** - Loop variables available during transformation
2. **Better SEO** - Fully expanded HTML in server-rendered output
3. **Svelte Compatibility** - Matches Svelte's build-time expansion behavior
4. **Performance** - No runtime loop evaluation needed for static content
5. **Flexibility** - Runtime fallback for dynamic content

### Implementation Files

- `transformer/loops.go` - Build-time loop expansion logic
- `transformer/scope.go` - Scope cloning utilities (`cloneScope`, `resolveCollectionFromScope`)
- `transformer/component_loop_integration_test.go` - Integration tests
- `tests/build_time_loop_expansion/` - Output validation tests

**See:** `.agent-os/specs/2025-10-19-build-time-loop-expansion/` for full specification

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
