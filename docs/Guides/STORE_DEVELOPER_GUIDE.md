# Global Store System - Developer Guide

This guide provides best practices, patterns, and conventions for using the global store system in the template engine.

## Table of Contents

1. [When to Use Stores vs Props](#when-to-use-stores-vs-props)
2. [Store File Organization](#store-file-organization)
3. [Store Structure Best Practices](#store-structure-best-practices)
4. [Common Store Patterns](#common-store-patterns)
5. [Testing Stores](#testing-stores)
6. [Performance Considerations](#performance-considerations)
7. [Debugging Store Issues](#debugging-store-issues)

---

## When to Use Stores vs Props

### Use Props When:

✅ **Data is component-specific**
```html
<!-- Each UserCard instance has different data -->
<UserCard name="John Doe" age={30} />
<UserCard name="Jane Smith" age={25} />
```

✅ **Parent controls the data**
```html
---
let users = [
  { name: "John", role: "admin" },
  { name: "Jane", role: "user" }
]
---

{for user in users}
  <UserCard name={user.name} role={user.role} />
{/for}
```

✅ **Data flows one direction (parent → child)**
```html
<!-- Parent passes data down -->
<Dashboard>
  <Header title={pageTitle} />
  <Content data={pageData} />
</Dashboard>
```

### Use Stores When:

✅ **State is shared across multiple components**
```html
<!-- Auth state used by multiple components -->
<LoginButton />    <!-- Uses $auth.login() -->
<UserMenu />       <!-- Shows $auth.user.name -->
<ProtectedRoute /> <!-- Checks $auth.isLoggedIn -->
```

✅ **Multiple components need to read AND write**
```html
<!-- Shopping cart updated from multiple places -->
<ProductCard @add="$store.cart.addItem(product)" />
<CartBadge /> <!-- Shows $cart.items.length -->
<CartPage />  <!-- Displays and modifies cart -->
```

✅ **State persists across component instances**
```html
<!-- Theme state persists across all pages -->
{if $theme.mode === 'dark'}
  <body class="dark-mode">
{else}
  <body class="light-mode">
{/if}
```

✅ **Global application state (auth, theme, cart, etc.)**

---

## Store File Organization

### Directory Structure

```
stores/
├── auth.js          # Authentication state
├── cart.js          # Shopping cart
├── theme.js         # Theme/UI preferences
├── notifications.js # Toast notifications
└── user.js          # User profile data
```

### Naming Conventions

- **File names**: lowercase, singular noun (`auth.js`, not `Auth.js` or `auths.js`)
- **Store names**: match file name without extension (`auth` from `auth.js`)
- **Properties**: camelCase (`isLoggedIn`, `userProfile`)
- **Methods**: camelCase verbs (`login()`, `updateProfile()`, `clearCart()`)

### Store File Template

```javascript
// stores/storeName.js
{
  // State properties
  propertyName: defaultValue,
  anotherProperty: null,

  // Computed properties (getters)
  get computedProperty() {
    return this.propertyName.toUpperCase();
  },

  // Methods
  methodName(params) {
    this.propertyName = params;
  },

  reset() {
    this.propertyName = defaultValue;
    this.anotherProperty = null;
  }
}
```

---

## Store Structure Best Practices

### 1. Keep Stores Focused

✅ **Good - Single responsibility:**
```javascript
// stores/auth.js
{
  isLoggedIn: false,
  user: null,
  token: null,

  login(user, token) {
    this.isLoggedIn = true;
    this.user = user;
    this.token = token;
  },

  logout() {
    this.isLoggedIn = false;
    this.user = null;
    this.token = null;
  }
}
```

❌ **Bad - Too many responsibilities:**
```javascript
// stores/app.js - TOO BROAD
{
  isLoggedIn: false,
  user: null,
  cartItems: [],
  theme: 'light',
  notifications: [],
  // ... everything in one store
}
```

### 2. Use Computed Properties for Derived State

✅ **Good - Use getters:**
```javascript
// stores/cart.js
{
  items: [],
  total: 0,

  get formattedTotal() {
    return this.total.toFixed(2);
  },

  get itemCount() {
    return this.items.length;
  },

  get isEmpty() {
    return this.items.length === 0;
  }
}
```

❌ **Bad - Method calls in templates:**
```html
<!-- Don't do this: -->
<p>Total: ${$cart.total.toFixed(2)}</p> <!-- ERROR -->

<!-- Do this instead: -->
<p>Total: ${$cart.formattedTotal}</p> <!-- ✅ -->
```

### 3. Provide Reset/Clear Methods

✅ **Good - Always include reset:**
```javascript
{
  items: [],
  total: 0,

  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  },

  clear() {
    this.items = [];
    this.total = 0;
  }
}
```

### 4. Use Nested Objects for Complex State

✅ **Good - Organized structure:**
```javascript
// stores/auth.js
{
  isLoggedIn: false,
  user: {
    name: "",
    email: "",
    role: "guest"
  },
  session: {
    token: null,
    expiresAt: null
  },

  login(userData, sessionData) {
    this.isLoggedIn = true;
    this.user = userData;
    this.session = sessionData;
  }
}
```

---

## Common Store Patterns

### 1. Authentication Store

```javascript
// stores/auth.js
{
  isLoggedIn: false,
  user: {
    id: null,
    name: "",
    email: "",
    role: "guest"
  },

  get isAdmin() {
    return this.user.role === 'admin';
  },

  login(userData) {
    this.isLoggedIn = true;
    this.user = {
      id: userData.id,
      name: userData.name,
      email: userData.email,
      role: userData.role || 'user'
    };
  },

  logout() {
    this.isLoggedIn = false;
    this.user = {
      id: null,
      name: "",
      email: "",
      role: "guest"
    };
  },

  updateProfile(updates) {
    this.user = { ...this.user, ...updates };
  }
}
```

**Usage in templates:**
```html
{if $auth.isLoggedIn}
  <p>Welcome, {$auth.user.name}!</p>
  {if $auth.isAdmin}
    <a href="/admin">Admin Panel</a>
  {/if}
  <button @click="$store.auth.logout()">Logout</button>
{else}
  <button @click="$store.auth.login({ name: 'User', email: 'user@example.com' })">
    Login
  </button>
{/if}
```

### 2. Shopping Cart Store

```javascript
// stores/cart.js
{
  items: [],
  total: 0,

  get formattedTotal() {
    return this.total.toFixed(2);
  },

  get itemCount() {
    return this.items.length;
  },

  get isEmpty() {
    return this.items.length === 0;
  },

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

**Usage in templates:**
```html
<!-- Cart badge component -->
<div class="cart-badge">
  🛒 {$cart.itemCount} items - ${$cart.formattedTotal}
</div>

<!-- Product card -->
<button @click="$store.cart.addItem({ name: 'Widget', price: 9.99 })">
  Add to Cart
</button>

<!-- Cart page -->
{if $cart.isEmpty}
  <p>Your cart is empty</p>
{else}
  {for item in $cart.items}
    <div>{item.name} - ${item.price}</div>
  {/for}
  <p>Total: ${$cart.formattedTotal}</p>
{/if}
```

### 3. Theme Store

```javascript
// stores/theme.js
{
  mode: "light",

  get colors() {
    return this.mode === 'dark' ? {
      background: '#1a1a1a',
      text: '#e0e0e0',
      primary: '#4a9eff'
    } : {
      background: '#ffffff',
      text: '#1a1a1a',
      primary: '#2563eb'
    };
  },

  toggle() {
    this.mode = this.mode === 'light' ? 'dark' : 'light';
  },

  setMode(mode) {
    this.mode = mode;
  }
}
```

**Usage in templates:**
```html
<body :style="`background: ${$store.theme.colors.background};
               color: ${$store.theme.colors.text}`">

  <button @click="$store.theme.toggle()">
    Switch to {$theme.mode === 'dark' ? 'Light' : 'Dark'} Mode
  </button>
</body>
```

### 4. Notification/Toast Store

```javascript
// stores/notifications.js
{
  messages: [],
  nextId: 1,

  add(message, type = 'info', duration = 3000) {
    const id = this.nextId++;
    this.messages.push({ id, message, type, duration });

    // Auto-remove after duration
    setTimeout(() => {
      this.remove(id);
    }, duration);
  },

  remove(id) {
    this.messages = this.messages.filter(m => m.id !== id);
  },

  clear() {
    this.messages = [];
  }
}
```

**Usage in templates:**
```html
<!-- Notification container -->
<div class="notifications">
  {for notification in $notifications.messages}
    <div class="notification {notification.type}">
      {notification.message}
      <button @click="$store.notifications.remove({notification.id})">×</button>
    </div>
  {/for}
</div>

<!-- Trigger from anywhere -->
<button @click="$store.notifications.add('Item added to cart!', 'success')">
  Add Item
</button>
```

---

## Testing Stores

### Store Testing Checklist

When creating a new store, test these scenarios:

1. **Initial state** - Verify default values
2. **State mutations** - Test each method updates state correctly
3. **Computed properties** - Verify getters return correct derived values
4. **Edge cases** - Test with empty, null, or invalid data
5. **Reset functionality** - Ensure clear/reset methods work
6. **Template integration** - Verify expressions transform correctly

### Testing Store Files

Test store files independently before integration:

```bash
# Create test page with inline store
cat > examples/pages/test-cart-store.html << 'EOF'
---
import store from './stores/cart.js'
---

<div>
  <p>Items: {$cart.itemCount}</p>
  <p>Total: ${$cart.formattedTotal}</p>
  <button @click="$store.cart.addItem({ name: 'Test', price: 10 })">Add</button>
  <button @click="$store.cart.clear()">Clear</button>
</div>
EOF

# Run server and test
go run cmd/server/main.go
# Visit http://localhost:3000/test-cart-store
```

---

## Performance Considerations

### 1. Avoid Large Arrays in Stores

❌ **Bad - Store thousands of items:**
```javascript
{
  products: [], // Don't store entire product catalog
  loadProducts() {
    fetch('/api/products').then(r => r.json())
      .then(data => this.products = data); // 10,000+ items
  }
}
```

✅ **Good - Store only what's needed:**
```javascript
{
  selectedProduct: null,
  recentProducts: [], // Keep only last 10

  selectProduct(product) {
    this.selectedProduct = product;
  },

  addToRecent(product) {
    this.recentProducts.unshift(product);
    this.recentProducts = this.recentProducts.slice(0, 10); // Keep last 10
  }
}
```

### 2. Use Computed Properties for Expensive Calculations

✅ **Good - Cache expensive calculations:**
```javascript
{
  items: [],

  get totalPrice() {
    return this.items.reduce((sum, item) => sum + item.price, 0);
  },

  get averagePrice() {
    return this.items.length > 0 ? this.totalPrice / this.items.length : 0;
  }
}
```

### 3. Minimize Store Dependencies

Keep stores independent when possible. If Store B depends on Store A, consider:

1. Can this be a computed property instead?
2. Can the component handle the relationship?
3. Is this actually one store with related state?

---

## Debugging Store Issues

### Common Issues and Solutions

#### Issue 1: Store Not Found

**Error:** `Cannot read properties of undefined (reading 'propertyName')`

**Debug steps:**
1. Check store is imported or defined in fence section
2. Verify store file path is correct
3. Ensure store name matches file name
4. Check for typos in store name

**Solution:**
```html
---
import store from './stores/auth.js'  <!-- ✅ Correct path -->
<!-- NOT: import store from './store/auth.js' --> <!-- ❌ Wrong path -->
---
```

#### Issue 2: Methods Not Working

**Error:** `$store.cart.addItem is not a function`

**Debug steps:**
1. Check store file has valid JavaScript syntax
2. Verify method uses `this` binding correctly
3. Ensure store file is a plain object, not wrapped in function

**Solution:**
```javascript
// ✅ Correct
{
  items: [],
  addItem(item) {
    this.items.push(item);
  }
}

// ❌ Wrong - function wrapper
function() {
  return {
    items: [],
    addItem(item) { ... }
  }
}
```

#### Issue 3: Computed Properties Not Updating

**Error:** Getter returns stale data

**Debug steps:**
1. Ensure getter uses `this.property` not local variable
2. Check that underlying properties are being mutated correctly
3. Verify getter logic is correct

**Solution:**
```javascript
// ✅ Correct - uses this.items
get itemCount() {
  return this.items.length;
}

// ❌ Wrong - uses external variable
let items = [];
get itemCount() {
  return items.length; // Won't react to this.items changes
}
```

#### Issue 4: Store Expressions Not Transforming

**Error:** Template shows `{$auth.isLoggedIn}` as literal text

**Debug steps:**
1. Check syntax: `{$storeName.property}` (no spaces)
2. Verify store is properly imported
3. Check for typos in store name
4. Look for syntax errors in template

**Solution:**
```html
<!-- ✅ Correct -->
{$auth.isLoggedIn}

<!-- ❌ Wrong - extra spaces -->
{ $auth.isLoggedIn }

<!-- ❌ Wrong - missing $ -->
{auth.isLoggedIn}
```

### Debugging Tools

1. **Check rendered HTML:** Look at generated `x-data` to see if stores are initialized
2. **Browser console:** Use `Alpine.store('storeName')` to inspect store state
3. **Test pages:** Create minimal test pages to isolate issues
4. **Server logs:** Check for store registration errors

---

## Best Practices Summary

### Do ✅

- Keep stores focused on a single domain (auth, cart, theme)
- Use computed properties for derived state
- Provide reset/clear methods
- Use meaningful, descriptive names
- Store minimal data (only what's needed globally)
- Test stores independently before integration
- Document complex store logic

### Don't ❌

- Mix multiple concerns in one store
- Call methods with arguments in templates (use getters)
- Store large datasets or entire API responses
- Create circular dependencies between stores
- Mutate store state from outside store methods
- Use stores for component-specific state (use props instead)

---

## Additional Resources

- [README.md](../README.md) - User guide with store syntax
- [CLAUDE.md](../CLAUDE.md) - Technical architecture
- [stores/](../stores/) - Example store files
- [examples/pages/store-components-demo.html](../examples/pages/store-components-demo.html) - Full demo
- [.agent-os/specs/2025-10-07-global-store-system/](../.agent-os/specs/2025-10-07-global-store-system/) - Implementation specs
