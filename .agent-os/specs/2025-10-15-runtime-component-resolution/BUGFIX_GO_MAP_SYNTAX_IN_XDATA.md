# BUGFIX: Go Map Syntax Appearing in x-data Attributes

**Status**: 🔴 OPEN - Requires deep investigation
**Priority**: HIGH
**Date**: 2025-10-15
**Affects**: Runtime component resolution, complex prop passing

---

## Problem Statement

When complex Go objects (maps, slices) are passed as props to dynamic components, they are being rendered with **Go map syntax** instead of proper **JSON/JavaScript syntax** in the HTML output.

### Symptoms

**Expected Output**:
```html
<div x-data="{ content: {components: [{fields: {...}, name: 'hero'}]} }">
```

**Actual Output**:
```html
<div x-data="{ content: map[components:[map[fields:map[...] name:hero]]] }">
```

This causes Alpine.js to throw syntax errors:
- `Uncaught SyntaxError: Unexpected token ':'`
- Alpine.js fails to initialize
- `$renderDynamicComponent` magic function doesn't register
- Dynamic components fail to render

---

## Context

### How It Happens

1. **Server** passes `content` object (Go map) to html.html wrapper:
   ```go
   props := map[string]interface{}{
       "content": contentData, // contentData is map[string]interface{}
   }
   ```

2. **html.html** receives and passes it to layout:
   ```html
   <Component:dynamic name={layout}
       content={content}  <!-- content is a Go map -->
       {...content.fields} />
   ```

3. **Transformer** processes `content={content}` prop:
   - Parser extracts: `ComponentProp{Name: "content", Value: "{content}", IsDynamic: true}`
   - `extractPropValue()` resolves to actual value from dataScope
   - Returns: `map[string]interface{}{...}` (the actual Go map)

4. **Somewhere between extraction and rendering**, the map gets stringified:
   - `map[string]interface{}` → `"map[components:[map[...]]"` (string)
   - This happens BEFORE it reaches formatting functions

5. **formatComponentData()** receives a STRING, not a map:
   - Type switch sees `case string:` instead of `default:`
   - Never calls `FormatGoValueToJS()` which would JSON-encode it
   - Outputs the Go map syntax as-is

---

## Investigation History

### Attempted Fixes

1. **✅ Modified `FormatGoValueToJS` default case** (transformer/alpine.go:452-468)
   - Added JSON marshaling instead of `fmt.Sprintf("%v")`
   - **Result**: Fix is correct but never reached (value already stringified)

2. **✅ Added debug logging to `formatComponentData`** (transformer/components.go:272)
   - Logs should show type of each value
   - **Result**: Logs not appearing (function may not be called for this case)

3. **🔍 Investigated prop resolution pipeline**:
   - `extractPropValue()` → returns `any` (can be map)
   - `propScope[prop.Name] = extractPropValue(...)` → stores as `any`
   - `formatComponentData(dataScope)` → should receive map as `any`
   - **Gap**: Where does map → string conversion happen?

### Key Files Examined

- `ast/ast.go:199-206` - `ComponentProp` struct (Value is string field)
- `transformer/components.go:622-669` - `extractPropValue()`
- `transformer/components.go:673-681` - `transformComponentProps()`
- `transformer/components.go:260-414` - `formatComponentData()`
- `transformer/alpine.go:295-469` - `FormatGoValueToJS()`

---

## Root Cause Hypothesis

The issue likely occurs in one of these locations:

### Hypothesis 1: ComponentProp.Value String Assignment
**Location**: Somewhere props are converted from `any` back to `ComponentProp`

When a map value is assigned to `ComponentProp.Value` (which is a `string` field), Go automatically calls `fmt.Sprintf("%v")` to convert it, producing `"map[...]"`.

**Evidence**:
- `ComponentProp.Value` is typed as `string` (ast/ast.go:201)
- If a map is assigned: `prop.Value = mapValue`, Go stringifies it

**Where to look**:
- Search for code that creates ComponentProp structs from resolved values
- Look for `ComponentProp{...Value: someMapValue...}` patterns
- Check `convertPropsMapToComponentProps()` in dynamic_component_by_name.go

### Hypothesis 2: Template Rendering Layer
**Location**: renderer/ package may stringify values when generating HTML

The renderer might be converting all values to strings before inserting them into HTML, bypassing the transformer's formatting logic.

**Where to look**:
- `renderer/render.go` - attribute rendering
- `renderer/component.go` - component rendering
- Any code that converts dataScope to HTML attributes

### Hypothesis 3: Attribute Value Rendering
**Location**: AST attribute rendering in renderer

When rendering `<div x-data="...">`, the Value field might be processed differently than expected.

**Where to look**:
- How `ast.Attribute.Value` is rendered to HTML
- Whether Value is expected to be pre-formatted or gets formatted during render
- Check if there's a separate path for x-data vs regular attributes

