# LoadAllContent Specification

**Date:** 2026-01-28
**Status:** Discovery
**Priority:** P1 - Content Access Enhancement
**Depends On:** Plenti Integration API

---

## Problem Statement

Currently, accessing content in templates requires:
1. Manual filtering of `allContent` in loops
2. No built-in sorting or pagination
3. Verbose patterns for common operations

**Example of current verbose pattern:**
```html
{for post in allContent}
  {if post.type === "blog"}
    {if post.fields.featured === true}
      <BlogCard title={post.fields.title} />
    {/if}
  {/if}
{/for}
```

**Solution:** Provide a clean content loader API with filtering, sorting, and pagination.

---

## Proposed API

### Basic Usage

```html
---
import { loadContent } from 'plenti:content'

// Load all blog posts
const blogPosts = loadContent('blog')

// Load with sorting
const latestPosts = loadContent('blog', {
  sort: 'date',
  order: 'desc'
})

// Load with limit
const recentPosts = loadContent('blog', {
  sort: 'date',
  order: 'desc',
  limit: 5
})

// Load with filter
const featuredPosts = loadContent('blog', {
  where: { featured: true }
})
---

<h2>Latest Posts</h2>
{for post in latestPosts}
  <BlogCard {...post.fields} />
{/for}
```

### Single Content Access

```html
---
import { getContent } from 'plenti:content'

// Load specific content by path
const aboutPage = getContent('pages/about')

// Load by filepath
const homeContent = getContent('_index')
---

<h1>{aboutPage.fields.title}</h1>
```

### Full API

```html
---
import {
  loadContent,    // Load multiple entries by type
  getContent,     // Load single entry by path
  allContent,     // Raw allContent array
  contentTypes    // Available content types
} from 'plenti:content'
---
```

---

## LoadOptions Interface

```go
type LoadOptions struct {
    // Sorting
    Sort  string // Field to sort by (e.g., "date", "title")
    Order string // "asc" or "desc" (default: "asc")

    // Pagination
    Limit  int // Max entries to return
    Offset int // Skip first N entries

    // Filtering
    Where map[string]interface{} // Field conditions

    // Advanced
    Deep bool // Include nested fields in filtering
}
```

### Filter Examples

```html
---
// Simple equality
const published = loadContent('blog', {
  where: { status: 'published' }
})

// Multiple conditions (AND)
const featuredPublished = loadContent('blog', {
  where: {
    status: 'published',
    featured: true
  }
})

// Nested field access
const authorPosts = loadContent('blog', {
  where: { 'author.name': 'Jim' },
  deep: true
})
---
```

---

## Implementation

### Phase 1: Content Loader Type

```go
// plenti/content_loader.go

package plenti

// ContentLoader provides filtered access to content
type ContentLoader struct {
    allContent []ContentEntry
}

// NewContentLoader creates a loader with the full content array
func NewContentLoader(allContent []ContentEntry) *ContentLoader {
    return &ContentLoader{allContent: allContent}
}

// LoadContent returns filtered content entries
func (cl *ContentLoader) LoadContent(contentType string, opts ...LoadOptions) []ContentEntry {
    var options LoadOptions
    if len(opts) > 0 {
        options = opts[0]
    }

    // Filter by type
    var results []ContentEntry
    for _, entry := range cl.allContent {
        if entry.Type == contentType {
            if cl.matchesWhere(entry, options.Where) {
                results = append(results, entry)
            }
        }
    }

    // Sort
    if options.Sort != "" {
        results = cl.sortBy(results, options.Sort, options.Order)
    }

    // Pagination
    if options.Offset > 0 && options.Offset < len(results) {
        results = results[options.Offset:]
    }
    if options.Limit > 0 && options.Limit < len(results) {
        results = results[:options.Limit]
    }

    return results
}

// GetContent returns a single content entry by path
func (cl *ContentLoader) GetContent(path string) *ContentEntry {
    for _, entry := range cl.allContent {
        if entry.Path == path || entry.Filepath == "content/"+path+".json" {
            return &entry
        }
    }
    return nil
}

// GetContentTypes returns all unique content types
func (cl *ContentLoader) GetContentTypes() []string {
    types := make(map[string]bool)
    for _, entry := range cl.allContent {
        types[entry.Type] = true
    }

    var result []string
    for t := range types {
        result = append(result, t)
    }
    sort.Strings(result)
    return result
}

// matchesWhere checks if an entry matches the filter conditions
func (cl *ContentLoader) matchesWhere(entry ContentEntry, where map[string]interface{}) bool {
    if where == nil {
        return true
    }

    for key, expected := range where {
        actual, exists := entry.Fields[key]
        if !exists || actual != expected {
            return false
        }
    }
    return true
}

// sortBy sorts entries by a field
func (cl *ContentLoader) sortBy(entries []ContentEntry, field, order string) []ContentEntry {
    sort.Slice(entries, func(i, j int) bool {
        vi := entries[i].Fields[field]
        vj := entries[j].Fields[field]

        // Compare based on type
        less := compareValues(vi, vj)

        if order == "desc" {
            return !less
        }
        return less
    })
    return entries
}
```

