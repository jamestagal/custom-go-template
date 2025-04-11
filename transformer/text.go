package transformer

import (
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// preserveWhitespace ensures that whitespace is properly preserved between elements
func preserveWhitespace(nodes []ast.Node) []ast.Node {
	if len(nodes) == 0 {
		return nodes
	}

	var result []ast.Node
	
	// Process each node
	for i, node := range nodes {
		if textNode, ok := node.(*ast.TextNode); ok {
			// Process text nodes to preserve meaningful whitespace
			content := textNode.Content
			
			// Preserve leading whitespace if this is not the first node
			// or if it contains more than just whitespace
			preserveLeading := i > 0 || !isOnlyWhitespace(content)
			
			// Preserve trailing whitespace if this is not the last node
			// or if it contains more than just whitespace
			preserveTrailing := i < len(nodes)-1 || !isOnlyWhitespace(content)
			
			// Apply whitespace preservation
			newContent := processWhitespace(content, preserveLeading, preserveTrailing)
			
			// Only add the node if it has content after processing
			if newContent != "" {
				result = append(result, &ast.TextNode{Content: newContent})
			}
		} else {
			// For non-text nodes, process their children recursively if applicable
			if element, ok := node.(*ast.Element); ok && len(element.Children) > 0 {
				// Create a copy of the element
				newElement := *element
				
				// Process children for whitespace
				newElement.Children = preserveWhitespace(element.Children)
				
				result = append(result, &newElement)
			} else {
				// Add the node as is
				result = append(result, node)
			}
		}
	}
	
	return result
}

// isOnlyWhitespace checks if a string contains only whitespace
func isOnlyWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

// processWhitespace handles whitespace preservation in text content
func processWhitespace(content string, preserveLeading, preserveTrailing bool) string {
	// If the content is empty, return empty string
	if content == "" {
		return ""
	}
	
	// If it's only whitespace and we don't need to preserve either end, return empty
	if isOnlyWhitespace(content) && !preserveLeading && !preserveTrailing {
		return ""
	}
	
	// Normalize whitespace (convert multiple spaces to single space)
	// but only for internal whitespace, not leading or trailing
	re := regexp.MustCompile(`\s+`)
	
	// Extract leading and trailing whitespace
	leadingWS := ""
	trailingWS := ""
	
	if preserveLeading {
		leadingRe := regexp.MustCompile(`^\s+`)
		leadingMatches := leadingRe.FindStringSubmatch(content)
		if len(leadingMatches) > 0 {
			leadingWS = leadingMatches[0]
		}
	}
	
	if preserveTrailing {
		trailingRe := regexp.MustCompile(`\s+$`)
		trailingMatches := trailingRe.FindStringSubmatch(content)
		if len(trailingMatches) > 0 {
			trailingWS = trailingMatches[0]
		}
	}
	
	// Normalize internal whitespace
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		// If it's only whitespace, preserve a single space if needed
		if preserveLeading || preserveTrailing {
			return " "
		}
		return ""
	}
	
	normalized := re.ReplaceAllString(trimmed, " ")
	
	// Combine with preserved whitespace
	return leadingWS + normalized + trailingWS
}
