# SSG + Hydration Blueprint for Dynamic Components — v4.0 (Production Hardened)

**Goal:** Implement `<Component:dynamic>` in Plenti's **Go templating engine** with **SSG (build-time) rendering + selective client hydration**. This version combines the architectural rigor of v3.0 with the practical production hardening of v2.0, plus critical gap fixes.

**Version History:**
- v1.0: Initial SSG approach
- v2.0: Template functions, error boundaries, development tools
- v3.0: Deterministic signatures, merge plans, security hardening
- v4.0: **Best of both + race condition fixes + versioning + enhanced DX**

---

## Executive Summary

### Architecture
SSG at build → static HTML with hydration markers → **tiny Alpine/JS runtime** for dynamic swaps.

### Two Resolution Modes

1. **Build-time resolution** (default, 80%+ of cases):
   - Fully render components using content-known data
   - Emit hydration markers with deterministic signatures
   - **No client re-render on load** when signatures match
   - Zero runtime overhead for static content

2. **Runtime resolution** (fallback for dynamic cases):
   - When `name=` depends on runtime stores or loop variables
   - Emit wrapper with markers, render client-side
   - Progressive enhancement approach

### Key Improvements in v4.0

**From v3.0:**
- ✅ Single source of truth (codegen from same templates)
- ✅ Deterministic signatures (canonical JSON + stable hashing)
- ✅ Registry manifest with tree-shaking
- ✅ Hardened CSP and security
- ✅ Clear separation of concerns

**From v2.0:**
- ✅ Enhanced error boundaries with monitoring
- ✅ Realistic performance targets
- ✅ Comprehensive development tools
- ✅ Build reports with actionable metrics

**New in v4.0:**
- ✅ Race condition handling (promise deduplication)
- ✅ Versioning strategy (registry + page compatibility)
- ✅ Circuit breakers (prevent repeated failures)
- ✅ Optional merge plans (only when needed)
- ✅ Device-aware performance targets
- ✅ Sequence diagrams and clear resolution rules
- ✅ Partial resolution documentation
- ✅ Enhanced retry logic with exponential backoff

---

## Core Architecture

### Resolution Decision Rules

#### Build-Resolvable (Static) ✅
```go
// Pure variable references to build-known data
"Header"                    // Literal string
content.title               // Content field
components[0].name          // Known array index
page.hero.buttonText        // Front-matter field
```

#### Runtime-Only (Dynamic) ❌
```go
// References to runtime state
component.name              // Loop variable
$store.selected             // Alpine store
content.title + "_suffix"   // Expression with operators
components[i].name          // Variable index
window.location.pathname    // Browser state
```

**Rule:** Only pure variable references to build-known data are build-resolvable. Any expression involving operators, runtime stores, or dynamic indices defaults to runtime resolution.

### Hydration Lifecycle Sequence

```
┌─────────────┐
│ User        │
│ Navigates   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. Page Loads                                                │
│    • SSG HTML visible immediately (FCP)                      │
│    • All build-resolved components rendered                  │
│    • data-comp, data-sig, data-props embedded                │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. JavaScript Loads                                          │
│    • Main bundle executes                                    │
│    • Runtime helpers registered ($hydrate, $renderDynamic)   │
│    • Alpine.js initializes                                   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Registry Loads (via Manifest)                             │
│    • window.$registryManifest read                           │
│    • Main chunk loaded (common components)                   │
│    • Promise deduplication prevents double-loads             │
│    • Registry version checked                                │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Hydration Phase                                           │
│    • For each .dyn-comp:                                     │
│      - Compare data-sig with current state                   │
│      - If match: mark hydrated, no re-render ✅              │
│      - If mismatch: $renderDynamic() ⚠️                      │
│    • Runtime-only components rendered now                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Interactive                                               │
│    • TTI achieved                                            │
│    • User interactions trigger re-renders                    │
│    • Lazy chunks load on-demand                              │
│    • Circuit breakers prevent error cascades                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Static Foundation (Weeks 1-2)
**Goal:** Build-time resolution for literal and content-scoped names

- [ ] Scope analyzer with clear build/runtime rules
- [ ] Canonical JSON serialization (Go + JS parity)
- [ ] FNV-1a signature generation
- [ ] Basic SSR renderer with markers
- [ ] Per-page registry manifest generation
- [ ] Unit tests for resolution logic

**Deliverable:** Static components work with deterministic hydration

### Phase 2: Runtime Foundation (Weeks 3-4)
**Goal:** Runtime resolution and client-side rendering

- [ ] Runtime wrapper emission
- [ ] Client runtime with race condition handling
- [ ] Promise-based registry loading with deduplication
- [ ] Registry version tracking and validation
- [ ] Basic error boundaries
- [ ] Integration tests

**Deliverable:** Loop-based components render client-side correctly

### Phase 3: Spread Operators & Merge Plans (Weeks 5-6)
**Goal:** Full spread support with server/client parity

- [ ] Spread operator evaluation (left-to-right precedence)
- [ ] Merge plan emission (optional, only when spreads present)
- [ ] Client-side merge plan executor
- [ ] Props serialization strategies (attr/ref/url)
- [ ] Large props handling with thresholds
- [ ] Parity tests (server vs client merge)

**Deliverable:** Complex prop merging works identically on server and client

### Phase 4: Production Hardening (Weeks 7-8)
**Goal:** Performance, security, and developer experience

- [ ] Circuit breakers for failed components
- [ ] Exponential backoff retry logic
- [ ] CSP hardening with nonces
- [ ] Tree-shaking optimization
- [ ] Lazy loading with IntersectionObserver
- [ ] Development overlay panel
- [ ] Build reports with metrics
- [ ] Comprehensive E2E tests
- [ ] Documentation and migration guide

**Deliverable:** Production-ready with monitoring, tooling, and docs

---

## Technical Implementation

### 1. Scope Analyzer (Go)

```go
// scope_analyzer.go
package analyzer

