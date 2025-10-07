# Comprehensive Template Pattern Test Checklist

**Date**: 2025-10-07 (Last Updated)
**Primary Test Files**:
- `examples/pages/comprehensive-simple.html` - http://localhost:3333/comprehensive-simple
- `examples/pages/home.html` - http://localhost:3333/home (advanced features)
- `examples/pages/comprehensive.html` - http://localhost:3333/comprehensive (original full test)

**Test Result**: ✅ **PASSED** - Zero console errors, all core patterns rendering correctly

This document tracks testing of all template patterns to ensure complete feature coverage before Plenti Integration (Spec 8).

## Test Status Summary

### Core Features
- **Console Errors**: ✅ Zero errors (both home.html and comprehensive-simple.html)
- **Page Styling**: ✅ Working (page + component style aggregation)
- **Basic Expressions**: ✅ Working (comprehensive-simple.html)
- **Conditionals**: ✅ Working (if/else-if/else, nested) (comprehensive-simple.html)
- **Loops (arrays)**: ✅ Working (simple, with index, nested) (comprehensive-simple.html)
- **Loops (objects)**: ✅ Working (`{for key, value of object}`) (home.html - NOT in comprehensive-simple)
- **Static Components**: ✅ Working (comprehensive-simple.html)
- **Dynamic Components**: ✅ Working (`<="path" />` syntax) (home.html - NOT in comprehensive-simple)
- **Functions**: ✅ **FIXED** - Working in fence section (comprehensive-simple.html)
- **Reactive State**: ✅ Working (login/logout demo) (comprehensive-simple.html)

### Advanced Features (Tested in home.html)
- **Array Spread in Loops**: ✅ Working `{for item of ["new", ...array]}`
- **Variable Component Paths**: ✅ Working `<='{pathVar}'` and `<="./path/{var}.html"`
- **Notification Components**: ✅ Working (loop with different types: success, info, warning, error)
- **Interactive State**: ✅ Working (onclick handlers updating reactive vars)
- **String Manipulation**: ✅ Working (`.split().reverse().join()`)
- **Array Methods**: ✅ Working (`.filter()`, inline arrays)

### Missing from comprehensive-simple.html
- **Object Loops**: ⚠️ NOT showcased (but works in home.html)
- **Dynamic Components**: ⚠️ NOT showcased (but works in home.html)
- **Computed Values**: ⚠️ NOT showcased (`const filtered = array.filter(...)`)
- **Complex Array Methods**: ⚠️ NOT showcased (`.reduce()`, chained `.filter().map()`)
- **Function Props**: ⚠️ NOT showcased (passing `formatPrice={formatPrice}` to components)
- **Array Spread**: ⚠️ NOT showcased (but works in home.html)

## Fixed Issues (Previously Bugs)

### ✅ FIXED #1: Server x-data Building
- **Was**: Server manually extracted fence data and built x-data as JSON
- **Fix**: Server now uses `renderer.Render()` and `buildXDataFromProps()` with proper function extraction
- **Status**: Functions work correctly in fence section (see comprehensive-simple.html login/logout demo)
- **Spec**: Completed in fix-server-xdata-building spec (2025-10-07)

### Still TODO: Component Props in Loops
- **Issue**: Props like `inStock={product.inStock}` in loops may not evaluate correctly
- **Status**: Not fully tested with complex scenarios
- **Workaround**: Use static prop values or test more thoroughly

### Still TODO: Attribute Expressions
- **Issue**: Expressions in attributes like `class="{dynamicClass}"` may need Alpine binding support
- **Status**: Not tested in current files
- **Future**: May need transformer enhancement for attribute-position expressions

---

## Feature Coverage by File

### home.html - Advanced Features Showcase
**URL**: http://localhost:3333/home

**Unique Features Tested**:
1. **Dynamic Components with `<=` syntax** (Lines 103-105):
   - Static path: `<="./components/UserProfile.html"`
   - Variable path: `<='{path}'`
   - Template literal: `<="./components/{comp}.html"`

2. **Object Loops with `of` syntax** (Lines 111+):
   - Simple: `{for animal of animals}`
   - Inline array: `{for word of ["item1", "item2"]}`
   - Spread operator: `{for animal of ["🦄", ...animals]}`

