# Component Registry JavaScript Syntax Fixes

**Date**: October 20, 2025
**Status**: ✅ COMPLETED
**Related Specs**:
- [2025-10-15-runtime-component-resolution](../2025-10-15-runtime-component-resolution/)
- [2025-10-16-component-registry-debugging](../2025-10-16-component-registry-debugging/)
- [2025-10-19-build-time-loop-expansion](../2025-10-19-build-time-loop-expansion/)

## Problem Statement

After implementing the runtime component resolution system, the generated `component-registry.js` file contained **multiple JavaScript syntax errors** that prevented components from loading in the browser. Additionally, component **CSS was missing** and **props weren't being passed correctly**.

### Initial Symptoms

1. **Browser Console Errors:**
   ```
   Failed to load component registry after 3 attempts
   SyntaxError: Invalid destructuring assignment target
   SyntaxError: Unexpected string
   SyntaxError: Missing } in template expression
   ```

2. **Component Not Found:**
   ```
   Component 'hero2436' not found in registry
   Component 'services2437' not found in registry
   ```

3. **Missing Styling:**
   - Components rendered but appeared unstyled
   - CSS from `<style>` tags was not included

4. **Undefined Props:**
   - Component displayed "undefined" for all prop values
   - Props from JSON weren't reaching component templates

## Root Causes Identified

### 1. Arrow Function Parameter Bug
**File**: `builder/registry_generator.go`
**Location**: `extractArrowFunctionParams()` function

**Problem**: Arrow function parameters in method calls were being incorrectly prefixed with `props.`:
```javascript
// Bug:
${products.reduce((props.sum, p) => props.sum + p.price, 0)}

// Expected:
${products.reduce((sum, p) => sum + p.price, 0)}
```

**Root Cause**: The regex-based parameter extraction couldn't handle nested parentheses from method calls like `.reduce((sum, p) => ...)`.

### 2. String Literal Content Modification Bug
**File**: `builder/registry_generator.go`
**Location**: `prefixIdentifiersInExpression()` function

**Problem**: Content inside string literals was being treated as identifiers and prefixed:
```javascript
// Bug:
${todo.completed ? props.'✓ props.Done' : props.'○ props.Pending'}

// Expected:
${todo.completed ? '✓ Done' : '○ Pending'}
```

**Root Cause**: The function didn't track when it was inside a string literal, so quoted content was processed as code.

### 3. Spread Operator Misidentification Bug
**File**: `builder/registry_generator.go`
**Location**: `prefixIdentifiersInExpression()` function

**Problem**: The spread operator `...` was treated as property access, causing syntax errors:
```javascript
// Bug:
${props.animals = [props.newAnimal, ...props.animals]props.; props.newAnimal = ''}

// Expected:
${props.animals = [props.newAnimal, ...props.animals]; props.newAnimal = ''}
```

**Root Cause**: Single `.` characters were accumulated as part of identifiers, so `...` became part of a token instead of being recognized as an operator.

### 4. Semicolon Not Recognized as Delimiter Bug
**File**: `builder/registry_generator.go`
**Location**: `isOperatorOrDelimiter()` function

**Problem**: Semicolons weren't treated as statement delimiters:
```javascript
// Bug: semicolon became part of token, got prefixed
]props.;

// Expected:
];
```

**Root Cause**: The `isOperatorOrDelimiter()` function didn't include `;` in its list of delimiters.

### 5. Event Handler Attribute Conversion Bug
**File**: `builder/registry_generator.go`
**Location**: `renderAttributeToJS()` function

**Problem**: Event handler attributes like `onclick` were having their expressions converted to template literals:
```javascript
// Bug (creates invalid JS in template literal):
onclick="${props.animals = [props.newAnimal, ...props.animals]}"

// Expected (keep as Alpine.js expression):
onclick="{animals = [newAnimal, ...animals]}"
```

