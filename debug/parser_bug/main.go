package main

import (
	"fmt"
	"log"
	
	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

func main() {
	template := `{for animal of animals}
  {if animal == "cat"}
    <div>Hi {animal}!</div>
  {else}
    <div>Bye {animal}.</div>
  {/if}
  <div class="type-{animal}">{name} likes: {animal}s</div>
  <br>
{/for}`

	log.SetFlags(log.Lshortfile)
	tree, err := parser.ParseTemplate(template)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}
	
	if len(tree.RootNodes) == 0 {
		log.Fatal("No root nodes!")
	}
	
	loop, ok := tree.RootNodes[0].(*ast.Loop)
	if !ok {
		log.Fatalf("Expected Loop, got %T", tree.RootNodes[0])
	}
	
	fmt.Printf("Loop has %d content nodes:\n", len(loop.Content))
	for i, node := range loop.Content {
		fmt.Printf("  [%d] %T\n", i, node)
	}
	
	if len(loop.Content) != 3 {
		log.Fatalf("Expected 3 content nodes (Conditional, Element, Element), got %d", len(loop.Content))
	}
	
	fmt.Println("✓ Parser correctly returns 3 siblings!")
}
