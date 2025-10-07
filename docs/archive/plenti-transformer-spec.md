# AST Transformation Layer for Plenti - Implementation Specification

## Overview

This specification outlines the implementation of a transformation layer that converts Plenti's Svelte-inspired template syntax into Alpine.js compatible HTML. The transformation happens after parsing but before rendering, converting AST nodes into Alpine.js-compatible structures.

## System Architecture

### Components

1. **Parser**: Existing component that converts template syntax to AST
2. **Transformer**: New component that transforms the AST to Alpine.js compatible nodes
3. **Renderer**: Existing component that generates HTML from the AST

### Data Flow

```
Template Source → Parser → AST → Transformer → Transformed AST → Renderer → HTML Output
```

### Directory Structure

```
custom_go_template/
├── ast/                   # AST definitions
├── parser/                # Parser implementation
├── transformer/           # New transformer/transpiler
│   ├── transformer.go     # Main entry point
│   ├── expressions.go     # Expression handling
│   ├── conditionals.go    # Conditional blocks
│   ├── loops.go           # Loop blocks
│   ├── components.go      # Component references
│   └── scope.go           # Data scope management
├── renderer/              # HTML generation
└── cmd/                   # Command-line applications
```

## Detailed Requirements

### 1. Core Transformation Functions

#### 1.1 Entry Point

**Function**: `TransformAST(template *ast.Template, props map[string]any) *ast.Template`

**Purpose**: Transform an AST with Plenti's syntax to an AST with Alpine.js directives

**Algorithm**:
1. Initialize data scope with provided props
2. Collect variables from fence section
3. Transform root nodes with appropriate scope
4. Return the transformed template

#### 1.2 Node Transformation

**Function**: `transformNodes(nodes []ast.Node, dataScope map[string]any) []ast.Node`

**Purpose**: Transform a slice of AST nodes to Alpine.js compatible nodes

**Algorithm**:
1. Determine if nodes need Alpine.js data wrapper
2. If wrapper needed, create root element with x-data
3. Process each node and append to appropriate parent
4. Return transformed nodes

### 2. Data Scope Management

#### 2.1 Initialize Data Scope

**Function**: `initDataScope(props map[string]any) map[string]any`

**Purpose**: Create initial data scope from props

**Requirements**:
- Include all props as top-level keys
- Set default values for props where provided
- Return a map ready for use as Alpine.js data object

#### 2.2 Collect Variables

**Function**: `collectFenceData(fence *ast.FenceSection, dataScope map[string]any)`

**Purpose**: Extract variables declared in fence section

**Algorithm**:
1. Process all prop declarations
2. Process all variable declarations (let, var, const)
3. Add to data scope with appropriate default values

#### 2.3 Track Expression Variables

**Function**: `addExprVarsToScope(expr string, dataScope map[string]any)`

**Purpose**: Extract variable references from expressions and add to scope

**Algorithm**:
1. Parse expression to identify variable references
2. For each variable not already in scope, add with nil value
3. Skip JavaScript keywords and literals

### 3. Transformation Types

#### 3.1 Text Expressions

**Input**: `Hello {name}! Your age is {age}.`

**Output**:
```html
Hello <span x-text="name"></span>! Your age is <span x-text="age"></span>.
```

**Function**: `transformTextWithExpressions(text string, dataScope map[string]any) []ast.Node`

**Algorithm**:
1. Parse text to identify expression blocks `{...}`
2. Split text into literal text and expressions
3. For expressions, create span with x-text
4. Add referenced variables to data scope
5. Return array of nodes (mixture of TextNodes and Elements)

#### 3.2 Conditional Blocks

**Input**:
```
{if condition}
  <p>True content</p>
{else}
  <p>False content</p>
{/if}
```

**Output**:
```html
<template x-if="condition">
  <p>True content</p>
</template>
<template x-if="!condition">
  <p>False content</p>
</template>
```

**Function**: `transformConditional(cond *ast.Conditional, dataScope map[string]any) []ast.Node`

