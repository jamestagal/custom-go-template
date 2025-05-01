package parser

import (
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// IfStartParser parses {if condition} or {#if condition} directives with various whitespace patterns
func IfStartParser() Parser {
	return func(input string) Result {
		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")
		
		// Try patterns with various whitespace arrangements
		ifPatterns := []string{"{if ", "{#if ", "{ if ", "{ #if "}
		
		for _, pattern := range ifPatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				// Find the matching closing brace
				closeBracePos := strings.Index(trimmedInput, "}")
				if closeBracePos < 0 {
					continue // No proper closing brace found, try next pattern
				}
				
				// Extract condition and create node
				condition := trimmedInput[len(pattern):closeBracePos]
				condition = strings.TrimSpace(condition)
				
				node := &ast.Conditional{
					IfCondition:      condition,
					IfContent:        []ast.Node{},
					ElseIfConditions: []string{},
					ElseIfContent:    [][]ast.Node{},
					ElseContent:      []ast.Node{},
				}
				
				// Calculate how much of the original input to consume
				consumed := len(input) - len(trimmedInput) + closeBracePos + 1
				
				return Result{
					Value:      node,
					Remaining:  input[consumed:],
					Successful: true,
					Error:      "",
				}
			}
		}
		
		return Result{nil, input, false, "not an if directive", false}
	}
}

// ElseIfParser parses {else if condition} or {:else if condition} directives with various whitespace patterns
func ElseIfParser() Parser {
	return func(input string) Result {
		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")
		
		// Try patterns with various whitespace arrangements
		elseIfPatterns := []string{"{else if ", "{:else if ", "{ else if ", "{ :else if "}
		
		for _, pattern := range elseIfPatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				// Find the matching closing brace
				closeBracePos := strings.Index(trimmedInput, "}")
				if closeBracePos < 0 {
					continue // No proper closing brace found, try next pattern
				}
				
				// Extract condition
				condition := trimmedInput[len(pattern):closeBracePos]
				condition = strings.TrimSpace(condition)
				
				// Create ElseIfNode
				node := &ast.ElseIfNode{
					Condition: condition,
				}
				
				// Calculate how much of the original input to consume
				consumed := len(input) - len(trimmedInput) + closeBracePos + 1
				
				return Result{
					Value:      node,
					Remaining:  input[consumed:],
					Successful: true,
					Error:      "",
				}
			}
		}
		
		return Result{nil, input, false, "not an else-if directive", false}
	}
}

// ElseParser parses {else} or {:else} directives with various whitespace patterns
func ElseParser() Parser {
	return func(input string) Result {
		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")
		
		// Try patterns with various whitespace arrangements
		elsePatterns := []string{"{else}", "{:else}", "{ else }", "{ :else }"}
		
		for _, pattern := range elsePatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				node := &ast.ElseNode{}
				
				// Calculate how much of the original input to consume
				consumed := len(input) - len(trimmedInput) + len(pattern)
				
				return Result{
					Value:      node,
					Remaining:  input[consumed:],
					Successful: true,
					Error:      "",
				}
			}
		}
		
		return Result{nil, input, false, "not an else directive", false}
	}
}

// IfEndParser parses {/if}, {/#if}, or {end} directives with various whitespace patterns
func IfEndParser() Parser {
	return func(input string) Result {
		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")
		
		// Try patterns with various whitespace arrangements
		endPatterns := []string{"{/if}", "{/#if}", "{ /if }", "{ /#if }"}
		
		for _, pattern := range endPatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				node := &ast.IfEndNode{}
				
				// Calculate how much of the original input to consume
				consumed := len(input) - len(trimmedInput) + len(pattern)
				
				return Result{
					Value:      node,
					Remaining:  input[consumed:],
					Successful: true,
					Error:      "",
				}
			}
		}
		
		// Check specifically for {end} with flexible whitespace
		if strings.HasPrefix(trimmedInput, "{") {
			i := 1
			// Skip whitespace after opening bracket
			for i < len(trimmedInput) && (trimmedInput[i] == ' ' || trimmedInput[i] == '\t') {
				i++
			}
			
			// Check for "end" keyword
			if i+3 <= len(trimmedInput) && trimmedInput[i:i+3] == "end" {
				i += 3
				
				// Skip whitespace before closing bracket
				for i < len(trimmedInput) && (trimmedInput[i] == ' ' || trimmedInput[i] == '\t') {
					i++
				}
				
				// Check for closing bracket
				if i < len(trimmedInput) && trimmedInput[i] == '}' {
					node := &ast.IfEndNode{}
					
					// Calculate how much of the original input to consume
					consumed := len(input) - len(trimmedInput) + i + 1
					
					return Result{
						Value:      node,
						Remaining:  input[consumed:],
						Successful: true,
						Error:      "",
					}
				}
			}
		}
		
		return Result{nil, input, false, "not an if-end directive", false}
	}
}

