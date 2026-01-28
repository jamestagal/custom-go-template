# Component Style Aggregation Specification

**Date:** 2025-10-07
**Status:** Draft
**Priority:** High (Fixes HeaderSimple flashing issue)

## Problem Statement

Currently, component scoped styles defined in `<style>` tags within component files are not automatically extracted and included in the parent page output. This causes:

1. **Visual Bugs**: Components like `HeaderSimple` flash and disappear because their styles aren't loaded
2. **Manual Workarounds**: Developers must manually copy component styles to parent pages
3. **Maintenance Burden**: Style changes require updates in multiple places
4. **Inconsistency**: Doesn't match Svelte/Plenti behavior where component styles are automatically bundled

### Current Behavior

```html
<!-- examples/components/HeaderSimple.html -->
---
---

<style>
  .header { background-color: #f8f9fa; }
  .brand svg { height: 32px; }
</style>

<header class="header">
  <a href="/" class="brand"><svg>...</svg></a>
</header>
```

**Problem**: When `HeaderSimple` is imported into `home.html`, the styles are NOT included in the rendered output, causing the component to render without styling.

**Current Workaround**: Manually copy styles from `HeaderSimple.html` to `home.html`'s `<style>` section.

### Desired Behavior

Automatically extract and aggregate styles from all components used in a page, similar to how Svelte/Plenti handles it:

1. Parse `<style>` blocks from component templates
2. Track component dependency tree (which components import which)
3. Aggregate styles in correct order (dependencies first)
4. Deduplicate identical style blocks
5. Inject aggregated styles into page output

## Architectural Overview

### Current System Architecture

```
Template Source → Parser → AST → Transformer → Renderer → HTML Output
                                                            └─ No styles from components
```

### Proposed Architecture

```
Template Source → Parser → AST (with StyleSection) → Transformer → Renderer → HTML Output
                                                                      ├─ Component tree traversal
                                                                      ├─ Style aggregation
                                                                      └─ Style injection
```

### Data Flow

1. **Parsing Phase**: Parser extracts `<style>` blocks into `StyleSection` AST nodes
2. **Registration Phase**: Components register with their parsed template (already happens)
3. **Rendering Phase**:
   - Renderer identifies which components are used on the page
   - Traverses component dependency tree
   - Aggregates styles from all components (bottom-up: dependencies first)
   - Deduplicates styles based on content hash
   - Injects aggregated styles into page `<head>` or main `<style>` section

## Implementation Details

### Phase 1: AST Enhancement

**Status**: ✅ Already exists!

The `StyleSection` node already exists in `ast/ast.go`:

```go
// StyleSection represents the style section
type StyleSection struct {
    Content string
}

func (s *StyleSection) NodeType() string { return "StyleSection" }
```

**No changes needed** - the AST is ready.

### Phase 2: Parser Enhancement

**File**: `parser/parser.go`

**Current State**: Parser likely extracts `<style>` but may not store in `Template.RootNodes`

**Changes Needed**:

1. Ensure `<style>` blocks are parsed into `StyleSection` nodes
2. Add `StyleSection` nodes to `Template.RootNodes`
3. Extract styles from fence section AND body

**Example**:

```go
// In parser.go
func parseStyleSection(input string, pos int) (*ast.StyleSection, int, error) {
    // Find <style> tag
    startTag := "<style"
    endTag := "</style>"

    startIdx := strings.Index(input[pos:], startTag)
    if startIdx == -1 {
        return nil, pos, nil // No style section
    }

    // Find closing >
    openEnd := strings.Index(input[pos+startIdx:], ">")
    contentStart := pos + startIdx + openEnd + 1

    // Find </style>
    endIdx := strings.Index(input[contentStart:], endTag)
    if endIdx == -1 {
        return nil, pos, fmt.Errorf("unclosed <style> tag")
    }

    content := input[contentStart : contentStart+endIdx]

    return &ast.StyleSection{
        Content: strings.TrimSpace(content),
    }, contentStart + endIdx + len(endTag), nil
}
```

### Phase 3: Component Template Storage

**File**: `transformer/components.go`

**Current State**: Already stores component templates in registry

```go
type ComponentTemplate struct {
    Name     string
    Template *ast.Template  // Contains RootNodes including StyleSection
    Props    []string
}
```

**Changes Needed**: ✅ None - already stores full template with styles

### Phase 4: Style Aggregation Logic

**New File**: `renderer/styles.go`

Create a new module responsible for:
1. Traversing component dependency tree
2. Collecting styles from components
3. Deduplication
4. Ordering