**Algorithm**:
1. Create template with x-if for the condition
2. Transform children of the if block
3. For else-if blocks, create templates with compound conditions
4. For else block, create template with negated conditions
5. Add all condition variables to data scope
6. Return array of template elements

#### 3.3 Loop Blocks

**Input**:
```
{for item in items}
  <li>{item.name}</li>
{/for}
```

**Output**:
```html
<template x-for="item in items">
  <li><span x-text="item.name"></span></li>
</template>
```

**Function**: `transformLoop(loop *ast.Loop, dataScope map[string]any) ast.Node`

**Algorithm**:
1. Format loop expression (iterator in collection)
2. Create template with x-for directive
3. Transform loop content with local scope
4. Add collection variable to data scope
5. Return template element

#### 3.4 Component References

**Input**:
```
<Component prop1="value" prop2={expr} />
```

**Output**:
```html
<div x-component="Component" 
     data-prop-prop1="value" 
     data-prop-prop2="expr"></div>
```

**Function**: `transformComponent(comp *ast.ComponentNode, dataScope map[string]any) ast.Node`

**Algorithm**:
1. Create element with x-component directive
2. Transform each prop into data attribute
3. Add expression variables to data scope
4. Return component element

### 4. Alpine.js Integration

#### 4.1 Alpine Data Wrapper

**Function**: `createAlpineWrapper(dataScope map[string]any, children []ast.Node) *ast.Element`

**Purpose**: Create root element with Alpine.js x-data directive

**Algorithm**:
1. Format data scope as JSON-compatible object
2. Create div element with x-data attribute
3. Set children as transformed nodes
4. Return Alpine.js root element

#### 4.2 Alpine.js Event Handlers

**Input**: `<button on:click={handler}>Click</button>`

**Output**: `<button @click="handler">Click</button>`

**Algorithm**:
1. Identify on: prefixed attributes
2. Transform to @ prefix syntax
3. Add referenced handlers to data scope

#### 4.3 Alpine.js Bindings

**Input**: `<input bind:value={text}>`

**Output**: `<input x-model="text">`

**Algorithm**:
1. Identify bind: prefixed attributes
2. Transform to x-model directive
3. Add bound variable to data scope

## Error Handling Strategy

### 1. Validation Checks

1. **AST Structure Validation**:
   - Validate AST before transformation
   - Check for unexpected node types
   - Ensure required fields are present

2. **Expression Validation**:
   - Check for syntax errors in expressions
   - Validate conditional expressions
   - Ensure loop variables are valid

### 2. Error Types

1. **TransformationError**: Base error type for all transformation errors
2. **ScopeError**: Error related to variable scoping
3. **SyntaxError**: Error in template syntax that wasn't caught by parser
4. **ComponentError**: Error related to component handling

### 3. Error Reporting

1. Include original node information in errors
2. Provide context about what was being transformed
3. Include line/column information where possible
4. Suggest potential fixes where applicable

## Testing Plan

### 1. Unit Tests

1. **Expression Tests**:
   - Test simple expressions: `{name}`
   - Test complex expressions: `{item.prop + 1}`
   - Test expressions with multiple variables

2. **Conditional Tests**:
   - Test if blocks: `{if condition}`
   - Test if-else blocks
   - Test if-else-if-else chains
   - Test nested conditionals

3. **Loop Tests**:
   - Test basic loops: `{for item in items}`
   - Test loops with index: `{for item, index in items}`
   - Test nested loops
   - Test loops with complex expressions

4. **Component Tests**:
   - Test basic component: `<Component />`
   - Test props passing: `<Component prop={value} />`
   - Test shorthand props: `<Component {prop} />`
   - Test dynamic components: `<{dynamicComponent} />`

### 2. Integration Tests

1. **Full Template Tests**:
   - Test complete templates with mixed syntax
   - Verify transformed output against expected Alpine.js HTML

