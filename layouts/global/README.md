# Global Layouts Directory

This directory contains the global wrapper components that appear on every page.

## File Structure

```
layouts/global/
├── html.html          # ⭐ Root wrapper (main application shell)
├── head.html          # <head> section (meta, title, CSS, JS)
├── nav.html           # Navigation menu
├── header.html        # Top branding/info (simple version)
└── footer.html        # Footer with links and social
```

## html.html - Root Application Shell

The `html.html` file is the **outermost template** that wraps every page on the site. It's equivalent to Plenti's `html.svelte`.

### Purpose & Role

- Defines the complete HTML document structure (`<!DOCTYPE html>` through `</html>`)
- Wraps all pages with global components (Head, Nav, Footer)
- Handles dynamic layout injection based on content type
- Manages authentication state (when implemented)

### Component Hierarchy

```
html.html (root)
├── Head (meta tags, title, CSS, Alpine.js CDN)
├── body
    ├── User Menu (conditional - if authenticated)
    ├── Login Modal (optional)
    └── main
        ├── Nav (global navigation)
        ├── Header (optional top branding)
        ├── [DYNAMIC CONTENT LAYOUT] ← Your page-specific content
        ├── Newsletter (optional subscription form)
        └── Footer (global footer)
```

### Props Received

```html
export let content, layout, allContent, allLayouts, env, user, shadowContent
```

**Injected by the Go server:**
- `content` - Current page's content data from JSON
- `layout` - The specific layout component name for this content type
- `allContent` - All content in the site (for navigation, listings, etc.)
- `allLayouts` - All compiled components (component registry)
- `env` - Environment variables
- `user` - Authentication state (if using auth system)
- `shadowContent` - Draft/preview content (for CMS editing)

### Dynamic Layout Injection

The key mechanism is the `Component:dynamic` syntax:

```html
<Component:dynamic name={layout}
    {...content.fields}
    {allContent}
    {allLayouts}
    {content}
    {env}
    bind:shadowContent={shadowContent} />
```

This dynamically renders the appropriate layout based on the content type:
- For pages → renders `layouts/content/pages.html`
- For courses → renders `layouts/content/courses.html`  
- For teachers → renders `layouts/content/teachers.html`
- etc.

The `{...content.fields}` spreads all content fields as props to the layout.

### Data Flow

```
Plenti/Go Build System
    ↓
html.html (receives all global props)
    ↓
Specific Layout (e.g., pages.html, courses.html)
    ↓
Components (e.g., hero.html, stats.html)
```

### Practical Example

When you visit `/course/modern-history`:

1. Go server loads `content/courses/modern-history.json`
2. Server determines layout: `layouts/content/courses.html`
3. `html.html` wraps everything:
   - Adds Head, Nav, Header
   - Injects `courses.html` via Component:dynamic
   - Adds Newsletter, Footer
4. `courses.html` renders the course-specific UI
5. User sees: **Nav → Course Content → Newsletter → Footer**

### Why This Matters

- **Single source of global structure** - Change header/footer here, affects all pages
- **Consistent wrapper** - Every page gets the same Nav, Footer, Newsletter
- **Layout switcher** - The `layout` prop determines which content template renders
- **Props propagation** - All layouts receive `allContent` and `allLayouts` for cross-referencing

### Implementation Status

**Current Status:** ✅ Created, ⏳ Needs server integration

**Next Steps:**
1. Update `cmd/server/main.go` to use `html.html` as the wrapper
2. Implement dynamic layout resolution based on content type
3. Pass required props (`content`, `layout`, `allContent`, etc.)
4. Test with existing pages

## Usage

This file is automatically applied by the server to every route. You don't need to manually include it in your content layouts.

### Server Integration Required

The server needs to:
1. Load `layouts/global/html.html` as the root wrapper
2. Determine the appropriate layout based on content type/route
3. Inject the layout component into the `Component:dynamic` slot
4. Pass all required props to the wrapper

See `cmd/server/main.go` for implementation details.