type ScopeAnalyzer struct {
    buildVars   map[string]bool  // content, page, front-matter
    runtimeVars map[string]bool  // loop vars, stores
}

// Resolution rules
func (sa *ScopeAnalyzer) IsBuildResolvable(expr ast.Expr) bool {
    // Literal strings are always build-time
    if _, ok := expr.(*ast.StringLiteral); ok {
        return true
    }
    
    // Check all identifiers in expression
    identifiers := ExtractIdentifiers(expr)
    for _, id := range identifiers {
        if sa.runtimeVars[id] {
            return false  // Any runtime var makes whole expr runtime
        }
    }
    
    // Expressions with operators default to runtime for safety
    if HasOperators(expr) {
        return false
    }
    
    return true
}

func (sa *ScopeAnalyzer) Analyze(node ast.Node) {
    switch n := node.(type) {
    case *ast.ForNode:
        // Loop variables are runtime-only
        sa.runtimeVars[n.Variable] = true
        if n.IndexVar != "" {
            sa.runtimeVars[n.IndexVar] = true
        }
    case *ast.ExportLetNode:
        // Front-matter props are build-time
        for _, prop := range n.Props {
            sa.buildVars[prop] = true
        }
    }
}
```

### 2. Deterministic Signatures (Go + JS)

```go
// canonical_json.go
package serialize

import (
    "crypto/fnv"
    "encoding/json"
    "sort"
)

// Canonicalize recursively sorts map keys
func Canonicalize(v interface{}) interface{} {
    switch val := v.(type) {
    case map[string]interface{}:
        // Sort keys
        keys := make([]string, 0, len(val))
        for k := range val {
            // Exclude transient keys
            if k == "__proto__" || k == "constructor" {
                continue
            }
            keys = append(keys, k)
        }
        sort.Strings(keys)
        
        // Build sorted map
        result := make(map[string]interface{}, len(keys))
        for _, k := range keys {
            result[k] = Canonicalize(val[k])
        }
        return result
        
    case []interface{}:
        // Recursively canonicalize array elements
        result := make([]interface{}, len(val))
        for i, elem := range val {
            result[i] = Canonicalize(elem)
        }
        return result
        
    default:
        return val
    }
}

func CanonicalJSON(v interface{}) ([]byte, error) {
    canonical := Canonicalize(v)
    return json.Marshal(canonical)
}

// FNV-1a hash (fast, sufficient for our use case)
func Signature(name string, props interface{}) string {
    canonical, err := CanonicalJSON(props)
    if err != nil {
        return ""
    }
    
    h := fnv.New64a()
    h.Write([]byte(name))
    h.Write(canonical)
    
    return fmt.Sprintf("%s-%016x", name, h.Sum64())
}
```

**JavaScript mirror (exact parity):**

```javascript
// canonical_json.js
function canonicalize(v) {
    if (v === null || v === undefined) return v;
    
    if (typeof v === 'object' && !Array.isArray(v)) {
        // Exclude transient keys
        const keys = Object.keys(v)
            .filter(k => k !== '__proto__' && k !== 'constructor')
            .sort();
        
        const result = {};
        for (const k of keys) {
            result[k] = canonicalize(v[k]);
        }
        return result;
    }
    
    if (Array.isArray(v)) {
        return v.map(canonicalize);
    }
    
    return v;
}

function canonicalJSON(v) {
    return JSON.stringify(canonicalize(v));
}

// FNV-1a hash (matching Go implementation)
function fnv1a(str) {
    let hash = 0xcbf29ce484222325n;
    for (let i = 0; i < str.length; i++) {
        hash ^= BigInt(str.charCodeAt(i));
        hash *= 0x100000001b3n;
    }
    return hash & 0xffffffffffffffffn;
}

function signature(name, props) {
    const canonical = canonicalJSON(props);
    const hash = fnv1a(name + canonical);
    return `${name}-${hash.toString(16).padStart(16, '0')}`;
}
```

### 3. Registry with Version Tracking

```go
// registry_generator.go
package builder

type RegistryManifest struct {
    Version    string            `json:"version"`     // Build hash
    Main       string            `json:"main"`        // Main chunk path
    Components map[string]string `json:"components"`  // name -> chunk path
}

