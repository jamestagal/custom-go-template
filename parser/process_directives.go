package parser

import (
	"log"

	"github.com/jimafisk/custom_go_template/ast"
)

// processDirectiveNodes processes directives like if/else/for to build nested structures
func processDirectiveNodes(nodes []ast.Node) []ast.Node {
	log.Printf("[processDirectiveNodes] Processing %d nodes", len(nodes))
	
	// First process conditionals (if/else-if/else)
	processedNodes := processConditionals(nodes)
	
	// Then process loops (for/each)
	processedNodes = processLoops(processedNodes)
	
	// Finally, handle nested directives (conditionals inside loops, loops inside conditionals)
	processedNodes = processNestedDirectives(processedNodes)
	
	log.Printf("[processDirectiveNodes] Completed processing, now have %d nodes", len(processedNodes))
	return processedNodes
}

// processConditionals converts flat conditional nodes into proper nested structures
func processConditionals(nodes []ast.Node) []ast.Node {
	var result []ast.Node
	var currentConditional *ast.Conditional
	var currentContent []ast.Node
	var inConditional bool
	var inElseIf bool
	var inElse bool
	var currentElseIfIndex int
	
	log.Printf("[processConditionals] Processing %d nodes", len(nodes))
	
	for i, node := range nodes {
		log.Printf("[processConditionals] Processing node %d: %T", i, node)
		
		switch n := node.(type) {
		case *ast.Conditional:
			// Start of a new conditional
			log.Printf("[processConditionals] Found start of conditional: %s", n.IfCondition)
			
			// If we're already in a conditional, finalize it first
			if inConditional && currentConditional != nil {
				finalizeConditional(currentConditional, currentContent, inElseIf, inElse, currentElseIfIndex)
				result = append(result, currentConditional)
			}
			
			inConditional = true
			currentConditional = n
			currentContent = []ast.Node{}
			inElseIf = false
			inElse = false
		
		case *ast.ElseIfNode:
			// Add current content to the appropriate section
			if inConditional && currentConditional != nil {
				log.Printf("[processConditionals] Found else-if: %s", n.Condition)
				
				if inElseIf {
					// Add content to the previous else-if section
					if currentElseIfIndex < len(currentConditional.ElseIfContent) {
						currentConditional.ElseIfContent[currentElseIfIndex] = append(
							currentConditional.ElseIfContent[currentElseIfIndex], 
							currentContent...,
						)
					}
				} else if inElse {
					// Add content to the else section
					currentConditional.ElseContent = append(currentConditional.ElseContent, currentContent...)
				} else {
					// Add content to the if section
					currentConditional.IfContent = append(currentConditional.IfContent, currentContent...)
				}
				
				// Add this else-if condition
				currentConditional.ElseIfConditions = append(currentConditional.ElseIfConditions, n.Condition)
				currentConditional.ElseIfContent = append(currentConditional.ElseIfContent, []ast.Node{})
				currentElseIfIndex = len(currentConditional.ElseIfConditions) - 1
				inElseIf = true
				inElse = false
				currentContent = []ast.Node{}
			} else {
				// Orphaned else-if, treat as text
				result = append(result, &ast.TextNode{Content: "{else if " + n.Condition + "}"})
			}
		
		case *ast.ElseNode:
			// Add current content to the appropriate section
			if inConditional && currentConditional != nil {
				log.Printf("[processConditionals] Found else")
				
				if inElseIf {
					// Add content to the previous else-if section
					if currentElseIfIndex < len(currentConditional.ElseIfContent) {
						currentConditional.ElseIfContent[currentElseIfIndex] = append(
							currentConditional.ElseIfContent[currentElseIfIndex], 
							currentContent...,
						)
					}
				} else if !inElse {
					// Add content to the if section
					currentConditional.IfContent = append(currentConditional.IfContent, currentContent...)
				}
				
				inElseIf = false
				inElse = true
				currentContent = []ast.Node{}
				// We'll collect else content now
			} else {
				// Orphaned else, treat as text
				result = append(result, &ast.TextNode{Content: "{else}"})
			}
		
		case *ast.IfEndNode:
			// End of conditional - finalize and add to result
			if inConditional && currentConditional != nil {
				log.Printf("[processConditionals] Found end of conditional")
				
				finalizeConditional(currentConditional, currentContent, inElseIf, inElse, currentElseIfIndex)
				result = append(result, currentConditional)
				
				inConditional = false
				inElseIf = false
				inElse = false
				currentConditional = nil
				currentContent = []ast.Node{}
			} else {
				// Orphaned end tag, treat as text
				result = append(result, &ast.TextNode{Content: "{/if}"})
			}
		
		default:
			// Regular node
			if inConditional {
				// Collect as content for the current conditional section
				log.Printf("[processConditionals] Collecting node for conditional content: %T", node)
				currentContent = append(currentContent, node)
			} else {
				// Add directly to result
				log.Printf("[processConditionals] Adding node directly to result: %T", node)
				result = append(result, node)
			}
		}
	}
	
	// Handle any unclosed conditionals (shouldn't happen in well-formed templates)
	if inConditional && currentConditional != nil {
		log.Printf("[processConditionals] WARNING: Unclosed conditional at end of template")
		finalizeConditional(currentConditional, currentContent, inElseIf, inElse, currentElseIfIndex)
		result = append(result, currentConditional)
	}
	
	log.Printf("[processConditionals] Processed %d nodes into %d result nodes", len(nodes), len(result))
	return result
}