### Phase 2: Fence Import Resolution

```go
// parser/fence.go - Add plenti:content import handling

func resolvePlentiImport(importPath string) map[string]interface{} {
    if importPath == "plenti:content" {
        return map[string]interface{}{
            "loadContent":  true, // Marker for content loader
            "getContent":   true,
            "allContent":   true,
            "contentTypes": true,
        }
    }
    return nil
}
```

### Phase 3: Transformer Integration

```go
// transformer/content_loader.go

func resolveContentLoaderCalls(fence *ast.FenceSection, allContent []ContentEntry) {
    loader := plenti.NewContentLoader(allContent)

    for _, varDecl := range fence.Variables {
        // Check if this is a loadContent call
        if call, ok := parseLoadContentCall(varDecl.Value); ok {
            contentType := call.Args[0]
            options := parseLoadOptions(call.Args[1:])

            // Resolve at build-time
            result := loader.LoadContent(contentType, options)

            // Replace variable value with resolved content
            varDecl.ResolvedValue = result
        }
    }
}
```

---

## Use Cases

### 1. Blog Listing Page

```html
---
import { loadContent } from 'plenti:content'

const posts = loadContent('blog', {
  sort: 'date',
  order: 'desc'
})
---

<h1>Blog</h1>
<div class="posts">
  {for post in posts}
    <article>
      <h2>{post.fields.title}</h2>
      <time>{post.fields.date}</time>
      <p>{post.fields.excerpt}</p>
      <a href="{post.path}">Read more</a>
    </article>
  {/for}
</div>
```

### 2. Navigation Menu

```html
---
import { loadContent } from 'plenti:content'

const pages = loadContent('pages', {
  where: { showInNav: true },
  sort: 'navOrder'
})
---

<nav>
  {for page in pages}
    <a href="/{page.path}">{page.fields.title}</a>
  {/for}
</nav>
```

### 3. Related Posts

```html
---
import { loadContent } from 'plenti:content'
export let category

const related = loadContent('blog', {
  where: { category: category },
  limit: 3,
  sort: 'date',
  order: 'desc'
})
---

<aside>
  <h3>Related Posts</h3>
  {for post in related}
    <a href="{post.path}">{post.fields.title}</a>
  {/for}
</aside>
```

### 4. Sitemap Generation

```html
---
import { allContent, contentTypes } from 'plenti:content'
---

<urlset>
  {for entry in allContent}
    <url>
      <loc>https://example.com/{entry.path}</loc>
    </url>
  {/for}
</urlset>
```

### 5. Footer with All Content Types

```html
---
import { loadContent, contentTypes } from 'plenti:content'
---

<footer>
  {for type in contentTypes}
    <div class="footer-section">
      <h4>{type}</h4>
      {for entry in loadContent(type, { limit: 5 })}
        <a href="/{entry.path}">{entry.fields.title}</a>
      {/for}
    </div>
  {/for}
</footer>
```

---

## Implementation Plan

### Phase 1: Content Loader (2-3 hours)

- [ ] Create `plenti/content_loader.go`
- [ ] Implement `LoadContent()` with filtering
- [ ] Implement `GetContent()` for single entry
- [ ] Implement sorting and pagination
- [ ] Write unit tests

### Phase 2: Import Resolution (2-3 hours)

- [ ] Parse `plenti:content` imports in fence
- [ ] Resolve `loadContent()` calls at build-time
- [ ] Inject resolved content into variable scope
- [ ] Write parser tests

### Phase 3: Integration (1-2 hours)

- [ ] Wire content loader into transformer
- [ ] Pass allContent to fence resolution
- [ ] Test with real content
- [ ] Integration tests

### Phase 4: Documentation (1 hour)

- [ ] Update CLAUDE.md with loadContent API
- [ ] Add examples for common patterns
- [ ] Document LoadOptions

---

## Success Criteria

1. **Clean API**: `loadContent('blog', { sort: 'date' })` works
2. **Build-time resolution**: Content resolved at build, not runtime
3. **Full filtering**: Where clauses filter correctly
4. **Sorting works**: Sort by any field, asc/desc
5. **Pagination works**: Limit and offset function correctly
6. **Backwards compatible**: `allContent` still works as before

---

## Estimated Effort

| Phase | Hours |
|-------|-------|
| Content Loader | 2-3 |
| Import Resolution | 2-3 |
| Integration | 1-2 |
| Documentation | 1 |
| **Total** | **6-9** |

---

## Dependencies

- ✅ Content system (complete)
- ✅ Fence parsing (complete)
- ⏳ Plenti Integration API (in progress)

---

## Future Enhancements

- **GraphQL-like queries**: `loadContent('blog', { fields: ['title', 'date'] })`
- **Full-text search**: `loadContent('blog', { search: 'keyword' })`
- **Relationships**: `loadContent('blog', { include: 'author' })`
- **Caching**: Cache resolved queries for faster rebuilds
