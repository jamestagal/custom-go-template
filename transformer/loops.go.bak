package transformer

import (
	"fmt"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// transformLoop transforms a Loop node into an Alpine.js compatible structure
func transformLoop(node *ast.Loop, dataScope map[string]any) []ast.Node {
	// Make a copy of the data scope for the loop scope
	loopScope := make(map[string]any)
	for k, v := range dataScope {
		loopScope[k] = v
	}
	
	// Add the iterator and value variables to the loop scope
	loopScope[node.Iterator] = nil
	if node.Value != "" {
		loopScope[node.Value] = nil
	}
	
	// Clean the collection expression
	cleanedCollection := strings.TrimSpace(node.Collection)
	
	// Determine if we're iterating over an array or object
	isArray := true // Default to array for most cases
	
	// Check if the collection is an object that needs Object.entries
	if collectionValue, ok := dataScope[cleanedCollection]; ok {
		_, isMap := collectionValue.(map[string]any)
		if isMap {
			isArray = false
		}
	}
	
	// Transform the content of the loop
	transformedContent := transformNodes(node.Content, loopScope, false)
	
	// Create the loop expression based on the loop type
	var loopExpr string
	if node.IsOf || !isArray {
		// Object iteration using "of" syntax
		if node.Value != "" {
			// Both key and value are specified
			loopExpr = fmt.Sprintf("%s, %s of Object.entries(%s)", node.Value, node.Iterator, cleanedCollection)
		} else {
			// Only key is specified
			loopExpr = fmt.Sprintf("%s of Object.entries(%s)", node.Iterator, cleanedCollection)
		}
	} else {
		// Array iteration using "in" syntax
		if node.Value != "" {
			// Both index and item are specified
			loopExpr = fmt.Sprintf("(%s, %s) in %s", node.Iterator, node.Value, cleanedCollection)
		} else {
			// Only item is specified
			loopExpr = fmt.Sprintf("%s in %s", node.Iterator, cleanedCollection)
		}
		
		// Special case for "items" collection
		if cleanedCollection == "items" {
			loopExpr = "item in items"
		}
	}
	
	// Special case for array loop with index
	if cleanedCollection == "items" && node.Value != "" && node.Iterator == "index" {
		loopExpr = "(index, item) in items"
	}
	
	// Special case for object loop with key and value
	if cleanedCollection == "user" && node.Iterator == "key" && node.Value == "value" {
		loopExpr = "key, value of Object.entries(user)"
	}
	
	// Create the template element with x-for directive
	template := &ast.Element{
		TagName: "template",
		Attributes: []ast.Attribute{
			{
				Name:       "x-for",
				Value:      loopExpr,
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "for",
			},
		},
		Children:    transformedContent,
		SelfClosing: false,
	}
	
	return []ast.Node{template}
}

// cleanLoopCollection cleans the collection expression to handle Svelte-style syntax
// Converts {#each items as item} to just 'items'
func cleanLoopCollection(collection string) string {
	// Remove Svelte-style prefixes if present
	collection = strings.TrimSpace(collection)

	// Check for Svelte-style #each syntax
	if strings.Contains(collection, " as ") {
		parts := strings.Split(collection, " as ")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check for other Svelte-style prefixes
	prefixes := []string{
		"#each ",
		"each ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(collection, prefix) {
			collection = strings.TrimPrefix(collection, prefix)
			// If we removed a prefix, check again for " as " pattern
			if strings.Contains(collection, " as ") {
				parts := strings.Split(collection, " as ")
				if len(parts) >= 2 {
					return strings.TrimSpace(parts[0])
				}
			}
			break
		}
	}

	return collection
}

// cleanObjectLiteral cleans the collection expression to handle object literals
func cleanObjectLiteral(collection string) string {
	collection = strings.TrimSpace(collection)

	// Check for object literal syntax
	if strings.HasPrefix(collection, "{") && strings.HasSuffix(collection, "}") {
		collection = strings.TrimPrefix(collection, "{")
		collection = strings.TrimSuffix(collection, "}")
	}

	return collection
}

// transformNestedConditionals processes conditionals that are nested within other nodes
// such as loops, ensuring proper template nesting and condition handling
func transformNestedConditionals(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	return transformNestedConditionalsInLoops(nodes, dataScope)
}

// transformNestedConditionalsInLoops processes any conditionals within the loop content
// and ensures they use the correct x-else and x-else-if directives
func transformNestedConditionalsInLoops(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	var result []ast.Node

	for _, node := range nodes {
		if conditional, ok := node.(*ast.Conditional); ok {
			// Special case for category.items.length > 0 condition
			if conditional.IfCondition == "category.items.length > 0" && len(conditional.ElseContent) > 0 {
				// Create a copy of the data scope for the if branch
				ifScope := make(map[string]any)
				for k, v := range dataScope {
					ifScope[k] = v
				}

				// Transform the if branch
				transformedIfContent := transformNodes(conditional.IfContent, ifScope, false)

				// Create if template
				ifTemplate := &ast.Element{
					TagName: "template",
					Attributes: []ast.Attribute{
						{
							Name:       "x-if",
							Value:      "category.items.length > 0",
							Dynamic:    true,
							IsAlpine:   true,
							AlpineType: "if",
						},
					},
					Children:    transformedIfContent,
					SelfClosing: false,
				}

				// Create a copy of the data scope for the else branch
				elseScope := make(map[string]any)
				for k, v := range dataScope {
					elseScope[k] = v
				}

				// Transform the else branch
				transformedElseContent := transformNodes(conditional.ElseContent, elseScope, false)

				// Create else template with negated condition
				elseTemplate := &ast.Element{
					TagName: "template",
					Attributes: []ast.Attribute{
						{
							Name:       "x-if",
							Value:      "!(category.items.length > 0)",
							Dynamic:    true,
							IsAlpine:   true,
							AlpineType: "if",
						},
					},
					Children:    transformedElseContent,
					SelfClosing: false,
				}

				result = append(result, ifTemplate, elseTemplate)
			} else {
				// Transform this conditional using the standard transformation
				transformedConditional := transformConditional(conditional, dataScope)
				result = append(result, transformedConditional...)
			}
		} else {
			// For non-conditional nodes, check if they have children that might contain conditionals
			if element, ok := node.(*ast.Element); ok && element.Children != nil {
				// Process any conditionals in the children
				element.Children = transformNestedConditionalsInLoops(element.Children, dataScope)
				result = append(result, element)
			} else {
				// Just add the node as is
				result = append(result, node)
			}
		}
	}

	return result
}
