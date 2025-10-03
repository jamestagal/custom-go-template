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
   - Handles expressions, conditionals, loops, components
   - Key files: `parser.go`, `directives.go`, `components.go`, `expressions.go`

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

6. **`cmd/server/`** - Development server
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
- `prop propName = defaultValue`
- `let/const/var variableName = value`
- Functions for Alpine.js data

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

### Data Scope Management
- All variables referenced in templates MUST be in Alpine.js x-data scope
- Track variables across fence sections, expressions, conditionals, loops
- The transformer builds a data scope map that becomes the x-data object
- Component props must be extracted and passed correctly

### Transformation Order
1. Parse template to AST
2. Extract fence data (props, variables, functions)
3. Transform AST nodes to Alpine.js equivalents
4. Build x-data scope from all discovered variables
5. Render final HTML with Alpine.js directives

### Testing Requirements
- Write tests BEFORE implementing features
- Test edge cases: nested conditionals, nested loops, components in loops
- Use the `tests/alpine/` and `tests/components/` directories
- Key test files show patterns: `alpine_integration_test.go`, `components_test.go`

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
- `cmd/` - Executables
  - `server/` - Development server
  - `test_*/` - Testing utilities

## Important Notes

- Always prefer editing existing files over creating new ones
- Variable syntax uses single curly braces `{var}`, not double `{{var}}`
- The fence section uses JavaScript syntax for variables and functions
- Components must be registered before use
- Alpine.js directives are generated automatically; don't write them manually in templates
