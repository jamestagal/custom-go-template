# Plenti Registry Parity Specification

## Overview

This specification ensures the Go template engine's component registry matches Plenti's existing patterns exactly. As a **drop-in replacement** for Plenti's Svelte frontend, the Go/Alpine.js implementation must use identical:

1. **Signature format** - Path-based component signatures
2. **Registry structure** - `allLayouts` object with semantic hierarchy
3. **Fingerprinted output** - Hash-based directory for cache busting
4. **Content integration** - Same `allContent`, `content`, `allLayouts` props

---

## Review Findings (2026-01-28)

The go-reviewer agent identified several critical issues that this spec addresses:

### BLOCKING Issues Resolved

1. **Dual Registry Problem** - The codebase has TWO `ComponentTemplate` types:
   - `transformer.ComponentTemplate` (server-side transformation)
   - `builder.ComponentTemplate` (client-side registry generation)

   **Solution:** Create a **unified shared type** in `types/component.go`

2. **Phase 2+3 Coupling** - Registry format change without runtime updates breaks lookups

   **Solution:** Merge Phase 2 and Phase 3 into a single atomic phase

### HIGH Priority Issues Resolved

3. **Breaking Change Risk** - Switching from simple names to signatures breaks existing code

   **Solution:** Implement **backward compatibility layer** with dual-lookup support

4. **Missing Regression Tests** - No tests to ensure existing functionality continues

   **Solution:** Add regression test requirements to each phase

### MEDIUM Priority Issues Resolved

5. **Edge Cases** - Nested paths, underscores in names, case sensitivity not addressed

   **Solution:** Document handling rules and add edge case tests

---

## Architecture: Shared Registry Type (Option C)

### Problem: Dual Registry Types

The current codebase has **two separate ComponentTemplate types** serving different purposes:

```go
// transformer/components.go (server-side)
type ComponentTemplate struct {
    Name     string
    Template *ast.Template
    Props    []string
}

// builder/registry_generator.go (client-side)
type ComponentTemplate struct {
    Name string
    AST  *ast.Template
}
```

This creates inconsistency when adding signature support.

### Solution: Unified Type in Shared Package

Create a single canonical type that both packages use:

```go
// types/component.go (NEW)
package types

import "github.com/jimafisk/custom_go_template/ast"

// ComponentTemplate is the canonical component representation used by both
// server-side transformation and client-side registry generation.
//
// This unified type ensures consistency between:
// - transformer.componentTemplateRegistry (build-time transformation)
// - builder.GenerateComponentRegistry() (runtime JavaScript)
type ComponentTemplate struct {
    // Identity
    Name      string // Short name (e.g., "Hero2436")
    Signature string // Plenti signature (e.g., "layouts_components_Hero2436_html")

    // Location
    FilePath  string // Original file path (e.g., "layouts/components/Hero2436.html")
    Category  string // "components", "content", "global", "scripts"

    // Content
    AST   *ast.Template // Parsed template AST
    Props []string      // List of prop names this component accepts
}

// CategoryFromPath extracts the category from a file path.
// Example: "layouts/components/Hero2436.html" → "components"
func CategoryFromPath(filePath string) string {
    // Implementation in types/component.go
}
```

### Migration Path

1. Create `types/component.go` with unified type
2. Update `transformer/components.go` to use `types.ComponentTemplate`
3. Update `builder/registry_generator.go` to use `types.ComponentTemplate`
4. Both registries now share the same type with signature support

---

## Current Plenti Architecture

### Folder Structure

```
layouts/
├── components/     ← Reusable UI components (dynamically loadable)
│   ├── ball.svelte
│   ├── block.svelte
│   └── grid.svelte
│
├── content/        ← Page-type templates (one per content type)
│   ├── _index.svelte
│   ├── blog.svelte
│   └── pages.svelte
│
├── global/         ← Site-wide wrappers (always loaded)
│   ├── html.svelte
│   ├── head.svelte
│   ├── nav.svelte
│   └── footer.svelte
│
└── scripts/        ← Utilities (stores, helpers)
    ├── stores.svelte
    └── make_title.svelte
```

### Signature Format

Plenti converts file paths to signatures by:
1. Replacing `/` with `_`
2. Replacing `.` with `_`
3. Preserving the full path from `layouts/`

| File Path | Signature |
|-----------|-----------|
| `layouts/components/ball.svelte` | `layouts_components_ball_svelte` |
| `layouts/content/blog.svelte` | `layouts_content_blog_svelte` |
| `layouts/global/nav.svelte` | `layouts_global_nav_svelte` |
| `layouts/scripts/stores.svelte` | `layouts_scripts_stores_svelte` |

