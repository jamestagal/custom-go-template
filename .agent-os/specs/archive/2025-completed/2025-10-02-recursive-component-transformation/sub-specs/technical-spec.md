# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-02-recursive-component-transformation/spec.md

## Technical Requirements

### 1. Component Registry Enhancements

**Location**: `transformer/components.go`

- Ensure `componentTemplateRegistry` is properly accessible during transformation
- `ComponentTemplate` struct must include parsed `*ast.Template`, component name, and props list
- `RegisterComponent()` and `GetComponentTemplate()` functions working correctly
- Consider thread-safety if transformations can be concurrent (future consideration)

### 2. Core transformComponent() Refactoring

**Location**: `transformer/components.go` - `transformComponent()` function

**Current Signature**: `func transformComponent(node *ast.ComponentNode, parentDataScope map[string]any) []ast.Node`

**Implementation Steps**:

1. **Component Lookup**:
   ```go
   componentTemplate, exists := GetComponentTemplate(node.Name)
   if !exists {
       log.Printf("Error: Component template '%s' not registered.", node.Name)
       return []ast.Node{&ast.TextNode{Content: fmt.Sprintf("<!-- Component %s not found -->", node.Name)}}
   }
   ```

2. **Isolated Scope Creation**:
   ```go
   componentDataScope := make(map[string]any)
   ```

3. **Process Component's Own Fence**:
   ```go
   if fence := FindFenceSection(componentTemplate.Template.RootNodes); fence != nil {
       collectComponentFenceData(fence, componentDataScope)
   }
   ```

4. **Resolve Passed Props**:
   ```go
   for _, passedProp := range node.Props {
       resolvedValue := resolvePropValue(passedProp, parentDataScope)
       componentDataScope[passedProp.Name] = resolvedValue
   }
   ```

5. **Transform Component Body**:
   ```go
   componentBodyNodes := filterOutFence(componentTemplate.Template.RootNodes)
   transformedChildren := transformNodes(componentBodyNodes, componentDataScope, false)
   ```

6. **Add x-data Wrapper**:
   ```go
   finalComponentNodes := addComponentDataWrapper(transformedChildren, componentDataScope)
   return finalComponentNodes
   ```

### 3. Helper Function: collectComponentFenceData()

**Location**: `transformer/components.go` or `transformer/scope.go`

**Signature**: `func collectComponentFenceData(fence *ast.FenceSection, scope map[string]any)`

**Purpose**: Extract variables, prop defaults, and functions from component's fence section

**Implementation**:
- Process `fence.Variables` - parse and add to scope using `parseValue()`
- Process `fence.Props` - add default values to scope (will be overwritten by passed props)
- Extract function definitions from `fence.RawContent` using regex patterns
- Store function definitions as strings in the scope (alpineDataFormatter will handle them)

**Regex Patterns Needed**:
- `function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{[^}]*}` - Function declarations
- `(const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*function\([^)]*\)\s*{[^}]*}` - Function expressions
- `(const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*\([^)]*\)\s*=>\s*{[^}]*}` - Arrow functions
- `([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{[^}]*}` - Method shorthand

### 4. Helper Function: resolvePropValue()

**Location**: `transformer/components.go`

**Signature**: `func resolvePropValue(prop ast.ComponentProp, parentScope map[string]any) any`

**Purpose**: Resolve prop values by evaluating expressions against parent scope

**Logic**:
- If `prop.IsDynamic == true`: prop.Value is an expression - look up in parentScope
- If `prop.IsShorthand == true`: look up `prop.Name` in parentScope
- If static: use `parseValue(prop.Value)` to parse literal
- Return resolved value or expression string if not found in parent

### 5. Helper Function: addComponentDataWrapper()

**Location**: `transformer/components.go` or `transformer/alpine.go`

**Signature**: `func addComponentDataWrapper(nodes []ast.Node, dataScope map[string]any) []ast.Node`

**Purpose**: Wrap component nodes with x-data attribute

