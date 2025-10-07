# Technical Spec: Fix Server x-data Building Architecture

**Date**: 2025-10-07
**Parent Spec**: `../spec.md`

## Overview

This technical spec provides detailed implementation guidance for removing manual x-data building from the server and using the proper rendering pipeline.

## Current Implementation Analysis

### Problem Code Locations

The server has **3 route handlers** with identical broken patterns:

1. **`/` handler** (lines 35-214)
2. **`/comprehensive-simple` handler** (lines 217-315)
3. **`/comprehensive` handler** (lines 318-425)

Each handler contains:

#### Manual Fence Data Extraction (Lines 60-120, 238-263, 340-372)

```go
// PROBLEM: Manual regex-based function extraction
functionRegex := regexp.MustCompile(`function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{[^}]*}`)
matches := functionRegex.FindAllStringSubmatch(fence.RawContent, -1)
for _, match := range matches {
    if len(match) >= 2 {
        props[match[1]] = match[0]  // Adds function as STRING
    }
}
```

**Issues**:
- Regex `{[^}]*}` stops at first `}`, truncating multi-line functions
- Can't handle nested braces: `if (x) { return { value: 1 }; }`
- Can't handle template literals: `` `${variable}` ``
- Arrow functions have separate buggy regex (lines 94-102)
- Method shorthand has another regex (lines 104-116)

#### Manual x-data Building (Lines 150-182, 273-294, 383-404)

```go
// PROBLEM: Manual JSON marshaling
propsJSON, err := json.Marshal(props)
if err != nil {
    propsJSON = []byte("{}")
}
alpineDataAttr := string(propsJSON)

// PROBLEM: Manual injection into body tag with regex
bodyTagRegex := regexp.MustCompile(`(?i)<body[^>]*>`)
htmlWithLinks = bodyTagRegex.ReplaceAllStringFunc(htmlWithLinks, func(match string) string {
    tagWithoutClose := strings.TrimSuffix(match, ">")
    return fmt.Sprintf(`%s x-data='%s'>`, tagWithoutClose, alpineDataAttr)
})
```

**Issues**:
- `json.Marshal(props)` converts functions to quoted strings: `"formatPrice": "function formatPrice(price) { return..."`
- Alpine.js receives invalid x-data: functions as strings, not methods
- Bypasses transformer's `alpineDataFormatter` which handles this correctly

### Correct Implementation (Already Exists)

The transformer already handles this correctly in `transformer/alpine.go`:

```go
// transformer/alpine.go (lines ~50-80)
func alpineDataFormatter(data map[string]interface{}) string {
    var parts []string
    for key, value := range data {
        switch v := value.(type) {
        case string:
            if strings.HasPrefix(v, "function ") {
                // Convert "function name(params) { body }" to "name(params) { body }"
                funcBody := strings.TrimPrefix(v, "function "+key)
                parts = append(parts, fmt.Sprintf("%s%s", key, funcBody))
            } else if strings.Contains(v, "=>") {
                // Arrow function: keep as-is
                parts = append(parts, fmt.Sprintf("%s: %s", key, v))
            } else {
                // Regular string value
                parts = append(parts, fmt.Sprintf("%s: %s", key, strconv.Quote(v)))
            }
        case int, int64, float64, bool:
            parts = append(parts, fmt.Sprintf("%s: %v", key, v))
        case nil:
            parts = append(parts, fmt.Sprintf("%s: null", key))
        default:
            // Arrays, objects - convert to JSON
            jsonBytes, _ := json.Marshal(v)
            parts = append(parts, fmt.Sprintf("%s: %s", key, string(jsonBytes)))
        }
    }
    return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
```

**Key Features**:
- Detects `function` keyword and converts to method shorthand
- Handles arrow functions correctly
- Properly quotes string values
- Preserves numeric/boolean types
- Marshals complex objects as JSON

## Implementation Plan

### Phase 1: Refactor Server Route Handlers

#### Step 1.1: Create Unified Route Handler Function

**New Function** (add after `main()` function, around line 435):