### Edge Cases

| File Path | Signature | Notes |
|-----------|-----------|-------|
| `layouts/content/_index.html` | `layouts_content__index_html` | Leading underscore preserved |
| `layouts/components/Hero_2436.html` | `layouts_components_Hero_2436_html` | Underscore in name preserved |
| `layouts/components/forms/Input.html` | `layouts_components_forms_Input_html` | Nested path flattened |

### Registry Output (`generated/layouts.js`)

```javascript
export { default as layouts_components_ball_svelte } from "../layouts/components/ball.js";
export { default as layouts_components_block_svelte } from "../layouts/components/block.js";
export { default as layouts_content_blog_svelte } from "../layouts/content/blog.js";
export { default as layouts_global_nav_svelte } from "../layouts/global/nav.js";
// ... all layouts
```

### Fingerprinted Output

```
plenti.json:
  "entrypoint_js": ":fingerprint"

Output:
  public/aQwupMmCDl/    ← Hash changes each build
    ├── bundle.css
    ├── core/main.js
    ├── generated/layouts.js
    └── layouts/components/*.js
```

### Runtime Usage

```svelte
<!-- In content template (blog.svelte) -->
<script>
  export let components, allLayouts;
</script>

{#each components as { name }}
  <svelte:component this={allLayouts["layouts_components_" + name + "_svelte"]} />
{/each}
```

---

## Required Changes for Go/Alpine.js Implementation

### 1. Shared Component Type

**File:** `types/component.go` (NEW)

```go
package types

import (
    "path/filepath"
    "strings"

    "github.com/jimafisk/custom_go_template/ast"
)

// ComponentTemplate is the canonical component representation.
// Used by both server-side transformation and client-side registry generation.
type ComponentTemplate struct {
    // Identity
    Name      string // Short name (e.g., "Hero2436")
    Signature string // Plenti signature (e.g., "layouts_components_Hero2436_html")

    // Location
    FilePath  string // Original file path (e.g., "layouts/components/Hero2436.html")
    Category  string // "components", "content", "global", "scripts"

    // Content
    AST   *ast.Template // Parsed template AST
    Props []string      // List of prop names this component accepts
}

// NewComponentTemplate creates a ComponentTemplate from a file path and AST.
// Automatically generates the Plenti-compatible signature.
func NewComponentTemplate(filePath string, template *ast.Template, props []string) *ComponentTemplate {
    name := ExtractNameFromPath(filePath)
    signature := GenerateSignature(filePath)
    category := CategoryFromPath(filePath)

    return &ComponentTemplate{
        Name:      name,
        Signature: signature,
        FilePath:  filePath,
        Category:  category,
        AST:       template,
        Props:     props,
    }
}

// ExtractNameFromPath extracts the component name from a file path.
// Example: "layouts/components/Hero2436.html" → "Hero2436"
func ExtractNameFromPath(filePath string) string {
    base := filepath.Base(filePath)
    ext := filepath.Ext(base)
    return strings.TrimSuffix(base, ext)
}

// CategoryFromPath extracts the category from a file path.
// Example: "layouts/components/Hero2436.html" → "components"
func CategoryFromPath(filePath string) string {
    normalized := filepath.ToSlash(filePath)
    parts := strings.Split(normalized, "/")

    // Find "layouts" and return the next part
    for i, part := range parts {
        if part == "layouts" && i+1 < len(parts) {
            return parts[i+1]
        }
    }
    return "components" // Default fallback
}

// GenerateSignature converts a file path to a Plenti-compatible signature.
// Example: "layouts/components/Hero2436.html" → "layouts_components_Hero2436_html"
func GenerateSignature(filePath string) string {
    normalized := filepath.ToSlash(filePath)
    normalized = strings.TrimPrefix(normalized, "./")
    signature := strings.ReplaceAll(normalized, "/", "_")
    signature = strings.ReplaceAll(signature, ".", "_")
    return signature
}
```

### 2. Signature Utilities

**File:** `builder/signature.go` (UPDATE existing)