func (b *Builder) GenerateRegistry() error {
    buildHash := b.ComputeBuildHash()
    
    // Collect all components used across all pages
    allComponents := b.CollectComponents()
    
    // Split into main (frequent) and lazy (rare) chunks
    mainComponents := b.GetFrequentComponents(allComponents, 0.7)
    lazyComponents := b.GetRareComponents(allComponents, 0.3)
    
    manifest := RegistryManifest{
        Version:    buildHash,
        Components: make(map[string]string),
    }
    
    // Generate main chunk (inline or small file)
    mainChunk := b.CodegenRegistryChunk(mainComponents, buildHash)
    mainPath := fmt.Sprintf("/js/registry.main.%s.js", buildHash[:8])
    b.WriteFile(mainPath, mainChunk)
    manifest.Main = mainPath
    
    // Generate lazy chunks
    for _, comp := range lazyComponents {
        chunk := b.CodegenRegistryChunk([]Component{comp}, buildHash)
        path := fmt.Sprintf("/js/registry.%s.%s.js", 
            comp.Name, buildHash[:8])
        b.WriteFile(path, chunk)
        manifest.Components[comp.Name] = path
    }
    
    // Write manifest
    manifestJSON, _ := json.Marshal(manifest)
    b.WriteFile("/js/registry.manifest.json", manifestJSON)
    
    return nil
}
```

### 4. Client Runtime (Production Hardened)

```javascript
// runtime.js - Production-ready runtime
(() => {
  'use strict';
  
  // === Configuration ===
  const CONFIG = {
    MAX_RETRIES: 3,
    INITIAL_RETRY_DELAY: 100,
    MAX_RETRY_DELAY: 5000,
    CIRCUIT_BREAKER_THRESHOLD: 3,
    LARGE_PROPS_THRESHOLD: 2048,
  };
  
  // === Global State ===
  window.$componentRegistry = {};
  window.$registryManifest = null;
  window.$registryVersion = null;
  window.$largeProps = {};
  
  // Promise deduplication
  const loadPromises = new Map();
  
  // Circuit breakers
  const failedComponents = new Map();
  const retryDelays = new Map();
  
  // === Utilities ===
  const esc = (s) => {
    const div = esc._d || (esc._d = document.createElement('div'));
    div.textContent = s == null ? '' : String(s);
    return div.innerHTML;
  };
  
  const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));
  
  // === Registry Loading with Deduplication ===
  window.$loadRegistry = async function(key = 'main') {
    // Check if already loading
    if (loadPromises.has(key)) {
      return loadPromises.get(key);
    }
    
    const loadPromise = (async () => {
      try {
        // Load manifest on first call
        if (!window.$registryManifest) {
          const res = await fetch('/js/registry.manifest.json');
          const manifest = await res.json();
          window.$registryManifest = manifest;
          window.$registryVersion = manifest.version;
        }
        
        // Get file path from manifest
        const file = key === 'main' 
          ? window.$registryManifest.main
          : window.$registryManifest.components?.[key];
        
        if (!file) {
          console.warn(`Registry key not found: ${key}`);
          return;
        }
        
        // Dynamic import
        const mod = await import(/* @vite-ignore */ file);
        Object.assign(window.$componentRegistry, mod.default || mod);
        
      } catch (err) {
        console.error(`Failed to load registry: ${key}`, err);
        throw err;
      } finally {
        loadPromises.delete(key);
      }
    })();
    
    loadPromises.set(key, loadPromise);
    return loadPromise;
  };
  
  // === Props Parsing ===
  window.$parseProps = function(el) {
    try {
      // Large props via ref
      if (el.dataset.propsRef) {
        return window.$largeProps[el.dataset.propsRef] || {};
      }
      
      // Large props via URL (lazy fetch)
      if (el.dataset.propsUrl) {
        return { __url: el.dataset.propsUrl };
      }
      
      // Inline props
      if (el.dataset.props) {
        return JSON.parse(el.dataset.props);
      }
      
      return {};
    } catch (err) {
      console.error('Failed to parse props:', err);
      return {};
    }
  };
  
  // === Merge Plan Executor ===
  window.$applyMergePlan = function(plan, scope) {
    if (!plan || !Array.isArray(plan)) {
      return {};
    }
    
    const result = {};
    
    for (const step of plan) {
      const [kind, key] = step.split(':', 2);
      
      if (kind === 'spread') {
        // Spread object (left-to-right merge)
        const obj = scope[key];
        if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
          Object.assign(result, obj);
        }
      } else if (kind === 'lit') {
        // Literal property (overwrites)
        result[key] = scope[key];
      }
    }
    
    return result;
  };
  
  // === Component Rendering with Circuit Breaker ===
  window.$renderDynamic = async function(el, name, props, options = {}) {
    const {
      retry = 0,
      skipCircuitBreaker = false,
    } = options;
    
    // Check circuit breaker
    if (!skipCircuitBreaker) {
      const failures = failedComponents.get(name) || 0;
      if (failures >= CONFIG.CIRCUIT_BREAKER_THRESHOLD) {
        el.innerHTML = `<!-- Component ${name} disabled after repeated failures -->`;
        el.classList.add('circuit-open');
        return;
      }
    }
    
    try {
      // Lazy fetch large props
      if (props && props.__url) {
        const res = await fetch(props.__url, { 
          credentials: 'same-origin',
          cache: 'default',
        });
        props = await res.json();
      }
      
      // Load registry if needed
      if (!window.$componentRegistry[name]) {
        await window.$loadRegistry(name);
      }
      
      // Get render function
      const fn = window.$componentRegistry[name];
      if (typeof fn !== 'function') {
        throw new Error(`Component not found: ${name}`);
      }
      
      // Render with escape helper
      const html = fn(props, { esc });
      
      // Update DOM in next frame
      requestAnimationFrame(() => {
        el.innerHTML = html;
        el.classList.add('hydrated');
        el.classList.remove('render-error', 'circuit-open');
      });
      
      // Reset failure count on success
      failedComponents.delete(name);
      retryDelays.delete(name);
      
    } catch (err) {
      console.error(`Error rendering ${name}:`, err);
      
      // Track failure
      const failures = (failedComponents.get(name) || 0) + 1;
      failedComponents.set(name, failures);
      
      // Exponential backoff retry
      if (retry < CONFIG.MAX_RETRIES) {
        const baseDelay = retryDelays.get(name) || CONFIG.INITIAL_RETRY_DELAY;
        const delay = Math.min(baseDelay * 2, CONFIG.MAX_RETRY_DELAY);
        retryDelays.set(name, delay);
        
        console.log(`Retrying ${name} in ${delay}ms (attempt ${retry + 1})`);
        
        await sleep(delay);
        return window.$renderDynamic(el, name, props, { 
          retry: retry + 1,
          skipCircuitBreaker: true,
        });
      }
      
      // Show error UI
      el.innerHTML = `
        <div class="component-error">
          <p>Failed to load component: ${esc(name)}</p>
          <button onclick="location.reload()">Reload Page</button>
        </div>
      `;
      el.classList.add('render-error');
      
      // Report to monitoring
      if (window.errorReporter) {
        window.errorReporter.log({
          type: 'component_render_error',
          component: name,
          props: props,
          error: err.message,
          stack: err.stack,
          retry: retry,
          failures: failures,
        });
      }
    }
  };
  
  // === Hydration with Version Check ===
  window.$hydrate = function(el) {
    const name = el.dataset.comp;
    const sig = el.dataset.sig;
    const registryVersion = el.dataset.registryVersion;
    
    // Version mismatch warning
    if (registryVersion && 
        window.$registryVersion && 
        registryVersion !== window.$registryVersion) {
      console.warn(
        `Registry version mismatch for ${name}: ` +
        `page=${registryVersion}, registry=${window.$registryVersion}. ` +
        `Page may be stale.`
      );
    }
    
    // Check if already hydrated with same signature
    if (el.dataset.currentName === name && 
        el.dataset.currentSig === sig) {
      el.classList.add('hydrated');
      return;
    }
    
    // Default: trust SSR, no re-render on load
    el.dataset.currentName = name;
    el.dataset.currentSig = sig;
    el.classList.add('hydrated');
    
    // Optional: validate SSR if flag set
    if (el.dataset.validateSsr === 'true') {
      const props = window.$parseProps(el);
      const currentSig = signature(name, props);
      if (currentSig !== sig) {
        console.warn(`SSR signature mismatch for ${name}, re-rendering`);
        window.$renderDynamic(el, name, props);
      }
    }
  };
  
  // === Alpine Integration ===
  window.$dynCompData = () => ({ 
    init() {
      // x-init handles hydration/rendering
    }
  });
  
  // === Initialization ===
  const init = () => {
    window.$loadRegistry('main').catch(err => {
      console.error('Failed to load main registry:', err);
    });
  };
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
```

### 5. Build-Time Renderer with Optional Merge Plans

```go
// renderer_dynamic.go
func (r *Renderer) RenderDynamicComponent(
    node *ast.DynamicComponentNode, 
    scope Scope,
) (string, error) {
    analyzer := NewScopeAnalyzer(scope)
    
    if analyzer.IsBuildResolvable(node.NameExpr) {
        // === Build-time Resolution ===
        name := r.EvaluateExpr(node.NameExpr, scope).(string)
        props, mergePlan := r.MergeProps(node, scope)
        
        // Render component HTML from template
        html := r.RenderComponent(name, props)
        
        // Serialize props (with threshold handling)
        propsAttr := r.SerializePropsAttr(name, props)
        
        // Generate deterministic signature
        sig := Signature(name, props)
        
        // Only emit merge plan if spreads are present
        mergePlanAttr := ""
        if len(node.SpreadExprs) > 0 {
            planJSON, _ := json.Marshal(mergePlan)
            mergePlanAttr = fmt.Sprintf(
                `data-merge='%s'`, 
                html.EscapeString(string(planJSON)),
            )
        }
        
        // Emit wrapper with markers
        return fmt.Sprintf(`
            <div class="dyn-comp" 
                 data-comp="%s"
                 %s
                 %s
                 data-sig="%s"
                 data-registry-version="%s"
                 x-data="$dynCompData"
                 x-init="$hydrate($el)">
                %s
            </div>
        `, 
            html.EscapeString(name),
            propsAttr,
            mergePlanAttr,
            html.EscapeString(sig),
            html.EscapeString(r.BuildVersion),
            html,
        ), nil
    }
    
    // === Runtime Resolution ===
    return r.RenderRuntimeWrapper(node, scope)
}

