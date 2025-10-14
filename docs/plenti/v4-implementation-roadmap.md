# V4.0 Implementation Roadmap & Integration Analysis

**Date:** October 14, 2025
**Status:** Blueprint Complete, Implementation Pending
**Estimated Total Effort:** 115-145 hours (8-10 weeks)

---

## Executive Summary

This document provides a comprehensive roadmap for implementing Plenti v4.0's SSG + Hydration architecture using the custom Go template engine. After extensive analysis of both systems, we've identified **perfect alignment** between:

1. The custom Go template engine's current capabilities
2. The hybrid SSR-hydration approach specification
3. The v4.0 SSG blueprint and rationale documents
4. Recent JavaScript formatting bug fixes

**Key Finding:** The foundation is already in place. We need to build the hydration layer on top of existing infrastructure.

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Alignment Assessment](#alignment-assessment)
3. [Gap Analysis](#gap-analysis)
4. [Implementation Phases](#implementation-phases)
5. [Critical Path](#critical-path)
6. [Risk Assessment](#risk-assessment)
7. [Success Metrics](#success-metrics)
8. [Appendix: Technical Details](#appendix-technical-details)

---

## Current State Analysis

### ✅ **What's Already Working**

#### 1. JSON Content Loading (`loader/loader.go`)
```go
// ✅ Handles Plenti's JSON format
func LoadContentJSON(filePath string) (map[string]interface{}, error)

// ✅ Maps routes to files (compatible with Plenti)
func RoutePathToFilePath(routePath string) string

// ✅ Extracts component fields from Plenti structure
func ExtractComponentFields(data map[string]interface{}, componentName string) map[string]interface{}
```

**Status:** ✅ **PRODUCTION READY**
**Compatibility:** 100% compatible with Plenti's JSON structure

#### 2. Parser (`parser/*.go`)
```go
// ✅ Parses {for} loops with variables
{for component in components}
  <Component:dynamic name={component.name} />
{/for}

// ✅ Handles expressions
{component.name}
{content.title}
{$store.selected}

// ✅ Scope tracking exists
transformer/scope.go
```

**Status:** ✅ **FUNCTIONAL**
**Needs:** Formalize build vs runtime resolution rules

#### 3. Component Registry (`transformer/components.go`)
```go
var componentTemplateRegistry = make(map[string]*ComponentTemplate)

func RegisterComponentTemplate(name string, template *ComponentTemplate)
func GetComponentTemplate(name string) (*ComponentTemplate, bool)
```

**Status:** ✅ **EXISTS**
**Needs:** Generate JavaScript functions, create manifest.json

#### 4. JavaScript Formatting (`transformer/alpine.go`, `utils/utils.go`)
```go
// ✅ RECENTLY FIXED!
func FormatGoValueToJS(value interface{}) string {
    // Handles map[string]interface{} → JavaScript objects
    // Escapes newlines, special characters
    // No more map[...] Go format in output!
}

func MakeGetter(comp_data map[string]any) string {
    // Uses AnyToJSValue for proper formatting
}
```

**Status:** ✅ **FIXED (October 13, 2025)**
**Achievement:** Resolved all JavaScript syntax errors

#### 5. Server (`cmd/server/main.go`)
```go
// ✅ Serves templates
// ✅ Loads JSON content
// ✅ Injects props into components
// ✅ Renders HTML with Alpine.js directives
```

**Status:** ✅ **FUNCTIONAL**
**Needs:** Batch processing for static builds

### ⚠️ **What's Commented Out (But Fixed!)**

#### Magic Variables in `layouts/global/html.html`
```html
<Component:dynamic name={layout}
    {...content.fields}
    bind:shadowContent={shadowContent} />
    <!-- TEMPORARILY DISABLED: Magic variables causing JS syntax errors
    {allContent}
    {allLayouts}
    {content}
    {env}
    -->
```

**WHY DISABLED:** JavaScript `map[...]` format causing syntax errors
**NOW FIXED:** We fixed the formatting bugs on October 13!
**ACTION NEEDED:** Uncomment and test (should work now)

### ❌ **What's Missing**

1. **Deterministic signatures** - No hash-based validation
2. **Client hydration runtime** - No `$hydrate()` or `$renderDynamic()`
3. **Registry code generation** - No JavaScript function generation from templates
4. **Registry manifest** - No manifest.json with versioning
5. **Scope analyzer** - No `IsBuildResolvable()` decision logic
6. **Build pipeline** - No batch processing of content directory
7. **Props serialization** - No threshold logic (inline/ref/url)
8. **Merge plans** - No spread operator precedence implementation
9. **Circuit breakers** - No failure handling for registry loads
10. **Development overlay** - No component inspector UI

---

## Alignment Assessment

### 🎯 **Perfect Alignment: Three Specs Converge**

#### Spec 1: Hybrid SSR-Hydration Approach
**Source:** `.agent-os/specs/2025-10-12-dynamic-component-iteration/hybrid-ssr-hydration-approach.md`

**Key Concept:**
- SSR renders initial HTML (fast first paint, SEO)
- Embed `data-comp` and `data-props` descriptors
- Client hydrates with signature validation
- Runtime updates without page reload

#### Spec 2: V4.0 SSG Blueprint
**Source:** `docs/plenti/v4-ssg-hydration-blueprint.md`

**Key Concept:**
- Build-time resolution when possible (static optimization)
- Runtime resolution when necessary (loop variables, stores)
- Deterministic signatures prevent hydration mismatches
- Component registry with tree-shaking

#### Spec 3: V4.0 Blueprint Rationale
**Source:** `docs/plenti/v4-blueprint-rationale.md`

**Key Concept:**
- Why pure client-side won't work (SEO, performance)
- Why pure server-side isn't enough (no dynamic updates)
- Why hybrid is the ONLY solution (best of both worlds)

### 📊 **Alignment Matrix**

| Feature | Hybrid Spec | V4.0 Blueprint | Current Code | Status |
|---------|-------------|----------------|--------------|--------|
| **SSR First** | ✅ | ✅ | ✅ | Ready |
| **Runtime Hydration** | ✅ | ✅ | ❌ | Need to build |
| **Signature Validation** | ✅ | ✅ | ❌ | Need to build |
| **Component Registry** | ✅ | ✅ | 🟡 | Exists, needs JS codegen |
| **Scope Analysis** | ✅ | ✅ | 🟡 | Partial, needs `IsBuildResolvable()` |
| **Props Merging** | ✅ | ✅ | ❌ | Need to build |
| **Magic Variables** | 🟡 | ✅ | 🟡 | Commented out, but fixed |
| **Build Pipeline** | N/A | ✅ | ❌ | Need to build |
| **Tree-shaking** | ✅ | ✅ | ❌ | Need to build |
| **Circuit Breakers** | N/A | ✅ | ❌ | Need to build |

**Legend:**
- ✅ = Implemented and working
- 🟡 = Partially implemented or disabled
- ❌ = Not implemented
- N/A = Not mentioned in spec

---

## Gap Analysis

### **Critical Gaps (Must Have for MVP)**

#### 1. Deterministic Signatures (PRIORITY: CRITICAL)
**Problem:** No way to validate SSR vs client rendering matches

**V4.0 Blueprint (Lines 246-311):**
```go
func Canonicalize(v interface{}) interface{} {
    // Sorts map keys for consistent JSON
    // Filters __proto__ and constructor
    // Recursively processes nested structures
}

func Signature(name string, props interface{}) string {
    canonical, err := CanonicalJSON(props)
    h := fnv.New64a()
    h.Write([]byte(name))
    h.Write(canonical)
    return fmt.Sprintf("%s-%016x", name, h.Sum64())
}
```

**Impact:** Without this, hydration mismatches are invisible (React-style bugs)

**Effort:** 8-10 hours

---

#### 2. Client Runtime (PRIORITY: CRITICAL)
**Problem:** No client-side code to hydrate or update components

**V4.0 Blueprint (Lines 413-700):**
```javascript
window.$componentRegistry = {};
window.$loadRegistry = async function(key) { /* ... */ }
window.$renderDynamic = async function(el, name, props) { /* ... */ }
window.$hydrate = function(el) { /* ... */ }
```

**Features Needed:**
- ✅ Registry loading with promise deduplication
- ✅ Component rendering with error handling
- ✅ Hydration with signature validation
- ✅ Circuit breaker for failed components
- ✅ Exponential backoff for retries

**Effort:** 12-15 hours

---

#### 3. Scope Analyzer Enhancement (PRIORITY: HIGH)
**Problem:** Can't decide build-time vs runtime resolution

**V4.0 Resolution Rules (Lines 62-82):**
```
BUILD-RESOLVABLE:
  - "Header"                 (literal string)
  - content.title            (content field)
  - components[0].name       (known array index)

RUNTIME-ONLY:
  - component.name           (loop variable)
  - $store.selected          (Alpine store)
  - content.title + "_suffix" (expression with operators)
```

**Implementation:**
```go
type ScopeAnalyzer struct {
    buildVars   map[string]bool  // content, page, front-matter
    runtimeVars map[string]bool  // loop vars, stores
}

func (sa *ScopeAnalyzer) IsBuildResolvable(expr ast.Expr) bool {
    // Rule 1: Literals → build-time
    // Rule 2: Check all identifiers
    // Rule 3: Operators → runtime (for safety)
}
```

**Effort:** 6-8 hours

---

#### 4. Registry Code Generation (PRIORITY: HIGH)
**Problem:** Component templates need to become JavaScript functions

**Example Transformation:**
```html
<!-- Input: layouts/components/hero.html -->
<section id="hero">
  <h1>{title}</h1>
  <p>{description}</p>
</section>
```

```javascript
// Output: registry.main.[hash].js
export function Hero(props, { esc }) {
    return `
<section id="hero">
  <h1>${esc(props.title)}</h1>
  <p>${esc(props.description)}</p>
</section>
    `.trim();
}
```

**Effort:** 12-15 hours

---

### **High Priority Gaps (Important for Production)**

#### 5. Registry Manifest (PRIORITY: HIGH)
**File:** `public/js/registry.manifest.json`

**Format:**
```json
{
  "version": "abc123def456",
  "main": "/js/registry.main.abc123de.js",
  "components": {
    "RareComponent": "/js/registry.RareComponent.abc123de.js"
  }
}
```

**Benefits:**
- Cache busting (version in filename)
- Code splitting (main chunk + lazy chunks)
- Tree-shaking (only ship used components)

**Effort:** 8-10 hours

---

#### 6. Build Pipeline (PRIORITY: HIGH)
**Problem:** No batch processing for static site generation

**Plenti Currently Does (cmd/build/data_source.go):**
```go
for _, currentContent := range allContent {
    err := createProps(currentContent, allContentStr, env)
    err = createHTML(currentContent, env)
    // Write HTML files
}
```

**We Need:**
```go
func RenderAllContent(contentDir, buildDir string) error {
    allContent := loadAllContent(contentDir)
    env := buildEnvVars()

    for _, content := range allContent {
        html, err := renderTemplate(content, allContent, env)
        outputPath := filepath.Join(buildDir, content["path"], "index.html")
        os.WriteFile(outputPath, []byte(html), 0644)
    }
}
```

**Effort:** 10-12 hours

---

#### 7. Props Serialization Strategy (PRIORITY: MEDIUM)
**V4.0 Approach (Lines 763-790):**

```go
func (r *Renderer) SerializePropsAttr(name string, props map[string]interface{}) string {
    propsJSON, _ := json.Marshal(filtered)

    // Small props: inline in data-props
    if len(propsJSON) <= CONFIG.LARGE_PROPS_THRESHOLD {
        return fmt.Sprintf(`data-props='%s'`, html.EscapeString(string(propsJSON)))
    }

    // Large props: store in registry, reference by ID
    id := GeneratePropsID()
    r.RegisterLargeProps(id, props)
    return fmt.Sprintf(`data-props-ref="%s"`, id)

    // Huge props: URL params (edge case)
    if len(propsJSON) > CONFIG.HUGE_PROPS_THRESHOLD {
        url := r.SerializePropsToURL(name, props)
        return fmt.Sprintf(`data-props-url="%s"`, url)
    }
}
```

**Effort:** 6-8 hours

---

#### 8. Merge Plans (PRIORITY: MEDIUM)
**Problem:** Spread operator precedence not implemented

**V4.0 Precedence (Lines 89-96):**
```
1. Left-to-right spreads    {...spread1} {...spread2}
2. Left-to-right regular    prop1="a" prop2="b"
3. Later overrides earlier  (rightmost wins)
```

**Example:**
```html
<Component:dynamic
    name="Card"
    {...defaults}
    title="Override"
    {...overrides}
    class="final" />

<!-- Evaluation order:
  1. defaults = {title: "Default", class: "default"}
  2. title="Override" → {title: "Override"}
  3. overrides = {class: "override"}
  4. class="final" → {class: "final"}
  Result: {title: "Override", class: "final"}
-->
```

**Effort:** 8-10 hours

---

### **Medium Priority Gaps (Nice to Have)**

#### 9. Circuit Breakers (PRIORITY: MEDIUM)
**Problem:** Failed component loads can cascade

**V4.0 Implementation (Lines 413-452):**
```javascript
const failedComponents = new Map();
const CONFIG = { CIRCUIT_BREAKER_THRESHOLD: 3 };

window.$renderDynamic = async function(el, name, props, options = {}) {
    const failures = failedComponents.get(name) || 0;
    if (failures >= CONFIG.CIRCUIT_BREAKER_THRESHOLD) {
        el.innerHTML = `<!-- Component ${name} disabled -->`;
        return;
    }

    try {
        // Attempt render
        failedComponents.delete(name); // Success!
    } catch (err) {
        failedComponents.set(name, failures + 1);

        if (options.retry < CONFIG.MAX_RETRIES) {
            const delay = CONFIG.INITIAL_RETRY_DELAY * Math.pow(2, options.retry);
            await new Promise(resolve => setTimeout(resolve, delay));
            return window.$renderDynamic(el, name, props, { retry: options.retry + 1 });
        }
    }
};
```

**Effort:** 6-8 hours

---

#### 10. Development Overlay (PRIORITY: LOW)
**V4.0 Spec (Lines 878-923):**

**Features:**
- Component inspector (name, props, signature)
- Hydration status indicator
- Performance metrics
- Error visualization

**Effort:** 8-10 hours

---

## Implementation Phases

### **Phase 1: Core Hydration (Weeks 1-2)** ⏰ **40-50 hours**

**Goal:** Enable basic SSR + hydration with signature validation

#### Task 1.1: Deterministic Signatures (8-10h)
**Files to Create:**
- `pkg/signature/canonical.go`
- `pkg/signature/hash.go`
- `pkg/signature/signature.go`
- `pkg/signature/signature_test.go`
- `assets/js/canonical.js`

**Implementation Checklist:**
- [ ] Implement `Canonicalize()` in Go
- [ ] Implement FNV-1a hashing
- [ ] Create `Signature(name, props)` function
- [ ] Mirror in JavaScript (`canonical.js`)
- [ ] Write determinism tests (same props → same sig)
- [ ] Test with edge cases (nested objects, arrays, nulls)
- [ ] Benchmark performance (<1ms for typical props)

**Success Criteria:**
```go
props1 := map[string]interface{}{
    "b": 2,
    "a": map[string]interface{}{"y": 1, "x": 2},
}
props2 := map[string]interface{}{
    "a": map[string]interface{}{"x": 2, "y": 1},
    "b": 2,
}
sig1 := signature.Signature("Hero", props1)
sig2 := signature.Signature("Hero", props2)
assert.Equal(t, sig1, sig2) // MUST match!
```

---

#### Task 1.2: Scope Analyzer Enhancement (6-8h)
**Files to Modify/Create:**
- `analyzer/scope.go` (create)
- `transformer/scope.go` (enhance)
- `analyzer/scope_test.go`

**Implementation Checklist:**
- [ ] Create `ScopeAnalyzer` struct
- [ ] Implement `IsBuildResolvable(expr)` logic
- [ ] Implement `ExtractIdentifiers(expr)` helper
- [ ] Implement `HasOperators(expr)` helper
- [ ] Wire into transformer decision point
- [ ] Test with literal strings → build-time
- [ ] Test with loop vars → runtime
- [ ] Test with stores ($) → runtime
- [ ] Test with operators → runtime
- [ ] Test with content fields → build-time

**Success Criteria:**
```go
analyzer := NewScopeAnalyzer()
analyzer.runtimeVars["component"] = true

assert.True(t, analyzer.IsBuildResolvable("Header"))        // literal
assert.False(t, analyzer.IsBuildResolvable("component.name")) // loop var
assert.False(t, analyzer.IsBuildResolvable("$store.theme"))  // Alpine store
assert.True(t, analyzer.IsBuildResolvable("content.title"))  // content field
```

---

#### Task 1.3: Client Runtime (12-15h)
**Files to Create:**
- `assets/js/runtime.js`
- `assets/js/hydration.js`
- `assets/js/registry.js`
- `assets/js/utils.js` (esc function)

**Implementation Checklist:**
- [ ] Implement `$loadRegistry(key)` with promise deduplication
- [ ] Implement `$renderDynamic(el, name, props)` with error handling
- [ ] Implement `$hydrate(el)` with signature validation
- [ ] Add circuit breaker logic
- [ ] Add exponential backoff for retries
- [ ] Implement `esc()` helper (HTML escape)
- [ ] Add dev mode logging
- [ ] Test registry loading
- [ ] Test component rendering
- [ ] Test hydration skipping (matching sigs)
- [ ] Test hydration re-rendering (mismatched sigs)

**Success Criteria:**
```javascript
// Load registry
await window.$loadRegistry('main');
assert.ok(window.$componentRegistry.Hero);

// Render component
const el = document.createElement('div');
await window.$renderDynamic(el, 'Hero', {title: 'Test'});
assert.includes(el.innerHTML, '<h1>Test</h1>');

// Hydration with matching sig
el.dataset.sig = 'Hero-abc123';
el.dataset.comp = 'Hero';
window.$hydrate(el);
// Should NOT re-render (sig matches)
```

---

#### Task 1.4: SSR Renderer Updates (8-10h)
**Files to Modify/Create:**
- `renderer/dynamic_component.go` (create)
- `renderer/render.go` (enhance)
- `renderer/props.go` (create)

**Implementation Checklist:**
- [ ] Create `RenderDynamicComponent()` function
- [ ] Integrate `ScopeAnalyzer.IsBuildResolvable()`
- [ ] Implement build-time resolution path
- [ ] Implement runtime wrapper emission
- [ ] Add signature generation
- [ ] Add props serialization
- [ ] Emit wrapper with `data-comp`, `data-props`, `data-sig`
- [ ] Test with static name → full HTML
- [ ] Test with runtime name → wrapper only
- [ ] Test props serialization

**Success Criteria:**
```go
// Static name (build-time resolution)
html := renderer.RenderDynamicComponent(node, scope)
assert.Contains(html, `data-comp="Hero"`)
assert.Contains(html, `data-sig="Hero-abc123"`)
assert.Contains(html, `<section id="hero">`) // SSR content

// Runtime name (runtime resolution)
html := renderer.RenderDynamicComponent(nodeWithLoopVar, scope)
assert.Contains(html, `data-comp=""`) // Empty, will be filled at runtime
assert.Contains(html, `x-data="{compName:component.name}"`)
```

---

#### Task 1.5: Integration Testing (6-8h)
**Files to Create:**
- `tests/hydration/basic_test.go`
- `tests/hydration/loop_test.go`
- `tests/hydration/signature_test.go`

**Test Scenarios:**
- [ ] Static component renders with SSR HTML
- [ ] Loop with dynamic components emits runtime wrappers
- [ ] Signatures match between Go and JavaScript
- [ ] Hydration skips re-render when signatures match
- [ ] Hydration re-renders when signatures mismatch
- [ ] Missing components show error message
- [ ] Failed component loads trigger circuit breaker

---

### **Phase 2: Registry & Codegen (Weeks 3-4)** ⏰ **30-40 hours**

**Goal:** Generate JavaScript component functions and manifest

#### Task 2.1: Registry Code Generation (12-15h)
**Files to Create:**
- `builder/registry_codegen.go`
- `builder/template_to_js.go`
- `builder/component_collector.go`

**Implementation Checklist:**
- [ ] Walk `layouts/components/` directory
- [ ] Parse each component template
- [ ] Convert AST to JavaScript template literal
- [ ] Handle expressions: `{var}` → `${esc(props.var)}`
- [ ] Handle attributes: `attr={var}` → `attr="${esc(props.var)}"`
- [ ] Handle conditionals: `{if}` → ternary or if-block
- [ ] Handle loops: `{for}` → `.map()`
- [ ] Generate named export functions
- [ ] Test with simple component (no expressions)
- [ ] Test with component containing expressions
- [ ] Test with component containing loops
- [ ] Test with nested components

**Success Criteria:**
```go
// Input: hero.html
// ---
// export let title, description
// ---
// <section id="hero">
//   <h1>{title}</h1>
//   <p>{description}</p>
// </section>

js := builder.GenerateRegistryJS(heroTemplate)

// Output: registry.main.js
// export function Hero(props, { esc }) {
//     return `
// <section id="hero">
//   <h1>${esc(props.title)}</h1>
//   <p>${esc(props.description)}</p>
// </section>
//     `.trim();
// }
```

---

#### Task 2.2: Registry Manifest (8-10h)
**Files to Create:**
- `builder/registry_manifest.go`
- `builder/chunk_splitter.go`
- `builder/usage_analyzer.go`

**Implementation Checklist:**
- [ ] Analyze component usage frequency
- [ ] Split into main chunk (frequent) and lazy chunks (rare)
- [ ] Generate main chunk with most-used components
- [ ] Generate individual lazy chunks for rare components
- [ ] Compute build hash (for versioning)
- [ ] Write manifest.json with paths
- [ ] Add cache headers for immutable chunks
- [ ] Test with single component
- [ ] Test with multiple components
- [ ] Test with unused components (should be excluded)

**Success Criteria:**
```json
{
  "version": "abc123def456",
  "main": "/js/registry.main.abc123de.js",
  "components": {
    "RareComponent": "/js/registry.RareComponent.abc123de.js"
  }
}
```

---

#### Task 2.3: Tree-Shaking (6-8h)
**Files to Create:**
- `builder/tree_shaker.go`
- `builder/dependency_graph.go`

**Implementation Checklist:**
- [ ] Scan all templates for component usage
- [ ] Build dependency graph (which components use which)
- [ ] Mark used components (transitive closure)
- [ ] Exclude unused components from registry
- [ ] Test with unused component → excluded
- [ ] Test with deeply nested dependencies → included
- [ ] Generate usage report

**Success Criteria:**
```
Analyzed 50 components:
- 15 used directly
- 8 used transitively (dependencies)
- 27 unused (excluded from registry)

Registry size: 125 KB (was 450 KB)
Savings: 72%
```

---

#### Task 2.4: Integration Testing (4-6h)
**Test Scenarios:**
- [ ] Build generates registry JavaScript files
- [ ] Manifest.json has correct paths and version
- [ ] Main chunk loads on page init
- [ ] Lazy chunk loads on demand
- [ ] Unused components not in any chunk
- [ ] Cache busting works (version in filename)

---

### **Phase 3: Build Pipeline (Weeks 5-6)** ⏰ **25-30 hours**

**Goal:** Enable static site generation (batch processing)

#### Task 3.1: Batch Content Processing (10-12h)
**Files to Create:**
- `cmd/build/render_all.go`
- `cmd/build/parallel.go`
- `cmd/build/progress.go`

**Implementation Checklist:**
- [ ] Walk `content/` directory
- [ ] Load all JSON files
- [ ] Build `allContent` array
- [ ] Build `env` object
- [ ] Render each template with magic vars
- [ ] Write HTML files to `public/`
- [ ] Show progress bar
- [ ] Support parallel rendering
- [ ] Handle errors gracefully
- [ ] Generate build report

**Success Criteria:**
```bash
$ go run cmd/build/main.go

Building site...
├── Loading content... [50 files]
├── Rendering templates...
│   ├── / (index) ✓
│   ├── /about ✓
│   ├── /blog/post-1 ✓
│   └── ... (47 more)
├── Generating registry...
│   ├── main chunk (15 components) ✓
│   └── lazy chunks (3 components) ✓
└── Done! (3.2s)

Output: public/
  50 pages rendered
  125 KB registry (72% savings)
```

---

#### Task 3.2: Magic Variables System (8-10h)
**Files to Modify/Create:**
- `cmd/server/magic_vars.go` (create)
- `cmd/build/magic_vars.go` (create)
- `loader/all_content.go` (create)

**Implementation Checklist:**
- [ ] Implement `loadAllContent()` (scan content dir)
- [ ] Implement `buildEnvVars()` (from config)
- [ ] Implement `buildMagicVars()` (combine all)
- [ ] Uncomment magic vars in `html.html`
- [ ] Test with `{allContent}`
- [ ] Test with `{allLayouts}`
- [ ] Test with `{content}`
- [ ] Test with `{env}`
- [ ] Verify no JavaScript errors

**Success Criteria:**
```html
<!-- html.html (uncommented) -->
<Component:dynamic name={layout}
    {...content.fields}
    {allContent}    <!-- ✅ WORKS NOW -->
    {allLayouts}    <!-- ✅ WORKS NOW -->
    {content}       <!-- ✅ WORKS NOW -->
    {env}           <!-- ✅ WORKS NOW -->
    bind:shadowContent={shadowContent} />
```

---

#### Task 3.3: Plenti Integration (7-9h)
**Files to Modify:**
- `plenti/cmd/build/data_source.go`
- `plenti/cmd/build/compile.go`

**Implementation Checklist:**
- [ ] Replace `createProps()` with Go renderer
- [ ] Replace `createHTML()` with Go renderer
- [ ] Remove V8 dependency (SSRctx.RunScript)
- [ ] Update imports
- [ ] Test build with Plenti CLI
- [ ] Verify output matches old behavior
- [ ] Performance comparison (Go vs V8)

**Success Criteria:**
```bash
$ plenti build

Building site...
✓ Content loaded (50 files)
✓ Templates rendered (using Go engine)
✓ Registry generated
✓ Build complete (2.1s)

Previous build (V8): 5.4s
New build (Go): 2.1s
Speedup: 2.6x faster!
```

---

### **Phase 4: Production Hardening (Weeks 7-8)** ⏰ **20-25 hours**

**Goal:** Add production-ready features

#### Task 4.1: Props Serialization Strategy (6-8h)
**Implementation Checklist:**
- [ ] Implement threshold logic (inline/ref/url)
- [ ] Store large props in registry
- [ ] Generate props URLs for huge data
- [ ] Test with small props → inline
- [ ] Test with large props → ref
- [ ] Test with huge props → url

---

#### Task 4.2: Merge Plans (8-10h)
**Implementation Checklist:**
- [ ] Parse spread operators
- [ ] Implement precedence rules
- [ ] Generate merge plan AST
- [ ] Execute in SSR
- [ ] Execute in client runtime
- [ ] Test precedence order

---

#### Task 4.3: Development Overlay (8-10h)
**Implementation Checklist:**
- [ ] Create component inspector UI
- [ ] Show name, props, signature
- [ ] Show hydration status
- [ ] Show performance metrics
- [ ] Add hotkey to toggle (Ctrl+Shift+D)

---

#### Task 4.4: Circuit Breakers (6-8h)
**Implementation Checklist:**
- [ ] Implement failure tracking
- [ ] Implement threshold logic
- [ ] Implement exponential backoff
- [ ] Add error reporting
- [ ] Test with failing component

---

## Critical Path

### **Minimum Viable Product (MVP)** - Weeks 1-4

**Goal:** Basic hydration working end-to-end

1. ✅ **Week 1:** Deterministic signatures + scope analyzer
2. ✅ **Week 1-2:** Client runtime + SSR renderer
3. ✅ **Week 3:** Registry code generation
4. ✅ **Week 4:** Registry manifest + integration testing

**Deliverable:** Component loops work with SSR + hydration

---

### **Production Ready** - Weeks 5-8

**Goal:** Static builds + Plenti integration

5. ✅ **Week 5:** Build pipeline + magic variables
6. ✅ **Week 6:** Plenti integration + performance testing
7. ✅ **Week 7:** Props serialization + merge plans
8. ✅ **Week 8:** Dev overlay + circuit breakers + final testing

**Deliverable:** Full Plenti v4.0 replacement

---

## Risk Assessment

### **High Risk Items**

#### 1. JavaScript Codegen Complexity
**Risk:** Converting Go AST to JavaScript template literals may miss edge cases

**Mitigation:**
- Start with simple components (no nesting)
- Add complexity incrementally
- Use comprehensive test suite
- Compare output with manual templates

**Contingency:** Use simpler "innerHTML assignment" approach instead of template literals

---

#### 2. Signature Mismatch Between Go/JS
**Risk:** Different canonicalization in Go vs JavaScript leads to different signatures

**Mitigation:**
- Mirror algorithms exactly (copy logic)
- Share test vectors between Go and JS tests
- Use well-defined hash algorithm (FNV-1a)

**Contingency:** Add "trust SSR" mode (skip validation in production)

---

#### 3. Performance Regression
**Risk:** Go renderer is slower than V8 SSR

**Mitigation:**
- Benchmark early and often
- Profile bottlenecks
- Optimize hot paths
- Use parallel rendering for builds

**Contingency:** Hybrid approach (Go for simple pages, V8 for complex)

---

### **Medium Risk Items**

#### 4. Breaking Changes in Plenti
**Risk:** Plenti v4.0 changes incompatible with current spec

**Mitigation:**
- Version lock Plenti dependency
- Test with real Plenti sites
- Create migration guide

---

#### 5. Registry Size Explosion
**Risk:** All components in registry causes large bundle

**Mitigation:**
- Tree-shaking (remove unused)
- Code splitting (lazy chunks)
- Component allowlist for critical pages

---

## Success Metrics

### **Phase 1 Success Criteria**
- [ ] Signatures are deterministic (Go and JS match)
- [ ] Hydration works with signature validation
- [ ] Static names render with SSR HTML
- [ ] Loop variables render with runtime wrappers
- [ ] Zero console errors in browser
- [ ] All integration tests pass

---

### **Phase 2 Success Criteria**
- [ ] Registry JavaScript is valid and runnable
- [ ] Manifest.json loads correctly
- [ ] Tree-shaking reduces bundle size by >50%
- [ ] Components render correctly from registry
- [ ] Lazy loading works on demand

---

### **Phase 3 Success Criteria**
- [ ] Static build generates all HTML files
- [ ] Magic variables work without JavaScript errors
- [ ] Plenti build works with Go renderer
- [ ] Performance is ≥2x faster than V8 approach
- [ ] All Plenti starter sites build successfully

---

### **Phase 4 Success Criteria**
- [ ] Props serialization handles all edge cases
- [ ] Merge plans respect precedence correctly
- [ ] Dev overlay provides useful debugging info
- [ ] Circuit breakers prevent cascading failures
- [ ] Production builds are stable and reliable

---

## Next Steps

### **Immediate Actions (This Week)**

#### 1. Quick Win: Test Magic Variables (1-2 hours)
**Why:** We fixed the JS formatting bugs, so magic variables SHOULD work now

**Steps:**
```bash
cd /Users/benjaminwaller/Projects/Jim\ Fisk/custom_go_template

# 1. Uncomment magic variables in html.html
# 2. Restart server
go run cmd/server/main.go

# 3. Test in browser
curl http://localhost:3333/
# Look for JavaScript errors in console
```

**Expected Outcome:**
- ✅ No `map[...]` in HTML source
- ✅ No JavaScript console errors
- ✅ Magic variables available in templates

**If Successful:** You have a working system RIGHT NOW!

---

#### 2. Decision Point: MVP vs Full Implementation
**Option A: MVP (Weeks 1-4)**
- Focus on core hydration only
- Get component loops working
- Skip build pipeline for now

**Option B: Full Implementation (Weeks 1-8)**
- Complete all phases
- Full Plenti integration
- Production-ready features

**Recommendation:** Start with MVP, assess, then continue to full implementation

---

### **Long-Term Roadmap**

#### Q1 2026: v4.0 MVP (Weeks 1-4)
- Core hydration working
- Component loops functional
- Basic registry

#### Q2 2026: v4.0 Production (Weeks 5-8)
- Build pipeline complete
- Plenti fully integrated
- Performance optimized

#### Q3 2026: v4.1 Enhancements
- Advanced features (dev overlay, circuit breakers)
- Performance monitoring
- Documentation complete

---

## Appendix: Technical Details

### **A. Canonical JSON Algorithm**

```go
func Canonicalize(v interface{}) interface{} {
    switch val := v.(type) {
    case nil:
        return nil
    case bool, string, float64, int, int64:
        return val
    case map[string]interface{}:
        // 1. Filter dangerous keys
        keys := []string{}
        for k := range val {
            if k != "__proto__" && k != "constructor" {
                keys = append(keys, k)
            }
        }
        // 2. Sort keys
        sort.Strings(keys)
        // 3. Recursively canonicalize values
        result := make(map[string]interface{}, len(keys))
        for _, k := range keys {
            result[k] = Canonicalize(val[k])
        }
        return result
    case []interface{}:
        result := make([]interface{}, len(val))
        for i, elem := range val {
            result[i] = Canonicalize(elem)
        }
        return result
    default:
        return fmt.Sprintf("%v", val)
    }
}
```

### **B. FNV-1a Hashing**

```go
func fnv1a(data []byte) uint64 {
    const offset = 14695981039346656037
    const prime = 1099511628211

    hash := offset
    for _, b := range data {
        hash ^= uint64(b)
        hash *= prime
    }
    return hash
}

func Signature(name string, props interface{}) string {
    canonical := Canonicalize(props)
    jsonBytes, _ := json.Marshal(canonical)

    h := fnv1a(append([]byte(name), jsonBytes...))
    return fmt.Sprintf("%s-%016x", name, h)
}
```

### **C. Build vs Runtime Resolution**

```
DECISION TREE:

Is name a literal string? (e.g., "Header")
├─ YES → Build-time resolution
└─ NO → Is expression build-resolvable?
    ├─ Are all identifiers in buildVars? (e.g., content.title)
    │   ├─ YES → Does expression have operators? (e.g., title + "_suffix")
    │   │   ├─ YES → Runtime resolution (for safety)
    │   │   └─ NO → Build-time resolution
    │   └─ NO → Are any identifiers runtime-only? (e.g., component.name, $store.*)
    │       ├─ YES → Runtime resolution
    │       └─ NO → Build-time resolution
```

### **D. Props Serialization Thresholds**

```go
const (
    SMALL_PROPS_THRESHOLD  = 2048   // 2 KB - inline in data-props
    LARGE_PROPS_THRESHOLD  = 102400 // 100 KB - store in registry, reference by ID
    HUGE_PROPS_THRESHOLD   = 1048576 // 1 MB - serialize to URL params
)

func SerializeProps(name string, props map[string]interface{}) string {
    jsonBytes, _ := json.Marshal(props)
    size := len(jsonBytes)

    if size <= SMALL_PROPS_THRESHOLD {
        // Inline in attribute
        return fmt.Sprintf(`data-props='%s'`, html.EscapeString(string(jsonBytes)))
    } else if size <= LARGE_PROPS_THRESHOLD {
        // Store in registry, reference by ID
        id := GeneratePropsID()
        propsRegistry[id] = props
        return fmt.Sprintf(`data-props-ref="%s"`, id)
    } else {
        // Serialize to URL (edge case)
        url := SerializePropsToURL(name, props)
        return fmt.Sprintf(`data-props-url="%s"`, url)
    }
}
```

---

## Conclusion

The v4.0 SSG + Hydration blueprint is **perfectly aligned** with the custom Go template engine's current capabilities. The foundation exists; we need to build the hydration layer on top.

**Key Insights:**

1. ✅ **JSON handling is compatible** with Plenti's format
2. ✅ **JavaScript formatting bugs are fixed** (October 13, 2025)
3. ✅ **Parser and transformer ready** for hybrid approach
4. ❌ **Hydration layer missing** (signatures, runtime, registry)
5. ❌ **Build pipeline missing** (static generation)

**Recommended Path:**

1. **Test magic variables** (1-2 hours) - they should work now!
2. **Implement Phase 1** (2 weeks) - core hydration
3. **Assess and continue** to Phases 2-4 (6 weeks) - production features

**Total Effort:** 115-145 hours over 8 weeks for full v4.0 implementation

The blueprint is solid. The path is clear. Let's build it! 🚀

---

**Document Version:** 1.0
**Last Updated:** October 14, 2025
**Next Review:** After Phase 1 completion
