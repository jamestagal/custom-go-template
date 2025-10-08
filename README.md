# Custom Go Template Engine

A custom Go template engine that transforms Svelte-inspired template syntax into Alpine.js-compatible HTML. The engine parses custom template syntax, transforms it through an AST, and renders reactive HTML components with global state management.

## Features

- **Svelte-like Syntax**: Write templates with familiar `{if}`, `{for}`, and `{variable}` syntax
- **Alpine.js Integration**: Automatically transforms to Alpine.js directives
- **Component System**: Reusable components with prop passing
- **Global Store System**: Shared reactive state across components
- **Fence Sections**: Front matter for props, imports, and data initialization
- **Development Server**: Live preview at http://localhost:3000

## Quick Start

### Installation

```bash
go build ./...
```

### Run Development Server

```bash
go run cmd/server/main.go
# Visit http://localhost:3000
```

### Run Tests

```bash
# All tests
go test ./... -v

# Specific package
go test ./transformer -v

# Integration tests
go test ./tests/alpine -v
```

## Template Syntax

### Expressions

Transform dynamic values into Alpine.js reactive text:

```html
<p>Hello, {name}!</p>
<!-- Renders as: -->
<p>Hello, <span x-text="name"></span>!</p>
```

### Conditionals

```html
{if isLoggedIn}
  <p>Welcome back!</p>
{else if isGuest}
  <p>Hello, Guest</p>
{else}
  <p>Please log in</p>
{/if}
```

Transforms to Alpine.js `<template x-if>` and `<template x-else>` directives.

### Loops

```html
{for item in items}
  <div>{item.name}</div>
{/for}
```

Transforms to Alpine.js `<template x-for="item in items">`.

### Components

```html
<UserProfile name="John" age={userAge} />
```

Components are imported from `examples/components/` and props are passed automatically.

## Global Store System

The template engine supports global reactive stores for sharing state across components.

### Store Syntax

Reference store data in templates using `{$storeName.property}`:

```html
{if $auth.isLoggedIn}
  <p>Welcome, {$auth.user.name}!</p>
{/if}
```

This transforms to Alpine.js store syntax: `$store.auth.isLoggedIn`

### Defining Stores

#### Option 1: Inline Store Definition

Define stores directly in the fence section:

```html
---
store auth = {
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'John' };
  },
  logout() {
    this.isLoggedIn = false;
    this.user = null;
  }
}

store cart = {
  items: [],
  total: 0,
  addItem(name, price) {
    this.items.push({ name, price });
    this.total += price;
  }
}
---
```

#### Option 2: External Store Files

Create reusable store files in `stores/` directory:

**stores/auth.js:**
```javascript
{
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'User', email: 'user@example.com' };
  },
  logout() {
    this.isLoggedIn = false;
    this.user = null;
  }
}
```

**Import in template fence:**
```html
---
import store from './stores/auth.js'
---
```

#### Option 3: Component Store Imports

Components can import stores they depend on:

**examples/components/CartBadge.html:**
```html
---
import store from './stores/cart.js'
---

<div class="cart-badge">
  {if $cart.items.length > 0}
    <span>{$cart.items.length} items</span>
    <span>${$cart.total}</span>
  {/if}
</div>
```

### Store Priority

When multiple store definitions exist, the priority is:

1. **Inline definitions** (`store name = { ... }` in fence)
2. **Imported files** (`import store from './stores/name.js'` in fence)
3. **External stores** (registered globally)

### Store Methods and Computed Properties

Stores can have methods and getters:

```javascript
// stores/cart.js
{
  items: [],
  total: 0,

  // Computed property
  get formattedTotal() {
    return this.total.toFixed(2);
  },

  // Method
  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  }
}
```

Use in templates:
```html
<p>Total: ${$cart.formattedTotal}</p>
<button @click="$store.cart.addItem({ name: 'Widget', price: 9.99 })">
  Add Item
</button>
```

### Store Expressions in Alpine Directives

Stores work with all Alpine.js directives:

```html
<!-- Conditional rendering -->
<div x-show="$store.auth.isLoggedIn">...</div>

<!-- Event handlers -->
<button @click="$store.auth.login()">Login</button>

<!-- Dynamic styling with :style -->
<body :style="`background: ${$store.theme.getCurrentColors().background}`">
```

## Fence Section

The fence section (between `---` markers) contains:

### Props

Define component props with default values:

```html
---
prop name = "Guest"
prop age = 0
prop isActive = false
---
```

### Variables