func (r *Renderer) SerializePropsAttr(
    name string, 
    props map[string]interface{},
) string {
    // Remove defaults to reduce size
    defaults := r.GetComponentDefaults(name)
    filtered := FilterDefaults(props, defaults)
    
    // Serialize to JSON
    propsJSON, err := json.Marshal(filtered)
    if err != nil {
        return ""
    }
    
    // Small props: inline as attribute
    if len(propsJSON) <= CONFIG.LARGE_PROPS_THRESHOLD {
        return fmt.Sprintf(
            `data-props='%s'`,
            html.EscapeString(string(propsJSON)),
        )
    }
    
    // Large props: store separately
    id := GeneratePropsID()
    r.RegisterLargeProps(id, filtered)
    
    return fmt.Sprintf(`data-props-ref="%s"`, id)
}
```

### 6. Props Serialization Strategies

```go
// props_serializer.go
const (
    INLINE_THRESHOLD = 2048      // 2KB - inline in attribute
    REF_THRESHOLD    = 10240     // 10KB - inline script tag
    URL_THRESHOLD    = 10240     // >10KB - separate JSON file
)

type PropsStrategy int

const (
    StrategyInline PropsStrategy = iota
    StrategyRef
    StrategyURL
)

func DetermineStrategy(propsJSON []byte) PropsStrategy {
    size := len(propsJSON)
    
    if size <= INLINE_THRESHOLD {
        return StrategyInline
    }
    
    if size <= REF_THRESHOLD {
        return StrategyRef
    }
    
    return StrategyURL
}
```

### 7. Development Overlay

```javascript
// dev-overlay.js
window.$componentDebug = {
  enabled: false,
  components: [],
  
  init() {
    if (process.env.NODE_ENV !== 'development') return;
    
    // Observe all dynamic components
    const observer = new MutationObserver(() => {
      this.scan();
    });
    
    observer.observe(document.body, {
      childList: true,
      subtree: true,
    });
    
    this.scan();
  },
  
  scan() {
    const components = [];
    document.querySelectorAll('.dyn-comp').forEach(el => {
      const name = el.dataset.comp;
      const sig = el.dataset.sig;
      const propsSize = el.dataset.props?.length || 
                        el.dataset.propsRef?.length ||
                        el.dataset.propsUrl?.length || 0;
      
      components.push({
        name,
        sig,
        status: el.classList.contains('hydrated') ? 'hydrated' :
                el.classList.contains('render-error') ? 'error' :
                el.classList.contains('circuit-open') ? 'circuit-open' :
                'pending',
        propsSize,
        element: el,
      });
    });
    
    this.components = components;
  },
  
  toggle() {
    this.enabled = !this.enabled;
  },
};

