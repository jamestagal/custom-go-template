# Spec: Dynamic Component Iteration with Spread Operator

**Date**: 2025-10-12
**Status**: Draft
**Priority**: High
**Agent**: go-backend (MANDATORY)

## Overview

Implement `<Component:dynamic>` syntax to enable Plenti-style component iteration in templates. This allows templates to dynamically render components by name with spread props, matching Svelte's `<svelte:component this={name} {...fields} />` pattern.

## Business Context

In Plenti, content types map directly to layout templates:
```
content/pages/*.json      → layouts/content/pages.svelte
content/blog/*.json       → layouts/content/blog.svelte
content/products/*.json   → layouts/content/products.svelte
```

Each layout uses the SAME iteration pattern to render multiple components:
```svelte
{#each components as {name, fields}}
  <svelte:component this={allLayouts["layouts_components_" + name + "_svelte"]}
                    {...fields}
                    {allContent}
                    {content}/>
{/each}
```

This Go template engine must replicate this pattern WITHOUT Svelte primitives.

## Goals

1. Enable dynamic component rendering by variable name
2. Support spread operator for props (`{...component.fields}`)
3. Support regular props (Plenti magic variables: `allContent`, `allLayouts`, `content`)
4. Maintain template-first architecture (iteration happens in template, not server)
5. Work seamlessly with existing `{for}` loop syntax

## Syntax Design

### Example Template (layouts/content/pages.html)

```html
---
export let components, allContent, allLayouts, content
---

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Page</title>
    <link rel="stylesheet" href="/styles/style.css">
</head>
<body>
    {for component in components}
        <Component:dynamic name={component.name} {...component.fields} allContent={allContent} content={content} />
    {/for}
</body>
</html>
```

### Syntax Elements

#### 1. Component:dynamic Tag
```html
<Component:dynamic name={expression} />
```

- **Tag name**: `Component:dynamic` (case-sensitive)
- **Required attribute**: `name={expression}` - the component name to render
- **Special semantics**: Colon (`:`) indicates a special built-in component type

#### 2. Spread Operator
```html
<Component:dynamic {...component.fields} />
```

- **Syntax**: `{...expression}`
- **Behavior**: Spreads all properties from object as individual props
- **Example**: `{...component.fields}` where `fields = {title: "Hello", link: "/about"}`
  - Equivalent to: `title="Hello" link="/about"`

#### 3. Mixed Props
```html
<Component:dynamic name={component.name} {...component.fields} allContent={allContent} />
```

- **Supports**: Mix of spread props and regular props
- **Order matters**: Later props override earlier ones (right-to-left precedence)
- **Example**: `{...{title: "A"}} title="B"` results in `title="B"`

### Comparison with Existing Syntax

| Feature | Current Syntax | New Syntax |
|---------|---------------|------------|
| Static component | `<Header title="Hello" />` | Same |
| Dynamic by path | `<='./views/{comp}.html' />` | Still supported |
| Dynamic by name | ❌ Not supported | `<Component:dynamic name={varName} />` |
| Spread props | ❌ Not supported | `{...props}` |

## Example Content JSON

```json
{
  "components": [
    {
      "name": "Hero2436",
      "fields": {
        "topper": "Welcome",
        "title": "Main Title",
        "description": "Description text",
        "buttonText": "Click Here",
        "buttonLink": "/contact"
      }
    },
    {
      "name": "ContactForm",
      "fields": {
        "title": "Get In Touch",
        "email": "info@example.com"
      }
    }
  ]
}
```

## Example Component (layouts/components/Hero2436.html)

```html
---
export let topper, title, description, buttonText, buttonLink, allContent, content
---

<section id="hero-2436">
    <div class="container">
        <span class="topper">{topper}</span>
        <h2>{title}</h2>
        <p>{description}</p>
        <a href="{buttonLink}">{buttonText}</a>
    </div>
</section>

<style>
#hero-2436 { padding: 2rem; }
</style>
```

## Technical Design

### 1. AST Nodes

#### DynamicComponentByNameNode
```go
// DynamicComponentByNameNode represents <Component:dynamic name={expr} {...spread} prop={val} />
type DynamicComponentByNameNode struct {
    NameExpression string          // Expression that evaluates to component name
    Props          []ComponentProp // Regular props (name={value})
    SpreadProps    []string        // Spread expressions ({...expr})
    SelfClosing    bool
}

func (d *DynamicComponentByNameNode) NodeType() string {
    return "DynamicComponentByName"
}
```