2. **Render Pipeline Tests**:
   - Test parser → transformer → renderer pipeline
   - Verify final HTML output

### 3. Edge Case Tests

1. **Empty/Invalid Cases**:
   - Test empty templates
   - Test invalid syntax that passes parser
   - Test edge cases in expressions

2. **Complex Nested Structures**:
   - Test deeply nested conditionals/loops
   - Test components within loops within conditionals

## Implementation Plan

### Phase 1: Core Framework (Week 1)

1. Set up transformer package structure
2. Implement basic AST traversal and transformation
3. Implement data scope management
4. Create basic error handling

### Phase 2: Basic Transformations (Week 1-2)

1. Implement text expression transformation
2. Implement basic element and attribute transformation
3. Write unit tests for basic transformations

### Phase 3: Control Structures (Week 2)

1. Implement conditional block transformation
2. Implement loop block transformation
3. Write unit tests for control structures

### Phase 4: Component System (Week 3)

1. Implement component reference transformation
2. Implement props passing mechanism
3. Write unit tests for component system

### Phase 5: Integration and Testing (Week 3-4)

1. Integrate transformer with renderer
2. Implement comprehensive test suite
3. Benchmark performance and optimize

## Acceptance Criteria

1. **Correctness**: All Plenti template features transform to correct Alpine.js code
2. **Completeness**: All syntax elements are supported
3. **Error Handling**: Clear, actionable error messages for invalid templates
4. **Performance**: Minimal overhead added to rendering pipeline
5. **Maintainability**: Clean, well-documented code with comprehensive tests

## Dependencies

1. **External Libraries**:
   - No external dependencies beyond standard library

2. **Internal Dependencies**:
   - AST package for node definitions
   - Parser package for initial AST
   - Renderer package for HTML generation

## Additional Considerations

1. **Alpine.js Version Compatibility**:
   - Target Alpine.js 3.x
   - Document any version-specific features used

2. **Extension Points**:
   - Provide hooks for adding custom transformations
   - Document extension process for future syntax additions

3. **Performance Optimization**:
   - Use efficient string manipulation for expressions
   - Minimize unnecessary AST node creation
   - Consider memoization for repeated transformations
## Core Files to Create

### 1. transformer.go

```go
package transformer

import (
	"github.com/jimafisk/custom_go_template/ast"
)

// TransformAST is the main entry point for AST transformation
func TransformAST(template *ast.Template, props map[string]any) *ast.Template {
	// Initialize data scope with props
	dataScope := initDataScope(props)
	
	// Collect variables from fence section if present
	if fenceNode := findFenceSection(template.RootNodes); fenceNode != nil {
		collectFenceData(fenceNode, dataScope)
	}
	
	// Transform the nodes
	transformedNodes := transformNodes(template.RootNodes, dataScope)
	
	// Create transformed template
	transformed := &ast.Template{
		RootNodes: transformedNodes,
	}
	
	return transformed
}

// transformNodes processes a slice of nodes
func transformNodes(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	// Implementation...
}

// Determine if nodes need Alpine.js data wrapper
func needsAlpineWrapper(nodes []ast.Node) bool {
	// Implementation...
}

// Create Alpine.js data wrapper
func createAlpineWrapper(dataScope map[string]any, children []ast.Node) *ast.Element {
	// Implementation...
}
```

## 2. expressions.go

```go
package transformer

import (
	"strings"
	"regexp"
	
	"github.com/jimafisk/custom_go_template/ast"
)

// Transform text with expressions {name} to Alpine.js x-text
func transformTextWithExpressions(text string, dataScope map[string]any) []ast.Node {
	// Implementation...
}

// Extract variables from expressions
func extractExpressionVariables(expr string) []string {
	// Implementation...
}

// Format data scope as Alpine.js x-data object
func formatDataObject(dataScope map[string]any) string {
	// Implementation...
}
```