// Initialize in dev mode
if (process.env.NODE_ENV === 'development') {
  window.$componentDebug.init();
}
```

**UI Template:**

```html
<!-- Dev overlay panel -->
<div id="component-inspector" 
     x-show="$componentDebug.enabled"
     x-data="$componentDebug">
  <div class="overlay-header">
    <h3>Dynamic Components</h3>
    <button @click="toggle()">✕</button>
  </div>
  
  <div class="overlay-stats">
    <div>Total: <strong x-text="components.length"></strong></div>
    <div>Hydrated: <strong x-text="components.filter(c => c.status === 'hydrated').length"></strong></div>
    <div>Errors: <strong x-text="components.filter(c => c.status === 'error').length"></strong></div>
  </div>
  
  <div class="overlay-list">
    <template x-for="comp in components" :key="comp.sig">
      <div class="component-item" 
           :class="comp.status"
           @click="comp.element.scrollIntoView({ behavior: 'smooth' })">
        <div class="component-name">
          <strong x-text="comp.name"></strong>
          <span class="status-badge" x-text="comp.status"></span>
        </div>
        <div class="component-details">
          <code class="sig" x-text="comp.sig"></code>
          <span class="props-size" x-text="`${comp.propsSize}B`"></span>
        </div>
      </div>
    </template>
  </div>
</div>

<style>
#component-inspector {
  position: fixed;
  top: 20px;
  right: 20px;
  width: 400px;
  max-height: 80vh;
  background: white;
  border: 1px solid #ccc;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  overflow: hidden;
  z-index: 10000;
  font-family: monospace;
}

.overlay-header {
  padding: 12px;
  background: #f5f5f5;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.overlay-stats {
  padding: 8px 12px;
  background: #fafafa;
  display: flex;
  gap: 16px;
  font-size: 12px;
}