```go
package builder

import (
    "strings"

    "github.com/jimafisk/custom_go_template/types"
)

// Re-export types for backward compatibility
type SignatureInfo = types.SignatureInfo

// GenerateSignature re-exports types.GenerateSignature
var GenerateSignature = types.GenerateSignature

// ParseSignature extracts components from a Plenti-compatible signature.
// Handles edge cases: nested paths, underscores in names.
//
// Examples:
//   layouts_components_Hero2436_html → {Category: "components", Name: "Hero2436"}
//   layouts_components_forms_Input_html → {Category: "components", Name: "forms_Input"}
//   layouts_content__index_html → {Category: "content", Name: "_index"}
func ParseSignature(signature string) SignatureInfo {
    parts := strings.Split(signature, "_")

    // Must have at least 4 parts: layouts, category, name, extension
    if len(parts) < 4 || parts[0] != "layouts" {
        return SignatureInfo{Valid: false}
    }

    category := parts[1]
    if !isValidCategory(category) {
        return SignatureInfo{Valid: false}
    }

    // Extension is always the last part
    extension := parts[len(parts)-1]

    // Name is everything between category and extension
    // This handles nested paths and underscores in names
    nameParts := parts[2 : len(parts)-1]
    name := strings.Join(nameParts, "_")

    return SignatureInfo{
        Valid:     true,
        Category:  category,
        Name:      name,
        Extension: extension,
    }
}

func isValidCategory(category string) bool {
    switch category {
    case "components", "content", "global", "scripts":
        return true
    default:
        return false
    }
}

// ShortNameToSignature converts a short component name to a full Plenti signature.
// This mirrors Plenti's runtime lookup: allLayouts["layouts_components_" + name + "_svelte"]
//
// Example: ShortNameToSignature("Hero2436", "html") → "layouts_components_Hero2436_html"
func ShortNameToSignature(name string, extension string) string {
    return "layouts_components_" + name + "_" + extension
}
```

### 3. Backward Compatible Registry

**File:** `transformer/components.go` (UPDATE)

The transformer registry must support **both** simple names and signatures for backward compatibility:

```go
package transformer

import (
    "log"
    "strings"

    "github.com/jimafisk/custom_go_template/types"
)

// componentTemplateRegistry stores registered component templates
// Keyed by BOTH short name AND full signature for backward compatibility
var componentTemplateRegistry = make(map[string]*types.ComponentTemplate)

// RegisterComponent registers a component template for later use.
// Registers under BOTH short name and full signature for backward compatibility.
func RegisterComponent(component *types.ComponentTemplate) {
    // Register by short name (backward compatibility)
    componentTemplateRegistry[component.Name] = component

    // Register by full signature (Plenti compatibility)
    if component.Signature != "" && component.Signature != component.Name {
        componentTemplateRegistry[component.Signature] = component
    }

    log.Printf("Registered component: %s (signature: %s)", component.Name, component.Signature)
}

// GetComponentTemplate retrieves a component template by name OR signature.
// Supports case-insensitive lookup for backward compatibility.
func GetComponentTemplate(nameOrSignature string) (*types.ComponentTemplate, bool) {
    // Try exact match first (works for both names and signatures)
    if template, exists := componentTemplateRegistry[nameOrSignature]; exists {
        return template, true
    }

    // Try case-insensitive match (backward compatibility)
    if len(nameOrSignature) > 0 {
        capitalizedName := strings.ToUpper(nameOrSignature[:1]) + nameOrSignature[1:]
        if template, exists := componentTemplateRegistry[capitalizedName]; exists {
            return template, true
        }

        lowercasedName := strings.ToLower(nameOrSignature[:1]) + nameOrSignature[1:]
        if template, exists := componentTemplateRegistry[lowercasedName]; exists {
            return template, true
        }
    }

    return nil, false
}
```

### 4. Registry Generator Updates

**File:** `builder/registry_generator.go` (UPDATE)

```go
package builder

import (
    "fmt"
    "strings"

    "github.com/jimafisk/custom_go_template/types"
)

// GenerateComponentRegistry produces Plenti-compatible registry using signatures as keys.
// Also generates short-name aliases for backward compatibility.
func GenerateComponentRegistry(components []*types.ComponentTemplate) string {
    if len(components) == 0 {
        return "export default {};"
    }

    var sb strings.Builder
    sb.WriteString("// Auto-generated Plenti-compatible component registry\n")
    sb.WriteString("// Signature format: layouts_{category}_{name}_{extension}\n")
    sb.WriteString("// Lookup: allLayouts['layouts_components_' + name + '_html']\n\n")
    sb.WriteString("const registry = {\n")

    for i, component := range components {
        // Use Plenti signature as primary key
        key := component.Signature
        if key == "" {
            key = component.Name // Fallback for legacy components
        }

        sb.WriteString(fmt.Sprintf("  '%s': (props) => `", key))
        templateContent := convertToJSTemplate(component.AST)
        sb.WriteString(templateContent)
        sb.WriteString("`")

        if i < len(components)-1 {
            sb.WriteString(",\n")
        } else {
            sb.WriteString("\n")
        }
    }

    sb.WriteString("};\n\n")

    // Add short-name aliases for backward compatibility
    sb.WriteString("// Short-name aliases for backward compatibility\n")
    for _, component := range components {
        if component.Category == "components" && component.Signature != component.Name {
            sb.WriteString(fmt.Sprintf("registry['%s'] = registry['%s'];\n",
                component.Name, component.Signature))
        }
    }

    sb.WriteString("\nexport default registry;\n")
    return sb.String()
}
```

### 5. Runtime Component Resolution

**File:** `static/js/runtime-components.js` (UPDATE)

```javascript
/**
 * Resolve component name to registry signature.
 * Supports both short names ("Hero2436") and full signatures ("layouts_components_Hero2436_html").
 *
 * @param {string} name - Component name or signature
 * @returns {string} - Full Plenti-compatible signature
 */
