# Plenti + Custom Go Template Integration Specification
## Replacing Svelte with Custom Go Templating Engine

---

## 1. Executive Summary

This specification outlines how to **integrate the custom Go templating engine into Plenti** to replace Svelte as the templating layer, while keeping Plenti's existing:
- ✅ Content discovery system (`content/` JSON files)
- ✅ Routing system (URL → content mapping)
- ✅ Path generation (custom routes from `plenti.json`)
- ✅ Build pipeline (Go backend)
- ✅ Magic variables system (`content`, `allContent`, `env`, etc.)
- ✅ CMS integration (Git-backed editing)

**What Changes:**
- ❌ Remove Svelte compilation (`compile.go`)
- ❌ Replace `.svelte` templates with `.html` templates
- ✅ Add custom Go template rendering
- ✅ Keep existing fence section syntax (with modifications)
- ✅ Simplify or remove client-side SPA hydration

---

## 2. Current Plenti Architecture (What to Keep)

### 2.1 Content Discovery (KEEP AS-IS)
**File:** `cmd/build/data_source.go`

Plenti already:
- Scans `content/` directory recursively
- Reads all `.json` files
- Extracts content types from directory structure
- Generates magic variables:
  - `content` - current page data
  - `allContent` - array of all content nodes
  - `env` - environment config

**Data Structure:**
```go
type content struct {
    contentType      string  // "blog", "pages", etc.
    contentPath      string  // "/blog/my-post"
    contentDest      string  // "public/blog/my-post/index.html"
    contentDetails   string  // JSON string of content node
    contentFilepath  string  // "content/blog/my-post.json"
    contentFilename  string  // "my-post"
    contentFields    string  // JSON string of user fields
    contentPagerDest string  // For pagination
    contentPagerPath string  // For pagination
    contentPagerNums []string // Pagination numbers
}
```

**Magic Variables Injected:**
```javascript
// Currently in JavaScript - we'll adapt to Go templates
{
  content: {
    pager: null,
    type: "blog",
    path: "/blog/post1",
    filepath: "content/blog/post1.json",
    filename: "post1",
    fields: { /* user-defined JSON */ }
  },
  allContent: [ /* array of all content nodes */ ],
  layout: /* component reference */,
  env: {
    local: true,
    baseurl: "/",
    routes: { /* custom routes */ },
    types: ["blog", "pages"],
    fingerprint: "abc123",
    sitevars: { /* custom vars */ },
    cms: { /* CMS config */ }
  }
}
```

### 2.2 Routing System (KEEP AS-IS)
**File:** `cmd/build/data_source.go` (functions: `evalRouteReplacementPatterns`, `generatePath`)

Plenti already handles:
- Default paths: `content/blog/post.json` → `/blog/post`
- Custom routes from `plenti.json`:
  ```json
  {
    "routes": {
      "pages": "/:filename",
      "blog": "/blog/:fields(author)/:fields(title)"
    }
  }
  ```
- Pagination: `:paginate(totalPages)` replacement
- Slugification (lowercase, hyphens, clean URLs)

### 2.3 Build Pipeline (MODIFY)
**Current Flow (Svelte):**
```
1. data_source.go → Read content JSON files
2. compile.go → Compile .svelte files to JS
3. bundle.go → Bundle all JS together
4. createHTML() → Use V8 SSR to generate static HTML
5. Write to public/ directory
```

**New Flow (Go Templates):**
```
1. data_source.go → Read content JSON files (KEEP)
2. render_templates.go → Render .html files with Go engine (NEW)
3. createHTML() → Direct HTML output (SIMPLIFIED)
4. Write to public/ directory (KEEP)
5. Optional: Minimal client-side JS for Alpine.js (SIMPLIFIED)
```

---

## 3. Integration Points

### 3.1 Template File Format Changes

#### Current Svelte Template
**File:** `layouts/content/blog.svelte`
```svelte
<script>
  export let title, author, date, body;
</script>

<article>
  <h1>{title}</h1>
  <p>By {author} on {date}</p>
  <div>{@html body}</div>
</article>

<style>
  article {
    max-width: 800px;
  }
</style>
```