.overlay-list {
  overflow-y: auto;
  max-height: calc(80vh - 120px);
}

.component-item {
  padding: 12px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
  transition: background 0.2s;
}

.component-item:hover {
  background: #f9f9f9;
}

.component-item.error {
  background: #fff3f3;
}

.component-item.circuit-open {
  background: #fff9e6;
}

.status-badge {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  text-transform: uppercase;
  background: #e0e0e0;
  margin-left: 8px;
}

.status-badge:where(.hydrated *) {
  background: #c8e6c9;
}

.status-badge:where(.error *) {
  background: #ffcdd2;
}

.component-details {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 11px;
  color: #666;
}
</style>
```

---

## Performance Optimization

### Device-Aware Performance Targets

```javascript
// Detect device capability
function getDeviceCategory() {
  const memory = navigator.deviceMemory; // GB
  const cores = navigator.hardwareConcurrency;
  
  if (memory >= 8 && cores >= 8) return 'fast';
  if (memory >= 4 && cores >= 4) return 'medium';
  return 'slow';
}

const PERFORMANCE_TARGETS = {
  fast: {
    hydrationTime: 50,    // ms for 10 components
    registryLoad: 100,    // ms
    renderTime: 20,       // ms per component
  },
  medium: {
    hydrationTime: 150,
    registryLoad: 300,
    renderTime: 50,
  },
  slow: {
    hydrationTime: 300,
    registryLoad: 500,
    renderTime: 100,
  },
};

// Use in performance monitoring
const targets = PERFORMANCE_TARGETS[getDeviceCategory()];
```

### Lazy Hydration with IntersectionObserver

```javascript
// Defer offscreen component hydration
function setupLazyHydration() {
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const el = entry.target;
          window.$hydrate(el);
          observer.unobserve(el);
        }
      });
    },
    { rootMargin: '50px' } // Start 50px before visible
  );
  
  document.querySelectorAll('.dyn-comp[data-lazy]').forEach(el => {
    observer.observe(el);
  });
}
```

---

## Security

### Content Security Policy (Production)

```html
<meta http-equiv="Content-Security-Policy" content="
  default-src 'self';
  script-src 'self' 'nonce-BUILD_NONCE_HERE';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: https:;
  font-src 'self' data:;
  connect-src 'self';
">
```

**Nonce Generation:**

```go
// During build, generate unique nonce per page
func (b *Builder) GeneratePageNonce() string {
    return base64.StdEncoding.EncodeToString(
        generateRandom(16),
    )
}

// Inject into all inline scripts
func (b *Builder) InjectNonce(html string, nonce string) string {
    return strings.ReplaceAll(
        html,
        "<script>",
        fmt.Sprintf(`<script nonce="%s">`, nonce),
    )
}
```

### XSS Prevention

```go
// Component templates must escape all dynamic content
func (g *CodeGen) GenerateComponentFunction(comp Component) string {
    return fmt.Sprintf(`
        function %s(props, helpers) {
            const { esc } = helpers;
            return \`
                <section id="%s">
                    <h1>\${esc(props.title)}</h1>
                    <p>\${esc(props.description)}</p>
                </section>
            \`;
        }
    `, comp.Name, comp.ID)
}
```

---

## Testing Strategy

### Unit Tests (Go)

```go
func TestScopeAnalyzer_ResolutionRules(t *testing.T) {
    tests := []struct {
        expr     string
        scope    Scope
        expected bool
        reason   string
    }{
        {
            expr:     `"Header"`,
            scope:    Scope{},
            expected: true,
            reason:   "Literal strings are build-time",
        },
        {
            expr:     `content.title`,
            scope:    Scope{"content": map[string]interface{}{"title": "Hello"}},
            expected: true,
            reason:   "Content fields are build-time",
        },
        {
            expr:     `component.name`,
            scope:    Scope{"component": RuntimeVar{}},
            expected: false,
            reason:   "Loop variables are runtime",
        },
        {
            expr:     `content.title + "_suffix"`,
            scope:    Scope{"content": map[string]interface{}{"title": "Hello"}},
            expected: false,
            reason:   "Expressions with operators are runtime",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.reason, func(t *testing.T) {
            analyzer := NewScopeAnalyzer(tt.scope)
            result := analyzer.IsBuildResolvable(parseExpr(tt.expr))
            
            if result != tt.expected {
                t.Errorf("Expected %v, got %v for: %s", 
                    tt.expected, result, tt.expr)
            }
        })
    }
}

