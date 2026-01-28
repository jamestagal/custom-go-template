# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-07-global-store-system/spec.md

> Created: 2025-10-07
> Version: 1.0.0

## Technical Requirements

### 1. Parser Changes (`parser/`)

**New Store Expression Parser**
- Detect `$storeName.property` syntax in expressions
- Create new AST node type: `StoreExpressionNode`
- Pattern: `$[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_.]*`

```go
// ast/store.go
type StoreExpressionNode struct {
    StoreName string   // "auth"
    Property  string   // "loggedIn" or "user.name" (dot-notation)
    Line      int
    Column    int
}
```

**Fence Section Store Parser**
- Extend fence section parser to recognize `store storeName = { ... }` syntax
- Parse store name and JSON/JS object literal
- Store declarations added to `FenceSection.Stores map[string]string`

```go
// Example fence section parsing
---
import Button from './components/Button.html'
store auth = {
    loggedIn: false,
    user: null
}
let message = "Hello"
---
```

**Integration with Expression Parser**
- Add `$` prefix detection to `parseExpression()` in `parser/expressions.go`
- Route store expressions through new `parseStoreExpression()` function
- Maintain existing variable expression logic for non-store expressions

### 2. AST Changes (`ast/`)

**New Node Types**

```go
// ast/store.go
type StoreExpressionNode struct {
    StoreName string
    Property  string
    Line      int
    Column    int
}

func (n *StoreExpressionNode) NodeType() string {
    return "StoreExpression"
}

func (n *StoreExpressionNode) String() string {
    return fmt.Sprintf("$%s.%s", n.StoreName, n.Property)
}
```

**FenceSection Extension**

```go
// ast/fence.go - Add to existing FenceSection struct
type FenceSection struct {
    Imports   []Import
    Props     []Prop
    Variables []Variable
    Functions []Function
    Stores    map[string]string  // NEW: storeName -> JSON object string
}
```

### 3. Transformer Changes (`transformer/`)

**Store Expression Transformer**
- Create `transformer/stores.go`
- Transform `StoreExpressionNode` to Alpine.js `$store` reference
- Generate wrapper element with `x-text` directive

```go
// transformer/stores.go
func (t *Transformer) transformStoreExpression(node *ast.StoreExpressionNode) ast.Node {
    // {$auth.loggedIn} → <span x-text="$store.auth.loggedIn"></span>
    return &ast.Element{
        Tag: "span",
        Attributes: []ast.Attribute{
            {
                Name:  "x-text",
                Value: fmt.Sprintf("$store.%s.%s", node.StoreName, node.Property),
            },
        },
    }
}
```

**Store Initialization Tracking**
- Track all store references during transformation
- Build map of `storeName -> storeDefinition`
- Store definitions come from fence section `Stores` map or external imports

**Integration with Transform Pipeline**
- Add store expression case to `Transform()` switch in `transformer/transformer.go`
- Call `transformStoreExpression()` for `StoreExpressionNode` types

### 4. Renderer Changes (`renderer/`)

**Store Initialization Script Generation**
- Create `renderer/stores.go`
- Generate `Alpine.store()` initialization calls
- Insert before Alpine.start() or in document ready

```go
// renderer/stores.go
func (r *Renderer) renderStoreInitializations(stores map[string]string) string {
    var sb strings.Builder
    sb.WriteString("<script>\n")
    sb.WriteString("document.addEventListener('alpine:init', () => {\n")

    for name, definition := range stores {
        sb.WriteString(fmt.Sprintf("  Alpine.store('%s', %s);\n", name, definition))
    }

    sb.WriteString("});\n")
    sb.WriteString("</script>\n")
    return sb.String()
}
```

**HTML Output Structure**

```html
<!DOCTYPE html>
<html>
<head>
    <script src="//unpkg.com/alpinejs" defer></script>
</head>
<body>
    <!-- Store initialization BEFORE Alpine loads -->
    <script>
    document.addEventListener('alpine:init', () => {
        Alpine.store('auth', {
            loggedIn: false,
            user: null
        });
    });
    </script>

    <!-- Component content with store references -->
    <div x-data="{ message: 'Hello' }">
        <span x-text="$store.auth.loggedIn"></span>
    </div>
</body>
</html>
```

### 5. Server Changes (`cmd/server/`)

**Store File Discovery**
- Scan `stores/` directory for `.js` files on server startup
- Parse store files to extract store name and definition
- Register stores in global store registry

```go
// cmd/server/main.go - Add to existing setup
func registerStores(storeDir string) map[string]string {
    stores := make(map[string]string)
    files, _ := os.ReadDir(storeDir)

    for _, file := range files {
        if filepath.Ext(file.Name()) == ".js" {
            content, _ := os.ReadFile(filepath.Join(storeDir, file.Name()))
            storeName := strings.TrimSuffix(file.Name(), ".js")
            stores[storeName] = extractStoreDefinition(content)
        }
    }

    return stores
}
```

**Store Import Resolution**
- Extend existing import resolution to handle store imports
- Pattern: `import store from './stores/storeName.js'`
- Load external store definition and merge with inline stores

**Request Handling**
- Merge inline stores (from fence section) with external stores
- Pass combined store map to renderer
- Inline stores override external stores with same name