function resolveComponentSignature(name) {
  // If already a full signature, return as-is
  if (name.startsWith('layouts_')) {
    return name;
  }

  // Convert short name to Plenti signature
  return `layouts_components_${name}_html`;
}

/**
 * Render a dynamic component by name.
 * @param {HTMLElement} el - Container element
 * @param {string} componentName - Component name or signature
 * @param {Object} props - Props to pass to component
 */
async function renderDynamicComponent(el, componentName, props) {
  const registry = await import('/js/component-registry.js').then(m => m.default);

  // Try full signature first
  const signature = resolveComponentSignature(componentName);
  let templateFn = registry[signature];

  // Fallback: try short name directly (backward compatibility)
  if (!templateFn) {
    templateFn = registry[componentName];
  }

  // Fallback: case-insensitive search
  if (!templateFn) {
    const lowerName = componentName.toLowerCase();
    const key = Object.keys(registry).find(k => {
      const parsed = k.split('_');
      const name = parsed.slice(2, -1).join('_');
      return name.toLowerCase() === lowerName;
    });
    if (key) {
      templateFn = registry[key];
    }
  }

  if (!templateFn) {
    console.error(`[Runtime] Component not found: ${componentName} (tried: ${signature})`);
    el.innerHTML = `<div class="component-error">Component not found: ${componentName}</div>`;
    return;
  }

  // Render and hydrate
  const html = templateFn(props);
  el.innerHTML = html;

  if (window.Alpine) {
    Alpine.initTree(el);
  }
}

// Register Alpine.js magic
document.addEventListener('alpine:init', () => {
  Alpine.magic('renderDynamicComponent', () => renderDynamicComponent);
});

export { resolveComponentSignature, renderDynamicComponent };
```

### 6. Fingerprinted Output Directory

**File:** `builder/fingerprint.go` (NEW)

```go
package builder

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
    "path/filepath"
)

// GenerateBuildFingerprint creates a hash of all source files.
func GenerateBuildFingerprint(sourceDir string) (string, error) {
    hasher := sha256.New()

    err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        // Only hash template files
        ext := filepath.Ext(path)
        if ext != ".html" && ext != ".js" && ext != ".css" {
            return nil
        }

        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()

        if _, err := io.Copy(hasher, file); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return "", err
    }

    // Return first 10 chars of hash (like Plenti)
    fullHash := hex.EncodeToString(hasher.Sum(nil))
    return fullHash[:10], nil
}

// BuildOutput represents the fingerprinted output structure.
type BuildOutput struct {
    Fingerprint string
    OutputDir   string // e.g., "public/aQwupMmCDl"
}

// CreateFingerprintedOutput sets up the output directory structure.
func CreateFingerprintedOutput(publicDir, fingerprint string) (*BuildOutput, error) {
    outputDir := filepath.Join(publicDir, fingerprint)

    // Create directory structure matching Plenti
    dirs := []string{
        outputDir,
        filepath.Join(outputDir, "core"),
        filepath.Join(outputDir, "generated"),
        filepath.Join(outputDir, "layouts", "components"),
        filepath.Join(outputDir, "layouts", "content"),
        filepath.Join(outputDir, "layouts", "global"),
    }

    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return nil, err
        }
    }

    return &BuildOutput{
        Fingerprint: fingerprint,
        OutputDir:   outputDir,
    }, nil
}
```

### 7. HTML Output with Fingerprint

**File:** `renderer/html.go` (UPDATE)

```go
// RenderHTMLPage produces HTML with fingerprinted asset paths.
func RenderHTMLPage(page Page, fingerprint string) string {
    return fmt.Sprintf(`<!doctype html>
<html data-content-filepath="%s" lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>%s</title>
  <base href="/">
  <script type="module" src="%s/core/main.js"></script>
  <link rel="stylesheet" href="%s/bundle.css">
</head>
<body>
  %s
</body>
</html>`,
        page.ContentFilePath,
        page.Title,
        fingerprint,
        fingerprint,
        page.Content,
    )
}
```

### 8. Runtime Main.js (Plenti Equivalent)

**File:** `static/js/core/main.js` (NEW)

```javascript
/**
 * Plenti-compatible runtime entry point
 *
 * Mirrors Plenti's core/main.js behavior:
 * 1. Import allContent and allLayouts
 * 2. Find content for current page
 * 3. Hydrate with Alpine.js
 */

