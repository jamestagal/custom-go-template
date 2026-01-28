# Technical Specification: Dynamic Component Rendering Fix

**Date**: 2025-10-04

## Root Cause Analysis

### Current Problem

Dynamic components (`<=` syntax) are creating placeholder divs instead of rendering actual component content.

**Example from `home.html`** (lines 54-56):
```html
<={`./components/UserProfile.html`} name={name} age={age} />
<='{path}' name="Dynamic User" age={25} />
<={`./components/{comp}.html`} name="Variable Path User" age={30} />
```

**Current Output** (WRONG):
```html
<div x-component="`./components/UserProfile.html`" data-prop-name="Jim"></div>
<div x-component="'{path}'" data-prop-name="Dynamic User"></div>
<div x-component="`./components/{comp}.html`" data-prop-name="Variable Path User"></div>
```

**Expected Output** (CORRECT):
```html
<div x-data="{name: 'Jim', age: 35, ...}">
  <!-- Actual UserProfile component HTML -->
</div>
```

### Root Cause

**Component Registry Mismatch**:

1. **Registration** (`cmd/server/main.go` lines 238-242):
   ```go
   transformer.RegisterComponent(componentName, componentAST, componentProps)
   // Registers as: "UserProfile"

   pathWithPrefix := fmt.Sprintf("./components/%s.html", componentName)
   transformer.RegisterComponent(pathWithPrefix, componentAST, componentProps)
   // Registers as: "./components/UserProfile.html"
   ```

2. **Lookup** (`transformer/components.go` line 561):
   ```go
   _, exists := GetComponentTemplate(resolvedPath)
   // Looks up: "`./components/UserProfile.html`" (WITH BACKTICKS!)
   // Fails because registry has: "./components/UserProfile.html" (WITHOUT BACKTICKS)
   ```

The backticks are being included in the path expression, causing lookup failures.

---

## Solution Architecture

### Phase 1: Component Registry Normalization

**Goal**: Register components with multiple path variants so lookup succeeds regardless of path format.

#### 1.1 Path Normalization Function

```go
// normalizeComponentPath generates all possible lookup keys for a component path
// This ensures components can be found by name, relative path, or absolute path
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (path parsing: 4, key generation: 4)
//
// Example:
//   normalizeComponentPath("./components/UserProfile.html")
//   Returns: []string{
//     "UserProfile",                    // name only
//     "./components/UserProfile.html",  // relative path
//     "UserProfile.html",               // filename
//   }
func normalizeComponentPath(path string) []string {
	keys := []string{}

	// Always add the original path as-is
	keys = append(keys, path)

	// Extract filename from path
	// "./components/UserProfile.html" -> "UserProfile.html"
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		filename := parts[len(parts)-1]
		keys = append(keys, filename)

		// Extract name without extension
		// "UserProfile.html" -> "UserProfile"
		if strings.Contains(filename, ".") {
			nameParts := strings.Split(filename, ".")
			name := nameParts[0]
			keys = append(keys, name)
		}
	}

	// If path has extension, also add without extension
	// "./components/UserProfile.html" -> "./components/UserProfile"
	if strings.Contains(path, ".") && strings.Contains(path, "/") {
		pathWithoutExt := path[:strings.LastIndex(path, ".")]
		keys = append(keys, pathWithoutExt)
	}

	return keys
}
```

#### 1.2 Updated Registration

```go
// In cmd/server/main.go registerComponents()

for _, file := range files {
	if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
		componentName := strings.TrimSuffix(file.Name(), ".html")
		componentPath := fmt.Sprintf("%s/%s", componentDir, file.Name())

		// Read and parse component...

		// Generate all possible lookup keys
		relativePath := fmt.Sprintf("./components/%s", file.Name())
		allKeys := []string{
			componentName,           // "UserProfile"
			relativePath,            // "./components/UserProfile.html"
			file.Name(),             // "UserProfile.html"
			componentPath,           // "examples/components/UserProfile.html"
		}

		// Register component with ALL keys
		for _, key := range allKeys {
			transformer.RegisterComponent(key, componentAST, componentProps)
			log.Printf("Registered component with key: %s", key)
		}
	}
}
```

### Phase 2: Path Variable Resolution

**Goal**: Properly resolve path variables at build time when possible.

#### 2.1 Enhanced Path Resolution

