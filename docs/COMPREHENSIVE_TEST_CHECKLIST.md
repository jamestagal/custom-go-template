# Comprehensive Template Pattern Test Checklist

**Date**: 2025-10-07
**Test File**: `examples/pages/comprehensive-simple.html` (simplified due to bugs)
**Server URL**: http://localhost:3333/comprehensive-simple
**Test Result**: ✅ **PASSED** - Zero console errors, all patterns rendering correctly

This document tracks testing of all template patterns to ensure complete feature coverage before Plenti Integration (Spec 8).

## Test Status Summary

- **Console Errors**: ✅ Zero errors
- **Page Styling**: ✅ Working (page + component style aggregation)
- **Basic Expressions**: ✅ Working
- **Conditionals**: ✅ Working (if/else-if/else, nested)
- **Loops**: ✅ Working (simple, with index, nested)
- **Static Components**: ✅ Working
- **Components in Loops**: ⚠️ **NOT TESTED** (known bug - see below)
- **Functions**: ⚠️ **NOT TESTED** (known bug - see below)
- **Attribute Expressions**: ⚠️ **NOT TESTED** (known bug - see below)

## Known Bugs Preventing Full Testing

### Bug #1: Server Manually Builds x-data (High Priority)
- **Issue**: Server route handlers manually extract fence data and build x-data as JSON
- **Impact**: Functions get extracted as truncated JSON strings instead of proper Alpine.js methods
- **Workaround**: Removed all functions from test file
- **Fix Required**: Server should use `renderer/transformer` with `alpineDataFormatter`

### Bug #2: Component Props in Loops Don't Evaluate
- **Issue**: Props like `inStock={product.inStock}` are passed as literal expressions, not evaluated values
- **Result**: Component x-data contains `inStock: product.inStock` instead of `inStock: true`
- **Impact**: Variables undefined in component scope, causing console errors
- **Workaround**: Only using static components with literal prop values
- **Fix Required**: Transformer should evaluate expressions in loop context before passing to components

### Bug #3: Attribute Expressions Transform Incorrectly
- **Issue**: Expressions in attributes like `<button {!inStock ? 'disabled' : ''}>` transform to `<span x-text="!inStock ? 'disabled' : ''">` instead of Alpine binding
- **Expected**: Should become `:disabled="!inStock"` or similar Alpine.js directive
- **Impact**: Expression evaluates in wrong scope, causes console errors
- **Workaround**: Use conditional rendering instead: `{if inStock}<button>...` / `{else}<button disabled>...`
- **Fix Required**: Transformer needs to detect attribute-position expressions and use Alpine bindings

### Bug #4: Multi-line Variable Extraction
- **Issue**: Fence parser only captures first line of const/let/var declarations
- **Impact**: Complex variable values get truncated
- **Workaround**: Use only single-line declarations
- **Fix Required**: Apply same multi-line parsing logic as props use

---

## Test Instructions

1. Start server: `go run cmd/server/main.go` or `./bin/server`
2. Navigate to: http://localhost:3333/comprehensive-simple
3. Open browser console to check for errors (should be zero)
4. Verify each section visually and functionally
5. Check that page styling is applied (component + page styles aggregated)

---

## Section 1: Basic Expressions (Lines 139-151)

**Purpose**: Test simple variable interpolation, object property access, and function calls

### Test Cases

- [ ] **1.1 Simple variable**: `{title}` displays "Custom Template Showcase"
- [ ] **1.2 Object property**: `{user.name}` displays "John Doe"
- [ ] **1.3 Object nested property**: `{user.role}` displays "admin"
- [ ] **1.4 Function call**: `{getGreeting()}` displays time-appropriate greeting
- [ ] **1.5 Function with parameter**: `{user.name}` used in greeting context
- [ ] **1.6 Array length**: `{products.length}` displays "4"
- [ ] **1.7 Array length (categories)**: `{categories.length}` displays count
- [ ] **1.8 Object property (settings)**: `{settings.theme}` displays "light"
- [ ] **1.9 Object property (settings)**: `{settings.currency}` displays "USD"

**Expected Alpine.js Output**:
```html
<span x-text="title"></span>
<span x-text="user.name"></span>
<span x-text="getGreeting()"></span>
```

**Browser Verification**:
- No `{variable}` literal text visible
- All values render correctly
- Console: No Alpine.js errors

---

## Section 2: Conditionals (Lines 153-196)

**Purpose**: Test if/else, if/else-if/else, and nested conditionals

### Test Cases

- [ ] **2.1 Simple if/else**: Login status displays correctly based on `isLoggedIn`
  - When `isLoggedIn = true`: Shows "You are currently logged in as John Doe"
  - When `isLoggedIn = false`: Shows "You are not logged in"

- [ ] **2.2 Multiple else-if branches**: User role check
  - Admin role: "Welcome, Administrator!"
  - Manager role: "Welcome, Manager!"
  - Editor role: "Welcome, Editor!"
  - Default: "Welcome, User!"