**Logic**:
- Format dataScope using `alpineDataFormatter(dataScope)`
- If single root element: add x-data attribute directly to it
- If multiple root nodes: create wrapper `<div x-data="...">` and add nodes as children
- Handle edge case of empty dataScope (still add wrapper for consistency)
- Return single element or wrapper as `[]ast.Node`

### 6. Helper Function: filterOutFence()

**Location**: `transformer/components.go` or `transformer/utils.go`

**Signature**: `func filterOutFence(nodes []ast.Node) []ast.Node`

**Purpose**: Remove FenceSection nodes from a slice

**Implementation**:
```go
var filtered []ast.Node
for _, n := range nodes {
    if _, isFence := n.(*ast.FenceSection); !isFence {
        filtered = append(filtered, n)
    }
}
return filtered
```

### 7. Enhanced parseValue() Helper

**Location**: `transformer/components.go`, `transformer/utils.go`, or extract from `cmd/server/main.go`

**Signature**: `func parseValue(value string) interface{}`

**Purpose**: Parse JavaScript literal values from fence section strings

**Must Handle**:
- Booleans: `"true"`, `"false"`
- Null: `"null"`
- Numbers: integers and floats
- Strings: quoted with `"` or `'`
- Arrays: `"[1, 2, 3]"` - return as string for Alpine
- Objects: `"{ key: 'value' }"` - return as string for Alpine
- Expressions: variable references, function calls - return as string

### 8. Integration with transformNodes()

**Location**: `transformer/transformer.go`

**Current Code** (line ~140):
```go
case *ast.ComponentNode:
    log.Printf("transformNodes: Transforming Component node %s", n.Name)
    componentNodes := transformComponent(n, dataScope)
    transformedNodes = append(transformedNodes, componentNodes...)
```

**Verification**: This is already correct - just ensure it's calling the refactored `transformComponent()`

### 9. Test Cleanup

**Location**: `transformer/components.go`

**Remove**:
- Lines 172-272: All `isAlpineIntegrationTest()` special-case code
- Any hardcoded component rendering logic for specific test cases
- `resetRenderedComponents()` calls if they're only for test workarounds

**Location**: `transformer/alpine.go`

**Remove**:
- Any test-specific hardcoded x-data formatting (lines 24-44)
- Special cases for `dynamic_components_with_conditionals` test

### 10. Function Definition Extraction

**Technical Detail**: When extracting functions from fence.RawContent, store them as raw strings:

```go
scope[fnName] = fnBody // e.g., "function formatPrice(price) { return '$' + price.toFixed(2); }"
```

The `alpineDataFormatter()` function (in the separate Function Expression Handling spec) will handle outputting these without quotes.

### 11. Dynamic Components

**Handle**: `<{componentVar} />` syntax where component name is a variable

**Implementation**:
```go
if strings.HasPrefix(componentName, "{") && strings.HasSuffix(componentName, "}") {
    varName := strings.Trim(componentName, "{} ")
    extractVariablesFromExpr(varName, componentDataScope)
    // Use x-component directive with variable binding
}
```

### 12. Error Handling

**Required**:
- Log warning if component not found in registry
- Return comment node as placeholder: `<!-- Component X not found -->`
- Log warning if prop not found in parent scope (for dynamic props)
- Handle nil/empty component templates gracefully

### 13. Logging Strategy

**Add Debug Logs**:
- Component transformation start/end
- Fence data collection results
- Prop resolution (name, value, type)
- Final data scope contents
- Number of transformed children

**Example**:
```go
log.Printf("Transforming component: %s", node.Name)
log.Printf("  Component fence data: %v", componentDataScope)
log.Printf("  Resolved prop '%s' to: %v (type: %T)", propName, resolvedValue, resolvedValue)
log.Printf("  Final component scope: %v", componentDataScope)
```

## External Dependencies

No new external dependencies are required. This spec uses only the existing Go standard library and internal packages.
