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
	// Alpine.js doesn't support x-else-if or x-else, so we need to use negated x-if conditions
	// Build up negation of all previous conditions
	var previousConditions []string
	previousConditions = append(previousConditions, node.IfCondition)

	if len(node.ElseIfConditions) > 0 {
		for i, condition := range node.ElseIfConditions {
			// Extract variables from the else-if condition
			extractVariablesFromExpr(condition, dataScope)

			// Build the negated condition: !(A) && (B)
			// Where A is all previous conditions and B is current condition
			negatedPrevious := ""
			for j, prev := range previousConditions {
				if j > 0 {
					negatedPrevious += " && "
				}
				negatedPrevious += "!(" + prev + ")"
			}

			elseIfCondition := "(" + negatedPrevious + ") && (" + condition + ")"

			// Create a template element for the else-if branch using x-if
			elseIfTemplate := &ast.Element{
				TagName: "template",
				Attributes: []ast.Attribute{
					{
						Name:  "x-if",
						Value: elseIfCondition,
					},
				},
				Children: []ast.Node{},
			}

			// Transform the content of the else-if branch
			elseIfContent := transformNodes(node.ElseIfContent[i], dataScope, false)
			elseIfTemplate.Children = elseIfContent

			// Add the else-if template to the result
			result = append(result, elseIfTemplate)

			// Track this condition for future else-if/else branches
			previousConditions = append(previousConditions, condition)
		}
	}

	// Handle the else branch if present
	if len(node.ElseContent) > 0 {
		// Build the negation of all previous conditions: !(A) && !(B) && ...
		negatedAll := ""
		for i, prev := range previousConditions {
			if i > 0 {
				negatedAll += " && "
			}
			negatedAll += "!(" + prev + ")"
		}

		// Create a template element for the else branch using negated x-if
		elseTemplate := &ast.Element{
			TagName: "template",
			Attributes: []ast.Attribute{
				{
					Name:  "x-if",
					Value: negatedAll,
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