// ForStartParser parses {for ...} or {#each ...} directives with various whitespace patterns
func ForStartParser() Parser {
	return func(input string) Result {
		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")
		
		// First try to match {for ...} syntax
		forStart := -1
		forPatterns := []string{"{for ", "{ for "}
		
		for _, pattern := range forPatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				forStart = 0
				break
			}
		}
		
		if forStart >= 0 {
			// Find the matching closing brace
			closeBracePos := strings.Index(trimmedInput, "}")
			if closeBracePos < 0 {
				return Result{nil, input, false, "no closing brace for for directive", false}
			}
			
			// Extract the for expression
			forExpr := trimmedInput[forStart+len("{for"):closeBracePos]
			forExpr = strings.TrimSpace(forExpr)
			
			// Parse "item in items" or "item, index in items" pattern
			parts := strings.Split(forExpr, " in ")
			if len(parts) != 2 {
				return Result{nil, input, false, "invalid for expression: " + forExpr, false}
			}
			
			itemPart := strings.TrimSpace(parts[0])
			collectionPart := strings.TrimSpace(parts[1])
			
			// Check if we have an index variable
			var itemVar, indexVar string
			if strings.Contains(itemPart, ",") {
				itemParts := strings.Split(itemPart, ",")
				itemVar = strings.TrimSpace(itemParts[0])
				if len(itemParts) > 1 {
					indexVar = strings.TrimSpace(itemParts[1])
				}
			} else {
				itemVar = itemPart
			}
			
			node := &ast.Loop{
				Collection: collectionPart,
				Value:      itemVar,
				Iterator:   indexVar,
				Content:    []ast.Node{},
				IsOf:       false, // Using "in" semantics
			}
			
			// Calculate how much of the original input to consume
			consumed := len(input) - len(trimmedInput) + closeBracePos + 1
			
			return Result{
				Value:      node,
				Remaining:  input[consumed:],
				Successful: true,
				Error:      "",
			}
		}
		
		// Next try to match {#each ...} syntax
		eachStart := -1
		eachPatterns := []string{"{#each ", "{ #each "}
		
		for _, pattern := range eachPatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				eachStart = 0
				break
			}
		}
		
		if eachStart >= 0 {
			// Find the matching closing brace
			closeBracePos := strings.Index(trimmedInput, "}")
			if closeBracePos < 0 {
				return Result{nil, input, false, "no closing brace for each directive", false}
			}
			
			// Extract the each expression
			eachExpr := trimmedInput[eachStart+len("{#each"):closeBracePos]
			eachExpr = strings.TrimSpace(eachExpr)
			
			// Parse "items as item" or "items as item, index" pattern
			parts := strings.Split(eachExpr, " as ")
			if len(parts) != 2 {
				return Result{nil, input, false, "invalid each expression: " + eachExpr, false}
			}
			
			collectionPart := strings.TrimSpace(parts[0])
			itemPart := strings.TrimSpace(parts[1])
			
			// Check if we have an index variable
			var itemVar, indexVar string
			if strings.Contains(itemPart, ",") {
				itemParts := strings.Split(itemPart, ",")
				itemVar = strings.TrimSpace(itemParts[0])
				if len(itemParts) > 1 {
					indexVar = strings.TrimSpace(itemParts[1])
				}
			} else {
				itemVar = itemPart
			}
			
			node := &ast.Loop{
				Collection: collectionPart,
				Value:      itemVar,
				Iterator:   indexVar,
				Content:    []ast.Node{},
				IsOf:       true, // Using "as" semantics (similar to "of")
			}
			
			// Calculate how much of the original input to consume
			consumed := len(input) - len(trimmedInput) + closeBracePos + 1
			
			return Result{
				Value:      node,
				Remaining:  input[consumed:],
				Successful: true,
				Error:      "",
			}
		}
		
		return Result{nil, input, false, "not a for/each directive", false}
	}
}

