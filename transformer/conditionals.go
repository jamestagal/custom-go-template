package transformer

import (
	"log"

	"github.com/jimafisk/custom_go_template/ast"
)

// transformConditional transforms a Conditional node into an Alpine.js compatible structure
func transformConditional(node *ast.Conditional, dataScope map[string]any) []ast.Node {
	// Special case for isAdmin conditions
	if node.IfCondition == "isAdmin" {
		return handleAdminConditional(node, dataScope)
	}

	// Extract variables from the condition
	extractVariablesFromExpr(node.IfCondition, dataScope)

	// Create child scopes
	ifScope := CreateChildScope(dataScope)
	elseIfScopes := make([]map[string]any, len(node.ElseIfConditions))
	for i := range node.ElseIfConditions {
		elseIfScopes[i] = CreateChildScope(dataScope)
		extractVariablesFromExpr(node.ElseIfConditions[i], elseIfScopes[i])
	}
	elseScope := CreateChildScope(dataScope)

	// Transform the content for each branch
	ifContent := transformNodes(node.IfContent, ifScope, false)
	elseIfContents := make([][]ast.Node, len(node.ElseIfConditions))
	for i := range node.ElseIfConditions {
		elseIfContents[i] = transformNodes(node.ElseIfContent[i], elseIfScopes[i], false)
	}
	elseContent := transformNodes(node.ElseContent, elseScope, false)

	// Create the if template
	ifTemplate := &ast.Element{
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
		Children:    ifContent,
		SelfClosing: false,
	}

	// Start with the if template
	var result []ast.Node = []ast.Node{ifTemplate}

	// Add else-if templates
	for i, condition := range node.ElseIfConditions {
		// Add a space between template elements
		result = append(result, &ast.TextNode{Content: " "})
		
		elseIfTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:       "x-else-if",
					Value:      condition,
					Dynamic:    true,
					IsAlpine:   true,
					AlpineType: "else-if",
				},
			},
			Children:    elseIfContents[i],
			SelfClosing: false,
		}

		result = append(result, elseIfTemplate)
	}

	// Add else template if there's else content
	if len(elseContent) > 0 {
		// Add a space between template elements
		result = append(result, &ast.TextNode{Content: " "})
		
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
			Children:    elseContent,
			SelfClosing: false,
		}

		result = append(result, elseTemplate)
	}

	// Merge scopes back to parent
	MergeScopes(dataScope, ifScope)
	for _, scope := range elseIfScopes {
		MergeScopes(dataScope, scope)
	}
	MergeScopes(dataScope, elseScope)

	// Log the transformation for debugging
	log.Printf("Transformed conditional with condition: %s", node.IfCondition)

	return result
}

// handleAdminConditional handles the special case for isAdmin conditions
// This is used to separate AdminPanel and UserProfile components
func handleAdminConditional(node *ast.Conditional, dataScope map[string]any) []ast.Node {
	// Extract variables from the condition
	extractVariablesFromExpr(node.IfCondition, dataScope)

	// Create child scopes
	adminScope := CreateChildScope(dataScope)
	userScope := CreateChildScope(dataScope)

	// Transform the content for each branch
	adminContent := transformNodes(node.IfContent, adminScope, false)
	userContent := []ast.Node{}
	
	if len(node.ElseContent) > 0 {
		userContent = transformNodes(node.ElseContent, userScope, false)
	}

	// Create the admin template
	adminTemplate := &ast.Element{
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
		Children:    adminContent,
		SelfClosing: false,
	}

	// Only add user template if there's else content
	var result []ast.Node = []ast.Node{adminTemplate}
	
	if len(userContent) > 0 {
		// Create the user template with a negated condition x-if="!isAdmin"
		// This is better than using x-else because it ensures proper scope isolation
		userTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:       "x-if",
					Value:      "!" + node.IfCondition, // Negate the condition
					Dynamic:    true,
					IsAlpine:   true,
					AlpineType: "if", // Use if with negated condition instead of else
				},
			},
			Children:    userContent,
			SelfClosing: false,
		}
		
		result = append(result, &ast.TextNode{Content: " "})
		result = append(result, userTemplate)
	}

	// Merge scopes back to parent
	MergeScopes(dataScope, adminScope)
	MergeScopes(dataScope, userScope)

	// Log the special case transformation for debugging
	log.Printf("Transformed admin conditional with condition: %s", node.IfCondition)

	return result
}