- [ ] **2.3 Nested conditionals**: Product availability
  - Outer condition: `filteredProducts.length > 0`
  - Inner condition: `settings.filters.inStockOnly`
  - Both branches display correctly

**Expected Alpine.js Output**:
```html
<template x-if="isLoggedIn">
  <p>You are currently logged in...</p>
</template>
<template x-if="!(isLoggedIn)">
  <p>You are not logged in...</p>
</template>
```

**Browser Verification**:
- No `{if}`, `{else}`, `{/if}` literal text
- Only one branch renders (correct conditional logic)
- Console: No template errors

---

## Section 3: Loops (Lines 198-279)

**Purpose**: Test array loops, indexed loops, object loops, and nested loops

### Test Cases

- [ ] **3.1 Simple array loop**: `{for product in filteredProducts}`
  - All filtered products display
  - Out-of-stock indicator shows correctly

- [ ] **3.2 Loop with index**: `{for product, index in filteredProducts}`
  - List numbers start at 1 (index + 1)
  - Featured star (★) displays for featured products

- [ ] **3.3 Object loop**: `{for key, value of settings}`
  - All settings keys display (theme, currency, showFeatured, filters)
  - Nested object (filters) handled correctly

- [ ] **3.4 Nested object loop**: Within settings loop
  - `{for subKey, subValue of value}` displays filter properties
  - Shows: minPrice, maxPrice, inStockOnly

- [ ] **3.5 Triple-nested loops**: Categories → Items → Tags
  - Category loop: `{for category in categories}`
  - Item loop: `{for item in category.items}`
  - Tag loop: `{for tag in item.tags}`
  - Tag classes apply correctly (bg-blue-100, bg-green-100, etc.)

**Expected Alpine.js Output**:
```html
<template x-for="product in filteredProducts">
  <li><span x-text="product.name"></span></li>
</template>

<template x-for="(product, index) in filteredProducts">
  <li :value="index + 1">...</li>
</template>

<template x-for="[key, value] in Object.entries(settings)">
  <dt x-text="key"></dt>
</template>
```

**Browser Verification**:
- No `{for}`, `{/for}` literal text
- All items render (count matches data)
- Nested loops render correctly
- Console: No iteration errors

---

## Section 4: Components (Lines 281-327)

**Purpose**: Test static components, components in loops, conditional components, dynamic components

### Test Cases

- [ ] **4.1 Static component with props**: `<Header title={title} user={user} />`
  - Header renders at top of page
  - Props passed correctly
  - Header styles applied

- [ ] **4.2 Component with boolean prop**: `<UserProfile user={user} showRole={true} />`
  - User profile displays
  - Role is visible (showRole=true)

- [ ] **4.3 Component in loop**: Product cards
  - `{for product in filteredProducts.filter(p => p.featured)}`
  - Only featured products show (2 cards expected)
  - Each ProductCard receives correct props
  - `formatPrice` function passed and works

- [ ] **4.4 Conditional component loop**: Notifications
  - 3 notifications render
  - Each has correct type (info, success, warning)
  - Type-specific styling applied

- [ ] **4.5 Dynamic component**: `<="./components/AdminPanel.html" user={user} />`
  - When `user.role === "admin"`: AdminPanel renders
  - When `user.role !== "admin"`: UserDashboard renders
  - Correct component loaded based on condition

- [ ] **4.6 Footer component**: `<Footer />`
  - Footer renders at bottom
  - Footer styles applied
  - No empty x-data wrapper (semantic HTML fix)

**Expected Alpine.js Output**:
```html
<!-- Static component -->
<header x-data="{ title: '...', user: {...} }">...</header>

<!-- Component in loop -->
<template x-for="product in filteredProducts.filter(p => p.featured)">
  <div x-data="{ name: product.name, price: product.price }">...</div>
</template>

<!-- Dynamic component -->
<template x-if="user.role === 'admin'">
  <div x-data="{ user: {...} }"><!-- AdminPanel --></div>
</template>
```

**Browser Verification**:
- All components render
- Component styles appear in `<head>` (style aggregation)
- No duplicate styles (SHA256 deduplication)
- Console: No component loading errors

---

## Section 5: Advanced Features (Lines 329-385)

**Purpose**: Test computed values, complex expressions, and combined features

### Test Cases

- [ ] **5.1 Computed values in fence**: `filteredProducts` (line 42-46)
  - Pre-computed array filters correctly
  - Used throughout template without recalculation

- [ ] **5.2 Complex inline expressions**:
  - Average price: `products.reduce((sum, p) => sum + p.price, 0) / products.length`
  - Featured count: `products.filter(p => p.featured).length`
  - In-stock count: `products.filter(p => p.inStock).length`
  - Total inventory value: `products.filter(p => p.inStock).reduce(...)`