// ForEndParser parses {/for}, {/each}, or {end} directive with various whitespace patterns
func ForEndParser() Parser {
	return func(input string) Result {
		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")
		
		// Try patterns with various whitespace arrangements
		endPatterns := []string{
			"{/for}", "{/each}", "{/#for}", "{/#each}",
			"{ /for }", "{ /each }", "{ /#for }", "{ /#each }",
		}
		
		for _, pattern := range endPatterns {
			if strings.HasPrefix(trimmedInput, pattern) {
				node := &ast.ForEndNode{}
				
				// Calculate how much of the original input to consume
				consumed := len(input) - len(trimmedInput) + len(pattern)
				
				return Result{
					Value:      node,
					Remaining:  input[consumed:],
					Successful: true,
					Error:      "",
				}
			}
		}
		
		// Check specifically for {end} with flexible whitespace
		if strings.HasPrefix(trimmedInput, "{") {
			i := 1
			// Skip whitespace after opening bracket
			for i < len(trimmedInput) && (trimmedInput[i] == ' ' || trimmedInput[i] == '\t') {
				i++
			}
			
			// Check for "end" keyword
			if i+3 <= len(trimmedInput) && trimmedInput[i:i+3] == "end" {
				i += 3
				
				// Skip whitespace before closing bracket
				for i < len(trimmedInput) && (trimmedInput[i] == ' ' || trimmedInput[i] == '\t') {
					i++
				}
				
				// Check for closing bracket
				if i < len(trimmedInput) && trimmedInput[i] == '}' {
					node := &ast.ForEndNode{}
					
					// Calculate how much of the original input to consume
					consumed := len(input) - len(trimmedInput) + i + 1
					
					return Result{
						Value:      node,
						Remaining:  input[consumed:],
						Successful: true,
						Error:      "",
					}
				}
			}
		}
		
		return Result{nil, input, false, "not a for/each end directive", false}
	}
}

// findIfEndNode finds the matching end node for an if directive
func findIfEndNode(nodes []ast.Node, startIndex int) int {
	// Count nesting level to handle nested if statements
	nestingLevel := 1
	for i := startIndex + 1; i < len(nodes); i++ {
		switch nodes[i].(type) {
		case *ast.Conditional:
			nestingLevel++
		case *ast.IfEndNode:
			nestingLevel--
			if nestingLevel == 0 {
				return i
			}
		}
	}
	return -1
}

// ProcessDirectives updates the parser to prioritize directives and handle their bodies
func ProcessDirectives(templateText string) *ast.Template {
	// This function will be implemented later
	// It will transform a template with directives into a proper AST
	return nil
}

// ConditionalParser parses if/else-if/else blocks
func ConditionalParser() Parser {
	return func(input string) Result {
		// Try to parse as if statement first
		ifRes := IfStartParser()(input)
		if ifRes.Successful {
			return ifRes
		}
		
		// Try to parse as else-if statement
		elseIfRes := ElseIfParser()(input)
		if elseIfRes.Successful {
			return elseIfRes
		}
		
		// Try to parse as else statement
		elseRes := ElseParser()(input)
		if elseRes.Successful {
			return elseRes
		}
		
		// Try to parse as if-end statement
		ifEndRes := IfEndParser()(input)
		if ifEndRes.Successful {
			return ifEndRes
		}
		
		// No conditional statement found
		return Result{nil, input, false, "not a conditional statement", false}
	}
}

// LoopParser parses for loops
func LoopParser() Parser {
	return func(input string) Result {
		// Try to parse as for loop start
		forStartRes := ForStartParser()(input)
		if forStartRes.Successful {
			return forStartRes
		}
		
		// Try to parse as for loop end
		forEndRes := ForEndParser()(input)
		if forEndRes.Successful {
			return forEndRes
		}
		
		// No loop statement found
		return Result{nil, input, false, "not a loop statement", false}
	}
}
