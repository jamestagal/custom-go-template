# Spec 4: Dynamic Component Paths - Technical Specification

**Date**: 2025-10-03

## Architecture

### AST Changes

**New Node Type** (`ast/ast.go`):

```go
// DynamicComponentNode represents a component with a dynamic path
// Example: <='./views/{comp}.html' prop={value} />
type DynamicComponentNode struct {
    PathExpression string      // The path with possible {variables}
    Props          []Attribute // Props to pass to component
    SelfClosing    bool         // Whether it's self-closing
}
```

**Type Switch Updates**:
- Add case in `transformer/transformer.go`
- Add case in any visitor patterns
- Add String() method for debugging

### Parser Implementation

**Location**: `parser/components.go`

```go
// DynamicComponentParser parses the <= syntax for dynamic components
// Syntax: <='path' props />
// Example: <='./views/{comp}.html' age={age + 1} />
func DynamicComponentParser() Parser {
    return func(input string) Result {
        trimmed := strings.TrimSpace(input)

        // 1. Check for <= prefix
        if !strings.HasPrefix(trimmed, "<=") {
            return Result{Success: false}
        }

        // 2. Extract quoted path
        pathStart := 2 // After <=
        for pathStart < len(trimmed) && unicode.IsSpace(rune(trimmed[pathStart])) {
            pathStart++
        }

        if pathStart >= len(trimmed) {
            return Result{Success: false, Error: "Expected path after <="}
        }

        // 3. Determine quote type
        quoteChar := trimmed[pathStart]
        if quoteChar != '\'' && quoteChar != '"' {
            return Result{Success: false, Error: "Path must be quoted"}
        }

        // 4. Find closing quote
        pathEnd := pathStart + 1
        for pathEnd < len(trimmed) && trimmed[pathEnd] != quoteChar {
            pathEnd++
        }

        if pathEnd >= len(trimmed) {
            return Result{Success: false, Error: "Unclosed path quote"}
        }

        pathExpression := trimmed[pathStart+1:pathEnd]

        // 5. Parse props after path
        afterPath := trimmed[pathEnd+1:]
        propsResult := ComponentPropsParser()(afterPath)

        var props []ast.Attribute
        if propsResult.Success {
            props = propsResult.Node.([]ast.Attribute)
        }

        // 6. Check for self-closing
        remaining := trimmed[pathEnd+1:]
        if propsResult.Success {
            remaining = propsResult.Remaining
        }

        selfClosing := strings.TrimSpace(remaining) == "/>"

        // 7. Create node
        node := &ast.DynamicComponentNode{
            PathExpression: pathExpression,
            Props:          props,
            SelfClosing:    selfClosing,
        }

        return Result{
            Success:   true,
            Node:      node,
            Remaining: "", // Consumed everything
        }
    }
}
```

**Parser Integration** (`parser/parser.go`):

```go
func TemplateParser() Parser {
    return func(input string) Result {
        // ... existing parsers ...

        // Try dynamic component BEFORE regular component
        // (because <= starts with <)
        if result := DynamicComponentParser()(input); result.Success {
            return result
        }

        if result := ComponentParser()(input); result.Success {
            return result
        }

        // ... rest ...
    }
}
```

### Transformer Implementation

**Location**: `transformer/components.go`