---

## Debugging Strategy for Next Investigation

### Step 1: Add Comprehensive Logging

Add logging at EVERY step where the content value is touched:

```go
// In extractPropValue (transformer/components.go:655)
if prop.Name == "content" {
    log.Printf("[CONTENT DEBUG] extractPropValue: returning type=%T, value=%#v", value, value)
}

// In transformComponentProps (transformer/components.go:677)
if prop.Name == "content" {
    extracted := extractPropValue(prop, parentDataScope)
    log.Printf("[CONTENT DEBUG] transformComponentProps: extracted type=%T", extracted)
    propScope[prop.Name] = extracted
}

// In formatComponentData (transformer/components.go:270)
if key == "content" {
    log.Printf("[CONTENT DEBUG] formatComponentData: type=%T, isString=%v", value, _, ok := value.(string); ok)
}

// In renderer (wherever x-data is rendered)
log.Printf("[CONTENT DEBUG] rendering x-data attribute: %s", attribute.Value)
```

### Step 2: Trace the Full Pipeline

Create a test case that:
1. Creates a component with `content={content}` where content is a map
2. Logs at every transformation step
3. Captures the exact point where `map[...]` string appears

### Step 3: Check Alternative Code Paths

The issue might be in a code path we haven't examined:

- **Dynamic component by name** - Different handling for runtime components
- **Loop context** - Special handling inside x-for loops
- **Spread props** - `{...content.fields}` might interfere with `content={content}`

### Step 4: Check for String Coercion

Search for these patterns in the codebase:

```bash
# Find where maps might be assigned to string fields
grep -rn "\.Value = " transformer/ | grep -v "string"

# Find fmt.Sprintf with maps
grep -rn "fmt.Sprintf.*%v" transformer/ renderer/

# Find string type assertions
grep -rn "value.(string)" transformer/ renderer/
```

---

## Workaround (Temporary)

Until this is fixed, avoid passing complex nested objects as props to dynamic components:

**Instead of**:
```html
<Component:dynamic name={layout} content={content} />
```

**Use**:
```html
<Component:dynamic name={layout} {...content} />
```

This spreads the content fields directly into props, avoiding the nested map issue.

**Limitation**: This only works if the component doesn't need the `content` object itself (e.g., to access `content.components`).

---

## Expected Fix

Once the stringification point is found, the fix should:

1. **Preserve type information** through the pipeline
   - Keep maps as `map[string]interface{}` until final formatting
   - Don't convert to string until `formatComponentData()` or `FormatGoValueToJS()`

2. **Use proper JSON encoding** for complex types
   - Call `FormatGoValueToJS()` for all non-string types
   - Ensure it reaches the `default:` case which JSON-encodes

3. **Update tests** to verify complex prop passing:
   ```go
   func TestComplexPropSerialization(t *testing.T) {
       content := map[string]interface{}{
           "components": []map[string]interface{}{
               {"name": "Hero", "fields": map[string]interface{}{"title": "Test"}},
           },
       }
       // Verify JSON output, not map[...] syntax
   }
   ```

---

## Related Issues

- ✅ **BUGFIX_CONTENT_NULL_ERROR.md** - Fixed by passing `content={content}` explicitly
- ⚠️ **Runtime Component Resolution** - This blocks proper runtime component rendering
- ⚠️ **Alpine.js Integration** - Syntax errors prevent Alpine from initializing

---

## Next Steps for go-backend Agent

1. **Add comprehensive logging** at every step of prop resolution pipeline
2. **Run server and capture logs** showing exact point where map→string conversion occurs
3. **Identify the code location** responsible for stringification
4. **Implement fix** to preserve type information until final formatting
5. **Add regression tests** for complex prop types (maps, slices, nested objects)
6. **Update documentation** on prop passing best practices

---

## Console Errors (For Reference)

```
cdn.min.js:5 Uncaught SyntaxError: Unexpected token ':'
[Alpine] content: map[components:[map[fields:map[buttonLink:/contact ...
```

```
cdn.min.js:5 Uncaught ReferenceError: allContent is not defined
[Alpine] {compName: component.name, compProps: {content: content, allContent: allContent}}
```

```
cdn.min.js:5 Uncaught ReferenceError: $renderDynamicComponent is not defined
[Alpine] $renderDynamicComponent($el, compName, compProps)
```

**Note**: The last two errors are CAUSED by the first - Alpine.js fails to initialize due to syntax error, so `allContent` and `$renderDynamicComponent` are never defined.

---

## Cognitive Load Assessment

**Current Investigation**: 8 (requires deep understanding of 4 packages)
**Expected Fix**: 4-6 (once location is identified, fix should be straightforward)
**Testing**: 5 (need comprehensive tests for complex prop types)

**Total Load for Fix**: ~15-19 (manageable within guidelines)
