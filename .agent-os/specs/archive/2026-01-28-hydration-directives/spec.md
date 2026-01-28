# Hydration Directives Specification

## Overview

This spec defines the implementation of **hydration directives** for the Go template engine, inspired by Astro's `client:*` directives. These directives control **when** a component's JavaScript initializes, enabling lazy hydration strategies that significantly improve page load performance.

## Problem Statement

Currently, all components with `x-data` hydrate immediately on page load:

```html
<!-- All 10 components hydrate at once, blocking main thread -->
<div x-data="{ ... }">Hero</div>
<div x-data="{ ... }">Features</div>
<div x-data="{ ... }">Testimonials</div>  <!-- Below fold -->
<div x-data="{ ... }">Pricing</div>       <!-- Below fold -->
<div x-data="{ ... }">FAQ</div>           <!-- Below fold -->
<div x-data="{ ... }">Contact</div>       <!-- Below fold -->
```

**Impact:**
- Time to Interactive (TTI) blocked by off-screen components
- Main thread saturated during initial load
- Memory allocated for components user may never see
- Poor Core Web Vitals (LCP, INP, TBT)

## Solution: Hydration Directives

### Supported Directives

| Directive | Trigger | Use Case |
|-----------|---------|----------|
| `client:load` | Immediate (default) | Critical above-fold components |
| `client:visible` | IntersectionObserver | Below-fold components |
| `client:idle` | requestIdleCallback | Non-critical components |
| `client:media` | matchMedia | Responsive components |
| `client:none` | Never | Static-only components |

---

## Template Syntax

### Component Syntax

```html
<!-- Immediate hydration (default, current behavior) -->
<Component:dynamic name="Hero" {...fields} />
<Hero title="Welcome" />

<!-- Visible hydration (lazy) -->
<Component:dynamic name="FAQ" {...fields} client:visible />
<FAQ items={faqItems} client:visible />

<!-- Idle hydration (low priority) -->
<Analytics trackingId="UA-123" client:idle />

<!-- Media query hydration -->
<MobileMenu client:media="(max-width: 768px)" />

<!-- No hydration (static only) -->
<StaticFooter client:none />
```

### Attribute Syntax

```html
<!-- Boolean form -->
client:visible
client:idle
client:none

<!-- Value form (for media) -->
client:media="(max-width: 768px)"

<!-- With rootMargin for visible -->
client:visible="200px"
```

---

## Architecture

### Phase 1: Parser Changes

#### 1.1 AST Node Changes

**File:** `ast/ast.go`

```go
// HydrationDirective represents a client:* hydration strategy
type HydrationDirective struct {
    Type  string // "load", "visible", "idle", "media", "none"
    Value string // Optional value (media query, rootMargin)
}

// ComponentNode - add HydrationDirective field
type ComponentNode struct {
    Name       string
    Props      []ComponentProp
    Children   []Node
    Hydration  *HydrationDirective // NEW: nil = immediate (default)
}

// DynamicComponentByNameNode - add HydrationDirective field
type DynamicComponentByNameNode struct {
    NameExpression string
    Props          []ComponentProp
    SpreadProps    []string
    Children       []Node
    Hydration      *HydrationDirective // NEW
}
```

#### 1.2 Parser Changes

**File:** `parser/components.go`

```go
// parseHydrationDirective extracts client:* directive from attributes
func parseHydrationDirective(attrs []Attribute) (*ast.HydrationDirective, []Attribute) {
    var directive *ast.HydrationDirective
    var remaining []Attribute

    for _, attr := range attrs {
        if strings.HasPrefix(attr.Name, "client:") {
            directiveType := strings.TrimPrefix(attr.Name, "client:")
            directive = &ast.HydrationDirective{
                Type:  directiveType,
                Value: attr.Value,
            }
        } else {
            remaining = append(remaining, attr)
        }
    }

    return directive, remaining
}
```

---

### Phase 2: Transformer Changes

#### 2.1 Build-Time Components

**File:** `transformer/components.go`

For components resolved at build-time, wrap output with hydration metadata:

```go
func transformComponentWithHydration(node *ast.ComponentNode, dataScope map[string]any) []ast.Node {
    // Transform component as usual
    rendered := transformComponent(node, dataScope)

    // If no hydration directive or client:load, return as-is
    if node.Hydration == nil || node.Hydration.Type == "load" {
        return rendered
    }

    // Wrap with hydration container
    return wrapWithHydrationContainer(rendered, node.Hydration, dataScope)
}

func wrapWithHydrationContainer(content []ast.Node, hydration *ast.HydrationDirective, dataScope map[string]any) []ast.Node {
    // Extract x-data from content (move to data attribute)
    xDataValue := extractXDataFromContent(content)
    contentWithoutXData := removeXDataFromContent(content)

    wrapper := &ast.Element{
        TagName: "div",
        Attributes: []ast.Attribute{
            {Name: "data-hydrate", Value: hydration.Type},
            {Name: "data-hydrate-value", Value: hydration.Value}, // For media queries
            {Name: "data-x-data", Value: xDataValue},
        },
        Children: contentWithoutXData,
    }

    return []ast.Node{wrapper}
}
```

#### 2.2 Runtime Components

**File:** `transformer/dynamic_component_by_name.go`

For runtime-resolved components, add hydration directive to wrapper:

```go
func emitRuntimeWrapper(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
    // ... existing code ...

    // Add hydration directive if present
    hydrateAttr := "load" // default
    hydrateValue := ""
    if node.Hydration != nil {
        hydrateAttr = node.Hydration.Type
        hydrateValue = node.Hydration.Value
    }

    wrapper := &ast.Element{
        TagName: "div",
        Attributes: []ast.Attribute{
            {Name: "class", Value: "dyn-comp-runtime"},
            {Name: "data-hydrate", Value: hydrateAttr},
            {Name: "data-hydrate-value", Value: hydrateValue},
            {Name: "x-data", Value: xDataValue, IsAlpine: true},
            // x-init is now conditional based on hydration
        },
        Children: []ast.Node{},
    }

    // Only add x-init for immediate hydration
    if hydrateAttr == "load" {
        wrapper.Attributes = append(wrapper.Attributes, ast.Attribute{
            Name:    "x-init",
            Value:   "$renderDynamicComponent($el, compName, compProps)",
            IsAlpine: true,
        })
    }

    return []ast.Node{wrapper}
}
```

---

### Phase 3: Runtime Implementation

#### 3.1 Hydration Manager

**File:** `static/js/hydration-manager.js`

```javascript
/**
 * Hydration Manager - Controls when components hydrate
 *
 * Strategies:
 * - load: Immediate (default Alpine behavior)
 * - visible: IntersectionObserver
 * - idle: requestIdleCallback
 * - media: matchMedia
 * - none: Never hydrate (static only)
 */

class HydrationManager {
    constructor() {
        this.observers = new Map();
        this.mediaListeners = new Map();
    }

    /**
     * Initialize hydration for all deferred components
     */
    init() {
        // Find all elements with data-hydrate (excluding "load")
        document.querySelectorAll('[data-hydrate]').forEach(el => {
            const strategy = el.dataset.hydrate;

            if (strategy === 'load') {
                this.hydrateImmediate(el);
            } else if (strategy === 'visible') {
                this.hydrateOnVisible(el);
            } else if (strategy === 'idle') {
                this.hydrateOnIdle(el);
            } else if (strategy === 'media') {
                this.hydrateOnMedia(el);
            }
            // 'none' = never hydrate
        });
    }

    /**
     * Immediate hydration
     */
    hydrateImmediate(el) {
        this.hydrate(el);
    }

    /**
     * Hydrate when element becomes visible
     */
    hydrateOnVisible(el) {
        const rootMargin = el.dataset.hydrateValue || '200px';

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.hydrate(el);
                    observer.disconnect();
                    this.observers.delete(el);
                }
            });
        }, {
            rootMargin,
            threshold: 0.01
        });

        observer.observe(el);
        this.observers.set(el, observer);
    }

    /**
     * Hydrate when browser is idle
     */
    hydrateOnIdle(el) {
        const callback = () => this.hydrate(el);

        if ('requestIdleCallback' in window) {
            requestIdleCallback(callback, { timeout: 2000 });
        } else {
            // Fallback for Safari
            setTimeout(callback, 200);
        }
    }

    /**
     * Hydrate when media query matches
     */
    hydrateOnMedia(el) {
        const query = el.dataset.hydrateValue;
        if (!query) {
            console.warn('[Hydration] client:media requires a media query value');
            this.hydrate(el);
            return;
        }

        const mql = window.matchMedia(query);

        const handler = (e) => {
            if (e.matches) {
                this.hydrate(el);
                mql.removeEventListener('change', handler);
                this.mediaListeners.delete(el);
            }
        };

        // Check immediately
        if (mql.matches) {
            this.hydrate(el);
        } else {
            mql.addEventListener('change', handler);
            this.mediaListeners.set(el, { mql, handler });
        }
    }

    /**
     * Perform hydration on an element
     */
    hydrate(el) {
        // Check if already hydrated
        if (el.dataset.hydrated === 'true') {
            return;
        }

        console.log('[Hydration] Hydrating:', el.dataset.hydrate);

        // Restore x-data from data attribute
        const xData = el.dataset.xData;
        if (xData) {
            el.setAttribute('x-data', xData);
            el.removeAttribute('data-x-data');
        }

        // Check if this is a runtime component
        if (el.classList.contains('dyn-comp-runtime')) {
            // Runtime component - need to call $renderDynamicComponent
            const compName = el.dataset.compName ||
                            this.extractFromXData(xData, 'compName');
            const compProps = el.dataset.compProps ||
                             this.extractFromXData(xData, 'compProps');

            if (window.$renderDynamicComponent) {
                window.$renderDynamicComponent(el, compName, compProps);
            }
        }

        // Initialize Alpine on this element
        if (window.Alpine) {
            window.Alpine.initTree(el);
        }

        // Mark as hydrated
        el.dataset.hydrated = 'true';
        el.removeAttribute('data-hydrate');
    }

    /**
     * Extract value from x-data JSON string
     */
    extractFromXData(xDataStr, key) {
        try {
            const data = JSON.parse(xDataStr.replace(/'/g, '"'));
            return data[key];
        } catch (e) {
            return null;
        }
    }

    /**
     * Cleanup observers
     */
    destroy() {
        this.observers.forEach(observer => observer.disconnect());
        this.observers.clear();

        this.mediaListeners.forEach(({ mql, handler }) => {
            mql.removeEventListener('change', handler);
        });
        this.mediaListeners.clear();
    }
}

// Initialize on DOMContentLoaded
const hydrationManager = new HydrationManager();

document.addEventListener('DOMContentLoaded', () => {
    hydrationManager.init();
});

// Export for manual control
window.HydrationManager = hydrationManager;
```