#### New Go Template
**File:** `layouts/content/blog.html`
```html
---
// No props needed - data comes from content.fields
// But we can still use fence for imports and functions
import Header from "../components/header.html";
import Footer from "../components/footer.html";
---

<!DOCTYPE html>
<html>
<head>
  <title>{content.fields.title}</title>
</head>
<body>
  <Header />
  
  <article>
    <h1>{content.fields.title}</h1>
    <p>By {content.fields.author} on {content.fields.date}</p>
    <div>{content.fields.body}</div>
  </article>
  
  <Footer />
</body>
</html>

<style>
  article {
    max-width: 800px;
  }
</style>
```

**Key Differences:**
1. **No `export let` statements** - data comes from injected `content` object
2. **Access via `content.fields.*`** - all user fields are under this key
3. **Keep fence section** for imports and helper functions
4. **Same component syntax** - `<Header />` still works
5. **Same conditionals/loops** - `{if}`, `{for}` syntax unchanged

### 3.2 Global Entrypoint

#### Current: `layouts/global/html.svelte`
```svelte
<script>
  import Head from './head.svelte';
  export let content, layout, env;
</script>

<html lang="en">
  <Head title={content.filename} {env} />
  <body>
    <svelte:component this={layout} {...content.fields} />
  </body>
</html>
```

#### New: `layouts/global/html.html`
```html
---
import Head from './head.html';
// Magic variables are auto-injected:
// - content (current page)
// - allContent (all pages)
// - env (environment)
// - layout (component signature)
---

<!DOCTYPE html>
<html lang="en">
  <Head title={content.filename} env={env} />
  <body>
    <!-- Dynamic component loading -->
    <={layout} {...content.fields} />
  </body>
</html>
```

**Changes Needed:**
- Replace `<svelte:component this={layout}>` with `<={layout}>`
- Keep spread operator `{...content.fields}` for passing all fields
- Auto-inject magic variables (no need to export them)

### 3.3 Magic Variables System

**What Plenti Currently Injects (JavaScript):**
```javascript
var props = {
  content: { /* current page */ },
  layout: layouts_content_blog_svelte,  // Component reference
  allContent: [ /* all pages */ ],
  shadowContent: {},
  env: { /* environment */ }
};
```

**What Custom Go Engine Should Receive:**
```go
// In Go, before rendering
magicVars := map[string]interface{}{
    "content":       currentContent,        // Current page struct
    "allContent":    allContentArray,      // All pages slice
    "allLayouts":    componentSignatures,  // Map of component names
    "env":           envConfig,            // Environment config
    "params":        urlParams,            // URL query params (client-side)
}

// Flatten content.fields for easier access (optional)
for key, value := range currentContent.Fields {
    magicVars[key] = value  // Now {title} works instead of {content.fields.title}
}
```

---

## 4. Implementation Changes Required

### 4.1 New File: `cmd/build/render_templates.go`

Replace Svelte compilation with Go template rendering:

```go
package build

import (
    "github.com/jimafisk/custom_go_template/renderer"
    "github.com/jimafisk/custom_go_template/parser"
)

// RenderTemplate replaces compileSvelte()
func RenderTemplate(layoutPath string, currentContent content, allContentStr string, env env) (string, string, string, error) {
    
    // Build magic variables map
    magicVars := map[string]interface{}{
        "content":    createContentObject(currentContent),
        "allContent": allContentStr, // Already JSON string
        "env":        createEnvObject(env),
        "layout":     getLayoutSignature(currentContent.contentType),
    }
    
    // Render with custom Go template engine
    markup, script, style := renderer.Render(layoutPath, magicVars)
    
    return markup, script, style, nil
}

func createContentObject(c content) map[string]interface{} {
    return map[string]interface{}{
        "pager":    c.contentPagerNums,
        "type":     c.contentType,
        "path":     c.contentPath,
        "filepath": c.contentFilepath,
        "filename": c.contentFilename,
        "fields":   parseJSONFields(c.contentFields),
    }
}

func createEnvObject(e env) map[string]interface{} {
    return map[string]interface{}{
        "local":          e.local,
        "baseurl":        e.baseurl,
        "routes":         e.routes,
        "fingerprint":    e.fingerprint,
        "entrypointHTML": e.entrypointHTML,
        "entrypointJS":   e.entrypointJS,
        "sitevars":       e.sitevars,
        "cms":            e.cms,
    }
}

func getLayoutSignature(contentType string) string {
    // Generate component signature: layouts_content_blog_html
    return "layouts_content_" + contentType + "_html"
}
```