```go
// resolvePathVariables resolves {variable} patterns in path expressions
// This is enhanced version of existing resolveDynamicPath()
//
// Pattern: Service Implementation Pattern [Load: 12]
// Cognitive Load: 12 (variable extraction: 5, value lookup: 4, substitution: 3)
//
// Example:
//   dataScope := map[string]any{"comp": "UserProfile", "path": "./components/Header.html"}
//
//   resolvePathVariables("./components/{comp}.html", dataScope)
//   // Returns: "./components/UserProfile.html"
//
//   resolvePathVariables("{path}", dataScope)
//   // Returns: "./components/Header.html"
func resolvePathVariables(pathExpr string, dataScope map[string]any) (string, error) {
	// Remove any surrounding backticks or quotes first
	cleanPath := strings.Trim(pathExpr, "`'\"")

	// Check if entire path is a single variable: {path} or {comp}
	if strings.HasPrefix(cleanPath, "{") && strings.HasSuffix(cleanPath, "}") &&
	   strings.Count(cleanPath, "{") == 1 {
		varName := strings.Trim(cleanPath, "{}")

		if val, exists := dataScope[varName]; exists && val != nil {
			if strVal, ok := val.(string); ok {
				return strVal, nil
			}
			return "", fmt.Errorf("resolvePathVariables: variable %s is not a string", varName)
		}
		return cleanPath, fmt.Errorf("resolvePathVariables: variable %s not found in scope", varName)
	}

	// Handle embedded variables: "./components/{comp}.html"
	resolved := cleanPath
	varPattern := regexp.MustCompile(`\{([a-zA-Z_$][a-zA-Z0-9_$]*)\}`)
	matches := varPattern.FindAllStringSubmatch(cleanPath, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]

			if val, exists := dataScope[varName]; exists && val != nil {
				var strVal string
				switch v := val.(type) {
				case string:
					strVal = v
				default:
					strVal = fmt.Sprintf("%v", v)
				}

				resolved = strings.Replace(resolved, match[0], strVal, 1)
				log.Printf("resolvePathVariables: Resolved {%s} to '%s'", varName, strVal)
			} else {
				return cleanPath, fmt.Errorf("resolvePathVariables: variable %s not found", varName)
			}
		}
	}

	return resolved, nil
}
```

### Phase 3: Component Inlining Implementation

**Goal**: Actually inline component content instead of creating placeholders.

#### 3.1 Rewritten transformDynamicComponent

```go
// transformDynamicComponent transforms a dynamic component node (<= syntax)
// by inlining the actual component content
//
// Pattern: Service Implementation Pattern [Load: 20]
// Cognitive Load: 20 (path resolution: 8, component lookup: 5, inlining: 7)
//
// This function MUST inline component content, not create placeholders.
func transformDynamicComponent(node *ast.DynamicComponentNode, parentDataScope map[string]any) []ast.Node {
	log.Printf("transformDynamicComponent: path=%s, props=%d", node.PathExpression, len(node.Props))

	// PHASE 1: Extract variables from path and add to parent scope
	extractVariablesFromPath(node.PathExpression, parentDataScope)

	// PHASE 2: Resolve path expression with variable substitution
	resolvedPath, err := resolvePathVariables(node.PathExpression, parentDataScope)
	if err != nil {
		log.Printf("transformDynamicComponent: Failed to resolve path: %v", err)
		// If we can't resolve, create error placeholder
		return createErrorPlaceholder(
			fmt.Sprintf("Failed to resolve component path: %s (%v)", node.PathExpression, err),
		)
	}

	log.Printf("transformDynamicComponent: Resolved path '%s' to '%s'", node.PathExpression, resolvedPath)

	// PHASE 3: Look up component template from registry
	componentTemplate, exists := GetComponentTemplate(resolvedPath)
	if !exists {
		log.Printf("transformDynamicComponent: Component not found: %s", resolvedPath)
		// Try alternative lookup keys
		alternativeKeys := normalizeComponentPath(resolvedPath)
		for _, key := range alternativeKeys {
			componentTemplate, exists = GetComponentTemplate(key)
			if exists {
				log.Printf("transformDynamicComponent: Found component with alternative key: %s", key)
				break
			}
		}

		if !exists {
			return createErrorPlaceholder(
				fmt.Sprintf("Component not found: %s (tried keys: %v)", resolvedPath, alternativeKeys),
			)
		}
	}

	// PHASE 4: Inline component using existing transformComponent logic
	// Create a ComponentNode to reuse existing transformation
	regularComponentNode := &ast.ComponentNode{
		Name:  resolvedPath, // Use resolved path as component name
		Props: node.Props,
	}

	// This will inline the actual component content
	return transformComponent(regularComponentNode, parentDataScope)
}
```

#### 3.2 Helper Functions

```go
// createErrorPlaceholder creates a visible error element for debugging
func createErrorPlaceholder(message string) []ast.Node {
	return []ast.Node{
		&ast.Element{
			TagName: "div",
			Attributes: []ast.Attribute{
				{Name: "class", Value: "component-error"},
				{Name: "style", Value: "border: 2px solid red; padding: 1rem; background: #fee;"},
			},
			Children: []ast.Node{
				&ast.TextNode{Content: "⚠️ Dynamic Component Error"},
				&ast.Element{
					TagName: "pre",
					Children: []ast.Node{
						&ast.TextNode{Content: message},
					},
				},
			},
		},
	}
}

