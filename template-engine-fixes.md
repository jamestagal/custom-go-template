# Custom Go Template Engine Fixes

## Main Issues

Based on the test output, there are several core issues with the template engine implementation:

### 1. Alpine Data Format

The `x-data` JSON format is inconsistent with expected output:

```
// Current output
<div x-data="{"categories":[{"items":[{"name":"Laptop","price":999.99},{"name":"Phone","price":699.99}],"name":"Electronics"},{"items":[],"name":"Books"}],"category":null,"item":null,"items":null,"name":null,"price":null,"title":"Product Catalog"}">

// Expected output
<div x-data='{"title":"Product Catalog","categories":[{"name":"Electronics","items":[{"name":"Laptop","price":999.99},{"name":"Phone","price":699.99}]},{"name":"Books","items":[]}]}'>
```

Key differences:
- Property order (title should be first)
- Extra null properties being inserted
- Different JSON formatting
- Single quotes vs double quotes

### 2. Loop Transformation

Loops are being transformed incorrectly:

```
// Current output
<template x-for="(category, ) of Object.entries(categories)">

// Expected output
<template x-for="category in categories">
```

The engine is using `Object.entries()` and destructuring where it shouldn't, especially for array-based loops.

### 3. Conditional Structure

Conditional logic is flawed, with templates being included incorrectly:

```
// Current output
<template x-if="category.items.length > 0">
  <ul><!-- content --></ul>
  <p>Noitemsinthiscategory</p> <!-- Shouldn't be here -->
</template>

// Expected output
<template x-if="category.items.length > 0">
  <ul><!-- content --></ul>
</template>
<template x-if="!(category.items.length > 0)">
  <p>No items in this category</p>
</template>
```

### 4. Component Duplication

Components are being duplicated in conditionals:

```
// Current output
<template x-if="!isAdmin">
  <div x-component="UserProfile" data-prop-user="currentUser"></div>
  <div x-component="AdminPanel" data-prop-user="currentUser"></div> <!-- Shouldn't be here -->
  <div x-component="UserProfile" data-prop-user="currentUser"></div> <!-- Shouldn't be here -->
</template>

// Expected output
<template x-if="!isAdmin">
  <div x-component="UserProfile" data-prop-user="currentUser"></div>
</template>
```

### 5. Whitespace Handling

Whitespace is being improperly compressed:

```
// Current output
<p>Noitemsinthiscategory</p>

// Expected output
<p>No items in this category</p>
```

## Fix Recommendations

### 1. Alpine Data Formatting

1. Create a consistent property order in x-data JSON
2. Remove nullified scope variables from the data object
3. Apply proper JSON formatting and quoting

```go
// Possible fix in utils.go or render.go
func formatAlpineData(props map[string]any) string {
    // Filter out null values
    filteredProps := make(map[string]any)
    for k, v := range props {
        if v != nil {
            filteredProps[k] = v
        }
    }
    
    // Use single quotes for the outer JSON string
    jsonData, _ := json.Marshal(filteredProps)
    return "'" + string(jsonData) + "'"
}
```

### 2. Loop Transformation

1. Fix the loop rendering logic to use correct syntax based on collection type
2. For arrays, use `item in array` syntax
3. For objects, use `(value, key) of Object.entries(obj)` syntax

```go
// Fix in transformNode or wherever loops are processed
func transformLoop(node *ast.Loop) string {
    if isArrayType(node.Collection) {
        return fmt.Sprintf(`<template x-for="%s in %s">`, node.Value, node.Collection)
    } else {
        return fmt.Sprintf(`<template x-for="(%s, %s) of Object.entries(%s)">`, 
                          node.Value, node.Iterator, node.Collection)
    }
}
```

### 3. Conditional Structure

1. Fix the conditional rendering to properly separate branches
2. Ensure else branches use the correct syntax
3. Properly scope the elements inside each branch

```go
// Fix in transformConditional
func transformConditional(node *ast.Conditional) string {
    result := fmt.Sprintf(`<template x-if="%s">%s</template>`, 
                         node.Condition, renderContent(node.IfContent))
    
    if node.ElseContent != nil {
        result += fmt.Sprintf(`<template x-if="!(%s)">%s</template>`, 
                            node.Condition, renderContent(node.ElseContent))
    }
    
    return result
}
```

### 4. Component Handling

1. Ensure components are only rendered once
2. Fix the component reference system to avoid duplication

```go
// Fix in component handling logic to avoid duplication
func renderComponentNode(node *ast.ComponentNode) string {
    // Add tracking to prevent duplicate rendering
    if components[node.Name] {
        log.Printf("Warning: Component %s already rendered, skipping duplicate", node.Name)
        return ""
    }
    
    components[node.Name] = true
    return generateComponentHtml(node)
}
```

### 5. Whitespace Handling

1. Preserve whitespace in text nodes
2. Ensure proper spacing between elements

```go
// Fix in TextNode handling
func renderTextNode(node *ast.TextNode) string {
    // Preserve whitespace in text content
    return html.EscapeString(node.Content)
}
```

## Implementation Order

1. Fix Alpine data formatting first as it affects all tests
2. Address loop transformation issues
3. Fix conditional structure
4. Resolve component duplication
5. Correct whitespace handling

These changes should address most of the failing tests in the codebase.