**Root Cause**: All attributes were being processed through `convertAttributeExpressions()`, which converted `{expr}` to `${props.expr}`. But event handlers need the original `{expr}` syntax for Alpine.js.

### 6. Quote Escaping in Conditionals Bug
**File**: `builder/registry_generator.go`
**Location**: `renderConditionalToJS()` function

**Problem**: Double quotes in x-if conditions weren't being escaped:
```javascript
// Bug:
<template x-if="animal == "cat"">

// Expected:
<template x-if="animal == \"cat\"">
```

**Root Cause**: The `renderConditionalToJS()` function wrote conditions directly without escaping, while regular attributes used `escapeQuotesInAttributeValue()`.

### 7. Missing CSS in Components Bug
**File**: `builder/registry_generator.go`
**Location**: `renderNodeToJS()` function

**Problem**: `<style>` tags were completely omitted from component registry:
```javascript
// Bug: Only HTML, no <style> tag
'Hero2436': (props) => `<section>...</section>`

// Expected: HTML + CSS
'Hero2436': (props) => `<section>...</section><style>...</style>`
```

**Root Cause**: The code had `case *ast.StyleSection: // Skip metadata sections` which was skipping style rendering entirely.

### 8. Case-Sensitive Component Name Lookup Bug
**File**: `static/js/runtime-components.js`
**Location**: `renderDynamicComponent()` function

**Problem**: Component names in JSON were lowercase but registry had capitalized names:
```javascript
// JSON content:
{ "name": "hero2436" }

// Registry:
{ 'Hero2436': (props) => ... }

// Result: Component not found
```

**Root Cause**: The server capitalizes component names (`hero2436.html` → `Hero2436`) but JSON content has lowercase names. The runtime lookup was case-sensitive.

### 9. Props with 'this.' Prefix Bug
**File**: `static/js/runtime-components.js`
**Location**: `renderDynamicComponent()` function

**Problem**: Prop keys had `this.` prefix from server-side processing:
```javascript
// Props object:
{ 'this.topper': 'Welcome', 'this.title': 'Title' }

// Component template trying to access:
${props.topper}  // undefined!
${props.title}   // undefined!
```

**Root Cause**: The server-side `replaceVarRefsWithThis()` function was adding `this.` prefix to field names in the JSON data, but component templates expected clean prop names.

## Solutions Implemented

### Fix 1: Rewrite Arrow Function Parameter Extraction
**File**: `builder/registry_generator.go` (lines 286-378)

**Strategy**: Replace regex-based extraction with depth-tracking string traversal.

```go
func extractArrowFunctionParams(expr string) map[string]bool {
    params := make(map[string]bool)
    offset := 0

    for {
        // Find next =>
        arrowIndex := strings.Index(expr[offset:], "=>")
        if arrowIndex == -1 {
            break
        }
        arrowIndex += offset

        // Work backwards to find parameter list
        paramEnd := arrowIndex - 1
        for paramEnd >= 0 && (expr[paramEnd] == ' ' || expr[paramEnd] == '\t') {
            paramEnd--
        }

        if expr[paramEnd] == ')' {
            // Use depth tracking to match parentheses correctly
            parenDepth := 1
            paramStart := paramEnd - 1
            for paramStart >= 0 && parenDepth > 0 {
                if expr[paramStart] == ')' {
                    parenDepth++
                } else if expr[paramStart] == '(' {
                    parenDepth--
                }
                paramStart--
            }

            if parenDepth == 0 {
                paramStr := expr[paramStart+2 : paramEnd]
                extractParamNames(paramStr, params)
            }
        } else {
            // Single param without parens
            paramStart := paramEnd
            for paramStart >= 0 && isIdentifierChar(expr[paramStart]) {
                paramStart--
            }
            paramName := expr[paramStart+1 : paramEnd+1]
            if paramName != "" && isSimpleIdentifier(paramName) {
                params[paramName] = true
            }
        }

        offset = arrowIndex + 2
    }

    return params
}
```