#### 3.2 Integration with Alpine.js

**File:** `static/js/runtime-components.js` (update)

```javascript
// Modify the Alpine.js initialization to work with hydration manager

document.addEventListener('alpine:init', () => {
    // Register magic for runtime component rendering
    window.Alpine.magic('renderDynamicComponent', () => {
        return renderDynamicComponent;
    });

    // Make it globally available for hydration manager
    window.$renderDynamicComponent = renderDynamicComponent;
});

// Don't auto-init elements with data-hydrate (let HydrationManager handle them)
// This requires Alpine.js configuration or custom initialization
```

---

### Phase 4: HTML Output Examples

#### Before (Current - All Immediate)

```html
<div x-data="{ count: 0 }">
    <button @click="count++">Count: <span x-text="count"></span></button>
</div>
```

#### After - Immediate (client:load or default)

```html
<div x-data="{ count: 0 }">
    <button @click="count++">Count: <span x-text="count"></span></button>
</div>
```

#### After - Visible (client:visible)

```html
<div data-hydrate="visible"
     data-hydrate-value="200px"
     data-x-data="{ count: 0 }">
    <!-- Static HTML rendered at build time -->
    <button>Count: <span>0</span></button>
</div>
```

#### After - Idle (client:idle)

```html
<div data-hydrate="idle"
     data-x-data="{ items: [...] }">
    <!-- Static HTML -->
    <ul>
        <li>Item 1</li>
        <li>Item 2</li>
    </ul>
</div>
```

#### After - Media (client:media)

```html
<div data-hydrate="media"
     data-hydrate-value="(max-width: 768px)"
     data-x-data="{ open: false }">
    <!-- Mobile menu - only hydrates on mobile -->
    <nav>...</nav>
</div>
```

#### After - None (client:none)

```html
<div>
    <!-- No x-data at all - purely static -->
    <footer>Copyright 2025</footer>
</div>
```

---

## Implementation Plan

### Phase 1: Foundation (4-6 hours)
- [ ] Add `HydrationDirective` to AST nodes
- [ ] Update parser to extract `client:*` attributes
- [ ] Unit tests for parser changes