#### ComponentProp Enhancement
```go
// ComponentProp represents a prop passed to a component
type ComponentProp struct {
    Name      string
    Value     string
    IsDynamic bool   // true if value is {expression}
    IsSpread  bool   // NEW: true if this is a spread prop {...expr}
    SpreadExpr string // NEW: expression for spread (e.g., "component.fields")
}
```

### 2. Parser

#### Grammar
```
DynamicComponent := '<Component:dynamic' Attributes '/>'
                  | '<Component:dynamic' Attributes '>' Children '</Component:dynamic>'

Attributes := (RegularProp | SpreadProp)*

RegularProp := Name '=' ('"' String '"' | '{' Expression '}')

SpreadProp := '{...' Expression '}'
```

#### Parser Implementation

**File**: `parser/dynamic_component_by_name.go`

```go
// ParseDynamicComponentByName parses <Component:dynamic name={expr} {...spread} />
func ParseDynamicComponentByName(input string) (*ast.DynamicComponentByNameNode, string, error) {
    // 1. Match opening tag: <Component:dynamic
    // 2. Parse attributes (name={}, {...spread}, regular props)
    // 3. Identify spread props by {...} pattern
    // 4. Handle self-closing /> or children </Component:dynamic>
    // 5. Return DynamicComponentByNameNode
}

// parseSpreadProp parses {...expression}
func parseSpreadProp(input string) (spreadExpr string, remaining string, error) {
    // Match pattern: {...someExpression}
    // Extract expression between {...}
    // Return expression and remaining input
}
```

**Integration**: Add to `AnyNodeParser` parsers list

```go
// In parser/parser.go
parsers := []parserFunc{
    // ... existing parsers
    tryParse(DynamicComponentByNameParser, "DynamicComponentByName"),
    // ... other parsers
}
```

### 3. Transformer

**File**: `transformer/dynamic_component_by_name.go`

```go
// TransformDynamicComponentByName transforms <Component:dynamic> nodes
func TransformDynamicComponentByName(node *ast.DynamicComponentByNameNode, dataScope map[string]any) []ast.Node {
    // 1. Evaluate name expression against dataScope
    //    Example: component.name → "Hero2436"

    // 2. Look up component template from registry
    //    componentTemplate := GetComponentTemplate(resolvedName)

    // 3. Build component props:
    //    a. Process spread props first
    //    b. Resolve regular props against dataScope
    //    c. Later props override earlier (right-to-left)

    // 4. Transform component with resolved props
    //    return transformComponent(componentTemplate, mergedProps)
}

// resolveSpreadProps evaluates spread expressions and returns flattened props
func resolveSpreadProps(spreadExprs []string, dataScope map[string]any) map[string]any {
    // For each spread expression:
    // 1. Evaluate expression against dataScope
    //    Example: "component.fields" → {title: "Hello", link: "/about"}
    // 2. Type assert to map[string]any
    // 3. Merge all spread results (later spreads override earlier)
    // 4. Return flattened props
}
```

**Prop Resolution Order**:
```
1. Spread props (left to right)
2. Regular props (left to right)
3. Result: Later props override earlier
```

**Example**:
```html
<Component:dynamic {...defaults} title="Override" {...extras} />
```
Results in: `defaults` → then `title="Override"` → then `extras` (extras wins conflicts)

### 4. Renderer

**File**: `renderer/dynamic_component_by_name.go`

```go
// RenderDynamicComponentByName renders <Component:dynamic> nodes
// Note: This should rarely be called directly since transformer converts
// DynamicComponentByNameNode into regular rendered component nodes
func RenderDynamicComponentByName(node *ast.DynamicComponentByNameNode) string {
    // Fallback rendering if node wasn't transformed
    // Return placeholder or error message
    return fmt.Sprintf("<!-- Dynamic component: name=%s (not resolved) -->",
                       node.NameExpression)
}
```

### 5. Server Integration

**File**: `cmd/server/main.go`

Update route handlers to pass magic variables:

```go
func renderTemplate(entrypoint string, w http.ResponseWriter, r *http.Request) {
    // ... existing code ...

    // Load content for route
    contentData, err := loadContentWithCache(r.URL.Path)

    // Build props with magic variables
    props := make(map[string]interface{})
    props["components"] = contentData["components"]  // Components array
    props["content"] = contentData                    // Full content object
    props["allContent"] = getAllContent()             // All site content
    props["allLayouts"] = componentTemplateRegistry   // All registered components

    // ... transform and render ...
}

// getAllContent loads all content files from content/ directory
func getAllContent() map[string]interface{} {
    // Walk content/ directory
    // Load all JSON files
    // Return map of all content indexed by path
}
```

## Implementation Phases

### Phase 1: AST & Parser (Task 1.1 - 1.3)
- [ ] Create `DynamicComponentByNameNode` in `ast/ast.go`
- [ ] Enhance `ComponentProp` with spread support
- [ ] Implement `ParseDynamicComponentByName` parser
- [ ] Implement `parseSpreadProp` helper
- [ ] Add to `AnyNodeParser` parsers list
- [ ] Write parser tests

**Tests**:
```go
// parser/dynamic_component_by_name_test.go
TestParseDynamicComponentByName_Basic
TestParseDynamicComponentByName_WithSpread
TestParseDynamicComponentByName_MixedProps
TestParseDynamicComponentByName_MultipleSpread
TestParseSpreadProp_Simple
TestParseSpreadProp_NestedProperty
```

### Phase 2: Transformer (Task 2.1 - 2.3)
- [ ] Implement `TransformDynamicComponentByName`
- [ ] Implement `resolveSpreadProps`
- [ ] Implement prop merging with override logic
- [ ] Add expression evaluation for component name
- [ ] Integrate with existing component transformer
- [ ] Write transformer tests

**Tests**:
```go
// transformer/dynamic_component_by_name_test.go
TestTransformDynamicComponentByName_ResolveComponentName
TestResolveSpreadProps_SingleSpread
TestResolveSpreadProps_MultipleSpread
TestPropMerging_SpreadThenRegular
TestPropMerging_RegularThenSpread
TestDynamicComponent_WithMagicVariables
```

### Phase 3: Renderer (Task 3.1)
- [ ] Implement `RenderDynamicComponentByName` (fallback)
- [ ] Add to main render switch statement
- [ ] Write renderer tests

### Phase 4: Server Integration (Task 4.1 - 4.3)
- [ ] Implement `getAllContent()` helper
- [ ] Update `renderTemplate` to pass magic variables
- [ ] Remove server-level component iteration (renderPlentiPage)
- [ ] Update route handlers to use template-based iteration
- [ ] Write integration tests

**Tests**:
```go
// tests/dynamic_component_integration_test.go
TestDynamicComponentIteration_PagesLayout
TestDynamicComponentIteration_WithMagicVariables
TestDynamicComponentIteration_MultipleComponents
TestSpreadOperator_OverrideProps
```

### Phase 5: Documentation & Examples (Task 5.1 - 5.2)
- [ ] Update CLAUDE.md with Component:dynamic syntax
- [ ] Create example pages.html layout
- [ ] Create example content JSON
- [ ] Add developer guide for dynamic components
- [ ] Update migration guide from Svelte/Plenti

## Edge Cases & Error Handling

### 1. Component Not Found
```html
<Component:dynamic name="NonExistent" />
```
**Behavior**: Render placeholder with warning
```html
<!-- Warning: Component 'NonExistent' not found in registry -->
```

### 2. Name Expression Evaluates to Non-String
```html
<Component:dynamic name={42} />
```
**Behavior**: Convert to string or error
**Error**: `"Component name must be a string, got number: 42"`

### 3. Spread Expression Not an Object
```html
<Component:dynamic {...someArray} />
```
**Behavior**: Skip spread or error
**Error**: `"Spread expression must evaluate to object, got array"`

### 4. Circular Component References
```html
<!-- In ComponentA.html -->
<Component:dynamic name="ComponentB" />

<!-- In ComponentB.html -->
<Component:dynamic name="ComponentA" />
```
**Behavior**: Detect cycle with depth limit
**Error**: `"Circular component reference detected: ComponentA → ComponentB → ComponentA"`