```go
// renderTemplate is a unified handler for rendering template files
// It replaces the manual x-data building logic with proper renderer.Render() usage
func renderTemplate(entrypoint string, w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()

    // Read template file
    templateContent, err := os.ReadFile(entrypoint)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to read template: %v", err), http.StatusInternalServerError)
        return
    }

    // Parse template to extract fence data
    template, err := parser.ParseTemplate(string(templateContent))
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
        return
    }

    // Extract initial props from fence section (for buildTime)
    props := make(map[string]interface{})
    for _, node := range template.RootNodes {
        if fence, ok := node.(*ast.FenceSection); ok {
            // Process variables
            for _, variable := range fence.Variables {
                props[variable.Name] = parseValue(variable.Value)
            }

            // Process props with default values
            for _, prop := range fence.Props {
                if _, exists := props[prop.Name]; !exists && prop.DefaultValue != "" {
                    props[prop.Name] = parseValue(prop.DefaultValue)
                }
            }
            break
        }
    }

    // Add build time as a prop
    buildTime := time.Since(startTime)
    buildTimeMs := float64(buildTime.Microseconds()) / 1000.0
    props["buildTime"] = fmt.Sprintf("%.2fms", buildTimeMs)

    // CRITICAL: Use renderer.Render() - this calls the transformer
    // which uses alpineDataFormatter for correct x-data generation
    markup, script, style := renderer.Render(entrypoint, props)

    // Build final HTML with inline styles and scripts
    finalHTML := markup

    // Inject styles into <head>
    if style != "" {
        headEndRegex := regexp.MustCompile(`(?i)</head>`)
        finalHTML = headEndRegex.ReplaceAllString(finalHTML, fmt.Sprintf("<style>\n%s\n</style></head>", style))
    }

    // Inject scripts before </body>
    if script != "" {
        bodyEndRegex := regexp.MustCompile(`(?i)</body>`)
        finalHTML = bodyEndRegex.ReplaceAllString(finalHTML, fmt.Sprintf("<script>\n%s\n</script></body>", script))
    }

    // Add Alpine.js CDN if not already present
    if !strings.Contains(finalHTML, "alpinejs") {
        headEndRegex := regexp.MustCompile(`(?i)</head>`)
        finalHTML = headEndRegex.ReplaceAllString(finalHTML,
            `<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script></head>`)
    }

    // Add build time comment
    totalBuildTime := time.Since(startTime)
    htmlComment := fmt.Sprintf("<!-- Build time: %v -->\n", totalBuildTime)
    finalHTML = htmlComment + finalHTML

    // Send response
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(finalHTML))
}
```

**Key Changes**:
- ❌ **REMOVED**: Manual regex extraction of functions
- ❌ **REMOVED**: Manual `json.Marshal(props)` for x-data
- ❌ **REMOVED**: Manual body tag x-data injection
- ✅ **ADDED**: `renderer.Render(entrypoint, props)` - single source of truth
- ✅ **ADDED**: Proper inline style/script injection
- ✅ **ADDED**: Build time tracking (preserved feature)

#### Step 1.2: Update Route Handlers to Use New Function

**Replace `/` handler** (lines 35-214):

```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // Serve static files from the public directory
    publicDir := "./public"
    if r.URL.Path != "/" {
        http.ServeFile(w, r, publicDir+r.URL.Path)
        return
    }

    // Render home page
    renderTemplate("examples/pages/home.html", w, r)
})
```

**Replace `/comprehensive-simple` handler** (lines 217-315):

```go
http.HandleFunc("/comprehensive-simple", func(w http.ResponseWriter, r *http.Request) {
    renderTemplate("examples/pages/comprehensive-simple.html", w, r)
})
```

**Replace `/comprehensive` handler** (lines 318-425):

```go
http.HandleFunc("/comprehensive", func(w http.ResponseWriter, r *http.Request) {
    renderTemplate("examples/pages/comprehensive.html", w, r)
})
```

**Result**: ~350 lines of manual code removed, ~3 lines per handler added

#### Step 1.3: Remove Obsolete Helper Functions

**Delete or comment out** (lines 522-567):
- `addLinksToHTML()` - No longer needed (Alpine CDN added in renderTemplate)
- `convertJSToJSON()` - No longer needed (transformer handles this)

**Keep**:
- `registerComponents()` (lines 436-473) - Still needed
- `extractComponentProps()` (lines 475-520) - Still needed
- `parseValue()` (lines 569-622) - Still needed for prop parsing

### Phase 2: Verify Transformer Integration

#### Step 2.1: Check renderer.Render() Flow

**File**: `renderer/render.go`

Verify the call chain:
1. `renderer.Render(entrypoint, props)` called
2. Renderer loads template and calls `transformer.Transform(template, props)`
3. Transformer calls `buildDataScope(template, props)` to aggregate all variables
4. Transformer calls `alpineDataFormatter(dataScope)` to generate x-data string
5. Renderer injects x-data into template body (not just body tag)

**Expected**: x-data contains functions as method shorthand, not JSON strings

#### Step 2.2: Add Debug Logging (Optional)

Add to `transformer/alpine.go` (after alpineDataFormatter call):

```go
// Debug: Log generated x-data
log.Printf("[Transformer] Generated x-data: %s", xDataString)
```

This helps verify functions are formatted correctly during development.

### Phase 3: Restore Functions to Test File

#### Step 3.1: Add Functions to comprehensive-simple.html

**File**: `examples/pages/comprehensive-simple.html`

**Location**: Fence section (after line 12, before `---`)

**Add**:

```javascript
// Helper function to get time-appropriate greeting
function getGreeting() {
  const hour = new Date().getHours();
  if (hour < 12) return 'Good morning';
  if (hour < 18) return 'Good afternoon';
  return 'Good evening';
}

// Helper function to format prices
function formatPrice(price) {
  return '$' + price.toFixed(2);
}
```

#### Step 3.2: Use Functions in Template Body

**Update line 105** (in Basic Expressions section):

```html
<!-- BEFORE -->
<p>Hello, {user.name}!</p>

<!-- AFTER -->
<p>{getGreeting()}, {user.name}!</p>
```