## Integration with Existing Code
In your `renderer/render.go`, add the transformation step:
```go
func Render(templatePath string, props map[string]any) (string, string, string) {
	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Error reading template: %v", err)
	}
	
	// Parse the template to AST
	ast, err := parser.ParseTemplate(string(content))
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	
	// *** Add this step ***
	// Transform the AST to Alpine.js compatible nodes
	transformedAST := transformer.TransformAST(ast, props)
	
	// Generate markup from the transformed AST
	markup := generateMarkup(transformedAST)
	script := generateScript(transformedAST)
	style := generateStyle(transformedAST)
	
	return markup, script, style
}
```


## 1.Core Transformation Functions

### Entry Point: Transform Function

```go
func TransformAST(ast *ast.Template) *ast.Template {
    // Create a new transformed template
    transformedTemplate := &ast.Template{}
    
    // Track variables that need to be in the Alpine.js data scope
    dataScope := collectDataScope(ast)
    
    // Transform root nodes
    transformedTemplate.RootNodes = transformNodes(ast.RootNodes, dataScope)
    
    return transformedTemplate
}
```

### Collecting the Data Scope

```go
func collectDataScope(ast *ast.Template) map[string]any {
    dataScope := make(map[string]any)
    
    // Extract props from fence section
    if fenceNode := findFenceSection(ast.RootNodes); fenceNode != nil {
        // Add props from fence to dataScope
        extractPropsFromFence(fenceNode, dataScope)
        
        // Add variables declared in fence
        extractVariablesFromFence(fenceNode, dataScope)
    }
    
    // Collect all variables used in expressions
    walkASTForExpressions(ast.RootNodes, dataScope)
    
    return dataScope
}
```

### Transforming Nodes

```go
func transformNodes(nodes []ast.Node, dataScope map[string]any) []ast.Node {
    transformedNodes := []ast.Node{}
    
    // Check if we need to add an x-data wrapper
    needsDataWrapper := needsAlpineDataWrapper(nodes, dataScope)
    
    if needsDataWrapper {
        // Create root element with x-data containing all variables
        rootEl := createRootElementWithAlpineData(dataScope)
        
        // Process other nodes as children of this root
        for _, node := range nodes {
            transformedNode := transformNode(node, dataScope)
            if transformedNode != nil {
                rootEl.Children = append(rootEl.Children, transformedNode)
            }
        }
        
        transformedNodes = append(transformedNodes, rootEl)
    } else {
        // Just transform each node
        for _, node := range nodes {
            transformedNode := transformNode(node, dataScope)
            if transformedNode != nil {
                transformedNodes = append(transformedNodes, transformedNode)
            }
        }
    }
    
    return transformedNodes
}
```

### Node-Specific Transformations
```go
func transformNode(node ast.Node, dataScope map[string]any) ast.Node {
    switch n := node.(type) {
    case *ast.TextNode:
        return transformTextNode(n, dataScope)
        
    case *ast.Element:
        return transformElementNode(n, dataScope)
        
    case *ast.ExpressionNode:
        return transformExpressionNode(n, dataScope)
        
    case *ast.Conditional:
        return transformConditionalNode(n, dataScope)
        
    case *ast.Loop:
        return transformLoopNode(n, dataScope)
        
    // Handle other node types...
        
    default:
        return node
    }
}
```

## 2. Specific Transformation Rules

### Text Expressions → Alpine.js x-text

```go
func transformTextNode(node *ast.TextNode, dataScope map[string]any) ast.Node {
    // If text contains {expression} patterns, transform them
    if containsExpression(node.Content) {
        // Split text and expressions
        parts := splitTextAndExpressions(node.Content)
        
        // Create elements for each part
        elements := []ast.Node{}
        for _, part := range parts {
            if isExpression(part) {
                // Create x-text span for expressions
                exprValue := extractExpressionValue(part)
                
                span := &ast.Element{
                    TagName: "span",
                    Attributes: []ast.Attribute{
                        {
                            Name:       "x-text",
                            Value:      exprValue,
                            Dynamic:    true,
                            IsAlpine:   true,
                            AlpineType: "text",
                        },
                    },
                }
                
                elements = append(elements, span)
                
                // Add to data scope if needed
                addToDataScope(exprValue, dataScope)
            } else {
                // Keep regular text as is
                elements = append(elements, &ast.TextNode{Content: part})
            }
        }
        
        // Return a document fragment with all parts
        return &ast.DocumentFragment{Children: elements}
    }
    
    return node
}
```

