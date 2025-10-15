# Focused Solution: Runtime Component Resolution for Loop Variables

## The Problem in Detail

### Current Behavior (Failing)

```go
// Input template
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}

// Go transformer output
<template x-for="(component, ) in components">
  <div>
    <!-- ERROR: component.name can't be resolved at build time -->
  </div>
</template>
```

### Why It Fails

```go
// transformer/dynamic_component_by_name.go (line ~40)
func TransformDynamicComponentByName(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
    // This tries to resolve at BUILD TIME
    componentName, err := evaluateNameExpression(node.NameExpression, dataScope)
    
    // But "component" is an ALPINE LOOP VARIABLE
    // It doesn't exist in Go's dataScope!
    // Result: ERROR
}
```

---

## The Solution: Three-Part Fix

### Part 1: Detect Loop Variables (Scope Analyzer)

Create a scope analyzer that tracks which variables are build-time vs runtime:

```go
// analyzer/scope.go (NEW FILE)
package analyzer

type ScopeAnalyzer struct {
    buildVars   map[string]bool  // Known at build: content, allContent, etc.
    runtimeVars map[string]bool  // Loop variables: component, item, etc.
}

func NewScopeAnalyzer() *ScopeAnalyzer {
    return &ScopeAnalyzer{
        buildVars:   make(map[string]bool),
        runtimeVars: make(map[string]bool),
    }
}

func (sa *ScopeAnalyzer) MarkRuntimeVar(varName string) {
    sa.runtimeVars[varName] = true
}

func (sa *ScopeAnalyzer) IsRuntimeVar(varName string) bool {
    return sa.runtimeVars[varName]
}

func (sa *ScopeAnalyzer) IsRuntimeExpression(expr string) bool {
    // Check if expression uses any runtime variables
    
    // Example: "component.name"
    parts := strings.Split(expr, ".")
    if len(parts) == 0 {
        return false
    }
    
    rootVar := parts[0]
    return sa.runtimeVars[rootVar]
}
```

### Part 2: Mark Loop Variables During Transformation

Update your loop transformer to mark loop variables as runtime:

```go
// transformer/loops.go (MODIFY EXISTING)
func transformFor(node *ast.ForNode, dataScope map[string]any, scopeAnalyzer *ScopeAnalyzer) []ast.Node {
    // ... existing code ...
    
    // CRITICAL: Mark loop variable as runtime
    scopeAnalyzer.MarkRuntimeVar(loopVar)
    
    log.Printf("Marked '%s' as runtime variable (x-for loop)", loopVar)
    
    // ... rest of existing code ...
}
```

### Part 3: Emit Runtime Wrappers for Runtime Expressions

Modify dynamic component transformer to handle runtime expressions:

```go
// transformer/dynamic_component_by_name.go (MODIFY EXISTING)
func TransformDynamicComponentByName(
    node *ast.DynamicComponentByNameNode, 
    dataScope map[string]any,
    scopeAnalyzer *ScopeAnalyzer,  // ADD THIS PARAMETER
) []ast.Node {
    log.Printf("TransformDynamicComponentByName: nameExpr=%q", node.NameExpression)
    
    // NEW: Check if this is a runtime expression
    if scopeAnalyzer.IsRuntimeExpression(node.NameExpression) {
        log.Printf("Detected RUNTIME expression: %q (contains loop variable)", node.NameExpression)
        return emitRuntimeWrapper(node, dataScope)
    }
    
    // EXISTING: Build-time resolution (for static names)
    componentName, err := evaluateNameExpression(node.NameExpression, dataScope)
    if err != nil {
        log.Printf("Failed to evaluate: %v", err)
        return createDynamicByNamePlaceholder(node, "ERROR: "+err.Error())
    }
    
    // ... rest of existing code ...
}

// NEW FUNCTION
func emitRuntimeWrapper(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
    log.Printf("Emitting runtime wrapper for: %q", node.NameExpression)
    
    // Create a placeholder that Alpine.js will fill in at runtime
    wrapper := &ast.Element{
        TagName: "div",
        Attributes: []ast.Attribute{
            {
                Name:    "x-data",
                Value:   fmt.Sprintf("{__componentName: %s}", node.NameExpression),
                Dynamic: true,
            },
            {
                Name:    "x-init",
                Value:   "$nextTick(() => $renderDynamicComponent($el, __componentName, $data))",
                Dynamic: true,
            },
        },
        Children: []ast.Node{
            &ast.CommentNode{
                Content: fmt.Sprintf(" Runtime component: %s ", node.NameExpression),
            },
        },
    }
    
    return []ast.Node{wrapper}
}
```

---

## Client-Side Runtime (Alpine.js Extension)

Create a simple Alpine.js helper that renders components at runtime:

```javascript
// static/js/runtime-components.js (NEW FILE)

document.addEventListener('alpine:init', () => {
    // Store component templates
    Alpine.store('componentRegistry', {
        // Will be populated by Go at build time
        Hero2436: `
            <section class="hero">
                <h1 x-text="title"></h1>
                <p x-text="subtitle"></p>
            </section>
        `,
        Services2437: `
            <section class="services">
                <h2 x-text="title"></h2>
                <div x-html="description"></div>
            </section>
        `,
        // More components...
    });
    
    // Runtime component renderer
    Alpine.magic('renderDynamicComponent', () => {
        return (el, componentName, props) => {
            const registry = Alpine.store('componentRegistry');
            
            if (!registry[componentName]) {
                console.error(`Component not found: ${componentName}`);
                el.innerHTML = `<!-- Component ${componentName} not found -->`;
                return;
            }
            
            // Get template
            const template = registry[componentName];
            
            // Create Alpine scope with props
            const scope = Alpine.reactive(props);
            
            // Render template into element
            el.innerHTML = template;
            
            // Initialize Alpine on the new content
            Alpine.initTree(el);
        };
    });
});
```

