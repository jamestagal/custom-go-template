# Global Store System - Lite Summary

Add Alpine.js store integration using `$storeName.prop` syntax and `store name = {}` declarations to enable cross-component reactive state sharing for client-side application state (auth, cart, theme) while maintaining the existing props system for content data.

## Key Points
- Inline `store auth = { loggedIn: false }` syntax in fence sections defines global stores
- `{$auth.loggedIn}` transforms to `<span x-text="$store.auth.loggedIn">`
- External store files supported via `import store from './stores/name.js'`
- Complements props system: props = content data, stores = application state
- Pure client-side reactivity using Alpine.js built-in store functionality