### Conditional Blocks → Alpine.js x-if

```go
func transformConditionalNode(node *ast.Conditional, dataScope map[string]any) ast.Node {
    // Create template element with x-if
    template := &ast.Element{
        TagName: "template",
        Attributes: []ast.Attribute{
            {
                Name:       "x-if",
                Value:      node.IfCondition,
                Dynamic:    true,
                IsAlpine:   true,
                AlpineType: "if",
            },
        },
        Children: transformNodes(node.IfContent, dataScope),
    }
    
    // Add to data scope
    addToDataScope(node.IfCondition, dataScope)
    
    // For else-if branches
    elseIfTemplates := []ast.Node{}
    for i, condition := range node.ElseIfConditions {
        // Need to negate previous conditions
        negatedPrevious := negateConditions(node.IfCondition, node.ElseIfConditions[:i])
        combinedCondition := fmt.Sprintf("(%s) && (%s)", negatedPrevious, condition)
        
        elseIfTemplate := &ast.Element{
            TagName: "template",
            Attributes: []ast.Attribute{
                {
                    Name:       "x-if",
                    Value:      combinedCondition,
                    Dynamic:    true,
                    IsAlpine:   true,
                    AlpineType: "if",
                },
            },
            Children: transformNodes(node.ElseIfContent[i], dataScope),
        }
        
        addToDataScope(condition, dataScope)
        elseIfTemplates = append(elseIfTemplates, elseIfTemplate)
    }
    
    // For else branch
    if len(node.ElseContent) > 0 {
        // Negate all previous conditions
        negatedAll := negateConditions(node.IfCondition, node.ElseIfConditions)
        
        elseTemplate := &ast.Element{
            TagName: "template",
            Attributes: []ast.Attribute{
                {
                    Name:       "x-if",
                    Value:      negatedAll,
                    Dynamic:    true,
                    IsAlpine:   true,
                    AlpineType: "if",
                },
            },
            Children: transformNodes(node.ElseContent, dataScope),
        }
        
        elseIfTemplates = append(elseIfTemplates, elseTemplate)
    }
    
    // Combine templates into a fragment
    allNodes := []ast.Node{template}
    allNodes = append(allNodes, elseIfTemplates...)
    
    return &ast.DocumentFragment{Children: allNodes}
}
```

### Loop Blocks → Alpine.js x-for
```go
func transformLoopNode(node *ast.Loop, dataScope map[string]any) ast.Node {
    // Create template element with x-for
    loopExpression := fmt.Sprintf("%s in %s", node.Iterator, node.Collection)
    
    template := &ast.Element{
        TagName: "template",
        Attributes: []ast.Attribute{
            {
                Name:       "x-for",
                Value:      loopExpression,
                Dynamic:    true,
                IsAlpine:   true,
                AlpineType: "for",
            },
        },
        Children: transformNodes(node.Content, dataScope),
    }
    
    // Add iterator variable to a special scope for this node
    loopScope := make(map[string]any)
    loopScope[node.Iterator] = nil
    
    // Add collection to data scope
    addToDataScope(node.Collection, dataScope)
    
    return template
}
```