---

## Build-Time Component Registry Generation

Generate the JavaScript registry at build time:

```go
// builder/registry_generator.go (NEW FILE)
package builder

func GenerateComponentRegistry(outputPath string) error {
    registry := make(map[string]string)
    
    // Walk component directory
    componentsDir := "layouts/components"
    files, err := os.ReadDir(componentsDir)
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if !strings.HasSuffix(file.Name(), ".html") {
            continue
        }
        
        // Read component template
        componentPath := filepath.Join(componentsDir, file.Name())
        content, err := os.ReadFile(componentPath)
        if err != nil {
            continue
        }
        
        // Parse component name
        componentName := strings.TrimSuffix(file.Name(), ".html")
        componentName = strings.Title(componentName) // hero → Hero
        
        // Convert to Alpine.js template
        alpineTemplate := convertToAlpineTemplate(string(content))
        
        registry[componentName] = alpineTemplate
    }
    
    // Generate JavaScript file
    return writeRegistryJS(registry, outputPath)
}

func writeRegistryJS(registry map[string]string, outputPath string) error {
    var buf strings.Builder
    
    buf.WriteString("// Auto-generated component registry\n")
    buf.WriteString("document.addEventListener('alpine:init', () => {\n")
    buf.WriteString("  Alpine.store('componentRegistry', {\n")
    
    for name, template := range registry {
        // Escape template for JavaScript string
        escaped := strings.ReplaceAll(template, "`", "\\`")
        escaped = strings.ReplaceAll(escaped, "${", "\\${")
        
        buf.WriteString(fmt.Sprintf("    %s: `%s`,\n", name, escaped))
    }
    
    buf.WriteString("  });\n")
    buf.WriteString("});\n")
    
    return os.WriteFile(outputPath, []byte(buf.String()), 0644)
}
```

---

## Implementation Phases

### Phase 1: Scope Tracking (4-6 hours)

1. Create `analyzer/scope.go`
2. Add `MarkRuntimeVar()` to loop transformer
3. Pass `scopeAnalyzer` through transformation chain
4. Test: Verify loop variables are marked as runtime

### Phase 2: Runtime Wrapper Emission (6-8 hours)

1. Add `IsRuntimeExpression()` check to dynamic component transformer
2. Implement `emitRuntimeWrapper()` function
3. Test: Verify runtime wrappers are emitted for loop variables
4. Test: Verify build-time resolution still works for static names

### Phase 3: Client-Side Runtime (6-8 hours)

1. Create `static/js/runtime-components.js`
2. Implement `$renderDynamicComponent` magic function
3. Test: Verify components render at runtime in browser
4. Test: Verify Alpine.js reactivity works in rendered components

### Phase 4: Registry Generation (6-8 hours)

1. Create `builder/registry_generator.go`
2. Implement component template scanning
3. Convert templates to Alpine.js format
4. Generate JavaScript registry file
5. Test: Verify registry loads and components render

**Total Effort: 22-30 hours** (NOT 115 hours!)

---

## Testing Strategy

### Test 1: Static Component Name (Should Work Now)

```html
<!-- Build time resolution -->
<Component:dynamic name="Hero2436" title="Welcome" />

<!-- Expected output: Rendered component HTML -->
<div x-data='{"title":"Welcome"}' class="hero">
  <section id="hero-2436">...</section>
</div>
```

### Test 2: Loop Variable (Currently Fails, Will Work After Fix)

```html
<!-- Runtime resolution -->
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}

<!-- Expected output: Runtime wrapper -->
<template x-for="component in components">
  <div x-data="{__componentName: component.name}" 
       x-init="$nextTick(() => $renderDynamicComponent($el, __componentName, component.fields))">
    <!-- Runtime component: component.name -->
  </div>
</template>
```

### Test 3: Mixed (Should Work After Fix)

```html
<!-- Build time + runtime mix -->
<Component:dynamic name="Header" />

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}

<Component:dynamic name="Footer" />
```

---

## Key Differences from Full v4.0

| Feature | Full v4.0 (115h) | Focused Solution (25h) |
|---------|------------------|------------------------|
| **Scope analyzer** | ✅ Complex | ✅ Simple (just loop vars) |
| **Runtime wrappers** | ✅ Full | ✅ Minimal |
| **Component registry** | ✅ Advanced | ✅ Basic |
| **Signatures** | ✅ Deterministic | ❌ Skip |
| **Circuit breakers** | ✅ Full | ❌ Skip |
| **Dev overlay** | ✅ Full | ❌ Skip |
| **Tree-shaking** | ✅ Advanced | ❌ Skip |
| **Hydration checks** | ✅ Full | ❌ Skip |

You only need the **focused solution** for your Plenti use case!

---

## Next Steps

1. **Implement scope analyzer** (track loop variables as runtime)
2. **Detect runtime expressions** in dynamic component transformer
3. **Emit runtime wrappers** instead of trying to resolve at build time
4. **Create simple Alpine.js runtime** ($renderDynamicComponent)
5. **Generate component registry** at build time

This will fix your immediate problem without requiring the full v4.0 implementation!

---

## Why This Is Simpler Than v4.0

**v4.0 assumes:**
- Users might change component types at runtime (filters, searches)
- Need signature validation for hydration mismatches
- Need advanced error handling (circuit breakers)
- Need development tools (overlay, debugging)

**Your actual needs:**
- Components are determined at BUILD TIME (from content JSON)
- Just need runtime rendering of those predetermined components
- No user-driven component switching
- Basic error handling is sufficient

**Result:** ~25% of the v4.0 effort gets you 100% of what you need!
