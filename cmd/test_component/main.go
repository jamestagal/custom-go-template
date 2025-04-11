package main

import (
	"fmt"
	"log"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

func main() {
	// Test the component parser
	fmt.Println("Testing Component Parser")
	
	// Test cases
	testCases := []string{
		"<Button />",
		"<Button label=\"Click me\" />",
		"<Button label=\"Click me\" onClick={handleClick} />",
		"<AdminPanel user={currentUser} />",
		"<UserProfile user={currentUser} />",
	}
	
	for _, tc := range testCases {
		fmt.Printf("\n=== Testing Component Parser on: %s ===\n", tc)
		result := parser.ComponentParser()(tc)
		if result.Successful {
			fmt.Printf("SUCCESS: Parsed component\n")
			compNode, ok := result.Value.(*ast.ComponentNode)
			if ok {
				fmt.Printf("Component Name: %s\n", compNode.Name)
				fmt.Printf("Dynamic: %v\n", compNode.Dynamic)
				if len(compNode.Props) > 0 {
					fmt.Printf("Props:\n")
					for _, prop := range compNode.Props {
						fmt.Printf("  - %s: %s (isDynamic: %v, isShorthand: %v)\n", 
							prop.Name, prop.Value, prop.IsDynamic, prop.IsShorthand)
					}
				} else {
					fmt.Printf("No props\n")
				}
			} else {
				fmt.Printf("ERROR: Result is not a ComponentNode\n")
			}
		} else {
			fmt.Printf("FAILED: %s\n", result.Error)
		}
		fmt.Printf("Remaining: %s\n", result.Remaining)
		fmt.Printf("=== End Test ===\n")
	}
	
	// Also test a complete component in a template context
	templateStr := `
<div>
  <h1>{{ title }}</h1>
  {#if isAdmin}
    <AdminPanel user={currentUser} />
  {:else}
    <UserProfile user={currentUser} />
  {/if}
</div>
`
	
	fmt.Printf("\n\n=== Testing Complete Template ===\n")
	templateAST, err := parser.ParseTemplate(templateStr)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	
	// Print the AST structure
	fmt.Printf("Template AST Root Nodes: %d\n", len(templateAST.RootNodes))
	for i, node := range templateAST.RootNodes {
		fmt.Printf("Node %d: %T\n", i, node)
	}
}
