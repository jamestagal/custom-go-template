# Store Tracking Flow - Before and After Fix

## Before Fix (BROKEN)

```
Template Source:
┌─────────────────────────────────────────────────┐
│ {$theme.mode}                                   │ ← Template syntax
│ @click="$store.theme.setLight()"                │ ← Alpine syntax
└─────────────────────────────────────────────────┘
                    ↓
        transformAttributesWithStores()
                    ↓
┌─────────────────────────────────────────────────┐
│ Process {$theme.mode}:                          │
│   - Match storeAttrPattern ✓                    │
│   - Extract "theme" ✓                           │
│   - TrackStoreReference("theme") ✓              │
│   - Transform to $store.theme.mode ✓            │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ Process @click="$store.theme.setLight()":       │
│   - IsAlpine = true                             │
│   - Skip without tracking ✗                     │
│   - Pass through unchanged                      │
└─────────────────────────────────────────────────┘
                    ↓
        GetTrackedStores()
                    ↓
┌─────────────────────────────────────────────────┐
│ Referenced stores: [auth, cart]                 │
│ ❌ theme NOT tracked (only used in @click)      │
└─────────────────────────────────────────────────┘
                    ↓
        Alpine.js Initialization
                    ↓
┌─────────────────────────────────────────────────┐
│ Alpine.store('auth', {...})   ✓                 │
│ Alpine.store('cart', {...})   ✓                 │
│ Alpine.store('theme', {...})  ✗ MISSING!        │
└─────────────────────────────────────────────────┘
                    ↓
        Browser Runtime
                    ↓
┌─────────────────────────────────────────────────┐
│ User clicks theme button                        │
│ → @click="$store.theme.setLight()" executes     │
│ → $store.theme is undefined                     │
│ → ❌ Error: Cannot read 'setLight' of undefined │
└─────────────────────────────────────────────────┘
```

## After Fix (WORKING)

```
Template Source:
┌─────────────────────────────────────────────────┐
│ {$theme.mode}                                   │ ← Template syntax
│ @click="$store.theme.setLight()"                │ ← Alpine syntax
└─────────────────────────────────────────────────┘
                    ↓
        transformAttributesWithStores()
                    ↓
┌─────────────────────────────────────────────────┐
│ Process {$theme.mode}:                          │
│   - Match storeAttrPattern ✓                    │
│   - Extract "theme" ✓                           │
│   - TrackStoreReference("theme") ✓              │
│   - Transform to $store.theme.mode ✓            │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ Process @click="$store.theme.setLight()":       │
│   - IsAlpine = true                             │
│   - Contains "$store." = true                   │
│   - ✓ NEW: trackAlpineStoreReferences()         │
│     - Match alpineStorePattern                  │
│     - Extract "theme"                           │
│     - TrackStoreReference("theme") ✓            │
│   - Pass through unchanged                      │
└─────────────────────────────────────────────────┘
                    ↓
        GetTrackedStores()
                    ↓
┌─────────────────────────────────────────────────┐
│ Referenced stores: [auth, cart, theme]          │
│ ✓ theme tracked from @click handler!            │
└─────────────────────────────────────────────────┘
                    ↓
        Alpine.js Initialization
                    ↓
┌─────────────────────────────────────────────────┐
│ Alpine.store('auth', {...})   ✓                 │
│ Alpine.store('cart', {...})   ✓                 │
│ Alpine.store('theme', {...})  ✓ PRESENT!        │
└─────────────────────────────────────────────────┘
                    ↓
        Browser Runtime
                    ↓
┌─────────────────────────────────────────────────┐
│ User clicks theme button                        │
│ → @click="$store.theme.setLight()" executes     │
│ → $store.theme exists                           │
│ → ✓ setLight() method called successfully       │
│ → ✓ Theme changes to light mode                 │
└─────────────────────────────────────────────────┘
```

## Key Insight

The fix recognizes that store references come in **two forms**:

1. **Template syntax**: `{$storeName.property}`
   - Needs transformation AND tracking
   - Original logic handled this ✓

2. **Alpine syntax**: `$store.storeName.property`
   - Already transformed, needs ONLY tracking
   - Original logic missed this ✗
   - **Fix added this** ✓

## Code Comparison

### Before (Missing Tracking)
```go
if attr.IsAlpine {
    transformedAttributes = append(transformedAttributes, attr)
    continue  // ← Skips without tracking!
}
```

### After (Tracks Before Skipping)
```go
if attr.IsAlpine && strings.Contains(attr.Value, "$store.") {
    trackAlpineStoreReferences(attr.Value)  // ← TRACKS FIRST!
    transformedAttributes = append(transformedAttributes, attr)
    continue
}
```

## Pattern Coverage

The fix now handles ALL these patterns:

```javascript
// Method calls
@click="$store.theme.setLight()"       ✓
@click="$store.cart.clear()"           ✓
@click="$store.auth.login()"           ✓

// Property access
x-show="$store.auth.isLoggedIn"        ✓
x-text="$store.theme.mode"             ✓
:class="$store.theme.isDark"           ✓

// Complex expressions
x-if="$store.auth.user && $store.cart.total > 0"  ✓
@click="count + $store.cart.itemCount"            ✓

// Multiple stores in one expression
x-show="$store.auth.isAdmin || $store.user.role === 'owner'"  ✓
```

All patterns now properly track their store references!
