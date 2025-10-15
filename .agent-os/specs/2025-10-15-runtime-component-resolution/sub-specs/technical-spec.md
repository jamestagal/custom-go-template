# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-15-runtime-component-resolution/spec.md

> Created: 2025-10-15
> Version: 1.0.0

## Technical Requirements

### 1. Scope Analyzer (`analyzer/scope.go`)

**Purpose**: Distinguish build-time resolvable expressions from runtime-only expressions

**Implementation**:
- `ScopeAnalyzer` struct with `buildVars` and `runtimeVars` maps
- `IsRuntimeExpression(expr ast.Expr) bool` - returns true if expression contains any runtime-only variable
- `TrackLoopVariable(name string)` - marks loop iterator as runtime-only
- `TrackContentProp(name string)` - marks content prop as build-resolvable

**Decision Logic**:
```
IsRuntimeExpression returns TRUE if:
  - Expression contains a loop variable (component, item, index)
  - Expression contains Alpine store reference ($store.*)
  - Expression contains operators (for safety, treat as runtime)

IsRuntimeExpression returns FALSE if:
  - Expression is a string literal ("ComponentName")
  - Expression only references content props (content.title)
  - Expression only references exported props (title, description)
```

**Integration Point**: Called by transformer before attempting component resolution

---

### 2. Runtime Wrapper Emission (`transformer/dynamic_component_by_name.go`)

**Purpose**: Emit Alpine.js-compatible wrapper for components with runtime-only names

**Modification to existing transformer**:
```go
func TransformDynamicComponentByName(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
    analyzer := NewScopeAnalyzer(dataScope)

    if analyzer.IsRuntimeExpression(node.NameExpression) {
        // NEW: Runtime resolution path
        return emitRuntimeWrapper(node, dataScope)
    }

    // EXISTING: Build-time resolution path (unchanged)
    componentName := evaluateNameExpression(node.NameExpression, dataScope)
    // ... existing logic
}
```

**Runtime Wrapper Output**:
```html
<div
  class="dyn-comp-runtime"
  x-data="{compName: component.name, compProps: component.fields}"
  x-init="$renderDynamicComponent($el, compName, compProps)">
  <!-- Alpine.js will fill this at runtime -->
</div>
```

**Props Handling**: Serialize props to JSON and pass via x-data

---

### 3. Client-Side Runtime (`static/js/runtime-components.js`)

**Purpose**: Alpine.js magic for runtime component rendering

**Core Function**:
```javascript
Alpine.magic('renderDynamicComponent', () => {
  return async (el, componentName, props) => {
    // 1. Load registry if not loaded
    if (!window.$componentRegistry) {
      await loadComponentRegistry();
    }

    // 2. Get component template function
    const templateFn = window.$componentRegistry[componentName];
    if (!templateFn) {
      el.innerHTML = `<!-- Warning: Component '${componentName}' not found -->`;
      return;
    }

    // 3. Render component
    const html = templateFn(props);
    el.innerHTML = html;
  };
});
```

**Registry Loading**:
```javascript
async function loadComponentRegistry() {
  const response = await fetch('/js/component-registry.js');
  const module = await import(response.url);
  window.$componentRegistry = module.default;
}
```

**Error Handling**:
- Missing component -> HTML comment warning (silent)
- Network error -> Retry with exponential backoff (3 attempts)
- Template error -> Log to console, show error message in dev mode

---

### 4. Registry Code Generation (`builder/registry_generator.go`)

**Purpose**: Convert Go component templates to JavaScript functions

**Process**:
```go
func GenerateComponentRegistry(components []ComponentTemplate) string {
    var jsCode strings.Builder

    jsCode.WriteString("export default {\n")

    for _, comp := range components {
        // Convert component AST to JavaScript template literal
        jsTemplate := convertToJSTemplate(comp.AST)

        jsCode.WriteString(fmt.Sprintf("  '%s': (props) => `%s`,\n",
            comp.Name, jsTemplate))
    }

    jsCode.WriteString("};\n")
    return jsCode.String()
}
```

**Template Conversion**:
- `{variable}` -> `${props.variable}`
- `<span x-text="title">` -> `<span x-text="title">${props.title}</span>`
- HTML elements -> preserved as-is
- Alpine directives -> preserved for client-side hydration

**Output Example**:
```javascript
// static/js/component-registry.js
export default {
  'Hero2436': (props) => `
    <section id="hero-2436">
      <h1>${props.title}</h1>
      <p>${props.description}</p>
    </section>
  `,
  'Services2437': (props) => `
    <section id="services-2437">
      <h2>Services</h2>
    </section>
  `
};
```

---

### 5. Server-Side Loop Expansion (`renderer/loop_expansion.go`)

**Purpose**: Pre-render component loops for static builds

**Implementation**:
```go
func ExpandComponentLoop(loopNode *ast.LoopNode, dataScope map[string]any) (string, error) {
    // Get the array to iterate
    array := dataScope[loopNode.Iterator].([]interface{})

    var html strings.Builder

    for index, item := range array {
        // Create loop scope
        loopScope := copyScope(dataScope)
        loopScope[loopNode.Variable] = item
        loopScope["index"] = index

        // Render loop body with loop scope
        for _, childNode := range loopNode.Children {
            childHTML := RenderNode(childNode, loopScope)
            html.WriteString(childHTML)
        }
    }

    return html.String(), nil
}
```

**When to Expand**:
- During static build (`plenti build`)
- When component data is available at build time
- For SEO and fast first paint

**When to Use Runtime**:
- During dev server (`go run cmd/server/main.go`)
- When component data changes frequently
- For interactive features (filters, sorting)

---

### 6. Integration Points

**Transformer Integration**:
```go
// transformer/transform.go
func Transform(template *ast.Template, dataScope map[string]any) *ast.Template {
    // Initialize scope analyzer
    analyzer := NewScopeAnalyzer()

    // Track loop variables during traversal
    for _, node := range template.Nodes {
        if loopNode, ok := node.(*ast.LoopNode); ok {
            analyzer.TrackLoopVariable(loopNode.Variable)
        }
    }

    // Transform nodes with scope awareness
    // ...
}
```

**Renderer Integration**:
```go
// renderer/render.go
func Render(template *ast.Template, dataScope map[string]any) string {
    // Detect build mode
    if isBuildMode() {
        // Server-side loop expansion
        return RenderWithExpansion(template, dataScope)
    }

    // Dev mode - emit runtime wrappers
    return RenderWithRuntimeWrappers(template, dataScope)
}
```

---

## Performance Criteria

1. **Registry Size**: Single component registry file < 100KB for typical site (10-20 components)
2. **Build Time**: Registry generation adds < 500ms to build time
3. **Runtime Performance**: Component rendering < 50ms per component
4. **Memory**: Scope analyzer overhead < 1MB per page

---

## External Dependencies

**None Required** - This spec uses only existing dependencies:
- Alpine.js 3.x (already in use)
- Go standard library (encoding/json, strings, etc.)
- Existing AST and parser infrastructure

No new external libraries needed.
