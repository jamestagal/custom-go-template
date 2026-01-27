# Plenti Registry Parity Specification

## Overview

This specification ensures the Go template engine's component registry matches Plenti's existing patterns exactly. As a **drop-in replacement** for Plenti's Svelte frontend, the Go/Alpine.js implementation must use identical:

1. **Signature format** - Path-based component signatures
2. **Registry structure** - `allLayouts` object with semantic hierarchy
3. **Fingerprinted output** - Hash-based directory for cache busting
4. **Content integration** - Same `allContent`, `content`, `allLayouts` props

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

### 1. File Structure Alignment

**Current Go SSG structure:**
```
layouts/
├── components/     ← ✅ Already matches
│   └── Hero2436.html
├── content/        ← ✅ Already matches
│   └── pages.html
└── global/         ← ✅ Already matches
    └── html.html
```

**Required:** Rename `.html` extension handling to produce `_html` in signatures (not `_svelte`).

### 2. Signature Generation

**File:** `builder/signature.go` (NEW)

```go
package builder

import (
    "path/filepath"
    "strings"
)

// GenerateSignature converts a layout file path to a Plenti-compatible signature.
//
// Examples:
//   layouts/components/Hero2436.html → layouts_components_Hero2436_html
//   layouts/content/blog.html → layouts_content_blog_html
//   layouts/global/nav.html → layouts_global_nav_html
//
// Pattern: Pure Function [Load: 3]
func GenerateSignature(filePath string) string {
    // Normalize path separators
    normalized := filepath.ToSlash(filePath)

    // Remove leading ./ if present
    normalized = strings.TrimPrefix(normalized, "./")

    // Replace / and . with _
    signature := strings.ReplaceAll(normalized, "/", "_")
    signature = strings.ReplaceAll(signature, ".", "_")

    return signature
}

// ParseSignature extracts components from a signature.
//
// Example:
//   layouts_components_Hero2436_html → {
//     Category: "components",
//     Name: "Hero2436",
//     Extension: "html",
//   }
//
// Pattern: Parser Function [Load: 5]
func ParseSignature(signature string) SignatureInfo {
    parts := strings.Split(signature, "_")

    if len(parts) < 4 || parts[0] != "layouts" {
        return SignatureInfo{Valid: false}
    }

    return SignatureInfo{
        Valid:     true,
        Category:  parts[1],  // components, content, global, scripts
        Name:      strings.Join(parts[2:len(parts)-1], "_"),
        Extension: parts[len(parts)-1],
    }
}

type SignatureInfo struct {
    Valid     bool
    Category  string  // "components", "content", "global", "scripts"
    Name      string  // "Hero2436", "blog", "nav"
    Extension string  // "html", "svelte"
}

// IsComponent returns true if the signature is a dynamically loadable component.
func (s SignatureInfo) IsComponent() bool {
    return s.Valid && s.Category == "components"
}

// IsGlobal returns true if the signature is a global layout (always loaded).
func (s SignatureInfo) IsGlobal() bool {
    return s.Valid && s.Category == "global"
}

// IsContentTemplate returns true if the signature is a content type template.
func (s SignatureInfo) IsContentTemplate() bool {
    return s.Valid && s.Category == "content"
}
```

### 3. Registry Generator Updates

**File:** `builder/registry_generator.go` (UPDATE)

```go
// ComponentTemplate now includes signature
type ComponentTemplate struct {
    Name      string        // Original name (e.g., "Hero2436")
    Signature string        // Plenti signature (e.g., "layouts_components_Hero2436_html")
    FilePath  string        // Original file path
    Category  string        // "components", "content", "global", "scripts"
    AST       *ast.Template
}

// GenerateComponentRegistry produces Plenti-compatible registry
func GenerateComponentRegistry(components []ComponentTemplate) string {
    if len(components) == 0 {
        return "export default {};"
    }

    var sb strings.Builder
    sb.WriteString("// Auto-generated Plenti-compatible component registry\n")
    sb.WriteString("// Signature format: layouts_{category}_{name}_{extension}\n\n")
    sb.WriteString("export default {\n")

    for i, component := range components {
        // Use Plenti signature format
        sb.WriteString(fmt.Sprintf("  '%s': (props) => `", component.Signature))

        templateContent := convertToJSTemplate(component.AST)
        sb.WriteString(templateContent)

        sb.WriteString("`")

        if i < len(components)-1 {
            sb.WriteString(",\n")
        } else {
            sb.WriteString("\n")
        }
    }

    sb.WriteString("};\n")
    return sb.String()
}
```

### 4. Registry Output Format

**File:** `static/js/component-registry.js` (UPDATED FORMAT)

```javascript
// Auto-generated Plenti-compatible component registry
// Signature format: layouts_{category}_{name}_{extension}

