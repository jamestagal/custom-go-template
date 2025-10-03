package transformer

import (
	"log"

	"github.com/jimafisk/custom_go_template/ast"
)

// transformConditional transforms a Conditional node into an Alpine.js compatible structure
func transformConditional(node *ast.Conditional, dataScope map[string]any) []ast.Node {
	// Extract variables from the condition expression
	extractVariablesFromExpr(node.IfCondition, dataScope)

	// Log the condition for debugging
	log.Printf("Transformed conditional with condition: %s", node.IfCondition)

	// Create a template element with x-if directive
	templateElement := &ast.Element{
		TagName: "template",
		Attributes: []ast.Attribute{
			{
				Name:  "x-if",
				Value: node.IfCondition,
			},
		},
		Children: []ast.Node{},
	}

	// Transform the content of the if branch
	transformedContent := transformNodes(node.IfContent, dataScope, false)
	templateElement.Children = transformedContent

	// Create a result slice with the if template
	result := []ast.Node{templateElement}

	// Handle else-if and else branches if present
	if len(node.ElseIfConditions) > 0 {
		for i, condition := range node.ElseIfConditions {
			// Extract variables from the else-if condition
			extractVariablesFromExpr(condition, dataScope)

			// Create a template element for the else-if branch
			elseIfTemplate := &ast.Element{
				TagName: "template",
				Attributes: []ast.Attribute{
					{
						Name:  "x-else-if",
						Value: condition,
					},
				},
				Children: []ast.Node{},
			}

			// Transform the content of the else-if branch
			elseIfContent := transformNodes(node.ElseIfContent[i], dataScope, false)
			elseIfTemplate.Children = elseIfContent

			// Add the else-if template to the result
			result = append(result, elseIfTemplate)
		}
	}

	// Handle the else branch if present
	if len(node.ElseContent) > 0 {
		// Create a template element for the else branch with x-else attribute
		elseTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:  "x-else",
					Value: "",
				},
			},
			Children: []ast.Node{},
		}

		// Transform the content of the else branch
		elseContent := transformNodes(node.ElseContent, dataScope, false)
		elseTemplate.Children = elseContent

		// Add the else template to the result
		result = append(result, elseTemplate)
	}

	return result
}