// cloneTemplate creates a deep copy of a template AST
// This prevents mutation of the registered component template
func cloneTemplate(template *ast.Template) *ast.Template {
	// Implementation depends on AST node types
	// For now, we can reuse the template directly since Go slices are reference types
	// and transformComponent creates new scope anyway
	return template
}

// mergeMaps merges two maps, with values from second map taking precedence
func mergeMaps(base, overlay map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy base map
	for k, v := range base {
		result[k] = v
	}

	// Overlay with second map
	for k, v := range overlay {
		result[k] = v
	}

	return result
}
```

---

## Testing Strategy

### Unit Tests

**File**: `transformer/path_resolution_test.go`

```go
func TestResolvePathVariables(t *testing.T) {
	tests := []struct {
		name      string
		pathExpr  string
		dataScope map[string]any
		want      string
		wantErr   bool
	}{
		{
			name:      "Static path with backticks",
			pathExpr:  "`./components/UserProfile.html`",
			dataScope: map[string]any{},
			want:      "./components/UserProfile.html",
			wantErr:   false,
		},
		{
			name:      "Variable path {path}",
			pathExpr:  "{path}",
			dataScope: map[string]any{"path": "./components/Header.html"},
			want:      "./components/Header.html",
			wantErr:   false,
		},
		{
			name:      "Variable in path {comp}",
			pathExpr:  "./components/{comp}.html",
			dataScope: map[string]any{"comp": "UserProfile"},
			want:      "./components/UserProfile.html",
			wantErr:   false,
		},
		{
			name:      "Missing variable",
			pathExpr:  "{missingVar}",
			dataScope: map[string]any{},
			want:      "{missingVar}",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePathVariables(tt.pathExpr, tt.dataScope)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("resolvePathVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

### Integration Tests

**File**: `tests/components/dynamic_components_test.go`

```go
func TestDynamicComponentInlining(t *testing.T) {
	// Register test component
	userProfileHTML := `
	---
	prop name = "Guest"
	prop age = 0
	---
	<div class="profile">
		<h3>{name}</h3>
		<p>Age: {age}</p>
	</div>
	`

	template, _ := parser.ParseTemplate(userProfileHTML)
	transformer.RegisterComponent("UserProfile", template, []string{"name", "age"})
	transformer.RegisterComponent("./components/UserProfile.html", template, []string{"name", "age"})

	tests := []struct {
		name     string
		input    string
		wantHTML string
		wantErr  bool
	}{
		{
			name: "Static path with backticks",
			input: `
			---
			let name = "Jim"
			let age = 35
			---
			<={` + "`./components/UserProfile.html`" + `} name={name} age={age} />
			`,
			wantHTML: `<div class="profile">`,
			wantErr:  false,
		},
		{
			name: "Variable path",
			input: `
			---
			let path = "./components/UserProfile.html"
			---
			<='{path}' name="Dynamic User" age={25} />
			`,
			wantHTML: `<div class="profile">`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := parser.ParseTemplate(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			transformed := transformer.Transform(template)
			html := renderer.Render(transformed)

			if !strings.Contains(html, tt.wantHTML) {
				t.Errorf("Expected HTML to contain %q, got:\n%s", tt.wantHTML, html)
			}
		})
	}
}
```

---

## Cognitive Load Budget

| Function | Load | Status |
|----------|------|--------|
| normalizeComponentPath | 8 | ✅ OK |
| resolvePathVariables | 12 | ✅ OK |
| transformDynamicComponent | 20 | ✅ OK |
| createErrorPlaceholder | 5 | ✅ OK |
| mergeMaps | 3 | ✅ OK |

**Total**: 48 (distributed across 5 functions, all < 30)

---

## Success Metrics

1. ✅ All three dynamic component examples render actual content
2. ✅ No placeholder divs with `x-component` attribute
3. ✅ Props correctly passed and displayed
4. ✅ Component styles applied
5. ✅ No regression in static component rendering
6. ✅ All tests pass

---

## Risk Mitigation

1. **Backward Compatibility**: Register components with multiple keys to support existing code
2. **Error Visibility**: Use visible error placeholders during development
3. **Logging**: Add detailed logging for debugging path resolution
4. **Incremental Testing**: Test each phase before moving to next

---

## References

- Original dynamic component spec: `.agent-os/specs/2025-10-03-dynamic-component-paths/`
- Component transformation: `transformer/components.go`
- Component registration: `cmd/server/main.go` lines 208-244
- Example template: `examples/pages/home.html` lines 52-58
