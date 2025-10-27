# CRITICAL LESSON: Route Registration Must Not Be Changed

## What Happened

While trying to implement Plenti content type support for the news page, we made a **critical mistake** that broke the jim-test showcase page.

## The Mistake

**File:** `cmd/server/main.go` - `registerContentRoutes()` function (around line 1440)

**Changed FROM (CORRECT):**
```go
http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
    renderTemplate(currentFilePath, w, r)
})
```

**Changed TO (BROKEN):**
```go
http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
    if err := renderWithWrapper(currentRouteName, w, r); err != nil {
        log.Printf("Error rendering %s: %v", route, err)
        http.Error(w, "Failed to render page", http.StatusInternalServerError)
    }
})
```

## Impact

This change broke **ALL content routes** registered from `layouts/content/*.html`, including:
- `/jim-test` - Showcase page demonstrating multiple features
- `/store-demo` - Store demonstration page
- `/pages` - Pages layout
- Any other custom content layouts

## Why It Broke

### How Content Routes Work

Content routes registered via `registerContentRoutes()` are designed to render **standalone layout templates directly**:

1. Files in `layouts/content/*.html` are discovered
2. Routes are auto-registered (e.g., `jim-test.html` → `/jim-test`)
3. Each route calls `renderTemplate(filePath, w, r)` directly
4. The template is rendered as-is with its own structure

### What renderTemplate Does

```go
renderTemplate("layouts/content/jim-test.html", w, r)
```

- Loads the template file
- Parses it to AST
- Loads content JSON for the route (if exists)
- Transforms AST to Alpine.js
- Renders the template directly
- Returns the rendered HTML

The jim-test template is a **complete HTML page** with:
- Its own `<!DOCTYPE html>`
- Its own `<head>` and `<body>`
- Its own x-data on the body
- All its content components

### What renderWithWrapper Does

```go
renderWithWrapper("jim-test", w, r)
```

- Loads the **global HTML wrapper** (`layouts/global/html.html`)
- Loads content JSON for the route
- Passes the layout name ("jim-test") as a prop
- Expects the wrapper to render the layout as a **component**
- The wrapper uses `<Component:dynamic name={layout} />`

### The Problem

When we changed content routes to use `renderWithWrapper`:

1. The global HTML wrapper was loaded
2. The jim-test layout was supposed to be rendered as a component
3. But jim-test is NOT a component - it's a **full HTML page**
4. The wrapper tried to inject it as a component inside its own HTML structure
5. This caused conflicts, missing content, and broken rendering

## The Root Cause

We were trying to make the **news page** work by forcing ALL content routes through the wrapper. But:

- **Collection-type pages** (like jim-test with `components:[]`) are designed to render directly
- **Single content type pages** (like news with `fields:{}`) need different handling
- Changing the route registration affected BOTH types, breaking working pages

## What We Should Have Done

Instead of changing the route registration globally, we should have:

1. **Created a separate route specifically for /news** that uses the wrapper
2. **Updated the news layout** to work with direct rendering (access via `content.fields`)
3. **Created a new route handler** specifically for Plenti-style single content types

## The Correct Architecture

### Direct Template Rendering (Current System)
```
Route: /jim-test
  ↓
registerContentRoutes() → renderTemplate()
  ↓
Load: layouts/content/jim-test.html
  ↓
Render as complete HTML page
  ↓
Output: Full HTML with <!DOCTYPE>, <html>, <body>, etc.
```

### Wrapper-Based Rendering (For special pages)
```
Route: / (home)
  ↓
Specific route handler → renderWithWrapper("Pages")
  ↓
Load: layouts/global/html.html (wrapper)
  ↓
Wrapper renders: <Component:dynamic name="Pages" />
  ↓
Output: HTML with wrapper's structure + Pages layout injected
```

## Lessons Learned

### 1. Never Change Core Route Registration Without Full Testing

The `registerContentRoutes()` function is **core infrastructure** that affects:
- All auto-registered content layouts
- All showcase pages
- All custom content types
- Potentially dozens of routes

Changing it requires testing EVERY registered route.

### 2. Test Existing Pages Before Adding New Features

Before implementing news page support, we should have:
- ✅ Verified jim-test still works
- ✅ Verified store-demo still works
- ✅ Verified all existing routes work
- ❌ We skipped this and broke working functionality

### 3. Understand the Difference Between Page Types

**Standalone Pages (jim-test):**
- Complete HTML documents
- Rendered directly via `renderTemplate()`
- Have their own structure and x-data
- Should NOT be wrapped

**Component-Based Pages (home):**
- Use global HTML wrapper
- Rendered via `renderWithWrapper()`
- Injected as components into wrapper
- Designed to work with wrapper's structure

### 4. New Features Should Not Break Existing Functionality

The news page is a **new feature**. It should:
- ✅ Be additive (add new capability)
- ❌ NOT modify existing core systems
- ✅ Have its own route handler if needed
- ❌ NOT change behavior of existing pages

### 5. Document Architectural Decisions

We should have documented:
- Why content routes use `renderTemplate()`
- Why the home route uses `renderWithWrapper()`
- What the difference is between these approaches
- When to use each approach

## Prevention Checklist

Before making changes to route registration:

- [ ] Identify all routes affected by the change
- [ ] Test each affected route before and after
- [ ] Understand why the current approach exists
- [ ] Consider if a new route handler is better than modifying existing ones
- [ ] Document the reason for the change
- [ ] Have a rollback plan

## Rollback Steps (What We Did)

To fix the broken jim-test page, we reverted:

1. ✅ Route registration in `registerContentRoutes()` - Changed back to `renderTemplate()`
2. ✅ Field extraction in `renderWithWrapper()` - Reverted single content type handling
3. ✅ Field spreading in `renderWithWrapper()` - Removed the spreading loop
4. ✅ Content injection in `renderTemplate()` - Reverted to original logic

After reverting all changes, jim-test page works again.

## Going Forward

### For the News Page

The news page needs a **different approach**:

**Option A: Dedicated Route Handler**
```go
http.HandleFunc("/news", func(w http.ResponseWriter, r *http.Request) {
    if err := renderWithWrapper("news", w, r); err != nil {
        http.Error(w, "Failed to render page", http.StatusInternalServerError)
    }
})
```

**Option B: Update News Layout**
Make news.html access fields via `content.fields.title` instead of `export let title`.

**Option C: Convert to Collection Type**
Change news JSON to use `components:[]` array like home page.

### General Principle

**When adding new page types:**
1. Create dedicated route handlers
2. Don't modify `registerContentRoutes()`
3. Test existing pages after every change
4. Revert immediately if anything breaks

## Summary

- ❌ **Mistake:** Changed `registerContentRoutes()` to use `renderWithWrapper()`
- 💥 **Impact:** Broke jim-test and all auto-registered content routes
- ✅ **Fix:** Reverted route registration back to `renderTemplate()`
- 📝 **Lesson:** Never change core infrastructure without full testing
- 🎯 **Rule:** New features should be additive, not destructive

This was a critical mistake that we must never repeat. The route registration system is core infrastructure that affects all content pages.
