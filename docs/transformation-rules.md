# Template Transformation Rules

This document details the technical implementation of the transformation rules used in the custom Go template engine to convert template syntax to Alpine.js compatible HTML.

## Table of Contents

1. [AST Transformation Overview](#ast-transformation-overview)
2. [Expression Transformation](#expression-transformation)
3. [Conditional Transformation](#conditional-transformation)
4. [Loop Transformation](#loop-transformation)
5. [Component Transformation](#component-transformation)
6. [Whitespace Handling](#whitespace-handling)
7. [Alpine.js Data Formatting](#alpine-js-data-formatting)
8. [Scope Management](#scope-management)

## AST Transformation Overview

The template engine parses templates into an Abstract Syntax Tree (AST) and then transforms this AST into Alpine.js compatible HTML. The transformation process follows these steps:

1. Parse the template into an AST
2. Transform each node in the AST based on its type
3. Combine the transformed nodes into the final output

## Expression Transformation

Expressions are transformed into Alpine.js `x-text` directives or attribute bindings.

### Text Expressions

```html
{{ variable }}
```

Transformation rule:
```go
func transformExpression(node *ast.Expression) string {
    // Clean the expression
    cleanedExpr := cleanExpression(node.Content)
    
    // Create a span with x-text directive
    return fmt.Sprintf("<span x-text=\"%s\"></span>", cleanedExpr)
}
```

### Attribute Expressions

```html
<div class="{{ dynamicClass }}">
```

Transformation rule:
```go
func transformAttributeExpression(attr string, value string) string {
    // Clean the expression
    cleanedValue := cleanExpression(value)
    
    // Create an attribute binding
    return fmt.Sprintf(":%s=\"%s\"", attr, cleanedValue)
}
```

## Conditional Transformation

Conditionals are transformed into Alpine.js `x-if`, `x-else-if`, and `x-else` directives.

### If Condition

```html
{{ if condition }}
  <div>Content</div>
{{ end }}
```

Transformation rule:
```go
func transformConditional(node *ast.Conditional) string {
    // Clean the condition
    cleanedCondition := cleanExpression(node.Condition)
    
    // Transform the content
    transformedContent := transformNodes(node.Children)
    
    // Create a template with x-if directive
    return fmt.Sprintf("<template x-if=\"%s\">%s</template>", 
                       cleanedCondition, transformedContent)
}
```

### If-Else Condition

```html
{{ if condition }}
  <div>If content</div>
{{ else }}
  <div>Else content</div>
{{ end }}
```

Transformation rule:
```go
func transformIfElseConditional(ifNode *ast.Conditional, elseNode *ast.Conditional) string {
    // Clean the condition
    cleanedCondition := cleanExpression(ifNode.Condition)
    
    // Transform the content
    transformedIfContent := transformNodes(ifNode.Children)
    transformedElseContent := transformNodes(elseNode.Children)
    
    // Create templates with x-if and x-else directives
    return fmt.Sprintf("<template x-if=\"%s\">%s</template><template x-else>%s</template>", 
                       cleanedCondition, transformedIfContent, transformedElseContent)
}
```

### If-Else If-Else Condition

```html
{{ if conditionA }}
  <div>If content</div>
{{ else if conditionB }}
  <div>Else if content</div>
{{ else }}
  <div>Else content</div>
{{ end }}
```

Transformation rule:
```go
func transformIfElseIfElseConditional(ifNode *ast.Conditional, elseIfNode *ast.Conditional, elseNode *ast.Conditional) string {
    // Clean the conditions
    cleanedIfCondition := cleanExpression(ifNode.Condition)
    cleanedElseIfCondition := cleanExpression(elseIfNode.Condition)
    
    // Transform the content
    transformedIfContent := transformNodes(ifNode.Children)
    transformedElseIfContent := transformNodes(elseIfNode.Children)
    transformedElseContent := transformNodes(elseNode.Children)
    
    // Create templates with x-if, x-else-if, and x-else directives
    return fmt.Sprintf("<template x-if=\"%s\">%s</template><template x-else-if=\"%s\">%s</template><template x-else>%s</template>", 
                       cleanedIfCondition, transformedIfContent, 
                       cleanedElseIfCondition, transformedElseIfContent, 
                       transformedElseContent)
}
```

## Loop Transformation

Loops are transformed into Alpine.js `x-for` directives.

### Array Loop

```html
{{ for item in items }}
  <div>{{ item }}</div>
{{ end }}
```

Transformation rule:
```go
func transformArrayLoop(node *ast.Loop) string {
    // Clean the collection
    cleanedCollection := cleanExpression(node.Collection)
    
    // Create the loop expression
    loopExpr := fmt.Sprintf("%s in %s", node.Iterator, cleanedCollection)
    
    // Transform the content
    transformedContent := transformNodes(node.Children)
    
    // Create a template with x-for directive
    return fmt.Sprintf("<template x-for=\"%s\">%s</template>", 
                       loopExpr, transformedContent)
}
```

### Array Loop with Index

```html
{{ for index, item in items }}
  <div>{{ index }}: {{ item }}</div>
{{ end }}
```

Transformation rule:
```go
func transformArrayLoopWithIndex(node *ast.Loop) string {
    // Clean the collection
    cleanedCollection := cleanExpression(node.Collection)
    
    // Create the loop expression
    loopExpr := fmt.Sprintf("(%s, %s) in %s", node.Iterator, node.Value, cleanedCollection)
    
    // Transform the content
    transformedContent := transformNodes(node.Children)
    
    // Create a template with x-for directive
    return fmt.Sprintf("<template x-for=\"%s\">%s</template>", 
                       loopExpr, transformedContent)
}
```

### Object Loop

```html
{{ for key, value of object }}
  <div>{{ key }}: {{ value }}</div>
{{ end }}
```

Transformation rule:
```go
func transformObjectLoop(node *ast.Loop) string {
    // Clean the collection
    cleanedCollection := cleanExpression(node.Collection)
    
    // Create the loop expression
    loopExpr := fmt.Sprintf("%s, %s of Object.entries(%s)", node.Iterator, node.Value, cleanedCollection)
    
    // Transform the content
    transformedContent := transformNodes(node.Children)
    
    // Create a template with x-for directive
    return fmt.Sprintf("<template x-for=\"%s\">%s</template>", 
                       loopExpr, transformedContent)
}
```

## Component Transformation

Components are transformed into their content with props passed as variables.

### Component Definition

```html
{{ component Button }}
  <button class="{{ class }}">{{ label }}</button>
{{ end }}
```

Transformation rule:
```go
func transformComponent(node *ast.Component) string {
    // Store the component definition
    components[node.Name] = node.Children
    
    // Return empty string as component definitions are not rendered directly
    return ""
}
```

### Component Usage

```html
{{ Button label="Click me" class="btn btn-primary" }}
```

Transformation rule:
```go
func transformComponentUsage(node *ast.ComponentUsage) string {
    // Get the component definition
    componentDef := components[node.Name]
    
    // Create a scope with the props
    scope := createScope(node.Props)
    
    // Transform the component content with the scope
    return transformNodesWithScope(componentDef, scope)
}
```

## Whitespace Handling

The template engine preserves meaningful whitespace between elements.

```go
func preserveWhitespace(text string) string {
    // Trim leading and trailing whitespace
    trimmed := strings.TrimSpace(text)
    
    // If the text is only whitespace, return empty string
    if trimmed == "" {
        return ""
    }
    
    // Otherwise, preserve the whitespace
    return text
}
```

## Alpine.js Data Formatting

The template engine formats data for Alpine.js compatibility.

```go
func alpineDataFormatter(data interface{}) string {
    // Convert data to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return "{}"
    }
    
    // Clean the JSON for Alpine.js
    cleanedJSON := cleanJSON(string(jsonData))
    
    // Wrap in single quotes for Alpine.js
    return fmt.Sprintf("'%s'", cleanedJSON)
}
```

## Scope Management

The template engine manages variable scopes for nested structures.

```go
func createScope(variables map[string]interface{}) *Scope {
    return &Scope{
        Variables: variables,
        Parent:    nil,
    }
}

func (s *Scope) CreateChildScope() *Scope {
    return &Scope{
        Variables: make(map[string]interface{}),
        Parent:    s,
    }
}

func (s *Scope) Get(name string) (interface{}, bool) {
    // Check if the variable exists in the current scope
    if value, ok := s.Variables[name]; ok {
        return value, true
    }
    
    // If not, check the parent scope
    if s.Parent != nil {
        return s.Parent.Get(name)
    }
    
    // Variable not found
    return nil, false
}

func (s *Scope) Set(name string, value interface{}) {
    s.Variables[name] = value
}
```

This document provides a technical overview of the transformation rules used in the custom Go template engine. For more information on the template syntax, see the [Template Syntax](./template-syntax.md) documentation.
