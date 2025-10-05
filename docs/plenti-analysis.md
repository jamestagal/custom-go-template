# Comprehensive Analysis of Plenti Build-Time Rendering Engine

## Executive Summary

Plenti is an open-source Build-Time Render (BTR) Engine that uniquely combines a **Go-based CLI**, **Svelte templating**, and a **Git-backed, browser-based CMS** into a unified static site generation platform. Its distinguishing feature is the "Discoverable CMS" - a zero-configuration content management system that automatically analyzes JSON data sources to generate appropriate editing interfaces.

## Core Architecture

### Technology Stack
- **Backend/CLI**: Go (for build performance)
- **Frontend Templates**: Svelte (reactive component framework)
- **Data Format**: JSON (no databases required)
- **Version Control**: Git (for content management)
- **Output**: Static HTML + hydrated SPA

### Key Architectural Decisions

1. **Build-time vs Runtime**: Most processing happens at build time, not in the browser
2. **No Bundlers**: Eliminates slow traditional bundling processes
3. **No External Database**: All content lives as JSON files in the repository
4. **Compiled Templates**: Svelte compiles to vanilla JavaScript (small runtime)

---

## Project Structure Deep Dive

### 1. Content Directory (`content/`)

**Purpose**: Single source of truth for all site data

**Structure**:
```
content/
├── index.json                    # Homepage (single file type)
├── 404.json                      # Error page (single file type)
├── blog/                         # Multi-file type
│   ├── _defaults.json           # Template for new posts
│   ├── _schema.json             # Widget overrides
│   ├── post1.json
│   └── post2.json
├── pages/
│   ├── _defaults.json
│   └── about.json
└── _components/                  # Global component definitions
    ├── hero/
    │   └── _defaults.json
    └── grid/
        └── _defaults.json
```

**Key Concepts**:

- **Types**: Organizational grouping (folder = type)
- **Flexibility**: No required fields, variable schemas within types
- **Special Files**:
  - `_defaults.json`: Scaffold for new content (enables "Add" in CMS)
  - `_schema.json`: Override default widget types
  - Files prefixed with `_schema_`: Override widgets for single types (e.g., `_schema_index.json`)

**Content Object Structure**:
```javascript
// Available on every content node
{
  pager: 1,                    // Pagination tracking
  type: "blog",                // Content type
  path: "/blog/post1",         // URL endpoint
  filepath: "content/blog/post1.json",
  filename: "post1",
  fields: {                    // User-defined data
    title: "My Post",
    author: "Jane",
    // ... any custom fields
  }
}
```

---

### 2. Layouts Directory (`layouts/`)

**Purpose**: Svelte templates that render content

**Structure**:
```
layouts/
├── global/
│   └── html.svelte           # CRITICAL: Main entry point
├── content/                  # Type-specific templates
│   ├── blog.svelte          # Matches content/blog/
│   ├── pages.svelte         # Matches content/pages/
│   └── index.svelte         # Matches content/index.json
├── components/              # Reusable UI elements
│   ├── grid.svelte
│   └── pager.svelte
└── scripts/                 # Helper functions
    └── utilities.js
```

**Critical File**: `layouts/global/html.svelte`
- Changing its name breaks the app
- Contains routing logic via `<svelte:component>`
- Common pattern:
```svelte
<script>
  export let content, user, allContent, allLayouts, env;
</script>

<html lang="en">
  <head>
    <base href="{env.baseurl}">
  </head>
  <body>
    {#if user && $user.isAuthenticated}
      <svelte:component this={$user.menu} bind:content {user} />
    {/if}
    <main>
      <!-- Dynamic routing: content.type determines which template loads -->
      <svelte:component this={layout} {...content.fields} {allContent} />
    </main>
  </body>
</html>
```

**Template Mapping Rule**: 
- `content/blog/post.json` → `layouts/content/blog.svelte`
- One template per type, feeds many content files

---

### 3. Media Directory (`media/`)

**Purpose**: Editor-manageable assets

**Behavior**:
- CMS-integrated (media browser)
- Folder structure = filtering mechanism
- Only folders with files appear in browser
- Reference in JSON with full path: `"image": "/media/people/photo.jpg"`

**Best Practices**:
- Store editor-changeable assets here
- Hardcoded theme assets go in `static/`

---

### 4. Static Directory (`static/`)

**Purpose**: Immutable theme files

**Common Contents**:
- Global CSS
- Logos, favicons
- `robots.txt`
- Hosting configs (e.g., `.domains`)

**Access**: Direct from domain root
```html
<!-- static/global.css accessed as: -->
<link rel="stylesheet" href="/global.css">
```

---

### 5. Themes Directory (`themes/`)

**Purpose**: Inherit from other Plenti projects

**Capability**: Any Plenti site can be a theme

**Configuration** (`plenti.json`):
```json
{
  "theme": "my-theme",
  "theme_config": {
    "my-theme": {
      "url": "git@github.com:user/repo",
      "commit": "abc123",
      "exclude": ["content", "media"]  // Optional exclusions
    }
  }
}
```

