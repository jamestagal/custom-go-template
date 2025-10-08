# Store File Format Documentation

## Overview

Store files define Alpine.js global reactive state that can be shared across all templates and components. The template engine automatically discovers, registers, and initializes stores from the `stores/` directory.

## File Location

All store files must be placed in the `stores/` directory at the project root:

```
custom_go_template/
├── stores/
│   ├── auth.js       # Authentication store
│   ├── cart.js       # Shopping cart store
│   ├── theme.js      # Theme/UI preferences store
│   └── README.md     # This documentation
```

## File Format

Store files must:
1. Be valid JavaScript files with `.js` extension
2. Contain a single JavaScript object literal
3. Use ES5+ syntax (methods, getters, etc.)
4. **NOT** include `Alpine.store()` wrapper (added automatically)
5. **NOT** include surrounding parentheses or semicolons

### Basic Structure

```javascript
{
  // State properties
  propertyName: initialValue,

  // Methods
  methodName(param) {
    // Access state with 'this'
    this.propertyName = newValue;
  },

  // Getters (computed values)
  get computedName() {
    return this.propertyName + ' computed';
  }
}
```

## Store Registration

### Automatic Discovery

The server automatically:
1. Scans the `stores/` directory on startup
2. Reads all `.js` files
3. Extracts store names from filenames (`auth.js` → `auth`)
4. Registers stores with Alpine.js: `Alpine.store('auth', { ... })`

### Naming Convention

- **Filename**: `storeName.js` (lowercase, hyphen-separated)
- **Store name**: Extracted from filename (e.g., `theme-settings.js` → `theme-settings`)
- **Template access**: `$storeName` (e.g., `$auth`, `$cart`, `$theme`)

## Usage in Templates

### In Fence Section

Declare which stores a template uses:

```html
---
store auth = { ... }  // Use existing auth.js store
store cart = { ... }  // Use existing cart.js store
---
```

**Note**: The inline definition is ignored; the store content comes from the `.js` file. The declaration is just to indicate which stores the template needs.

### In Template Body

Access stores with `$storeName` prefix:

```html
<!-- Read store property -->
<span>{$auth.user.name}</span>

<!-- Call store method -->
<button @click="$store.auth.login()">Login</button>

<!-- Use in conditionals -->
{if $auth.isLoggedIn}
  <p>Welcome back!</p>
{/if}

<!-- Use in loops -->
{for item in $cart.items}
  <div>{item.name}</div>
{/for}
```

## Example Store Files

### 1. auth.js - Authentication Store

```javascript
{
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'Test User', email: 'test@example.com' };
  },
  logout() {
    this.isLoggedIn = false;
    this.user = null;
  }
}
```

**Usage**:
```html
{if $auth.isLoggedIn}
  <span>Welcome, {$auth.user.name}</span>
  <button @click="$store.auth.logout()">Logout</button>
{else}
  <button @click="$store.auth.login()">Login</button>
{/if}
```

### 2. cart.js - Shopping Cart Store

```javascript
{
  items: [],
  total: 0,
  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  },
  removeItem(index) {
    const item = this.items[index];
    this.items.splice(index, 1);
    this.total -= item.price;
  },
  clear() {
    this.items = [];
    this.total = 0;
  }
}
```

**Usage**:
```html
<span>Cart: {$cart.items.length} items</span>
<span>Total: ${$cart.total}</span>

{for item in $cart.items}
  <div>
    <span>{item.name}</span>
    <button @click="$store.cart.removeItem({item})">Remove</button>
  </div>
{/for}
```

### 3. theme.js - Theme Store

```javascript
{
  mode: 'light',
  colors: {
    light: {
      background: '#ffffff',
      text: '#1a202c',
      primary: '#3182ce',
      secondary: '#718096'
    },
    dark: {
      background: '#1a202c',
      text: '#f7fafc',
      primary: '#63b3ed',
      secondary: '#a0aec0'
    }
  },
  getCurrentColors() {
    return this.colors[this.mode];
  },
  setLight() {
    this.mode = 'light';
    localStorage.setItem('theme', 'light');
  },
  setDark() {
    this.mode = 'dark';
    localStorage.setItem('theme', 'dark');
  },
  toggle() {
    this.mode = this.mode === 'light' ? 'dark' : 'light';
    localStorage.setItem('theme', this.mode);
  },
  init() {
    const saved = localStorage.getItem('theme');
    if (saved) {
      this.mode = saved;
    } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      this.mode = 'dark';
    }
  }
}
```

**Usage**:
```html
<div :style="`background: ${$store.theme.getCurrentColors().background}; color: ${$store.theme.getCurrentColors().text}`">
  <button @click="$store.theme.toggle()">
    Toggle Theme (Current: {$theme.mode})
  </button>
</div>
```

## Best Practices

