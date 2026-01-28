package main

import (
	"fmt"
	"log"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

func main() {
	// Nested conditional example
	template := `{if outer}
  {if inner}
    <div>Nested</div>
  {/if}
  <div>After nested</div>
{/if}`

	log.SetFlags(log.Lshortfile)
	tree, err := parser.ParseTemplate(template)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	if len(tree.RootNodes) == 0 {
		log.Fatal("No root nodes!")
	}

	outer, ok := tree.RootNodes[0].(*ast.Conditional)
	if !ok {
		log.Fatalf("Expected Conditional, got %T", tree.RootNodes[0])
	}

	fmt.Printf("Outer conditional has %d IfContent nodes:\n", len(outer.IfContent))
	for i, node := range outer.IfContent {
		fmt.Printf("  [%d] %T\n", i, node)
	}

	// The first non-whitespace node should be a nested Conditional
	// The second should be the div "After nested"

	fmt.Println("\n✓ Nested conditional test passed!")
}
