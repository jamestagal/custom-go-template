package transformer

import (
	"fmt"
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// ComponentTemplate represents a registered component template
type ComponentTemplate struct {
	Name     string
	Template *ast.Template
	Props    []string // List of prop names this component accepts
}

// componentTemplateRegistry stores registered component templates
var componentTemplateRegistry = make(map[string]*ComponentTemplate)

// Track rendered component instances to prevent duplication
// This is now managed by the componentRegistry in alpine.go

// resetComponentTemplateRegistry resets the component template registry
// Note: This doesn't clear registered templates, only the instance tracking
func resetComponentTemplateRegistry() {
	// In a more complex implementation, we might want to clear
	// certain aspects of the registry but keep the templates
	log.Printf("Component template registry reset")
}

// RegisterComponent registers a component template for later use
func RegisterComponent(name string, template *ast.Template, props []string) {
	componentTemplateRegistry[name] = &ComponentTemplate{
		Name:     name,
		Template: template,
		Props:    props,
	}
	log.Printf("Registered component template: %s with %d props", name, len(props))
}

// GetComponentTemplate retrieves a component template by name
func GetComponentTemplate(name string) (*ComponentTemplate, bool) {
	template, exists := componentTemplateRegistry[name]
	return template, exists
}

// formatComponentData formats the component data scope for the x-data attribute
// This is a special formatter for components to match the expected output format
func formatComponentData(dataScope map[string]any) string {
	// For test cases, we need to format the data in a specific way
	// Create a simple string representation
	var result strings.Builder
	result.WriteString("{ ")
	
	// Add each key-value pair
	first := true
	for key, value := range dataScope {
		// Skip internal Alpine.js variables
		if strings.HasPrefix(key, "$") {
			continue
		}
		
		// Add comma if not the first item
		if !first {
			result.WriteString(", ")
		}
		first = false
		
		// Add key
		result.WriteString(key)
		result.WriteString(": ")
		
		// Add value based on type
		switch v := value.(type) {
		case string:
			// Check if this is a dynamic expression (no quotes)
			// We need to handle variable references without quotes
			if isDynamicExpression(v) {
				// This is a variable reference or expression, don't quote it
				result.WriteString(v)
			} else {
				// This is a literal string, add quotes
				result.WriteString("'")
				result.WriteString(v)
				result.WriteString("'")
			}
		case int, int64, float64:
			// Format numbers directly
			result.WriteString(fmt.Sprintf("%v", v))
		case bool:
			// Format booleans
			if v {
				result.WriteString("true")
			} else {
				result.WriteString("false")
			}
		default:
			// For other types, use a generic string representation
			result.WriteString(fmt.Sprintf("'%v'", v))
		}
	}
	
	result.WriteString(" }")
	return result.String()
}

// isDynamicExpression checks if a string is a dynamic expression that should not be quoted
func isDynamicExpression(s string) bool {
	// Common patterns that indicate a dynamic expression
	if strings.Contains(s, ".") || 
	   strings.Contains(s, "[") || 
	   strings.Contains(s, "(") ||
	   strings.Contains(s, "+") ||
	   strings.Contains(s, "-") ||
	   strings.Contains(s, "*") ||
	   strings.Contains(s, "/") {
		return true
	}
	
	// Check if it's a simple variable name (no spaces, quotes, etc.)
	if len(s) > 0 && isValidVariableName(s) {
		return true
	}
	
	return false
}

// isValidVariableName checks if a string is a valid JavaScript variable name
func isValidVariableName(s string) bool {
	if len(s) == 0 {
		return false
	}
	
	// First character must be a letter, underscore, or dollar sign
	firstChar := s[0]
	if !((firstChar >= 'a' && firstChar <= 'z') || 
		 (firstChar >= 'A' && firstChar <= 'Z') || 
		 firstChar == '_' || 
		 firstChar == '$') {
		return false
	}
	
	// Rest of the characters must be letters, numbers, underscores, or dollar signs
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || 
			 (c >= 'A' && c <= 'Z') || 
			 (c >= '0' && c <= '9') || 
			 c == '_' || 
			 c == '$') {
			return false
		}
	}
	
	return true
}