### 4.2 Modify: `cmd/build/data_source.go`

Replace `createProps()` and `createHTML()` functions:

```go
// REMOVE: createProps() - no longer needed (no V8 SSR)
// REMOVE: SSRctx references - no JavaScript engine needed

// MODIFY: createHTML()
func createHTML(currentContent content, allContentStr string, env env) error {
    
    // Determine layout path based on content type
    layoutPath := fmt.Sprintf("layouts/content/%s.html", currentContent.contentType)
    
    // Render template with Go engine
    markup, script, style, err := RenderTemplate(layoutPath, currentContent, allContentStr, env)
    if err != nil {
        return fmt.Errorf("failed to render template: %w", err)
    }
    
    // Add DOCTYPE if not present
    if !strings.Contains(markup, "<!DOCTYPE") {
        markup = "<!DOCTYPE html>" + markup
    }
    
    // Add data-content-filepath attribute
    markup = strings.Replace(markup, "<html", 
        fmt.Sprintf("<html data-content-filepath='%s' ", currentContent.contentFilepath), 1)
    
    // Inject live-reload script if in dev mode
    if Doreload {
        markup = strings.Replace(markup, "</body>", 
            fmt.Sprintf("<script src='%s%s/core/live-reload.js'></script></body>", 
                env.baseurl, env.entrypointJS), 1)
    }
    
    // Write CSS to bundle
    if style != "" {
        if err := appendToFile("public/bundle.css", style); err != nil {
            return err
        }
    }
    
    // Write JS to bundle (if any Alpine.js data/functions)
    if script != "" {
        if err := appendToFile("public/bundle.js", script); err != nil {
            return err
        }
    }
    
    // Create output directory
    destDir := filepath.Dir(currentContent.contentDest)
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return err
    }
    
    // Write final HTML
    return ioutil.WriteFile(currentContent.contentDest, []byte(markup), 0644)
}
```

### 4.3 Remove: Svelte-specific Files

**Delete or Skip:**
- `cmd/build/compile.go` - Svelte compilation
- `cmd/build/bundle.go` - JavaScript bundling (or simplify for minimal Alpine.js)
- `cmd/build/gopack.go` - Go bundler (if only for Svelte)
- `defaults/compiler/` - Svelte compiler
- `defaults/core/router.svelte` - Client-side routing (or replace with minimal version)

**Keep:**
- `cmd/build/data_source.go` - Content discovery ✅
- `cmd/build/media_copy.go` - Media file copying ✅
- `cmd/build/static_copy.go` - Static file copying ✅
- `cmd/build/themes_*.go` - Theme support ✅
- `defaults/core/live-reload.js` - Dev server ✅

### 4.4 Simplify: Client-Side Routing

**Option A: Static-Only (No SPA)**
- Remove all client-side routing
- Pure static HTML pages
- No JavaScript hydration needed
- Simplest approach

**Option B: Minimal Alpine.js Routing**
Create `defaults/core/router.html`:
```html
<script>
  // Minimal routing for Alpine.js data reactivity
  // Only needed if you want client-side interactivity
  import Alpine from 'alpinejs';
  
  Alpine.start();
</script>
```

**Recommendation:** Start with Option A (static-only), add Option B later if needed.

---

## 5. Fence Section Changes

### 5.1 Current Fence Section (Svelte)
```javascript
---
export let title, author;
let greeting = "Hello";
---
```

### 5.2 New Fence Section (Go Template)
```javascript
---
// NO prop declarations (data from content.fields)
// YES to imports
import Header from "../components/header.html";

// YES to helper functions (for template logic)
function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString();
}

// YES to local variables (if needed for logic)
let formattedDate = formatDate(content.fields.date);
---
```