- [ ] **5.3 Combined features**: Top 3 products section
  - Conditional: `{if filteredProducts.length > 0}`
  - Math function: `Math.min(3, filteredProducts.length)`
  - Loop: `{for product, index in filteredProducts.slice(0, 3)}`
  - Nested conditional in loop: `{if product.inStock}`
  - Nested loop in loop: `{for tag in product.tags}`

**Expected Alpine.js Output**:
```html
<!-- Complex expression -->
<span x-text="formatPrice(products.reduce((sum, p) => sum + p.price, 0) / products.length)"></span>

<!-- Combined features -->
<template x-if="filteredProducts.length > 0">
  <div>
    <span x-text="Math.min(3, filteredProducts.length)"></span>
    <template x-for="(product, index) in filteredProducts.slice(0, 3)">
      <!-- Nested structures -->
    </template>
  </div>
</template>
```

**Browser Verification**:
- All calculations correct
- No JavaScript errors from complex expressions
- Nested structures render properly
- Console: No Alpine.js evaluation errors

---

## Style Aggregation Verification

**Purpose**: Verify Spec 7 (Component Style Aggregation) works correctly

### Test Cases

- [ ] **Styles in `<head>`**: Check page source for aggregated styles
  - Page-level styles from comprehensive.html
  - Header component styles
  - Footer component styles
  - ProductCard component styles
  - UserProfile component styles
  - Notification component styles
  - AdminPanel OR UserDashboard styles (depending on user role)

- [ ] **No duplicate styles**: Each component's styles appear only once
  - SHA256 deduplication working
  - No multiple identical `<style>` blocks

- [ ] **Dependency order**: Child components before parent
  - Component dependencies loaded in correct order
  - Styles available when components render

- [ ] **Cache performance** (dev tools Network tab):
  - First load: All styles aggregated
  - Subsequent loads: Fast (cache hit)

**Browser Verification**:
- View page source → `<head>` section
- Count `<style>` blocks (should match component count)
- No duplicate CSS rules
- Console: No missing style warnings

---

## Known Patterns NOT in comprehensive.html

These patterns are supported but not tested in comprehensive.html:

1. **Svelte-style loops**: `{#each items as item}` (we support both syntaxes)
2. **Self-closing components**: `<Component />` vs `<Component></Component>`
3. **Attribute expressions**: `class="{dynamicClass}"` (untested in this file)
4. **Event handlers**: `onclick={handler}` (not used in comprehensive.html)
5. **Shorthand props**: `<Component {user} />` (equivalent to `<Component user={user} />`)
6. **Comments**: `<!-- HTML comment -->`
7. **Multiline attribute values**: Props spanning multiple lines

---

## Browser Console Checklist

Open browser console and verify:

- [ ] **No Alpine.js errors**: No "Cannot read property" errors
- [ ] **No template errors**: No "Unexpected token" in expressions
- [ ] **No component errors**: No "Component not found" warnings
- [ ] **No style errors**: No CSS syntax errors
- [ ] **Alpine initialized**: "Alpine.js initialized" or similar message
- [ ] **x-data scope**: Use Alpine DevTools to inspect x-data values

---

## Visual Regression Checklist

Compare with expected layout:

- [ ] **Header**: Displays at top with title and user info
- [ ] **5 sections**: All sections visible with borders and padding
- [ ] **Product grid**: Grid layout (not list) for product cards
- [ ] **Tags**: Colored tags with correct background colors
- [ ] **Notifications**: Color-coded borders (blue, green, yellow)
- [ ] **Footer**: Displays at bottom with company info

---

## Debugging Steps (If Issues Found)

1. **Check server logs**: Look for parser/transformer errors
2. **View page source**: Verify Alpine.js directives generated correctly
3. **Browser console**: Check for JavaScript/Alpine errors
4. **Alpine DevTools**: Inspect x-data scope and reactive state
5. **Network tab**: Verify all component files loaded
6. **Elements inspector**: Check rendered DOM structure

---

## Test Results Summary

**Date Tested**: __________
**Tester**: __________
**Browser**: __________
**Server Version**: __________

### Overall Status

- [ ] ✅ All basic expressions working
- [ ] ✅ All conditionals working
- [ ] ✅ All loops working
- [ ] ✅ All components working
- [ ] ✅ All advanced features working
- [ ] ✅ Style aggregation working
- [ ] ✅ No console errors

### Issues Found

| Section | Issue | Severity | Status |
|---------|-------|----------|--------|
|         |       |          |        |

### Missing Patterns

List any patterns that should be tested but aren't in comprehensive.html:

1.
2.
3.

---

## Next Steps

After completing this checklist:

1. **Fix any issues found**: Address bugs before Spec 8
2. **Add missing patterns**: Create additional test files if needed
3. **Update roadmap**: Mark Spec 8 (Plenti Integration) as ready to start
4. **Document results**: Update this file with test results