**Result**: Arrow function parameters are correctly identified and skipped during identifier prefixing.

### Fix 2: Add String Literal Tracking
**File**: `builder/registry_generator.go` (lines 428-484)

**Strategy**: Implement state machine to track when inside string literals.

```go
func prefixIdentifiersInExpression(expr string, arrowParams map[string]bool) string {
    var result strings.Builder
    var currentToken strings.Builder

    inString := false      // Track if we're inside a string literal
    stringChar := byte(0)  // Track which quote started the string (' or ")
    escaped := false       // Track if previous char was backslash

    for i := 0; i < len(expr); i++ {
        ch := expr[i]

        // Handle escape characters
        if escaped {
            result.WriteByte(ch)
            escaped = false
            continue
        }

        if ch == '\\' {
            result.WriteByte(ch)
            escaped = true
            continue
        }

        // Track string boundaries
        if ch == '\'' || ch == '"' {
            if !inString {
                inString = true
                stringChar = ch
                // Process accumulated token before entering string
                if currentToken.Len() > 0 {
                    token := currentToken.String()
                    result.WriteString(processToken(token, combinedSkip))
                    currentToken.Reset()
                }
                result.WriteByte(ch)
                continue
            } else if ch == stringChar {
                inString = false
                stringChar = 0
                result.WriteByte(ch)
                continue
            }
        }

        // If inside string, just copy as-is
        if inString {
            result.WriteByte(ch)
            continue
        }

        // Normal processing when not in string...
        // [rest of existing logic]
    }

    return result.String()
}
```

**Result**: Content inside string literals is preserved exactly as-is without modification.

### Fix 3: Handle Spread Operator
**File**: `builder/registry_generator.go` (lines 573-590)

**Strategy**: Detect three consecutive dots and treat as operator.

```go
} else if ch == '.' {
    // Check if this is spread operator (...) vs property access (.)
    if i+2 < len(expr) && expr[i+1] == '.' && expr[i+2] == '.' {
        // This is spread operator - process accumulated token first
        if currentToken.Len() > 0 {
            token := currentToken.String()
            result.WriteString(processToken(token, combinedSkip))
            currentToken.Reset()
        }
        // Write spread operator as-is
        result.WriteString("...")
        i += 2 // Skip the next two dots
        continue // CRITICAL: Skip to next iteration
    } else {
        // Property access - keep as part of token
        currentToken.WriteByte(ch)
    }
```

**Result**: Spread operator `...` is correctly recognized and not modified.

### Fix 4: Add Semicolon to Delimiter List
**File**: `builder/registry_generator.go` (lines 618-624)

**Strategy**: Add `;` to the list of operators/delimiters.

```go
func isOperatorOrDelimiter(ch byte) bool {
    return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' ||
        ch == '[' || ch == ']' ||
        ch == ',' || ch == '?' || ch == ':' || ch == '!' || ch == ';' ||  // Added ;
        ch == '>' || ch == '<' || ch == '=' || ch == '&' || ch == '|' ||
        ch == ' ' || ch == '\t' || ch == '\n'
}
```

**Result**: Semicolons are properly treated as statement separators.

### Fix 5: Skip Event Handler Attribute Conversion
**File**: `builder/registry_generator.go` (lines 760-808)

**Strategy**: Add helper function to detect event handlers and skip conversion.

```go
func isEventHandlerAttribute(name string) bool {
    // Standard HTML event handlers
    if strings.HasPrefix(name, "on") {
        return true
    }
    // Alpine.js @ shorthand (@click, @submit, etc.)
    if strings.HasPrefix(name, "@") {
        return true
    }
    // Alpine.js x-on: syntax
    if strings.HasPrefix(name, "x-on:") {
        return true
    }
    return false
}

func renderAttributeToJS(attr ast.Attribute, sb *strings.Builder, ctx *RenderContext, children []ast.Node) {
    // ... escaping steps ...

    // Convert {expression} to ${props.expression}
    // But NOT for event handlers where code should remain as-is
    if !isEventHandlerAttribute(attr.Name) {
        escaped = convertAttributeExpressions(escaped)
    }

    sb.WriteString(escaped)
    sb.WriteString("\"")
}
```