**Rules:**
1. ❌ **No `export let` or `prop` declarations** - all data from magic variables
2. ✅ **Keep imports** - component imports work the same
3. ✅ **Keep functions** - helper functions for data transformation
4. ✅ **Keep local variables** - computed values before rendering
5. ✅ **Access magic variables** - `content`, `allContent`, `env` available

### 5.3 Fence Parser Changes

**Modify:** `parser/parser.go` in custom Go template project

```go
// When parsing fence section:
// - Keep: ImportStatements
// - Keep: Functions
// - Keep: Variables (local only)
// - Remove: Prop declarations (or ignore them)

type FenceSection struct {
    Imports    []ImportStatement  // KEEP
    Functions  []FunctionDef      // KEEP
    Variables  []Variable         // KEEP (for local vars)
    Props      []PropDeclaration  // REMOVE or IGNORE
    RawContent string            // KEEP
}
```

**Instead of extracting props**, the renderer should:
1. Parse fence for imports and functions
2. Inject magic variables from Plenti's data
3. Make `content.fields.*` available for data access

---

## 6. Component System Integration

### 6.1 Component Registration

**Current (Plenti):**
```go
// Plenti doesn't pre-register components
// They're dynamically imported at runtime via Svelte
```

**New (Custom Go Template):**
```go
// Pre-register all components at build time
func registerAllComponents() error {
    componentsDir := "layouts/components"
    return filepath.Walk(componentsDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && strings.HasSuffix(path, ".html") {
            return registerComponent(path)
        }
        return nil
    })
}

func registerComponent(path string) error {
    // Use existing transformer.RegisterComponent()
    content, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }
    
    ast, err := parser.ParseTemplate(string(content))
    if err != nil {
        return err
    }
    
    name := componentNameFromPath(path)
    transformer.RegisterComponent(name, ast, extractProps(ast))
    
    return nil
}
```

### 6.2 Component Signatures

**Keep the same signature system:**
```
layouts/components/header.html → layouts_components_header_html
layouts/content/blog.html      → layouts_content_blog_html
```

**Update file extension matching:**
```go
// Change from .svelte to .html
componentSignature := strings.ReplaceAll(
    strings.ReplaceAll(layoutPath, "/", "_"),
    ".", "_")
```

---

## 7. allLayouts Magic Variable

### 7.1 Current Implementation (Svelte)
Generated during Svelte compilation in `compile.go`:
```javascript
// Generated JavaScript
export let allLayouts = {
  layouts_components_header_svelte: Header,
  layouts_components_footer_svelte: Footer,
  layouts_content_blog_svelte: Blog
};
```

### 7.2 New Implementation (Go Template)

**Build Time (Go):**
```go
func generateAllLayoutsMap() map[string]string {
    layouts := make(map[string]string)
    
    // Walk layouts directory
    filepath.Walk("layouts", func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && strings.HasSuffix(path, ".html") {
            signature := makeSignature(path)
            layouts[signature] = path
        }
        return nil
    })
    
    return layouts
}

func makeSignature(path string) string {
    sig := strings.ReplaceAll(path, "/", "_")
    sig = strings.ReplaceAll(sig, ".", "_")
    return sig
}
```

**Runtime (Template):**
```html
<!-- Access in template -->
{for comp of content.fields.components}
  <={allLayouts["layouts_components_" + comp.name + "_html"]} {...comp} />
{/for}
```

**In Renderer:**
```go
magicVars := map[string]interface{}{
    "allLayouts": generateAllLayoutsMap(),
    // ... other vars
}
```

---

## 8. Build Command Integration

### 8.1 Modify: `cmd/build.go`