export default {
  // Components (dynamically loadable)
  'layouts_components_Hero2436_html': (props) => `<section class="hero">...</section>`,
  'layouts_components_Services2437_html': (props) => `<section class="services">...</section>`,
  'layouts_components_FAQ2438_html': (props) => `<div class="faq">...</div>`,

  // Content templates (page types)
  'layouts_content_pages_html': (props) => `<div class="page">...</div>`,
  'layouts_content_blog_html': (props) => `<article class="blog">...</article>`,
  'layouts_content__index_html': (props) => `<div class="home">...</div>`,

  // Global (always loaded - may not need in registry)
  'layouts_global_html_html': (props) => `<!DOCTYPE html>...`,
  'layouts_global_nav_html': (props) => `<nav>...</nav>`,
  'layouts_global_footer_html': (props) => `<footer>...</footer>`,
};
```

### 5. Content JSON Compatibility

**Current Plenti format:**
```json
{
  "components": [
    { "name": "ball" },
    { "name": "block" }
  ]
}
```

**Template usage (Plenti Svelte):**
```svelte
{#each components as { name }}
  <svelte:component this={allLayouts["layouts_components_" + name + "_svelte"]} />
{/each}
```

**Template usage (Go/Alpine.js):**
```html
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Runtime resolution must build the signature:**
```javascript
// In runtime-components.js
function resolveComponentSignature(name) {
  // Convert short name to full signature
  // "Hero2436" → "layouts_components_Hero2436_html"
  return `layouts_components_${name}_html`;
}

async function renderDynamicComponent(el, componentName, props) {
  const signature = resolveComponentSignature(componentName);
  const templateFn = registry[signature];
  // ...
}
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
//
// Pattern: Hash Function [Load: 8]
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
    OutputDir   string  // e.g., "public/aQwupMmCDl"
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
            return allLayouts[signature];
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

## Migration Path

### Phase 1: Signature System (4 hours)

1. Create `builder/signature.go`
2. Update `builder/registry_generator.go` to use signatures
3. Update component registration to include file paths
4. Unit tests for signature generation/parsing

### Phase 2: Registry Format (4 hours)

1. Update registry output format to use signatures
2. Update `runtime-components.js` to resolve signatures
3. Add signature → name and name → signature helpers
4. Integration tests

### Phase 3: Fingerprinting (4 hours)

1. Create `builder/fingerprint.go`
2. Update build process to generate fingerprinted output
3. Update HTML renderer to use fingerprinted paths
4. Create `core/main.js` entry point

### Phase 4: Content Integration (4 hours)

1. Create `generated/content.js` output
2. Ensure `allContent` format matches Plenti
3. Add `data-content-filepath` to HTML output
4. End-to-end tests with real Plenti content

---

## Verification Checklist

### Signature Compatibility

- [ ] `layouts/components/Hero2436.html` → `layouts_components_Hero2436_html`
- [ ] `layouts/content/pages.html` → `layouts_content_pages_html`
- [ ] `layouts/global/html.html` → `layouts_global_html_html`
- [ ] Signature can be parsed back to path

### Registry Compatibility

- [ ] Registry exports `allLayouts` object
- [ ] Signatures are keys (not component names)
- [ ] `allLayouts["layouts_components_" + name + "_html"]` works
- [ ] Template functions accept `props` parameter

### Output Compatibility

- [ ] Fingerprinted directory created (e.g., `public/aQwupMmCDl/`)
- [ ] `core/main.js` exists and initializes runtime
- [ ] `generated/layouts.js` contains registry
- [ ] `generated/content.js` contains allContent
- [ ] HTML includes `data-content-filepath` attribute

### Runtime Compatibility

- [ ] `window.allLayouts` available
- [ ] `window.allContent` available
- [ ] `window.content` contains current page data
- [ ] Dynamic component resolution works with short names

---

## Example: Full Build Output

```
public/
├── aQwupMmCDl/                    ← Fingerprinted directory
│   ├── core/
│   │   └── main.js                ← Runtime entry point
│   ├── generated/
│   │   ├── content.js             ← allContent data
│   │   └── layouts.js             ← allLayouts registry
│   ├── layouts/
│   │   ├── components/
│   │   │   ├── Hero2436.js        ← Individual component (optional)
│   │   │   └── Services2437.js
│   │   ├── content/
│   │   │   └── pages.js
│   │   └── global/
│   │       └── nav.js
│   └── bundle.css
├── index.html
├── about/
│   └── index.html
└── blog/
    └── index.html
```

**index.html:**
```html
<!doctype html>
<html data-content-filepath="content/_index.json" lang="en">
<head>
  <script type="module" src="aQwupMmCDl/core/main.js"></script>
  <link rel="stylesheet" href="aQwupMmCDl/bundle.css">
</head>
<body>
  <!-- Pre-rendered static HTML -->
  <main x-data>
    <!-- Alpine.js hydrates this -->
  </main>
</body>
</html>
```

---

## Success Criteria

1. **Drop-in replacement** - Existing Plenti content JSON works without modification
2. **Same signatures** - `layouts_components_ball_svelte` pattern (with `_html` extension)
3. **Same runtime API** - `allLayouts`, `allContent`, `content` available globally
4. **Same output structure** - Fingerprinted directory with `core/`, `generated/`, `layouts/`
5. **Same hydration pattern** - `data-content-filepath` attribute for page identification