```go
// transformDynamicComponent transforms a dynamic component node
// It resolves the path (if possible) and transforms like a regular component
func transformDynamicComponent(node *ast.DynamicComponentNode, parentDataScope map[string]any) []ast.Node {
    log.Printf("transformDynamicComponent: path=%s, props=%d",
        node.PathExpression, len(node.Props))

    // PHASE 1: Extract variables from path expression
    // Example: "./views/{comp}.html" extracts "comp"
    extractVariablesFromPath(node.PathExpression, parentDataScope)

    // PHASE 2: Try to resolve path at compile time if possible
    // If path is static or all variables have known values, resolve now
    resolvedPath := resolveDynamicPath(node.PathExpression, parentDataScope)

    // PHASE 3: Look up component template
    componentTemplate, exists := GetComponentTemplate(resolvedPath)
    if !exists {
        // For truly dynamic paths (runtime-only resolution),
        // we might need a placeholder or runtime resolution strategy
        log.Printf("WARNING: Dynamic component path not resolved: %s", node.PathExpression)

        // Option 1: Return placeholder (for dev mode)
        // Option 2: Return error node
        // Option 3: Skip rendering (return empty)

        return []ast.Node{
            &ast.TextNode{
                Content: fmt.Sprintf("<!-- Dynamic component not found: %s -->", resolvedPath),
            },
        }
    }

    // PHASE 4: Transform like regular component
    // Use the same logic as transformComponent
    return transformComponentWithTemplate(componentTemplate, node.Props, parentDataScope)
}

// extractVariablesFromPath extracts {variable} references from path string
func extractVariablesFromPath(path string, dataScope map[string]any) {
    // Find all {variable} patterns
    varPattern := regexp.MustCompile(`\{([a-zA-Z_$][a-zA-Z0-9_$]*)\}`)
    matches := varPattern.FindAllStringSubmatch(path, -1)

    for _, match := range matches {
        if len(match) > 1 {
            varName := match[1]
            if _, exists := dataScope[varName]; !exists {
                dataScope[varName] = nil // Add to scope
            }
        }
    }
}

// resolveDynamicPath attempts to resolve path at compile time
func resolveDynamicPath(pathExpr string, dataScope map[string]any) string {
    // Replace {variable} with actual values if known
    resolved := pathExpr

    varPattern := regexp.MustCompile(`\{([a-zA-Z_$][a-zA-Z0-9_$]*)\}`)
    matches := varPattern.FindAllStringSubmatch(pathExpr, -1)

    for _, match := range matches {
        if len(match) > 1 {
            varName := match[1]
            if val, exists := dataScope[varName]; exists && val != nil {
                // Replace {varName} with actual value
                if strVal, ok := val.(string); ok {
                    resolved = strings.Replace(resolved, match[0], strVal, 1)
                }
            }
        }
    }

    return resolved
}

// transformComponentWithTemplate is refactored logic from transformComponent
// to be reusable for both static and dynamic components
func transformComponentWithTemplate(
    template *ast.Template,
    props []ast.Attribute,
    parentDataScope map[string]any,
) []ast.Node {
    // PHASE 1: Create component scope
    componentDataScope := make(map[string]any)

    // PHASE 2: Extract fence data
    if template.Fence != nil {
        collectComponentFenceData(template.Fence, componentDataScope)
    }

    // PHASE 3: Resolve props
    for _, prop := range props {
        resolvedValue := resolvePropValue(prop, parentDataScope)
        componentDataScope[prop.Name] = resolvedValue
    }

    // PHASE 4: Transform body
    transformedChildren := transformNodes(template.Body, componentDataScope, false)

    // PHASE 5: Add data wrapper
    return addComponentDataWrapper(transformedChildren, componentDataScope)
}
```

**Transformer Integration** (`transformer/transformer.go`):

```go
func transformNode(node ast.Node, dataScope map[string]any, isTopLevel bool) []ast.Node {
    switch n := node.(type) {
    // ... existing cases ...

    case *ast.DynamicComponentNode:
        return transformDynamicComponent(n, dataScope)

    // ... rest ...
    }
}
```

### Error Handling

**Missing Component**:
```go
if !exists {
    return []ast.Node{
        &ast.Element{
            TagName: "div",
            Attributes: []ast.Attribute{
                {Name: "class", Value: "error"},
                {Name: "style", Value: "color: red;"},
            },
            Children: []ast.Node{
                &ast.TextNode{
                    Content: fmt.Sprintf("Error: Component not found: %s", resolvedPath),
                },
            },
        },
    }
}
```

**Invalid Path**:
```go
if !isValidComponentPath(resolvedPath) {
    log.Printf("ERROR: Invalid component path: %s", resolvedPath)
    return []ast.Node{}
}
```

## Testing Strategy

### Parser Tests

**File**: `parser/components_test.go`

```go
func TestDynamicComponentParser(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantPath string
        wantProps int
        wantFail bool
    }{
        {
            name:     "Basic dynamic component",
            input:    "<='./path.html' />",
            wantPath: "./path.html",
            wantProps: 0,
        },
        {
            name:     "With variable interpolation",
            input:    "<='./views/{comp}.html' />",
            wantPath: "./views/{comp}.html",
            wantProps: 0,
        },
        {
            name:     "With props",
            input:    "<='path' prop={value} />",
            wantPath: "path",
            wantProps: 1,
        },
        {
            name:     "Double quotes",
            input:    `<="path" />`,
            wantPath: "path",
            wantProps: 0,
        },
        {
            name:     "Multiple variables",
            input:    "<='./views/{dir}/{comp}.html' />",
            wantPath: "./views/{dir}/{comp}.html",
            wantProps: 0,
        },
        {
            name:     "Missing quotes",
            input:    "<=path />",
            wantFail: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := DynamicComponentParser()(tt.input)

            if tt.wantFail {
                assert.False(t, result.Success)
                return
            }

            assert.True(t, result.Success)
            node := result.Node.(*ast.DynamicComponentNode)
            assert.Equal(t, tt.wantPath, node.PathExpression)
            assert.Equal(t, tt.wantProps, len(node.Props))
        })
    }
}
```

