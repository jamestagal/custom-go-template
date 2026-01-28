# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-12-dynamic-component-iteration/spec.md
**MANDATORY: Use go-backend agent for all Go implementation**
> Created: 2025-10-12
> Status: Ready for Implementation
> Agent: go-backend (MANDATORY)

## Tasks

- [x] 1. AST Nodes & Data Structures
  - [x] 1.1 Write tests for DynamicComponentByNameNode
  - [x] 1.2 Create DynamicComponentByNameNode in ast/ast.go
  - [x] 1.3 Enhance ComponentProp with IsSpread and SpreadExpr fields
  - [x] 1.4 Add NodeType() method implementation
  - [x] 1.5 Verify all AST tests pass
  - [x] 1.6 **MANDATORY: Use go-backend agent for all Go implementation**

- [x] 2. Parser Implementation
  - [x] 2.1 Write tests for ParseDynamicComponentByName (basic, spread, mixed props)
  - [x] 2.2 Create parser/dynamic_component_by_name.go
  - [x] 2.3 Implement ParseDynamicComponentByName function
  - [x] 2.4 Implement parseSpreadProp helper function
  - [x] 2.5 Add DynamicComponentByNameParser to AnyNodeParser parsers list
  - [x] 2.6 Handle Component:dynamic tag detection with colon
  - [x] 2.7 Parse name={} attribute (required)
  - [x] 2.8 Parse spread props {...expr}
  - [x] 2.9 Parse mixed regular and spread props
  - [x] 2.10 Verify all parser tests pass
  - [x] 2.11 **MANDATORY: Use go-backend agent for all Go implementation**

- [x] 3. Transformer Implementation
  - [x] 3.1 Write tests for TransformDynamicComponentByName
  - [x] 3.2 Write tests for resolveSpreadProps
  - [x] 3.3 Write tests for prop merging/override logic
  - [x] 3.4 Create transformer/dynamic_component_by_name.go
  - [x] 3.5 Implement TransformDynamicComponentByName function
  - [x] 3.6 Implement resolveSpreadProps function
  - [x] 3.7 Implement expression evaluation for component name
  - [x] 3.8 Implement prop merging with right-to-left override
  - [x] 3.9 Add to main transformer switch statement
  - [x] 3.10 Verify all transformer tests pass
  - [x] 3.11 **MANDATORY: Use go-backend agent for all Go implementation**

- [x] 4. Renderer Implementation
  - [x] 4.1 Write tests for RenderDynamicComponentByName fallback
  - [x] 4.2 Create renderer/dynamic_component_by_name.go
  - [x] 4.3 Implement RenderDynamicComponentByName (fallback placeholder)
  - [x] 4.4 Add to renderer switch statement
  - [x] 4.5 Verify all renderer tests pass
  - [x] 4.6 **MANDATORY: Use go-backend agent for all Go implementation**

- [x] 5. Server Integration & Magic Variables
  - [x] 5.1 Write integration tests for getAllContent helper
  - [x] 5.2 Write integration tests for magic variables passing
  - [x] 5.3 Implement getAllContent() function in cmd/server/main.go
  - [x] 5.4 Update renderTemplate to pass magic variables (components, content, allContent, allLayouts)
  - [x] 5.5 Create layouts/content/pages.html with Component:dynamic iteration
  - [x] 5.6 Update route handlers to use template-based iteration
  - [x] 5.7 Remove renderPlentiPage function (superseded by template iteration)
  - [x] 5.8 Verify all integration tests pass
  - [x] 5.9 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 6. End-to-End Testing
  - [ ] 6.1 Write E2E test for pages.html with multiple components
  - [ ] 6.2 Write E2E test for spread operator prop merging
  - [ ] 6.3 Write E2E test for magic variables (allContent, content)
  - [ ] 6.4 Test with real content JSON (content/pages/_index.json)
  - [ ] 6.5 Test component not found edge case
  - [ ] 6.6 Test invalid spread expression edge case
  - [ ] 6.7 Verify all E2E tests pass
  - [ ] 6.8 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 7. Documentation & Examples
  - [ ] 7.1 Update CLAUDE.md with Component:dynamic syntax
  - [ ] 7.2 Create example pages.html layout
  - [ ] 7.3 Create example blog.html layout (demonstrating content type mapping)
  - [ ] 7.4 Add developer guide section in docs/
  - [ ] 7.5 Update migration notes from Svelte/Plenti patterns
  - [ ] 7.6 **MANDATORY: Use go-backend agent for all Go implementation**