### 6. Store File Format

**External Store File Structure**

```javascript
// stores/auth.js
{
    loggedIn: false,
    user: null,
    login(userData) {
        this.loggedIn = true;
        this.user = userData;
    },
    logout() {
        this.loggedIn = false;
        this.user = null;
    }
}
```

**Store Import Syntax in Templates**

```html
---
import store from './stores/auth.js'
import store from './stores/cart.js'
---

<div>
    {if $auth.loggedIn}
        <p>Welcome, {$auth.user.name}!</p>
    {else}
        <p>Please log in</p>
    {/if}

    <p>Cart items: {$cart.itemCount}</p>
</div>
```

### 7. Syntax Examples

**Basic Store Reference**
```
Input:  {$auth.loggedIn}
Output: <span x-text="$store.auth.loggedIn"></span>
```

**Nested Property Access**
```
Input:  {$auth.user.name}
Output: <span x-text="$store.auth.user.name"></span>
```

**Store in Conditional**
```
Input:
{if $auth.loggedIn}
    <p>Welcome!</p>
{/if}

Output:
<template x-if="$store.auth.loggedIn">
    <p>Welcome!</p>
</template>
```

**Store in Loop**
```
Input:
{for item in $cart.items}
    <div>{item.name}</div>
{/for}

Output:
<template x-for="item in $store.cart.items">
    <div>
        <span x-text="item.name"></span>
    </div>
</template>
```

**Store in Attribute**
```
Input:  <div class="{$theme.mode}">
Output: <div :class="$store.theme.mode">
```

**Inline Store Definition**
```
Input:
---
store auth = {
    loggedIn: false,
    user: null
}
---

Output (in rendered HTML):
<script>
document.addEventListener('alpine:init', () => {
    Alpine.store('auth', {
        loggedIn: false,
        user: null
    });
});
</script>
```

### 8. Integration with Existing Systems

**Props vs Stores Distinction**
- **Props**: Content data passed from Plenti's layout system (page title, content, frontmatter)
- **Stores**: Client-side application state (auth, cart, theme, UI state)
- Props are component-scoped, stores are global
- Props are immutable from template perspective, stores are mutable

**Plenti Layout Integration**
```go
// Plenti passes content via props
<BlogPost
    title={content.title}
    author={content.author}
/>

// Component uses both props and stores
---
prop title = ""
prop author = ""
store theme = { mode: "light" }
---

<article class="{$theme.mode}">
    <h1>{title}</h1>
    <p>By {author}</p>
</article>
```

**Fence Section Priority**
1. Imports processed first (components and stores)
2. Props extracted (for component interface)
3. Stores merged (inline overrides external)
4. Variables and functions added to x-data scope

**Scope Management**
- Store references use Alpine's `$store` magic property
- Regular variables remain in local x-data scope
- No collision between `$store.auth` and local `auth` variable

### 9. Error Handling

**Parser Errors**
- Invalid store syntax: `{$123invalid}` → "Store name must start with letter or underscore"
- Missing property: `{$auth}` → "Store reference must include property access"
- Invalid store definition: `store auth = invalid` → "Store definition must be valid JSON object"

**Runtime Validation**
- Undefined store reference: `{$nonexistent.prop}` → Warning in console, graceful degradation
- Store initialization order: External stores loaded before inline stores
- Circular dependencies: Not applicable (stores are data-only in v1)

**Development Feedback**
- Server logs all registered stores at startup
- Template parsing errors include line/column information
- Clear distinction between parser errors and transformation errors

## Approach

### Phase 1: Parser Foundation (Week 1)
1. Create `StoreExpressionNode` AST type
2. Add `$` prefix detection to expression parser
3. Parse `store name = {}` syntax in fence sections
4. Write parser tests for all syntax variations

### Phase 2: Transformation (Week 1-2)
1. Create store expression transformer
2. Generate `x-text` wrappers for store references
3. Handle stores in conditionals, loops, attributes
4. Track all store references during transformation

### Phase 3: Rendering & Server (Week 2)
1. Generate Alpine.store() initialization scripts
2. Add store file discovery in server
3. Merge inline and external store definitions
4. Render stores before Alpine.js initialization

### Phase 4: Integration Testing (Week 2-3)
1. Test cross-component reactivity
2. Validate props vs stores separation
3. Test nested property access
4. Integration with existing conditional/loop tests

### Phase 5: Documentation & Examples (Week 3)
1. Add store examples to `examples/components/`
2. Document syntax in project README
3. Create example store files in `stores/`
4. Update CLAUDE.md with store system details

## External Dependencies

**No new external dependencies required.**

The implementation uses existing dependencies:
- Go 1.22.2 standard library for parsing and file operations
- Alpine.js 3.x (already integrated) provides `$store` magic property and `Alpine.store()` API
- Existing parser combinator framework in `parser/` package

**Why no new dependencies:**
- Store syntax parsing extends existing expression parser
- Alpine.js built-in store functionality handles all client-side reactivity
- File system operations use Go stdlib `os` and `path/filepath`
- JSON parsing uses Go stdlib `encoding/json`

**Future considerations (out of scope for v1):**
- Alpine.js persist plugin for localStorage (optional user addition)
- Alpine.js morph plugin for server-state sync (if needed later)