**Update line 166** (in Product List loop):

```html
<!-- BEFORE -->
<li>{product.name} - ${product.price}

<!-- AFTER -->
<li>{product.name} - {formatPrice(product.price)}
```

#### Step 3.3: Add Function Call Test Section

**Add after Section 5** (around line 275):

```html
<!-- 6. Function Tests Section -->
<div class="section">
  <h2 class="section-title">6. Function Tests</h2>

  <div class="card">
    <h3>Helper Functions</h3>
    <p>Current greeting: <strong>{getGreeting()}</strong></p>
    <p>Formatted price example: <strong>{formatPrice(999.99)}</strong></p>
    <p>Total value: <strong>{formatPrice(products.reduce((sum, p) => sum + p.price, 0))}</strong></p>
  </div>
</div>
```

**Expected Result**: Functions execute correctly, no console errors

### Phase 4: Integration Testing

#### Step 4.1: Build and Run Server

```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

**Expected Output**:
```
Starting server...
Registering component: Header from examples/components/Header.html
Registering component: Footer from examples/components/Footer.html
...
Server starting on http://localhost:3333
```

#### Step 4.2: Test comprehensive-simple Page

**URL**: http://localhost:3333/comprehensive-simple

**Browser Console Checklist**:
- [ ] Zero errors (no `ReferenceError: formatPrice is not defined`)
- [ ] Zero warnings
- [ ] Alpine.js initialized successfully

**Visual Verification**:
- [ ] Greeting displays correctly: "Good morning, John Doe!" (time-dependent)
- [ ] Prices formatted: "$999.99" (not "999.99")
- [ ] Section 6 displays function test results

**View Page Source** (Ctrl+U):
- [ ] Check for x-data attribute with correct function syntax
- [ ] Verify: `formatPrice(price) { return '$' + price.toFixed(2); }`
- [ ] NOT: `"formatPrice": "function formatPrice(price) { return..."`

#### Step 4.3: Run Test Suite

```bash
# Run all tests
go test ./... -v

# Specifically test transformer
go test ./transformer -run TestAlpineDataFormatter -v

# Test renderer
go test ./renderer -v
```

**Expected**: All tests pass (no regressions)

#### Step 4.4: Performance Verification

**Check build times** (in HTML comment):
```html
<!-- Build time: 2.5ms -->
```

**Expected**: No significant performance regression (should be similar or faster than before)

## Code Review Checklist

Before marking as complete:

- [ ] All 3 route handlers use `renderTemplate()` function
- [ ] No manual regex extraction of functions in server code
- [ ] No manual `json.Marshal()` for x-data in server code
- [ ] `renderer.Render()` called for all template rendering
- [ ] Functions added to comprehensive-simple.html fence section
- [ ] Functions used in at least 2 places in template body
- [ ] Browser console shows zero errors
- [ ] x-data contains functions as method shorthand (not JSON strings)
- [ ] All existing features work (props, variables, conditionals, loops, components)
- [ ] All tests pass
- [ ] Build time comparable to before

## Rollback Plan

If issues occur:

1. **Immediate Rollback**: Revert `cmd/server/main.go` to previous version
2. **Partial Rollback**: Keep new `renderTemplate()` but remove function tests
3. **Debug Mode**: Add debug logging to transformer to see x-data generation

Backup current version before changes:
```bash
cp cmd/server/main.go cmd/server/main.go.backup
```

## Performance Considerations

### Expected Performance Impact

- **Positive**: Removing manual regex parsing reduces overhead
- **Neutral**: `renderer.Render()` already exists and is optimized
- **Positive**: Single code path easier to cache/optimize

### Monitoring

Add timing logs in `renderTemplate()`:
```go
log.Printf("[Server] Template %s rendered in %v", entrypoint, time.Since(startTime))
```

## Security Considerations

### Removed Attack Vectors

- **Regex Injection**: Manual regex patterns no longer exposed to user input
- **Template Injection**: All parsing now goes through proper parser (already secure)

### Maintained Security

- File path validation in route handlers (keep `r.URL.Path` checks)
- Component registration still validates file extensions
- No eval() or dangerous JavaScript generation

## Future Enhancements (Out of Scope)

After this fix:
- Bug #2: Fix component prop evaluation in loops
- Bug #3: Fix attribute expression transformation
- Bug #4: Fix multi-line variable extraction in fence parser
- Performance: Add template caching in renderer
- Developer Experience: Hot reload on file changes

## Related Files

- `cmd/server/main.go` - PRIMARY CHANGES HERE
- `transformer/alpine.go` - Verify alpineDataFormatter is called
- `renderer/render.go` - Verify x-data injection
- `examples/pages/comprehensive-simple.html` - Add function tests
- `docs/COMPREHENSIVE_TEST_CHECKLIST.md` - Update Bug #1 status

## Approval

This technical spec is ready for implementation when:
- [ ] Parent spec (`../spec.md`) approved
- [ ] Technical approach reviewed by maintainer
- [ ] Success criteria validated
- [ ] Timeline confirmed (2-4 hours)