### 5. Missing Required name Attribute
```html
<Component:dynamic {...fields} />
```
**Behavior**: Parse error
**Error**: `"Component:dynamic requires name attribute"`

## Testing Strategy

### Unit Tests
- Parser: All syntax variations
- Transformer: Prop resolution and merging
- Renderer: Fallback cases

### Integration Tests
- Full template → component iteration
- Magic variables passed correctly
- Multiple components in loop
- Nested components with dynamic rendering

### End-to-End Tests
- Real pages.html layout
- Real content JSON
- Server rendering
- Browser rendering with Alpine.js

## Success Criteria

1. ✅ Can parse `<Component:dynamic name={expr} />` syntax
2. ✅ Can parse spread operator `{...expr}` in props
3. ✅ Correctly resolves component by name at runtime
4. ✅ Spread props merge correctly with regular props
5. ✅ Works inside `{for}` loops
6. ✅ Magic variables (`allContent`, `content`, `allLayouts`) pass through
7. ✅ Matches Plenti's Svelte component iteration behavior
8. ✅ Performance: No significant slowdown vs static components

## Non-Goals (Out of Scope)

- ❌ Dynamic component by path expression (already supported via `<=` syntax)
- ❌ Slots/children content (future enhancement)
- ❌ Component lazy loading
- ❌ Client-side component hydration

## Migration Path

### Current Server-Level Iteration
```go
// cmd/server/main.go
func renderPlentiPage(w http.ResponseWriter, r *http.Request) {
    // Iterate components server-side
}
```

### New Template-Level Iteration
```html
<!-- layouts/content/pages.html -->
{for component in components}
    <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Steps**:
1. Implement Component:dynamic feature
2. Create pages.html template
3. Update server to use renderTemplate instead of renderPlentiPage
4. Remove renderPlentiPage function
5. Test with existing content JSON

## Related Specifications

- `.agent-os/specs/2025-10-11-export-let-content-injection/` - Content loading system
- `.agent-os/specs/2025-10-07-global-store-system/` - Store system integration
- `.agent-os/specs/2025-10-04-dynamic-component-rendering/` - Dynamic component by path

## Questions & Decisions

### Q1: Should Component:dynamic support children?
```html
<Component:dynamic name={component.name}>
    <p>Child content</p>
</Component:dynamic>
```
**Decision**: Not in v1. Add in future if needed (requires slots system).

### Q2: Spread operator elsewhere (loops, conditionals)?
```html
{for {...item} in items}  <!-- Spread destructuring? -->
```
**Decision**: No. Only in component props for v1.

### Q3: Should name attribute accept string literals?
```html
<Component:dynamic name="Header" />  <!-- String literal instead of expression -->
```
**Decision**: Yes, support both. If static string, use directly. If expression, evaluate.

### Q4: Component name case sensitivity?
**Decision**: Case-sensitive. "Header" ≠ "header". Matches component registration.

## Appendix: Example Use Cases

### Use Case 1: Basic Page Layout
```html
<!-- layouts/content/pages.html -->
---
export let components
---
<body>
    {for component in components}
        <Component:dynamic name={component.name} {...component.fields} />
    {/for}
</body>
```

### Use Case 2: Blog Post Layout
```html
<!-- layouts/content/blog.html -->
---
export let components, content
---
<article>
    <h1>{content.title}</h1>
    <time>{content.date}</time>

    {for component in components}
        <Component:dynamic name={component.name} {...component.fields} content={content} />
    {/for}
</article>
```

### Use Case 3: Override Props
```html
<!-- Force a specific prop value -->
<Component:dynamic name={component.name} {...component.fields} theme="dark" />
```

### Use Case 4: Conditional Rendering
```html
{if component.name == "Hero"}
    <Component:dynamic name="Hero" {...component.fields} featured={true} />
{else}
    <Component:dynamic name={component.name} {...component.fields} />
{/if}
```

## Agent Assignment

**MANDATORY: All implementation tasks MUST use the go-backend agent.**

The go-backend agent is required for:
- All Go code implementation
- AST node creation
- Parser implementation
- Transformer implementation
- Renderer implementation
- Test writing
- Integration work

Do NOT use general-purpose agent for Go implementation tasks.
