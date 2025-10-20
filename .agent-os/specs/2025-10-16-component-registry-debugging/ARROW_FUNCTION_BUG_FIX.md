# Arrow Function Parameter Extraction Bug Fix

## Issue

**File**: `builder/registry_generator.go`
**Function**: `extractArrowFunctionParams()` and `prefixIdentifiersInExpression()`

### Problem
When converting template expressions to JavaScript template literals, arrow function parameters were being incorrectly prefixed with `props.`:

```javascript
// Template expression:
{formatPrice(products.reduce((sum, p) => sum + p.price, 0) / products.length)}

// Generated (BROKEN):
${props.formatPrice(products.reduce((props.sum, p) => props.sum + p.price, 0) / props.products.length)}
//                                    ^^^^^^^^^^^ Invalid parameter name!
```

### Root Causes

1. **Arrow Parameter Extraction**: The regex pattern didn't handle nested method calls correctly. When it saw `.reduce((sum, p) =>`, it captured the opening paren from `.reduce(` along with the arrow function params.

2. **Method Call Argument Processing**: The `prefixIdentifiersInExpression()` function treated entire method calls as single tokens, without recursively processing the arguments inside.

## Solution

### Part 1: Improved Arrow Function Parameter Extraction

Updated `extractArrowFunctionParams()` to use a more robust approach:
- Find all `=>` occurrences first
- Work backwards from each `=>` to locate the parameter list
- Use depth tracking to correctly match parentheses

```go
// Before: Regex-based, failed on nested calls
var arrowFunctionPattern = regexp.MustCompile(`\(([^)]+)\)\s*=>|([a-zA-Z_$][\w]*)\s*=>`)

// After: String traversal with depth tracking
offset := 0
for {
    arrowIndex := strings.Index(expr[offset:], "=>")
    if arrowIndex == -1 {
        break
    }
    // Look backwards to find parameter list
    // Use parenDepth to match braces correctly
}
```

### Part 2: Recursive Method Call Argument Processing

Updated `prefixIdentifiersInExpression()` to recursively process method call arguments:

```go
if isMethodCall {
    // Process the method name/chain first
    methodName := currentToken.String()
    result.WriteString(processToken(methodName, combinedSkip))
    result.WriteByte('(')

    // Find matching closing paren
    // Recursively process arguments
    if argEnd > argStart {
        args := expr[argStart:argEnd]
        processedArgs := prefixIdentifiersInExpression(args, combinedSkip)
        result.WriteString(processedArgs)
    }

    result.WriteByte(')')
}
```

## Results

### Fixed Output
```javascript
// Template expression:
{formatPrice(products.reduce((sum, p) => sum + p.price, 0) / products.length)}

// Generated (FIXED):
${props.formatPrice(props.products.reduce((sum, p) => sum + p.price, 0) / props.products.length)}
//                   ^^^^^^^^^^^^^ Correctly prefixed!
//                                 ^^^^^^^^ ✅ No prefix on arrow params
```

### Test Cases

Added comprehensive test coverage:

1. **Arrow Function Parameter Extraction**:
   - Single params: `x =>`
   - Multiple params: `(sum, p) =>`
   - Nested in method calls: `.reduce((sum, p) =>`
   - Multiple arrow functions: `.map(x => x.id).filter(id => id > 0)`

2. **Method Call Argument Processing**:
   - `{Math.floor(x)}` → `${Math.floor(props.x)}`
   - `{products.map(p => p.price * multiplier)}` → `${props.products.map(p => p.price * props.multiplier)}`

3. **Regression Test**:
   - `TestArrowFunctionBugFix` - Exact bug case from the issue

## Improvements

The fix not only resolved the arrow function bug but also improved the overall expression processing:

1. **Recursive Argument Processing**: Method call arguments are now correctly prefixed, which was previously a "KNOWN LIMITATION"
2. **Better Arrow Parameter Detection**: More robust handling of complex nested expressions
3. **Maintained Cognitive Load**: Despite the improvements, cognitive load scores remain within acceptable limits (< 30)

## Files Modified

- `builder/registry_generator.go`:
  - `extractArrowFunctionParams()` - Improved arrow parameter extraction
  - `prefixIdentifiersInExpression()` - Added recursive method call processing
  - `processToken()` - Updated to work with recursive processing

- `builder/registry_generator_test.go`:
  - `TestArrowFunctionBugFix()` - New regression test for the specific bug
  - Updated test expectations for improved behavior

## Cognitive Load Analysis

| Function | Before | After | Status |
|----------|--------|-------|--------|
| `extractArrowFunctionParams()` | 10 | 12 | ✅ Still < 30 |
| `prefixIdentifiersInExpression()` | 18 | 18 | ✅ No change |
| `processToken()` | 10 | 10 | ✅ No change |

Total file cognitive load: 28 (within acceptable limit of 30)

## Verification

All tests pass:
```bash
$ go test ./builder
ok  	github.com/jimafisk/custom_go_template/builder	0.227s
```