// transformComponent transforms a component node into an Alpine.js compatible structure
func transformComponent(node *ast.ComponentNode, dataScope map[string]any) []ast.Node {
	// Create a unique key for this component based on name and props
	componentKey := node.Name
	for _, prop := range node.Props {
		// Use the cleaned prop value for the key to avoid duplicates with different brace formats
		cleanedPropValue := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(prop.Value, "}"), "{"))
		componentKey += ":" + prop.Name + "=" + cleanedPropValue
	}
	
	// Check if we've rendered this exact component before in the current transformation
	if isDuplicate := componentRegistry[componentKey]; isDuplicate {
		log.Printf("Warning: Duplicate component detected: %s", componentKey)
		// Return empty node to avoid duplication
		return []ast.Node{}
	}
	
	// Mark this component as rendered
	componentRegistry[componentKey] = true
	
	// Create a child scope for this component
	// Note: We're creating an empty scope rather than inheriting from parent
	// This ensures only explicitly passed props are included
	componentScope := make(map[string]any)
	
	// Add props to the component scope
	for _, prop := range node.Props {
		propName := prop.Name
		propValue := prop.Value
		
		// Handle dynamic props (with curly braces)
		if prop.IsDynamic || strings.HasPrefix(propValue, "{") && strings.HasSuffix(propValue, "}") {
			// Clean the expression by removing curly braces
			cleanedExpr := strings.TrimSpace(propValue)
			cleanedExpr = strings.TrimPrefix(strings.TrimSuffix(cleanedExpr, "}"), "{")
			cleanedExpr = strings.TrimSpace(cleanedExpr)
			
			// Extract variables from the expression and add to parent scope
			// but don't add them to the component scope
			extractVariablesFromExpr(cleanedExpr, dataScope)
			
			// For dynamic props, we set the value to the expression itself
			// This will be evaluated in the Alpine.js context
			componentScope[propName] = cleanedExpr
		} else if prop.IsShorthand {
			// For shorthand props like {propName}, use the prop name as the value
			// This is a reference to a variable in the parent scope
			componentScope[propName] = propName
			
			// Also add to parent scope
			extractVariablesFromExpr(propName, dataScope)
		} else {
			// For static props, use the literal value
			componentScope[propName] = propValue
		}
	}
	
	// Try to find the component template
	var componentChildren []ast.Node
	
	// Check if this is a registered component
	if componentTemplate, exists := GetComponentTemplate(node.Name); exists {
		log.Printf("Found registered component template: %s", node.Name)
		
		// Transform the component template with the component scope
		// We need to avoid calling TransformAST directly to prevent circular dependency
		// Instead, transform the nodes directly
		childNodes := componentTemplate.Template.RootNodes
		transformedNodes := transformNodes(childNodes, componentScope, false)
		componentChildren = transformedNodes
	} else {
		log.Printf("Component template not found: %s, using placeholder", node.Name)
		
		// Create a placeholder for unknown components
		placeholder := &ast.Element{
			TagName: "div",
			Attributes: []ast.Attribute{
				{
					Name:  "x-component-placeholder",
					Value: node.Name,
				},
			},
			Children: []ast.Node{
				&ast.TextNode{
					Content: fmt.Sprintf("Component not found: %s", node.Name),
				},
			},
			SelfClosing: false,
		}
		
		componentChildren = []ast.Node{placeholder}
	}
	
	// Create the component wrapper element
	element := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:  "x-component",
				Value: node.Name,
			},
		},
		Children:    componentChildren,
		SelfClosing: false,
	}
	
	// Add props as data attributes for debugging and reference
	for propName, propValue := range componentScope {
		// Skip internal Alpine.js variables
		if strings.HasPrefix(propName, "$") {
			continue
		}
		
		// Add the prop as a data attribute
		attrName := fmt.Sprintf("data-prop-%s", propName)
		attrValue := fmt.Sprintf("%v", propValue)
		
		element.Attributes = append(element.Attributes, ast.Attribute{
			Name:  attrName,
			Value: attrValue,
		})
	}
	
	// Merge any new variables from the component scope back to the parent scope
	// This allows child components to affect parent state if needed
	MergeScopes(dataScope, componentScope)
	
	return []ast.Node{element}
}