---

## The Discoverable CMS Concept

### Philosophy
**Zero-configuration content management** - the system automatically generates appropriate editing interfaces by analyzing the JSON data structure.

### How It Works

**1. Automatic Widget Detection**

The CMS inspects field values to determine appropriate input widgets:

| Data Pattern | Widget Type | Example Values |
|--------------|-------------|----------------|
| Date formats | Date picker | `6/7/2008`, `6-7-2008`, `Saturday, June 7, 2008` |
| Time formats | Time picker | `14:30`, `02:30pm`, `2:30 PM` |
| Numbers | Number input | `1`, `-1`, `1.5` |
| Booleans | Checkbox | `true`, `false` |
| Media paths | Media browser | `/media/image.jpg` |
| Arrays | Component list | `[]` |
| Text | Text input | Anything else |

**2. Schema Overrides** (`_schema.json`)

Customize default behavior:
```json
{
  "author": {
    "type": "reference",
    "options": [{
      "type": "docs",
      "search": ["group"],
      "result": "fields.group"
    }]
  },
  "body": {
    "type": "wysiwyg",
    "options": ["all"]
  }
}
```

**Available Widget Types**:
- `text`, `number`, `boolean`
- `date`, `time`
- `wysiwyg` (rich text)
- `select`, `checkbox`, `radio`
- `reference` / `references`
- `media`
- `component`
- `list_text`

**3. Component Architecture**

Special folder for component definitions:
```
content/
└── _components/
    ├── hero/
    │   └── _defaults.json
    └── grid/
        └── _defaults.json
```

In parent content schema:
```json
{
  "page_components": {
    "type": "component",
    "options": ["hero", "grid", "slider"]
  }
}
```

Result: Editors can add/remove/reorder components via CMS without touching code.

---

## Magic Variables System

Auto-injected props available in templates:

### `allContent`
```svelte
<script>
  export let allContent;
  
  // Filter by type
  $: blogPosts = allContent.filter(c => c.type === "blog");
</script>

{#each blogPosts as post}
  <a href={post.path}>{post.fields.title}</a>
{/each}
```

### `allLayouts`
Dynamic component loading without imports:

**Component Signatures**: Path with `/` and `.` → `_`
- `layouts/components/grid.svelte` → `layouts_components_grid_svelte`

```svelte
<script>
  export let components, allLayouts;
</script>

{#each components as {name}}
  <svelte:component 
    this={allLayouts[`layouts_components_${name}_svelte`]} 
  />
{/each}
```

### `content`
Current page data with standardized structure

### `params`
URL query string as `URLSearchParams` object (client-side only)

### `user`
CMS authentication state and methods:
- `$user.isAuthenticated`
- `$user.login()`
- `$user.logout()`
- `$user.menu` (admin interface component)

### `env`
Environment variables (e.g., `env.baseurl`)

---

## Routing & Path Generation

### Default Behavior
File structure = URL structure:
```
content/index.json          → /
content/blog/post1.json     → /blog/post1
content/events/My_Event.json → /events/my-event
```

### Custom Routes (`plenti.json`)
```json
{
  "routes": {
    "pages": "/:filename",           // Top-level pages
    "blog": "/blog/:fields(author)/:fields(title)",
    "index": ":paginate(totalPages)" // Pagination
  }
}
```

### Disable Endpoints
Delete corresponding template in `layouts/content/` or:
```bash
plenti new type YOUR_TYPE --endpoint=false
```

---

## Pagination System

### Configuration
```json
{
  "routes": {
    "index": ":paginate(totalPages)"
  }
}
```

### Implementation Pattern
```svelte
<script>
  export let content, allContent;
  
  $: currentPage = content.pager || 1;
  const postsPerPage = 3;
  
  $: allPosts = allContent.filter(c => c.type === "blog");
  $: totalPages = Math.ceil(allPosts.length / postsPerPage);
  $: postRangeHigh = currentPage * postsPerPage;
  $: postRangeLow = postRangeHigh - postsPerPage;
</script>

<!-- Display component -->
<Grid {allPosts} {postRangeLow} {postRangeHigh} />

<!-- Navigation component -->
<Pager {currentPage} {totalPages} />
```

---

## CMS Integration

### Local Setup
```svelte
<!-- layouts/global/html.svelte -->
<script>
  export let content, user;
</script>

{#if user && $user.isAuthenticated}
  <svelte:component this={$user.menu} bind:content {user} />
{/if}

<!-- Login trigger -->
<button on:click|preventDefault={$user.login()}>Login</button>
```

### Remote/Deployed CMS
```json
{
  "cms": {
    "repo": "https://gitlab.com/org/repo",
    "redirect_url": "https://my-site.com",
    "app_id": "oauth_id_here",
    "branch": "main"
  }
}
```

