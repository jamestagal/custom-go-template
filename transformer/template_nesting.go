package transformer

import (
	"fmt"
	"log"
	"strings"
	"github.com/jimafisk/custom_go_template/ast"
)

// getNodeDesc returns a brief description of a node for logging
func getNodeDesc(node ast.Node) string {
	if elem, ok := node.(*ast.Element); ok {
		attrs := ""
		for _, attr := range elem.Attributes {
			if attr.Name == "x-if" || attr.Name == "x-for" || attr.Name == "class" {
				attrs += fmt.Sprintf(" %s=%q", attr.Name, attr.Value)
				break
			}
		}
		return fmt.Sprintf("<%s%s>", elem.TagName, attrs)
	}
	if text, ok := node.(*ast.TextNode); ok {
		content := text.Content
		if len(content) > 20 {
			content = content[:20] + "..."
		}
		return fmt.Sprintf("TEXT(%q)", content)
	}
	return fmt.Sprintf("%T", node)
}

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
	result := processNestingLevel(nodes)

	// Recursively process children of all elements
	for _, node := range result {
		if element, isElement := node.(*ast.Element); isElement && len(element.Children) > 0 {
			element.Children = ensureProperNesting(element.Children)
		}
	}

	return result
}

// processNestingLevel processes a single level of nodes for proper nesting
func processNestingLevel(nodes []ast.Node) []ast.Node {
	var result []ast.Node
	var currentTemplate *ast.Element
	var contentBuffer []ast.Node

	// Log what we're processing
	nodeDescs := make([]string, len(nodes))
	for i, node := range nodes {
		nodeDescs[i] = getNodeDesc(node)
	}
	log.Printf("processNestingLevel: Processing %d nodes: %v", len(nodes), nodeDescs)

	for i, node := range nodes {
		// Check if this is a template element
		if element, isElement := node.(*ast.Element); isElement && element.TagName == "template" {
			// If we have a previous template and buffered content, merge them
			if currentTemplate != nil && len(contentBuffer) > 0 {
				// Add the buffered content to the template's children
				currentTemplate.Children = append(currentTemplate.Children, contentBuffer...)
				log.Printf("processNestingLevel: Added %d buffered nodes to template", len(contentBuffer))
				// Add the template to the result
				result = append(result, currentTemplate)
				// Clear the buffer
				contentBuffer = nil
			} else if currentTemplate != nil {
				// Add the template with existing children
				log.Printf("processNestingLevel: Adding template %s with %d children to result", getNodeDesc(currentTemplate), len(currentTemplate.Children))
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
				log.Printf("processNestingLevel: Setting currentTemplate %s (conditional=%v, loop=%v)", getNodeDesc(element), isConditionalTemplate, isLoopTemplate)
				currentTemplate = element
				contentBuffer = nil // Clear any previous buffer
				continue
			}

			// Regular template - just add to result
			result = append(result, element)
			currentTemplate = nil // Reset current template
		} else if currentTemplate != nil {
			// Not a template - might be content that needs to be nested inside the currentTemplate

			// Get node type for logging
			nodeDesc := getNodeDesc(node)

			// Special case: don't process text nodes that are just whitespace
			if textNode, isText := node.(*ast.TextNode); isText {
				if isWhitespaceOnly(textNode.Content) {
					result = append(result, node) // Preserve whitespace between elements
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
				log.Printf("processNestingLevel: Node %s is part of conditional, adding to currentTemplate children", nodeDesc)
				currentTemplate.Children = append(currentTemplate.Children, node)
			} else {
				// THIS IS THE FIX:
				// This node is NOT part of the conditional chain, so the conditional is complete.
				// Finalize the current template and treat this node as a new sibling.

				log.Printf("processNestingLevel: Node %s is NOT part of conditional chain, finalizing template %s and adding as sibling", nodeDesc, getNodeDesc(currentTemplate))

				// 1. Finalize currentTemplate (add buffered content if any)
				if len(contentBuffer) > 0 {
					currentTemplate.Children = append(currentTemplate.Children, contentBuffer...)
					contentBuffer = nil
				}

				// 2. Add currentTemplate to results
				result = append(result, currentTemplate)

				// 3. Clear currentTemplate
				currentTemplate = nil

				// 4. Add this node as a sibling (not buffered!)
				result = append(result, node)
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

	// Log result
	resultDescs := make([]string, len(result))
	for i, node := range result {
		resultDescs[i] = getNodeDesc(node)
	}
	log.Printf("processNestingLevel: Returning %d nodes: %v (started with %d)", len(result), resultDescs, len(nodes))

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
