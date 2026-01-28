# System Architecture

**Last Updated:** 2026-01-29

## Build Time vs Compile Time vs Runtime

These terms are often confused. Here's what each means in the context of web development and our Go template engine:

| Phase | When It Happens | Who Executes | Our System |
|-------|-----------------|--------------|------------|
| **Compile Time** | Developer builds Go binary | Go compiler | `go build ./...` |
| **Build Time** | Server processes request | Go server | Template → HTML transformation |
| **Runtime** | User views page in browser | JavaScript (Alpine.js) | Reactive updates, events |

### Compile Time (Go Binary)

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Go Source      │ ──▶ │  Go Compiler    │ ──▶ │  Server Binary  │
│  *.go files     │     │  go build       │     │  ./server       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**What happens:** Go source code is compiled into machine code. Type checking, optimization, and linking occur here.

**Relevant to us:** Our parser, transformer, renderer packages are compiled. Errors in Go code are caught here.

---

### Build Time (Template Processing)

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Template       │ ──▶ │  Go Server      │ ──▶ │  HTML Output    │
│  {for x in y}   │     │  Parse/Transform│     │  <div>...</div> │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                      │
         │              ┌───────┴───────┐
         │              │  JSON Content │
         │              │  content/*.json│
         │              └───────────────┘
         │
    On each HTTP request (dev) or pre-build (production)
```

**What happens:** Templates are parsed, AST is built, data is injected, loops are expanded, components are resolved, HTML is generated.

**This is our core innovation:** Most work happens here, not in the browser.

**Build-time operations:**
- Template parsing → AST
- JSON content loading
- Loop expansion (when data is available)
- Component inlining
- CSS/JS scoping
- x-data optimization

---

### Runtime (Browser JavaScript)

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  HTML + Alpine  │ ──▶ │  Browser        │ ──▶ │  Reactive UI    │
│  x-data, x-text │     │  Alpine.js      │     │  User interacts │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**What happens:** Alpine.js initializes, binds data, handles events, updates DOM reactively.

**Runtime operations:**
- `x-data` scope initialization
- `x-text`, `x-html` binding
- `x-if`, `x-show` conditional display
- `x-for` loop rendering (only for runtime-only collections)
- `@click`, `@submit` event handling
- `$store` global state access

---

## Our Philosophy: Maximize Build Time, Minimize Runtime

```
┌────────────────────────────────────────────────────────────────┐
│                        WORK DISTRIBUTION                        │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  BUILD TIME (Go Server)                    RUNTIME (Browser)   │
│  ████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░  │
│  ▲                                   ▲                         │
│  │                                   │                         │
│  │ • Parse templates                 │ • Reactive binding      │
│  │ • Expand loops                    │ • Event handling        │
│  │ • Resolve components              │ • Store updates         │
│  │ • Inject content                  │ • Dynamic x-for         │
│  │ • Optimize x-data                 │   (runtime collections) │
│  │ • Scope CSS/JS                    │                         │
│                                                                │
│  ~80% of work                        ~20% of work              │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## The Transformation Pipeline

```
┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   Template   │   │    Parser    │   │ Transformer  │   │   Renderer   │   │    Output    │
│   Source     │──▶│              │──▶│              │──▶│              │──▶│              │
│              │   │   ast/*      │   │ transformer/*│   │  renderer/*  │   │    HTML      │
│  .html file  │   │  parser/*    │   │  analyzer/*  │   │  scoping/*   │   │  + Alpine.js │
└──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘
                          │                  │                  │
                          │                  │                  │
                   ┌──────┴──────┐    ┌──────┴──────┐    ┌──────┴──────┐
                   │   loader/*  │    │  builder/*  │    │ cmd/server  │
                   │ JSON content│    │  Registry   │    │   Serving   │
                   └─────────────┘    └─────────────┘    └─────────────┘
```

---

## Package Reference Matrix

### Core Pipeline Packages

| Package | Phase | Purpose | Key Files |
|---------|-------|---------|-----------|
| [`ast/`](../../ast/) | Build | AST node definitions | `template.go`, `element.go`, `loop.go`, `conditional.go` |
| [`parser/`](../../parser/) | Build | Template → AST | `parser.go`, `expressions.go`, `components.go` |
| [`transformer/`](../../transformer/) | Build | AST → Alpine.js AST | `transformer.go`, `loops.go`, `scope.go` |
| [`renderer/`](../../renderer/) | Build | AST → HTML string | `render.go`, `component.go`, `plenti_html.go` |
| [`scoping/`](../../scoping/) | Build | CSS/JS isolation | `css.go`, `js.go`, `html.go` |

### Support Packages

| Package | Phase | Purpose | Key Files |
|---------|-------|---------|-----------|
| [`loader/`](../../loader/) | Build | JSON content loading | `loader.go` |
| [`analyzer/`](../../analyzer/) | Build | Runtime expression detection | `scope.go` |
| [`builder/`](../../builder/) | Build | Component registry generation | `registry_generator.go` |
| [`cmd/server/`](../../cmd/server/) | Build | HTTP server, orchestration | `main.go` |

### Runtime Files (Browser)

| File | Phase | Purpose |
|------|-------|---------|
| [`core/runtime-components.js`](../../core/runtime-components.js) | Runtime | `$renderDynamicComponent` Alpine magic |
| [`generated/layouts.js`](../../generated/layouts.js) | Runtime | Component registry (auto-generated) |
| Alpine.js (CDN) | Runtime | Reactive framework |

---

## Package Deep Dive

### 1. `ast/` - Abstract Syntax Tree

**Phase:** Build Time (data structures)

**Purpose:** Defines the node types that represent parsed template structures.

```go
// Key types
type Template struct {
    Nodes []Node           // All top-level nodes
    Fence *FenceSection    // Front matter (---)
}

type Loop struct {
    Iterator   string      // "item" in {for item in items}
    Collection string      // "items" in {for item in items}
    Body       []Node      // Nodes inside the loop
}

type Conditional struct {
    Condition string       // Expression to evaluate
    ThenBlock []Node       // Nodes when true
    ElseBlock []Node       // Nodes when false (or else-if chain)
}
```

**Key Files:**
| File | Contains |
|------|----------|
| `template.go` | Template, FenceSection |
| `element.go` | Element, Attribute |
| `loop.go` | Loop node |
| `conditional.go` | Conditional, ElseIf |
| `component.go` | ComponentNode |
| `expression.go` | ExpressionNode |

---

### 2. `parser/` - Template Parser

**Phase:** Build Time

**Purpose:** Converts template source text into AST nodes.

```
Input:  "{for item in items}<div>{item}</div>{/for}"
         │
         ▼
Output: Loop{
          Iterator: "item",
          Collection: "items",
          Body: [Element{Tag: "div", Children: [Expression{Value: "item"}]}]
        }
```

**Architecture:**
- **Unified parsing path** via `AnyNodeParser`
- **Block parsers** with depth tracking (handles nesting)
- **Parser combinators** for composable parsing

**Key Files:**
| File | Parses |
|------|--------|
| `parser.go` | Entry point, AnyNodeParser |
| `expressions.go` | `{variable}`, `{obj.prop}` |
| `conditionals.go` | `{if}...{/if}` blocks |
| `loops.go` | `{for}...{/for}` blocks |
| `components.go` | `<Component />`, `<Component:dynamic>` |
| `html.go` | HTML elements and attributes |

---

### 3. `transformer/` - AST Transformer

**Phase:** Build Time (the critical phase)

**Purpose:** Transforms template AST into Alpine.js-compatible AST.

```
Input:  Loop{Iterator: "item", Collection: "items", Body: [...]}
         │
         │  (if items resolvable from dataScope)
         ▼
Output: [Element{...}, Element{...}, Element{...}]  // Expanded!

         │  (if items is runtime-only, e.g., $store.cart.items)
         ▼
Output: Element{
          Tag: "template",
          Attributes: [{Name: "x-for", Value: "item in $store.cart.items"}],
          Children: [...]
        }
```

**Key Responsibilities:**
- **Loop expansion** (build-time when possible)
- **Component resolution** (inline or runtime wrapper)
- **Expression transformation** (`{var}` → `x-text="var"`)
- **Scope management** (track what goes in x-data)
- **x-data optimization** (RuntimeVarTracker)

**Key Files:**
| File | Transforms |
|------|------------|
| `transformer.go` | Main entry, orchestration |
| `loops.go` | Loop expansion logic |
| `conditionals.go` | `{if}` → `<template x-if>` |
| `expressions.go` | `{var}` → `<span x-text>` |
| `components.go` | Component inlining |
| `scope.go` | Data scope, RuntimeVarTracker |
| `alpine.go` | Alpine.js directive generation |
| `dynamic_component_by_name.go` | Component name resolution |

---

### 4. `analyzer/` - Expression Analyzer

**Phase:** Build Time

**Purpose:** Determines if expressions can be resolved at build time or require runtime.

```go
// Key function
func IsRuntimeExpression(expr string, dataScope map[string]interface{}) bool

// Examples:
IsRuntimeExpression("items", scope)           // false if items in scope
IsRuntimeExpression("$store.cart.items", _)   // true (always runtime)
IsRuntimeExpression("component.name", scope)  // depends on scope
```

**Decision Tree:**
```
Expression
    │
    ├── Starts with "$store." ────────────▶ RUNTIME
    │
    ├── Base variable in scope?
    │       │
    │       ├── Yes, value is nil ─────────▶ RUNTIME (loop iterator marker)
    │       │
    │       └── Yes, value exists ─────────▶ BUILD-TIME
    │
    └── Not in scope ──────────────────────▶ RUNTIME
```

---

### 5. `renderer/` - HTML Renderer

**Phase:** Build Time (final output generation)

**Purpose:** Converts transformed AST into HTML string.

```go
// Main functions
func Render(template *ast.Template) string
func RenderWithWrapper(content, wrapper *ast.Template) string
func RenderFromJSON(template *ast.Template, data map[string]interface{}) string
```

**Key Files:**
| File | Renders |
|------|---------|
| `render.go` | Main rendering logic |
| `component.go` | Component rendering |
| `plenti_html.go` | Wrapper injection, script tags |
| `fence.go` | Fence section processing |
| `content_injection.go` | JSON → export let injection |

---

### 6. `loader/` - Content Loader

**Phase:** Build Time

**Purpose:** Loads and parses JSON content files.

```go
// Route to file mapping
RoutePathToFilePath("/about")     // → "content/pages/about.json"
RoutePathToFilePath("/news/item") // → "content/news/item.json"

// Content loading
data, _ := LoadContentJSON("content/pages/index.json")

// Component field extraction (Plenti pattern)
fields := ExtractComponentFields(data, "hero2436")
```

---

### 7. `builder/` - Registry Generator

**Phase:** Build Time (generates runtime assets)

**Purpose:** Generates JavaScript component registry for runtime resolution.

```go
// Generates: generated/layouts.js
export default {
  'hero2436': (props) => `<section class="hero">
    <h1>${props.title}</h1>
  </section>`,
  'services2437': (props) => `...`,
  // ... all components
}
```

**Note:** This registry is only loaded when runtime component resolution is needed.

---

### 8. `scoping/` - CSS/JS Scoping

**Phase:** Build Time

**Purpose:** Isolates component styles and scripts.

```html
<!-- Input -->
<style>
  .hero { color: red; }
</style>

<!-- Output (scoped) -->
<style>
  .hero[data-scope-abc123] { color: red; }
</style>
```

---

### 9. `cmd/server/` - Development Server

**Phase:** Build Time (orchestration)

**Purpose:** HTTP server that orchestrates the entire pipeline.

**Request Flow:**
```
HTTP Request
     │
     ▼
┌─────────────────────┐
│ Route Matching      │  /about → content/pages/about.json
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Load JSON Content   │  loader.LoadContentJSON()
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Parse Template      │  parser.Parse()
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Transform AST       │  transformer.Transform()
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Render HTML         │  renderer.Render()
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Inject Scripts      │  (conditional based on HasRuntimeComponents)
└──────────┬──────────┘
           │
           ▼
HTTP Response (HTML)
```

---

## Key Architectural Decisions

| Decision | Rationale | Implementation |
|----------|-----------|----------------|
| **Build-time loop expansion** | Zero runtime JS for static content | `transformer/loops.go` |
| **Alpine.js over server JS** | Security, performance, reactivity | All templates → Alpine directives |
| **Opt-in magic variables** | Bandwidth optimization | `export let` in fence section |
| **Conditional script injection** | Smaller payloads for static pages | `cmd/server/main.go` |
| **x-data optimization** | Reduce payload size | `transformer/scope.go` RuntimeVarTracker |
| **Unified parser path** | Consistent behavior, no edge cases | `parser/parser.go` AnyNodeParser |

---

## When Each Package Runs

```
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP REQUEST LIFECYCLE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. cmd/server   ─▶  Route matching, request handling           │
│        │                                                        │
│        ▼                                                        │
│  2. loader       ─▶  Load JSON content from content/*.json      │
│        │                                                        │
│        ▼                                                        │
│  3. parser       ─▶  Parse template to AST                      │
│        │              (uses ast/ for node types)                │
│        ▼                                                        │
│  4. transformer  ─▶  Transform AST, expand loops, resolve comps │
│        │              (uses analyzer/ for runtime detection)    │
│        ▼                                                        │
│  5. renderer     ─▶  Render HTML from transformed AST           │
│        │              (uses scoping/ for CSS/JS isolation)      │
│        ▼                                                        │
│  6. cmd/server   ─▶  Inject scripts (if needed), send response  │
│                                                                 │
│  ─────────────────── BUILD TIME COMPLETE ─────────────────────  │
│                                                                 │
│  7. Browser      ─▶  Alpine.js initializes, binds, reacts       │
│                      (uses core/*.js, generated/*.js if needed) │
│                                                                 │
│  ─────────────────── RUNTIME ─────────────────────────────────  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## See Also

- [BUILD_TIME_LOOP_EXPANSION.md](./BUILD_TIME_LOOP_EXPANSION.md) - Deep dive on loop expansion
- [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md) - Full developer documentation
- [STORE_DEVELOPER_GUIDE.md](./STORE_DEVELOPER_GUIDE.md) - Global store system
- [CLAUDE.md](../../CLAUDE.md) - AI assistant context (includes package details)
