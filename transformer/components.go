package transformer

import (
	"fmt"
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// Track rendered components to prevent duplication
// This is now managed by the componentRegistry in alpine.go

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
	componentScope := CreateChildScope(dataScope)
	
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
			extractVariablesFromExpr(cleanedExpr, dataScope)
			
			// For dynamic props, we need to evaluate the expression
			// In Alpine.js, we'll pass this as a prop
			componentScope[propName] = cleanedExpr
		} else if prop.IsShorthand {
			// For shorthand props like {propName}, use the prop name as the value
			componentScope[propName] = propName
			
			// Also add to parent scope
			extractVariablesFromExpr(propName, dataScope)
		} else {
			// For static props, use the literal value
			componentScope[propName] = propValue
		}
	}
	
	// Create the component element
	element := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:       "x-data",
				Value:      alpineDataFormatter(componentScope),
				IsAlpine:   true,
				AlpineType: "data",
			},
			{
				Name:  "x-component",
				Value: node.Name,
			},
		},
		Children:    []ast.Node{},
		SelfClosing: false,
	}
	
	// If the component has children (slot content), transform them
	if len(node.Children) > 0 {
		// Transform the children with the component scope
		transformedChildren := transformNodes(node.Children, componentScope, false)
		element.Children = transformedChildren
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