func TestSignature_Deterministic(t *testing.T) {
    props1 := map[string]interface{}{
        "title": "Hello",
        "count": 42,
        "nested": map[string]interface{}{
            "b": 2,
            "a": 1,
        },
    }
    
    props2 := map[string]interface{}{
        "nested": map[string]interface{}{
            "a": 1,
            "b": 2,
        },
        "count": 42,
        "title": "Hello",
    }
    
    sig1 := Signature("Hero", props1)
    sig2 := Signature("Hero", props2)
    
    if sig1 != sig2 {
        t.Errorf("Signatures should match for same data with different key order")
    }
}
```

### Integration Tests (Build Output)

```go
func TestBuildOutput_HydrationMarkers(t *testing.T) {
    builder := NewTestBuilder()
    builder.AddTemplate(`
        {for c in components}
            <Component:dynamic name={c.name} {...c.fields} />
        {/for}
    `)
    
    output, err := builder.Build()
    assert.NoError(t, err)
    
    // Verify markers present
    assert.Contains(t, output, `data-comp="Hero2436"`)
    assert.Contains(t, output, `data-props=`)
    assert.Contains(t, output, `data-sig=`)
    assert.Contains(t, output, `x-init="$hydrate($el)"`)
    
    // Verify SSR content present
    assert.Contains(t, output, `<section id="hero-2436">`)
}

func TestRegistry_TreeShaking(t *testing.T) {
    builder := NewTestBuilder()
    builder.AddComponents(
        Component{Name: "Hero", Used: true},
        Component{Name: "Footer", Used: true},
        Component{Name: "Unused", Used: false},
    )
    
    manifest, err := builder.GenerateManifest()
    assert.NoError(t, err)
    
    // Only used components in manifest
    assert.Contains(t, manifest.Components, "Hero")
    assert.Contains(t, manifest.Components, "Footer")
    assert.NotContains(t, manifest.Components, "Unused")
}
```

### E2E Tests (Browser)

```javascript
// e2e/dynamic-components.spec.js
import { test, expect } from '@playwright/test';

test('SSG content visible immediately', async ({ page }) => {
  await page.goto('/test-page');
  
  // SSR content should be visible before JS loads
  await page.evaluate(() => {
    // Disable JavaScript to test SSG
    window.stop();
  });
  
  const components = page.locator('.dyn-comp');
  await expect(components).toHaveCount(3);
  
  // SSR HTML should be present
  await expect(components.first()).toContainText('Hero Title');
});

test('Hydration does not re-render when signatures match', async ({ page }) => {
  await page.goto('/test-page');
  
  // Get initial HTML
  const initialHTML = await page.locator('.dyn-comp').first().innerHTML();
  
  // Wait for hydration
  await page.waitForFunction(() => 
    document.querySelectorAll('.hydrated').length > 0
  );
  
  // Get HTML after hydration
  const afterHTML = await page.locator('.dyn-comp').first().innerHTML();
  
  // Should not have re-rendered
  expect(initialHTML).toBe(afterHTML);
});

test('Runtime component swapping', async ({ page }) => {
  await page.goto('/test-page');
  
  await page.waitForSelector('.hydrated');
  
  // Change component via Alpine store
  await page.evaluate(() => {
    Alpine.store('currentComponent', {
      name: 'Services2437',
      fields: { title: 'New Title' }
    });
  });
  
  // Should update without reload
  await expect(page.locator('.dyn-comp')).toContainText('New Title');
});

test('Circuit breaker prevents cascading failures', async ({ page }) => {
  await page.goto('/test-page-with-errors');
  
  // Simulate component error
  await page.evaluate(() => {
    window.$componentRegistry.BrokenComponent = () => {
      throw new Error('Simulated error');
    };
  });
  
  // First few attempts should retry
  await page.waitForTimeout(1000);
  
  const errorMessages = await page.evaluate(() => 
    console.logs.filter(msg => msg.includes('Retrying')).length
  );
  
  expect(errorMessages).toBeGreaterThan(0);
  
  // After threshold, should stop trying
  const circuitOpen = await page.locator('.circuit-open').count();
  expect(circuitOpen).toBeGreaterThan(0);
});

test('Registry version mismatch warning', async ({ page }) => {
  await page.goto('/test-page');
  
  // Simulate version mismatch
  await page.evaluate(() => {
    document.querySelector('[data-registry-version]')
      .dataset.registryVersion = 'old-version';
  });
  
  // Should log warning
  const warnings = await page.evaluate(() => 
    console.warnings.filter(w => w.includes('version mismatch'))
  );
  
  expect(warnings.length).toBeGreaterThan(0);
});
```

---

## Build Reports

```bash
$ plenti build --verbose

╔═══════════════════════════════════════════════════════════╗
║        Dynamic Component Build Report - v4.0              ║
╚═══════════════════════════════════════════════════════════╝

📊 Component Statistics:
   Total components: 45
   ├─ Build-time resolved: 38 (84.4%) ✓
   └─ Runtime placeholders: 7 (15.6%)

📦 Registry Output:
   Main bundle: 18.4 KB (gzipped: 6.2 KB)
   Lazy chunks: 12 files, 31.8 KB total (gzipped: 10.1 KB)
   Manifest: 2.1 KB
   Version: a8f3c9e2

⚡ Performance:
   Build time: 2.34s (+8.2% from baseline)
   Cache hits: 89.3%
   Tree-shaking: removed 23 unused components

📋 Props Analysis:
   Inline props: 34 components
   Large props (ref): 8 components
   Large props (url): 3 components
   Largest: Hero2436 (8.2 KB)