### 1. State Management

✅ **DO**: Keep state minimal and focused
```javascript
{
  count: 0,
  increment() { this.count++; }
}
```

❌ **DON'T**: Mix unrelated concerns
```javascript
{
  count: 0,
  username: '',  // Different concern - use separate store
  theme: 'dark'  // Different concern - use separate store
}
```

### 2. Method Naming

✅ **DO**: Use clear, action-based method names
```javascript
{
  login() { ... },
  logout() { ... },
  updateProfile(data) { ... }
}
```

❌ **DON'T**: Use ambiguous names
```javascript
{
  do() { ... },        // What does it do?
  handle() { ... },    // Handle what?
  process() { ... }    // Process what?
}
```

### 3. Persistence

✅ **DO**: Persist important state to localStorage
```javascript
{
  mode: 'light',
  setMode(mode) {
    this.mode = mode;
    localStorage.setItem('theme-mode', mode);  // Persist
  },
  init() {
    const saved = localStorage.getItem('theme-mode');
    if (saved) this.mode = saved;  // Restore
  }
}
```

### 4. Computed Values

✅ **DO**: Use getters for derived values
```javascript
{
  items: [],
  get total() {
    return this.items.reduce((sum, item) => sum + item.price, 0);
  }
}
```

❌ **DON'T**: Store computed values as state
```javascript
{
  items: [],
  total: 0,  // Must manually update when items change
  addItem(item) {
    this.items.push(item);
    this.total += item.price;  // Easy to forget or get wrong
  }
}
```

### 5. Nested Objects

✅ **DO**: Use nested objects for related data
```javascript
{
  user: {
    profile: {
      name: 'John',
      email: 'john@example.com'
    },
    preferences: {
      theme: 'dark',
      notifications: true
    }
  }
}
```

Access in templates: `{$auth.user.profile.name}`

## Store Lifecycle

### 1. Server Startup
```
Server starts → registerStores() called
  ↓
Scans stores/ directory
  ↓
Reads *.js files
  ↓
Stores in memory (map[string]string)
```

### 2. Template Rendering
```
Template parsed → Fence section declares stores
  ↓
Transformer tracks store references
  ↓
GetReferencedStoreDefinitions() called
  ↓
Renderer combines store scripts
  ↓
Output includes Alpine.store() initialization
```

### 3. Browser Runtime
```
Page loads → alpine:init event fires
  ↓
Alpine.store() calls execute
  ↓
Stores available globally as $store.storeName
  ↓
Components can read/modify store
  ↓
Alpine.js reactivity updates all components
```

## Troubleshooting

### Store not found

**Problem**: Template references `$myStore` but it's not initialized

**Solutions**:
1. Check file exists: `stores/myStore.js`
2. Check file format (valid JavaScript object)
3. Check fence section declares: `store myStore = { ... }`
4. Restart server to reload stores

### Store not updating

**Problem**: Store changes don't trigger UI updates

**Solutions**:
1. Ensure using `$store.storeName` prefix (not just `$storeName`)
2. Check Alpine.js is loaded
3. Verify store methods use `this.property = value` (triggers reactivity)
4. Check browser console for Alpine.js errors

### Syntax error in store file

**Problem**: Store file has JavaScript syntax error

**Solutions**:
1. Validate JSON/JavaScript syntax
2. Remove `Alpine.store()` wrapper (added automatically)
3. Remove trailing semicolons
4. Check method syntax: `methodName() { }` not `methodName: function() { }`

## Advanced Patterns

### Store with API Integration

```javascript
{
  users: [],
  loading: false,
  error: null,
  async fetchUsers() {
    this.loading = true;
    this.error = null;
    try {
      const response = await fetch('/api/users');
      this.users = await response.json();
    } catch (err) {
      this.error = err.message;
    } finally {
      this.loading = false;
    }
  }
}
```

### Store with Computed Properties

```javascript
{
  items: [],
  get count() {
    return this.items.length;
  },
  get isEmpty() {
    return this.items.length === 0;
  },
  get total() {
    return this.items.reduce((sum, item) => sum + item.price, 0);
  }
}
```

### Store with Validation

```javascript
{
  email: '',
  password: '',
  get isValidEmail() {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.email);
  },
  get isValidPassword() {
    return this.password.length >= 8;
  },
  get canSubmit() {
    return this.isValidEmail && this.isValidPassword;
  }
}
```

## Related Documentation

- **Template Syntax**: See `CLAUDE.md` for full template syntax guide
- **Alpine.js Stores**: https://alpinejs.dev/globals/alpine-store
- **Integration Tests**: See `tests/integration/store_*.go` for usage examples
- **Phase 3 Completion**: See `.agent-os/specs/2025-10-07-global-store-system/TASK3.3_COMPLETION_REPORT.md`