### Transformer Tests

**File**: `transformer/components_test.go`

```go
func TestTransformDynamicComponent(t *testing.T) {
    tests := []struct {
        name          string
        pathExpr      string
        dataScope     map[string]any
        expectPath    string
        expectFound   bool
    }{
        {
            name:        "Static path",
            pathExpr:    "./Card.html",
            dataScope:   map[string]any{},
            expectPath:  "./Card.html",
            expectFound: true,
        },
        {
            name:        "Dynamic path with variable",
            pathExpr:    "./views/{comp}.html",
            dataScope:   map[string]any{"comp": "Card"},
            expectPath:  "./views/Card.html",
            expectFound: true,
        },
        {
            name:        "Variable not resolved",
            pathExpr:    "./views/{comp}.html",
            dataScope:   map[string]any{},
            expectPath:  "./views/{comp}.html",
            expectFound: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            node := &ast.DynamicComponentNode{
                PathExpression: tt.pathExpr,
                Props:          []ast.Attribute{},
            }

            result := transformDynamicComponent(node, tt.dataScope)

            if tt.expectFound {
                // Should transform to actual component
                assert.NotEmpty(t, result)
            } else {
                // Should return error/placeholder
                assert.Contains(t, renderNodes(result), "not found")
            }
        })
    }
}
```

### Integration Tests

**File**: `tests/alpine/dynamic_components_test.go`

```go
func TestDynamicComponents(t *testing.T) {
    // Register test components
    renderer.RegisterComponent("Card", cardTemplate)
    renderer.RegisterComponent("List", listTemplate)

    tests := []struct {
        name     string
        template string
        want     string
    }{
        {
            name: "Static dynamic component",
            template: `
                ---
                ---
                <='./Card.html' title="Test" />
            `,
            want: `<div x-data='{...}'>Card content</div>`,
        },
        {
            name: "Dynamic with variable",
            template: `
                ---
                let comp = "Card"
                ---
                <='./views/{comp}.html' />
            `,
            want: `<div x-data='{...}'>Card content</div>`,
        },
        {
            name: "Dynamic in loop",
            template: `
                ---
                let components = ["Card", "List"]
                ---
                {for comp in components}
                    <='./views/{comp}.html' />
                {/for}
            `,
            want: `<template x-for="comp in components">...`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := renderer.Render(tt.template)
            assert.NoError(t, err)
            assert.Contains(t, result, tt.want)
        })
    }
}
```

## Performance Considerations

### Caching Strategy

```go
// Cache resolved component templates
var dynamicComponentCache = make(map[string]*ast.Template)
var cacheMutex sync.RWMutex

func getCachedComponent(path string) (*ast.Template, bool) {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    template, exists := dynamicComponentCache[path]
    return template, exists
}

func cacheComponent(path string, template *ast.Template) {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    dynamicComponentCache[path] = template
}
```

### Path Resolution Optimization

- Resolve paths at compile time when possible
- Only defer to runtime when variables are truly dynamic
- Cache resolved paths to avoid repeated lookups

## Cognitive Load Management

| Function | Complexity | Status |
|----------|-----------|--------|
| DynamicComponentParser | 15 | ✅ OK |
| transformDynamicComponent | 20 | ✅ OK |
| extractVariablesFromPath | 8 | ✅ OK |
| resolveDynamicPath | 12 | ✅ OK |
| transformComponentWithTemplate | 18 | ✅ OK |

Total: 73 (distributed across 5 functions)

## Dependencies

- Component registry: `cmd/server/main.go` (existing)
- GetComponentTemplate: `transformer/components.go` (existing)
- Parser combinators: `parser/parser.go` (existing)
- Transformation pipeline: `transformer/transformer.go` (existing)

## References

- Original implementation: Lines 708-733 of original main.go
- Parser patterns: `parser/components.go`
- Component transformation: `transformer/components.go`
- Comparison doc: `docs/OriginalVsCurrentComparison.md`