```go
func Build(buildPath string, spaPath string, siteConfig readers.SiteConfig) error {
    
    // Step 1: Register all components (NEW)
    if err := registerAllComponents(); err != nil {
        return err
    }
    
    // Step 2: Process content (EXISTING - keep as-is)
    if err := DataSource(buildPath, spaPath, siteConfig); err != nil {
        return err
    }
    
    // Step 3: Copy static files (EXISTING - keep as-is)
    if err := StaticCopy(buildPath); err != nil {
        return err
    }
    
    // Step 4: Copy media files (EXISTING - keep as-is)
    if err := MediaCopy(buildPath); err != nil {
        return err
    }
    
    // Step 5: REMOVE/SKIP Svelte compilation
    // SKIP: if err := Bundle(buildPath, siteConfig); err != nil {
    
    // Step 6: Minify output (EXISTING - keep as-is)
    if err := Minify(buildPath); err != nil {
        return err
    }
    
    return nil
}
```

### 8.2 Update: `cmd/serve.go`

```go
// Keep live-reload functionality
// Simplify or remove client-side SPA routing
// Watch for changes in:
// - content/*.json
// - layouts/**/*.html
// - media/**/*
// - static/**/*

func Serve(port int, buildPath string) error {
    // Existing file watcher (keep)
    // Existing server (keep)
    // REMOVE: JavaScript bundling watchers
    // ADD: Watch for .html template changes
}
```

---

## 9. Testing Strategy

### 9.1 Unit Tests
```go
// Test magic variable injection
func TestMagicVariableInjection(t *testing.T)

// Test component registration
func TestComponentRegistration(t *testing.T)

// Test template rendering with Plenti data
func TestRenderWithPlentiData(t *testing.T)

// Test allLayouts generation
func TestAllLayoutsMap(t *testing.T)
```

### 9.2 Integration Tests
Create test Plenti site:
```
test-site/
├── content/
│   ├── _index.json
│   └── blog/
│       └── test.json
├── layouts/
│   ├── global/
│   │   └── html.html
│   ├── content/
│   │   ├── _index.html
│   │   └── blog.html
│   └── components/
│       └── header.html
└── plenti.json
```

**Test Cases:**
1. ✅ Build site successfully
2. ✅ Generate correct HTML output
3. ✅ Magic variables accessible in templates
4. ✅ Components render correctly
5. ✅ Custom routes work
6. ✅ Pagination works
7. ✅ Styles extracted correctly
8. ✅ Live reload works in dev mode

---

## 10. Migration Path

### Phase 1: Setup Infrastructure
**Duration:** 2-3 hours

1. Create `cmd/build/render_templates.go`
2. Modify `cmd/build/data_source.go` to use new renderer
3. Add component registration to build process
4. Update imports in all build files

### Phase 2: Template Conversion
**Duration:** 1-2 hours

1. Convert `layouts/global/html.svelte` → `.html`
2. Convert one content template (e.g., `_index.svelte` → `.html`)
3. Test basic rendering with magic variables
4. Verify component imports work

### Phase 3: Remove Svelte Dependencies
**Duration:** 1 hour

1. Comment out Svelte compilation in `build.go`
2. Remove or simplify JavaScript bundling
3. Test static HTML generation
4. Verify CSS extraction

### Phase 4: Full Template Migration
**Duration:** 2-3 hours

1. Convert all content templates
2. Convert all components
3. Test each page type
4. Verify all features work

### Phase 5: Client-Side Refinement
**Duration:** 1-2 hours

1. Decide on SPA vs static-only approach
2. Add minimal Alpine.js if needed
3. Test interactivity
4. Test live reload

**Total Estimated Time:** 7-11 hours

---

## 11. Configuration Changes

### 11.1 plenti.json (Minimal Changes)
```json
{
  "entrypointHTML": "global/html.html",  // Changed from .svelte
  "entrypointJS": "spa",                 // Keep or simplify
  "routes": { /* same */ },
  "baseurl": "/",
  "fingerprint": "abc123",
  "cms": { /* same */ }
}
```

### 11.2 package.json (Remove Svelte)
```json
{
  "dependencies": {
    // REMOVE: "svelte": "^3.x"
    // KEEP: "alpinejs": "^3.x" (if using client-side)
  }
}
```

---

## 12. Key Differences Summary