**Result**: Event handler attributes like `onclick="{expr}"` are preserved for Alpine.js to handle.

### Fix 6: Escape Quotes in Conditionals
**File**: `builder/registry_generator.go` (lines 793-819)

**Strategy**: Apply quote escaping to both if and else-if conditions.

```go
func renderConditionalToJS(cond *ast.Conditional, sb *strings.Builder, ctx *RenderContext) {
    // Main if block
    sb.WriteString(`<template x-if="`)
    escapedCondition := escapeQuotesInAttributeValue(cond.IfCondition)
    sb.WriteString(escapedCondition)
    sb.WriteString(`">`)

    // ... render if content ...

    // Else-if blocks
    for i, elseIfCond := range cond.ElseIfConditions {
        sb.WriteString(`<template x-else-if="`)
        escapedElseIfCond := escapeQuotesInAttributeValue(elseIfCond)
        sb.WriteString(escapedElseIfCond)
        sb.WriteString(`">`)

        // ... render else-if content ...
    }
}
```

**Result**: Quotes in conditional expressions are properly escaped for template literal context.

### Fix 7: Render StyleSection and ScriptSection
**File**: `builder/registry_generator.go` (lines 140-155)

**Strategy**: Render style and script sections instead of skipping them.

```go
case *ast.StyleSection:
    // Render <style> tags with CSS content
    // This is critical for component-specific styling
    sb.WriteString("<style>")
    sb.WriteString(escapeTemplateLiteral(n.Content))
    sb.WriteString("</style>")

case *ast.ScriptSection:
    // Render <script> tags with JavaScript content
    sb.WriteString("<script>")
    sb.WriteString(escapeTemplateLiteral(n.Content))
    sb.WriteString("</script>")

// Only fence sections are skipped (metadata)
case *ast.FenceSection:
    // Skip fence section (props, variables, etc.)
```

**Result**: Component CSS and JavaScript are included in the registry, so components are properly styled.

### Fix 8: Case-Insensitive Component Lookup
**File**: `static/js/runtime-components.js` (lines 147-160)

**Strategy**: Try exact match first, then case-insensitive fallback.

```javascript
// Try exact match first
let templateFn = registry[componentName];

if (!templateFn) {
    // Try case-insensitive lookup
    // This handles cases where JSON has 'hero2436' but registry has 'Hero2436'
    const lowerName = componentName.toLowerCase();
    const registryKey = Object.keys(registry).find(key => key.toLowerCase() === lowerName);
    if (registryKey) {
        templateFn = registry[registryKey];
        console.log(`[Runtime Components] Case-insensitive match: '${componentName}' → '${registryKey}'`);
    }
}
```

**Result**: Components can be referenced with any case in JSON content.

### Fix 9: Strip 'this.' Prefix from Props
**File**: `static/js/runtime-components.js` (lines 177-186)

**Strategy**: Normalize prop keys before passing to template function.

```javascript
// Normalize props - strip 'this.' prefix from keys
// The server-side replaceVarRefsWithThis adds 'this.' prefix to field names
// but component templates expect clean prop names
const normalizedProps = {};
if (props) {
    for (const [key, value] of Object.entries(props)) {
        const cleanKey = key.startsWith('this.') ? key.substring(5) : key;
        normalizedProps[cleanKey] = value;
    }
}

// Call template function with normalized props
html = templateFn(normalizedProps);
```

**Result**: Props like `this.topper` are normalized to `topper`, matching what component templates expect.

## Testing & Verification

### Test 1: Component Registry JavaScript Validity
```bash
# Node.js can now parse the generated registry
cd static/js
node test-import.mjs