3. **Interactive Notifications** (Lines 166-186):
   - Loop over notification types: success, info, warning, error
   - Reactive state: `onclick="{currentNotification = notif}"`
   - Conditional component rendering: `{if currentNotification}`

4. **Advanced String/Array Operations**:
   - String manipulation: `animal.split('').reverse().join('')`
   - Array filter in onclick: `animals.filter(a => a !== animal)`
   - Array spread in state update: `[newAnimal, ...animals]`

5. **Interactive State Management**:
   - Add/remove items from arrays
   - x-model for input binding
   - onclick handlers with complex expressions

### comprehensive-simple.html - Core Features Showcase
**URL**: http://localhost:3333/comprehensive-simple

**Features Tested**:
1. Basic expressions (variables, object properties, function calls)
2. Conditionals (if/else, if/else-if/else, nested)
3. Array loops (simple, with index, nested)
4. Static components with props
5. Functions in fence section (getGreeting, formatPrice)
6. **Reactive authentication demo** (login/logout with state updates)
7. Alpine.js directives (@click, :disabled, :class, x-text)
8. Component style aggregation

**Missing** (available in home.html or comprehensive.html):
- Object loops (`{for key, value of object}`)
- Dynamic components (`<="path" />`)
- Computed values (`const filtered = ...`)
- Complex array methods (reduce, chained filter/map)
- Passing functions as props to components
- Array spread operator

### comprehensive.html - Original Full Test
**URL**: http://localhost:3333/comprehensive

**Additional Features**:
1. **Object property loops** (Lines 243-257):
   - `{for key, value of settings}`
   - Nested object loops: `{for subKey, subValue of value}`
   - Type checking: `{if typeof value === 'object'}`

2. **Categories with filtered data** (Lines 264-286):
   - Triple-nested loops: categories → items → tags
   - Inline filtering: `categories.filter(p => p.tags.includes("..."))`

3. **Complex expressions** (Lines 355-358):
   - `products.reduce((sum, p) => sum + p.price, 0) / products.length`
   - `products.filter(p => p.featured).length`
   - Chained methods: `products.filter(p => p.inStock).reduce(...)`

4. **Math functions** (Line 366):
   - `Math.min(3, filteredProducts.length)`

5. **Array slicing in loops** (Line 368):
   - `{for product, index in filteredProducts.slice(0, 3)}`

6. **Function props to components** (Line 309):
   - `<ProductCard formatPrice={formatPrice} />`

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

## Recommendations for comprehensive-simple.html Enhancement

To make comprehensive-simple.html a more complete showcase, consider adding:

### Priority 1: Core Missing Features
1. **Section 5: Object Loops**
   - Add: `{for key, value of settings}` example
   - Showcase nested object iteration
   - Test: Verify Object.entries() transformation works

2. **Section 6: Dynamic Components**
   - Add: `<="./components/AdminPanel.html" user={user} />` conditional example
   - Add: Variable path example: `<='{componentPath}'`
   - Test: Verify component loading based on conditions

3. **Section 7: Computed Values**
   - Add fence section: `const filtered = products.filter(...)`
   - Show computed values used in multiple places
   - Test: Verify const values in x-data scope

### Priority 2: Advanced Showcases
4. **Section 8: Complex Array Methods**
   - Average price with reduce
   - Chained filter().map() operations
   - Array slicing in loops: `.slice(0, 3)`

5. **Enhanced Component Section**
   - Pass `formatPrice` function as prop to ProductCard
   - Test function props work correctly
   - Verify function scope in child components

### Priority 3: Edge Cases
6. **Math Functions**: `Math.min()`, `Math.max()`, `Math.floor()`
7. **Type Checking**: `{if typeof value === 'object'}`
8. **Array Spread**: `{for item of ["new", ...existingArray]}`

### Files to Reference
- **home.html**: For dynamic components, object loops, array spread examples
- **comprehensive.html**: For object loops, computed values, complex expressions

---

## Next Steps

After completing this checklist:

1. ✅ **Update checklist**: Document current feature coverage (DONE 2025-10-07)
2. **Enhance comprehensive-simple.html**: Add missing showcase sections (optional)
3. **Test comprehensive.html**: Verify all original patterns still work
4. **Update roadmap**: Mark template engine features as complete
5. **Prepare for Spec 8**: Plenti Integration ready to start