### Phase 2: Transformer (4-6 hours)
- [ ] Implement `wrapWithHydrationContainer()` for build-time components
- [ ] Update `emitRuntimeWrapper()` for runtime components
- [ ] Handle `client:none` (strip all x-data)
- [ ] Unit tests for transformer changes

### Phase 3: Runtime (4-6 hours)
- [ ] Create `hydration-manager.js`
- [ ] Implement `hydrateOnVisible()` with IntersectionObserver
- [ ] Implement `hydrateOnIdle()` with requestIdleCallback
- [ ] Implement `hydrateOnMedia()` with matchMedia
- [ ] Integration with Alpine.js initialization

### Phase 4: Integration (2-4 hours)
- [ ] Update server to serve `hydration-manager.js`
- [ ] Ensure proper script loading order
- [ ] End-to-end tests
- [ ] Performance benchmarking

### Phase 5: Documentation (2 hours)
- [ ] Update CLAUDE.md with hydration directive syntax
- [ ] Add examples to component documentation
- [ ] Performance guidelines

---

## Test Plan

### Unit Tests

```go
// parser/hydration_directive_test.go
func TestParseHydrationDirective_Visible(t *testing.T) {
    input := `<Hero title="Welcome" client:visible />`
    // Assert directive.Type == "visible"
}

func TestParseHydrationDirective_Media(t *testing.T) {
    input := `<Menu client:media="(max-width: 768px)" />`
    // Assert directive.Value == "(max-width: 768px)"
}

func TestParseHydrationDirective_None(t *testing.T) {
    input := `<Footer client:none />`
    // Assert no x-data in output
}
```

### Integration Tests

```go
// tests/hydration/hydration_integration_test.go
func TestHydration_VisibleOutputsCorrectHTML(t *testing.T) {
    // Verify data-hydrate="visible" in output
}

func TestHydration_IdleOutputsCorrectHTML(t *testing.T) {
    // Verify data-hydrate="idle" in output
}

func TestHydration_NoneStripsXData(t *testing.T) {
    // Verify no Alpine directives in output
}
```

### E2E Tests (Browser)

```javascript
// tests/e2e/hydration.spec.js
test('client:visible only hydrates when scrolled into view', async () => {
    // Load page with below-fold component
    // Assert component not hydrated
    // Scroll into view
    // Assert component now hydrated
});

test('client:idle hydrates after page is idle', async () => {
    // Load page
    // Assert component not immediately hydrated
    // Wait for idle
    // Assert component hydrated
});
```

---

## Performance Metrics

### Before (All Immediate)

| Metric | Value |
|--------|-------|
| Time to Interactive | ~800ms |
| Total Blocking Time | ~600ms |
| Components hydrated at load | 10/10 |
| Initial JS execution | 100% |

### After (With Directives)

| Metric | Value |
|--------|-------|
| Time to Interactive | ~350ms |
| Total Blocking Time | ~200ms |
| Components hydrated at load | 4/10 |
| Initial JS execution | 40% |

**Expected Improvement: 50-60% reduction in TTI**

---

## Migration Guide

### Existing Components (No Change Required)

Components without `client:*` directives continue to work exactly as before (immediate hydration).

### Opt-in Lazy Hydration

```html
<!-- Before -->
<FAQ items={faqItems} />

<!-- After (add directive) -->
<FAQ items={faqItems} client:visible />
```

### Best Practices

1. **Above-fold components:** No directive (or `client:load`)
2. **Below-fold components:** `client:visible`
3. **Analytics/Tracking:** `client:idle`
4. **Mobile-only components:** `client:media="(max-width: 768px)"`
5. **Static content:** `client:none`

---

## Open Questions

1. **Should `client:visible` be the default for `<Component:dynamic>` in loops?**
   - Pro: Automatic performance optimization
   - Con: Might surprise developers

2. **Root margin configuration?**
   - Default: `200px` (start hydrating 200px before visible)
   - Should this be configurable globally?

3. **Interaction with existing x-data?**
   - How to handle components that already have complex x-data?
   - Preserve reactivity during deferred hydration?

---

## References

- [Astro Client Directives](https://docs.astro.build/en/reference/directives-reference/#client-directives)
- [IntersectionObserver API](https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API)
- [requestIdleCallback](https://developer.mozilla.org/en-US/docs/Web/API/Window/requestIdleCallback)
- [Patterns for Partial Hydration](https://www.patterns.dev/posts/progressive-hydration)