## Key Implementation Notes

**MANDATORY**: All Go code implementation MUST use the go-backend agent.

### Phase Order

1. **AST Foundation (Task 1)** - Create node structures first
2. **Parser (Task 2)** - Parse syntax into AST nodes
3. **Transformer (Task 3)** - Resolve dynamic components and spread props
4. **Renderer (Task 4)** - Fallback rendering for untransformed nodes
5. **Server Integration (Task 5)** - Wire magic variables and content loading
6. **Testing & Validation (Task 6)** - E2E testing with real content
7. **Documentation (Task 7)** - Developer guides and examples

### Critical Dependencies

- **Task 2 depends on Task 1**: AST nodes must exist before parsing
- **Task 3 depends on Task 2**: Transformer needs parsed nodes
- **Task 4 depends on Task 3**: Renderer handles transformed output
- **Task 5 depends on Tasks 1-4**: Full pipeline must be complete
- **Task 6 depends on Task 5**: Integration must work before E2E testing
- **Task 7 can run in parallel after Task 5**: Documentation can be written once integration is complete

### Test-First Approach

Each major task follows this pattern:
1. **Write tests first** (subtask X.1) - Define expected behavior
2. **Implement functionality** (subtasks X.2 to X.n-1) - Build to pass tests
3. **Verify tests pass** (subtask X.n) - Confirm implementation is correct

### Key Technical Decisions

**Component Name Resolution**:
- Evaluate `name={expression}` against current data scope
- Support both string literals (`name="Header"`) and expressions (`name={component.name}`)
- Case-sensitive matching ("Header" ≠ "header")
- Error if component not found in registry

**Spread Operator Semantics**:
- Syntax: `{...expression}` where expression evaluates to object
- Multiple spreads allowed: `{...defaults} {...overrides}`
- Merge order: left-to-right (later spreads override earlier)
- Regular props override spread props when declared after

**Prop Merging Order**:
```
1. First spread props (left to right)
2. Regular props (left to right)
3. Later spread props (left to right)
Result: Rightmost declaration wins
```

Example:
```html
<Component:dynamic {...defaults} title="Override" {...extras} />
```
Resolution: `defaults` → `title="Override"` → `extras` (extras.title wins if present)

**Magic Variables**:
- `components` - Array of component definitions from content JSON
- `content` - Current page/post content object
- `allContent` - All site content indexed by path
- `allLayouts` - Registry of all component templates

### Error Handling Strategy

**Component Not Found**:
```html
<!-- Warning: Component 'NonExistent' not found in registry -->
```

**Invalid Spread Expression**:
```
Error: "Spread expression must evaluate to object, got array"
```

**Missing Required Attribute**:
```
Error: "Component:dynamic requires name attribute"
```

**Name Expression Wrong Type**:
```
Error: "Component name must be a string, got number: 42"
```

**Circular References**:
```
Error: "Circular component reference detected: ComponentA → ComponentB → ComponentA"
```
(Detect with depth limit during transformation)

### Integration Points

**Parser Integration** (`parser/parser.go`):
```go
parsers := []parserFunc{
    // ... existing parsers
    tryParse(DynamicComponentByNameParser, "DynamicComponentByName"),
    // ... other parsers
}
```

**Transformer Integration** (`transformer/transformer.go`):
```go
switch node := n.(type) {
case *ast.DynamicComponentByNameNode:
    return t.TransformDynamicComponentByName(node)
// ... other cases
}
```

**Renderer Integration** (`renderer/render.go`):
```go
switch node := n.(type) {
case *ast.DynamicComponentByNameNode:
    return r.RenderDynamicComponentByName(node)
// ... other cases
}
```

### Server-Side Changes

**Before (server-level iteration)**:
```go
func renderPlentiPage(w http.ResponseWriter, r *http.Request) {
    // Iterate components server-side in Go
}
```