| Aspect | Svelte (Current) | Go Template (New) |
|--------|------------------|-------------------|
| **Template Extension** | `.svelte` | `.html` |
| **Prop Declaration** | `export let title` | None (from `content.fields`) |
| **Data Access** | `{title}` | `{content.fields.title}` |
| **Component Import** | `import Header` | `import Header` (same) |
| **Magic Variables** | Exported props | Auto-injected |
| **Compilation** | Svelte → JS (V8) | Direct Go rendering |
| **Client-Side** | Full SPA | Static or minimal |
| **Build Time** | Slower (JS compilation) | Faster (native Go) |
| **Output Size** | Larger (Svelte runtime) | Smaller (no runtime) |

---

## 13. Benefits of This Integration

### 13.1 Performance
- ✅ **Faster builds** - No JavaScript compilation
- ✅ **Smaller output** - No Svelte runtime (~40KB saved)
- ✅ **Native Go** - Better concurrency for parallel builds

### 13.2 Simplicity
- ✅ **Single language** - All Go, no JavaScript toolchain
- ✅ **Fewer dependencies** - No npm packages for build
- ✅ **Easier debugging** - Native Go errors, no V8 runtime

### 13.3 Compatibility
- ✅ **Keep Plenti features** - Content, routing, CMS all work
- ✅ **Same data model** - JSON content unchanged
- ✅ **Familiar syntax** - Similar to Svelte templates
- ✅ **Progressive enhancement** - Can add Alpine.js later

---

## 14. Potential Challenges

### 14.1 Dynamic Components
**Challenge:** Svelte's `<svelte:component this={layout}>` is powerful

**Solution:** Use custom Go template dynamic component syntax:
```html
<!-- Instead of Svelte's dynamic component -->
<svelte:component this={layout} {...content.fields} />

<!-- Use Go template's dynamic syntax -->
<={layout} {...content.fields} />
```

**Implementation in renderer:**
```go
// Parse <={variableName} syntax
// Look up component from allLayouts map
// Render with provided props
```

### 14.2 Scoped CSS
**Challenge:** Svelte auto-scopes CSS to components

**Solution:** 
- Option A: Collect all CSS into single bundle (simple)
- Option B: Add component-specific class prefixes (more work)
- Option C: Use CSS-in-JS approach with Alpine.js (optional)

**Recommendation:** Start with Option A (global CSS bundle)

### 14.3 Reactivity
**Challenge:** Svelte has reactive statements (`$: computed = value * 2`)

**Solution:**
- Server-side: Compute in fence section before rendering
- Client-side: Use Alpine.js reactive data if needed

```html
---
// Compute server-side
let doubled = content.fields.count * 2;
---

<div x-data="{ count: {content.fields.count} }">
  <!-- Client-side reactivity via Alpine -->
  <span x-text="count * 2"></span>
</div>
```

---

## 15. Success Criteria

The integration is complete when:

1. ✅ **Build works** - `plenti build` generates correct HTML
2. ✅ **Serve works** - `plenti serve` runs dev server with live reload
3. ✅ **Content loads** - JSON data accessible in templates
4. ✅ **Routing works** - URLs map to correct pages
5. ✅ **Components work** - Imports and dynamic components render
6. ✅ **Magic variables work** - `content`, `allContent`, `env` accessible
7. ✅ **Styles work** - CSS extracted and applied
8. ✅ **No Svelte** - All `.svelte` files replaced with `.html`
9. ✅ **Tests pass** - All integration tests green
10. ✅ **Docs updated** - Migration guide complete

---

## 16. Next Steps

1. **Review this spec** with team
2. **Create feature branch** for integration
3. **Start Phase 1** - Infrastructure setup
4. **Test incrementally** - One phase at a time
5. **Document changes** - Update Plenti docs
6. **Create migration guide** - For existing Plenti users

---

## 17. Questions to Resolve

1. **Client-side routing:** Keep SPA behavior or go static-only?
2. **Alpine.js:** Include by default or optional?
3. **CSS bundling:** Single file or per-component?
4. **Backward compatibility:** Support `.svelte` templates temporarily?
5. **Performance targets:** What build time is acceptable?

---

## End of Specification