// finalizeConditional adds the current content to the appropriate section of a conditional
func finalizeConditional(conditional *ast.Conditional, content []ast.Node, inElseIf bool, inElse bool, elseIfIndex int) {
	if inElseIf {
		// Add content to the last else-if
		if elseIfIndex < len(conditional.ElseIfContent) {
			conditional.ElseIfContent[elseIfIndex] = append(conditional.ElseIfContent[elseIfIndex], content...)
		}
	} else if inElse {
		// Add content to the else section
		conditional.ElseContent = append(conditional.ElseContent, content...)
	} else {
		// Add content to the if section
		conditional.IfContent = append(conditional.IfContent, content...)
	}
}

// processLoops converts flat loop nodes into proper nested structures
func processLoops(nodes []ast.Node) []ast.Node {
	var result []ast.Node
	var currentLoop *ast.Loop
	var currentContent []ast.Node
	var inLoop bool
	
	log.Printf("[processLoops] Processing %d nodes", len(nodes))
	
	for i, node := range nodes {
		log.Printf("[processLoops] Processing node %d: %T", i, node)
		
		switch n := node.(type) {
		case *ast.Loop:
			// Start of a new loop
			log.Printf("[processLoops] Found start of loop: %s in %s", n.Value, n.Collection)
			
			// If we're already in a loop, finalize it first
			if inLoop && currentLoop != nil {
				currentLoop.Content = append(currentLoop.Content, currentContent...)
				result = append(result, currentLoop)
			}
			
			inLoop = true
			currentLoop = n
			currentContent = []ast.Node{}
		
		case *ast.ForEndNode:
			// End of loop - finalize and add to result
			if inLoop && currentLoop != nil {
				log.Printf("[processLoops] Found end of loop")
				currentLoop.Content = append(currentLoop.Content, currentContent...)
				result = append(result, currentLoop)
				inLoop = false
				currentLoop = nil
				currentContent = []ast.Node{}
			} else {
				// Orphaned end tag, treat as text
				result = append(result, &ast.TextNode{Content: "{/for}"})
			}
		
		default:
			// Regular node
			if inLoop {
				// Collect as content for the current loop
				log.Printf("[processLoops] Collecting node for loop content: %T", node)
				currentContent = append(currentContent, node)
			} else {
				// Add directly to result
				log.Printf("[processLoops] Adding node directly to result: %T", node)
				result = append(result, node)
			}
		}
	}
	
	// Handle any unclosed loops (shouldn't happen in well-formed templates)
	if inLoop && currentLoop != nil {
		log.Printf("[processLoops] WARNING: Unclosed loop at end of template")
		currentLoop.Content = append(currentLoop.Content, currentContent...)
		result = append(result, currentLoop)
	}
	
	log.Printf("[processLoops] Processed %d nodes into %d result nodes", len(nodes), len(result))
	return result
}

// processNestedDirectives handles complex nesting of directives (conditionals in loops, loops in conditionals)
func processNestedDirectives(nodes []ast.Node) []ast.Node {
	var result []ast.Node
	
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.Conditional:
			// Process nested directives in each branch of the conditional
			n.IfContent = processDirectiveNodes(n.IfContent)
			
			for i := range n.ElseIfContent {
				n.ElseIfContent[i] = processDirectiveNodes(n.ElseIfContent[i])
			}
			
			n.ElseContent = processDirectiveNodes(n.ElseContent)
			result = append(result, n)
			
		case *ast.Loop:
			// Process nested directives in the loop content
			n.Content = processDirectiveNodes(n.Content)
			result = append(result, n)
			
		default:
			// Regular node, add as is
			result = append(result, node)
		}
	}
	
	return result
}
