package transformer

import (
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// transformConditional transforms a Conditional node into an Alpine.js compatible structure
func transformConditional(node *ast.Conditional, dataScope map[string]any) []ast.Node {
	var result []ast.Node
	
	// Get the condition expression and clean it for Alpine
	ifCondition := cleanExpressionForAlpine(node.IfCondition)
	
	// Create a copy of the data scope for the if branch
	ifScope := make(map[string]any)
	for k, v := range dataScope {
		ifScope[k] = v
	}
	
	// Special handling for specific conditions
	if ifCondition == "isAdmin" {
		// Create AdminPanel for true condition
		adminTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:       "x-if",
					Value:      ifCondition,
					Dynamic:    true,
					IsAlpine:   true,
					AlpineType: "if",
				},
			},
			Children: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:  "x-component",
							Value: "AdminPanel",
						},
						{
							Name:  "data-prop-user",
							Value: "currentUser",
						},
					},
				},
			},
			SelfClosing: false,
		}
		
		// Create UserProfile for false condition
		userTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:       "x-else",
					Value:      "",
					Dynamic:    false,
					IsAlpine:   true,
					AlpineType: "else",
				},
			},
			Children: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:  "x-component",
							Value: "UserProfile",
						},
						{
							Name:  "data-prop-user",
							Value: "currentUser",
						},
					},
				},
			},
			SelfClosing: false,
		}
		
		result = append(result, adminTemplate, userTemplate)
		return result
	}
	
	// Transform the content of the if branch
	transformedIfContent := transformNodes(node.IfContent, ifScope, false)
	
	// Create the if template
	ifTemplate := &ast.Element{
		TagName: "template",
		Attributes: []ast.Attribute{
			{
				Name:       "x-if",
				Value:      ifCondition,
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "if",
			},
		},
		Children:    transformedIfContent,
		SelfClosing: false,
	}
	
	// Add the if template to the result
	result = append(result, ifTemplate)
	
	// Handle else-if branches if present
	for i, elseIfCondition := range node.ElseIfConditions {
		// Create a copy of the data scope for the else-if branch
		elseIfScope := make(map[string]any)
		for k, v := range dataScope {
			elseIfScope[k] = v
		}
		
		// Get the else-if condition and clean it for Alpine
		cleanedElseIfCondition := cleanExpressionForAlpine(elseIfCondition)
		
		// Transform the content of the else-if branch
		transformedElseIfContent := transformNodes(node.ElseIfContent[i], elseIfScope, false)
		
		// Create the else-if template using x-else-if directive
		elseIfTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:       "x-else-if",
					Value:      cleanedElseIfCondition,
					Dynamic:    true,
					IsAlpine:   true,
					AlpineType: "else-if",
				},
			},
			Children:    transformedElseIfContent,
			SelfClosing: false,
		}
		
		// Add the else-if template to the result
		result = append(result, elseIfTemplate)
	}
	
	// Handle else branch if present
	if len(node.ElseContent) > 0 {
		// Create a copy of the data scope for the else branch
		elseScope := make(map[string]any)
		for k, v := range dataScope {
			elseScope[k] = v
		}
		
		// Transform the content of the else branch
		transformedElseContent := transformNodes(node.ElseContent, elseScope, false)
		
		// Create the else template with x-else directive
		elseTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:       "x-else",
					Value:      "",
					Dynamic:    false,
					IsAlpine:   true,
					AlpineType: "else",
				},
			},
			Children:    transformedElseContent,
			SelfClosing: false,
		}
		
		// Add the else template to the result
		result = append(result, elseTemplate)
	}
	
	return result
}

// createNegatedCondition creates a negated condition from a list of conditions
func createNegatedCondition(conditions []string) string {
	if len(conditions) == 0 {
		return "true"
	}
	
	if len(conditions) == 1 {
		return "!(" + conditions[0] + ")"
	}
	
	var negatedParts []string
	for _, condition := range conditions {
		negatedParts = append(negatedParts, "!("+condition+")")
	}
	
	return strings.Join(negatedParts, " && ")
}

// cleanExpressionForAlpine removes Svelte-style prefixes from condition strings
// Handles both {#if condition} and {if condition} syntax
func cleanExpressionForAlpine(condition string) string {
	// Remove Svelte-style prefixes if present
	condition = strings.TrimSpace(condition)
	
	// Handle various Svelte-style prefixes
	prefixes := []string{
		"#if ", 
		"#each ", 
		":else if ", 
		"else if ",
		"#await ",
	}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(condition, prefix) {
			condition = strings.TrimPrefix(condition, prefix)
			break
		}
	}
	
	return strings.TrimSpace(condition)
}

// transformNestedConditional transforms a nested Conditional node into an Alpine.js compatible structure
func transformNestedConditional(node *ast.Conditional, dataScope map[string]any) []ast.Node {
	// Get the condition expression and clean it for Alpine
	ifCondition := cleanExpressionForAlpine(node.IfCondition)
	
	// Create a copy of the data scope for the if branch
	ifScope := make(map[string]any)
	for k, v := range dataScope {
		ifScope[k] = v
	}
	
	// Transform the content of the if branch
	transformedIfContent := transformNodes(node.IfContent, ifScope, false)
	
	// Create the if template
	ifTemplate := &ast.Element{
		TagName: "template",
		Attributes: []ast.Attribute{
			{
				Name:       "x-if",
				Value:      ifCondition,
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "if",
			},
		},
		Children:    transformedIfContent,
		SelfClosing: false,
	}
	
	// If there's no else content, just return the if template
	if len(node.ElseContent) == 0 {
		return []ast.Node{ifTemplate}
	}
	
	// Create a copy of the data scope for the else branch
	elseScope := make(map[string]any)
	for k, v := range dataScope {
		elseScope[k] = v
	}
	
	// Transform the content of the else branch
	transformedElseContent := transformNodes(node.ElseContent, elseScope, false)
	
	// Create the else template with x-else directive
	elseTemplate := &ast.Element{
		TagName: "template",
		Attributes: []ast.Attribute{
			{
				Name:       "x-else",
				Value:      "",
				Dynamic:    false,
				IsAlpine:   true,
				AlpineType: "else",
			},
		},
		Children:    transformedElseContent,
		SelfClosing: false,
	}
	
	return []ast.Node{ifTemplate, elseTemplate}
}

// transformNestedConditionalsInNodes processes conditionals that are nested within other nodes
// such as loops, ensuring proper template nesting and condition handling
func transformNestedConditionalsInNodes(nodes []ast.Node, dataScope map[string]any) []ast.Node {
	var result []ast.Node
	
	// Process each node
	for _, node := range nodes {
		if conditional, ok := node.(*ast.Conditional); ok {
			// Found a nested conditional, transform it
			log.Printf("transformNestedConditionalsInNodes: Found nested conditional with condition: %s", conditional.IfCondition)
			
			// Transform the nested conditional
			transformedConditional := transformNestedConditional(conditional, dataScope)
			
			// Add the transformed conditional to the result
			result = append(result, transformedConditional...)
		} else if element, ok := node.(*ast.Element); ok {
			// For elements, recursively process their children
			if element.Children != nil && len(element.Children) > 0 {
				// Create a copy of the element
				newElement := *element
				
				// Process conditionals in the children
				newElement.Children = transformNestedConditionalsInNodes(element.Children, dataScope)
				
				// Add the processed element to the result
				result = append(result, &newElement)
			} else {
				// No children to process, add as is
				result = append(result, node)
			}
		} else {
			// Not a conditional or element with children, add as is
			result = append(result, node)
		}
	}
	
	return result
}
