package transformer

import (
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// InitDataScope initializes the data scope with provided props
func InitDataScope(props map[string]any) map[string]any {
	// Create a new map to avoid modifying the original props
	dataScope := make(map[string]any)

	// Copy all props to the data scope
	for key, value := range props {
		dataScope[key] = value
	}

	return dataScope
}

// FindFenceSection locates a fence section in the AST nodes
func FindFenceSection(nodes []ast.Node) *ast.FenceSection {
	for _, node := range nodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			return fence
		}
	}
	return nil
}

// CollectFenceData extracts variables from fence section and adds them to data scope
func CollectFenceData(fence *ast.FenceSection, dataScope map[string]any) {
	// Process variables directly from the FenceSection struct
	for _, variable := range fence.Variables {
		varName := variable.Name
		varValue := variable.Value

		// Try to parse the value (simple cases only)
		// For complex cases (objects, arrays, expressions), store the raw value as a string
		// so Alpine.js can evaluate it at runtime
		if strings.HasPrefix(varValue, "\"") || strings.HasPrefix(varValue, "'") {
			// String value
			dataScope[varName] = strings.Trim(varValue, "\"'")
		} else if varValue == "true" {
			dataScope[varName] = true
		} else if varValue == "false" {
			dataScope[varName] = false
		} else if varValue == "null" {
			dataScope[varName] = nil
		} else if regexp.MustCompile(`^[0-9]+$`).MatchString(varValue) {
			// Integer value
			dataScope[varName] = varValue // Keep as string, Alpine.js will convert
		} else if regexp.MustCompile(`^[0-9]*\.[0-9]+$`).MatchString(varValue) {
			// Float value
			dataScope[varName] = varValue // Keep as string, Alpine.js will convert
		} else {
			// CRITICAL FIX: For complex expressions (objects, arrays, etc.),
			// store the actual value string instead of nil.
			// This allows components to receive the expression string and
			// output it in their x-data for Alpine.js to evaluate.
			//
			// Example: let user1 = {name:"Benjamin"}
			//   Before: dataScope["user1"] = nil
			//   After:  dataScope["user1"] = "{name:\"Benjamin\"}"
			//
			// This fixes the bug where dynamic component props like user={user1}
			// were being passed as nil instead of the object expression.
			dataScope[varName] = varValue
		}
	}

	// Process props
	for _, prop := range fence.Props {
		if _, exists := dataScope[prop.Name]; !exists {
			// Only add if not already provided in props
			if prop.DefaultValue != "" {
				// Try to parse the default value (simple cases only)
				if strings.HasPrefix(prop.DefaultValue, "\"") || strings.HasPrefix(prop.DefaultValue, "'") {
					dataScope[prop.Name] = strings.Trim(prop.DefaultValue, "\"'")
				} else if prop.DefaultValue == "true" {
					dataScope[prop.Name] = true
				} else if prop.DefaultValue == "false" {
					dataScope[prop.Name] = false
				} else if prop.DefaultValue == "null" {
					dataScope[prop.Name] = nil
				} else {
					// CRITICAL FIX: Same as above - store the expression string
					// instead of nil for complex prop defaults
					dataScope[prop.Name] = prop.DefaultValue
				}
			} else {
				dataScope[prop.Name] = nil
			}
		}
	}

	// Extract variables from raw content in the fence
	extractVariablesFromExpr(fence.RawContent, dataScope)
}

// CreateChildScope creates a new scope that inherits from the parent scope
func CreateChildScope(parentScope map[string]any) map[string]any {
	childScope := make(map[string]any)

	// Copy all values from parent scope to child scope
	for key, value := range parentScope {
		childScope[key] = value
	}

	return childScope
}

// MergeScopes merges variables from child scope back to parent scope
// This is useful when we need to track variables defined in nested blocks
func MergeScopes(parentScope, childScope map[string]any) {
	// Only add variables that don't exist in parent scope
	// This prevents overwriting existing values in the parent scope
	for key, value := range childScope {
		if _, exists := parentScope[key]; !exists {
			parentScope[key] = value
		}
	}
}