**After (template-level iteration)**:
```html
<!-- layouts/content/pages.html -->
{for component in components}
    <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

Server only needs to:
1. Load content JSON
2. Extract components array
3. Pass magic variables to template
4. Let template handle iteration

### Testing Coverage Requirements

**Unit Tests**:
- Parser: All syntax variations (basic, spread, mixed, multiple spreads)
- Transformer: Component name resolution, prop merging, spread evaluation
- Renderer: Fallback cases for untransformed nodes

**Integration Tests**:
- Full pipeline: Template → AST → Transformation → Rendering
- Magic variables passed correctly through pipeline
- Multiple components in loop iteration
- Nested dynamic components

**End-to-End Tests**:
- Real pages.html layout with Component:dynamic
- Real content JSON from content/pages/
- Server rendering complete pages
- Alpine.js reactivity in browser

### Performance Considerations

**Component Registry Lookup**:
- O(1) map lookup by component name
- Pre-register all components on server startup
- No filesystem access during rendering

**Spread Prop Evaluation**:
- Evaluate spread expressions once per component instance
- Cache component templates (already implemented)
- No re-parsing of component definitions

**Memory Management**:
- Reuse component AST across instances
- Don't clone entire AST for each dynamic component
- Stream rendering where possible

## Example Implementation Patterns

### Pattern 1: Basic Dynamic Component
```html
<Component:dynamic name={component.name} />
```
Transforms to: Look up component "Hero2436" → Render with empty props

### Pattern 2: Spread Props
```html
<Component:dynamic name="Header" {...component.fields} />
```
Transforms to: Render Header with fields = {title: "Hello", link: "/about"}

### Pattern 3: Mixed Props
```html
<Component:dynamic name={comp.name} {...comp.fields} theme="dark" />
```
Transforms to: Merge comp.fields + override theme="dark"

### Pattern 4: In Loop Context
```html
{for component in components}
    <Component:dynamic name={component.name} {...component.fields} allContent={allContent} />
{/for}
```
Transforms to: Iterate components, render each with merged props

### Pattern 5: Magic Variables
```html
<Component:dynamic name="BlogPost" {...fields} content={content} allContent={allContent} />
```
Transforms to: Pass magic variables to component for global state access

## Success Metrics

1. All parser tests pass (Task 2.10)
2. All transformer tests pass (Task 3.10)
3. All renderer tests pass (Task 4.5)
4. All integration tests pass (Task 5.8)
5. All E2E tests pass (Task 6.7)
6. Real pages.html renders correctly with real content
7. No performance regression vs static components
8. Documentation complete and accurate
9. Migration from renderPlentiPage to template iteration complete
10. Matches Plenti's Svelte component iteration behavior

## Related Files & References

**Core Implementation Files**:
- `ast/ast.go` - DynamicComponentByNameNode definition
- `parser/dynamic_component_by_name.go` - Parser implementation
- `transformer/dynamic_component_by_name.go` - Transformer implementation
- `renderer/dynamic_component_by_name.go` - Renderer implementation
- `cmd/server/main.go` - Server integration and magic variables

**Related Specs**:
- `.agent-os/specs/2025-10-11-export-let-content-injection/` - Content loading system
- `.agent-os/specs/2025-10-07-global-store-system/` - Store system integration
- `.agent-os/specs/2025-10-04-dynamic-component-rendering/` - Dynamic component by path

**Example Templates**:
- `layouts/content/pages.html` - Page layout with Component:dynamic
- `layouts/components/Hero2436.html` - Example component with export let
- `content/pages/_index.json` - Example content JSON

**Test Files**:
- `parser/dynamic_component_by_name_test.go` - Parser tests
- `transformer/dynamic_component_by_name_test.go` - Transformer tests
- `tests/dynamic_component_integration_test.go` - Integration tests

## Notes & Warnings

**CRITICAL**: This feature replaces server-level component iteration with template-level iteration. The renderPlentiPage function will be REMOVED after this implementation.

**Breaking Change**: Projects using direct renderPlentiPage calls will need to migrate to template-based iteration using Component:dynamic.

**Scope Limitation**: v1 does NOT support:
- Children/slots in Component:dynamic
- Spread operator in loops or conditionals
- Client-side lazy loading
- Dynamic component hydration

These features may be added in future versions if needed.

**Agent Assignment**: ALL Go code implementation MUST use the go-backend agent. Do NOT use general-purpose agent for implementation tasks.