```go
package renderer

import (
    "crypto/sha256"
    "fmt"
    "strings"

    "github.com/jimafisk/custom_go_template/ast"
    "github.com/jimafisk/custom_go_template/transformer"
)

// StyleBlock represents a style block with metadata
type StyleBlock struct {
    Content    string // CSS content
    Source     string // Component name
    Hash       string // Content hash for deduplication
}

// AggregateComponentStyles traverses the component tree and aggregates styles
//
// Algorithm:
// 1. Use depth-first traversal to collect styles (dependencies first)
// 2. Track visited components to prevent infinite loops
// 3. Deduplicate based on content hash
// 4. Return aggregated CSS string
func AggregateComponentStyles(rootTemplate *ast.Template, componentName string) string {
    visited := make(map[string]bool)
    styleBlocks := make(map[string]*StyleBlock) // Hash -> StyleBlock
    var orderedHashes []string // Preserve order

    // Recursive collection function
    var collectStyles func(template *ast.Template, name string)
    collectStyles = func(template *ast.Template, name string) {
        // Prevent infinite loops
        if visited[name] {
            return
        }
        visited[name] = true

        // First, collect styles from imported components (dependencies)
        for _, node := range template.RootNodes {
            if fence, ok := node.(*ast.FenceSection); ok {
                for _, imp := range fence.Imports {
                    if compTemplate, exists := transformer.GetComponentTemplate(imp.Name); exists {
                        collectStyles(compTemplate.Template, imp.Name)
                    }
                }
            }
        }

        // Then collect styles from this component
        for _, node := range template.RootNodes {
            if styleSection, ok := node.(*ast.StyleSection); ok {
                if styleSection.Content == "" {
                    continue
                }

                // Create hash for deduplication
                hash := fmt.Sprintf("%x", sha256.Sum256([]byte(styleSection.Content)))

                // Only add if not already present
                if _, exists := styleBlocks[hash]; !exists {
                    styleBlocks[hash] = &StyleBlock{
                        Content: styleSection.Content,
                        Source:  name,
                        Hash:    hash,
                    }
                    orderedHashes = append(orderedHashes, hash)
                }
            }
        }
    }

    // Start collection from root
    collectStyles(rootTemplate, componentName)

    // Build aggregated CSS with source comments
    var result strings.Builder
    for _, hash := range orderedHashes {
        block := styleBlocks[hash]

        // Add comment indicating source component
        result.WriteString(fmt.Sprintf("/* Styles from: %s */\n", block.Source))
        result.WriteString(block.Content)
        result.WriteString("\n\n")
    }

    return result.String()
}
```

### Phase 5: Renderer Integration

**File**: `renderer/render.go`

**Changes Needed**: Modify rendering to inject aggregated styles

```go
// In RenderTemplate function
func RenderTemplate(template *ast.Template, componentName string) (string, error) {
    // ... existing rendering logic ...

    // Aggregate styles from component tree
    aggregatedStyles := AggregateComponentStyles(template, componentName)

    // Inject styles into output
    // Option A: In <head> section
    // Option B: In main <style> block after existing styles

    var output strings.Builder

    // Render HTML structure...

    // Inject aggregated styles
    if aggregatedStyles != "" {
        output.WriteString("<style>\n")
        output.WriteString(aggregatedStyles)
        output.WriteString("</style>\n")
    }

    // ... rest of rendering ...

    return output.String(), nil
}
```

### Phase 6: Caching Strategy

For performance, cache aggregated styles per component:

```go
// Global cache
var componentStyleCache = make(map[string]string)
var styleCacheMutex sync.RWMutex

func GetAggregatedStyles(template *ast.Template, componentName string) string {
    // Check cache first
    styleCacheMutex.RLock()
    cached, exists := componentStyleCache[componentName]
    styleCacheMutex.RUnlock()

    if exists {
        return cached
    }

    // Generate and cache
    aggregated := AggregateComponentStyles(template, componentName)

    styleCacheMutex.Lock()
    componentStyleCache[componentName] = aggregated
    styleCacheMutex.Unlock()

    return aggregated
}

// Clear cache when components are re-registered (dev mode)
func ClearStyleCache() {
    styleCacheMutex.Lock()
    componentStyleCache = make(map[string]string)
    styleCacheMutex.Unlock()
}
```

## Developer Experience

### No Changes to Template Syntax

Developers continue writing components exactly as before:

```html
<!-- Component: examples/components/MyComponent.html -->
---
prop title = "Default"
---

<style>
  .my-component { color: blue; }
</style>

<div class="my-component">{title}</div>
```

### Automatic Style Inclusion

When `MyComponent` is imported and used in a page, its styles are automatically included:

```html
<!-- Page: examples/pages/about.html -->
---
import MyComponent from "./components/MyComponent.html"
---

<MyComponent title="About Us" />

<!-- Output HTML will include MyComponent's styles automatically -->
```

### Debugging: Source Comments

The aggregated styles include comments showing which component contributed each block:

```html
<style>
/* Styles from: HeaderSimple */
.header { background-color: #f8f9fa; }
.brand svg { height: 32px; }

/* Styles from: Footer */
.footer { background-color: #333; }
</style>
```

## Testing Strategy

### Unit Tests

**File**: `renderer/styles_test.go`

```go
func TestStyleAggregation_SingleComponent(t *testing.T) {
    // Test: Component with one style block
}

func TestStyleAggregation_NestedComponents(t *testing.T) {
    // Test: Parent imports Child, both have styles
    // Verify: Child styles come before Parent styles
}

func TestStyleAggregation_Deduplication(t *testing.T) {
    // Test: Two components with identical styles
    // Verify: Only one copy in output
}

func TestStyleAggregation_CircularDependency(t *testing.T) {
    // Test: A imports B, B imports A
    // Verify: No infinite loop, styles included once
}

func TestStyleAggregation_EmptyStyles(t *testing.T) {
    // Test: Component with empty <style> block
    // Verify: No empty comments in output
}

func TestStyleAggregation_NoStyles(t *testing.T) {
    // Test: Component without <style> block
    // Verify: No errors, empty string returned
}
```

### Integration Tests

**File**: `tests/components/style_aggregation_test.go`

```go
func TestStyleAggregation_RealWorld_HeaderSimple(t *testing.T) {
    // Test: HeaderSimple component in home.html
    // Verify: Header styles are included in output
    // Verify: No flashing/disappearing
}

func TestStyleAggregation_MultipleComponentsOnPage(t *testing.T) {
    // Test: Page uses Header, Footer, Sidebar
    // Verify: All component styles included
    // Verify: Correct order (dependencies first)
}
```

### Manual Testing

1. Start dev server: `go run cmd/server/main.go`
2. Visit `http://localhost:3000`
3. Verify HeaderSimple displays correctly without flashing
4. Inspect HTML source - verify styles from HeaderSimple are present
5. Remove HeaderSimple import - verify styles are not included

## Performance Considerations

### Build-Time Overhead

- **Style Parsing**: Minimal - already parsing HTML
- **Tree Traversal**: O(n) where n = number of components used
- **Deduplication**: O(m) where m = number of style blocks
- **Hash Calculation**: SHA256 on CSS strings - fast for small styles

### Runtime Performance

- **Caching**: Aggregated styles cached per component
- **Cache Invalidation**: Only in dev mode when components change
- **Production**: Cache never cleared, styles aggregated once

### Memory Usage

- **Style Content**: Stored in AST nodes (already in memory)
- **Cache**: Stores aggregated CSS strings per component
- **Typical Size**: ~1-10KB per component = negligible

## Edge Cases

### Multiple Style Blocks in One Component

```html
<style>
  .header { }
</style>

<header>...</header>

<style>
  .footer { }
</style>
```

**Handling**: Extract all `<style>` blocks, concatenate in order

### Styles in Fence Section vs Body

Fence sections don't support `<style>` tags. Only body `<style>` blocks are extracted.

### Dynamic Component Imports

For dynamic components (`<='./path/{var}.html' />`), styles are aggregated at render time based on resolved path.

### Circular Dependencies

A imports B, B imports A:

**Handling**: `visited` map prevents infinite loops. Each component's styles included exactly once in dependency order.

## Migration Path

### Phase 1: Implementation (This Spec)
- Implement style aggregation
- Add tests
- Fix HeaderSimple flashing issue

### Phase 2: Remove Manual Workarounds
- Remove duplicate styles from `home.html`
- Clean up other pages with manual style copying

### Phase 3: Documentation
- Update CLAUDE.md with style aggregation behavior
- Add developer guide for component styles

## Future Enhancements (Out of Scope)

### Optional CSS Scoping

Add opt-in scoping for components that need isolation:

```html
---
scoped = true
---

<style>
  /* These styles get scoped to this component */
  .item { color: blue; }
</style>
```

See `docs/FutureDevelopment.md` for detailed scoping plans.

### CSS Minification

Minify aggregated styles in production builds.

### Source Maps

Generate CSS source maps for debugging.

## Success Criteria

1. ✅ HeaderSimple displays correctly without flashing
2. ✅ Component styles automatically included in pages
3. ✅ No manual style copying needed
4. ✅ Correct ordering (dependencies first)
5. ✅ Deduplication prevents duplicate styles
6. ✅ All tests pass
7. ✅ No performance degradation (<10ms overhead)
8. ✅ Documentation updated

## References

- **Current Issue**: HeaderSimple flashing (`examples/components/HeaderSimple.html`)
- **Svelte Behavior**: Component styles automatically scoped and bundled
- **AST Definition**: `ast/ast.go` - StyleSection already exists
- **Component Registry**: `transformer/components.go` - Already stores templates
- **Future Scoping**: `docs/FutureDevelopment.md` - Optional scoping enhancement
