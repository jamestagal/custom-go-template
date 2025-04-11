# Custom Go Template Engine Documentation

This document provides comprehensive documentation for the custom Go template engine, focusing on the template syntax and transformation rules for Alpine.js integration.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Syntax](#basic-syntax)
3. [Expressions](#expressions)
4. [Conditionals](#conditionals)
5. [Loops](#loops)
6. [Components](#components)
7. [Alpine.js Integration](#alpine-js-integration)
8. [Transformation Rules](#transformation-rules)
9. [Examples](#examples)

## Introduction

The custom Go template engine provides a powerful way to create dynamic HTML templates that integrate seamlessly with Alpine.js. The engine transforms template syntax into Alpine.js compatible HTML, allowing for reactive components without writing Alpine.js code directly.

## Basic Syntax

### Text Content

Plain text in templates is rendered as-is:

```html
<div>Hello, world!</div>
```

### HTML Elements

Standard HTML elements work as expected:

```html
<div class="container">
  <h1>Title</h1>
  <p>Paragraph content</p>
</div>
```

## Expressions

Expressions allow you to output dynamic content using Go template syntax.

### Basic Expression

```html
<div>{{ variable }}</div>
```

This will be transformed to:

```html
<div><span x-text="variable"></span></div>
```

### Inline Expressions

```html
<div>Hello, {{ name }}!</div>
```

This will be transformed to:

```html
<div>Hello, <span x-text="name"></span>!</div>
```

### Attribute Expressions

```html
<div class="{{ dynamicClass }}">Content</div>
```

This will be transformed to:

```html
<div :class="dynamicClass">Content</div>
```

## Conditionals

Conditionals allow you to render content based on conditions.

### If Statement

```html
{{ if condition }}
  <div>Conditional content</div>
{{ end }}
```

This will be transformed to:

```html
<template x-if="condition">
  <div>Conditional content</div>
</template>
```

### If-Else Statement

```html
{{ if isActive }}
  <div>This is active</div>
{{ else }}
  <div>This is inactive</div>
{{ end }}
```

This will be transformed to:

```html
<template x-if="isActive">
  <div>This is active</div>
</template>
<template x-else>
  <div>This is inactive</div>
</template>
```

### If-Else If-Else Statement

```html
{{ if status === 'active' }}
  <div>This is active</div>
{{ else if status === 'pending' }}
  <div>This is pending</div>
{{ else }}
  <div>This is inactive</div>
{{ end }}
```

This will be transformed to:

```html
<template x-if="status === 'active'">
  <div>This is active</div>
</template>
<template x-else-if="status === 'pending'">
  <div>This is pending</div>
</template>
<template x-else>
  <div>This is inactive</div>
</template>
```

## Loops

Loops allow you to iterate over arrays and objects.

### Array Loop

```html
{{ for item in items }}
  <div>{{ item }}</div>
{{ end }}
```

This will be transformed to:

```html
<template x-for="item in items">
  <div><span x-text="item"></span></div>
</template>
```

### Array Loop with Index

```html
{{ for index, item in items }}
  <div>{{ index }}: {{ item }}</div>
{{ end }}
```

This will be transformed to:

```html
<template x-for="(index, item) in items">
  <div><span x-text="index"></span>: <span x-text="item"></span></div>
</template>
```

### Object Loop

```html
{{ for key, value of user }}
  <div>{{ key }}: {{ value }}</div>
{{ end }}
```

This will be transformed to:

```html
<template x-for="key, value of Object.entries(user)">
  <div><span x-text="key"></span>: <span x-text="value"></span></div>
</template>
```

## Components

Components allow you to create reusable template fragments.

### Component Definition

```html
{{ component Button }}
  <button class="{{ class }}">{{ label }}</button>
{{ end }}
```

### Component Usage

```html
{{ Button label="Click me" class="btn btn-primary" }}
```

This will be transformed to include the component content with the provided props.

## Alpine.js Integration

The template engine automatically integrates with Alpine.js by transforming template syntax into Alpine.js directives.

### Data Binding

```html
<input value="{{ inputValue }}">
```

This will be transformed to:

```html
<input :value="inputValue">
```

### Event Handling

```html
<button @click="{{ handleClick }}">Click me</button>
```

This will be transformed to:

```html
<button @click="handleClick">Click me</button>
```

## Transformation Rules

The template engine follows these transformation rules:

### Expression Transformation

1. Text expressions (`{{ variable }}`) are transformed to `<span x-text="variable"></span>`
2. Attribute expressions (`attribute="{{ value }}"`) are transformed to `:attribute="value"`

### Conditional Transformation

1. `if` conditions are transformed to `<template x-if="condition">`
2. `else if` conditions are transformed to `<template x-else-if="condition">`
3. `else` conditions are transformed to `<template x-else>`

### Loop Transformation

1. Array loops (`for item in items`) are transformed to `<template x-for="item in items">`
2. Array loops with index (`for index, item in items`) are transformed to `<template x-for="(index, item) in items">`
3. Object loops (`for key, value of object`) are transformed to `<template x-for="key, value of Object.entries(object)">`

### Component Transformation

1. Components are tracked to prevent duplication
2. Component props are passed to the component template
3. Components are rendered once and referenced in the output

## Examples

### Complete Example: User Profile

```html
<div class="user-profile">
  {{ if user }}
    <h1>{{ user.name }}</h1>
    
    <div class="user-details">
      {{ for key, value of user.details }}
        <div class="detail">
          <strong>{{ key }}:</strong> {{ value }}
        </div>
      {{ end }}
    </div>
    
    {{ if user.isAdmin }}
      {{ AdminPanel }}
    {{ else }}
      {{ UserProfile }}
    {{ end }}
    
    <div class="user-posts">
      <h2>Recent Posts</h2>
      {{ if user.posts.length > 0 }}
        {{ for index, post in user.posts }}
          <div class="post">
            <h3>{{ post.title }}</h3>
            <p>{{ post.content }}</p>
            {{ if post.comments.length > 0 }}
              <div class="comments">
                <h4>Comments ({{ post.comments.length }})</h4>
                {{ for comment in post.comments }}
                  <div class="comment">{{ comment.text }}</div>
                {{ end }}
              </div>
            {{ end }}
          </div>
        {{ end }}
      {{ else }}
        <p>No posts found.</p>
      {{ end }}
    </div>
  {{ else }}
    <p>User not found.</p>
  {{ end }}
</div>
```

This example demonstrates the use of conditionals, loops, expressions, and nested structures in a real-world template.
