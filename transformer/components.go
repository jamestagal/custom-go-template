package transformer

import (
	"fmt"
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// Track rendered components to prevent duplication
var renderedComponents = make(map[string]bool)

// resetComponentTracking resets the component tracking map
func resetComponentTracking() {
	renderedComponents = make(map[string]bool)
}

// transformComponent transforms a component node into an Alpine.js compatible structure
func transformComponent(node *ast.ComponentNode, dataScope map[string]any) []ast.Node {
	// Create a unique key for this component based on name and props
	componentKey := node.Name
	for _, prop := range node.Props {
		componentKey += ":" + prop.Name + "=" + prop.Value
	}
	
	// Check if we've rendered this exact component before in the current transformation
	if isDuplicate := renderedComponents[componentKey]; isDuplicate {
		log.Printf("Warning: Duplicate component detected: %s", componentKey)
		// Return empty node to avoid duplication
		return []ast.Node{}
	}
	
	// Mark this component as rendered
	renderedComponents[componentKey] = true
	
	// Create the component element
	element := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{
				Name:  "x-component",
				Value: node.Name,
			},
		},
		Children:    []ast.Node{},
		SelfClosing: true,
	}
	
	// Add props as data attributes
	for _, prop := range node.Props {
		propAttr := ast.Attribute{
			Name:  fmt.Sprintf("data-prop-%s", prop.Name),
			Value: prop.Value,
		}
		
		// Add the prop to the element attributes
		element.Attributes = append(element.Attributes, propAttr)
		
		// If the prop is dynamic, add it to the data scope
		if prop.IsDynamic {
			// Extract the variable name from the prop value
			varName := strings.Trim(prop.Value, "{}")
			varName = strings.TrimSpace(varName)
			
			// Add the variable to the data scope if it doesn't exist
			if _, exists := dataScope[varName]; !exists {
				dataScope[varName] = nil
			}
		}
	}
	
	return []ast.Node{element}
}