Local reactive variables:

```html
---
let count = 0
const apiUrl = "https://api.example.com"
---
```

### Functions

Alpine.js data methods:

```html
---
function increment() {
  this.count++;
}

function formatDate(date) {
  return new Date(date).toLocaleDateString();
}
---
```

### Imports

Import components and stores:

```html
---
import UserProfile from './components/UserProfile.html'
import store from './stores/auth.js'
---
```

## Architecture

The codebase follows a pipeline architecture:

```
Template Source → Parser → AST → Transformer → Rendered HTML/CSS/JS
```

### Key Packages

- **`ast/`** - Abstract Syntax Tree node definitions
- **`parser/`** - Converts template syntax to AST
- **`transformer/`** - Transforms AST to Alpine.js nodes
- **`renderer/`** - Generates final HTML/CSS/JS
- **`scoping/`** - CSS and JS scoping utilities
- **`cmd/server/`** - Development server

See [CLAUDE.md](CLAUDE.md) for detailed architecture documentation.

## Props vs Stores: When to Use Which

### Use Props When:

- Data is **component-specific** and passed from parent
- Each instance needs **different values**
- Data flows **one direction** (parent → child)

Example:
```html
<UserCard name="John" age={30} />
<UserCard name="Jane" age={25} />
```

### Use Stores When:

- State is **shared across multiple components**
- Multiple components need to **read and write** the same data
- State persists across **component instances**
- You need **global application state** (auth, theme, cart, etc.)

Example:
```html
<!-- Multiple components accessing same auth state -->
<LoginButton />    <!-- Uses $auth.login() -->
<UserMenu />       <!-- Shows $auth.user.name -->
<ProtectedRoute /> <!-- Checks $auth.isLoggedIn -->
```

## Examples

### Complete Page with Store

```html
---
prop title = "My Page"

import store from './stores/auth.js'
import store from './stores/cart.js'

store theme = {
  mode: "light",
  toggle() {
    this.mode = this.mode === "light" ? "dark" : "light";
  }
}
---

<!DOCTYPE html>
<html>
<head>
  <title>{title}</title>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
  <header>
    {if $auth.isLoggedIn}
      <p>Welcome, {$auth.user.name}!</p>
      <button @click="$store.auth.logout()">Logout</button>
    {else}
      <button @click="$store.auth.login()">Login</button>
    {/if}
  </header>

  <main>
    <h1>{title}</h1>

    <div class="cart-summary">
      <p>Cart: {$cart.items.length} items</p>
      <p>Total: ${$cart.formattedTotal}</p>
    </div>

    <button @click="$store.theme.toggle()">
      Switch to {$theme.mode === 'light' ? 'Dark' : 'Light'} Mode
    </button>
  </main>
</body>
</html>
```

## Troubleshooting

### Common Store Issues

#### 1. Store Not Found

**Error:** `Cannot read properties of undefined (reading 'propertyName')`

**Solution:** Ensure store is imported or defined:
```html
---
import store from './stores/auth.js'
---
```

#### 2. Store Methods Not Working

**Error:** `$store.cart.addItem is not a function`

**Solution:** Check that:
- Store file exports valid JavaScript object
- Methods use proper `this` binding
- Store is properly registered

#### 3. Double Store Prefix

**Error:** Console shows `$store.store.theme` instead of `$store.theme`

**Solution:** This was a bug fixed in the transformer. Update to latest version.

#### 4. Computed Properties Undefined

**Error:** `.toFixed is not a function`

**Solution:** Use computed properties (getters) instead of method calls in templates:
```javascript
// ✅ Good
get formattedTotal() {
  return this.total.toFixed(2);
}

// ❌ Bad - don't call .toFixed(2) in template
```

### Known Issues

See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for active issues and regressions.

## Testing

### Test Structure

```
tests/
├── alpine/              # Alpine.js integration tests
├── components/          # Component tests
└── integration/         # End-to-end tests

transformer/
├── *_test.go           # Unit tests for transformers
└── stores_test.go      # Store transformation tests

parser/
└── *_test.go           # Parser tests
```

### Running Specific Tests

```bash
# Store system tests
go test ./transformer -run TestStores -v

# Component tests
go test ./tests/components -v

# Alpine integration
go test ./tests/alpine -v
```

## Contributing

When making changes:

1. Write tests first
2. Update CLAUDE.md if architecture changes
3. Add examples for new features
4. Run full test suite before committing
5. Update this README if adding user-facing features

## License

[Your License Here]
