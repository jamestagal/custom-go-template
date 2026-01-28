# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-03-fence-multiline-props-fix/spec.md

> Created: 2025-10-03
> Version: 1.0.0

## Technical Requirements

### 1. Multi-line Value Detection

The parser must detect when a prop value begins with:
- Opening bracket `[` (array literal)
- Opening brace `{` (object literal)
- Any expression that may span multiple lines

### 2. Bracket/Brace Matching Algorithm

Implement a stack-based bracket matching algorithm:

```go
type bracketMatcher struct {
    stack []rune
    inString bool
    stringChar rune
}

func (bm *bracketMatcher) push(char rune) {
    if char == '[' || char == '{' || char == '(' {
        bm.stack = append(bm.stack, char)
    }
}

func (bm *bracketMatcher) pop(char rune) bool {
    if len(bm.stack) == 0 {
        return false
    }

    expected := bm.matchingBracket(char)
    if bm.stack[len(bm.stack)-1] == expected {
        bm.stack = bm.stack[:len(bm.stack)-1]
        return true
    }
    return false
}

func (bm *bracketMatcher) isComplete() bool {
    return len(bm.stack) == 0
}
```

### 3. String Literal Handling

The parser must ignore brackets/braces inside string literals:
- Track when inside single quotes `'`
- Track when inside double quotes `"`
- Handle escaped quotes `\'` and `\"`
- Do NOT count brackets/braces while `inString == true`

### 4. Line Accumulation Logic

Modified parsing flow:

```go
func parseMultiLinePropValue(lines []string, startIndex int) (value string, endIndex int) {
    matcher := newBracketMatcher()
    accumulator := strings.Builder{}

    for i := startIndex; i < len(lines); i++ {
        line := lines[i]
        accumulator.WriteString(line)

        // Process each character
        for _, char := range line {
            matcher.processChar(char)
        }

        // Check if complete
        if matcher.isComplete() {
            return accumulator.String(), i
        }

        // Add newline for multi-line continuation
        if i < len(lines)-1 {
            accumulator.WriteString("\n")
        }
    }

    return "", -1 // Error: unclosed brackets
}
```

### 5. Prop Parsing Integration

Update `extractComponentProps()` in `cmd/server/main.go`:

**Current behavior:**
```go
parts := strings.SplitN(line, "=", 2)
value := strings.TrimSpace(parts[1])
props[propName] = value // Only gets first line!
```

**New behavior:**
```go
parts := strings.SplitN(line, "=", 2)
firstLineValue := strings.TrimSpace(parts[1])

// Check if multi-line value
if needsMultiLineCapture(firstLineValue) {
    fullValue, endLine := parseMultiLinePropValue(lines, currentLineIndex)
    props[propName] = fullValue
    currentLineIndex = endLine
} else {
    props[propName] = firstLineValue
}
```

### 6. Function Expression Preservation

Function expressions must be kept as strings:

```go
// Input
prop year = new Date().getFullYear()

// Output in props map
"year": "new Date().getFullYear()"

// Alpine.js will evaluate at runtime
x-data="{ year: new Date().getFullYear() }"
```

### 7. Error Handling

Provide clear error messages:
- Unclosed brackets: "Prop 'propName' has unclosed bracket on line X"
- Mismatched brackets: "Prop 'propName' has mismatched brackets: expected ']' but found '}' on line X"
- Unexpected end of fence: "Prop 'propName' value incomplete: fence section ended before closing bracket"

## Approach

### Phase 1: Implement Bracket Matcher (1-2 hours)

1. Create `parser/bracket_matcher.go` with:
   - `BracketMatcher` struct
   - Character processing logic
   - String literal detection
   - Stack-based bracket tracking

2. Write unit tests for bracket matcher:
   - Simple arrays: `[1, 2, 3]`
   - Simple objects: `{a: 1, b: 2}`
   - Nested structures: `{a: [1, 2], b: {c: 3}}`
   - Strings with brackets: `{msg: "array [1,2]"}`
   - Edge cases: empty arrays, empty objects

### Phase 2: Update Fence Parser (2-3 hours)

1. Modify `cmd/server/main.go`:
   - Add `parseMultiLinePropValue()` function
   - Update `extractComponentProps()` to use multi-line parser
   - Handle line index advancement correctly

2. Ensure backward compatibility:
   - Single-line props continue to work
   - No performance regression for simple props

### Phase 3: Testing (2-3 hours)

1. Create test file: `parser/fence_multiline_test.go`
2. Test cases:
   - Multi-line arrays with objects
   - Multi-line objects with nested arrays
   - Function expressions
   - Mixed single and multi-line props in same fence
   - Malformed input (unclosed brackets)

3. Integration testing:
   - Test Footer.html rendering with full links array
   - Verify props are correctly passed to Alpine.js x-data
   - Confirm no regression in existing components

### Phase 4: Documentation (1 hour)

1. Update CLAUDE.md with fence parsing behavior
2. Add examples of valid multi-line prop syntax
3. Document error messages and troubleshooting

## External Dependencies

**None.** This is purely an internal parser enhancement using existing Go standard library:
- `strings` - String manipulation
- `unicode` - Character classification
- Standard Go testing framework

## Performance Considerations

- Multi-line parsing only activated when needed (opening bracket detected)
- Single-line props use fast path (no overhead)
- Bracket matcher uses stack with minimal allocations
- Expected performance impact: negligible (<1ms per component)

## Backward Compatibility

**Fully backward compatible:**
- Single-line props parse exactly as before
- No changes to prop syntax
- No changes to component API
- Existing components continue to work unchanged

## Migration Path

**No migration required.** Existing components work as-is. Components can be updated to use multi-line formatting at developer's discretion.

**Example migration:**

```html
<!-- Before (works but cramped) -->
---
prop links = [{ label: "Home", url: "/" }, { label: "About", url: "/about" }]
---

<!-- After (works with better readability) -->
---
prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]
---
```
