package renderer

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/utils"
)

// InjectContentProps injects content data from JSON into exported props.
// This function is called after fence parsing but before transformation.
//
// Process:
// 1. For each prop name in fence.ExportedProps:
//    - Look up value in contentData
//    - If found: create a variable with the JSON value
//    - If not found and prop has default: use default value
//    - If not found and no default: return error
// 2. Return a modified fence section with injected variables
//
// Cognitive Load: 14
// - Loop through exported props: 2
// - Check if value exists in content: 2
// - Check if default exists: 3
// - Create variable node: 2
// - Error handling: 3
// - Append to result: 2
func InjectContentProps(fence *ast.FenceSection, contentData map[string]interface{}) (*ast.FenceSection, error) {
	if fence == nil {
		return nil, fmt.Errorf("InjectContentProps: fence section is nil")
	}

	// Create a copy of the fence section to avoid modifying the original
	result := &ast.FenceSection{
		Imports:       fence.Imports,
		Props:         fence.Props,
		ExportedProps: fence.ExportedProps,
		Variables:     make([]ast.VariableNode, len(fence.Variables)),
		Functions:     fence.Functions,
		Stores:        fence.Stores,
		RawContent:    fence.RawContent,
	}

	// Copy existing variables
	copy(result.Variables, fence.Variables)

	// If no exported props, return as-is (backward compatibility)
	if len(fence.ExportedProps) == 0 {
		return result, nil
	}

	// Build a map of prop defaults for quick lookup
	propDefaults := make(map[string]string)
	for _, prop := range fence.Props {
		propDefaults[prop.Name] = prop.DefaultValue
	}

	// Magic variables that should be injected later by the server (not from content)
	magicVars := map[string]bool{
		"allContent": true,
		"content":    true,
		"allLayouts": true,
		"components": true,
	}

	// Process each exported prop
	for _, propName := range fence.ExportedProps {
		// Skip magic variables - they're injected later by the server
		if magicVars[propName] {
			log.Printf("Skipping magic variable '%s' during content injection (will be added later)", propName)
			continue
		}

		// Check if value exists in content data
		value, exists := contentData[propName]

		if exists {
			// Content provides value - inject it
			// Convert to string representation for VariableNode
			// For complex types (maps, arrays), use JSON; for primitives, use string representation
			var jsValue string
			switch v := value.(type) {
			case bool:
				// Store boolean as "true" or "false" without quotes
				jsValue = fmt.Sprintf("%t", v)
			case string:
				// Store strings with quotes
				jsValue = strconv.Quote(v)
			default:
				// For complex types, use utils.AnyToJSValue
				jsValue = utils.AnyToJSValue(value)
			}
			result.Variables = append(result.Variables, ast.VariableNode{
				Keyword: "let",
				Name:    propName,
				Value:   jsValue,
			})
		} else {
			// Check if prop has a default value
			defaultValue, hasDefault := propDefaults[propName]

			if hasDefault {
				// Use default value with warning
				log.Printf("Warning: exported prop '%s' not found in content, using default value", propName)
				result.Variables = append(result.Variables, ast.VariableNode{
					Keyword: "let",
					Name:    propName,
					Value:   defaultValue,
				})
			} else {
				// No value and no default - error
				return nil, fmt.Errorf("exported prop '%s' not found in content and has no default value", propName)
			}
		}
	}

	return result, nil
}
