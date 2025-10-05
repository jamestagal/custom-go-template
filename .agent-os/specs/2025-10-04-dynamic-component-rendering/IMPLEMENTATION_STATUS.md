# Dynamic Component Rendering - Implementation Status

**Date**: 2025-10-04
**Status**: Partially Complete - Placeholders Still Being Created

## What Was Implemented

### ✅ Task 1: Component Registry Normalization (COMPLETE)
- **File**: `cmd/server/main.go:237-250`
- **Implementation**: Components are now registered with multiple keys:
  - Component name: "UserProfile"
  - Relative path: "./components/UserProfile.html"
  - Full path: "examples/components/UserProfile.html"

### ✅ Task 2: Path Variable Resolution (COMPLETE)
- **File**: `transformer/components.go:697-747`
- **Function**: `resolveDynamicPath()`
- **Capabilities**:
  - Removes surrounding backticks, quotes
  - Handles single variable paths: `{path}` → resolves to value
  - Handles embedded variables: `./components/{comp}.html` → `./components/UserProfile.html`
  - Returns unresolved if variable not in scope

### ✅ Backtick Support in Parser (COMPLETE)
- **File**: `parser/components.go:42-53`
- **Implementation**: Parser now recognizes backticks (`` ` ``) as valid quote characters for dynamic component paths

### ⚠️ Task 3: Component Inlining (INCOMPLETE - Wrong Approach)
- **File**: `transformer/components.go:584-651`
- **What Was Implemented**:
  1. `normalizeComponentPath()` helper (lines 44-79) ✅
  2. `transformDynamicComponent()` with path resolution ✅
  3. Component lookup with normalized keys (lines 612-624) ✅
  4. **BUT**: Still creates placeholders via `createDynamicComponentPlaceholder()` (lines 766-798) ❌

### ❌ Task 4: End-to-End Validation (NOT COMPLETE)
- Dynamic components still render as placeholder divs with `x-component-dynamic` attribute
- Browser shows empty divs instead of component content

## The Core Problem

The implementation creates placeholder divs instead of inlining component content:

### Current Behavior (Lines 599-651)
```go
func transformDynamicComponent(node *ast.DynamicComponentNode, parentDataScope map[string]any) []ast.Node {
    // ... path resolution ...

    _, exists := GetComponentTemplate(resolvedPath)
    if !exists {
        // Try normalized keys
        alternativeKeys := normalizeComponentPath(resolvedPath)
        for _, key := range alternativeKeys {
            _, exists = GetComponentTemplate(key)
            if exists {
                resolvedPath = key
                break
            }
        }

        if !exists {
            // ❌ PROBLEM: Returns placeholder even when component NOT found
            return createDynamicComponentPlaceholder(node, resolvedPath, parentDataScope)
        }
    }

    // ✅ This part is correct - delegates to transformComponent
    regularComponentNode := &ast.ComponentNode{
        Name:  resolvedPath,
        Props: node.Props,
    }
    return transformComponent(regularComponentNode, parentDataScope)
}
```

### What's Wrong

The issue is on **line 612**: the code retrieves the component template but **discards it** (`_,  exists :=`).

Then on **line 650**, it calls `transformComponent()` with the component NAME, expecting `transformComponent` to look it up again.

### Expected Behavior (From Spec)

The spec called for:
1. ✅ Resolve path variables → **DONE**
2. ✅ Normalize path for lookup → **DONE**
3. ✅ Retrieve component AST → **DONE** (but discarded!)
4. ❌ Clone component AST → **NOT DONE**
5. ❌ Merge props into component scope → **NOT DONE**
6. ❌ Recursively transform inlined component → **NOT DONE** (delegates to transformComponent instead)
7. ❌ Wrap in x-data div if needed → **NOT DONE**

## Why It's Still Creating Placeholders

Looking at the `transformDynamicComponent` logic:

1. Component is found with normalized key
2. **BUT** the template is discarded (`_, exists :=`)
3. Code creates a `ComponentNode` with the resolved name
4. Calls `transformComponent()` which must be looking it up again
5. **Hypothesis**: `transformComponent()` might not be finding it, OR it's creating placeholders for dynamic components

## Next Steps to Fix

### Option 1: Fix transformDynamicComponent (Recommended)

Change lines 612-651 to:

```go
// PHASE 3: Look up component template (COGNITIVE LOAD: 7)
componentTemplate, exists := GetComponentTemplate(resolvedPath)
if !exists {
    // Try with normalized path variants
    log.Printf("transformDynamicComponent: Component not found with path '%s', trying normalized keys", resolvedPath)
    alternativeKeys := normalizeComponentPath(resolvedPath)
    for _, key := range alternativeKeys {
        componentTemplate, exists = GetComponentTemplate(key)
        if exists {
            log.Printf("transformDynamicComponent: Found component with alternative key: %s", key)
            resolvedPath = key // Use the key that worked
            break
        }
    }

    if !exists {
        // Path couldn't be resolved at build time - check if it still has variables
        if strings.Contains(resolvedPath, "{") {
            log.Printf("transformDynamicComponent: Path contains unresolved variables: %s", resolvedPath)
            // Return error - we can't inline components with unresolved variables
            return []ast.Node{&ast.TextNode{
                Content: fmt.Sprintf("<!-- ERROR: Dynamic component path has unresolved variables: %s -->", resolvedPath),
            }}
        }

        log.Printf("ERROR: Dynamic component not found: %s (tried keys: %v)", resolvedPath, alternativeKeys)
        return []ast.Node{&ast.TextNode{
            Content: fmt.Sprintf("<!-- ERROR: Component not found: %s -->", resolvedPath),
        }}
    }
}

// PHASE 4: Inline the component (COGNITIVE LOAD: 8)
// We have the component template - inline it directly
log.Printf("transformDynamicComponent: Inlining component template for: %s", resolvedPath)

// Clone the component AST to avoid mutation
clonedTemplate := cloneTemplate(componentTemplate.Template)

// Build prop scope for component
propScope := make(map[string]any)
for _, prop := range node.Props {
    if prop.IsDynamic {
        // Evaluate dynamic prop from current scope
        value, exists := parentDataScope[prop.Value]
        if exists {
            propScope[prop.Name] = value
        } else {
            // It's an expression, keep as-is for Alpine
            propScope[prop.Name] = prop.Value
        }
    } else {
        // Static value
        propScope[prop.Name] = prop.Value
    }
}

// Merge parent scope with prop scope
mergedScope := mergeMaps(parentDataScope, propScope)

// Transform the component's nodes with merged scope
transformedNodes := []ast.Node{}
for _, childNode := range clonedTemplate.RootNodes {
    // Recursively transform each node
    transformed := transformNode(childNode, mergedScope)
    transformedNodes = append(transformedNodes, transformed...)
}

// Wrap in x-data div if component has props
if len(propScope) > 0 {
    wrapper := &ast.Element{
        Tag: "div",
        Attributes: []ast.Attribute{
            {
                Name:  "x-data",
                Value: formatComponentProps(propScope),
            },
        },
        Children: transformedNodes,
    }
    return []ast.Node{wrapper}
}

// No wrapper needed, return inlined nodes directly
return transformedNodes
```

### Required Helper Functions

1. **`cloneTemplate()`** - Deep copy AST to avoid mutation
2. **`mergeMaps()`** - Merge parent and prop scopes
3. **`transformNode()`** - Recursively transform a single node

## Test Plan

After fixing, test with:
```bash
curl http://localhost:3000 | grep -A 20 "Dynamic Component Examples"
```

Should see actual UserProfile HTML content, NOT:
```html
<div x-component-dynamic="./components/UserProfile.html" ...></div>
```

## Summary

**What Works**:
- ✅ Component registry with multiple lookup keys
- ✅ Path variable resolution
- ✅ Backtick parsing
- ✅ Normalized path lookups

**What's Broken**:
- ❌ Component template is found but discarded
- ❌ No AST cloning
- ❌ No direct component inlining
- ❌ Still creates placeholder divs

**Root Cause**: Line 612 retrieves but discards the component template, then line 645-650 delegates to `transformComponent()` which likely can't find it or creates placeholders.

**Fix**: Inline the component directly in `transformDynamicComponent()` instead of delegating.