### Component Inclusion → Alpine.js Components
```go
func transformComponentNode(node *ast.ComponentNode, dataScope map[string]any) ast.Node {
    // This is more complex and depends on how you want to handle components
    
    // Basic approach: create a div with x-data that loads the component
    componentLoader := &ast.Element{
        TagName: "div",
        Attributes: []ast.Attribute{
            {
                Name:       "x-component",
                Value:      node.Name,
                Dynamic:    node.Dynamic,
                IsAlpine:   true,
                AlpineType: "component",
            },
        },
    }
    
    // Add props as data attributes
    for _, prop := range node.Props {
        attr := ast.Attribute{
            Name:     "data-prop-" + prop.Name,
            Value:    prop.Value,
            Dynamic:  prop.IsDynamic,
            IsAlpine: false,
        }
        
        componentLoader.Attributes = append(componentLoader.Attributes, attr)
        
        // Also add to data scope if it's a variable reference
        if !prop.IsDynamic {
            addToDataScope(prop.Value, dataScope)
        }
    }
    
    return componentLoader
}
```

## 3. Creating the Alpine.js Data Wrapper
```go
func createRootElementWithAlpineData(dataScope map[string]any) *ast.Element {
    // Format data scope as JSON-like object for x-data
    dataObj := formatDataObject(dataScope)
    
    rootEl := &ast.Element{
        TagName: "div",
        Attributes: []ast.Attribute{
            {
                Name:       "x-data",
                Value:      dataObj,
                Dynamic:    true,
                IsAlpine:   true,
                AlpineType: "data",
            },
        },
    }
    
    return rootEl
}

func formatDataObject(dataScope map[string]any) string {
    parts := []string{}
    
    for key, value := range dataScope {
        if value == nil {
            // For variables with unknown value
            parts = append(parts, fmt.Sprintf("%s: undefined", key))
        } else {
            // For variables with known value (like props)
            parts = append(parts, fmt.Sprintf("%s: %v", key, value))
        }
    }
    
    return "{ " + strings.Join(parts, ", ") + " }"
}
```

## 4. Handling Props and Integration with Alpine.js
```go
func extractPropsFromFence(fence *ast.FenceSection, dataScope map[string]any) {
    for _, prop := range fence.Props {
        // Add prop to data scope with default value if available
        if prop.DefaultValue != "" {
            // Parse default value and store
            parsedValue := parseJSValue(prop.DefaultValue)
            dataScope[prop.Name] = parsedValue
        } else {
            // Just record that we need this prop
            dataScope[prop.Name] = nil
        }
    }
}

func extractVariablesFromFence(fence *ast.FenceSection, dataScope map[string]any) {
    for _, variable := range fence.Variables {
        // Add variable to data scope with its value
        if variable.Value != "" {
            parsedValue := parseJSValue(variable.Value)
            dataScope[variable.Name] = parsedValue
        } else {
            dataScope[variable.Name] = nil
        }
    }
}
```

## 5. Integration into Your Renderer
```go
func Render(template string, props map[string]any) (string, string, string) {
    // Parse template
    ast, err := parser.ParseTemplate(template)
    if err != nil {
        log.Fatalf("Error parsing template: %v", err)
    }
    
    // Transform the AST
    transformedAST := transformer.TransformAST(ast)
    
    // Apply props to the transformed AST
    applyProps(transformedAST, props)
    
    // Generate HTML, JS, and CSS
    html := generateHTML(transformedAST)
    js := generateJS(transformedAST)
    css := generateCSS(transformedAST)
    
    return html, js, css
}
```

## 6. Utility Functions
```go
// Determines if an expression contains a variable reference
func extractVariablesFromExpression(expr string) []string {
    // This would need a proper JS expression parser
    // For simplicity, a naive implementation might use regex
    re := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
    matches := re.FindAllString(expr, -1)
    
    // Filter out JS keywords and literal values
    variables := []string{}
    for _, match := range matches {
        if !isJSKeyword(match) && !isLiteralValue(match) {
            variables = append(variables, match)
        }
    }
    
    return variables
}

// Add a variable to the data scope
func addToDataScope(expr string, dataScope map[string]any) {
    variables := extractVariablesFromExpression(expr)
    for _, variable := range variables {
        if _, exists := dataScope[variable]; !exists {
            dataScope[variable] = nil
        }
    }
}
```