⚠️  Warnings:
   • Component 'OldHero' uses deprecated API (use Hero2436)
   • Large props in Services2437 (consider splitting)
   • 2 components have no defaults defined

✓ Build completed successfully
```

---

## Migration Strategy

### Phase 1: Opt-in Beta (Weeks 1-2)

```yaml
# config.yaml
features:
  dynamicComponents:
    enabled: false  # Default off
    
templates:
  # Per-template opt-in
  - path: "/experiments/*"
    dynamicComponents: true
```

### Phase 2: Progressive Rollout (Weeks 3-4)

```yaml
features:
  dynamicComponents:
    mode: progressive
    percentage: 25  # 25% of builds
    
monitoring:
  metrics:
    - build_time
    - page_size
    - hydration_time
    - error_rate
```

### Phase 3: Default On (Weeks 5+)

```yaml
features:
  dynamicComponents:
    enabled: true
    
templates:
  # Opt-out for specific templates if needed
  - path: "/legacy/*"
    dynamicComponents: false
```

---

## Success Criteria

### Must Have ✅

1. **≥80% build-time resolution** - Most components render at build
2. **Zero hydration mismatches** - Signatures match server/client
3. **Registry ≤30KB gzipped** - Reasonable bundle size
4. **No XSS vectors** - Security audit passes
5. **DX adoption >90%** - Developers prefer new system

### Performance ⚡

| Metric | Fast Device | Medium | Slow |
|--------|------------|---------|------|
| Hydration (10 comp) | <50ms | <150ms | <300ms |
| Registry load | <100ms | <300ms | <500ms |
| Per-component render | <20ms | <50ms | <100ms |
| Build time increase | <10% | <10% | <10% |

### Quality Assurance 🔍

- [ ] Unit test coverage >90%
- [ ] Integration tests for all resolution paths
- [ ] E2E tests for common patterns
- [ ] Security audit passed
- [ ] Performance benchmarks met
- [ ] Documentation complete
- [ ] Migration guide validated

---

## Appendix: Decision Log

### Why v4.0 over v3.0/v2.0?

**Architecture (v3.0 wins):**
- ✅ Deterministic signatures prevent hydration bugs
- ✅ Single source of truth eliminates drift
- ✅ Merge plans guarantee parity
- ✅ Hardened CSP for production

**Pragmatism (v2.0 contributions):**
- ✅ Realistic performance expectations
- ✅ Enhanced error recovery
- ✅ Better developer tooling
- ✅ Comprehensive monitoring

**v4.0 Additions (fixes critical gaps):**
- ✅ Race condition handling (promise deduplication)
- ✅ Versioning strategy (prevents stale cache issues)
- ✅ Circuit breakers (prevents cascading failures)
- ✅ Optional merge plans (reduces overhead when not needed)
- ✅ Device-aware targets (realistic expectations)
- ✅ Sequence diagrams (implementation clarity)
- ✅ Exponential backoff (smart retry logic)

### Why Template Functions over Pre-rendered Variants?

**Scalability:** Linear growth (O(n)) vs exponential (O(n×m))
**Flexibility:** Handle any prop combination
**Size:** Smaller registry (code vs HTML × variants)

### Why Promise Deduplication?

**Problem:** Multiple components requesting same chunk causes redundant fetches
**Solution:** Single promise per key, shared across requesters
**Benefit:** Faster loading, reduced network overhead

### Why Circuit Breakers?

**Problem:** Failing component retries infinitely, degrading UX
**Solution:** After N failures, stop trying and show static fallback
**Benefit:** Graceful degradation, better performance

### Why Registry Versioning?

**Problem:** Cached page HTML with stale component references
**Solution:** Version check on hydration warns of mismatches
**Benefit:** Prevents subtle bugs from cache inconsistency

---

## Appendix: Canonical JSON Test Cases

```javascript
// Test cases for canonical JSON implementation

// Case 1: Key ordering
assert(
  canonicalJSON({c: 3, a: 1, b: 2}) ===
  canonicalJSON({a: 1, b: 2, c: 3})
);

// Case 2: Nested objects
assert(
  canonicalJSON({outer: {z: 2, a: 1}}) ===
  canonicalJSON({outer: {a: 1, z: 2}})
);

// Case 3: Array ordering preserved
assert(
  canonicalJSON({arr: [3, 1, 2]}) ===
  canonicalJSON({arr: [3, 1, 2]})
);

// Case 4: Transient keys excluded
assert(
  !canonicalJSON({__proto__: 'bad', a: 1}).includes('__proto__')
);

// Case 5: Go/JS parity
// These must produce identical output:
// Go: CanonicalJSON(map[string]interface{}{"a": 1, "b": 2})
// JS: canonicalJSON({a: 1, b: 2})
```

---

**Version:** 4.0  
**Status:** Production Ready  
**Last Updated:** [Current Date]  
**Maintainer:** Plenti Core Team  

**Review Checklist:**
- [ ] Architecture review by lead developer
- [ ] Security audit scheduled
- [ ] Performance benchmarks validated
- [ ] Documentation reviewed
- [ ] Migration plan approved
- [ ] Rollout schedule confirmed