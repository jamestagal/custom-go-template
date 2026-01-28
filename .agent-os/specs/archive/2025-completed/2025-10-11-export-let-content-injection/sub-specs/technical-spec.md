# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-11-export-let-content-injection/spec.md

## Technical Requirements

### 1. Fence Parser Enhancement

**File**: `parser/fence.go`

- Add recognition of `export let` keyword at start of fence lines
- Parse exported prop names (support comma-separated list: `export let title, description, author`)
- Store exported props separately from regular props in `ast.FenceSection`
- Maintain backward compatibility with existing `prop` declarations

**AST Changes** (`ast/ast.go`):
```go
type FenceSection struct {
    // ... existing fields ...
    ExportedProps []string // New field: prop names that should come from content JSON
}
```

### 2. Content JSON Loader

**New File**: `content/loader.go`

Create content loading utilities:
```go
// LoadContentJSON loads JSON from content/ directory based on path
// Example: "/store-demo" -> "content/pages/store-demo.json"
func LoadContentJSON(routePath string) (map[string]interface{}, error)

// ExtractComponentFields extracts fields for a specific component from Plenti structure
// Supports both flat JSON and components array format
func ExtractComponentFields(data map[string]interface{}, componentName string) map[string]interface{}
```

**JSON Structure Support**:

**IMPORTANT**: Plenti uses a strict structure where ALL content MUST be wrapped in a `components` array. This is non-negotiable for Plenti compatibility.

**Plenti-compliant format** (REQUIRED for Plenti integration):
```json
{
  "components": [
    {
      "name": "page_header",
      "fields": {
        "title": "My Page",
        "description": "Page description"
      }
    },
    {
      "name": "stats",
      "fields": {
        "teachers": {"title": "TEACHERS", "amount": "60"},
        "courses": {"title": "COURSES", "amount": "50"}
      }
    }
  ]
}
```

**Simple format** (for standalone use, NOT Plenti-compatible):
```json
{
  "title": "My Page",
  "description": "Page description"
}
```

**Implementation Note**: For MVP, we can support both formats (detect if `components` array exists), but eventual Plenti integration will require the `components` array structure exclusively.

### 3. Prop Injection in renderTemplate

**File**: `cmd/server/main.go`

Modify `renderTemplate()` signature and implementation:
```go
// Old
func renderTemplate(entrypoint string, w http.ResponseWriter, r *http.Request)

// New
func renderTemplate(entrypoint string, w http.ResponseWriter, r *http.Request, contentData map[string]interface{})
```

**Injection Logic**:
1. After parsing template, check fence section for `ExportedProps`
2. For each exported prop name, look up value in `contentData`
3. If found, add/override prop in fence section before transformation
4. If not found, use default value if specified, otherwise error

### 4. Route Handler Updates

Update all route handlers to load content JSON:
```go
http.HandleFunc("/store-demo", func(w http.ResponseWriter, r *http.Request) {
    // Load content JSON for this route
    contentData, err := content.LoadContentJSON("/store-demo")
    if err != nil {
        // Handle error or use empty data
        contentData = make(map[string]interface{})
    }

    // Pass content data to template renderer
    renderTemplate("layouts/content/store-demo.html", w, r, contentData)
})
```

### 5. Error Handling

- **Missing JSON file**: Log warning, continue with empty content data
- **Invalid JSON**: Return error with details about parsing failure
- **Missing exported prop**: Error if no default value, warning if default exists
- **Type mismatches**: Convert JSON types to appropriate prop types where possible

### 6. Testing Requirements

**New Test File**: `content/loader_test.go`
- Test simple flat JSON loading
- Test Plenti components array format
- Test missing files (should not error)
- Test invalid JSON (should error)
- Test nested field extraction

**Integration Test**: `tests/content_injection_test.go`
- Test export let parsing in fence section
- Test prop injection from JSON
- Test mixed export let and regular props
- Test Plenti structure with multiple components

## Performance Considerations

- Cache loaded JSON files in memory (similar to component registration)
- Only reload on file changes in development mode
- Minimize JSON parsing overhead by reusing parsed data

## Backward Compatibility

- Existing templates without `export let` continue to work unchanged
- `prop` declarations still work as before
- `export let` is additive, not breaking
