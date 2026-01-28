# Plenti Content Type Mapping - Real-World Example

This document demonstrates the Plenti content type pattern with a real-world news/blog system example.

## Overview

In Plenti (and our engine), content types follow this pattern:

```
content/{type}/*.json → layouts/content/{type}.html
```

Each JSON file in a content type directory becomes a separate page, all using the same layout template.

## Example: News Content Type

### Directory Structure

```
content/
  news/
    quarterly-results.json    ← Becomes /news/quarterly-results
    new-product-launch.json   ← Becomes /news/new-product-launch
    team-expansion.json       ← Becomes /news/team-expansion

layouts/
  content/
    news.html                 ← Layout used by ALL news pages
  components/
    featured_posts_sidebar.html  ← Sidebar using allContent
```

### Content File Example

**`content/news/quarterly-results.json`**

```json
{
  "type": "news",
  "layout": "news",
  "path": "/news/quarterly-results",
  "title": "Quarterly Results Announced",
  "published": true,
  "fields": {
    "title": "Quarterly Results Announced",
    "published": true,
    "author": {
      "name": "Jane Smith",
      "image": {
        "src": "https://picsum.photos/32/32?random=author1",
        "alt": "Jane Smith profile"
      }
    },
    "publish": {
      "date": "2025-01-15"
    },
    "blogImage": {
      "src": "https://picsum.photos/1200/600?random=4",
      "alt": "Quarterly results chart",
      "width": 1200,
      "height": 600
    },
    "figcaption": {
      "caption": "Q4 2024 results demonstrate strong growth",
      "attribution": {
        "text": "Company Analytics Team",
        "link": "#"
      }
    },
    "textItems": [
      {
        "order": 1,
        "title": "Strong Financial Performance",
        "paragraph": "We're proud to announce our quarterly results..."
      }
    ],
    "photos": [
      {
        "src": "https://picsum.photos/400/300?random=41",
        "alt": "Team celebrating results",
        "width": 400,
        "height": 300
      }
    ]
  }
}
```

### Layout Template

**`layouts/content/news.html`**

The layout receives fields as individual props via `export let`:

```html
---
<!-- Declare props that come from JSON content -->
export let title, author, publish, blogImage, figcaption, textItems, photos, published, allContent
---

{if published === true}
<div class="blog-container main-content-wrapper">
  <div class="main-content">
    <article class="blog-article">
      <!-- Hero Image -->
      <picture class="blog-mainImage">
        <img
          alt="{blogImage?.alt || 'Blog image'}"
          src="{blogImage?.src}"
          width="{blogImage?.width || 1200}"
          height="{blogImage?.height || 600}"
        />
      </picture>

      <!-- Article Header -->
      <div class="article-group">
        <h1 class="blog-h1">{title}</h1>

        <div class="blog-authorGroup">
          {if author?.image}
          <picture class="blog-author-img">
            <img
              alt="{author.image.alt}"
              src="{author.image.src}"
              width="32"
              height="32"
            />
          </picture>
          {/if}
          <span class="blog-author">{author?.name || 'Author'}</span>
          <span class="blog-date">{publish?.date}</span>
        </div>
      </div>

      <!-- Article Content -->
      {if textItems}
        {for item in textItems}
        <section id="blog-content">
          <h4>{item.title}</h4>
          <p>{item.paragraph}</p>
        </section>
        {/for}
      {/if}

      <!-- Photo Gallery (if present) -->
      {if photos && photos.length > 0}
      <section id="photo-gallery">
        <div class="photo-grid">
          {for photo in photos}
          <div class="photo-item">
            <img
              src="{photo.src}"
              alt="{photo.alt || 'Gallery photo'}"
              width="{photo.width || 400}"
              height="{photo.height || 300}"
            />
          </div>
          {/for}
        </div>
      </section>
      {/if}
    </article>
  </div>

  <!-- Sidebar with Featured Posts (uses allContent) -->
  <FeaturedPostsSidebar allContent={allContent} />
</div>
{else}
<div class="draft-message">
  <h2>This post is currently a draft</h2>
  <p>Published posts will appear here.</p>
</div>
{/if}
```

### Sidebar Component Using allContent

**`layouts/components/featured_posts_sidebar.html`**

The sidebar requests `allContent` to display posts from OTHER pages:

```html
---
<!-- Request allContent magic variable -->
export let allContent
---

<div class="blog-sidebar">
  <div class="blog-featured-group">
    <span class="blog-header">Featured Posts</span>

    <!-- Filter posts by type and display featured ones -->
    {for post in Object.values(allContent).filter(c => c.type === 'news')}
      {if post.fields?.post?.id === '6' || post.fields?.post?.id === '5' || post.fields?.post?.id === '4'}
        <a href="{post.path}" class="blog-feature">
          <picture class="blog-featureImage">
            <img
              alt="{post.fields.blogImage?.alt || 'Blog image'}"
              src="{post.fields.blogImage?.src}"
              width="{post.fields.blogImage?.width || 400}"
              height="{post.fields.blogImage?.height || 300}"
            />
          </picture>

          <div class="content-group">
            <h3 class="feature-h3">{post.fields.title}</h3>
            <span class="feature-date">{post.fields.publish?.date}</span>
          </div>
        </a>
      {/if}
    {/for}
  </div>
</div>
```

## Key Patterns from Plenti

### 1. Content Type Mapping

- **Pattern**: `content/{type}/*.json` → `layouts/content/{type}.html`
- **Behavior**: Each JSON file becomes a page using the same layout
- **Example**: 3 JSON files in `content/news/` create 3 pages, all using `layouts/content/news.html`

### 2. Field Props via export let

- **Pattern**: `export let field1, field2, field3`
- **Source**: Fields come from `fields` object in JSON content
- **Behavior**: Each field is extracted and passed as individual prop to layout

### 3. Magic Variables

- **`content`**: The current page's full content object
- **`allContent`**: All content from all pages (opt-in via `export let`)
- **`allLayouts`**: All layout/component templates (for dynamic rendering)

### 4. Opt-In allContent

- **Trigger**: Component declares `export let allContent`
- **Load Behavior**: Server only loads allContent when explicitly requested
- **Use Case**: Sidebars, navigation, filtering/searching across content

### 5. Component Data Flow

```
JSON File
  ↓
Extract fields object
  ↓
Pass fields as individual props (export let)
  ↓
Layout renders with props
  ↓
Layout passes allContent to child components
  ↓
Child components filter/display related content
```

## Comparison: Plenti vs Our Engine

| Feature | Plenti (Svelte) | Our Engine |
|---------|----------------|------------|
| **Syntax** | `{#each}` / `{#if}` | `{for}` / `{if}` |
| **Props** | `export let prop` | `export let prop` ✅ Same |
| **Output** | Alpine.js directives | Alpine.js directives ✅ Same |
| **Content Loading** | Svelte component props | Fence section props ✅ Same |
| **Loops** | Runtime `x-for` | Build-time expansion + runtime fallback |
| **Components** | Svelte components | HTML components with fence |

## Real-World Example: capitaltigers Project

The example above is based on the **capitaltigers** Plenti project:

**File**: `/Users/benjaminwaller/Projects/PlentifyWebsites/capitaltigers/layouts/components/feature_blog_sidebar.svelte`

```svelte
<script>
  export let allContent;
  let posts = allContent.filter((content) => content.type === "news");
</script>

<div class="blog-sidebar">
  <div class="blog-featured-group">
    <span class="blog-header">Featured Posts</span>
    {#each posts as post}
    {#if post.fields.post.id === "6" || post.fields.post.id === "5" || post.fields.post.id === "4"}
        <a href={post.path} class="blog-feature">
          <picture class="blog-featureImage">
            <img
              alt={post.fields.blogImage.alt}
              src={post.fields.blogImage.src}
              width={post.fields.blogImage.width}
              height={post.fields.blogImage.height}
            />
          </picture>
          <div class="content-group">
            <h3 class="feature-h3">{post.fields.title}</h3>
            <span class="feature-date">{post.fields.publish.date}</span>
          </div>
        </a>
      {/if}
    {/each}
  </div>
</div>
```

Our implementation matches this pattern exactly, just with our syntax:
- `{#each posts as post}` → `{for post in posts}`
- `{#if condition}` → `{if condition}`
- Same `export let allContent` prop pattern
- Same filtering and display logic

## Benefits of This Pattern

### 1. **Content Separation**
- Content (JSON) separate from presentation (HTML)
- Non-developers can edit content files
- Developers can update layouts without touching content

### 2. **Reusability**
- One layout serves all pages of that type
- Changes to layout automatically apply to all pages
- DRY principle for content presentation

### 3. **Scalability**
- Add new pages by creating new JSON files
- No code changes needed for new content
- Automatic routing based on file structure

### 4. **SEO Benefits**
- Build-time loop expansion generates full HTML
- No client-side rendering required for content
- All text visible to search engines

### 5. **Component Ecosystem**
- Layouts can use components (sidebars, headers, etc.)
- Components can access allContent for cross-page features
- Modular, maintainable codebase

## Summary

The Plenti content type pattern provides:
- **Clear convention**: `content/{type}/*.json` → `layouts/content/{type}.html`
- **Simple data flow**: JSON fields → layout props → component props
- **Powerful querying**: `allContent` enables cross-page features
- **Developer experience**: Svelte-inspired syntax with Go backend performance

This example demonstrates real-world usage matching the **capitaltigers** Plenti project patterns.