**Flow**:
1. Editors make changes in browser
2. CMS commits JSON updates to Git
3. CI/CD triggers build
4. Site rebuilds with new content

---

## Build & Deployment

### Local Development
```bash
plenti serve              # Start dev server (port 3000)
plenti serve -p 8080      # Custom port
plenti serve -L           # Live reload
plenti serve -M           # In-memory builds (faster)
```

### Production Build
```bash
plenti build              # Creates public/ folder
plenti build -o dist      # Custom output folder
plenti build -b           # Show benchmarks
```

### CI/CD Patterns

**GitHub Actions** (`.github/workflows/gh-pages.yml`):
```yaml
- name: Build
  uses: docker://plentico/plenti:latest
  with:
    entrypoint: /plenti
    args: build
```

**GitLab CI** (`.gitlab-ci.yml`):
```yaml
script:
  - wget [plenti release]
  - ./plenti build
```

**Version Locking** (recommended for production):
```bash
wget https://github.com/plentico/plenti/releases/download/v0.6.62/plenti_0.6.62_linux_64-bit.tar.gz
```

---

## Key Patterns for Go Templating Implementation

### 1. Type System
- **Directory = Type** paradigm
- Support for single-file types vs multi-file types
- Optional `_defaults.json` and `_schema.json` per type

### 2. Template Resolution
- One template per content type
- Dynamic component loading via signatures
- Spread operator pattern for field access

### 3. Data Access Patterns
```go
// Pseudocode structure
type Content struct {
    Pager    int
    Type     string
    Path     string
    Filepath string
    Filename string
    Fields   map[string]interface{}  // User-defined
}

// Magic variables to inject
type Context struct {
    Content    Content
    AllContent []Content
    AllLayouts map[string]Component
    Params     url.Values
    User       *AuthState
    Env        map[string]string
}
```

### 4. Component Signature Algorithm
```go
func ComponentSignature(path string) string {
    // "layouts/components/grid.svelte" 
    // → "layouts_components_grid_svelte"
    sig := strings.ReplaceAll(path, "/", "_")
    sig = strings.ReplaceAll(sig, ".", "_")
    return sig
}
```

### 5. Discoverable Widget Logic
```go
func DetectWidgetType(value interface{}) string {
    switch v := value.(type) {
    case bool:
        return "boolean"
    case float64, int:
        return "number"
    case string:
        if isDateFormat(v) { return "date" }
        if isTimeFormat(v) { return "time" }
        if isMediaPath(v) { return "media" }
        return "text"
    case []interface{}:
        return "component"
    default:
        return "text"
    }
}
```

### 6. Route Generation
```go
type RoutePattern struct {
    Type    string
    Pattern string  // e.g., "/:filename" or "/blog/:fields(author)"
}

func GeneratePath(content Content, pattern string) string {
    // Parse pattern and substitute variables
    // Handle :filename, :fields(key), etc.
}
```

---

## Differences from Traditional SSGs

| Aspect | Traditional SSGs | Plenti |
|--------|------------------|--------|
| **Data Format** | Markdown + YAML frontmatter | Pure JSON |
| **CMS** | Separate service (Netlify CMS, etc.) | Built-in, Git-backed |
| **Configuration** | Explicit setup required | Discoverable/zero-config |
| **Templates** | Go/Liquid/etc. | Svelte (compiled) |
| **CLI** | Often Node.js | Go (faster) |
| **Component System** | Import-based | Signature-based dynamic loading |

---

## Critical Implementation Notes

### Must-Have Features
1. **Automatic type detection** from `content/` folder structure
2. **Magic variable injection** (allContent, allLayouts, etc.)
3. **Component signature** generation and resolution
4. **Dynamic template** loading without explicit imports
5. **Route pattern** parsing and substitution
6. **Widget auto-detection** from field values

### Flexibility Points
1. **No required fields** in content JSON
2. **Variable schemas** within same type
3. **Nested component** support (unlimited depth)
4. **Custom route** patterns per type
5. **Selective theme** inheritance (can exclude folders)

### Performance Considerations
1. **Build-time processing** (not runtime)
2. **Static HTML** fallbacks for every page
3. **Minimal JavaScript** (Svelte's small runtime)
4. **Go's concurrency** for parallel builds

---

## Summary for LLM Agents

When building Go templating solutions for Plenti-like systems, remember:

1. **Content is king**: Everything derives from JSON structure in `content/`
2. **Convention over configuration**: Folder names, file names, and data patterns drive behavior
3. **Template mapping is 1:1**: Each content type folder maps to one layout template
4. **Magic happens at build time**: Analyze, generate, optimize - all before deployment
5. **CMS is emergent**: The editing interface emerges from analyzing the data structure
6. **Components are composable**: Signature-based loading enables content-driven UI
7. **No database needed**: Git + JSON = version-controlled, queryable data store

The core insight: **By making data structure itself the schema**, Plenti eliminates configuration overhead while maintaining flexibility. The Go implementation should preserve this discoverability while adding type safety and performance.