import allContent from '../generated/content.js';
import allLayouts from '../generated/layouts.js';

// Get content filepath from HTML attribute (Plenti pattern)
const contentFilepath = document.documentElement.dataset.contentFilepath;

// Find content for current page
const content = allContent.find(c => c.filepath === contentFilepath);

if (!content) {
    console.error(`[Plenti] Content not found for filepath: ${contentFilepath}`);
}

// Initialize Alpine.js with Plenti-compatible data
document.addEventListener('alpine:init', () => {
    // Register global data available to all components
    Alpine.store('plenti', {
        content,
        allContent,
        allLayouts,
        path: location.pathname,
        params: new URLSearchParams(location.search),
    });

    // Register component resolver magic
    Alpine.magic('component', () => {
        return (name) => {
            const signature = `layouts_components_${name}_html`;
            return allLayouts[signature] || allLayouts[name];
        };
    });
});

// Export for direct access
window.allContent = allContent;
window.allLayouts = allLayouts;
window.content = content;

console.log('[Plenti] Runtime initialized', {
    contentFilepath,
    componentCount: Object.keys(allLayouts).length
});
```

---

## Migration Path (Revised)

### Phase 1: Shared Type & Signature System (4 hours)

1. Create `types/component.go` with unified ComponentTemplate
2. Create/update `builder/signature.go` with signature utilities
3. Unit tests for signature generation/parsing
4. Edge case tests (nested paths, underscores, case sensitivity)

### Phase 2+3: Registry & Runtime (Combined - 6 hours)

**CRITICAL:** These must be deployed together to prevent broken intermediate state.

1. Update `transformer/components.go` to use shared type
2. Update `builder/registry_generator.go` to use shared type with signatures
3. Update `static/js/runtime-components.js` with signature resolution
4. Add backward compatibility aliases
5. Regression tests for existing functionality
6. Integration tests for Plenti lookup pattern

### Phase 4: Fingerprinting (4 hours)

1. Create `builder/fingerprint.go`
2. Update build process to generate fingerprinted output
3. Update HTML renderer to use fingerprinted paths
4. Create `core/main.js` entry point

### Phase 5: Content Integration (4 hours)

1. Create `generated/content.js` output
2. Ensure `allContent` format matches Plenti
3. Add `data-content-filepath` to HTML output
4. End-to-end tests with real Plenti content

---

## Verification Checklist

### After Phase 1

- [ ] `types.ComponentTemplate` is the single source of truth
- [ ] `GenerateSignature()` produces Plenti-compatible output
- [ ] `ParseSignature()` handles all edge cases
- [ ] All existing tests still pass

### After Phase 2+3

- [ ] `component-registry.js` uses signatures as keys
- [ ] Short-name aliases work (backward compatibility)
- [ ] `allLayouts["layouts_components_" + name + "_html"]` works
- [ ] Existing `registry["Hero2436"]` still works
- [ ] All existing tests still pass

### After Phase 4

- [ ] Fingerprinted directory created
- [ ] Hash changes when source changes
- [ ] HTML uses fingerprinted paths
- [ ] All tests pass

### After Phase 5

- [ ] `content.js` format matches Plenti
- [ ] `data-content-filepath` enables page lookup
- [ ] Real Plenti content works without modification
- [ ] All tests pass

---

## Success Criteria

1. **Drop-in replacement** - Existing Plenti content JSON works without modification
2. **Same signatures** - `layouts_components_ball_svelte` pattern (with `_html` extension)
3. **Same runtime API** - `allLayouts`, `allContent`, `content` available globally
4. **Same output structure** - Fingerprinted directory with `core/`, `generated/`, `layouts/`
5. **Same hydration pattern** - `data-content-filepath` attribute for page identification
6. **Backward compatible** - Existing code using simple names continues to work