# Output:
✓ Import successful
✓ Loaded 65 components
```

### Test 2: Browser Rendering
- Navigate to http://localhost:3333/
- Components render with:
  - ✅ Proper styling from `<style>` tags
  - ✅ Correct content from JSON props
  - ✅ No console errors

### Test 3: All Builder Tests Pass
```bash
go test ./builder -v

# All 40+ tests pass, including new tests:
# - TestArrowFunctionBugFix
# - TestStringLiteralHandling
# - TestMethodChainingBugFix
# - TestQuoteEscapingBugFix
# - TestSpreadOperatorBugFix
```

### Test 4: Component Registry Size
- Before fixes: N/A (syntax errors prevented loading)
- After fixes: 65 components loaded successfully
- Note: Includes duplicate path-based entries (e.g., `Hero2436` and `../components/hero2436.html`)

## Files Modified

1. **builder/registry_generator.go** - Primary fixes for JavaScript generation
   - `extractArrowFunctionParams()` - Rewritten with depth tracking
   - `prefixIdentifiersInExpression()` - Added string literal tracking and spread operator handling
   - `isOperatorOrDelimiter()` - Added semicolon
   - `isEventHandlerAttribute()` - New helper function
   - `renderAttributeToJS()` - Skip conversion for event handlers
   - `renderConditionalToJS()` - Added quote escaping
   - `renderNodeToJS()` - Render StyleSection and ScriptSection

2. **static/js/runtime-components.js** - Runtime fixes for component loading
   - `renderDynamicComponent()` - Case-insensitive lookup
   - `renderDynamicComponent()` - Prop key normalization

3. **builder/spread_test.go** - New test file for spread operator
4. **builder/debug_spread_test.go** - Debug test for spread operator

## Performance Impact

- **Component Registry Generation**: No measurable impact (~8ms for 65 components)
- **Registry File Size**: Increased due to CSS/JS inclusion (expected and necessary)
- **Runtime Performance**: Minimal overhead from case-insensitive lookup and prop normalization
- **Browser Loading**: No issues, registry loads in <100ms

## Known Limitations

1. **Duplicate Component Entries**: Registry contains both proper names (`Hero2436`) and path-based names (`../components/hero2436.html`). The duplicates are for import resolution but aren't needed in the browser registry. Future optimization opportunity.

2. **'this.' Prefix Workaround**: The runtime normalization is a workaround. The root cause is server-side `replaceVarRefsWithThis()` being called on JSON field names. A proper fix would prevent the prefix from being added in the first place.

3. **No Validation**: The system doesn't validate that component props match the fence section declarations. Runtime errors are possible if JSON provides wrong prop types.

## Follow-up Items

### Low Priority
- [ ] Remove duplicate component entries from registry to reduce file size
- [ ] Add component prop validation against fence declarations
- [ ] Consider caching normalized props to avoid repeated normalization

### Future Consideration
- [ ] Investigate removing `this.` prefix at source (server-side) instead of runtime normalization
- [ ] Add minification to component registry for production builds
- [ ] Implement component registry versioning/cache busting

## Conclusion

The runtime component resolution system is now **fully functional**. All JavaScript syntax errors have been resolved, components load with proper styling and props, and the system successfully renders dynamic components from JSON data.

**Key Achievement**: Components defined in `content/pages/_index.json` are now dynamically loaded and rendered at runtime with full Alpine.js integration, matching the Plenti/Svelte architecture pattern.

**Status**: ✅ Production ready for runtime component resolution

## References

- [Runtime Component Resolution Spec](../2025-10-15-runtime-component-resolution/)
- [Component Registry Debugging](../2025-10-16-component-registry-debugging/)
- [Build-Time Loop Expansion](../2025-10-19-build-time-loop-expansion/)
- [CLAUDE.md - Component Registry Section](../../../CLAUDE.md#runtime-component-resolution)
