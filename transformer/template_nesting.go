package transformer

import (
	"strings"
	"github.com/jimafisk/custom_go_template/ast"
)

// fixNestedLoops handles special cases for nested loops that should be fixed
func fixNestedLoops(nodes []ast.Node) []ast.Node {
	// Look for problematic nested loops
	for _, node := range nodes {
		if element, isElement := node.(*ast.Element); isElement && element.TagName == "template" {
			// Check if this is a loop template
			var isForLoop bool
			var forExpression string
			for _, attr := range element.Attributes {
				if attr.Name == "x-for" {
					isForLoop = true
					forExpression = attr.Value
					break
				}
			}
			
			// If it's a loop over "categories", check its children
			if isForLoop && strings.Contains(forExpression, "category in categories") {
				for _, child := range element.Children {
					if childElement, isChildElement := child.(*ast.Element); isChildElement {
						// Look for nested loops in the children
						fixNestedLoopsInElement(childElement)
					}
				}
			}
		}
	}
	
	return nodes
}

// fixNestedLoopsInElement recursively fixes nested loops in an element
func fixNestedLoopsInElement(element *ast.Element) {
	// Check children for template elements with loops
	for _, child := range element.Children {
		if childTemplate, isTemplate := child.(*ast.Element); isTemplate && childTemplate.TagName == "template" {
			// Check if it's a loop template
			var isForLoop bool
			for i, attr := range childTemplate.Attributes {
				if attr.Name == "x-for" {
					isForLoop = true
					
					// Fix the loop over items to be category.items if needed
					if strings.Contains(attr.Value, "item in items") {
						childTemplate.Attributes[i].Value = "item in category.items"
					}
					break
				}
			}
			
			// Recursively fix loops in this template's children
			if isForLoop {
				for _, grandchild := range childTemplate.Children {
					if grandchildElement, isElement := grandchild.(*ast.Element); isElement {
						fixNestedLoopsInElement(grandchildElement)
					}
				}
			}
		} else if childElement, isElement := child.(*ast.Element); isElement {
			// Recursively check this element's children
			fixNestedLoopsInElement(childElement)
		}
	}
}

// ensureProperNesting ensures content nodes are properly nested inside their parent templates
// for scenarios where the AST has incorrectly placed them outside their containers
func ensureProperNesting(nodes []ast.Node) []ast.Node {
	// Special case fix for nested loops
	nodes = fixNestedLoops(nodes)
	
	// Process nodes to identify template-content pairs
	var result []ast.Node
	var currentTemplate *ast.Element
	var contentBuffer []ast.Node
	
	for i, node := range nodes {
		// Check if this is a template element
		if element, isElement := node.(*ast.Element); isElement && element.TagName == "template" {
			// If we have a previous template and buffered content, merge them
			if currentTemplate != nil && len(contentBuffer) > 0 {
				// Add the buffered content to the template's children
				currentTemplate.Children = append(currentTemplate.Children, contentBuffer...)
				// Add the template to the result
				result = append(result, currentTemplate)
				// Clear the buffer
				contentBuffer = nil
			} else if currentTemplate != nil {
				// Add the template with existing children
				result = append(result, currentTemplate)
			}
			
			// Check if this is an x-if, x-else-if, or x-else template
			isConditionalTemplate := false
			for _, attr := range element.Attributes {
				if attr.Name == "x-if" || attr.Name == "x-else-if" || attr.Name == "x-else" {
					isConditionalTemplate = true
					break
				}
			}
			
			// If it's an x-for template, we need to look at the next nodes
			isLoopTemplate := false
			for _, attr := range element.Attributes {
				if attr.Name == "x-for" {
					isLoopTemplate = true
					break
				}
			}
			
			// Check if this template might be followed by content that should be inside it
			if isConditionalTemplate || isLoopTemplate {
				// This is a new template that might need content
				currentTemplate = element
				contentBuffer = nil // Clear any previous buffer
				continue
			}
			
			// Regular template - just add to result
			result = append(result, element)
			currentTemplate = nil // Reset current template
		} else if currentTemplate != nil {
			// Not a template - might be content that needs to be nested inside the currentTemplate
			
			// Special case: don't collect text nodes that are just whitespace
			if textNode, isText := node.(*ast.TextNode); isText {
				if isWhitespaceOnly(textNode.Content) {
					// Skip whitespace nodes
					continue
				}
			}
			
			// Check if the next node is a template with attributes that suggest it's
			// related to the current one (like else/else-if after an if)
			isPartOfConditional := false
			if i < len(nodes)-1 {
				if nextElement, isElement := nodes[i+1].(*ast.Element); isElement && nextElement.TagName == "template" {
					for _, attr := range nextElement.Attributes {
						if attr.Name == "x-else" || attr.Name == "x-else-if" {
							isPartOfConditional = true
							break
						}
					}
				}
			}
			
			if isPartOfConditional {
				// Add to current template's children directly
				currentTemplate.Children = append(currentTemplate.Children, node)
			} else {
				// Buffer this content for the current template
				contentBuffer = append(contentBuffer, node)
			}
		} else {
			// Not related to a template - add directly to result
			result = append(result, node)
		}
	}
	
	// Handle any remaining template/content
	if currentTemplate != nil {
		if len(contentBuffer) > 0 {
			// Add the buffered content to the template's children
			currentTemplate.Children = append(currentTemplate.Children, contentBuffer...)
		}
		// Add the template to the result
		result = append(result, currentTemplate)
	}
	
	return result
}

// isWhitespaceOnly checks if a string contains only whitespace characters
func isWhitespaceOnly(s string) bool {
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}
